# Carry
[English Document](https://github.com/joyant/carry/blob/master/Readme.md)

Carry 是一个操作mysql的命令行工具，它让数据传输变得更方便。

```bash
$ carry trans -f dev.database_name.table_name -t dev2.database_name.table_name
```

## 安装
你可以clone后自己编译
```bash
go build main.go
```
你也可以在这里[下载](https://github.com/joyant/carry/releases)到编译过的64位可执行文件，有mac，windows和linux版本。


## 特性
* 在不同的主机间传输数据。
* 登录mysql更快捷。
* 不需要登录就能删除表和数据库。

## 快速开始

在你开始使用Carry前，需要先输入mysql的参数让Carry知道，比如，你现在要登录在主机192.168.0.1上的mysql server, 账号是root，密码是root，端口是3306, 你给这个连接起名叫dev， 这在Carry里有一个专门的术语叫：section。
```bash
$ carry store -s dev -H 192.168.0.1 -u root -p root -P 3306
```
如果你看到错误提示，请确保Carry对目录/usr/local/etc/有读写权限，它需要把你刚才输入的账号保存在这个目录下。

如果你能看到success的提示，证明你已经保存成功了，现在你可以登录mysql了:

```bash
$ carry login -s dev
```
登录的前提是，你已经安装了mysql-client, 并且设置了正确的环境变量，否则你将会看到错误提示。

如果你想修改dev的密码，也很简单：
```bash
$ carry store -s dev -p 123456
```

查看已经保存的section:
```bash
$ carry list # watch all
$ carry list -s dev # only watch section dev
```

当然，你也可以删除已经废弃的section:

```bash
carry del -s dev
```
你可以在不同主机间传输数据， 现在，我们为了演示传输，再加一个section:
```bash
$ carry store -s test -H 192.168.0.10 -u root -p root -h 3307
```
现在，我们可以使用传输命令了，首先，我们把dev上的erp库复制到test主机上：
```bash
$ carry trans -f dev.erp -t test.erp
```
复制一个单独的表也是可以的：
```bash
$ carry trans -f dev.erp.user -t test.erp.user
```
如果user表不存在，Carry会询问你是否要创建它。

你可以不用登录就删除一个表：
```bash
$ carry drop -s dev.erp.user
```
也可以不用登录就删除一个库：
```bash
$ carry drop -s dev.erp
```
当你执行drop操作时，Carry会询问你，以防止误操作。

## License
[MIT](LICENSE)
