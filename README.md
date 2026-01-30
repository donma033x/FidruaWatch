# FidruaWatch

<p align="center">
  <img src="logo.png" alt="FidruaWatch Logo" width="200">
</p>

<p align="center">
  <strong>Professional Batch File Upload Monitor</strong>
</p>

<p align="center">
  <a href="README_CN.md">中文文档</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-00d9ff" alt="Platform">
  <img src="https://img.shields.io/badge/version-2.1.2-00ff88" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="License">
</p>

---

## ✨ Features

- 🌐 **Cross-platform** - Windows / macOS / Linux
- 📁 **Smart Batching** - Files in same directory auto-grouped into batches
- 🔔 **Instant Notifications** - System alerts on upload start/complete
- ⏱️ **Configurable Timeout** - Custom inactivity threshold (default 30s)
- ✅ **Batch Sign-off** - Confirm processed upload batches
- 📊 **Size Statistics** - Real-time batch file size display
- 🚫 **Temp File Filter** - Auto-ignore .tmp/.part files
- 🔄 **FTP Friendly** - Supports FTP temp file rename scenarios
- 🚀 **Lightweight** - ~25MB, no WebView dependency
- 🚀 **Auto Start** - Launch on system startup

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

1. **Select Folder** - Choose the upload directory to monitor (e.g., FTP root)
2. **Start Monitoring** - Click the "Start" button
3. **Upload Starts** - Receive notification when new files detected
4. **Upload Complete** - Auto-marked complete after inactivity timeout
5. **Sign Off Batch** - Click to confirm processed batches

---

## ⚙️ Settings

### File Types

| Type | Extensions |
|------|------------|
| 🎬 Video | `.mp4` `.avi` `.mkv` `.mov` `.wmv` `.flv` `.webm` `.m4v` `.mpeg` `.mpg` `.3gp` `.ts` |
| 🖼 Image | `.jpg` `.jpeg` `.png` `.gif` `.bmp` `.webp` `.svg` `.ico` `.tiff` `.psd` |
| 🎵 Audio | `.mp3` `.wav` `.flac` `.aac` `.ogg` `.wma` `.m4a` `.opus` |
| 📄 Document | `.pdf` `.doc` `.docx` `.xls` `.xlsx` `.ppt` `.pptx` `.txt` `.md` `.csv` |
| 📦 Archive | `.zip` `.rar` `.7z` `.tar` `.gz` `.bz2` `.xz` |
| ✏️ Custom | Add any extension in settings |

### Other Settings

- **Monitor Subdirectories** - Recursively monitor subdirectories
- **Notify on Start** - Send notification when new batch detected
- **Notify on Complete** - Send notification when batch completes
- **Completion Timeout** - Seconds of inactivity before marking complete (default 30s, min 10s)
- **Auto Start** - Launch application on system startup

---

## 🔧 FTP Monitoring

This tool is ideal for monitoring FTP/SFTP server upload directories:

- ✅ Auto-filter FTP client temporary files
- ✅ Support post-upload rename scenarios
- ✅ Large file uploads won't trigger false completion
- ✅ Configurable timeout for different network conditions

**Note**: Monitor directory must be a locally mounted path. Network mapped drives may not support real-time file monitoring.

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
- **File Watching**: fsnotify
- **Language**: Go

---

## 📝 Changelog

### v2.1.1 (2025-01-14)
- ✨ **Tab Bar Layout** - Three tabs evenly distributed
- ✨ **Auto Start** - Support Windows/macOS/Linux startup
- 👍 **Dialog Improvements** - Larger folder picker and settings dialogs
- 🐛 **Fix** - About page version display

### v2.0.0
- 🎉 Brand new UI design
- 📁 Smart batch grouping
- 🔔 System notification support
- 🔄 FTP upload scenario optimization

---

## 📄 License

MIT License

---

<p align="center">
  Made with 💜 by Fidrua
</p>
