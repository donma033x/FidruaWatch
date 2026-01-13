# FidruaWatch

<p align="center">
  <strong>Professional Batch Video Upload Monitor</strong>
</p>

<p align="center">
  <a href="README_CN.md">中文</a> | English
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-00d9ff" alt="Platform">
  <img src="https://img.shields.io/badge/version-2.0.0-00ff88" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="License">
</p>

---

## ✨ Features

- 🌐 **Cross-platform** - Windows / macOS / Linux
- 📁 **Smart Batching** - Files in same directory grouped as one batch
- 🔔 **Start Notification** - Alert when upload detected
- ⏱️ **Completion Detection** - Auto-complete after 30s of no activity
- ✅ **Batch Acknowledgment** - Confirm processed upload batches
- 🎬 **Video Focused** - Monitor common video formats by default
- 🚀 **Lightweight** - ~15MB, no WebView dependency

---

## 📥 Download

Go to [Releases](https://github.com/donma033x/FidruaWatch/releases) to download:

| Platform | File |
|----------|------|
| 🪟 Windows | `fidruawatch-windows-amd64.zip` |
| 🍎 macOS (Intel) | `fidruawatch-darwin-amd64.tar.gz` |
| 🍎 macOS (Apple Silicon) | `fidruawatch-darwin-arm64.tar.gz` |
| 🐧 Linux | `fidruawatch-linux-amd64.tar.gz` |

---

## 🚀 Usage

1. **Select Folder** - Choose the video upload folder to monitor
2. **Start Monitoring** - Click "Start" button
3. **Upload Started** - Get notified when new video files detected
4. **Upload Completed** - Get notified after 30s of no new file activity
5. **Acknowledge** - Click to confirm processed batches

---

## 🎬 Supported File Formats

Default monitored formats:

```
.mp4  .avi  .mkv  .mov  .wmv  .flv  .webm  .m4v  .mpeg  .mpg  .3gp  .ts
```

> 💡 **Tip**: You can monitor any file type by modifying the `videoExts` variable in `main.go` and rebuilding. For example, add `.jpg`, `.png` for images, or `.pdf`, `.doc` for documents.

---

## 🛠️ Build from Source

### Requirements

- [Go](https://golang.org/) >= 1.21
- GCC (for CGO)
  - Windows: MinGW-w64
  - macOS: Xcode Command Line Tools
  - Linux: `gcc`, `libgl1-mesa-dev`, `xorg-dev`

### Build

```bash
git clone https://github.com/donma033x/FidruaWatch.git
cd FidruaWatch
go build -o fidruawatch .
```

### Tech Stack

- **GUI**: [Fyne](https://fyne.io/) v2
- **File Watcher**: fsnotify
- **Language**: Go

---

## 📄 License

MIT License

---

<p align="center">
  Made with 💙 by Fidrua
</p>
