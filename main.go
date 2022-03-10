package main

import (
	"embed"
	"io/fs"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
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

func AddressesController(c *gin.Context) {
	addrs, _ := net.InterfaceAddrs() //当前电脑的所有Ip地址
	var result []string
	for _, address := range addrs {
		//address.(*net.IPNet)是一个断言，判断address的类型是不是IPV4
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				result = append(result, ipnet.IP.String())
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"addresses": result})
}

func GetUploadsDir() (uploads string) {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Dir(exe)
	uploads = filepath.Join(dir, "uploads")
	return
}

func UploadsController(c *gin.Context) {
	if path := c.Param("path"); path != "" {
		target := filepath.Join(GetUploadsDir(), path)
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", "attachment; filename="+path)
		c.Header("Content-Type", "application/octet-stream")
		c.File(target) //给前端发送一个文件
	} else {
		c.Status(http.StatusNotFound)
	}
}

func QrcodesController(c *gin.Context) {
	if content := c.Query("content"); content != "" {
		png, err := qrcode.Encode(content, qrcode.Medium, 256)
		if err != nil {
			log.Fatal(err)
		}
		c.Data(http.StatusOK, "image/png", png)
	} else {
		c.Status(http.StatusBadRequest)
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
		router.GET("/api/v1/qrcodes", QrcodesController)
		router.GET("/uploads/:path", UploadsController)
		router.POST("/api/v1/texts", TextsController)
		router.GET("/api/v1/addresses", AddressesController)
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
