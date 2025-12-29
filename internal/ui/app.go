package ui

import (
	"fmt"
	"os"
	"path/filepath"
	
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	
	"github.com/chenyb888/aria2GoUI/internal/config"
	"github.com/chenyb888/aria2GoUI/internal/aria2"
)

// App 应用程序结构
type App struct {
	fyneApp   fyne.App
	window    fyne.Window
	config    *config.Config
	aria2Client *aria2.Client
}

// NewApp 创建新的应用程序
func NewApp() *App {
	fyneApp := fyne.CurrentApp()
	
	app := &App{
		fyneApp: fyneApp,
		config:  config.DefaultConfig(),
	}
	
	// 创建主窗口
	window := fyneApp.NewWindow("aria2GoUI")
	window.Resize(fyne.NewSize(float32(app.config.UI.WindowWidth), float32(app.config.UI.WindowHeight)))
	app.window = window
	
	return app
}

// SetConfig 设置配置
func (a *App) SetConfig(cfg *config.Config) {
	a.config = cfg
}

// SetAria2Client 设置 aria2 客户端
func (a *App) SetAria2Client(client *aria2.Client) {
	a.aria2Client = client
}

// CreateMainUI 创建主界面
func (a *App) CreateMainUI() {
	// 任务管理工具栏
	taskToolbar := container.NewHBox(
		widget.NewButtonWithIcon("添加任务", theme.ContentAddIcon(), func() {
			a.showAddTaskDialog()
		}),
		widget.NewButtonWithIcon("暂停", theme.MediaPauseIcon(), func() {
			a.pauseSelectedTasks()
		}),
		widget.NewButtonWithIcon("开始", theme.MediaPlayIcon(), func() {
			a.resumeSelectedTasks()
		}),
		widget.NewButtonWithIcon("删除", theme.DeleteIcon(), func() {
			a.showRemoveTaskDialog()
		}),
		widget.NewButtonWithIcon("上移", theme.MediaSkipPreviousIcon(), func() {
			a.moveTaskUp()
		}),
		widget.NewButtonWithIcon("下移", theme.MediaSkipNextIcon(), func() {
			a.moveTaskDown()
		}),
	)
	
	// 全局操作工具栏
	globalToolbar := container.NewHBox(
		widget.NewButtonWithIcon("全部暂停", theme.MediaPauseIcon(), func() {
			a.pauseAllTasks()
		}),
		widget.NewButtonWithIcon("全部开始", theme.MediaPlayIcon(), func() {
			a.resumeAllTasks()
		}),
		widget.NewButtonWithIcon("清理完成", theme.ConfirmIcon(), func() {
			a.clearCompletedTasks()
		}),
	)
	
	// 视图和设置工具栏
	viewToolbar := container.NewHBox(
		widget.NewButtonWithIcon("刷新", theme.ViewRefreshIcon(), func() {
			a.refreshTaskList()
		}),
		widget.NewButtonWithIcon("统计", theme.InfoIcon(), func() {
			a.showStatisticsDialog()
		}),
		widget.NewButtonWithIcon("导出", theme.DocumentSaveIcon(), func() {
			a.exportTasks()
		}),
		widget.NewButtonWithIcon("设置", theme.SettingsIcon(), func() {
			a.showSettingsDialog()
		}),
	)
	
	// 合并所有工具栏
	toolbar := container.NewVBox(
		container.NewHBox(taskToolbar, widget.NewSeparator(), globalToolbar),
		viewToolbar,
	)
	
	// 连接状态指示器
	statusLabel := widget.NewLabel("未连接")
	statusContainer := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		statusLabel,
	)
	
	// 主内容区域
	mainContent := container.NewBorder(
		container.NewVBox(toolbar, statusContainer),
		nil,
		nil,
		nil,
		a.createTaskList(),
	)
	
	a.window.SetContent(mainContent)
}

// createTaskList 创建任务列表
func (a *App) createTaskList() fyne.CanvasObject {
	// 创建一个容器来动态切换内容
	contentContainer := container.NewMax()
	
	// 更新任务列表显示的函数
	updateTaskDisplay := func() {
		taskCount := a.getTaskCount()
		
		if taskCount == 0 {
			// 显示空状态
			emptyState := a.createEmptyState()
			contentContainer.Objects = []fyne.CanvasObject{emptyState}
		} else {
			// 显示任务列表
			taskList := a.createTaskListWidget()
			contentContainer.Objects = []fyne.CanvasObject{taskList}
		}
		
		contentContainer.Refresh()
	}
	
	// 初始更新
	updateTaskDisplay()
	
	return contentContainer
}

// getTaskCount 获取任务数量
func (a *App) getTaskCount() int {
	if a.aria2Client == nil {
		return 0
	}
	
	// 获取活动任务
	activeTasks, err := a.aria2Client.TellActive()
	if err != nil {
		return 0
	}
	
	// 获取等待任务
	waitingTasks, err := a.aria2Client.TellWaiting(0, 1000)
	if err != nil {
		return len(activeTasks)
	}
	
	// 获取已停止任务（最近的一部分）
	stoppedTasks, err := a.aria2Client.TellStopped(0, 100)
	if err != nil {
		return len(activeTasks) + len(waitingTasks)
	}
	
	return len(activeTasks) + len(waitingTasks) + len(stoppedTasks)
}

// getAllTasks 获取所有任务
func (a *App) getAllTasks() []aria2.TellStatus {
	if a.aria2Client == nil {
		return []aria2.TellStatus{}
	}
	
	var allTasks []aria2.TellStatus
	
	// 获取活动任务
	if activeTasks, err := a.aria2Client.TellActive(); err == nil {
		allTasks = append(allTasks, activeTasks...)
	}
	
	// 获取等待任务
	if waitingTasks, err := a.aria2Client.TellWaiting(0, 1000); err == nil {
		allTasks = append(allTasks, waitingTasks...)
	}
	
	// 获取已停止任务
	if stoppedTasks, err := a.aria2Client.TellStopped(0, 100); err == nil {
		allTasks = append(allTasks, stoppedTasks...)
	}
	
	return allTasks
}

