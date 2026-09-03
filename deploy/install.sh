#!/usr/bin/env bash
#
# Installs unlock on the Arch box. Run as root, from a checkout, after `make dist`.
#
#   sudo UNLOCK_MANIFEST_URL=https://.../manifest.json ./deploy/install.sh
#
# Idempotent: re-run it after every change to the units or the binary.
set -euo pipefail

PLAYER="${PLAYER:-player}"
BINARY_SRC="${BINARY_SRC:-dist/unlock-linux-amd64}"
BINARY_DST="/usr/local/bin/unlock"
MANIFEST_URL="${UNLOCK_MANIFEST_URL:-}"

[[ $EUID -eq 0 ]] || { echo "run me as root" >&2; exit 1; }
[[ -f "$BINARY_SRC" ]] || { echo "missing $BINARY_SRC; run make dist first" >&2; exit 1; }

echo "==> installing $BINARY_DST"
install -Dm0755 "$BINARY_SRC" "$BINARY_DST"

echo "==> ensuring the $PLAYER account exists"
if ! id -u "$PLAYER" >/dev/null 2>&1; then
    # No sudo, no wheel. The ladder is the only way up.
    useradd --create-home --shell /bin/bash "$PLAYER"
    passwd --delete "$PLAYER"
fi

echo "==> installing the launcher session"
install -Dm0644 deploy/unlock-session.sh "/home/$PLAYER/.unlock-session.sh"
chown "$PLAYER:$PLAYER" "/home/$PLAYER/.unlock-session.sh"
profile="/home/$PLAYER/.bash_profile"
touch "$profile"
chown "$PLAYER:$PLAYER" "$profile"
if ! grep -q '.unlock-session.sh' "$profile"; then
    printf '\n. "$HOME/.unlock-session.sh"\n' >> "$profile"
fi

echo "==> installing systemd units"
install -Dm0644 deploy/systemd/getty@tty1.service.d/autologin.conf \
    /etc/systemd/system/getty@tty1.service.d/autologin.conf
# The unit file names the account, so keep it in step with $PLAYER.
sed -i "s/--autologin [A-Za-z0-9_-]*/--autologin $PLAYER/" \
    /etc/systemd/system/getty@tty1.service.d/autologin.conf
install -Dm0644 deploy/systemd/unlock-update.service /etc/systemd/system/unlock-update.service
install -Dm0644 deploy/systemd/unlock-update.timer /etc/systemd/system/unlock-update.timer

echo "==> writing /etc/unlock/unlock.env"
install -d -m0755 /etc/unlock
if [[ -n "$MANIFEST_URL" ]]; then
    printf 'UNLOCK_MANIFEST_URL=%s\nUNLOCK_PLAYER=%s\n' "$MANIFEST_URL" "$PLAYER" > /etc/unlock/unlock.env
elif [[ ! -f /etc/unlock/unlock.env ]]; then
    # The update unit reads this file unconditionally, so it has to exist even
    # when updates are not configured yet.
    printf 'UNLOCK_MANIFEST_URL=\nUNLOCK_PLAYER=%s\n' "$PLAYER" > /etc/unlock/unlock.env
    echo "    no UNLOCK_MANIFEST_URL set; update checks stay off until you fill it in"
fi
chmod 0644 /etc/unlock/unlock.env

echo "==> reloading systemd"
systemctl daemon-reload
systemctl enable --now unlock-update.timer
systemctl restart getty@tty1.service

echo
echo "done. switch to tty1 (ctrl-alt-f1) to see the launcher."
echo "a grown-up can still log in on tty2 (ctrl-alt-f2)."
