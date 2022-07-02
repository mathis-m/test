package spank_local

import (
	log "github.com/sirupsen/logrus"
	"unsafe"
)

func Exit(spank unsafe.Pointer) error {
	log.Info("spank_local.Exit: Nothing to do!")

	return nil
}