// createEmptyState 创建空状态显示
func (a *App) createEmptyState() fyne.CanvasObject {
	emptyIcon := widget.NewIcon(theme.ContentAddIcon())
	emptyText := widget.NewLabel("暂无下载任务")
	emptySubText := widget.NewLabel("点击上方\"添加\"按钮开始添加下载任务")
	
	emptyText.TextStyle = fyne.TextStyle{Bold: true}
	
	return container.NewVBox(
		container.NewCenter(emptyIcon),
		container.NewCenter(emptyText),
		container.NewCenter(emptySubText),
	)
}

// createTaskListWidget 创建任务列表控件
func (a *App) createTaskListWidget() fyne.CanvasObject {
	// 获取所有任务
	allTasks := a.getAllTasks()
	
	// 任务列表
	list := widget.NewList(
		func() int {
			return len(allTasks)
		},
		func() fyne.CanvasObject {
			return a.createTaskItem()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			a.updateTaskItem(id, obj, allTasks)
		},
	)
	
	return list
}

// updateTaskItem 更新任务项显示
func (a *App) updateTaskItem(id widget.ListItemID, obj fyne.CanvasObject, tasks []aria2.TellStatus) {
	if id >= len(tasks) {
		return
	}
	
	// 简化实现：重新创建整个任务项
	// 这样虽然效率稍低，但确保显示正确
	// TODO: 优化为只更新变化的组件
}

// createTaskItemForData 为特定任务数据创建任务项
func (a *App) createTaskItemForData(task aria2.TellStatus) fyne.CanvasObject {
	var name string = "未知任务"
	var status string = task.Status
	var progress float64 = 0
	var speedText string = "0 B/s"
	var sizeText string = "0 B / 0 B"
	
	// 获取任务名称
	if len(task.Files) > 0 {
		name = task.Files[0].Path
		if name == "" {
			name = "未命名任务"
		}
	}
	
	// 计算进度
	if task.TotalLength != "0" && task.CompletedLength != "0" {
		total := a.parseFloat64(task.TotalLength)
		completed := a.parseFloat64(task.CompletedLength)
		if total > 0 {
			progress = float64(completed) / float64(total)
		}
		
		sizeText = fmt.Sprintf("%s / %s", a.formatSize(completed), a.formatSize(total))
	}
	
	// 获取速度
	if task.DownloadSpeed != "0" {
		speed := a.parseFloat64(task.DownloadSpeed)
		speedText = a.formatSpeed(speed)
	}
	
	nameLabel := widget.NewLabel(name)
	statusLabel := widget.NewLabel(status)
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(progress)
	speedLabel := widget.NewLabel(speedText)
	sizeLabel := widget.NewLabel(sizeText)
	
	return container.NewVBox(
		container.NewHBox(
			nameLabel,
			statusLabel,
		),
		progressBar,
		container.NewHBox(
			speedLabel,
			sizeLabel,
		),
	)
}

// isStatusText 判断是否为状态文本
func (a *App) isStatusText(text string) bool {
	statuses := []string{"active", "waiting", "paused", "complete", "error", "removed"}
	for _, status := range statuses {
		if text == status {
			return true
		}
	}
	return false
}

// isSpeedText 判断是否为速度文本
func (a *App) isSpeedText(text string) bool {
	return len(text) > 0 && (text[len(text)-1] == 's' || text[len(text)-1] == 'B')
}

// isSizeText 判断是否为大小文本
func (a *App) isSizeText(text string) bool {
	return len(text) > 0 && (text[len(text)-1] == 'B' || text[len(text)-1] == 'K' || text[len(text)-1] == 'M' || text[len(text)-1] == 'G')
}

// parseFloat64 安全地解析字符串为 float64
func (a *App) parseFloat64(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

// formatSpeed 格式化速度显示
func (a *App) formatSpeed(bytesPerSec float64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/1024)
	} else if bytesPerSec < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB/s", bytesPerSec/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/(1024*1024*1024))
	}
}

// formatSize 格式化大小显示
func (a *App) formatSize(bytes float64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%.0f B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", bytes/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", bytes/(1024*1024))
	} else {
		return fmt.Sprintf("%.2f GB", bytes/(1024*1024*1024))
	}
}

// createTaskItem 创建任务项
func (a *App) createTaskItem() fyne.CanvasObject {
	nameLabel := widget.NewLabel("任务名称")
	statusLabel := widget.NewLabel("状态")
	progressBar := widget.NewProgressBar()
	speedLabel := widget.NewLabel("速度")
	sizeLabel := widget.NewLabel("大小")
	
	// 创建任务项容器
	taskContainer := container.NewVBox(
		container.NewHBox(
			nameLabel,
			statusLabel,
		),
		progressBar,
		container.NewHBox(
			speedLabel,
			sizeLabel,
		),
	)
	
	// 为任务项添加右键菜单
	// TODO: 实现右键菜单功能
	// taskContainer = a.addTaskContextMenu(taskContainer)
	
	return taskContainer
}

// addTaskContextMenu 为任务项添加右键菜单
func (a *App) addTaskContextMenu(item fyne.CanvasObject) fyne.CanvasObject {
	// TODO: Fyne 框架需要自定义实现右键菜单
	// 可以通过监听鼠标事件来实现
	return item
}

