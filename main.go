package main

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TextsController(c *gin.Context) {
	var json struct {
		Raw string
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		exe, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		dir := filepath.Dir(exe)
		if err != nil {
			log.Fatal(err)
		}
		filename := uuid.New().String()
		uploads := filepath.Join(dir, "uploads")
		err = os.MkdirAll(uploads, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		fullpath := path.Join("uploads", filename+".txt")
		err = ioutil.WriteFile(filepath.Join(dir, fullpath), []byte(json.Raw), 0644)
		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, gin.H{"url": "/" + fullpath})
	}
}

//把前端打包到go生产的文件上
//go:embed frontend/dist/*
var FS embed.FS

func main() {
	go func() {
		gin.SetMode(gin.ReleaseMode)
		router := gin.Default()
		staticFiles, _ := fs.Sub(FS, "frontend/dist")
		router.POST("api/v1/texts", TextsController)
		router.StaticFS("/static", http.FS(staticFiles))
		router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/static/") {
				reader, err := staticFiles.Open("index.html")
				if err != nil {
					log.Fatal(err)
				}
				defer reader.Close()
				stat, err := reader.Stat()
				if err != nil {
					log.Fatal(err)
				}
				c.DataFromReader(http.StatusOK, stat.Size(), "text/html;charset=utf-8", reader, nil)
			} else {
				c.Status(http.StatusNotFound)
			}
		})
		router.Run(":8080")
	}()
	//开一个子进程启动chrome
	chromePath := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	cmd := exec.Command(chromePath, "--app=http://localhost:8080/static/index.html")
	cmd.Start()
	//<-chSignal//如果channel中没有值的传入，那就会阻塞在这里
	//cmd.Process.Kill()
	//select {} //目的是不让mian进行结束，因为如果main进程结束了，里面开的gin进程
	//也就会结束，所以就是导致无法访问
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt) //监听Ctrl+C的信号，有信号就传给channel
	select {
	case <-chSignal:
		cmd.Process.Kill()
	}
}

//接口是对方法的约束
//type是对属性的约束
//go中 & 一般接在一个值的前面
//*一般接在一个类型前面
