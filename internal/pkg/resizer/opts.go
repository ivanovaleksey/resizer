package resizer

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