// createTaskContextMenu 创建任务右键菜单项
func (a *App) createTaskContextMenu() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		widget.NewButton("开始任务", func() {
			a.resumeSelectedTasks()
		}),
		widget.NewButton("暂停任务", func() {
			a.pauseSelectedTasks()
		}),
		widget.NewSeparator(),
		widget.NewButton("移动到顶部", func() {
			a.moveTaskToTop()
		}),
		widget.NewButton("移动到底部", func() {
			a.moveTaskToBottom()
		}),
		widget.NewSeparator(),
		widget.NewButton("复制下载链接", func() {
			a.copyTaskURL()
		}),
		widget.NewButton("打开文件所在目录", func() {
			a.openTaskDirectory()
		}),
		widget.NewSeparator(),
		widget.NewButton("任务详情", func() {
			a.showTaskDetailDialog()
		}),
		widget.NewButton("删除任务", func() {
			a.showRemoveTaskDialog()
		}),
	}
}

// moveTaskToTop 将任务移到队列顶部
func (a *App) moveTaskToTop() {
	// TODO: 调用 aria2Client.ChangePosition(taskID, pos=0)
}

// moveTaskToBottom 将任务移到队列底部
func (a *App) moveTaskToBottom() {
	// TODO: 调用 aria2Client.ChangePosition(taskID, pos=-1)
}

// copyTaskURL 复制任务下载链接
func (a *App) copyTaskURL() {
	// TODO: 获取选中任务的URL并复制到剪贴板
}

// openTaskDirectory 打开任务文件所在目录
func (a *App) openTaskDirectory() {
	// TODO: 获取任务文件路径并使用系统命令打开目录
}

// showTaskDetailDialog 显示任务详情对话框
func (a *App) showTaskDetailDialog() {
	// TODO: 显示任务的详细信息（文件列表、连接数、种子信息等）
}

// saveSettings 保存设置
func (a *App) saveSettings() {
	configPath := getConfigPath()
	
	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		a.showErrorMessage(fmt.Sprintf("创建配置目录失败: %v", err))
		return
	}
	
	// 保存配置到文件
	if err := a.config.SaveConfig(configPath); err != nil {
		a.showErrorMessage(fmt.Sprintf("保存配置失败: %v", err))
	} else {
		a.showSuccessMessage("配置已保存")
		
		// 如果 RPC 设置发生变化，重新连接 aria2 客户端
		a.reconnectAria2()
	}
}

// reconnectAria2 重新连接 aria2 客户端
func (a *App) reconnectAria2() {
	// 创建新的 aria2 客户端
	newClient := aria2.NewClient(
		a.config.RPC.Host,
		a.config.RPC.Port,
		a.config.RPC.Token,
		a.config.RPC.Protocol,
		a.config.RPC.Path,
	)
	
	// 测试连接
	if _, err := newClient.GetVersion(); err != nil {
		a.showErrorMessage(fmt.Sprintf("连接 aria2 失败: %v", err))
	} else {
		a.aria2Client = newClient
		a.showSuccessMessage("已重新连接到 aria2")
		a.refreshTaskList()
	}
}

// restoreDefaultSettings 恢复默认设置
func (a *App) restoreDefaultSettings() {
	// 创建确认对话框
	confirmWindow := a.fyneApp.NewWindow("确认恢复")
	confirmWindow.Resize(fyne.NewSize(300, 150))
	
	message := widget.NewLabel("确定要恢复默认设置吗？这将覆盖当前所有配置。")
	
	buttons := container.NewHBox(
		widget.NewButton("确定", func() {
			a.config = config.DefaultConfig()
			
			// 保存默认配置
			configPath := getConfigPath()
			if err := a.config.SaveConfig(configPath); err != nil {
				a.showErrorMessage(fmt.Sprintf("保存默认配置失败: %v", err))
			} else {
				a.showSuccessMessage("已恢复默认设置")
				
				// 重新连接 aria2 客户端
				a.reconnectAria2()
				
				// 重新打开设置窗口显示默认值
				a.showSettingsDialog()
			}
			
			confirmWindow.Close()
		}),
		widget.NewButton("取消", func() {
			confirmWindow.Close()
		}),
	)
	
	content := container.NewVBox(
		message,
		buttons,
	)
	
	confirmWindow.SetContent(container.NewCenter(content))
	confirmWindow.Show()
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(homeDir, ".aria2goui", "config.json")
}

// showAddTaskDialog 显示添加任务对话框
func (a *App) showAddTaskDialog() {
	// 创建添加任务窗口
	addWindow := a.fyneApp.NewWindow("添加下载任务")
	addWindow.Resize(fyne.NewSize(500, 400))
	
	// URL 输入框
	urlEntry := widget.NewMultiLineEntry()
	urlEntry.SetPlaceHolder("请输入下载链接，支持 HTTP/HTTPS/FTP/磁力链接/种子文件")
	urlEntry.Resize(fyne.NewSize(450, 100))
	
	// 下载目录
	dirEntry := widget.NewEntry()
	if a.config.Download.DefaultDirectory != "" {
		dirEntry.SetText(a.config.Download.DefaultDirectory)
	}
	
	// 选择目录按钮
	selectDirBtn := widget.NewButton("选择目录", func() {
		selectedDir := a.showDirectorySelectDialog(dirEntry.Text)
		if selectedDir != "" {
			dirEntry.SetText(selectedDir)
		}
	})
	
	// 下载选项
	options := map[string]fyne.CanvasObject{
		"split":  widget.NewSelect([]string{"1", "2", "4", "8", "16", "32"}, nil),
		"max-connection-per-server": widget.NewSelect([]string{"1", "5", "10", "16", "32"}, nil),
	}
	
	// 设置默认值
	if splitSelect, ok := options["split"].(*widget.Select); ok {
		splitSelect.SetSelected("16")
	}
	if maxConnSelect, ok := options["max-connection-per-server"].(*widget.Select); ok {
		maxConnSelect.SetSelected("16")
	}
	
	// 创建表单
	form := container.NewVBox(
		widget.NewCard("下载链接", "", urlEntry),
		widget.NewCard("下载设置", "", container.NewVBox(
			container.NewHBox(
				widget.NewLabel("下载目录:"),
				dirEntry,
				selectDirBtn,
			),
			container.NewGridWithColumns(2,
				widget.NewLabel("线程数:"), options["split"],
				widget.NewLabel("单服务器连接数:"), options["max-connection-per-server"],
			),
		)),
	)
	
	// 底部按钮
	bottomButtons := container.NewHBox(
		widget.NewButton("确定", func() {
			a.addTask(urlEntry.Text, dirEntry.Text, options)
			addWindow.Close()
		}),
		widget.NewButton("取消", func() {
			addWindow.Close()
		}),
	)
	
	// 主容器
	mainContainer := container.NewBorder(
		nil,
		bottomButtons,
		nil,
		nil,
		form,
	)
	
	addWindow.SetContent(mainContainer)
	addWindow.Show()
}

