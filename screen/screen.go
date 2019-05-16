package screen

import (
    "bufio"
    "fmt"
    "os"
    "time"
)

var spinStop = make(chan bool)

func Spin()  {
    go func() {
        for {
            select {
            case <- spinStop:
                fmt.Printf("\r")
                return
            default:
                for _, s := range "-\\|/" {
                    fmt.Printf("\r%c",s)
                    time.Sleep(time.Millisecond * 50)
                }
            }
        }
    }()
}

func ClearSpin()  {
    spinStop <- true
}

func Confirm(question string) (bool, error) {
    fmt.Print(question + "? [Y/N] ")
    reader := bufio.NewReader(os.Stdin)
    b, err := reader.ReadString('\n')
    if err != nil {
        return false, err
    }
    if b[0] == 'Y' || b[0] == 'y' {
        return true, nil
    } else {
        return false, nil
    }
}

func Ask(question string) (answer string, err error) {
    fmt.Print(question)
    reader := bufio.NewReader(os.Stdin)
    answer, err = reader.ReadString('\n')
    return
}

