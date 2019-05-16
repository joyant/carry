package tool

import "fmt"

func Progress(percent int)  {
    fmt.Printf("\r%d%%", percent)
}

func ClearProgress()  {
    fmt.Print("\r\r\r")
}
