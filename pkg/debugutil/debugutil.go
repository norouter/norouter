package debugutil

import (
	"io/ioutil"
	"runtime"
)

func NumFDs() int {
	if runtime.GOOS != "linux" {
		// unimplemented
		return -1
	}
	ents, err := ioutil.ReadDir("/proc/self/fd")
	if err != nil {
		return -1
	}
	return len(ents)
}
