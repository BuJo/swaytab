# swaytab

Attempts to recreate the alt-tab behaviour from Openbox for sway.

### Building

    go build .

### Usage

See `doc/sway.conf` for a sway configuration.  You can copy this to an
appropriate location (e.g. `~/.config/sway/config.d/alttab`.

Then simply run the daemon.

### TODO

* Does not detect `alt`-key up, so cannnot use that to detect finished cycle
* First window after toggling when moving to cycle mod is wrong
* Focuses windows outside of current workspace

