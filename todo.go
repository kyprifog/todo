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

func render_todos(s tcell.Screen, todos []map[string]interface{}) {
	green := tcell.StyleDefault.Foreground(tcell.ColorLawnGreen)
	grey := tcell.StyleDefault.Foreground(tcell.ColorGrey)
	orange := tcell.StyleDefault.Foreground(tcell.ColorOrange)
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)

	emitStr(s, 0,0, green, "TODO")

	index := 1

	for _, el := range todos {
		name := el["name"].(string)
		checked := el["done"].(bool)
		if checked == true {
			emitStr(s, 0, index, grey, "[x] " + name + " (-)")
		} else {
			emitStr(s, 0, index, blue, "[ ] " + name)
		}
		index += 1
	}
	emitStr(s, 0, index, orange, "Add +")
}

func add_new_todo(s tcell.Screen, new_todo string) {
	blue := tcell.StyleDefault.Foreground(tcell.ColorBlue)
	emitStr(s, 0, 2, blue, "New Todo: " + new_todo)
}

func tick_todos(x int, y int, todos []map[string]interface{}) []map[string]interface{} {
	new_todos := []map[string]interface{}{}
	done_todos := []map[string]interface{}{}
	undone_todos := []map[string]interface{}{}
	for i, el := range todos {
		checked := el["done"].(bool)
		name := el["name"].(string)
		if (1 + i) == y {
			if checked == true {
				if (len(name) + 6) != x {
					el["done"] = false
					undone_todos = append(undone_todos, el)
				}
			} else {
				el["done"] = true
				done_todos = append(done_todos, el)
			}
		} else {
			if checked == true {
				done_todos = append(done_todos, el)
			} else {
				undone_todos = append(undone_todos, el)
			}
		}
	}

	for _, el2 := range undone_todos {
		new_todos = append(new_todos, el2)
	}

	for _, el3 := range done_todos {
		new_todos = append(new_todos, el3)
	}

	save_todos(new_todos)
	return new_todos
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

	defStyle = tcell.StyleDefault.
		Background(tcell.ColorReset).
		Foreground(tcell.ColorReset)

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
				if y < (l + 1) {
					s.Clear()
					new_todos := tick_todos(x, y, atodos)
					render_todos(s, new_todos)
					s.Sync()
					s.Show()
				} else if y == l+1 {
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
