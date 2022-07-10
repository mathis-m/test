package slurm

/*
#include <stdlib.h>
#include <stdint.h>
#include <slurm/spank.h>
#include <types.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type Option struct {
	Name    string
	ArgInfo string
	Usage   string

	// Value a unique id for this spank plugin to identify the parameter
	Value int

	// HasArg
	// - 0 if no arguments
	// - 1 if argument is required
	// - 2 if argument is optional
	HasArg int

	Callback func(val int, optArg string, remote int)
}

func WriteInfo(msg string) {
	cStr := C.CString(msg)
	defer C.free(unsafe.Pointer(cStr))

	C.slurm_info_wrapper(cStr)
}

func WriteError(msg string) {
	cStr := C.CString(msg)
	defer C.free(unsafe.Pointer(cStr))

	C.slurm_info_wrapper(cStr)
}

func GetSlurmJobUid(spank C.spank_t) (uint32, error) {
	var userIdC C.uint32_t

	statusC := C.spank_get_item_uint32(spank, C.S_JOB_UID, &userIdC)
	status := int(statusC)
	if status > 0 {
		return 0, fmt.Errorf("spank_get_item failed with code %d", status)
	}

	return uint32(userIdC), nil
}

func GetSlurmEnvVar(spank unsafe.Pointer, name string) (string, error) {
	nameC := C.CString(name)
	defer C.free(unsafe.Pointer(nameC))

	const bufSize = 16 * 1024
	var buf [bufSize]C.char

	statusC := C.spank_getenv(*(*C.spank_t)(spank), nameC, &buf[0], bufSize)
	status := int(statusC)
	if status > 0 {
		return "", fmt.Errorf("spank_getenv failed with code %d", status)
	}

	return C.GoString(&buf[0]), nil
}

var optionCallbacks = make(map[int]func(val int, optArg string, remote int))

func RegisterOption(spank unsafe.Pointer, option *Option) error {
	spankTyped := *(*C.spank_t)(spank)

	cName := C.CString(option.Name)
	defer C.free(unsafe.Pointer(cName))

	cUsage := C.CString(option.Usage)
	defer C.free(unsafe.Pointer(cUsage))

	cArgInfo := C.CString(option.ArgInfo)
	defer C.free(unsafe.Pointer(cArgInfo))

	cOption := &C.struct_spank_option{
		name:    cName,
		arginfo: cArgInfo,
		usage:   cUsage,
		has_arg: C.int(option.HasArg),
		val:     C.int(option.Value),
		cb:      C.spank_opt_cb_f(C.optionCallback_cgo),
	}

	optionCallbacks[option.Value] = option.Callback

	result := C.spank_option_register(spankTyped, cOption)
	switch result {
	case C.ESPANK_SUCCESS:
		return nil
	default:
		return fmt.Errorf("unable to add slurm option")
	}
}

//export optionCallback
func optionCallback(value C.int, optarg *C.char, remote C.int) {
	goOptArg := string(C.GoString(optarg))
	goValue := int(value)
	goRemote := int(remote)

	if callback, ok := optionCallbacks[goValue]; ok {
		callback(goValue, goOptArg, goRemote)
	}
}
