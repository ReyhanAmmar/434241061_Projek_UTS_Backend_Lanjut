package helper

import "time"

func timeFromMicro(micro int64) time.Time {
	return time.UnixMicro(micro).UTC()
}