package slurm

/*

#include <stdlib.h>
#include <stdint.h>
#include <slurm/spank.h>

typedef spank_t SpankT;

void slurm_info_wrapper(const char *string) {
    slurm_info("%s", string);
}

void slurm_error_wrapper(const char *string) {
    slurm_error("%s", string);
}

spank_err_t spank_get_item_uint32(spank_t spank, spank_item_t item, uid_t *result) {
	return spank_get_item(spank, item, result);
}

SPANK_PLUGIN(kubernetes, 1);

*/
import "C"

import (
	"fmt"
	"unsafe"
)

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

	const bufSize = 2048
	var buf [bufSize]C.char

	statusC := C.spank_getenv(*(*C.spank_t)(spank), nameC, &buf[0], bufSize)
	status := int(statusC)
	if status > 0 {
		return "", fmt.Errorf("spank_getenv failed with code %d", status)
	}

	return C.GoString(&buf[0]), nil
}
