package resizer

import "go.uber.org/zap"

type ServiceOption func(*Service)

func WithImageProvider(provider ImageProvider) ServiceOption {
	return func(service *Service) {
		service.imageProvider = provider
	}
}

func WithImageResizer(resizer ImageResizer) ServiceOption {
	return func(service *Service) {
		service.imageResizer = resizer
	}
}

func WithLogger(logger *zap.Logger) ServiceOption {
	return func(service *Service) {
		service.logger = logger
	}
}
