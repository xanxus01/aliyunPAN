package aliyunDriver

import (
	"fmt"
	"runtime"
)

func log(v ...interface{})  {
	_, file, lineNo, _ := runtime.Caller(1)
	logString := fmt.Sprintf("%s:%d\n", file, lineNo)
	fmt.Println(logString)
	fmt.Println(v...)
}
