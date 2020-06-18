package drivers

import (
	"context"
	"sync"
)

// NewDeviceFunc is passed to NewServer and is called to create a new Device struct whenever the Server needs to control with a new Device.
type NewDeviceFunc func(context.Context, string) (Device, error)

func NewServer(newDev NewDeviceFunc) (Server, error) {
	newDev = saveDevicesFunc(newDev)
	server := newGrpcServer(newDev)
	return server, nil
}

func saveDevicesFunc(newDev NewDeviceFunc) NewDeviceFunc {
	m := &sync.Map{}

	return func(ctx context.Context, addr string) (Device, error) {
		if dev, ok := m.Load(addr); ok {
			return dev, nil
		}

		dev, err := newDev(ctx, addr)
		if err != nil {
			return dev, err
		}

		m.Store(addr, dev)
		return dev, nil
	}
}
