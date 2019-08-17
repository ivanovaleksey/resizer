package cache

type Entity string

func (e Entity) Key() string {
	return string(e)
}
