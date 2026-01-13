package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
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

var (
	monitorPath  string
	isMonitoring bool
	batches      = make(map[string]*Batch)
	batchesMu    sync.RWMutex
	watcher      *fsnotify.Watcher
	videoExts    = []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm", ".m4v", ".mpeg", ".mpg", ".3gp", ".ts"}
	colorCyan    = color.NRGBA{R: 0, G: 217, B: 255, A: 255}
	colorGreen   = color.NRGBA{R: 0, G: 255, B: 136, A: 255}
	colorGray    = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
)

func main() {
	a := app.NewWithID("com.fidrua.watch")
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("FidruaWatch")
	w.Resize(fyne.NewSize(420, 680))
	w.CenterOnScreen()

	// Title
	title := canvas.NewText("FidruaWatch", colorCyan)
	title.TextSize = 24
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	subtitle := widget.NewLabel("专业的批量视频上传监控工具")
	subtitle.Alignment = fyne.TextAlignCenter

	// Status widgets
	statusIcon := canvas.NewText("⏸", colorGray)
	statusIcon.TextSize = 36
	statusIcon.Alignment = fyne.TextAlignCenter

	statusTitle := widget.NewLabel("监控已停止")
	statusTitle.Alignment = fyne.TextAlignCenter
	statusTitle.TextStyle = fyne.TextStyle{Bold: true}

	statusDesc := widget.NewLabel("点击开始监控视频上传")
	statusDesc.Alignment = fyne.TextAlignCenter

	// Folder path display
	folderPath := widget.NewLabel("未选择文件夹")
	folderPath.Alignment = fyne.TextAlignCenter
	folderPath.Wrapping = fyne.TextWrapWord

	// Batch list
	batchList := container.NewVBox()
	batchScroll := container.NewVScroll(batchList)
	batchScroll.SetMinSize(fyne.NewSize(380, 250))

	// Update batch list UI
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
			for _, b := range batches {
				batch := b
				statusText := ""
				switch batch.Status {
				case "uploading":
					statusText = "📤 上传中"
				case "completed":
					statusText = "✅ 已完成"
				case "signed":
					statusText = "✔️ 已签收"
				}

				folderName := filepath.Base(batch.Folder)
				headerText := fmt.Sprintf("%s (%d 个文件) - %s", folderName, len(batch.Files), statusText)

				details := container.NewVBox(
					widget.NewLabel(fmt.Sprintf("📁 %s", batch.Folder)),
					widget.NewLabel(fmt.Sprintf("⏰ %s", batch.StartTime.Format("15:04:05"))),
				)

				if batch.Status == "completed" {
					signBtn := widget.NewButton("✅ 签收此批次", func() {
						batchesMu.Lock()
						if b, ok := batches[batch.ID]; ok {
							b.Status = "signed"
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

	// Initialize batch list
	updateBatchList()

	// Buttons
	var startBtn, stopBtn *widget.Button

	startBtn = widget.NewButton("▶ 开始监控", func() {
		if monitorPath == "" {
			dialog.ShowInformation("提示", "请先选择监控文件夹", w)
			return
		}
		if err := startMonitor(monitorPath); err != nil {
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

		// Start completion checker
		go checkCompletions(updateBatchList)
	})

	stopBtn = widget.NewButton("⏹ 停止", func() {
		stopMonitor()
		isMonitoring = false
		statusIcon.Text = "⏸"
		statusIcon.Color = colorGray
		statusIcon.Refresh()
		statusTitle.SetText("监控已停止")
		statusDesc.SetText("点击开始监控视频上传")
		startBtn.Enable()
		stopBtn.Disable()
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

	// Folder select button
	folderBtn := widget.NewButton("📁 选择监控文件夹", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			monitorPath = uri.Path()
			folderPath.SetText(monitorPath)
		}, w)
	})

	// Layout
	header := container.NewVBox(
		container.NewCenter(title),
		container.NewCenter(subtitle),
		widget.NewSeparator(),
	)

	statusCard := container.NewVBox(
		container.NewCenter(statusIcon),
		container.NewCenter(statusTitle),
		container.NewCenter(statusDesc),
	)

	folderCard := container.NewVBox(
		folderBtn,
		container.NewCenter(folderPath),
	)

	btnRow1 := container.NewGridWithColumns(2, startBtn, stopBtn)
	btnRow2 := container.NewGridWithColumns(2, signAllBtn, clearBtn)

	batchHeader := widget.NewLabel("📋 上传批次 (同目录文件为一批，30秒无变动视为完成)")
	batchHeader.TextStyle = fyne.TextStyle{Bold: true}

	batchSection := container.NewVBox(
		batchHeader,
		batchScroll,
	)

	content := container.NewVBox(
		header,
		statusCard,
		widget.NewSeparator(),
		folderCard,
		btnRow1,
		btnRow2,
		widget.NewSeparator(),
		batchSection,
	)

	w.SetContent(container.NewPadded(content))

	// Set up file event handler
	go handleFileEvents(updateBatchList)

	w.ShowAndRun()
}

func startMonitor(path string) error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Add directory recursively
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			watcher.Add(p)
		}
		return nil
	})

	return err
}

func stopMonitor() {
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
}

func handleFileEvents(updateUI func()) {
	for {
		if watcher == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Write) != 0 {
				if isVideoFile(event.Name) {
					addFileToBatch(event.Name)
					updateUI()
				}
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, ve := range videoExts {
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

	// Find or create batch for this folder
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

	// Add file if not exists
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

func checkCompletions(updateUI func()) {
	for isMonitoring {
		time.Sleep(5 * time.Second)

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
