#include "types.h"

void slurm_info_wrapper(const char *string) {
    slurm_info("%s", string);
}

void slurm_error_wrapper(const char *string) {
    slurm_error("%s", string);
}

spank_err_t spank_get_item_uint32(spank_t spank, spank_item_t item, uid_t *result) {
	return spank_get_item(spank, item, result);
}

void optionCallback_cgo(int value, char *optarg, int remote) {
    optionCallback(value, optarg, remote);
}