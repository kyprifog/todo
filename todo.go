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
	"sort"
	"strings"
	"time"
)

var defStyle tcell.Style

func todos_path() string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	return filepath.Join(dir, "/.todos.yaml")
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
func sort_list(todos []map[string]interface{}) {
	sort.Slice(todos, func(i, j int) bool {
		return todos[i]["name"].(string) < todos[j]["name"].(string)
	})
}

func sort_todo(todos []map[string]interface{}) [][]map[string]interface{}{
	goal := []map[string]interface{}{}
	important := []map[string]interface{}{}
	todo := []map[string]interface{}{}
	shopping := []map[string]interface{}{}
	done := []map[string]interface{}{}

	for _, el := range todos {
		name := el["name"].(string)
		checked := el["done"].(bool)
		if checked == true {
			done = append(done, el)
		} else if strings.Contains(name, "Goal:") {
			goal = append(goal, el)
		} else if strings.Contains(name, "*") {
			important = append(important, el)
		} else if strings.Contains(name, "Shopping:") {
			shopping = append(shopping, el)
		} else {
			todo = append(todo, el)
		}
	}

	sort_list(goal)
	sort_list(important)
	sort_list(shopping)
	sort_list(todo)
	sort_list(done)

	all_todos := [][]map[string]interface{}{}

	all_todos = append(all_todos, goal)
	all_todos = append(all_todos, important)
	all_todos = append(all_todos, todo)
	all_todos = append(all_todos, shopping)
	all_todos = append(all_todos, done)

	return all_todos
}

func render_todos(s tcell.Screen, todos []map[string]interface{}) {
	green := tcell.StyleDefault.Foreground(tcell.ColorLawnGreen)
	yellow := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	sea_green := tcell.StyleDefault.Foreground(tcell.ColorDarkSeaGreen)
	purple := tcell.StyleDefault.Foreground(tcell.ColorPurple)
	grey := tcell.StyleDefault.Foreground(tcell.ColorGrey)
	orange := tcell.StyleDefault.Foreground(tcell.ColorOrange)
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)

	emitStr(s, 0,0, green, "TODO")

	index := 1

	all_todos := sort_todo(todos)
	goal := all_todos[0]
	important := all_todos[1]
	todo := all_todos[2]
	shopping := all_todos[3]
	done := all_todos[4]

	for _, el := range goal {
		name := el["name"].(string)
		emitStr(s, 0, index, sea_green, "-- " + name + " --")
		index += 1
	}

	for _, el := range important {
		name := el["name"].(string)
		emitStr(s, 0, index, yellow, "[ ] " + name)
		index += 1
	}

	for _, el := range todo {
		name := el["name"].(string)
		emitStr(s, 0, index, blue, "[ ] " + name)
		index += 1
	}

	for _, el := range shopping {
		name := el["name"].(string)
		emitStr(s, 0, index, purple, "[ ] " + name)
		index += 1
	}

	for _, el := range done {
		name := el["name"].(string)
		emitStr(s, 0, index, grey, "[x] " + name + " (-)")
		index += 1
	}

	emitStr(s, 0, index, blue, "")
	index += 1
	emitStr(s, 0, index, orange, "Add +")
}

func add_new_todo(s tcell.Screen, new_todo string) {
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	emitStr(s, 0, 2, blue, "New Todo: " + new_todo)
}

