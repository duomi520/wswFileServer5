# FileServer

## 部署

在win系统中，从控制台进入项目目录，执行如下操作

```bash
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build main.go
```

生成linux的二进制文件，将整个目录拷贝到linux服务器下，

```bash
chmod 777 main
nohup ./main >a.log &
```

赋予main执行权限，执行 nohup ./main >a.log &，这样程序在后台运行了。

## 卸载

将该目录删除即可。

## 配置

在spaces目录下建目录，访问方式为http：//xxx.xxx.xxx.xxx/space/目录名

## 功能

点击选择文件按键可以选择本地上传文件（可多选），按蓝色按键上传到服务器。列表为服务器上的文件清单，点击链接可以下载到本地。在右侧的选择框可以选中文件，然后点击红色的按键，即可删除服务器上的文件。
