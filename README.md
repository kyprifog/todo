
# Todo

![Screenshot](https://github.com/kyprifog/todo/blob/master/images/screenshot.png)

Todo is a minimal todo app with click interaction provided by ![tcell](https://github.com/gdamore/tcell)

# Usage

```
go build
touch ~/.todos.yaml
./todo
```
Press `Add +` to add a new item, type your new item and press enter to add it.

To toggle an item as done simply click it.  Press `(-)` while checked done to permanently 
delete it.

Press Escape twice to exit

# Special keywords
`Shopping:`, `Goal:`, and `*` (corresponding to an important item) get formatting specially and the ordering is determined by this, Goals going first, then important items, then regular todos, then shopping items, and finally already finished items.
