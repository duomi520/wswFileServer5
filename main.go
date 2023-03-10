// FileServer project main.go
package main

//多文件上传
//作者：
//邮箱：
//日期：2015-4-3
//Bootstrap + Golang + HTML5 实现带进度条的多文件上传

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/gin-contrib/pprof"

	"github.com/gin-gonic/gin"
)

// ListFiles 文件列表
type ListFiles struct {
	Name    string `json:"name"`
	Size    string `json:"size"`
	ModTime string `json:"time"`
	modTime int64
}

// ByModTime 排序
type ByModTime []ListFiles

func (a ByModTime) Len() int           { return len(a) }
func (a ByModTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByModTime) Less(i, j int) bool { return a[i].modTime > a[j].modTime }

var listFilesMap sync.Map
var memoryFile map[string][]byte

func cachedFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	memoryFile[filename] = data
}

func main() {
	memoryFile = make(map[string][]byte)
	log.Println("Server start")
	defer log.Println("Server stop")
	gin.SetMode(gin.ReleaseMode)
	//路由设置
	r := gin.Default()
	r.Static("/static", "./static")
	r.Static("/spaces", "./spaces")
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("favicon.ico")
	})
	cachedFile("./spaces/space.html")
	r.GET("/space/:name", func(c *gin.Context) {
		name := c.Param("name")
		//判断目录是否存在
		_, err := os.Stat("./spaces/" + name)
		if err != nil {
			c.String(http.StatusNotFound, "404 page not found")
		} else {
			var jsondata []byte
			lm, ok := listFilesMap.Load(name)
			if !ok {
				lm = refreshCache(name)
			}
			//返回目录json数据
			jsondata, err = json.Marshal(lm.([]ListFiles))
			if err != nil {
				log.Println("main :", err.Error())
				c.String(http.StatusInternalServerError, "InternalServerError")
				return
			}
			c.Writer.Write(memoryFile["./spaces/space.html"])
			c.Writer.Write([]byte("\nvar datas = "))
			c.Writer.Write(jsondata)
			c.Writer.Write([]byte(";\n</script>\n</body>\n</html>"))
			//c.File("./spaces/space.html")
		}
	})
	//r.GET("/directory/:name", List)
	r.POST("/directory/:name", Upload)
	r.DELETE("/directory/:name", Delete)
	//启动服务
	srv := &http.Server{
		Handler:        r,
		ReadTimeout:    3600 * time.Second,
		WriteTimeout:   3600 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	//监听信号 ctrl+c kill
	exitChan := make(chan os.Signal, 16)
	signal.Notify(exitChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-exitChan
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("main :", err.Error())
		}
	}()
	//debug 模式
	if len(os.Args) > 1 {
		if strings.EqualFold(os.Args[1], "debug") {
			pprof.Register(r)
			log.Println("Debug mode")
		}
	}
	//-------------------------------------
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	if err := srv.ListenAndServe(); err != nil {
		log.Println(err.Error())
	}
}

// DeleteFiles 删除文件
type DeleteFiles struct {
	FileName []string `json:"filename" binding:"required"`
}

// Delete 删除
func Delete(c *gin.Context) {
	name := c.Param("name")
	var fl DeleteFiles
	lost := ""
	c.Bind(&fl)
	for _, file := range fl.FileName {
		//删除文件
		err := os.Remove("./spaces/" + name + "/" + file)
		if err != nil {
			log.Println(file, " 删除失败：", err)
			lost = lost + " | " + file
		}
	}
	lm := refreshCache(name)
	if len(lost) == 0 {
		c.JSON(http.StatusOK, lm)
	} else {
		c.JSON(http.StatusInternalServerError, lm)
	}
}

func refreshCache(directory string) []ListFiles {
	lm := make([]ListFiles, 0)
	//遍历目录，读出文件名、大小
	filepath.Walk("./spaces/"+directory, func(path string, fi os.FileInfo, err error) error {
		if fi == nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		var m ListFiles
		//避免XSS
		m.Name = template.HTMLEscapeString(fi.Name())
		m.Size = strconv.FormatInt(fi.Size()/1024, 10)
		m.modTime = fi.ModTime().Unix()
		m.ModTime = fi.ModTime().Format("2006-01-02 15:04:05")
		lm = append(lm, m)
		return nil
	})
	//排序
	sort.Sort(ByModTime(lm))
	//缓存
	listFilesMap.Store(directory, lm)
	return lm
}

/*
// List 列出文件清单
func List(c *gin.Context) {
	name := c.Param("name")
	lm, ok := listFilesMap.Load(name)
	if ok {
		//返回目录json数据
		c.JSON(http.StatusOK, lm.([]ListFiles))
	} else {
		c.JSON(http.StatusOK, refreshCache(name))
	}
}
*/
// Upload 上传
func Upload(c *gin.Context) {
	name := c.Param("name")
	//在使用r.MultipartForm前必须先调用ParseMultipartForm方法，参数为最大缓存
	c.Request.ParseMultipartForm(32 << 20)
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
		//获取所有上传文件信息
		fhs := c.Request.MultipartForm.File["userfile"]
		num := len(fhs)
		log.Printf("总文件数：%d 个文件", num)
		//循环对每个文件进行处理
		for n, fheader := range fhs {
			str := fheader.Filename
			//替换"/"
			str = strings.Replace(str, "/", "", -1)
			//替换"\"
			str = strings.Replace(str, "\\", "", -1)
			//避免XSS
			str = template.HTMLEscapeString(str)
			//设置文件名
			newFileName := "./spaces/" + name + "/" + str
			//打开上传文件
			uploadFile, err := fheader.Open()
			defer func() {
				if err = uploadFile.Close(); err != nil {
					log.Println(err)
				}
			}()
			if err != nil {
				log.Println(err)
				c.String(http.StatusBadRequest, "上传失败:", err.Error())
				return
			}
			//保存文件
			saveFile, err := os.OpenFile(newFileName, os.O_WRONLY|os.O_CREATE, 0666)
			defer func() {
				if err = saveFile.Close(); err != nil {
					log.Println(err)
				}
			}()
			if err != nil {
				log.Println(err)
				c.String(http.StatusBadRequest, "上传失败:", err.Error())
				return
			}
			io.Copy(saveFile, uploadFile)
			//获取文件状态信息
			fileStat, _ := saveFile.Stat()
			//打印接收信息
			log.Printf(" NO.: %d  Size: %d KB  Name：%s\n", n, fileStat.Size()/1024, newFileName)

		}
		lm := refreshCache(name)
		c.JSON(http.StatusOK, lm)
	}

}

//参考:
// https://developer.mozilla.org/zh-CN/docs/Web/Guide/Using_FormData_Objects
