package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
)

// Custom dark theme with blue tint
type customTheme struct{}

func (t *customTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 20, G: 22, B: 35, A: 255} // Dark blue background
	case theme.ColorNameButton:
		return color.NRGBA{R: 45, G: 50, B: 80, A: 255}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 35, G: 40, B: 60, A: 255}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 30, G: 35, B: 55, A: 255}
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 25, G: 28, B: 45, A: 255}
	case theme.ColorNameMenuBackground:
		return color.NRGBA{R: 30, G: 35, B: 55, A: 255}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 60, G: 65, B: 90, A: 255}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 138, G: 43, B: 226, A: 255} // Purple
	case theme.ColorNameForeground:
		return color.NRGBA{R: 220, G: 220, B: 230, A: 255}
	}
	return theme.DarkTheme().Color(name, variant)
}

func (t *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DarkTheme().Font(style)
}

func (t *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DarkTheme().Icon(name)
}

func (t *customTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DarkTheme().Size(name)
}

// Batch represents an upload batch
type Batch struct {
	ID        string
	Folder    string
	Files     []string
	FileSizes map[string]int64
	TotalSize int64
	Status    string
	StartTime time.Time
	LastTime  time.Time
}

// Config represents app settings
type Config struct {
	VideoEnabled      bool   `json:"video_enabled"`
	ImageEnabled      bool   `json:"image_enabled"`
	AudioEnabled      bool   `json:"audio_enabled"`
	DocEnabled        bool   `json:"doc_enabled"`
	ArchiveEnabled    bool   `json:"archive_enabled"`
	CustomExts        string `json:"custom_exts"`
	MonitorSubdirs    bool   `json:"monitor_subdirs"`
	CompletionTimeout int    `json:"completion_timeout"`
	NotifyOnStart     bool   `json:"notify_on_start"`
	NotifyOnComplete  bool   `json:"notify_on_complete"`
	SoundEnabled      bool   `json:"sound_enabled"`
	SaveHistory       bool   `json:"save_history"`
}

var tempFilePatterns = []string{".tmp", ".temp", ".part", ".partial", ".crdownload", "~$", ".swp", ".lock"}

var (
	monitorPath   string
	isMonitoring  bool
	batches       = make(map[string]*Batch)
	batchesMu     sync.RWMutex
	watcher       *fsnotify.Watcher
	watcherMu     sync.Mutex
	config        Config
	configPath    string
	monitorCtx    context.Context
	monitorCancel context.CancelFunc

	videoExts   = []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpeg", ".mpg", ".3gp", ".ts"}
	imageExts   = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico", ".tiff", ".psd"}
	audioExts   = []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a", ".opus"}
	docExts     = []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".csv"}
	archiveExts = []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz"}

	colorPurple = color.NRGBA{R: 138, G: 43, B: 226, A: 255}
	colorCyan   = color.NRGBA{R: 0, G: 220, B: 255, A: 255}
	colorGreen  = color.NRGBA{R: 0, G: 230, B: 118, A: 255}
	colorGray   = color.NRGBA{R: 100, G: 100, B: 120, A: 255}
)

func init() {
	config = Config{
		VideoEnabled:      true,
		ImageEnabled:      false,
		AudioEnabled:      false,
		DocEnabled:        false,
		ArchiveEnabled:    false,
		CustomExts:        "",
		MonitorSubdirs:    true,
		CompletionTimeout: 30,
		NotifyOnStart:     true,
		NotifyOnComplete:  true,
		SoundEnabled:      true,
		SaveHistory:       true,
	}
	configDir, _ := os.UserConfigDir()
	configPath = filepath.Join(configDir, "fidruawatch", "config.json")
	loadConfig()
}

func loadConfig() {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &config)
}

func saveConfig() {
	os.MkdirAll(filepath.Dir(configPath), 0755)
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configPath, data, 0644)
}

