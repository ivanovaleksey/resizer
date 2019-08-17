package singleflight

import "go.uber.org/zap"

type Option func(*SingleFlight)

func WithImageProvider(provider ImageProvider) Option {
	return func(s *SingleFlight) {
		s.imageProvider = provider
	}
}

func WithCacheProvider(provider CacheProvider) Option {
	return func(s *SingleFlight) {
		s.cache = provider
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(s *SingleFlight) {
		s.logger = logger
	}
}
