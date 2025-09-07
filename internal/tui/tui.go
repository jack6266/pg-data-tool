package tui

import (
	"fmt"
	"strconv"
	"strings"

	"pg-data-tool/internal/config"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TUIApp TUI应用程序结构
type TUIApp struct {
	app        *tview.Application
	pages      *tview.Pages
	config     *config.Config
	onComplete func(*config.Config, string) // 完成回调，返回配置和操作类型
}

// NewTUIApp 创建新的TUI应用
func NewTUIApp() *TUIApp {
	app := tview.NewApplication()
	app.EnableMouse(true) // 启用鼠标支持

	// 设置高对比度主题以增强光标和界面可见性
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack  // 黑色背景
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlue    // 蓝色对比背景
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorRed // 红色高对比背景
	tview.Styles.BorderColor = tcell.ColorYellow              // 黄色边框
	tview.Styles.TitleColor = tcell.ColorAqua                 // 青色标题
	tview.Styles.GraphicsColor = tcell.ColorLime              // 亮绿色图形
	tview.Styles.PrimaryTextColor = tcell.ColorWhite          // 白色主文本
	tview.Styles.SecondaryTextColor = tcell.ColorYellow       // 黄色次要文本
	tview.Styles.TertiaryTextColor = tcell.ColorLime          // 亮绿色第三级文本
	tview.Styles.InverseTextColor = tcell.ColorBlack          // 反色文本
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorAqua // 青色对比次要文本

	return &TUIApp{
		app:    app,
		pages:  tview.NewPages(),
		config: config.NewConfig(),
	}
}

// SetCompleteCallback 设置完成回调
func (t *TUIApp) SetCompleteCallback(callback func(*config.Config, string)) {
	t.onComplete = callback
}

// Run 运行TUI应用
func (t *TUIApp) Run() error {
	// 创建主菜单
	t.createMainMenu()

	// 设置根组件
	t.app.SetRoot(t.pages, true)

	// 运行应用
	return t.app.Run()
}

// createMainMenu 创建主菜单
func (t *TUIApp) createMainMenu() {
	// 创建主菜单标题
	title := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("🗃️  PostgreSQL 数据工具 (pg-data-tool)").
		SetTextColor(tview.Styles.PrimaryTextColor)
	title.SetBorder(true).SetTitle("欢迎")

	// 创建操作选择列表
	operationList := tview.NewList().
		ShowSecondaryText(false).
		SetSelectedTextColor(tview.Styles.PrimitiveBackgroundColor).
		SetSelectedBackgroundColor(tview.Styles.PrimaryTextColor)

	// 添加操作选项
	operations := []struct {
		main, desc string
		action     func()
	}{
		{"📤 逻辑备份", "使用pg_dump进行逻辑备份", func() { t.showBackupConfig("logical") }},
		{"🔥 物理热备", "使用pg_basebackup进行物理热备", func() { t.showBackupConfig("physical") }},
		{"📈 增量备份", "基于WAL日志的增量备份", func() { t.showBackupConfig("incremental") }},
		{"📥 逻辑还原", "从逻辑备份文件还原数据库", func() { t.showRestoreConfig("logical") }},
		{"🔄 物理还原", "从物理备份还原数据目录", func() { t.showRestoreConfig("physical") }},
		{"⚡ 增量恢复", "应用增量备份进行时间点恢复", func() { t.showRestoreConfig("incremental") }},
		{"🔀 数据传输", "在不同数据库间传输数据", func() { t.showTransferConfig() }},
		{"❌ 退出", "退出程序", func() { t.app.Stop() }},
	}

	for i, op := range operations {
		operationList.AddItem(op.main, op.desc, rune('1'+i), op.action)
	}

	operationList.SetBorder(true).SetTitle("选择操作类型 (支持鼠标点击)")

	// 创建说明文本
	infoText := tview.NewTextView().
		SetText(`🎯 操作说明:
• 使用 ↑↓ 键或鼠标选择操作
• 按 Enter 或点击确认选择
• 按 Esc 返回上级菜单
• 按 Ctrl+C 退出程序

💡 提示:
• 逻辑备份：适合小到中型数据库，支持跨版本
• 物理热备：适合大型数据库，速度快
• 增量备份：适合连续数据保护，节省空间`).
		SetTextColor(tview.Styles.SecondaryTextColor)
	infoText.SetBorder(true).SetTitle("帮助信息")

	// 创建布局
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 3, 0, false).
		AddItem(tview.NewFlex().
			AddItem(operationList, 0, 1, true).
			AddItem(infoText, 50, 0, false), 0, 1, true)

	// 设置快捷键
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			t.app.Stop()
			return nil
		}
		return event
	})

	t.pages.AddPage("main", flex, true, true)
}

