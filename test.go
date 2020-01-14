package zlsgo

import (
	"path"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestUtil Test aid
type TestUtil struct {
	T *testing.T
}

// NewTest testing object
func NewTest(t *testing.T) *TestUtil {
	return &TestUtil{t}
}

// GetCallerInfo GetCallerInfo
func (u *TestUtil) GetCallerInfo() string {
	var info string

	for i := 0; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		basename := path.Base(file)
		if !strings.HasSuffix(basename, "_test.go") {
			continue
		}

		funcName := runtime.FuncForPC(pc).Name()
		index := strings.LastIndex(funcName, ".Test")
		if -1 == index {
			index = strings.LastIndex(funcName, ".Benchmark")
			if index == -1 {
				continue
			}
		}
		funcName = funcName[index+1:]

		if index := strings.IndexByte(funcName, '.'); index > -1 {
			funcName = funcName[:index]
			info = funcName + "(" + basename + ":" + strconv.Itoa(line) + ")"
			continue
		}

		info = funcName + "(" + basename + ":" + strconv.Itoa(line) + ")"
		break
	}

	if info == "" {
		info = "<Unable to get information>"
	}
	return info
}

// Equal Equal
func (u *TestUtil) Equal(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		u.T.Errorf("%s 期待:%v (type %v) - 结果:%v (type %v)", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

// EqualExit EqualExit
func (u *TestUtil) EqualExit(expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		u.T.Fatalf("%s 期待:%v (type %v) - 结果:%v (type %v)", u.PrintMyName(), expected, reflect.TypeOf(expected), actual, reflect.TypeOf(actual))
	}
}

// Log log
func (u *TestUtil) Log(v ...interface{}) {
	tip := []interface{}{"\n  " + u.PrintMyName()}
	va := append(tip, v...)
	u.T.Log(va...)
}

// Fatal Fatal
func (u *TestUtil) Fatal(v ...interface{}) {
	tip := []interface{}{"\n  " + u.PrintMyName()}
	va := append(tip, v...)
	u.T.Fatal(va...)
}

// PrintMyName PrintMyName
func (u *TestUtil) PrintMyName() string {
	return "@" + u.GetCallerInfo()
}
