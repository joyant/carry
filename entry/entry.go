package entry

import (
    "errors"
    "github.com/joyant/carry/ini"
    "strings"
)

type Entry struct {
    Section  string
    Database string
    Table    string
    IsSql    bool
}

func (e *Entry) Parse(str string) error {
    slice := strings.Split(str, ".")
    if len(slice) < 2 {
        return errors.New("not indicate database")
    }

    e.Section = slice[0]
    e.Database = slice[1]

    if len(slice) > 2 {
        if slice[2] == "sql" {
            e.IsSql = true
        } else {
            e.Table = slice[2]
        }
    }

    return nil
}

func (e *Entry) SelfExamine() error {
    sec, err := ini.Ini().GetSection(e.Section)
    if err != nil {
        return err
    }
    if sec.Key("user").String() == "" {
        return errors.New("user can not be empty")
    }
    if sec.Key("host").String() == "" {
        return errors.New("host can not be empty")
    }
    if sec.Key("password").String() == "" {
        return errors.New("password can not be empty")
    }
    return nil
}