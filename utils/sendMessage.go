package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func Message(config *Config) {
	logfile := "/var/log/waf_alert/waf_alert.log"
	lastLine, err := readLastLine(logfile)
	if err != nil {
		fmt.Println("Error reading last line:", err)
		return
	}
	// 解析最后一条日志
	logEntry := parseLogEntry(lastLine)
	// 获取当前时间
	currentTime := time.Now()
	// 格式化时间为 "YYYY-MM-DD HH:MM:SS" 格式
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	// 构造message信息
	message := fmt.Sprintf(
		"**入侵检测事件**\n\n- **告警通知时间:**\n  %s\n- **日志告警时间:**\n  %s\n\n- **受影响设备地址:**\n\n  %s\n- **攻击源地址:**\n\n  %s\n- **攻击来源:**\n\n  %s\n- **触发规则:**\n\n  %s\n- **被攻击路径:**\n\n  %s\n- **原始攻击日志:**\n\n  %s",
		formattedTime,      // 告警通知时间
		logEntry.Timestamp, // 日志告警时间
		logEntry.DstHost,   // 受影响的设备地址
		logEntry.SrcCity,   // 攻击来源地址
		logEntry.SrcIP,     // 攻击来源IP
		logEntry.RuleID,    // 规则ID
		logEntry.URL,       // 被攻击路径
		lastLine,           // 原始攻击日志
	)
	data := DingTalkMessage{
		MsgType: "markdown",
		Markdown: struct {
			Title string `json:"title"`
			Text  string `json:"text"`
		}{
			Title: "入侵检测事件",
			Text:  message,
		},
	}
	switch config.Method {
	case "dingtalk":
		sendDingTalkMessage(config.Token, data)
	case "serverchan":
		sendServerChatMessage(config.Token, data)
	default:
		log.Fatalf("未知的推送方法: %s", config.Method)
	}
}

// 钉钉推送
func sendDingTalkMessage(token string, message DingTalkMessage) {
	data, err := json.Marshal(message) // 将消息结构体转换为JSON
	if err != nil {
		log.Printf("Error sending DingTalk message: %v", err)
		WriteError(fmt.Sprintf("Error sending DingTalk message: %v", err))
		return
	}
	webhookURL := fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", token) // 构建Webhook URL
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))            // 发送HTTP POST请求
	if err != nil {
		log.Printf("Error sending DingTalk message: %v", err)
		WriteError(fmt.Sprintf("Error sending DingTalk message: %v", err))
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error sending DingTalk message: %v", resp.Status)              // 打印错误状态码
		WriteError(fmt.Sprintf("Error sending DingTalk message: %v", resp.Status)) // 写入错误日志
	}
}

// Server酱推送
func sendServerChatMessage(token string, message DingTalkMessage) {
	data := url.Values{}
	data.Add("title", message.Markdown.Title)
	data.Add("text", message.Markdown.Text)
	webhookURL := fmt.Sprintf("https://sctapi.ftqq.com/%s.send", token) // 构建Webhook URL
	resp, err := http.PostForm(webhookURL, data)                        // 发送HTTP POST请求
	if err != nil {
		log.Printf("Error sending ServerChat message: %v", err)
		WriteError(fmt.Sprintf("Error sending ServerChat message: %v", resp.Status))
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error sending ServerChat message: %v", resp.Status)
		WriteError(fmt.Sprintf("Error sending ServerChat message: %v", resp.Status))
	}
}
