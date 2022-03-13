package controller

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
