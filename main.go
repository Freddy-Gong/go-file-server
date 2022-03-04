package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/zserge/lorca"
)

func main() {
	var ui lorca.UI //接口：只要声明这个借口的变量，他就具有接口中的方法
	currentDir, _ := os.Getwd()
	dir := filepath.Join(currentDir, ".cache")
	ui, _ = lorca.New("https://baidu.com", dir, 800, 600, "--disable-sync", "--disable-translate")
	//os.signal 操作系统的信号
	chSignal := make(chan os.Signal, 1)
	//订阅信号，监听系统调用的终止信号，把信号传给chSignal
	signal.Notify(chSignal, syscall.SIGINT, syscall.SIGTERM) //中断信号 终止信号
	//专门用来进行channel的筛选的
	//同时尝试读取两个channel的信号，把首先读到的信号进行执行
	//select会阻塞当前线程，不会继续执行了
	//-------------------------------------------------------------
	// select {
	// case <-ui.Done(): //:后面为党case进行选择之后执行的代码
	// 	ui.Close()
	// 	return
	// case <-chSignal:
	// 	ui.Close()
	// 	return
	// }
	select {
	case <-ui.Done(): //:后面为党case进行选择之后执行的代码
	case <-chSignal:
	}
	ui.Close()
}
