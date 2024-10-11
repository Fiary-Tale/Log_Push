package utils

import (
	"os"
	"path/filepath"
	"sync"
)

var fileMutex sync.Mutex

func WriteError(err string) {
	name := "/var/log/waf_alert/Log_Push_error.log"
	fileMutex.Lock()
	defer fileMutex.Unlock()
	// 获取目录路径
	dir := filepath.Dir(name)
	// 创建目录
	_ = os.MkdirAll(dir, 0755)
	f, _ := os.OpenFile(name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	f.Write([]byte(err))
	f.Close()
}
