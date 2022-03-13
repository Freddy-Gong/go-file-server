package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/Freddy-Gong/file-server/server/controller"
	"github.com/gin-gonic/gin"
)

//把前端打包到go生产的文件上
//go:embed frontend/dist/*
var FS embed.FS

func Run() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	staticFiles, _ := fs.Sub(FS, "frontend/dist")
	router.POST("/api/v1/files", controller.FilesController)
	router.GET("/api/v1/qrcodes", controller.QrcodesController)
	router.GET("/uploads/:path", controller.UploadsController)
	router.POST("/api/v1/texts", controller.TextsController)
	router.GET("/api/v1/addresses", controller.AddressesController)
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
}
