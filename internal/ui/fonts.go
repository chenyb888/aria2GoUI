package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// 中文字体主题
type ChineseFontTheme struct {
	fyne.Theme
}

// NewChineseFontTheme 创建支持中文的主题
func NewChineseFontTheme() fyne.Theme {
	return &ChineseFontTheme{Theme: theme.DefaultTheme()}
}

func (t *ChineseFontTheme) Font(style fyne.TextStyle) fyne.Resource {
	// 返回默认主题的字体，避免 nil 导致的错误
	// 环境变量 FYNE_FONT 会覆盖这个设置
	return t.Theme.Font(style)
}

// 为了更好的中文支持，我们可以添加字体资源
// 这里提供几个常见的中文字体选项，宋体优先
var chineseFonts = []string{
	"SimSun",
	"SimHei",
	"Microsoft YaHei",
	"KaiTi",
	"FangSong",
}