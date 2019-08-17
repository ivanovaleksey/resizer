package cache

const ErrCacheMiss = Error("cache miss")

type Error string

func (e Error) Error() string {
	return string(e)
}
