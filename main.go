// wswFileServer5 project main.go
package main

//多文件上传
//作者：
//邮箱：
//日期：2015-4-3
//Bootstrap + Golang + HTML5 实现带进度条的多文件上传
//参考:
// https://developer.mozilla.org/zh-CN/docs/Web/Guide/Using_FormData_Objects
// http://www.cnblogs.com/fredlau/archive/2008/08/12/1266089.html

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

var port = ":8086"      //端口
var storage = "./files" //上传文件目录
var currentDirectory string

func main() {
	//读取本地局域网IP
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Error : " + err.Error())
	}
	for _, inter := range interfaces {
		temp, _ := inter.Addrs()
		for _, addr := range temp {
			if addr.String() != "0.0.0.0" {
				fmt.Println("Server local Ip address:" + addr.String() + port)
			}
		}
	}
	//读取当前目录
	tempFile, _ := exec.LookPath(os.Args[0])
	currentDirectory = filepath.Dir(tempFile)
	//路由设置
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.Static("/static", "./static")
	r.Static("/files", "./files")
	r.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})
	r.GET("/file", List)
	r.POST("/file", Upload)
	r.DELETE("/file", Delete)
	//启动服务
	s := &http.Server{
		Addr:           port,
		Handler:        r,
		ReadTimeout:    3600 * time.Second,
		WriteTimeout:   3600 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

type DeleteFiles struct {
	FileName []string `json:"filename" binding:"required"`
}

func Delete(c *gin.Context) {
	var fl DeleteFiles
	c.Bind(&fl)
	for _, file := range fl.FileName {
		err := os.Remove(storage + "/" + file) //删除文件
		if err != nil {
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"), file, "失败删除文件", err)
		}
	}
	c.String(http.StatusOK, "删除文件结束")
}

type ListFiles struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

func List(c *gin.Context) {
	lm := make([]ListFiles, 0)
	//遍历目录，读出文件名、大小
	filepath.Walk(storage, func(path string, fi os.FileInfo, err error) error {
		if nil == fi {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		//	fmt.Println(fi.Name(), fi.Size()/1024)
		var m ListFiles
		m.Name = fi.Name()
		m.Size = strconv.FormatInt(fi.Size()/1024, 10)
		lm = append(lm, m)
		return nil
	})
	//返回目录json数据
	c.JSON(http.StatusOK, lm)
}
func Upload(c *gin.Context) {
	c.Request.ParseMultipartForm(32 << 20) //在使用r.MultipartForm前必须先调用ParseMultipartForm方法，参数为最大缓存
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		os.Chdir(storage)                               //进入存储目录
		defer os.Chdir(currentDirectory)                //退出存储目录
		fhs := c.Request.MultipartForm.File["userfile"] //获取所有上传文件信息
		num := len(fhs)
		fmt.Printf("总文件数：%d 个文件", num)
		//循环对每个文件进行处理
		for n, fheader := range fhs {
			//设置文件名
			//newFileName := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(fheader.Filename) //十进制
			newFileName := fheader.Filename
			//打开上传文件
			uploadFile, err := fheader.Open()
			if err != nil {
				fmt.Println(err)
				c.String(http.StatusBadRequest, "上传失败:", err.Error())
				return
			}
			defer uploadFile.Close()
			//保存文件
			saveFile, err := os.Create(newFileName)
			if err != nil {
				fmt.Println(err)
				c.String(http.StatusBadRequest, "上传失败:", err.Error())
				return
			}
			defer saveFile.Close()
			io.Copy(saveFile, uploadFile)

			//获取文件状态信息
			fileStat, _ := saveFile.Stat()
			//打印接收信息
			fmt.Printf("%s  NO.: %d  Size: %d KB  Name：%s\n", time.Now().Format("2006-01-02 15:04:05"), n, fileStat.Size()/1024, newFileName)

		}
		c.String(http.StatusOK, "上传成功")
	}

}