// showBackupConfig 显示备份配置界面
func (t *TUIApp) showBackupConfig(backupType string) {
	t.config.BackupType = backupType

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(fmt.Sprintf("📤 %s 配置 (Tab切换字段, Esc返回)", getBackupTypeName(backupType)))

	// 应用表单样式
	t.applyFormStyle(form)

	// 通用连接配置
	t.addConnectionFields(form)

	// 根据备份类型添加特定配置
	switch backupType {
	case "logical":
		t.addLogicalBackupFields(form)
	case "physical":
		t.addPhysicalBackupFields(form)
	case "incremental":
		t.addIncrementalBackupFields(form)
	}

	// 添加操作按钮
	form.AddButton("开始备份", func() {
		if t.validateBackupConfig(backupType) {
			t.executeAction("backup")
		}
	})
	form.AddButton("返回", func() {
		t.pages.SwitchToPage("main")
	})

	t.pages.AddPage("backup_config", form, true, true)
	t.pages.SwitchToPage("backup_config")
}

// showRestoreConfig 显示还原配置界面
func (t *TUIApp) showRestoreConfig(restoreType string) {
	t.config.RestoreType = restoreType

	form := tview.NewForm()
	form.SetBorder(true).SetTitle(fmt.Sprintf("📥 %s 配置 (Tab切换字段, Esc返回)", getRestoreTypeName(restoreType)))

	// 应用表单样式
	t.applyFormStyle(form)

	// 通用连接配置
	t.addConnectionFields(form)

	// 根据还原类型添加特定配置
	switch restoreType {
	case "logical":
		t.addLogicalRestoreFields(form)
	case "physical":
		t.addPhysicalRestoreFields(form)
	case "incremental":
		t.addIncrementalRestoreFields(form)
	}

	// 添加操作按钮
	form.AddButton("开始还原", func() {
		if t.validateRestoreConfig(restoreType) {
			t.executeAction("restore")
		}
	})
	form.AddButton("返回", func() {
		t.pages.SwitchToPage("main")
	})

	t.pages.AddPage("restore_config", form, true, true)
	t.pages.SwitchToPage("restore_config")
}

// showTransferConfig 显示传输配置界面
func (t *TUIApp) showTransferConfig() {
	form := tview.NewForm()
	form.SetBorder(true).SetTitle("🔀 数据传输配置 (Tab切换字段, Esc返回)")

	// 应用表单样式
	t.applyFormStyle(form)

	// 源数据库配置
	form.AddTextView("源数据库配置:", "", 0, 1, true, false)
	t.addConnectionFields(form)

	// 目标数据库配置
	form.AddTextView("目标数据库配置:", "", 0, 1, true, false)
	form.AddInputField("目标主机", t.config.TargetHost, 50, nil, func(text string) {
		t.config.TargetHost = text
	})
	form.AddInputField("目标端口", t.config.TargetPort, 10, nil, func(text string) {
		t.config.TargetPort = text
	})
	form.AddInputField("目标用户", t.config.TargetUser, 20, nil, func(text string) {
		t.config.TargetUser = text
	})
	form.AddPasswordField("目标密码", t.config.TargetPassword, 20, '*', func(text string) {
		t.config.TargetPassword = text
	})

	// 传输选项
	form.AddCheckbox("传输所有数据库", t.config.TransferAll, func(checked bool) {
		t.config.TransferAll = checked
	})
	form.AddInputField("指定数据库", t.config.DBName, 30, nil, func(text string) {
		t.config.DBName = text
	})
	form.AddCheckbox("包含大对象", t.config.IncludeBlobs, func(checked bool) {
		t.config.IncludeBlobs = checked
	})
	form.AddCheckbox("包含索引", t.config.IncludeIndexes, func(checked bool) {
		t.config.IncludeIndexes = checked
	})
	form.AddCheckbox("包含权限", t.config.IncludePrivileges, func(checked bool) {
		t.config.IncludePrivileges = checked
	})

	// 添加操作按钮
	form.AddButton("开始传输", func() {
		t.executeAction("transfer")
	})
	form.AddButton("返回", func() {
		t.pages.SwitchToPage("main")
	})

	t.pages.AddPage("transfer_config", form, true, true)
	t.pages.SwitchToPage("transfer_config")
}

// addConnectionFields 添加数据库连接字段
func (t *TUIApp) addConnectionFields(form *tview.Form) {
	form.AddInputField("主机地址", t.config.Host, 50, nil, func(text string) {
		t.config.Host = text
	})
	form.AddInputField("端口", t.config.Port, 10, nil, func(text string) {
		t.config.Port = text
	})
	form.AddInputField("用户名", t.config.User, 20, nil, func(text string) {
		t.config.User = text
	})
	form.AddPasswordField("密码", t.config.Password, 20, '*', func(text string) {
		t.config.Password = text
	})
}

