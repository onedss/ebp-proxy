package buildtime

/*
const char* utils_build_time(void)
{
    static const char* psz_build_time = __DATE__ " " __TIME__;
    return psz_build_time;
}
*/
import "C"

import (
	"log"
	"time"
)

const (
	CBuildTimeLayout = "Jan  _2 2006 15:04:05"
)

func GetBuildTime() time.Time {
	value := C.GoString(C.utils_build_time())
	log.Println("BuildTime:", value)
	t, _ := time.ParseInLocation(CBuildTimeLayout, value, time.Local)
	return t
}
