package utils

import "testing"


func TestLogger(t *testing.T) {
    handles, err := GetHandles("info", "/etc/yascrapy/core.log")
    if err != nil {
        t.Error(err.Error())
    }
    InitLoggers(handles[0], handles[1], handles[2], handles[3])
    Debug.Println("ok")
    Info.Println("ok")
    Warning.Println("ok")
    Error.Println("ok")
}
