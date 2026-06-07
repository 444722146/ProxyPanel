package controllers

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"proxypanel/config"
	"proxypanel/utils"
)

// GetAccessLog 获取访问日志
func GetAccessLog(c *gin.Context) {
	domain := c.Query("domain")
	port := c.Query("port")
	lines := c.DefaultQuery("lines", "100")
	
	logPath := filepath.Join(config.AppConfig.Nginx.LogDir, getLogFilename(domain, port, "access"))
	
	logs, err := readLogFile(logPath, lines)
	if err != nil {
		utils.Error(c, 500, "读取访问日志失败: "+err.Error())
		return
	}
	
	utils.Success(c, gin.H{
		"path": logPath,
		"logs": logs,
	})
}

// GetErrorLog 获取错误日志
func GetErrorLog(c *gin.Context) {
	domain := c.Query("domain")
	port := c.Query("port")
	lines := c.DefaultQuery("lines", "100")
	
	logPath := filepath.Join(config.AppConfig.Nginx.LogDir, getLogFilename(domain, port, "error"))
	
	logs, err := readLogFile(logPath, lines)
	if err != nil {
		utils.Error(c, 500, "读取错误日志失败: "+err.Error())
		return
	}
	
	utils.Success(c, gin.H{
		"path": logPath,
		"logs": logs,
	})
}

// GetGeneralAccessLog 获取通用访问日志
func GetGeneralAccessLog(c *gin.Context) {
	lines := c.DefaultQuery("lines", "100")
	
	logPath := filepath.Join(config.AppConfig.Nginx.LogDir, "access.log")
	
	logs, err := readLogFile(logPath, lines)
	if err != nil {
		utils.Error(c, 500, "读取访问日志失败: "+err.Error())
		return
	}
	
	utils.Success(c, gin.H{
		"path": logPath,
		"logs": logs,
	})
}

// GetGeneralErrorLog 获取通用错误日志
func GetGeneralErrorLog(c *gin.Context) {
	lines := c.DefaultQuery("lines", "100")
	
	logPath := filepath.Join(config.AppConfig.Nginx.LogDir, "error.log")
	
	logs, err := readLogFile(logPath, lines)
	if err != nil {
		utils.Error(c, 500, "读取错误日志失败: "+err.Error())
		return
	}
	
	utils.Success(c, gin.H{
		"path": logPath,
		"logs": logs,
	})
}

// ClearLog 清空日志
func ClearLog(c *gin.Context) {
	logType := c.Query("type") // access 或 error
	domain := c.Query("domain")
	port := c.Query("port")
	
	logPath := filepath.Join(config.AppConfig.Nginx.LogDir, getLogFilename(domain, port, logType))
	
	// 清空日志文件（不删除，保留文件）
	if err := os.WriteFile(logPath, []byte{}, 0644); err != nil {
		utils.Error(c, 500, "清空日志失败: "+err.Error())
		return
	}
	
	utils.SuccessWithMsg(c, "日志已清空", nil)
}

// SearchLog 搜索日志
func SearchLog(c *gin.Context) {
	domain := c.Query("domain")
	port := c.Query("port")
	logType := c.Query("type") // access 或 error
	keyword := c.Query("keyword")
	lines := c.DefaultQuery("lines", "100")
	
	if keyword == "" {
		utils.BadRequest(c, "搜索关键词不能为空")
		return
	}
	
	logPath := filepath.Join(config.AppConfig.Nginx.LogDir, getLogFilename(domain, port, logType))
	
	logs, err := searchLogFile(logPath, keyword, lines)
	if err != nil {
		utils.Error(c, 500, "搜索日志失败: "+err.Error())
		return
	}
	
	utils.Success(c, gin.H{
		"path":    logPath,
		"keyword": keyword,
		"logs":    logs,
	})
}

// readLogFile 读取日志文件
func readLogFile(path string, linesStr string) ([]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []string{"日志文件不存在"}, nil
	}

	// 解析行数限制
	maxLines := 100
	if linesStr != "" {
		if parsed, err := parseLines(linesStr); err == nil {
			maxLines = parsed
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 使用bufio Scanner读取，避免内存溢出
	var lines []string
	scanner := bufio.NewScanner(file)
	
	// 先读取所有行到数组（为了从末尾开始）
	allLines := []string{}
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	
	// 获取最后N行
	start := len(allLines) - maxLines
	if start < 0 {
		start = 0
	}
	
	lines = allLines[start:]
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return lines, nil
}

// searchLogFile 搜索日志文件
func searchLogFile(path string, keyword string, linesStr string) ([]string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []string{"日志文件不存在"}, nil
	}

	maxLines := 100
	if linesStr != "" {
		if parsed, err := parseLines(linesStr); err == nil {
			maxLines = parsed
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matchedLines []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, keyword) {
			matchedLines = append(matchedLines, line)
		}
	}
	
	// 限制返回数量
	if len(matchedLines) > maxLines {
		matchedLines = matchedLines[len(matchedLines)-maxLines:]
	}
	
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	
	return matchedLines, nil
}

// getLogFilename 获取日志文件名（与 Nginx 模板生成的文件名保持一致）
func getLogFilename(domain string, port string, logType string) string {
	if domain != "" {
		if port != "" {
			return "proxy_" + domain + "_" + port + "_" + logType + ".log"
		}
		return "proxy_" + domain + "_" + logType + ".log"
	}
	return logType + ".log"
}

// parseLines 解析行数参数
func parseLines(linesStr string) (int, error) {
	var lines int
	_, err := sscanf(linesStr, "%d", &lines)
	return lines, err
}

// sscanf 简单的字符串解析函数
func sscanf(str string, format string, dest *int) (int, error) {
	// 简化实现：直接转换
	var result int
	var err error
	
	for i := 0; i < len(str); i++ {
		if str[i] >= '0' && str[i] <= '9' {
			result = result * 10 + int(str[i] - '0')
		}
	}
	
	if result == 0 && str != "0" {
		err = os.ErrInvalid
	}
	
	*dest = result
	return 1, err
}