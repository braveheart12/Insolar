package helper

import (
	"fmt"
	"strconv"
	"time"
)

func ParseTimeStamp(timeStr string) (time.Time, error) {

	i, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Unix(0, 0), fmt.Errorf("Failed to parse time ")
	}
	return time.Unix(i, 0), nil
}
