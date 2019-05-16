package db

import (
    "errors"
    "fmt"
    "github.com/jmoiron/sqlx"
    "github.com/joyant/carry/entry"
    "github.com/joyant/carry/ini"
    "github.com/joyant/carry/screen"
    "github.com/joyant/carry/tool"
    v1 "gopkg.in/ini.v1"
    "os"
    "os/exec"
    "strconv"
    "time"
)

const (
    pageSize = 1000
    connectTimeout = 5
)

func Trans(from entry.Entry, to entry.Entry) error {
    err := from.SelfExamine()
    if err != nil {
        return err
    }

    err = to.SelfExamine()
    if err != nil {
        return err
    }

    fromSec, _ := ini.Ini().GetSection(from.Section)
    toSec, _ := ini.Ini().GetSection(to.Section)

    fromDB,err := Connect(fromSec)
    if err != nil {
        return fmt.Errorf("connect from db %s", err.Error())
    }

    toDB,err := Connect(toSec)
    if err != nil {
        return fmt.Errorf("connect to db %s", err.Error())
    }

    if from.Table == "" && to.Table == ""{
        return Database2database(from, to)
    } else if from.Table != "" && to.Table != ""{
        if from.IsSql {
            return Sql2table(from, to, fromDB, toDB)
        } else {
            return Table2table(from, to, fromDB, toDB)
        }
    } else {
        return errors.New("less table, please indicate a table")
    }
}

func mysqlExist() bool {
    _,err := exec.LookPath("mysql")
    return err == nil
}

func Database2database(from entry.Entry, to entry.Entry) error {
    sec,err := ini.Ini().GetSection(from.Section)
    if err != nil {
        return err
    }

    fromDB,err := connect(sec.Key("host").String(),
        sec.Key("user").String(),
        sec.Key("password").String(),
        from.Database)
    if err != nil {
        return err
    }

    if err := UseDatabase(fromDB, from.Database);err != nil {
        return err
    }

    sec2, err := ini.Ini().GetSection(to.Section)
    if err != nil {
        return err
    }

    toDB, err := connect(sec2.Key("host").String(),
        sec2.Key("user").String(),
        sec2.Key("password").String(),
        to.Database)
    if err != nil {
        return err
    }

    if err := UseDatabase(toDB, to.Database);err != nil {
        return err
    }

    tables, err := GetTables(fromDB)
    if err != nil {
        return err
    }

    for _, table := range tables {
        needCreate := false

        if TableExist(toDB, table) {
            yes,err := screen.Confirm("table " + table + " has exist, drop")
            if err != nil {
                return err
            }
            if yes {
                if err := dropTable(toDB, table);err != nil {
                    return err
                }
                needCreate = true
            }
        } else {
            needCreate = true
        }

        if needCreate {
            err := CreateTable(fromDB, toDB, table)
            if err != nil {
                return err
            }
        }

        err := CopyData(fromDB, table, false, toDB, table, func(percent int) {
            tool.Progress(percent)
        })
        if err != nil {
            return err
        } else {
            fmt.Printf("\rtable %s done\n",table)
        }
    }

    return nil
}

func getKeys(results []Result) []string {
    if len(results) == 0 {
        return nil
    } else {
        var keys []string
        for k := range results[0] {
            keys = append(keys, k)
        }
        return keys
    }
}

func CopyData(fromDB *sqlx.DB, fromTable string, isSql bool, toDB *sqlx.DB, toTable string, f func(int)) error {
    if isSql {
        fromTable = fmt.Sprintf("(%s) ss", fromTable)
    }

    count, err := GetMap(fromDB, "select count(*) count from " + fromTable)
    if err != nil {
        return err
    }
    total,err := strconv.ParseInt(count["count"].(string), 10, 32)
    if err != nil {
        return err
    }
    if total == 0 {
        return nil
    }

    toTx,err := toDB.Beginx()
    if err != nil {
        return err
    }

    page := 1

    for (page-1) * pageSize < int(total) {
        results,err := SelectMap(fromDB,
            fmt.Sprintf("select * from %s limit %d, %d", fromTable, (page - 1) * pageSize, pageSize))

        if err != nil {
            return err
        }
        if len(results) > 0 {
            keys := getKeys(results)
            affectRow, err := BatchExecTx(toTx, toTable, keys, results)
            if err != nil {
                toTx.Rollback()
                return err
            }
            if int(affectRow) != len(results) {
                toTx.Rollback()
                return fmt.Errorf("affect rows:%d, expect:%d", affectRow, len(results))
            }
        }

        percent := page * pageSize * 100 / int(total)
        if percent > 100 {
            percent = 100
        }
        f(percent)

        page ++
    }

    err = toTx.Commit()
    if err != nil {
        return errors.New("tx commit error:" + err.Error())
    }

    return nil
}

func UseDatabase(db *sqlx.DB, database string) error {
    _,err := db.Exec("use " + database)
    return err
}

