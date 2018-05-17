# Mybarista

Mybarista is my implementation of a i3 status bar using [Barista](https://github.com/soumya92/barista)
As this is my personal status bar i use daily it is subject to change at any time.

Current Fields:

 * rhythmbox
 * wifi
 * vpn
 * net
 * temperature
 * backlight
 * Memory
 * Load
 * Volume
 * Weather
 * Battery
 * Date/Time


## Installation

`go get https://github.com/toke/mybarista`

I actually use [dep](https://golang.github.io/dep/) for tracking dependencies,
use `dep ensure` and `go build` for using this exact versions.

i3 / sway configuraion:
```
bar {
    …
    status_command $i3helper/mybarista
    …
}
```

## Other work

 * https://github.com/glebtv/custom_barista