func tick_todos(x int, y int, todos []map[string]interface{}) []map[string]interface{} {

	sorted_todos := sort_todo(todos)
	goal := sorted_todos[0]
	important := sorted_todos[1]
	todo := sorted_todos[2]
	shopping := sorted_todos[3]
	checked := sorted_todos[4]

	all_todos := []map[string]interface{}{}

	index := 1
	for _, el := range goal {
		if index == y {
			el["done"] = true
		}
		index += 1
		all_todos = append(all_todos, el)
	}

	for _, el := range important {
		if index == y {
			el["done"] = true
		}
		index += 1
		all_todos = append(all_todos, el)
	}

	for _, el := range todo {
		if index == y {
			el["done"] = true
		}
		index += 1
		all_todos = append(all_todos, el)
	}

	for _, el := range shopping {
		if index == y {
			el["done"] = true
		}
		index += 1
		all_todos = append(all_todos, el)
	}

	for _, el := range checked {
		if index == y {
			name := el["name"].(string)
			if (len(name) + 6) <= x {
				el["done"] = false
				all_todos = append(all_todos, el)
				index += 1
			}
		} else {
			all_todos = append(all_todos, el)
			index += 1
		}
	}

	save_todos(all_todos)
	return all_todos
}

func save_todos(todos []map[string]interface{}) {
	b := make(map[string]interface{})
	b["todos"] = todos
	d, err := yaml.Marshal(b)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	f, err := os.Create(todos_path())
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	f.Write(d)
}


func get_todos() ([]map[string]interface{}, error) {
	yamlFile, err := ioutil.ReadFile(todos_path())

	type AllToDos struct {
		ToDos []map[string]interface{}
	}

	todos := AllToDos{}
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &todos)
	}

	return todos.ToDos, err
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

	defStyle = tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)

	s.SetStyle(defStyle)
	s.EnableMouse()
	s.EnablePaste()
	s.Clear()

	ecnt := 0
	add_new := false
	new_todo := ""

	todos, err := get_todos()
	if err != nil {
		s.Fini()
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(0)
	}

	render_todos(s, todos)
	s.Show()

	defer s.Fini()
	events := make(chan tcell.Event)
	go func() {
		for {
			ev := s.PollEvent()
			events <- ev
		}
	}()
	go func() {
	for {
		ev := <-events
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape {
				ecnt++
				if ecnt > 1 {
					s.Fini()
					os.Exit(0)
				}
			} else if ev.Key() == tcell.KeyEnter {
				if add_new == true {
					todos, _ := get_todos()
					new_value := make(map[string]interface{})
					new_value["name"] = new_todo
					new_value["done"] = false
					add_new = false
					new_todo = ""
					todos = append(todos, new_value)
					save_todos(todos)
					atodos, _ := get_todos()
					s.Clear()
					render_todos(s, atodos)
					s.Show()
				}
			} else if ev.Key() == tcell.KeyRune {
				if add_new {
					key_value := strings.Replace(strings.Replace(ev.Name(), "Rune[", "", 1), "]",
						"", 1)
					new_todo += key_value
					add_new_todo(s, new_todo)
					add_new = true
					s.Show()
				}
			} else if (ev.Key() == tcell.KeyBackspace2 || ev.Key() == tcell.KeyBackspace) {
				if len(new_todo) > 0 {
					new_todo = strings.TrimSuffix(new_todo, new_todo[len(new_todo)-1:])
					s.Clear()
					add_new_todo(s, new_todo)
					s.Show()
				} else {
					add_new = false
					atodos, _ := get_todos()
					s.Clear()
					render_todos(s, atodos)
					s.Show()
				}
			}
		case *tcell.EventMouse:
			x, y := ev.Position()
			switch ev.Buttons() {
			case tcell.Button1, tcell.Button2, tcell.Button3:
				atodos, _ := get_todos()
				l := len(atodos)
				if y < (l + 2) {
					s.Clear()
					tick_todos(x, y, atodos)
					todos, _ := get_todos()
					render_todos(s, todos)
					s.Show()
				} else if y == l+2 {
					s.Clear()
					add_new_todo(s, new_todo)
					add_new = true
					s.Show()
				}
			}
		}

	}
	}()

	t := time.NewTicker(time.Second)
	for {
		select {
		case <-t.C:
			if !add_new {
				s.Clear()
				atodos, _ := get_todos()
				render_todos(s, atodos)
				s.Sync()
				s.Show()
			}
		}

	}


}
