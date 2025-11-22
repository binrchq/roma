package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"binrc.com/roma/core/global"
)

// HistoryManager 历史管理器
type HistoryManager struct {
	historyFile string
	maxLines    int
	maxSize     int
}

// NewHistoryManager 创建历史管理器
func NewHistoryManager(username string) *HistoryManager {
	historyDir := global.CONFIG.Common.HistoryTmpDir
	if historyDir == "" {
		historyDir = "/tmp/roma_history"
	}

	// 创建用户目录
	userDir := filepath.Join(historyDir, username)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		// 如果创建失败，使用默认目录
		userDir = historyDir
	}

	// 历史文件路径: {HistoryTmpDir}/{username}/.roma_history
	historyFile := filepath.Join(userDir, ".roma_history")

	maxLines := global.CONFIG.Common.HistoryTmpMaxLine
	if maxLines <= 0 {
		maxLines = 1000 // 默认1000行
	}

	maxSize := global.CONFIG.Common.HistoryTmpMaxSize
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // 默认10MB
	}

	return &HistoryManager{
		historyFile: historyFile,
		maxLines:    maxLines,
		maxSize:     maxSize,
	}
}

// LoadHistory 从文件加载历史记录
func (hm *HistoryManager) LoadHistory() []string {
	var history []string

	file, err := os.Open(hm.historyFile)
	if err != nil {
		// 文件不存在或无法打开，返回空历史
		return history
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			history = append(history, line)
		}
	}

	if err := scanner.Err(); err != nil {
		// 读取错误，返回已加载的历史
		return history
	}

	// 限制历史记录数量
	if len(history) > hm.maxLines {
		history = history[len(history)-hm.maxLines:]
	}

	return history
}

// SaveHistory 保存历史记录到文件
func (hm *HistoryManager) SaveHistory(history []string) error {
	// 确保目录存在
	dir := filepath.Dir(hm.historyFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// 限制历史记录数量
	if len(history) > hm.maxLines {
		history = history[len(history)-hm.maxLines:]
	}

	// 打开文件进行写入（覆盖模式）
	file, err := os.Create(hm.historyFile)
	if err != nil {
		return fmt.Errorf("failed to create history file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// 写入历史记录
	for _, line := range history {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write history: %w", err)
		}
	}

	return nil
}

// AppendHistory 追加单条历史记录到文件
func (hm *HistoryManager) AppendHistory(command string) error {
	if command == "" {
		return nil
	}

	// 确保目录存在
	dir := filepath.Dir(hm.historyFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// 检查文件大小
	fileInfo, err := os.Stat(hm.historyFile)
	if err == nil && fileInfo.Size() > int64(hm.maxSize) {
		// 文件太大，需要清理
		history := hm.LoadHistory()
		// 只保留最近的一半
		if len(history) > hm.maxLines/2 {
			history = history[len(history)-hm.maxLines/2:]
		}
		if err := hm.SaveHistory(history); err != nil {
			return err
		}
	}

	// 追加模式打开文件
	file, err := os.OpenFile(hm.historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer file.Close()

	// 写入命令
	if _, err := file.WriteString(command + "\n"); err != nil {
		return fmt.Errorf("failed to append history: %w", err)
	}

	return nil
}
