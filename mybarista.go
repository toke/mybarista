// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// sample-bar demonstrates a sample i3bar built using barista.
package main

import (
	"fmt"
	"image/color"
	"math"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"github.com/martinlindhe/unit"
	"github.com/soumya92/barista"
	"github.com/soumya92/barista/bar"
	"github.com/soumya92/barista/colors"
	"github.com/soumya92/barista/modules/battery"
	"github.com/soumya92/barista/modules/clock"
	"github.com/soumya92/barista/modules/cputemp"
	"github.com/soumya92/barista/modules/group"
	"github.com/soumya92/barista/modules/media"
	"github.com/soumya92/barista/modules/meminfo"
	"github.com/soumya92/barista/modules/netinfo"
	"github.com/soumya92/barista/modules/netspeed"
	"github.com/soumya92/barista/modules/sysinfo"
	"github.com/soumya92/barista/modules/volume"
	"github.com/soumya92/barista/modules/vpn"
	"github.com/soumya92/barista/modules/weather"
	"github.com/soumya92/barista/modules/weather/openweathermap"
	"github.com/soumya92/barista/modules/wlan"
	"github.com/soumya92/barista/outputs"
	"github.com/soumya92/barista/pango"
	"github.com/soumya92/barista/pango/icons/fontawesome"
	"github.com/soumya92/barista/pango/icons/ionicons"
	"github.com/soumya92/barista/pango/icons/material"
	"github.com/soumya92/barista/pango/icons/mdi"
	"github.com/soumya92/barista/pango/icons/typicons"
)

var spacer = pango.Text(" ").XXSmall()

func truncate(in string, l int) string {
	if len([]rune(in)) <= l {
		return in
	}
	return string([]rune(in)[:l-1]) + "⋯"
}

func hms(d time.Duration) (h int, m int, s int) {
	h = int(d.Hours())
	m = int(d.Minutes()) % 60
	s = int(d.Seconds()) % 60
	return
}

func formatMediaTime(d time.Duration) string {
	h, m, s := hms(d)
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func mediaFormatFunc(m media.Info) bar.Output {
	if m.PlaybackStatus == media.Stopped || m.PlaybackStatus == media.Disconnected {
		return nil
	}
	artist := truncate(m.Artist, 20)
	title := truncate(m.Title, 40-len(artist))
	if len(title) < 20 {
		artist = truncate(m.Artist, 40-len(title))
	}
	iconAndPosition := pango.Icon("fa-music").Color(colors.Hex("#f70"))
	if m.PlaybackStatus == media.Playing {
		iconAndPosition.Append(
			spacer, pango.Textf("%s/%s",
				formatMediaTime(m.Position()),
				formatMediaTime(m.Length)),
		)
	}
	return outputs.Pango(iconAndPosition, spacer, title, " - ", artist)
}

func startTaskManager(e bar.Event) {
	if e.Button == bar.ButtonLeft {
		exec.Command("xfce4-taskmanager").Run()
	}
}

func home(path string) string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return filepath.Join(usr.HomeDir, path)
}

