package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

var defStyle tcell.Style

func bar_path() string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return filepath.Join(dir, "/.bars.yaml")
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

func days_since(s string) int {
	const layout = "2006-01-02"
	t, _ := time.Parse(layout, s)
	return int((time.Since(t)).Hours() / (24))
}


func render_bars(s tcell.Screen, max_bar_length int, bars []map[string]interface{}) {


	green := tcell.StyleDefault.Foreground(tcell.ColorLawnGreen)
	yellow := tcell.StyleDefault.Foreground(tcell.Color184)
	orange := tcell.StyleDefault.Foreground(tcell.ColorDarkOrange)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)

	theme := []string{"█", " ", "|", "▐"}

	index := 3
	maxBarLength := max_bar_length

	emitStr(s, 1, 2, blue, "STATUS BAR")

	for _, el := range bars {
		length := el["length"].(int)


		days_since := days_since(el["start_date"].(string))

		if days_since > maxBarLength {
			days_since = maxBarLength
		}

		neg_length := maxBarLength-length
		if neg_length < 0 {
			neg_length = maxBarLength
		}

		neg_day_length := maxBarLength-days_since
		if neg_day_length < 0 {
			neg_length = maxBarLength
		}

		barString := fmt.Sprintf("\r %s%s%s",
			theme[3],
			strings.Repeat(theme[0], length),
			" + | -",
		)
		errorBarString := fmt.Sprintf("\r %s%s%s",
			theme[3],
			strings.Repeat(theme[0], days_since),
			strings.Repeat(theme[1], neg_day_length),
		)
		name_string := fmt.Sprintf("\r %s (%v) ", el["name"].(string), length)

		var bar_color = green
		if days_since <= length {
			bar_color = green
		} else {
			if (days_since - length) < 5 {
				bar_color = yellow
			} else if (days_since - length) < 10 {
				bar_color = orange
			} else {
				bar_color = red
			}

		}
		emitStr(s, 2, index + 1, bar_color, name_string)
		emitStr(s, len(name_string) + 1, index + 1, blue, fmt.Sprintf("   %v day(s)", days_since))
		emitStr(s, 2, index + 2, bar_color, fmt.Sprintf(barString))
		emitStr(s, 2, index + 3, blue, fmt.Sprintf(errorBarString))
		index += 4
	}
}

func inc_bars_by_index(max_bar_length int, index int,
	bars []map[string]interface{}) []map[string]interface{}{
	new_bars := []map[string]interface{}{}
	for i, el := range bars {
		inc := el["inc"].(int)
		if (i + 1) == index {
			new_length := el["length"].(int) + inc
			if new_length > max_bar_length {
				el["length"] = max_bar_length
			} else {
				el["length"] = new_length
			}
			new_bars = append(new_bars, el)
		} else {
			new_bars = append(new_bars, el)
		}
	}
	save_bars(new_bars)


	return new_bars
}

func dec_bars_by_index(max_bar_length int, index int,
	bars []map[string]interface{}) []map[string]interface{}{
	new_bars := []map[string]interface{}{}
	for i, el := range bars {
		inc := el["inc"].(int)
		if (i+1) == index {
			new_length := el["length"].(int) - inc
			if new_length < 0 {
				el["length"] = max_bar_length
			} else {
				el["length"] = new_length
			}
			new_bars = append(new_bars, el)
		} else {
			new_bars = append(new_bars, el)
		}
	}
	save_bars(new_bars)

	return new_bars
}


func save_bars(bars []map[string]interface{}) {
	b := make(map[string]interface{})
	b["bars"] = bars
	d, err := yaml.Marshal(b)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	f, err := os.Create(bar_path())
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	f.Write(d)
}


func inc_dec_bars(max_bar_length int, x int, y int,
	bars []map[string]interface{}) []map[string]interface{}{
	new_bars := []map[string]interface{}{}
	for i, el := range bars {
		if 4 * (i + 1) + 1 == y {
			length := el["length"].(int)
			inc := el["inc"].(int)
			if x <=  (length + 6){
				new_length := el["length"].(int) + inc
				if new_length > max_bar_length {
					el["length"] = max_bar_length
				} else {
					el["length"] = new_length
				}
				new_bars = append(new_bars, el)
			} else if x > (length + 6){
				new_length := el["length"].(int) - inc
				if new_length < 0 {
					el["length"] = 0
				} else {
					el["length"] = new_length
				}

				new_bars = append(new_bars, el)
			} else {
				new_bars = append(new_bars, el)
			}
		} else {
			new_bars = append(new_bars, el)
		}
	}

	save_bars(new_bars)

	return new_bars
}

func get_bars() ([]map[string]interface{}, error) {
	yamlFile, err := ioutil.ReadFile(bar_path())

	type BarConfig struct {
		Bars []map[string]interface{}
	}

	bars := BarConfig{}
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &bars)
	}

	return bars.Bars, err
}



func main() {

	s, e := tcell.NewScreen()

	encoding.Register()

	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	defStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorReset)

	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	max_bar_length := 60

	ecnt := 0

	bars, err := get_bars()
	if err != nil {
		s.Fini()
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(0)
	}

	render_bars(s, max_bar_length, bars)
	s.Show()


	for {
		ev := s.PollEvent()

		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				ecnt++
				if ecnt > 1 {
					s.Fini()
					os.Exit(0)
				}
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			switch ev.Buttons() {
			case tcell.Button1, tcell.Button2, tcell.Button3:
				s.Clear()
				new_bars := inc_dec_bars(max_bar_length, x, y, bars)
				render_bars(s, max_bar_length, new_bars)
				s.Sync()
				s.Show()
			}
		}

	}


}