func getEnabledExts() []string {
	var exts []string
	if config.VideoEnabled {
		exts = append(exts, videoExts...)
	}
	if config.ImageEnabled {
		exts = append(exts, imageExts...)
	}
	if config.AudioEnabled {
		exts = append(exts, audioExts...)
	}
	if config.DocEnabled {
		exts = append(exts, docExts...)
	}
	if config.ArchiveEnabled {
		exts = append(exts, archiveExts...)
	}
	if config.CustomExts != "" {
		custom := strings.Split(config.CustomExts, ",")
		for _, ext := range custom {
			ext = strings.TrimSpace(ext)
			if ext != "" {
				if !strings.HasPrefix(ext, ".") {
					ext = "." + ext
				}
				exts = append(exts, strings.ToLower(ext))
			}
		}
	}
	return exts
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func main() {
	a := app.NewWithID("com.fidrua.watch")
	a.Settings().SetTheme(&customTheme{})
	w := a.NewWindow("FidruaWatch")
	w.Resize(fyne.NewSize(420, 700))
	w.CenterOnScreen()

	// ========== MONITOR TAB ==========
	title := canvas.NewText("FidruaWatch", colorPurple)
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	statusText := widget.NewLabel("点击开始监控")
	statusText.Alignment = fyne.TextAlignCenter

	// Play button - simple large button
	var playBtn *widget.Button
	playBtn = widget.NewButton("  ▶  ", nil)
	playBtn.Importance = widget.HighImportance

	// Folder selection
	folderLabel := widget.NewLabel("未选择文件夹")
	folderLabel.Alignment = fyne.TextAlignCenter

	var folderBtn *widget.Button
	folderBtn = widget.NewButton("📁 选择监控文件夹", nil)
	folderBtn.Importance = widget.HighImportance

	// Batch list
	batchList := container.NewVBox()
	batchScroll := container.NewVScroll(batchList)
	batchScroll.SetMinSize(fyne.NewSize(390, 250))

	uiUpdateChan := make(chan struct{}, 1)

	var updateBatchList func()
	updateBatchList = func() {
		batchList.Objects = nil
		batchesMu.RLock()
		defer batchesMu.RUnlock()

		if len(batches) == 0 {
			emptyLabel := widget.NewLabel("暂无上传批次")
			emptyLabel.Alignment = fyne.TextAlignCenter
			batchList.Add(container.NewCenter(emptyLabel))
		} else {
			sortedBatches := make([]*Batch, 0, len(batches))
			for _, b := range batches {
				sortedBatches = append(sortedBatches, b)
			}
			sort.Slice(sortedBatches, func(i, j int) bool {
				return sortedBatches[i].StartTime.After(sortedBatches[j].StartTime)
			})
			for _, batch := range sortedBatches {
				card := createBatchCard(batch, updateBatchList)
				batchList.Add(card)
			}
		}
		batchList.Refresh()
	}
	updateBatchList()

	requestUIUpdate := func() {
		select {
		case uiUpdateChan <- struct{}{}:
		default:
		}
	}

	go func() {
		for range uiUpdateChan {
			updateBatchList()
		}
	}()

	folderBtn.OnTapped = func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			monitorPath = uri.Path()
			// 显示路径，如果太长则截断
			displayPath := monitorPath
			if len(displayPath) > 40 {
				displayPath = "..." + displayPath[len(displayPath)-37:]
			}
			folderLabel.SetText(displayPath)
		}, w)
	}

	playBtn.OnTapped = func() {
		if !isMonitoring {
			if monitorPath == "" {
				dialog.ShowInformation("提示", "请先选择监控文件夹", w)
				return
			}
			if len(getEnabledExts()) == 0 {
				dialog.ShowInformation("提示", "请先在设置中启用至少一种文件类型", w)
				return
			}

			monitorCtx, monitorCancel = context.WithCancel(context.Background())
			if err := startMonitor(monitorPath); err != nil {
				monitorCancel()
				dialog.ShowError(err, w)
				return
			}

			isMonitoring = true
			playBtn.SetText("  ⏹  ")
			playBtn.Importance = widget.DangerImportance
			playBtn.Refresh()
			statusText.SetText("正在监控: " + filepath.Base(monitorPath))
			folderBtn.Disable()

			go handleFileEvents(monitorCtx, requestUIUpdate, a)
			go checkCompletions(monitorCtx, requestUIUpdate, a)
		} else {
			if monitorCancel != nil {
				monitorCancel()
			}
			stopMonitor()
			isMonitoring = false
			playBtn.SetText("  ▶  ")
			playBtn.Importance = widget.HighImportance
			playBtn.Refresh()
			statusText.SetText("点击开始监控")
			folderBtn.Enable()
		}
	}

	signAllBtn := widget.NewButton("✅ 全部签收", func() {
		batchesMu.Lock()
		for _, b := range batches {
			if b.Status == "completed" {
				b.Status = "signed"
			}
		}
		batchesMu.Unlock()
		updateBatchList()
	})

	clearBtn := widget.NewButton("🗑", func() {
		batchesMu.Lock()
		for id, b := range batches {
			if b.Status == "signed" {
				delete(batches, id)
			}
		}
		batchesMu.Unlock()
		updateBatchList()
	})

	batchHeader := container.NewHBox(
		widget.NewLabelWithStyle("📋 上传批次", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		signAllBtn,
		clearBtn,
	)

	monitorContent := container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(playBtn),
		container.NewCenter(statusText),
		widget.NewSeparator(),
		folderBtn,
		container.NewCenter(folderLabel),
		widget.NewSeparator(),
		batchHeader,
		batchScroll,
	)

	// ========== SETTINGS TAB ==========
	fileTypeBtn := widget.NewButton("⚙️ 设置监控的文件类型", func() {
		showFileTypeDialog(w)
	})

	subdirCheck := widget.NewCheck("📁 监控子文件夹", func(checked bool) {
		config.MonitorSubdirs = checked
	})
	subdirCheck.Checked = config.MonitorSubdirs

	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(fmt.Sprintf("%d", config.CompletionTimeout))
	timeoutEntry.Resize(fyne.NewSize(60, timeoutEntry.MinSize().Height))

	timeoutRow := container.NewHBox(
		widget.NewLabel("⏱️ 完成判定"),
		timeoutEntry,
		widget.NewLabel("秒"),
	)

	soundCheck := widget.NewCheck("🔊 声音提醒", func(checked bool) {
		config.SoundEnabled = checked
	})
	soundCheck.Checked = config.SoundEnabled

	startNotifyCheck := widget.NewCheck("📤 上传开始提醒", func(checked bool) {
		config.NotifyOnStart = checked
	})
	startNotifyCheck.Checked = config.NotifyOnStart

	completeNotifyCheck := widget.NewCheck("✅ 上传完成提醒", func(checked bool) {
		config.NotifyOnComplete = checked
	})
	completeNotifyCheck.Checked = config.NotifyOnComplete

	historyCheck := widget.NewCheck("📝 保存历史记录", func(checked bool) {
		config.SaveHistory = checked
	})
	historyCheck.Checked = config.SaveHistory

	saveBtn := widget.NewButton("💾 保存设置", func() {
		if t := timeoutEntry.Text; t != "" {
			var timeout int
			if _, err := fmt.Sscanf(t, "%d", &timeout); err == nil && timeout >= 10 {
				config.CompletionTimeout = timeout
			}
		}
		saveConfig()
		dialog.ShowInformation("成功", "设置已保存", w)
	})
	saveBtn.Importance = widget.HighImportance

	settingsContent := container.NewVBox(
		widget.NewLabelWithStyle("📁 文件监控", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		fileTypeBtn,
		subdirCheck,
		timeoutRow,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("🔔 通知设置", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		soundCheck,
		startNotifyCheck,
		completeNotifyCheck,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("⚙️ 其他", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		historyCheck,
		widget.NewSeparator(),
		saveBtn,
	)

	// ========== ABOUT TAB ==========
	// Use bundled logo
	var logoImage fyne.CanvasObject
	if resourceLogoPng != nil {
		logoRes := resourceLogoPng
		if logoRes != nil {
			img := canvas.NewImageFromResource(logoRes)
			img.SetMinSize(fyne.NewSize(120, 120))
			img.FillMode = canvas.ImageFillContain
			logoImage = img
		}
	}
	if logoImage == nil {
		// Fallback text logo
		logoText := canvas.NewText("🐕", colorPurple)
		logoText.TextSize = 60
		logoText.Alignment = fyne.TextAlignCenter
		logoImage = container.NewCenter(logoText)
	}

	aboutTitle := canvas.NewText("FidruaWatch", colorPurple)
	aboutTitle.TextSize = 26
	aboutTitle.TextStyle = fyne.TextStyle{Bold: true}
	aboutTitle.Alignment = fyne.TextAlignCenter

	versionLabel := canvas.NewText("v2.0.0", colorCyan)
	versionLabel.TextSize = 14
	versionLabel.Alignment = fyne.TextAlignCenter

	githubBtn := widget.NewButton("💻 GitHub 仓库", func() {
		u, _ := url.Parse("https://github.com/donma033x/FidruaWatch")
		_ = a.OpenURL(u)
	})

	downloadBtn := widget.NewButton("📥 下载最新版本", func() {
		u, _ := url.Parse("https://github.com/donma033x/FidruaWatch/releases")
		_ = a.OpenURL(u)
	})

	feedbackBtn := widget.NewButton("📧 反馈问题", func() {
		u, _ := url.Parse("https://github.com/donma033x/FidruaWatch/issues")
		_ = a.OpenURL(u)
	})

	copyrightLabel := widget.NewLabel("© 2024 Fidrua · donma033x")
	copyrightLabel.Alignment = fyne.TextAlignCenter

	licenseLabel := widget.NewLabel("Made with 💜 · MIT License")
	licenseLabel.Alignment = fyne.TextAlignCenter

	aboutContent := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(logoImage),
		container.NewCenter(aboutTitle),
		container.NewCenter(versionLabel),
		layout.NewSpacer(),
		githubBtn,
		downloadBtn,
		feedbackBtn,
		layout.NewSpacer(),
		container.NewCenter(copyrightLabel),
		container.NewCenter(licenseLabel),
	)

	// ========== TABS ==========
	tabs := container.NewAppTabs(
		container.NewTabItem("📡 监控", container.NewPadded(monitorContent)),
		container.NewTabItem("⚙️ 设置", container.NewPadded(settingsContent)),
		container.NewTabItem("ℹ️", container.NewPadded(aboutContent)),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
}

func createBatchCard(b *Batch, updateUI func()) fyne.CanvasObject {
	var statusColor color.Color
	var statusLabel string
	switch b.Status {
	case "uploading":
		statusColor = colorCyan
		statusLabel = "上传中"
	case "completed":
		statusColor = colorGreen
		statusLabel = "已完成"
	case "signed":
		statusColor = colorGray
		statusLabel = "已签收"
	}

	colorBar := canvas.NewRectangle(statusColor)
	colorBar.SetMinSize(fyne.NewSize(5, 70))

	folderName := filepath.Base(b.Folder)
	titleLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("📁 %s（%d个文件）", folderName, len(b.Files)),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	sizeStr := formatSize(b.TotalSize)
	infoLabel := widget.NewLabel(fmt.Sprintf("🕐 %s · %s · %s", b.StartTime.Format("15:04:05"), sizeStr, statusLabel))

	content := container.NewVBox(titleLabel, infoLabel)

	if b.Status == "completed" {
		signBtn := widget.NewButton("✅ 签收此批次", func() {
			batchesMu.Lock()
			b.Status = "signed"
			batchesMu.Unlock()
			updateUI()
		})
		signBtn.Importance = widget.SuccessImportance
		content.Add(signBtn)
	}

	// Card background
	cardBg := canvas.NewRectangle(color.NRGBA{R: 35, G: 40, B: 60, A: 255})
	cardBg.CornerRadius = 8

	cardContent := container.NewHBox(colorBar, container.NewPadded(content))
	card := container.NewStack(cardBg, cardContent)

	return container.NewPadded(card)
}

func showFileTypeDialog(w fyne.Window) {
	videoCheck := widget.NewCheck("🎬 视频 (.mp4, .avi, .mkv...)", func(checked bool) {
		config.VideoEnabled = checked
	})
	videoCheck.Checked = config.VideoEnabled

	imageCheck := widget.NewCheck("🖼️ 图片 (.jpg, .png, .gif...)", func(checked bool) {
		config.ImageEnabled = checked
	})
	imageCheck.Checked = config.ImageEnabled

	audioCheck := widget.NewCheck("🎵 音频 (.mp3, .wav, .flac...)", func(checked bool) {
		config.AudioEnabled = checked
	})
	audioCheck.Checked = config.AudioEnabled

	docCheck := widget.NewCheck("📄 文档 (.pdf, .doc, .xls...)", func(checked bool) {
		config.DocEnabled = checked
	})
	docCheck.Checked = config.DocEnabled

	archiveCheck := widget.NewCheck("📦 压缩包 (.zip, .rar, .7z...)", func(checked bool) {
		config.ArchiveEnabled = checked
	})
	archiveCheck.Checked = config.ArchiveEnabled

	customEntry := widget.NewEntry()
	customEntry.SetPlaceHolder("自定义后缀，如: .psd, .ai")
	customEntry.SetText(config.CustomExts)

	content := container.NewVBox(
		widget.NewLabel("选择要监控的文件类型："),
		videoCheck, imageCheck, audioCheck, docCheck, archiveCheck,
		widget.NewSeparator(),
		widget.NewLabel("自定义后缀（逗号分隔）："),
		customEntry,
	)

	dialog.ShowCustomConfirm("文件类型设置", "确定", "取消", content, func(ok bool) {
		if ok {
			config.CustomExts = customEntry.Text
		}
	}, w)
}

func startMonitor(path string) error {
	watcherMu.Lock()
	defer watcherMu.Unlock()

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if config.MonitorSubdirs {
		err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				watcher.Add(p)
			}
			return nil
		})
	} else {
		err = watcher.Add(path)
	}
	return err
}

