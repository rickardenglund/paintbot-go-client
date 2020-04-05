package timeHelper

import "time"

func Now() int {
	return int(time.Now().UnixNano() / int64(time.Millisecond))
}
