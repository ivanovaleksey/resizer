package cache

import "strconv"

type Entity struct {
	URL    string
	Width  int
	Height int
}

func (e Entity) Key() string {
	return e.URL + "-" + strconv.Itoa(e.Width) + "x" + strconv.Itoa(e.Height)
}
