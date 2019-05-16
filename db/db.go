package db

import (
    "bytes"
    "fmt"
    "github.com/jmoiron/sqlx"
    "strings"
)

type Result map[string]interface{}

func GetMap(db *sqlx.DB, query string, args ...interface{}) (Result, error) {
    row := db.QueryRowx(query, args...)
    if row.Err() != nil {
        return nil, row.Err()
    }
    m := make(map[string]interface{})
    row.MapScan(m)
    convert(m)
    return m, nil
}

func GetMapTx(db *sqlx.Tx, query string, args ...interface{}) (Result, error) {
    row := db.QueryRowx(query, args...)
    if row.Err() != nil {
        return nil, row.Err()
    }
    m := make(map[string]interface{})
    row.MapScan(m)
    convert(m)
    return m, nil
}

func convert(m map[string]interface{}) {
    for k, v := range m {
        if v == nil {
            continue
        }
        switch v.(type) {
        case []uint8:
            buff := bytes.Buffer{}
            for _, v := range v.([]uint8) {
                buff.WriteByte(byte(v))
            }
            m[k] = buff.String()
        }
    }
}

func SelectMap(db *sqlx.DB, query string, args ...interface{}) ([]Result, error) {
    var rows *sqlx.Rows
    var err error

    rows, err = db.Queryx(query, args...)

    if err != nil {
        return nil, err
    }
    var ret []Result
    for rows.Next() {
        one := make(map[string]interface{})
        rows.MapScan(one)
        convert(one)
        ret = append(ret, one)
    }

    return ret, nil
}

func SelectMapTx(db *sqlx.Tx, query string, args ...interface{}) ([]Result, error) {
    var rows *sqlx.Rows
    var err error

    rows, err = db.Queryx(query, args...)

    if err != nil {
        return nil, err
    }
    var ret []Result
    for rows.Next() {
        one := make(map[string]interface{})
        rows.MapScan(one)
        convert(one)
        ret = append(ret, one)
    }

    return ret, nil
}

func BatchExec(db *sqlx.DB, table string,keys []string,result []Result) (int64,error) {
    buff := bytes.Buffer{}
    buff.WriteString("insert into "+table+" (")
    for index,k := range keys {
        buff.WriteString(k)
        if index < len(keys) - 1 {
            buff.WriteString(",")
        }
    }
    buff.WriteString(")values")
    for oneIndex,one := range result {
        buff.WriteString("(")
        for index,k := range keys {
            if value,ok := one[k];ok{
                if value == nil {
                    buff.WriteString("null")
                }else{
                    buff.WriteString(fmt.Sprintf("'%v'",one[k]))
                }
            }else{
                buff.WriteString("null")
            }
            if index < len(keys) - 1 {
                buff.WriteString(",")
            }
        }
        buff.WriteString(")")
        if oneIndex < len(result) - 1 {
            buff.WriteString(",")
        }
    }

    statement := buff.String()
    affected,_,err := Exec(db, statement)
    if err != nil {
        return 0,err
    }

    return affected,nil
}

func BatchExecTx(db *sqlx.Tx, table string,keys []string,result []Result) (int64,error) {
    buff := bytes.Buffer{}
    buff.WriteString("insert into "+table+" (")
    for index,k := range keys {
        buff.WriteString(k)
        if index < len(keys) - 1 {
            buff.WriteString(",")
        }
    }
    buff.WriteString(")values")
    for oneIndex,one := range result {
        buff.WriteString("(")
        for index,k := range keys {
            if value,ok := one[k];ok{
                if value == nil {
                    buff.WriteString("null")
                }else{
                    buff.WriteString(fmt.Sprintf("'%v'",one[k]))
                }
            }else{
                buff.WriteString("null")
            }
            if index < len(keys) - 1 {
                buff.WriteString(",")
            }
        }
        buff.WriteString(")")
        if oneIndex < len(result) - 1 {
            buff.WriteString(",")
        }
    }

    statement := buff.String()
    affected,_,err := ExecTx(db, statement)
    if err != nil {
        return 0,err
    }

    return affected,nil
}

func ExecTx(db *sqlx.Tx, query string, args ...interface{}) (rowsAffected int64,lastInsertId int64,err error) {
    result,err := db.Exec(query, args...)
    if err != nil {
        return 0,0,err
    }

    a,err := result.RowsAffected()
    if err != nil {
        return 0,0,err
    }

    b,err := result.LastInsertId()
    if err != nil {
        return 0,0,err
    }

    return a,b,nil
}

func Exec(db *sqlx.DB, query string, args ...interface{}) (rowsAffected int64,lastInsertId int64,err error) {
    result,err := db.Exec(query, args...)
    if err != nil {
        return 0,0,err
    }

    a,err := result.RowsAffected()
    if err != nil {
        return 0,0,err
    }

    b,err := result.LastInsertId()
    if err != nil {
        return 0,0,err
    }

    return a,b,nil
}

func TableExist(db *sqlx.DB, table string) bool {
    _,err := GetMap(db, "desc " + table)
    if err != nil && !strings.Contains(err.Error(), "doesn't exist"){
        fmt.Println("func TableExist:", err)
    }
    return err == nil
}

func GetTables(db *sqlx.DB) ([]string, error) {
    results, err := SelectMap(db,"show tables")
    if err != nil {
        return nil, err
    }
    var tables []string
    for _, result := range results {
        for _,table := range result {
            tables = append(tables, table.(string))
        }
    }
    return tables, nil
}