func stopMonitor() {
	watcherMu.Lock()
	defer watcherMu.Unlock()
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
}

func handleFileEvents(ctx context.Context, updateUI func(), app fyne.App) {
	watcherMu.Lock()
	w := watcher
	watcherMu.Unlock()

	if w == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) != 0 {
				if config.MonitorSubdirs {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						watcherMu.Lock()
						if watcher != nil {
							watcher.Add(event.Name)
						}
						watcherMu.Unlock()
						continue
					}
				}
				if isMonitoredFile(event.Name) {
					isNewBatch := addFileToBatch(event.Name)
					if isNewBatch && config.NotifyOnStart {
						app.SendNotification(&fyne.Notification{
							Title:   "FidruaWatch - 新上传",
							Content: fmt.Sprintf("检测到新文件: %s", filepath.Base(event.Name)),
						})
					}
					updateUI()
				}
			}
		case _, ok := <-w.Errors:
			if !ok {
				return
			}
		}
	}
}

func isMonitoredFile(path string) bool {
	if isTempFile(path) {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	for _, ve := range getEnabledExts() {
		if ext == ve {
			return true
		}
	}
	return false
}

func isTempFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	for _, pattern := range tempFilePatterns {
		if strings.Contains(name, pattern) || strings.HasPrefix(name, pattern) {
			return true
		}
	}
	return false
}

