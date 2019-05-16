# Carry
[中文文档](https://github.com/joyant/carry/blob/master/Readme_zh.md)

Carry is a mysql command line tool, use it to transport data conveniently.

```bash
$ carry trans -f dev.database_name.table_name -t dev2.database_name.table_name
```

## Installation
You can compile by yourself.
```bash
go build main.go
```
Or you can download [here](https://github.com/joyant/carry/releases)


## Features
* transport data between different server.
* login mysql fast.
* drop table or database without login.

## Quick Start

Before you use carry, store mysql parameters first, for example, you have a mysql account on 192.168.0.1, user is root, password is root and the port is 3306, you call the section dev.
```bash
$ carry store -s dev -H 192.168.0.1 -u root -p root -P 3306
```
if you see error tips, please make sure carry have write permission in dir /usr/local/etc/.
if you see success, that prove you have store a section, if you want login mysql, you can input:

```bash
$ carry login -s dev
```
the condition is that you have installed mysql-client , and set correct environment variables, otherwise you will see error tips.

If you want modify your account password, just need this:
```bash
$ carry store -s dev -p 123456
```

you can watch the sections that you have stored.
```bash
$ carry list # watch all
$ carry list -s dev # only watch section dev
```

You can delete section abandoned.

```bash
carry del -s dev
```

You can transport data between different server, first of all, we add another section named test:
```bash
$ carry store -s test -H 192.168.0.10 -u root -p root -h 3307
```
then, we can transport data between dev server and test server, now we copy database erp from dev to test:
```bash
$ carry trans -f dev.erp -t test.erp
```
also we can copy table user from dev to test.
```bash
$ carry trans -f dev.erp.user -t test.erp.user
```
if table user is not exist on server test, carry will create it.

You can drop table user without login:
```bash
$ carry drop -s dev.erp.user
```
also you can drop database erp without login:
```bash
$ carry drop -s dev.erp
```

When you drop something, carry will confirm you, which protect you from misoperation.

## License
[MIT](LICENSE)
