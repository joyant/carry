package main

import (
    "errors"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    my "github.com/joyant/carry/db"
    "github.com/joyant/carry/entry"
    "github.com/joyant/carry/ini"
    "github.com/joyant/carry/screen"
    "github.com/urfave/cli"
    "os"
)

func main() {
    parseCommand()
}

func parseCommand()  {
    app := cli.NewApp()
    app.Name = "carry"
    app.Usage = "carry mysql data"
    
    app.Commands = []cli.Command{
        //command store
        cli.Command{
            Name:"store",
            Usage:"store section",
            UsageText:"carry store -s dev -H localhost -u root -p root -P 3306",
            Description:"section is a unit of connect",
            Action: func(c *cli.Context) error {
                section := c.String("section")
                if len(section) == 0 {
                    return errors.New("section can not be empty")
                }

                sec := ini.Ini().Section(section)

                host := c.String("host")
                user := c.String("user")
                password := c.String("password")
                port := c.String("port")

                if host != "" {
                    sec.NewKey("host", host)
                }
                if user != "" {
                    sec.NewKey("user", user)
                }
                if password != "" {
                    sec.NewKey("password", password)
                }
                if port != "" {
                    sec.NewKey("port", port)
                }

                fmt.Println("save success")

                return ini.Save()
            },
            Flags:[]cli.Flag{
                cli.StringFlag{
                    Name:"section, s",
                    Usage:"mysql section",
                },
                cli.StringFlag{
                    Name:"host, H",
                    Value:"localhost",
                    Usage:"mysql host",
                },
                cli.StringFlag{
                    Name:"user, u",
                    Usage:"mysql user",
                },
                cli.StringFlag{
                    Name:"password, p",
                    Usage:"mysql password",
                },
                cli.StringFlag{
                    Name:"port, P",
                    Value:"3306",
                    Usage:"mysql port",
                },
            },
        },
        //command trans
        cli.Command{
           Name:"trans",
           Usage:"transport data between database or table",
           UsageText:"carry trans -f section.database[.table, .sql] -to section.database[.table]",
           Description:"besides transport between database table, also support transport data from a sql statement to a table ",
           Action: func(c *cli.Context) error {
               from := c.String("from")
               to := c.String("to")

               if from == "" {
                   return errors.New("please input from section")
               }
               if to == "" {
                   return errors.New("please input to section")
               }

               fromEntry := entry.Entry{}
               err := fromEntry.Parse(from)
               if err != nil {
                   return err
               }

               toEntry := entry.Entry{}
               err = toEntry.Parse(to)
               if err != nil {
                   return err
               }

               if fromEntry.IsSql {
                   sql,err := screen.Ask("please input your sql: ")
                   if err != nil {
                       return err
                   }
                   if sql == "" {
                       return errors.New("sql can not be empty")
                   }
                   fromEntry.Table = sql
               }

               return my.Trans(fromEntry, toEntry)
           },
           Flags:[]cli.Flag{
               cli.StringFlag{
                   Name:"from, f",
                   Usage:"from section",
               },
               cli.StringFlag{
                   Name:"to, t",
                   Usage:"to section",
               },
           },
        },
        //command del
        cli.Command{
            Name:"del",
            Usage:"del one section",
            UsageText:"carry del -s dev",
            Action: func(c *cli.Context) error {
                section := c.String("section")
                if section == "" {
                    return errors.New("please input section")
                }
                if _,err := ini.Ini().GetSection(section);err != nil {
                    return err
                }
                ini.Ini().DeleteSection(section)
                err := ini.Save()
                if err == nil {
                    fmt.Println("delete success")
                }
                return err
            },
            Flags:[]cli.Flag{
                cli.StringFlag{
                    Name:"section, s",
                    Usage:"section name",
                },
            },
        },
        //command login
        cli.Command{
            Name:"login",
            Usage:"login section you stored",
            UsageText:"carry login -s dev",
            Description:"will inter into mysql interface after login success",
            Action: func(c *cli.Context) error {
                section := c.String("section")
                if section == "" {
                    return errors.New("please input section")
                }
                return my.Login(section)
            },
            Flags:[]cli.Flag{
                cli.StringFlag{
                    Name:"section, s",
                    Usage:"section",
                },
            },
        },
        //command list
        cli.Command{
            Name:"list",
            Usage:"show section list you stored",
            UsageText:"carry list | carry list -s dev",
            Description:"command carry list will show all sections, carry list -s dev will show only dev section",
            Action: func(c *cli.Context) error {
                section := c.String("section")
                if section == "" {
                    secs := ini.Ini().Sections()
                    for _, sec := range secs {
                        if sec.Name() != "DEFAULT" {
                            fmt.Println("section:",sec.Name())
                            keys := sec.Keys()
                            for _,k := range keys {
                                fmt.Println(k.Name() + "=" + k.Value())
                            }
                        }
                    }
                } else {
                    sec,err := ini.Ini().GetSection(section)
                    if err != nil {
                        return err
                    }
                    keys := sec.KeyStrings()
                    for _, k := range keys {
                        v := sec.Key(k).String()
                        fmt.Println(k + "=" + v)
                    }
                }
                return nil
            },
            Flags:[]cli.Flag{
                cli.StringFlag{
                    Name:"section, s",
                    Usage:"show section",
                },
            },
        },
        cli.Command{
            Name:"drop",
            Usage:"drop table or database",
            UsageText:"carry drop -s section.database[.table]",
            Description:"section.database means drop database, section.database.table means drop table",
            Action: func(c *cli.Context) error {
                section := c.String("section")
                fromEntry := entry.Entry{}
                if err := fromEntry.Parse(section);err != nil {
                    return err
                }
                return my.Drop(fromEntry)
            },
            Flags:[]cli.Flag{
                cli.StringFlag{
                    Name:"section, s",
                    Usage:"drop section",
                },
            },
        },
    }
    
    err := app.Run(os.Args)
    if err != nil {
        fmt.Println("err:", err)
        os.Exit(0)
    }
}