// applyFormStyle 应用表单样式以增强光标和输入框可见性
func (t *TUIApp) applyFormStyle(form *tview.Form) {
	// 设置高对比度的表单样式以增强光标可见性
	form.SetFieldBackgroundColor(tcell.ColorBlack) // 黑色背景使光标更明显
	form.SetFieldTextColor(tcell.ColorWhite)       // 白色文字
	form.SetLabelColor(tcell.ColorLime)            // 亮绿色标签，更突出
	form.SetButtonBackgroundColor(tcell.ColorBlue) // 蓝色按钮背景
	form.SetButtonTextColor(tcell.ColorWhite)      // 白色按钮文字

	// 设置边框颜色为亮色
	form.SetBorderColor(tcell.ColorYellow) // 黄色边框更醒目
	form.SetTitleColor(tcell.ColorAqua)    // 青色标题

	// 设置焦点样式，让当前输入框更明显
	form.SetFocus(0) // 默认焦点在第一个字段

	// 添加输入事件处理以增强交互体验
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// 当按下Tab键时，确保焦点切换时有视觉反馈
		if event.Key() == tcell.KeyTab {
			// Tab键会自动处理焦点切换，我们只需要确保样式正确
			return event
		}
		return event
	})
}

// addLogicalBackupFields 添加逻辑备份字段
func (t *TUIApp) addLogicalBackupFields(form *tview.Form) {
	form.AddCheckbox("备份所有数据库", t.config.BackupAll, func(checked bool) {
		t.config.BackupAll = checked
	})
	form.AddInputField("数据库名称", t.config.DBName, 30, nil, func(text string) {
		t.config.DBName = text
	})
	form.AddDropDown("备份格式", []string{"plain", "custom", "directory", "tar"}, 1, func(option string, optionIndex int) {
		formats := []string{"plain", "custom", "directory", "tar"}
		t.config.Format = formats[optionIndex]
	})
	form.AddInputField("输出路径", t.config.File, 80, nil, func(text string) {
		t.config.SetFile(text)
	})
	form.AddCheckbox("使用并行处理", t.config.UseParallel, func(checked bool) {
		t.config.UseParallel = checked
	})
	form.AddInputField("并行作业数", strconv.Itoa(t.config.ParallelJobs), 10, nil, func(text string) {
		if jobs, err := strconv.Atoi(text); err == nil {
			t.config.ParallelJobs = jobs
		}
	})
}

// addPhysicalBackupFields 添加物理备份字段
func (t *TUIApp) addPhysicalBackupFields(form *tview.Form) {
	form.AddInputField("备份目录", t.config.File, 80, nil, func(text string) {
		t.config.SetFile(text)
	})
	form.AddInputField("WAL归档目录", t.config.WALArchiveDir, 80, nil, func(text string) {
		t.config.SetWALArchiveDir(text)
	})
	form.AddInputField("最大WAL大小", t.config.MaxWalSize, 20, nil, func(text string) {
		t.config.MaxWalSize = text
	})
}

// addIncrementalBackupFields 添加增量备份字段
func (t *TUIApp) addIncrementalBackupFields(form *tview.Form) {
	form.AddInputField("输出目录", t.config.File, 80, nil, func(text string) {
		t.config.SetFile(text)
	})
	form.AddInputField("基础备份路径", t.config.BaseBackupPath, 80, nil, func(text string) {
		t.config.SetBaseBackupPath(text)
	})
	form.AddInputField("检查点LSN", t.config.CheckpointLSN, 30, nil, func(text string) {
		t.config.CheckpointLSN = text
	})
}

// addLogicalRestoreFields 添加逻辑还原字段
func (t *TUIApp) addLogicalRestoreFields(form *tview.Form) {
	form.AddCheckbox("还原所有数据库", t.config.RestoreAll, func(checked bool) {
		t.config.RestoreAll = checked
	})
	form.AddInputField("数据库名称", t.config.DBName, 30, nil, func(text string) {
		t.config.DBName = text
	})
	form.AddInputField("备份文件路径", t.config.File, 80, nil, func(text string) {
		t.config.SetFile(text)
	})
	form.AddCheckbox("自动创建数据库", t.config.AutoCreateDB, func(checked bool) {
		t.config.AutoCreateDB = checked
	})
	form.AddCheckbox("删除并重新创建数据库", t.config.DropDatabase, func(checked bool) {
		t.config.DropDatabase = checked
	})
	form.AddCheckbox("清理数据库结构", t.config.CleanStructure, func(checked bool) {
		t.config.CleanStructure = checked
	})
	form.AddCheckbox("清理数据", t.config.CleanData, func(checked bool) {
		t.config.CleanData = checked
	})
}

