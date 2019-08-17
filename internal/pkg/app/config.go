package app

type ImageProviderType int

const (
	ImageProviderHTTP ImageProviderType = iota + 1
	ImageProviderFile
)

type Config struct {
	ImageProvider ImageProviderType // 1 - http, 2 - file
}
