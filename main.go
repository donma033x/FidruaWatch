package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/fsnotify/fsnotify"
)

// Batch represents an upload batch
type Batch struct {
	ID        string
	Folder    string
	Files     []string
	Status    string // "uploading", "completed", "signed"
	StartTime time.Time
	LastTime  time.Time
}

// Config represents app settings
type Config struct {
	VideoEnabled    bool     `json:"video_enabled"`
	ImageEnabled    bool     `json:"image_enabled"`
	AudioEnabled    bool     `json:"audio_enabled"`
	DocEnabled      bool     `json:"doc_enabled"`
	ArchiveEnabled  bool     `json:"archive_enabled"`
	CustomExts      string   `json:"custom_exts"`
	MonitorSubdirs  bool     `json:"monitor_subdirs"`
}

var (
	monitorPath  string
	isMonitoring bool
	batches      = make(map[string]*Batch)
	batchesMu    sync.RWMutex
	watcher      *fsnotify.Watcher
	watcherMu    sync.Mutex
	config       Config
	configPath   string

	// Context for controlling goroutines
	monitorCtx    context.Context
	monitorCancel context.CancelFunc

	// File type categories
	videoExts   = []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpeg", ".mpg", ".3gp", ".ts"}
	imageExts   = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".ico", ".tiff", ".psd"}
	audioExts   = []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a", ".opus"}
	docExts     = []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".md", ".csv"}
	archiveExts = []string{".zip", ".rar", ".7z", ".tar", ".gz", ".bz2", ".xz"}

	colorCyan  = color.NRGBA{R: 0, G: 217, B: 255, A: 255}
	colorGreen = color.NRGBA{R: 0, G: 255, B: 136, A: 255}
	colorGray  = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
)