// addPhysicalRestoreFields 添加物理还原字段
func (t *TUIApp) addPhysicalRestoreFields(form *tview.Form) {
	form.AddInputField("数据目录路径", t.config.File, 80, nil, func(text string) {
		t.config.SetFile(text)
	})
	form.AddInputField("备份文件路径", t.config.BaseBackupPath, 80, nil, func(text string) {
		t.config.SetBaseBackupPath(text)
	})
}

// addIncrementalRestoreFields 添加增量还原字段
func (t *TUIApp) addIncrementalRestoreFields(form *tview.Form) {
	form.AddInputField("数据目录路径", t.config.File, 80, nil, func(text string) {
		t.config.SetFile(text)
	})
	form.AddInputField("基础备份路径", t.config.BaseBackupPath, 80, nil, func(text string) {
		t.config.SetBaseBackupPath(text)
	})
	form.AddInputField("WAL归档目录", t.config.WALArchiveDir, 80, nil, func(text string) {
		t.config.SetWALArchiveDir(text)
	})
	form.AddInputField("目标时间", t.config.TargetTime, 30, nil, func(text string) {
		t.config.TargetTime = text
	})
}

// validateBackupConfig 验证备份配置
func (t *TUIApp) validateBackupConfig(backupType string) bool {
	if strings.TrimSpace(t.config.Host) == "" {
		t.showErrorDialog("错误", "请输入主机地址")
		return false
	}
	if strings.TrimSpace(t.config.User) == "" {
		t.showErrorDialog("错误", "请输入用户名")
		return false
	}
	if strings.TrimSpace(t.config.File) == "" {
		t.showErrorDialog("错误", "请输入输出路径")
		return false
	}

	switch backupType {
	case "logical":
		if !t.config.BackupAll && strings.TrimSpace(t.config.DBName) == "" {
			t.showErrorDialog("错误", "请输入数据库名称或选择备份所有数据库")
			return false
		}
	case "incremental":
		if strings.TrimSpace(t.config.BaseBackupPath) == "" {
			t.showErrorDialog("错误", "增量备份需要指定基础备份路径")
			return false
		}
	}
	return true
}

// validateRestoreConfig 验证还原配置
func (t *TUIApp) validateRestoreConfig(restoreType string) bool {
	if strings.TrimSpace(t.config.Host) == "" {
		t.showErrorDialog("错误", "请输入主机地址")
		return false
	}
	if strings.TrimSpace(t.config.User) == "" {
		t.showErrorDialog("错误", "请输入用户名")
		return false
	}

	switch restoreType {
	case "logical":
		if !t.config.RestoreAll && strings.TrimSpace(t.config.DBName) == "" {
			t.showErrorDialog("错误", "请输入数据库名称或选择还原所有数据库")
			return false
		}
		if strings.TrimSpace(t.config.File) == "" {
			t.showErrorDialog("错误", "请输入备份文件路径")
			return false
		}
	case "physical", "incremental":
		if strings.TrimSpace(t.config.File) == "" {
			t.showErrorDialog("错误", "请输入数据目录路径")
			return false
		}
		if strings.TrimSpace(t.config.BaseBackupPath) == "" {
			t.showErrorDialog("错误", "请输入备份文件路径")
			return false
		}
	}

	if restoreType == "incremental" && strings.TrimSpace(t.config.WALArchiveDir) == "" {
		t.showErrorDialog("错误", "增量恢复需要指定WAL归档目录")
		return false
	}

	return true
}

// showErrorDialog 显示错误对话框
func (t *TUIApp) showErrorDialog(title, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"确定"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			t.pages.RemovePage("error")
		})
	modal.SetBorder(true).SetTitle(title)

	t.pages.AddPage("error", modal, false, true)
}

// executeAction 执行操作
func (t *TUIApp) executeAction(actionType string) {
	if t.onComplete != nil {
		t.app.Stop()
		t.onComplete(t.config, actionType)
	}
}

// 辅助函数
func getBackupTypeName(backupType string) string {
	switch backupType {
	case "logical":
		return "逻辑备份"
	case "physical":
		return "物理热备"
	case "incremental":
		return "增量备份"
	default:
		return "备份"
	}
}

func getRestoreTypeName(restoreType string) string {
	switch restoreType {
	case "logical":
		return "逻辑还原"
	case "physical":
		return "物理还原"
	case "incremental":
		return "增量恢复"
	default:
		return "还原"
	}
}
