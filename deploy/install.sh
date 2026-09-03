#!/usr/bin/env bash
#
# Installs aido on the Arch box. Run as root, from a checkout, after `make dist`.
#
#   sudo AIDO_MANIFEST_URL=https://.../manifest.json ./deploy/install.sh
#
# Idempotent: re-run it after every change to the units or the binary.
set -euo pipefail

PLAYER="${PLAYER:-aido}"
BINARY_SRC="${BINARY_SRC:-dist/aido-linux-amd64}"
BINARY_DST="/usr/local/bin/aido"
MANIFEST_URL="${AIDO_MANIFEST_URL:-}"

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
install -Dm0644 deploy/aido-session.sh "/home/$PLAYER/.aido-session.sh"
chown "$PLAYER:$PLAYER" "/home/$PLAYER/.aido-session.sh"
profile="/home/$PLAYER/.bash_profile"
touch "$profile"
chown "$PLAYER:$PLAYER" "$profile"
if ! grep -q '.aido-session.sh' "$profile"; then
    printf '\n. "$HOME/.aido-session.sh"\n' >> "$profile"
fi

echo "==> installing systemd units"
install -Dm0644 deploy/systemd/getty@tty1.service.d/autologin.conf \
    /etc/systemd/system/getty@tty1.service.d/autologin.conf
# The unit file names the account, so keep it in step with $PLAYER.
sed -i "s/--autologin [A-Za-z0-9_-]*/--autologin $PLAYER/" \
    /etc/systemd/system/getty@tty1.service.d/autologin.conf
install -Dm0644 deploy/systemd/aido-update.service /etc/systemd/system/aido-update.service
install -Dm0644 deploy/systemd/aido-update.timer /etc/systemd/system/aido-update.timer

echo "==> writing /etc/aido/aido.env"
install -d -m0755 /etc/aido
if [[ -n "$MANIFEST_URL" ]]; then
    printf 'AIDO_MANIFEST_URL=%s\nAIDO_PLAYER=%s\n' "$MANIFEST_URL" "$PLAYER" > /etc/aido/aido.env
elif [[ ! -f /etc/aido/aido.env ]]; then
    # The update unit reads this file unconditionally, so it has to exist even
    # when updates are not configured yet.
    printf 'AIDO_MANIFEST_URL=\nAIDO_PLAYER=%s\n' "$PLAYER" > /etc/aido/aido.env
    echo "    no AIDO_MANIFEST_URL set; update checks stay off until you fill it in"
fi
chmod 0644 /etc/aido/aido.env

echo "==> reloading systemd"
systemctl daemon-reload
systemctl enable --now aido-update.timer
systemctl restart getty@tty1.service

echo
echo "done. switch to tty1 (ctrl-alt-f1) to see the launcher."
echo "a grown-up can still log in on tty2 (ctrl-alt-f2)."