func main() {
	// See https://materialdesignicons.com/ for a mdi reference
	material.Load(home(".fonts/material-design-icons"))
	mdi.Load(home(".fonts/MaterialDesign-Webfont"))
	typicons.Load(home(".fonts/typicons.font"))
	ionicons.LoadMd(home(".fonts/ionicons"))
	fontawesome.Load(home(".fonts/Font-Awesome"))

	colors.LoadFromMap(map[string]string{
		"good":     "#6d6",
		"degraded": "#dd6",
		"bad":      "#d66",
		"dim-icon": "#777",
	})

	localtime := clock.Local().
		Output(time.Second, func(now time.Time) bar.Output {
			return outputs.Pango(
				pango.Icon("material-today").Color(colors.Scheme("dim-icon")),
				spacer,
				now.Format("Mon Jan 2 "),
				pango.Icon("material-access-time").Color(colors.Scheme("dim-icon")),
				spacer,
				now.Format("15:04:05"),
			)
		})
	localtime.OnClick(func(e bar.Event) {
		if e.Button == bar.ButtonLeft {
			exec.Command("gsimplecal").Run()
		}
	})

	// Weather information comes from OpenWeatherMap.
	// https://openweathermap.org/api.
	wthr := weather.New(
		openweathermap.Zipcode("76131", "DE").Build(),
	).Output(func(w weather.Weather) bar.Output {
		iconName := ""
		switch w.Condition {
		case weather.Thunderstorm,
			weather.TropicalStorm,
			weather.Hurricane:
			iconName = "stormy"
		case weather.Drizzle,
			weather.Hail:
			iconName = "shower"
		case weather.Rain:
			iconName = "downpour"
		case weather.Snow,
			weather.Sleet:
			iconName = "snow"
		case weather.Mist,
			weather.Smoke,
			weather.Whirls,
			weather.Haze,
			weather.Fog:
			iconName = "windy-cloudy"
		case weather.Clear:
			if !w.Sunset.IsZero() && time.Now().After(w.Sunset) {
				iconName = "night"
			} else {
				iconName = "sunny"
			}
		case weather.PartlyCloudy:
			iconName = "partly-sunny"
		case weather.Cloudy, weather.Overcast:
			iconName = "cloudy"
		case weather.Tornado,
			weather.Windy:
			iconName = "windy"
		}
		if iconName == "" {
			iconName = "warning-outline"
		} else {
			iconName = "weather-" + iconName
		}
		return outputs.Pango(
			pango.Icon("typecn-"+iconName), spacer,
			pango.Textf("%d℃", int(w.Temperature.Celsius())),
		)
	})

	vol := volume.DefaultMixer().Output(func(v volume.Volume) bar.Output {
		if v.Mute {
			return outputs.
				Pango(pango.Icon("mdi-volume-mute"), "MUT").
				Color(colors.Scheme("degraded"))
		}
		iconName := "low"
		pct := v.Pct()
		if pct > 66 {
			iconName = "high"
		} else if pct > 33 {
			iconName = "medium"
		}
		return outputs.Pango(
			pango.Icon("mdi-volume-"+iconName),
			spacer,
			pango.Textf("%2d%%", pct),
		)
	})

	loadAvg := sysinfo.New().Output(func(s sysinfo.Info) bar.Output {
		out := outputs.Textf("%0.2f %0.2f", s.Loads[0], s.Loads[2])
		// Load averages are unusually high for a few minutes after boot.
		if s.Uptime < 10*time.Minute {
			// so don't add colours until 10 minutes after system start.
			return out
		}
		switch {
		case s.Loads[0] > 128, s.Loads[2] > 64:
			out.Urgent(true)
		case s.Loads[0] > 64, s.Loads[2] > 32:
			out.Color(colors.Scheme("bad"))
		case s.Loads[0] > 32, s.Loads[2] > 16:
			out.Color(colors.Scheme("degraded"))
		}
		return out
	})
	loadAvg.OnClick(startTaskManager)

	freeMem := meminfo.New().Output(func(m meminfo.Info) bar.Output {
		out := outputs.Pango(pango.Icon("mdi-memory"), spacer, outputs.IBytesize(m.Available()))
		freeGigs := m.Available().Gigabytes()
		switch {
		case freeGigs < 0.5:
			out.Urgent(true)
		case freeGigs < 1:
			out.Color(colors.Scheme("bad"))
		case freeGigs < 2:
			out.Color(colors.Scheme("degraded"))
		case freeGigs > 12:
			out.Color(colors.Scheme("good"))
		}
		return out
	})
	freeMem.OnClick(startTaskManager)

	temp := cputemp.DefaultZone().
		RefreshInterval(2 * time.Second).
		UrgentWhen(func(temp unit.Temperature) bool {
			return temp.Celsius() > 90
		}).
		OutputColor(func(temp unit.Temperature) color.Color {
			switch {
			case temp.Celsius() > 70:
				return colors.Scheme("bad")
			case temp.Celsius() > 60:
				return colors.Scheme("degraded")
			default:
				return nil
			}
		}).
		Output(func(temp unit.Temperature) bar.Output {
			return outputs.Pango(
				pango.Icon("mdi-fan"), spacer,
				pango.Textf("%2d℃", int(temp.Celsius())),
			)
		})

	net := netspeed.New("eno1").
		RefreshInterval(2 * time.Second).
		Output(func(s netspeed.Speeds) bar.Output {
			return outputs.Pango(
				pango.Icon("mdi-arrow-up"), spacer, pango.Textf("%5s", outputs.Byterate(s.Tx)),
				spacer,
				pango.Icon("mdi-arrow-down"), spacer, pango.Textf("%5s", outputs.Byterate(s.Rx)),
			)
		})

	audioplayer := media.New("DeaDBeeF").Output(mediaFormatFunc)

	wifi := wlan.Any().Output(func(w wlan.Info) bar.Output {
		switch {
		case w.Connected():
			out := fmt.Sprintf("%s", w.SSID)
			return outputs.Pango(pango.Icon("mdi-wifi"), spacer, pango.Text(out))
		case w.Connecting():
			return outputs.Pango(pango.Icon("mdi-wifi-off"), spacer, pango.Text("connecting…"))
		case w.Enabled():
			return outputs.Pango(pango.Icon("mdi-wifi-off"), spacer, pango.Text("down"))
		default:
			return nil
		}
	})

	vpn := vpn.New("vpn0")

	netinfo := netinfo.New().Output(func(s netinfo.State) bar.Output {
		if !s.Enabled() {
			return nil
		}
		for _, ip := range s.IPs {
			if ip.To4() == nil && ip.IsGlobalUnicast() {
				return outputs.Text(ip.String()).Color(colors.Scheme("good"))
			}
		}
		return outputs.Text("no IPv6").Color(colors.Scheme("bad"))
	})

	statusName := map[battery.Status]string{
		battery.Charging:    "CHR",
		battery.Discharging: "BAT",
		battery.NotCharging: "NOT",
		battery.Unknown:     "UNK",
	}

	batt := battery.All().Output(func(b battery.Info) bar.Output {
		// Round to the nearest decimal
		perc := math.Round(float64(b.RemainingPct())/float64(10)) * float64(10)
		mode := ""
		if b.Discharging() {
			mode = ""
		} else {
			mode = "-charging"
		}
		icon := fmt.Sprintf("mdi-battery%s-%.0f", mode, perc)

		if icon == "mdi-battery-100" {
			icon = "mdi-battery"
		} else if icon == "mdi-battery-0" {
			icon = "mdi-battery-outline"
		}

		if b.Status == battery.Disconnected {
			return nil
		}

		out := outputs.Pango(
			pango.Icon(icon),
			pango.Textf("%s %d%% %s",
				statusName[b.Status],
				b.RemainingPct(),
				b.RemainingTime(),
			),
		)
		if b.Discharging() {
			if b.RemainingPct() < 20 || b.RemainingTime() < 30*time.Minute {
				out.Color(colors.Scheme("bad"))
			}
		}
		return out
	})

	/* 	thirtySeconds, _ := time.ParseDuration("30s")
	   	backlight := funcs.Every(thirtySeconds, func(m funcs.Channel) {
	   		maxBrightnessBytes, err := ioutil.ReadFile("/sys/class/backlight/intel_backlight/max_brightness")
	   		brightnessBytes, err := ioutil.ReadFile("/sys/class/backlight/intel_backlight/brightness")
	   		if err != nil {
	   			m.Error(err)
	   			return
	   		}
	   		level, err := strconv.ParseFloat(
	   			strings.Trim(string(brightnessBytes), "\n "),
	   			32,
	   		)
	   		maxLevel, err := strconv.ParseFloat(
	   			strings.Trim(string(maxBrightnessBytes), "\n "),
	   			32,
	   		)

	   		if err != nil {
	   			m.Error(err)
	   			return
	   		}
	   		backlightPct := (level / maxLevel) * 100

	   		m.Output(outputs.Pango(
	   			pango.Icon("lightbulb"),
	   			spacer,
	   			pango.Textf("%.0f%%", backlightPct),
	   		))
	   	}) */

	g := group.Collapsing()

	panic(barista.Run(
		audioplayer,
		g.Add(wifi),
		g.Add(vpn),
		g.Add(net),
		g.Add(netinfo),
		g.Add(temp),
		//g.Add(backlight),
		g.Add(freeMem),
		g.Button(outputs.Pango(pango.Icon("mdi-chevron-left")), outputs.Pango(pango.Icon("mdi-chevron-right"))),
		loadAvg,
		vol,
		wthr,
		batt,
		localtime,
	))
}