func init() {
	// Default config
	config = Config{
		VideoEnabled:   true,
		ImageEnabled:   false,
		AudioEnabled:   false,
		DocEnabled:     false,
		ArchiveEnabled: false,
		CustomExts:     "",
		MonitorSubdirs: true,
	}

	// Config file path
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
	// Custom extensions
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

func main() {
	a := app.NewWithID("com.fidrua.watch")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("FidruaWatch")
	w.Resize(fyne.NewSize(420, 720))
	w.CenterOnScreen()

	// ===== Monitor Tab =====
	title := canvas.NewText("FidruaWatch", colorCyan)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("专业的批量文件上传监控工具")
	subtitle.Alignment = fyne.TextAlignCenter

	statusIcon := canvas.NewText("⏸", colorGray)
	statusIcon.TextSize = 36
	statusIcon.Alignment = fyne.TextAlignCenter

	statusTitle := widget.NewLabel("监控已停止")
	statusTitle.Alignment = fyne.TextAlignCenter
	statusTitle.TextStyle = fyne.TextStyle{Bold: true}

	statusDesc := widget.NewLabel("点击开始监控文件上传")
	statusDesc.Alignment = fyne.TextAlignCenter

	folderPath := widget.NewLabel("未选择文件夹")
	folderPath.Alignment = fyne.TextAlignCenter
	folderPath.Wrapping = fyne.TextWrapWord

	batchList := container.NewVBox()
	batchScroll := container.NewVScroll(batchList)
	batchScroll.SetMinSize(fyne.NewSize(380, 200))

	// Channel for thread-safe UI updates
	uiUpdateChan := make(chan struct{}, 1)

	var updateBatchList func()
	updateBatchList = func() {
		batchList.Objects = nil
		batchesMu.RLock()
		defer batchesMu.RUnlock()

		if len(batches) == 0 {
			emptyLabel := widget.NewLabel("暂无上传批次")
			emptyLabel.Alignment = fyne.TextAlignCenter
			batchList.Add(emptyLabel)
		} else {
			// Sort batches by start time (newest first)
			sortedBatches := make([]*Batch, 0, len(batches))
			for _, b := range batches {
				sortedBatches = append(sortedBatches, b)
			}
			sort.Slice(sortedBatches, func(i, j int) bool {
				return sortedBatches[i].StartTime.After(sortedBatches[j].StartTime)
			})

			for _, batch := range sortedBatches {
				b := batch // capture for closure
				statusText := ""
				switch b.Status {
				case "uploading":
					statusText = "📤 上传中"
				case "completed":
					statusText = "✅ 已完成"
				case "signed":
					statusText = "✔️ 已签收"
				}

				folderName := filepath.Base(b.Folder)
				headerText := fmt.Sprintf("%s (%d 个文件) - %s", folderName, len(b.Files), statusText)

				details := container.NewVBox(
					widget.NewLabel(fmt.Sprintf("📁 %s", b.Folder)),
					widget.NewLabel(fmt.Sprintf("⏰ %s", b.StartTime.Format("15:04:05"))),
				)

				if b.Status == "completed" {
					signBtn := widget.NewButton("✅ 签收此批次", func() {
						batchesMu.Lock()
						if batch, ok := batches[b.ID]; ok {
							batch.Status = "signed"
						}
						batchesMu.Unlock()
						updateBatchList()
					})
					details.Add(signBtn)
				}

				card := widget.NewCard(headerText, "", details)
				batchList.Add(card)
			}
		}
		batchList.Refresh()
	}
	updateBatchList()

	// Request UI update (non-blocking, thread-safe)
	requestUIUpdate := func() {
		select {
		case uiUpdateChan <- struct{}{}:
		default:
			// Already has pending update
		}
	}

	// Background goroutine to process UI updates on main thread
	go func() {
		for range uiUpdateChan {
			updateBatchList()
		}
	}()

	var startBtn, stopBtn, folderBtn *widget.Button

	startBtn = widget.NewButton("▶ 开始监控", func() {
		if monitorPath == "" {
			dialog.ShowInformation("提示", "请先选择监控文件夹", w)
			return
		}
		// Check if any file type is enabled
		if len(getEnabledExts()) == 0 {
			dialog.ShowInformation("提示", "请先在设置中启用至少一种文件类型", w)
			return
		}
		// Create context for this monitoring session
		monitorCtx, monitorCancel = context.WithCancel(context.Background())
		
		if err := startMonitor(monitorPath); err != nil {
			monitorCancel()
			dialog.ShowError(err, w)
			return
		}
		isMonitoring = true
		statusIcon.Text = "🟢"
		statusIcon.Color = colorGreen
		statusIcon.Refresh()
		statusTitle.SetText("正在监控")
		statusDesc.SetText(filepath.Base(monitorPath))
		startBtn.Disable()
		stopBtn.Enable()
		folderBtn.Disable() // Disable folder selection during monitoring
		// Start goroutines with context
		go handleFileEvents(monitorCtx, requestUIUpdate)
		go checkCompletions(monitorCtx, requestUIUpdate)
	})

	stopBtn = widget.NewButton("⏹ 停止", func() {
		// Cancel context first to stop goroutines
		if monitorCancel != nil {
			monitorCancel()
		}
		stopMonitor()
		isMonitoring = false
		statusIcon.Text = "⏸"
		statusIcon.Color = colorGray
		statusIcon.Refresh()
		statusTitle.SetText("监控已停止")
		statusDesc.SetText("点击开始监控文件上传")
		startBtn.Enable()
		stopBtn.Disable()
		folderBtn.Enable() // Re-enable folder selection
	})
	stopBtn.Disable()

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

	clearBtn := widget.NewButton("🗑 清空已签收", func() {
		batchesMu.Lock()
		for id, b := range batches {
			if b.Status == "signed" {
				delete(batches, id)
			}
		}
		batchesMu.Unlock()
		updateBatchList()
	})

	folderBtn = widget.NewButton("📁 选择监控文件夹", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			monitorPath = uri.Path()
			folderPath.SetText(monitorPath)
		}, w)
	})

	monitorTab := container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
		container.NewCenter(statusIcon),
		container.NewCenter(statusTitle),
		container.NewCenter(statusDesc),
		widget.NewSeparator(),
		folderBtn,
		container.NewCenter(folderPath),
		container.NewGridWithColumns(2, startBtn, stopBtn),
		container.NewGridWithColumns(2, signAllBtn, clearBtn),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("📋 上传批次", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		batchScroll,
	)

	// ===== Settings Tab =====
	videoCheck := widget.NewCheck("🎬 视频 (.mp4, .avi, .mkv, .mov...)", func(checked bool) {
		config.VideoEnabled = checked
	})
	videoCheck.Checked = config.VideoEnabled

	imageCheck := widget.NewCheck("🖼 图片 (.jpg, .png, .gif, .webp...)", func(checked bool) {
		config.ImageEnabled = checked
	})
	imageCheck.Checked = config.ImageEnabled

	audioCheck := widget.NewCheck("🎵 音频 (.mp3, .wav, .flac, .aac...)", func(checked bool) {
		config.AudioEnabled = checked
	})
	audioCheck.Checked = config.AudioEnabled

	docCheck := widget.NewCheck("📄 文档 (.pdf, .doc, .xls, .ppt...)", func(checked bool) {
		config.DocEnabled = checked
	})
	docCheck.Checked = config.DocEnabled

	archiveCheck := widget.NewCheck("📦 压缩包 (.zip, .rar, .7z, .tar...)", func(checked bool) {
		config.ArchiveEnabled = checked
	})
	archiveCheck.Checked = config.ArchiveEnabled

	customEntry := widget.NewEntry()
	customEntry.SetPlaceHolder("例如: .psd, .ai, .sketch")
	customEntry.SetText(config.CustomExts)

	subdirCheck := widget.NewCheck("📂 监控子文件夹", func(checked bool) {
		config.MonitorSubdirs = checked
	})
	subdirCheck.Checked = config.MonitorSubdirs

	saveBtn := widget.NewButton("💾 保存设置", func() {
		config.CustomExts = customEntry.Text
		saveConfig()
		dialog.ShowInformation("成功", "设置已保存", w)
	})

	settingsTab := container.NewVBox(
		widget.NewLabelWithStyle("📁 监控的文件类型", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		videoCheck,
		imageCheck,
		audioCheck,
		docCheck,
		archiveCheck,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("✏️ 自定义后缀 (逗号分隔)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		customEntry,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("⚙️ 其他设置", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		subdirCheck,
		widget.NewSeparator(),
		saveBtn,
	)

	// ===== About Tab =====
	aboutTitle := canvas.NewText("FidruaWatch", colorCyan)
	aboutTitle.TextSize = 28
	aboutTitle.TextStyle = fyne.TextStyle{Bold: true}
	aboutTitle.Alignment = fyne.TextAlignCenter

	aboutTab := container.NewVBox(
		container.NewCenter(aboutTitle),
		widget.NewLabelWithStyle("v2.0.0", fyne.TextAlignCenter, fyne.TextStyle{}),
		widget.NewSeparator(),
		widget.NewLabel("专业的批量文件上传监控工具"),
		widget.NewLabel(""),
		widget.NewLabel("✨ 同目录文件自动归批"),
		widget.NewLabel("✨ 开始上传即时通知"),
		widget.NewLabel("✨ 30秒无变动自动完成"),
		widget.NewLabel("✨ 批次签收管理"),
		widget.NewLabel("✨ 跨平台支持"),
		widget.NewLabel("✨ 自定义文件类型"),
		widget.NewLabel(""),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("© 2024 Fidrua", fyne.TextAlignCenter, fyne.TextStyle{}),
	)

	// ===== Tabs =====
	tabs := container.NewAppTabs(
		container.NewTabItem("📡 监控", container.NewPadded(monitorTab)),
		container.NewTabItem("⚙️ 设置", container.NewPadded(settingsTab)),
		container.NewTabItem("ℹ️ 关于", container.NewPadded(aboutTab)),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
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

func handleFileEvents(ctx context.Context, updateUI func()) {
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
			// Handle file create/write events
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				// Check if new directory was created (for subdirectory monitoring)
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
				// Handle monitored files
				if isMonitoredFile(event.Name) {
					addFileToBatch(event.Name)
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
	ext := strings.ToLower(filepath.Ext(path))
	for _, ve := range getEnabledExts() {
		if ext == ve {
			return true
		}
	}
	return false
}

func addFileToBatch(filePath string) {
	folder := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)

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
			Status:    "uploading",
			StartTime: time.Now(),
		}
		batches[batch.ID] = batch
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
	batch.LastTime = time.Now()
}

func checkCompletions(ctx context.Context, updateUI func()) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batchesMu.Lock()
			for _, b := range batches {
				if b.Status == "uploading" && time.Since(b.LastTime) > 30*time.Second {
					b.Status = "completed"
				}
			}
			batchesMu.Unlock()

			updateUI()
		}
	}
}
