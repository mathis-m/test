#include <stdlib.h>
#include <stdint.h>
#include <slurm/spank.h>

extern void optionCallback(int value, char *optarg, int remote);



void slurm_info_wrapper(const char *string);
void slurm_error_wrapper(const char *string);
spank_err_t spank_get_item_uint32(spank_t spank, spank_item_t item, uid_t *result);

void optionCallback_cgo(int value, char *optarg, int remote);


