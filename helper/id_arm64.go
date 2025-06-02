package helper

import (
	"fmt"
	"strconv"

	"lukechampine.com/frand"
)

/*
#include <mach/mach_time.h>
C.mach_absolute_time()
*/

func Now(physical bool) uint64

func Cputicks() uint64 {

	val := fmt.Sprintf("%d%d", Now(true), 100+frand.Intn(900))

	t, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}
	return t
}
