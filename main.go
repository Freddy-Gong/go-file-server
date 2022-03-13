package main

import (
	"os"
	"os/exec"
	"os/signal"

	"github.com/Freddy-Gong/file-server/config"
	"github.com/Freddy-Gong/file-server/server"
)

func main() {
	go server.Run()

	cmd := startBrowser()
	chSignal := listenToInterrupt()
	select {
	case <-chSignal:
		cmd.Process.Kill()
	}
	//<-chSignal//如果channel中没有值的传入，那就会阻塞在这里
	//cmd.Process.Kill()
	//select {} //目的是不让mian进行结束，因为如果main进程结束了，里面开的gin进程
	//也就会结束，所以就是导致无法访问
}

func startBrowser() *exec.Cmd {
	chromePath := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	cmd := exec.Command(chromePath, "--app=http://localhost:"+config.GetPort()+"/static/index.html")
	cmd.Start()
	return cmd
}

func listenToInterrupt() chan os.Signal {
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt) //监听Ctrl+C的信号，有信号就传给channel
	return chSignal
}

//接口是对方法的约束
//type是对属性的约束
//go中 & 一般接在一个值的前面
//*一般接在一个类型前面