func Table2table(from entry.Entry, to entry.Entry, fromDB, toDB *sqlx.DB) error {
    if err := UseDatabase(fromDB, from.Database);err != nil {
        return err
    }
    if err := UseDatabase(toDB, to.Database);err != nil {
        return err
    }

    needCreate := false

    if TableExist(toDB, to.Table) {
        //confirm if drop the exist table
        yes,err := screen.Confirm("table " + to.Table + " has exist, drop")
        if err != nil {
            return err
        }
        if yes {
            if err := dropTable(toDB, to.Table);err != nil {
                return err
            }
            needCreate = true
        }
    } else {
        needCreate = true
    }

    if needCreate {
        if from.Table != to.Table {
            return errors.New("from table must same as to table")
        }
        err := CreateTable(fromDB, toDB, to.Table)
        if err != nil {
            return err
        }
    }

    err := CopyData(fromDB, from.Table, false, toDB, to.Table, func(percent int) {
        tool.Progress(percent)
    })
    if err == nil {
        fmt.Printf("\rtable %s done\n", to.Table)
    }

    return err
}

func GetTableCreateStatement(db *sqlx.DB, table string) (string, error) {
    m,err := GetMap(db, "show create table " + table)
    if err != nil {
        return "", err
    }
    return m["Create Table"].(string), nil
}

func CreateTable(fromDB *sqlx.DB, toDB *sqlx.DB, table string) error {
    create, err := GetTableCreateStatement(fromDB, table)
    if err != nil {
        return err
    }
    _,err = toDB.Exec(create)
    return err
}

func Sql2table(from entry.Entry, to entry.Entry, fromDB, toDB *sqlx.DB) error {
    if err := UseDatabase(fromDB, from.Database);err != nil {
        return err
    }
    if err := UseDatabase(toDB, to.Database);err != nil {
        return err
    }

    if ! TableExist(toDB, to.Table) {
        //can not create table cause unknown table structure
        return errors.New("table " + to.Table + " not exist")
    }

    results,err := SelectMap(fromDB, from.Table)
    if err != nil {
        return err
    }

    if len(results) > 0 {
        err := CopyData(fromDB, from.Table, true, toDB, to.Table, func(percent int) {
            tool.Progress(percent)
        })
        if err != nil {
            return err
        }
    }

    fmt.Printf("\rtable %s done\n", to.Table)

    return nil
}

func Connect(section *v1.Section) (*sqlx.DB, error) {
    screen.Spin()

    d, err := connect(section.Key("host").String(),
        section.Key("user").String(),
        section.Key("password").String(),
        section.Key("database").String())

    screen.ClearSpin()

    return d, err
}

func connect(host, user, password string, args ...string) (*sqlx.DB, error) {
    dbChan := make(chan *sqlx.DB)
    errChan := make(chan error)

    database := ""
    if len(args) > 0 {
        database = args[0]
    }

    source := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, database)

    go func() {
        db,err := sqlx.Connect("mysql", source)
        if err != nil {
            errChan <- err
        } else {
            dbChan <- db
        }
    }()

    select {
    case e := <- errChan:
        return nil, e
    case d := <- dbChan:
        return d, nil
    case <- time.After(time.Second * connectTimeout):
        return nil, errors.New("connect " + host + " time out")
    }
}

func Login(section string) error {
    sec, err := ini.Ini().GetSection(section)
    if err != nil {
        return err
    }

    host := sec.Key("host").String()
    user := sec.Key("user").String()
    password := sec.Key("password").String()
    port := sec.Key("port").String()

    if mysqlExist() {
        //try to connect database before execute shell, Connect will handle timeout problem
        if _, err := Connect(sec);err != nil {
            return err
        }
        cmd := exec.Command("mysql","-h"+host,"-u"+user,"-p"+password,"-P"+port)
        cmd.Stdin = os.Stdin
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        err := cmd.Run()
        return err
    } else {
        return errors.New("please install mysql-client first")
    }
}

func Drop(entry entry.Entry) (error) {
    if entry.Section == "" {
        return errors.New("section can not be empty")
    }
    sec,err := ini.Ini().GetSection(entry.Section)
    if err != nil {
        return err
    }

    fromDB, err := connect(sec.Key("host").String(),
        sec.Key("user").String(),
        sec.Key("password").String())
    if err != nil {
        return err
    }

    if entry.Table != "" {
        if err := UseDatabase(fromDB, entry.Database);err != nil {
            return err
        }
        err = dropTable(fromDB, entry.Table)
    } else {
        err = dropDatabase(fromDB, entry.Database)
    }

    if err == nil {
        if entry.Table == "" {
            fmt.Println("drop database " + entry.Database + " done")
        } else {
            fmt.Println("drop table " + entry.Table + " done")
        }
    }

    return err
}

func dropTable(db *sqlx.DB, table string) error {
    _,err := db.Exec("drop table " + table)
    return err
}

func dropDatabase(db *sqlx.DB, database string) error {
    _,err := db.Exec("drop database " + database)
    return err
}
