package ini

import (
    "fmt"
    "gopkg.in/ini.v1"
    "os"
    "runtime"
)

var iniPath = "/usr/local/etc/carry.ini"

var iniFile *ini.File

func Ini() *ini.File {
    return iniFile
}

func init()  {
    if runtime.GOOS == "windows" {
        iniPath = "C:/carry.ini"
    }
    f, err := os.OpenFile(iniPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 766)
    if err != nil {
        fmt.Println("err:", err.Error(), "please make sure dir /usr/local/etc exist, and has read-write permission")
        os.Exit(0)
    }

    iniFile, err = ini.Load(f)
    if err != nil {
        fmt.Println("load ini err:", err.Error())
        os.Exit(0)
    }
}

func Save() error {
    return iniFile.SaveTo(iniPath)
}
