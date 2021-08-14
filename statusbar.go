package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/mattn/go-runewidth"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var defStyle tcell.Style

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

func days_since_string(i int) string {
	return fmt.Sprintf("%v day(s)", i)
}

func get_max_length(bars []map[string]interface{}) int {
	maxLength := 0
	for _, el := range bars {
		length := el["length"].(int)
		name := fmt.Sprintf("%s (%v)", el["name"].(string), length)
		start_date := el["start_date"].(string)
		days_since := days_since(start_date)
		days_since_string := days_since_string(days_since)

		l := len(name)
		l2 := len(days_since_string)
		if l > maxLength {
			maxLength = l
		}
		if l2 > maxLength {
			maxLength = l2
		}
	}
	return maxLength

}


func render_bars(s tcell.Screen, max_bar_length int, bars []map[string]interface{}) {

	green := tcell.StyleDefault.Foreground(tcell.ColorGreenYellow)
	red := tcell.StyleDefault.Foreground(tcell.ColorRed)
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	//purple := tcell.StyleDefault.Foreground(tcell.ColorPurple)

	theme := []string{"█", " ", "|", "|"}

	index := 3
	maxLength := get_max_length(bars)
	maxBarLength := max_bar_length

	emitStr(s, 2, 2, green, "STATUS BAR")


	for _, el := range bars {
		length := el["length"].(int)



		name := fmt.Sprintf("%s (%v)", el["name"].(string), length)

		days_since := days_since(el["start_date"].(string))

		if days_since > maxBarLength {
			days_since = maxBarLength
		}

		days_since_string := days_since_string(days_since)
		l := len(name)


		neg_length := maxBarLength-length
		if neg_length < 0 {
			neg_length = maxBarLength
		}

		neg_day_length := maxBarLength-days_since
		if neg_day_length < 0 {
			neg_length = maxBarLength
		}

		barString := fmt.Sprintf("\r %s%s%s",
			" "+name + strings.Repeat(" ", (maxLength + 2) - l),
			strings.Repeat(theme[0], length),
			strings.Repeat(theme[1], neg_length) + " - / +",
		)
		l2 := len(days_since_string)
		errorBarString := fmt.Sprintf("\r %s%s%s",
			" " + days_since_string + strings.Repeat(" ", (maxLength + 2) - l2),
			strings.Repeat(theme[0], days_since),
			strings.Repeat(theme[1], neg_day_length),
		)
		emitStr(s, 2, index + 1, blue, fmt.Sprintf(barString))
		emitStr(s, 2, index + 2, red, fmt.Sprintf(errorBarString))
		emitStr(s, 2, index + 3, red, "\r")
		index += 3
	}
}

func inc_dec_bars(max_bar_length int, x int, y int,
	bars []map[string]interface{}) []map[string]interface{}{
	new_bars := []map[string]interface{}{}
	for i, el := range bars {
		if 3 * (i + 1) + 1 == y {
			if x >  (max_bar_length + get_max_length(bars) + 8){
				new_length := el["length"].(int) +1
				if new_length > max_bar_length {
					el["length"] = max_bar_length
				} else {
					el["length"] = new_length
				}
				new_bars = append(new_bars, el)
			} else if x == (max_bar_length + get_max_length(bars) + 5){
				new_length := el["length"].(int) - 1
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
	return new_bars
}

func get_bars() ([]map[string]interface{}, error) {
	yamlFile, err := ioutil.ReadFile("/Users/kprifogle/.bars.yaml")

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

	max_bar_length := 40

	ecnt := 0

	bars, err := get_bars()
	if err != nil {
		s.Fini()
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(0)
	}

	render_bars(s, max_bar_length, bars)

	for {
		ev := s.PollEvent()
		s.Show()
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
			case tcell.Button1:
				s.Clear()
				new_bars := inc_dec_bars(max_bar_length, x, y, bars)
				render_bars(s, max_bar_length, new_bars)
				s.Show()
				s.Sync()
			}
		}

	}


}