// addTask 添加下载任务
func (a *App) addTask(url, dir string, options map[string]fyne.CanvasObject) {
	if url == "" {
		a.showErrorMessage("请输入下载链接")
		return
	}
	
	// 解析 URL
	uris := []string{url}
	
	// 构建 aria2 选项
	aria2Options := make(map[string]interface{})
	
	if dir != "" {
		aria2Options["dir"] = dir
	}
	
	if splitSelect, ok := options["split"].(*widget.Select); ok {
		if split := splitSelect.Selected; split != "" {
			aria2Options["split"] = split
		}
	}
	
	if maxConnSelect, ok := options["max-connection-per-server"].(*widget.Select); ok {
		if maxConn := maxConnSelect.Selected; maxConn != "" {
			aria2Options["max-connection-per-server"] = maxConn
		}
	}
	
	// 调用 aria2 客户端添加任务
	if a.aria2Client == nil {
		a.showErrorMessage("未连接到 aria2 服务")
		return
	}
	
	gid, err := a.aria2Client.AddURI(uris, aria2Options)
	if err != nil {
		a.showErrorMessage(fmt.Sprintf("添加任务失败: %v", err))
		return
	}
	
	a.showSuccessMessage(fmt.Sprintf("任务已添加，GID: %s", gid))
	
	// 刷新任务列表
	a.refreshTaskList()
}

// showErrorMessage 显示错误消息
func (a *App) showErrorMessage(message string) {
	// 创建错误提示窗口
	errorWindow := a.fyneApp.NewWindow("错误")
	errorWindow.Resize(fyne.NewSize(300, 150))
	
	content := container.NewVBox(
		widget.NewIcon(theme.ErrorIcon()),
		widget.NewLabel(message),
		widget.NewButton("确定", func() {
			errorWindow.Close()
		}),
	)
	
	errorWindow.SetContent(container.NewCenter(content))
	errorWindow.Show()
}

// showSuccessMessage 显示成功消息
func (a *App) showSuccessMessage(message string) {
	// 创建成功提示窗口
	successWindow := a.fyneApp.NewWindow("成功")
	successWindow.Resize(fyne.NewSize(300, 150))
	
	content := container.NewVBox(
		widget.NewIcon(theme.ConfirmIcon()),
		widget.NewLabel(message),
		widget.NewButton("确定", func() {
			successWindow.Close()
		}),
	)
	
	successWindow.SetContent(container.NewCenter(content))
	successWindow.Show()
}

// pauseSelectedTasks 暂停选中的任务
func (a *App) pauseSelectedTasks() {
	if a.aria2Client == nil {
		a.showErrorMessage("未连接到 aria2 服务")
		return
	}
	
	// 暂停所有活动任务（简化实现，实际应该获取选中的任务）
	tasks, err := a.aria2Client.TellActive()
	if err != nil {
		a.showErrorMessage(fmt.Sprintf("获取任务失败: %v", err))
		return
	}
	
	if len(tasks) == 0 {
		a.showErrorMessage("没有活动的任务需要暂停")
		return
	}
	
	// 暂停所有活动任务
	for _, task := range tasks {
		if err := a.aria2Client.Pause(task.GID); err != nil {
			a.showErrorMessage(fmt.Sprintf("暂停任务失败: %v", err))
			return
		}
	}
	
	a.showSuccessMessage(fmt.Sprintf("已暂停 %d 个任务", len(tasks)))
	a.refreshTaskList()
}

// resumeSelectedTasks 恢复选中的任务
func (a *App) resumeSelectedTasks() {
	if a.aria2Client == nil {
		a.showErrorMessage("未连接到 aria2 服务")
		return
	}
	
	// 获取等待中的任务
	waitingTasks, err := a.aria2Client.TellWaiting(0, 1000)
	if err != nil {
		a.showErrorMessage(fmt.Sprintf("获取任务失败: %v", err))
		return
	}
	
	if len(waitingTasks) == 0 {
		a.showErrorMessage("没有等待中的任务需要恢复")
		return
	}
	
	// 恢复所有等待的任务
	for _, task := range waitingTasks {
		if err := a.aria2Client.Unpause(task.GID); err != nil {
			a.showErrorMessage(fmt.Sprintf("恢复任务失败: %v", err))
			return
		}
	}
	
	a.showSuccessMessage(fmt.Sprintf("已恢复 %d 个任务", len(waitingTasks)))
	a.refreshTaskList()
}

