package lua

import (
	"fmt"
	"runtime"
)

func printStackTrace() {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	fmt.Printf("Stack trace:\n%s\n", buf[:n])
}