func addFileToBatch(filePath string) (isNewBatch bool) {
	folder := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)

	var fileSize int64
	if info, err := os.Stat(filePath); err == nil {
		fileSize = info.Size()
	}

	batchesMu.Lock()
	defer batchesMu.Unlock()

	var batch *Batch
	for _, b := range batches {
		if b.Folder == folder && b.Status == "uploading" {
			batch = b
			break
		}
	}

	if batch == nil {
		batch = &Batch{
			ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
			Folder:    folder,
			Files:     []string{},
			FileSizes: make(map[string]int64),
			Status:    "uploading",
			StartTime: time.Now(),
		}
		batches[batch.ID] = batch
		isNewBatch = true
	}

	exists := false
	for _, f := range batch.Files {
		if f == fileName {
			exists = true
			break
		}
	}
	if !exists {
		batch.Files = append(batch.Files, fileName)
	}

	oldSize := batch.FileSizes[fileName]
	if fileSize > oldSize {
		batch.TotalSize += fileSize - oldSize
		batch.FileSizes[fileName] = fileSize
	}

	batch.LastTime = time.Now()
	return
}

func checkCompletions(ctx context.Context, updateUI func(), app fyne.App) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	timeout := time.Duration(config.CompletionTimeout) * time.Second
	if timeout < 10*time.Second {
		timeout = 30 * time.Second
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batchesMu.Lock()
			for _, b := range batches {
				if b.Status == "uploading" && time.Since(b.LastTime) > timeout {
					b.Status = "completed"
					if config.NotifyOnComplete {
						app.SendNotification(&fyne.Notification{
							Title:   "FidruaWatch - 上传完成",
							Content: fmt.Sprintf("批次完成: %s (%d个文件)", filepath.Base(b.Folder), len(b.Files)),
						})
					}
				}
			}
			batchesMu.Unlock()
			updateUI()
		}
	}
}