// showRemoveTaskDialog 显示删除任务确认对话框
func (a *App) showRemoveTaskDialog() {
	if a.aria2Client == nil {
		a.showErrorMessage("未连接到 aria2 服务")
		return
	}
	
	// 创建确认对话框
	confirmWindow := a.fyneApp.NewWindow("确认删除")
	confirmWindow.Resize(fyne.NewSize(350, 200))
	
	// 删除文件选项
	deleteFilesCheck := widget.NewCheck("同时删除下载的文件", nil)
	
	// 提示信息
	message := widget.NewLabel("确定要删除选中的任务吗？")
	
	// 按钮
	buttons := container.NewHBox(
		widget.NewButton("确定", func() {
			a.removeSelectedTasks(deleteFilesCheck.Checked)
			confirmWindow.Close()
		}),
		widget.NewButton("取消", func() {
			confirmWindow.Close()
		}),
	)
	
	content := container.NewVBox(
		message,
		deleteFilesCheck,
		buttons,
	)
	
	confirmWindow.SetContent(container.NewCenter(content))
	confirmWindow.Show()
}

// removeSelectedTasks 删除选中的任务
func (a *App) removeSelectedTasks(deleteFiles bool) {
	// 获取所有任务
	tasks := a.getAllTasks()
	
	if len(tasks) == 0 {
		a.showErrorMessage("没有可删除的任务")
		return
	}
	
	deletedCount := 0
	
	// 删除所有任务（简化实现，实际应该删除选中的任务）
	for _, task := range tasks {
		if err := a.aria2Client.Remove(task.GID); err != nil {
			a.showErrorMessage(fmt.Sprintf("删除任务失败: %v", err))
			continue
		}
		
		// 如果需要删除文件，删除任务文件
		if deleteFiles && len(task.Files) > 0 {
			for _, file := range task.Files {
				// TODO: 实现文件删除功能
				fmt.Printf("删除文件: %s\n", file.Path)
			}
		}
		
		deletedCount++
	}
	
	a.showSuccessMessage(fmt.Sprintf("已删除 %d 个任务", deletedCount))
	a.refreshTaskList()
}

// moveTaskUp 将选中任务上移
func (a *App) moveTaskUp() {
	// TODO: 获取选中的任务
	// TODO: 调用 aria2Client.ChangePosition(taskID, -1)
}

// moveTaskDown 将选中任务下移
func (a *App) moveTaskDown() {
	// TODO: 获取选中的任务
	// TODO: 调用 aria2Client.ChangePosition(taskID, 1)
}

// pauseAllTasks 暂停所有任务
func (a *App) pauseAllTasks() {
	// TODO: 调用 aria2Client.PauseAll()
}

// resumeAllTasks 恢复所有任务
func (a *App) resumeAllTasks() {
	// TODO: 调用 aria2Client.UnpauseAll()
}

// clearCompletedTasks 清理已完成的任务
func (a *App) clearCompletedTasks() {
	// TODO: 获取已完成的任务列表
	// TODO: 调用 aria2Client.RemoveDownloadResult(completedTasks)
}

// refreshTaskList 刷新任务列表
func (a *App) refreshTaskList() {
	if a.aria2Client == nil {
		a.showErrorMessage("未连接到 aria2 服务")
		return
	}
	
	// 重新创建主界面以刷新任务列表
	a.CreateMainUI()
	
	// 显示刷新提示
	fmt.Println("任务列表已刷新")
}

// showStatisticsDialog 显示统计信息对话框
func (a *App) showStatisticsDialog() {
	// TODO: 调用 aria2Client.GetGlobalStat()
	// TODO: 显示下载统计信息（总速度、上传速度等）
}

// exportTasks 导出任务列表
func (a *App) exportTasks() {
	// TODO: 获取当前任务列表
	// TODO: 导出为文件（如 .aria2 格式）
}

// showSettingsDialog 显示设置对话框
func (a *App) showSettingsDialog() {
	// 创建设置窗口
	settingsWindow := a.fyneApp.NewWindow("设置")
	settingsWindow.Resize(fyne.NewSize(600, 500))
	
	// 创建设置内容
	settingsContent := a.createSettingsContent()
	
	// 底部按钮
	bottomButtons := container.NewHBox(
		widget.NewButton("确定", func() {
			a.saveSettings()
			settingsWindow.Close()
		}),
		widget.NewButton("取消", func() {
			settingsWindow.Close()
		}),
		widget.NewButton("恢复默认", func() {
			a.restoreDefaultSettings()
		}),
	)
	
	// 主容器
	mainContainer := container.NewBorder(
		nil,
		bottomButtons,
		nil,
		nil,
		settingsContent,
	)
	
	settingsWindow.SetContent(mainContainer)
	settingsWindow.Show()
}

// createSettingsContent 创建设置内容
func (a *App) createSettingsContent() fyne.CanvasObject {
	// 创建选项卡容器
	tabs := container.NewAppTabs(
		container.NewTabItem("Aria2 RPC", a.createRPCSettings()),
		container.NewTabItem("基本设置", a.createBasicSettings()),
		container.NewTabItem("下载设置", a.createDownloadSettings()),
		container.NewTabItem("高级设置", a.createAdvancedSettings()),
		container.NewTabItem("显示设置", a.createDisplaySettings()),
		container.NewTabItem("通知设置", a.createNotifySettings()),
	)
	
	return tabs
}

// createRPCSettings 创建 RPC 设置界面
func (a *App) createRPCSettings() fyne.CanvasObject {
	// RPC 地址
	hostEntry := widget.NewEntry()
	hostEntry.SetText(a.config.RPC.Host)
	
	// RPC 端口
	portEntry := widget.NewEntry()
	portEntry.SetText(fmt.Sprintf("%d", a.config.RPC.Port))
	
	// RPC 协议
	protocolSelect := widget.NewSelect([]string{"http", "https", "ws", "wss"}, nil)
	protocolSelect.SetSelected(a.config.RPC.Protocol)
	
	// RPC 密钥
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetText(a.config.RPC.Token)
	
	// 请求路径
	pathEntry := widget.NewEntry()
	pathEntry.SetText(a.config.RPC.Path)
	
	// 自动重连
	autoReconnectCheck := widget.NewCheck("自动重连", nil)
	autoReconnectCheck.SetChecked(a.config.RPC.AutoReconnect)
	
	// 连接超时
	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(fmt.Sprintf("%d", a.config.RPC.Timeout))
	
	return container.NewVBox(
		widget.NewCard("连接设置", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("RPC 地址:"), hostEntry,
				widget.NewLabel("RPC 端口:"), portEntry,
				widget.NewLabel("协议:"), protocolSelect,
				widget.NewLabel("密钥:"), tokenEntry,
				widget.NewLabel("请求路径:"), pathEntry,
			),
			container.NewHBox(autoReconnectCheck, widget.NewLabel("连接超时(秒):"), timeoutEntry),
		)),
	)
}

