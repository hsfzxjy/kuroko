package bluez

import (
	"github.com/godbus/dbus/v5"
	"github.com/muka/go-bluetooth/bluez/profile"
)

type bluezError string

var (
	// NotReady map to org.bluez.Error.NotReady
	ErrNotReady = bluezError(profile.ErrNotReady.Name)
	// InvalidArguments map to org.bluez.Error.InvalidArguments
	ErrInvalidArguments = bluezError(profile.ErrInvalidArguments.Name)
	// Failed map to org.bluez.Error.Failed
	ErrFailed = bluezError(profile.ErrFailed.Name)
	// DoesNotExist map to org.bluez.Error.DoesNotExist
	ErrDoesNotExist = bluezError(profile.ErrDoesNotExist.Name)
	// DoesNotExist map to org.bluez.Error.AlreadyExists
	ErrAlreadyExists = bluezError("org.bluez.Error.AlreadyExists")
	// Rejected map to org.bluez.Error.Rejected
	ErrRejected = bluezError(profile.ErrRejected.Name)
	// NotConnected map to org.bluez.Error.NotConnected
	ErrNotConnected = bluezError(profile.ErrNotConnected.Name)
	// NotAcquired map to org.bluez.Error.NotAcquired
	ErrNotAcquired = bluezError(profile.ErrNotAcquired.Name)
	// NotSupported map to org.bluez.Error.NotSupported
	ErrNotSupported = bluezError(profile.ErrNotSupported.Name)
	// NotAuthorized map to org.bluez.Error.NotAuthorized
	ErrNotAuthorized = bluezError(profile.ErrNotAuthorized.Name)
	// NotAvailable map to org.bluez.Error.NotAvailable
	ErrNotAvailable = bluezError(profile.ErrNotAvailable.Name)
	// AlreadyConnected map to org.bluez.Error.AlreadyConnected
	ErrAlreadyConnected = bluezError(profile.ErrAlreadyConnected.Name)
)

func (be bluezError) Is(err error) bool {
	if e, ok := err.(dbus.Error); ok {
		return e.Name == string(be)
	}
	return false
}
