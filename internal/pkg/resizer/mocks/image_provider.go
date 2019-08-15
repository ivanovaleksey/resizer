// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import image "image"
import mock "github.com/stretchr/testify/mock"

// ImageProvider is an autogenerated mock type for the ImageProvider type
type ImageProvider struct {
	mock.Mock
}

// GetImage provides a mock function with given fields: ctx, target
func (_m *ImageProvider) GetImage(ctx context.Context, target string) (image.Image, error) {
	ret := _m.Called(ctx, target)

	var r0 image.Image
	if rf, ok := ret.Get(0).(func(context.Context, string) image.Image); ok {
		r0 = rf(ctx, target)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(image.Image)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, target)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}