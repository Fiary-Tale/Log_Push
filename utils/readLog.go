package utils

import (
	"bufio"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

// Config 结构体，用于存储从配置文件读取的配置信息
type Config struct {
	Token  string `yaml:"token"` // DingTalk 机器人access_token
	Method string `yaml:"method"`
}

// 定义日志结构

type LogEntry struct {
	Timestamp string
	SrcCity   string
	SrcIP     string
	DstHost   string
	URL       string
	RuleID    string
	LogID     string
}

// 解析日志行，返回 LogEntry 对象
func parseLogEntry(line string) LogEntry {
	// 示例日志格式:
	// 2024-10-11 09:59:13 ubuntu safeline-ce: src_city: , src_ip:10.10.0.244, dst_host:10.10.0.246:80, url:/etc/shadow, rule_id:m_rule/33e75ff09dd601bbe69f351039152189, log_id: 5
	parts := strings.Split(line, " ")
	timestamp := parts[0] + " " + parts[1]
	logDetails := strings.Split(strings.Join(parts[4:], " "), ", ")

	var entry LogEntry
	entry.Timestamp = timestamp

	for _, detail := range logDetails {
		keyValue := strings.Split(detail, ":")
		key := keyValue[0]
		value := strings.TrimSpace(keyValue[1])
		switch key {
		case "src_city":
			entry.SrcCity = value
		case "src_ip":
			entry.SrcIP = value
		case "dst_host":
			entry.DstHost = value
		case "url":
			entry.URL = value
		case "rule_id":
			entry.RuleID = value
		case "log_id":
			entry.LogID = value
		}
	}

	return entry
}

// 读取文件的最后一行
func readLastLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var lastLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastLine = scanner.Text() // 不断更新，最后得到的就是最后一行
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}
	return lastLine, nil
}

// 解析yaml文件

func ReadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var Conf Config
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		return nil, err
	}
	return &Conf, nil
}

// 读取yaml文件，并执行消息推送

func Exec(name string) {
	conf, err := ReadConfig(name)
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		WriteError(fmt.Sprintf("%s Error reading config file: %v\n", time.Now().Format("2006-01-02 15:04:05"), err))
	}
	Message(conf)
}