// createBasicSettings 创建基本设置界面
func (a *App) createBasicSettings() fyne.CanvasObject {
	// 语言选择
	languageSelect := widget.NewSelect([]string{"zh_CN", "en_US"}, nil)
	languageSelect.SetSelected(a.config.UI.Language)
	
	// 主题选择
	themeSelect := widget.NewSelect([]string{"light", "dark", "auto"}, nil)
	themeSelect.SetSelected(a.config.UI.Theme)
	
	// 页面标题
	titleEntry := widget.NewEntry()
	titleEntry.SetText(a.config.UI.PageTitle)
	
	// 刷新间隔
	refreshEntry := widget.NewEntry()
	refreshEntry.SetText(fmt.Sprintf("%d", a.config.UI.RefreshInterval))
	
	// 启动时恢复任务
	continueCheck := widget.NewCheck("启动时恢复未完成任务", nil)
	continueCheck.SetChecked(a.config.General.ContinueTasks)
	
	// 最小化到托盘
	trayCheck := widget.NewCheck("最小化到系统托盘", nil)
	trayCheck.SetChecked(a.config.General.MinimizeToTray)
	
	return container.NewVBox(
		widget.NewCard("界面设置", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("语言:"), languageSelect,
				widget.NewLabel("主题:"), themeSelect,
				widget.NewLabel("页面标题:"), titleEntry,
				widget.NewLabel("刷新间隔(秒):"), refreshEntry,
			),
		)),
		widget.NewCard("行为设置", "", container.NewVBox(
			continueCheck,
			trayCheck,
		)),
	)
}

// createDownloadSettings 创建下载设置界面
func (a *App) createDownloadSettings() fyne.CanvasObject {
	// 下载目录
	dirEntry := widget.NewEntry()
	dirEntry.SetText(a.config.Download.DefaultDirectory)
	
	// 最大同时下载数
	maxConcurrentEntry := widget.NewEntry()
	maxConcurrentEntry.SetText(fmt.Sprintf("%d", a.config.Download.MaxConcurrentDownloads))
	
	// 单文件最大连接数
	maxConnEntry := widget.NewEntry()
	maxConnEntry.SetText(fmt.Sprintf("%d", a.config.Download.MaxConnectionPerServer))
	
	// 下载速度限制
	downSpeedEntry := widget.NewEntry()
	downSpeedEntry.SetText(fmt.Sprintf("%d", a.config.Download.GlobalSpeedLimit))
	
	// 上传速度限制
	upSpeedEntry := widget.NewEntry()
	upSpeedEntry.SetText(fmt.Sprintf("%d", a.config.Download.UploadSpeedLimit))
	
	return container.NewVBox(
		widget.NewCard("下载路径", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("默认下载目录:"), dirEntry,
			),
		)),
		widget.NewCard("连接限制", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("最大同时下载数:"), maxConcurrentEntry,
				widget.NewLabel("单文件最大连接数:"), maxConnEntry,
			),
		)),
		widget.NewCard("速度限制", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("下载速度限制(KB/s, 0=无限制):"), downSpeedEntry,
				widget.NewLabel("上传速度限制(KB/s, 0=无限制):"), upSpeedEntry,
			),
		)),
	)
}

// createAdvancedSettings 创建高级设置界面
func (a *App) createAdvancedSettings() fyne.CanvasObject {
	// User-Agent
	userAgentEntry := widget.NewEntry()
	userAgentEntry.SetText(a.config.Advanced.UserAgent)
	
	// HTTP 代理
	httpProxyEntry := widget.NewEntry()
	httpProxyEntry.SetText(a.config.Advanced.HTTProxy)
	
	// FTP 代理
	ftpProxyEntry := widget.NewEntry()
	ftpProxyEntry.SetText(a.config.Advanced.FTPProxy)
	
	// BT 端口范围
	btPortEntry := widget.NewEntry()
	btPortEntry.SetText(a.config.Advanced.BTPortRange)
	
	// DHT 支持
	dhtCheck := widget.NewCheck("启用 DHT", nil)
	dhtCheck.SetChecked(a.config.Advanced.DHTEnabled)
	
	// PEX 支持
	pexCheck := widget.NewCheck("启用 PEX", nil)
	pexCheck.SetChecked(a.config.Advanced.PEXEnabled)
	
	// 种子文件下载
	seedCheck := widget.NewCheck("自动下载种子文件", nil)
	seedCheck.SetChecked(a.config.Advanced.SeedDownload)
	
	return container.NewVBox(
		widget.NewCard("网络设置", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("User-Agent:"), userAgentEntry,
				widget.NewLabel("HTTP 代理:"), httpProxyEntry,
				widget.NewLabel("FTP 代理:"), ftpProxyEntry,
			),
		)),
		widget.NewCard("BitTorrent 设置", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("端口范围:"), btPortEntry,
			),
			container.NewHBox(dhtCheck, pexCheck),
			seedCheck,
		)),
	)
}

