package slurm

/*

#include <stdlib.h>
#include <stdint.h>
#include <slurm/spank.h>


*/
import "C"

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/user"
	"unsafe"
)

func GetJobUser(spank unsafe.Pointer) (*user.User, error) {
	userId, err := GetSlurmJobUid(*(*C.spank_t)(spank))
	if err != nil {
		return nil, errors.New(fmt.Sprint("Unable to get S_JOB_UID: ", err))
	}

	jobUser, err := user.LookupId(fmt.Sprint(userId))
	if err != nil {
		return nil, err
	}

	log.Info("UserId is", userId, ", UserName:", jobUser.Username, ", HomeDir:", jobUser.HomeDir)
	return jobUser, nil
}
