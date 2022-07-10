package main

/*

#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <slurm/spank.h>

void slurm_info_wrapper(const char *string);

*/
import "C"

import (
	"fmt"
	"github.com/s-bauer/slurm-k8s/internal/slurm"
	"github.com/s-bauer/slurm-k8s/internal/spank_local"
	"github.com/s-bauer/slurm-k8s/internal/spank_remote"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"unsafe"
)

func init() {
	log.SetOutput(io.Discard)
	log.AddHook(&slurm.LogHook{})
	log.SetFormatter(&log.TextFormatter{DisableColors: true, DisableQuote: true})
	log.SetLevel(log.DebugLevel)
}

func main() {}

//export slurm_spank_init
func slurm_spank_init(spank C.spank_t, ac C.int, av **C.char) C.int {
	log.Info("slurm_spank_init start")

	spankUnsafe := unsafe.Pointer(&spank)

	_ = slurm.RegisterOption(spankUnsafe, &slurm.Option{
		Name:    "k8s-init-cluster",
		ArgInfo: "",
		Usage:   "Initializes a new k8s cluster",
		Value:   0,
		HasArg:  0,
		Callback: func(val int, optArg string, remote int) {
			viper.Set("k8s-init-cluster", true)
			log.Info("OPTION: k8s-init-cluster set")
		},
	})

	_ = slurm.RegisterOption(spankUnsafe, &slurm.Option{
		Name:    "k8s-join-cluster",
		ArgInfo: "",
		Usage:   "Joins an existing k8s cluster setup with --k8s-init-cluster",
		Value:   1,
		HasArg:  0,
		Callback: func(val int, optArg string, remote int) {
			viper.Set("k8s-join-cluster", true)
			log.Info("OPTION: k8s-join-cluster set")
		},
	})

	return C.ESPANK_SUCCESS
}

//export slurm_spank_init_post_opt
func slurm_spank_init_post_opt(spank C.spank_t, ac C.int, av **C.char) C.int {
	log.Info("slurm_spank_init_post_opt start")

	switch ctx := C.spank_context(); ctx {

	case C.S_CTX_REMOTE:
		if err := spank_remote.Init(unsafe.Pointer(&spank)); err != nil {
			log.Error("spank_remote.Init:", err)
			return C.ESPANK_SUCCESS // TODO: Change to error
		}
		return C.ESPANK_SUCCESS

	case C.S_CTX_LOCAL:
		if err := spank_local.Init(unsafe.Pointer(&spank)); err != nil {
			log.Error("spank_local.Init:", err)
			return C.ESPANK_SUCCESS // TODO: Change to error
		}
		return C.ESPANK_SUCCESS

	default:
		log.Error(fmt.Sprintf("Unsupported Context: %v", uint32(ctx)))
		return C.ESPANK_SUCCESS

	}
}

// export slurm_spank_task_init

//export slurm_spank_user_init
func slurm_spank_user_init(spank C.spank_t, ac C.int, av **C.char) C.int {
	log.Info("slurm_spank_user_init start")

	switch ctx := C.spank_context(); ctx {
	case C.S_CTX_REMOTE:
		if err := spank_remote.UserInit(unsafe.Pointer(&spank)); err != nil {
			log.Error("spank_remote.UserInit: ", err)
			return C.ESPANK_SUCCESS // TODO: Change to error
		}
		return C.ESPANK_SUCCESS

	default:
		log.Error(fmt.Sprintf("Unsupported Context: %v", uint32(ctx)))
		return C.ESPANK_SUCCESS
	}
}

//export slurm_spank_exit
func slurm_spank_exit(spank C.spank_t, ac C.int, av **C.char) C.int {
	log.Info("slurm_spank_exit start")

	switch ctx := C.spank_context(); ctx {
	case C.S_CTX_REMOTE:
		if err := spank_remote.Exit(unsafe.Pointer(&spank)); err != nil {
			log.Error("spank_remote.Init:", err)
			return C.ESPANK_SUCCESS // TODO: Change to error
		}
		return C.ESPANK_SUCCESS

	case C.S_CTX_LOCAL:
		if err := spank_local.Exit(unsafe.Pointer(&spank)); err != nil {
			log.Error("spank_local.Init:", err)
			return C.ESPANK_SUCCESS // TODO: Change to error
		}
		return C.ESPANK_SUCCESS

	default:
		log.Error(fmt.Sprintf("Unsupported Context: %v", uint32(ctx)))
		return C.ESPANK_SUCCESS
	}

}
