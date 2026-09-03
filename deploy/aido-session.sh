# Sourced from the player's ~/.bash_profile.
#
# On tty1 this replaces the shell with the launcher, so there is no prompt to
# escape to. On any other tty it does nothing, which leaves a way in for the
# grown-up with the root password.
#
# `exec` means quitting the launcher ends the session; agetty then logs the
# player straight back in and the launcher reappears. That is the intended
# kiosk loop, and it is also how a freshly installed update gets picked up.

if [ "$(tty)" = "/dev/tty1" ]; then
    # A console font with box-drawing and block characters, for the progress
    # bars. Failure is fine: the default font renders everything else.
    setfont ter-116n 2>/dev/null || true

    exec /usr/local/bin/aido
fi