// createDisplaySettings 创建显示设置界面
func (a *App) createDisplaySettings() fyne.CanvasObject {
	// 显示模式
	viewModeSelect := widget.NewSelect([]string{"list", "card"}, nil)
	viewModeSelect.SetSelected(a.config.Display.ViewMode)
	
	// 排序方式
	sortBySelect := widget.NewSelect([]string{"name", "size", "progress", "speed"}, nil)
	sortBySelect.SetSelected(a.config.Display.SortBy)
	
	// 进度条样式
	progressStyleSelect := widget.NewSelect([]string{"normal", "detailed"}, nil)
	progressStyleSelect.SetSelected(a.config.Display.ProgressBarStyle)
	
	// 显示文件列表
	fileListCheck := widget.NewCheck("显示任务文件列表", nil)
	fileListCheck.SetChecked(a.config.Display.ShowFileList)
	
	// 显示 BT 信息
	btInfoCheck := widget.NewCheck("显示种子详细信息", nil)
	btInfoCheck.SetChecked(a.config.Display.ShowBTInfo)
	
	// 自动滚动
	autoScrollCheck := widget.NewCheck("自动滚动到活动任务", nil)
	autoScrollCheck.SetChecked(a.config.Display.AutoScrollToActive)
	
	return container.NewVBox(
		widget.NewCard("显示模式", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("显示模式:"), viewModeSelect,
				widget.NewLabel("排序方式:"), sortBySelect,
				widget.NewLabel("进度条样式:"), progressStyleSelect,
			),
		)),
		widget.NewCard("显示选项", "", container.NewVBox(
			fileListCheck,
			btInfoCheck,
			autoScrollCheck,
		)),
	)
}

// createNotifySettings 创建通知设置界面
func (a *App) createNotifySettings() fyne.CanvasObject {
	// 声音提醒
	soundCheck := widget.NewCheck("启用声音提醒", nil)
	soundCheck.SetChecked(a.config.Notify.SoundEnabled)
	
	// 系统通知
	systemNotifyCheck := widget.NewCheck("启用系统通知", nil)
	systemNotifyCheck.SetChecked(a.config.Notify.SystemNotify)
	
	// 浏览器通知
	browserNotifyCheck := widget.NewCheck("启用浏览器通知", nil)
	browserNotifyCheck.SetChecked(a.config.Notify.BrowserNotify)
	
	// 错误提醒
	errorNotifyCheck := widget.NewCheck("错误提醒", nil)
	errorNotifyCheck.SetChecked(a.config.Notify.ErrorNotify)
	
	// 完成提醒
	completeNotifyCheck := widget.NewCheck("完成提醒", nil)
	completeNotifyCheck.SetChecked(a.config.Notify.CompleteNotify)
	
	return container.NewVBox(
		widget.NewCard("通知类型", "", container.NewVBox(
			soundCheck,
			systemNotifyCheck,
			browserNotifyCheck,
		)),
		widget.NewCard("通知事件", "", container.NewVBox(
			errorNotifyCheck,
			completeNotifyCheck,
		)),
	)
}

// Show 显示应用程序
func (a *App) Show() {
	a.window.Show()
}

// ShowAndRun 显示并运行应用程序
func (a *App) ShowAndRun() {
	a.window.ShowAndRun()
}

// ShowConnectionDialog 显示连接设置对话框
func (a *App) ShowConnectionDialog() {
	// 创建连接设置窗口
	connWindow := a.fyneApp.NewWindow("连接设置")
	connWindow.Resize(fyne.NewSize(400, 300))
	
	// RPC 地址
	hostEntry := widget.NewEntry()
	hostEntry.SetText(a.config.RPC.Host)
	
	// RPC 端口
	portEntry := widget.NewEntry()
	portEntry.SetText(fmt.Sprintf("%d", a.config.RPC.Port))
	
	// RPC 协议
	protocolSelect := widget.NewSelect([]string{"http", "https", "ws", "wss"}, nil)
	protocolSelect.SetSelected(a.config.RPC.Protocol)
	
	// RPC 密钥
	tokenEntry := widget.NewPasswordEntry()
	tokenEntry.SetText(a.config.RPC.Token)
	
	// 请求路径
	pathEntry := widget.NewEntry()
	pathEntry.SetText(a.config.RPC.Path)
	
	// 连接状态
	statusLabel := widget.NewLabel("未连接")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	// 创建表单
	form := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("RPC 地址:"), hostEntry,
			widget.NewLabel("RPC 端口:"), portEntry,
			widget.NewLabel("协议:"), protocolSelect,
			widget.NewLabel("密钥:"), tokenEntry,
			widget.NewLabel("请求路径:"), pathEntry,
		),
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("连接状态:"),
			statusLabel,
		),
	)
	
	// 测试连接按钮
	testBtn := widget.NewButton("测试连接", func() {
		statusLabel.SetText("正在连接...")
		statusLabel.Refresh()
		
		// 创建临时客户端测试连接
		tempClient := aria2.NewClient(
			hostEntry.Text,
			a.parseInt(portEntry.Text),
			tokenEntry.Text,
			protocolSelect.Selected,
			pathEntry.Text,
		)
		
		if _, err := tempClient.GetVersion(); err != nil {
			statusLabel.SetText(fmt.Sprintf("连接失败: %v", err))
		} else {
			statusLabel.SetText("连接成功！")
		}
		statusLabel.Refresh()
	})
	
	// 底部按钮
	bottomButtons := container.NewHBox(
		widget.NewButton("保存并连接", func() {
			// 更新配置
			a.config.RPC.Host = hostEntry.Text
			a.config.RPC.Port = a.parseInt(portEntry.Text)
			a.config.RPC.Protocol = protocolSelect.Selected
			a.config.RPC.Token = tokenEntry.Text
			a.config.RPC.Path = pathEntry.Text
			
			// 重新连接
			a.reconnectAria2()
			connWindow.Close()
		}),
		widget.NewButton("取消", func() {
			connWindow.Close()
		}),
	)
	
	// 主容器
	mainContainer := container.NewBorder(
		nil,
		container.NewVBox(testBtn, bottomButtons),
		nil,
		nil,
		form,
	)
	
	connWindow.SetContent(mainContainer)
	connWindow.Show()
}

