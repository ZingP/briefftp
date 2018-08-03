package main

import (
	"fmt"
	"sync"
	"time"
	"github.com/astaxie/beego/logs"
)

var (
	appConf   AppConf
	timeStamp string
	timeDay   string
	timeMon   string
	hostMap   map[string]string
	rsyncChan = make(chan *rsyncLog, 60)
	lzopChan  = make(chan *lzopLog)
	hadpChan  = make(chan *hadpLog)

	waitGroup sync.WaitGroup
)

func initTimeStr() {
	h, _ := time.ParseDuration("-1h")
	timeStamp = time.Now().Add(h).Format("2006-01-02_15")
	timeDay = time.Now().Add(h).Format("20060102")
	timeMon = time.Now().Add(h).Format("200601")
}

func init() {
	confFile := "./tohadoop.cfg"
	err := initConfig(confFile)
	if err != nil {
		fmt.Printf("load conf failed:%v\n", err)
		return
	}

	initTimeStr()

	logfile := fmt.Sprintf("%s.%s", appConf.logFile, timeStamp)
	err = initLogs(logfile)
	if err != nil {
		fmt.Printf("init log failed:%v\n", err)
		return
	}

	err = GetHostMap(appConf.hostFile)
	if err != nil {
		fmt.Printf("get host map failed:%v\n", err)
	}

}

func main() {
	genRsyncLogObj(appConf.remoteAddr, appConf.localAddr, appConf.logName)

	for i := 0; i < appConf.threadNum; i++ {
		waitGroup.Add(1)
		go rsyncToLocal()
	}
	for i := 0; i < appConf.threadNum; i++ {
		waitGroup.Add(1)
		go lzopToLocal()
	}
	for i := 0; i < appConf.threadNum; i++ {
		waitGroup.Add(1)
		go putToHdfs()
	}

	waitGroup.Wait()

	logs.Info("all done.")
}