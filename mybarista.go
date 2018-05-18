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
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/martinlindhe/unit"
	"github.com/soumya92/barista"
	"github.com/soumya92/barista/bar"
	"github.com/soumya92/barista/colors"
	"github.com/soumya92/barista/modules/battery"
	"github.com/soumya92/barista/modules/clock"
	"github.com/soumya92/barista/modules/cputemp"
	"github.com/soumya92/barista/modules/funcs"
	"github.com/soumya92/barista/modules/group"
	"github.com/soumya92/barista/modules/media"
	"github.com/soumya92/barista/modules/meminfo"
	"github.com/soumya92/barista/modules/netspeed"
	"github.com/soumya92/barista/modules/sysinfo"
	"github.com/soumya92/barista/modules/volume"
	"github.com/soumya92/barista/modules/vpn"
	"github.com/soumya92/barista/modules/weather"
	"github.com/soumya92/barista/modules/weather/openweathermap"
	"github.com/soumya92/barista/modules/wlan"
	"github.com/soumya92/barista/outputs"
	"github.com/soumya92/barista/pango"
	"github.com/soumya92/barista/pango/icons/ionicons"
	"github.com/soumya92/barista/pango/icons/material"
	"github.com/soumya92/barista/pango/icons/material_community"
)

var spacer = pango.Span(" ", pango.XXSmall)

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
	var iconAndPosition pango.Node
	if m.PlaybackStatus == media.Playing {
		iconAndPosition = pango.Span(
			pango.Color(colors.Hex("#f70")),
			materialCommunity.Icon("music"),
			spacer,
			formatMediaTime(m.Position()),
			"/",
			formatMediaTime(m.Length),
		)
	} else {
		iconAndPosition = materialCommunity.Icon("music", pango.Color(colors.Hex("#f70"))...)
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
	//material.Load(home("Github/material-design-icons"))
	material.Load(home(".fonts/material-design-icons"))
	materialCommunity.Load(home(".fonts/MaterialDesign-Webfont"))
	//typicons.Load(home(".fonts/typicons.font"))
	//ionicons.Load(home(".fonts/ionicons"))
	//fontawesome.Load(home(".fonts/Font-Awesome"))

	colors.LoadFromMap(map[string]string{
		"good":     "#6d6",
		"degraded": "#dd6",
		"bad":      "#d66",
		"dim-icon": "#777",
	})

	localtime := clock.Local().
		OutputFunc(time.Second, func(now time.Time) bar.Output {
			return outputs.Pango(
				material.Icon("today", pango.Color(colors.Scheme("dim-icon"))...), spacer,
				now.Format("Mon Jan 2 "),
				material.Icon("access-time", pango.Color(colors.Scheme("dim-icon"))...), spacer,
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
	).OutputFunc(func(w weather.Weather) bar.Output {
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
			materialCommunity.Icon(iconName), spacer,
			pango.Textf("%d℃", int(w.Temperature.Celsius())),
			//pango.Span(" (provided by ", w.Attribution, ")", pango.XSmall),
		)
	})

	vol := volume.DefaultMixer().OutputFunc(func(v volume.Volume) bar.Output {
		if v.Mute {
			return outputs.
				Pango(ionicons.Icon("volume-mute"), "MUT").
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
			ionicons.Icon("volume-"+iconName),
			spacer,
			pango.Textf("%2d%%", pct),
		)
	})

	loadAvg := sysinfo.New().OutputFunc(func(s sysinfo.Info) bar.Output {
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

	freeMem := meminfo.New().OutputFunc(func(m meminfo.Info) bar.Output {
		out := outputs.Pango(material.Icon("memory"), outputs.IBytesize(m.Available()))
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
		OutputFunc(func(temp unit.Temperature) bar.Output {
			return outputs.Pango(
				materialCommunity.Icon("fan"), spacer,
				pango.Textf("%2d℃", int(temp.Celsius())),
			)
		})

	net := netspeed.New("eno1").
		RefreshInterval(2 * time.Second).
		OutputFunc(func(s netspeed.Speeds) bar.Output {
			return outputs.Pango(
				materialCommunity.Icon("arrow-up"), spacer, pango.Textf("%5s", outputs.Byterate(s.Tx)),
				pango.Span(" ", pango.Small),
				materialCommunity.Icon("arrow-down"), spacer, pango.Textf("%5s", outputs.Byterate(s.Rx)),
			)
		})

	audioplayer := media.New("DeaDBeeF").OutputFunc(mediaFormatFunc)
	wifi := wlan.New("wlp2s0").OutputFunc(func(w wlan.Info) bar.Output {
		if w.SSID != "" {
			return outputs.Pango(
				material.Icon("wifi"), spacer,
				pango.Textf("%s", w.SSID),
			)
		}
		return outputs.Empty()
	})
	vpn := vpn.New("vpn0")

	batt := battery.Default().OutputFunc(func(b battery.Info) bar.Output {
		var pstate pango.Node
		if b.PluggedIn() {
			pstate = materialCommunity.Icon("power-plug")
		} else {
			pstate = materialCommunity.Icon("power-plug-off")
		}
		out := outputs.Pango(pstate, pango.Textf("%d%%", b.RemainingPct()))
		switch {
		case b.RemainingTime() < time.Duration(5)*time.Minute:
			out.Urgent(true)
		case b.RemainingTime() < time.Duration(10)*time.Minute:
			out.Color(colors.Scheme("bad"))
		case b.RemainingTime() < time.Duration(30)*time.Minute:
			out.Color(colors.Scheme("degraded"))
		case b.RemainingTime() > time.Duration(45)*time.Minute:
			out.Color(colors.Scheme("good"))
		}
		return out
	})

	thirtySeconds, _ := time.ParseDuration("30s")
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
			materialCommunity.Icon("lightbulb"),
			spacer,
			pango.Textf("%.0f%%", backlightPct),
		))
	})

	g := group.Collapsing()

	panic(barista.Run(
		audioplayer,
		g.Add(wifi),
		g.Add(vpn),
		g.Add(net),
		g.Add(temp),
		g.Add(backlight),
		g.Add(freeMem),
		g.Button(outputs.Text("+"), outputs.Text("-")),
		loadAvg,
		vol,
		wthr,
		batt,
		localtime,
	))
}