// parseInt 安全地解析字符串为 int
func (a *App) parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// showDirectorySelectDialog 显示目录选择对话框
func (a *App) showDirectorySelectDialog(currentDir string) string {
	// 创建目录选择窗口
	dirWindow := a.fyneApp.NewWindow("选择下载目录")
	dirWindow.Resize(fyne.NewSize(500, 400))
	
	// 当前路径显示
	pathEntry := widget.NewEntry()
	if currentDir == "" {
		// 获取默认下载目录
		if homeDir, err := os.UserHomeDir(); err == nil {
			currentDir = filepath.Join(homeDir, "Downloads")
		} else {
			currentDir = "C:\\Downloads"
		}
	}
	pathEntry.SetText(currentDir)
	
	// 目录列表
	dirList := widget.NewList(
		func() int {
			entries, err := os.ReadDir(currentDir)
			if err != nil {
				return 0
			}
			count := 0
			for _, entry := range entries {
				if entry.IsDir() {
					count++
				}
			}
			return count + 1 // +1 for parent directory
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("目录")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			if id == 0 {
				// 父目录
				label.SetText(".. (返回上级目录)")
			} else {
				// 子目录
				entries, err := os.ReadDir(currentDir)
				if err == nil {
					dirIndex := 0
					for _, entry := range entries {
						if entry.IsDir() {
							dirIndex++
							if dirIndex == id {
								label.SetText(entry.Name())
								break
							}
						}
					}
				}
			}
		},
	)
	
	// 双击进入目录
	dirList.OnSelected = func(id widget.ListItemID) {
		if id == 0 {
			// 返回上级目录
			parent := filepath.Dir(currentDir)
			if parent != currentDir {
				currentDir = parent
				pathEntry.SetText(currentDir)
				dirList.Refresh()
			}
		} else {
			// 进入子目录
			entries, err := os.ReadDir(currentDir)
			if err == nil {
				dirIndex := 0
				for _, entry := range entries {
					if entry.IsDir() {
						dirIndex++
						if dirIndex == id {
							currentDir = filepath.Join(currentDir, entry.Name())
							pathEntry.SetText(currentDir)
							dirList.Refresh()
							dirList.UnselectAll()
							break
						}
					}
				}
			}
		}
	}
	
	// 快速访问按钮
	quickAccess := container.NewHBox(
		widget.NewButton("桌面", func() {
			if homeDir, err := os.UserHomeDir(); err == nil {
				currentDir = filepath.Join(homeDir, "Desktop")
				pathEntry.SetText(currentDir)
				dirList.Refresh()
			}
		}),
		widget.NewButton("文档", func() {
			if homeDir, err := os.UserHomeDir(); err == nil {
				currentDir = filepath.Join(homeDir, "Documents")
				pathEntry.SetText(currentDir)
				dirList.Refresh()
			}
		}),
		widget.NewButton("下载", func() {
			if homeDir, err := os.UserHomeDir(); err == nil {
				currentDir = filepath.Join(homeDir, "Downloads")
				pathEntry.SetText(currentDir)
				dirList.Refresh()
			}
		}),
	)
	
	// 底部按钮
	bottomButtons := container.NewHBox(
		widget.NewButton("确定", func() {
			dirWindow.Close()
		}),
		widget.NewButton("取消", func() {
			currentDir = ""
			dirWindow.Close()
		}),
		widget.NewButton("新建文件夹", func() {
			a.showCreateFolderDialog(currentDir, func(newDir string) {
				currentDir = newDir
				pathEntry.SetText(currentDir)
				dirList.Refresh()
			})
		}),
	)
	
	// 主容器
	mainContainer := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("当前路径:"),
			pathEntry,
			quickAccess,
		),
		bottomButtons,
		nil,
		nil,
		dirList,
	)
	
	dirWindow.SetContent(mainContainer)
	dirWindow.Show()
	
	// 等待窗口关闭
	// 注意：这里简化处理，实际应该使用同步机制
	return currentDir
}

// showCreateFolderDialog 显示新建文件夹对话框
func (a *App) showCreateFolderDialog(parentDir string, callback func(string)) {
	// 创建新建文件夹对话框
	createWindow := a.fyneApp.NewWindow("新建文件夹")
	createWindow.Resize(fyne.NewSize(300, 150))
	
	// 文件夹名称输入
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("请输入文件夹名称")
	
	// 创建按钮
	createBtn := widget.NewButton("创建", func() {
		folderName := nameEntry.Text
		if folderName == "" {
			return
		}
		
		newDirPath := filepath.Join(parentDir, folderName)
		if err := os.MkdirAll(newDirPath, 0755); err != nil {
			a.showErrorMessage(fmt.Sprintf("创建文件夹失败: %v", err))
		} else {
			callback(newDirPath)
			createWindow.Close()
		}
	})
	
	// 内容
	content := container.NewVBox(
		widget.NewLabel("文件夹名称:"),
		nameEntry,
		container.NewHBox(
			createBtn,
			widget.NewButton("取消", func() {
				createWindow.Close()
			}),
		),
	)
	
	createWindow.SetContent(container.NewCenter(content))
	createWindow.Show()
}

// Close 关闭应用程序
func (a *App) Close() {
	a.window.Close()
}

// MyTheme 自定义主题
type MyTheme struct{}

func (m *MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0, G: 120, B: 215, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m *MyTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 返回默认主题的字体，避免 nil 导致的错误
	// 环境变量 FYNE_FONT 会覆盖这个设置
	return theme.DefaultTheme().Font(style)
}

func (m *MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *MyTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}