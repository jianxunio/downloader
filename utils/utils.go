package utils

import "os"
import "log"
import "io"
import "io/ioutil"
import "errors"

var (
    Debug *log.Logger
    Info *log.Logger
    Warning *log.Logger
    Error *log.Logger
)


func GetHandles(logLevel string) ([]io.Writer, error) {
    handles := make([]io.Writer, 0)
    /*
    f, err := os.Create(logFile)
    if err != nil {
        return handles, err
    }*/
    levels := []string{"debug", "info", "warning", "error"}
    foundIndex := -1
    for i, level := range levels {
        if level == logLevel {
            foundIndex = i
            break
        }
    }
    if foundIndex < 0 {
        return handles, errors.New("logLevel error")
    } else {
        for i, _ := range levels {
            if i < foundIndex {
                handles = append(handles, ioutil.Discard)
            } else {
                handles = append(handles, os.Stdout)
            }
        }
    }
    return handles, nil
}

func InitLoggers(debugHandle io.Writer, infoHandle io.Writer, warnHandle io.Writer, errorHandle io.Writer) {
    Debug = log.New(debugHandle, "[debug]", log.Ldate|log.Ltime|log.Lshortfile)
    Info = log.New(infoHandle, "[info]", log.Ldate|log.Ltime|log.Lshortfile)
    Warning = log.New(warnHandle, "[warning]", log.Ldate|log.Ltime|log.Lshortfile)
    Error = log.New(errorHandle, "[error]", log.Ldate|log.Ltime|log.Lshortfile)
}
