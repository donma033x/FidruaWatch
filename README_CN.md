# FidruaWatch

<p align="center">
  <strong>专业的批量视频上传监控工具</strong>
</p>

<p align="center">
  中文 | <a href="README.md">English</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-00d9ff" alt="Platform">
  <img src="https://img.shields.io/badge/version-2.0.0-00ff88" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="License">
</p>

---

## ✨ 功能特点

- 🌐 **跨平台支持** - Windows / macOS / Linux
- 📁 **智能批次** - 同目录文件自动归为一批
- 🔔 **开始通知** - 检测到上传立即提醒
- ⏱️ **完成检测** - 30秒无变动自动判定上传完成
- ✅ **批次签收** - 确认已处理的上传批次
- 🎬 **视频专用** - 默认监控常见视频格式
- 🚀 **轻量级** - 约 15MB，无 WebView 依赖

---

## 📥 下载

前往 [Releases](https://github.com/donma033x/FidruaWatch/releases) 页面下载：

| 平台 | 文件 |
|------|------|
| 🪟 Windows | `fidruawatch-windows-amd64.zip` |
| 🍎 macOS (Intel) | `fidruawatch-darwin-amd64.tar.gz` |
| 🍎 macOS (Apple Silicon) | `fidruawatch-darwin-arm64.tar.gz` |
| 🐧 Linux | `fidruawatch-linux-amd64.tar.gz` |

---

## 🚀 使用方法

1. **选择监控目录** - 点击选择要监控的视频上传文件夹
2. **开始监控** - 点击“开始监控”按钮
3. **开始上传通知** - 检测到新视频文件时会提醒
4. **完成通知** - 30秒内无新文件变动则提醒“上传完成”
5. **签收确认** - 点击签收确认已处理的批次

---

## 🎬 支持的视频格式

```
.mp4  .avi  .mkv  .mov  .wmv  .flv  .webm  .m4v  .mpeg  .mpg  .3gp  .ts
```

---

## 🛠️ 从源码构建

### 环境要求

- [Go](https://golang.org/) >= 1.21
- GCC (用于 CGO)
  - Windows: MinGW-w64
  - macOS: Xcode Command Line Tools
  - Linux: `gcc`, `libgl1-mesa-dev`, `xorg-dev`

### 构建

```bash
git clone https://github.com/donma033x/FidruaWatch.git
cd FidruaWatch
go build -o fidruawatch .
```

### 技术栈

- **GUI**: [Fyne](https://fyne.io/) v2
- **文件监控**: fsnotify
- **语言**: Go

---

## 📄 许可证

MIT License

---

<p align="center">
  Made with 💙 by Fidrua
</p>
