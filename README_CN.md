# FidruaWatch

<p align="center">
  <img src="logo.png" alt="FidruaWatch Logo" width="200">
</p>

<p align="center">
  <strong>专业的批量文件上传监控工具</strong>
</p>

<p align="center">
  中文 | <a href="README.md">English</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-00d9ff" alt="Platform">
  <img src="https://img.shields.io/badge/version-2.2.0-00ff88" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-blue" alt="License">
</p>

---

## ✨ 功能特性

- 🌐 **跨平台** - Windows / macOS / Linux
- 📁 **智能归批** - 同目录文件自动归为一个批次
- 🔔 **即时通知** - 新上传开始和完成时系统通知
- ⏱️ **可配置超时** - 自定义无活动判定时间（默认30秒）
- ✅ **批次签收** - 确认已处理的上传批次
- 📊 **大小统计** - 实时显示批次文件总大小
- 🚫 **临时文件过滤** - 自动忽略 .tmp/.part 等临时文件
- 🔄 **FTP友好** - 支持FTP上传的临时文件重命名场景
- 🚀 **轻量级** - ~25MB，无 WebView 依赖
- 🚀 **开机自启** - 支持开机自动启动

---

## 📥 下载

前往 [Releases](https://github.com/donma033x/FidruaWatch/releases) 下载:

| 平台 | 文件 |
|------|------|
| 🪟 Windows | `fidruawatch-windows-amd64.zip` |
| 🍎 macOS (Intel) | `fidruawatch-darwin-amd64.tar.gz` |
| 🍎 macOS (Apple Silicon) | `fidruawatch-darwin-arm64.tar.gz` |
| 🐧 Linux | `fidruawatch-linux-amd64.tar.gz` |

---

## 🚀 使用方法

1. **选择文件夹** - 选择要监控的上传目录（如FTP根目录）
2. **开始监控** - 点击“开始监控”按钮
3. **上传开始** - 检测到新文件时收到系统通知
4. **上传完成** - 无新文件活动超过设定时间后自动标记完成
5. **批次签收** - 点击签收确认已处理的批次

---

## ⚙️ 设置选项

### 文件类型

| 类型 | 扩展名 |
|------|--------|
| 🎬 视频 | `.mp4` `.avi` `.mkv` `.mov` `.wmv` `.flv` `.webm` `.m4v` `.mpeg` `.mpg` `.3gp` `.ts` |
| 🖼 图片 | `.jpg` `.jpeg` `.png` `.gif` `.bmp` `.webp` `.svg` `.ico` `.tiff` `.psd` |
| 🎵 音频 | `.mp3` `.wav` `.flac` `.aac` `.ogg` `.wma` `.m4a` `.opus` |
| 📄 文档 | `.pdf` `.doc` `.docx` `.xls` `.xlsx` `.ppt` `.pptx` `.txt` `.md` `.csv` |
| 📦 压缩包 | `.zip` `.rar` `.7z` `.tar` `.gz` `.bz2` `.xz` |
| ✏️ 自定义 | 在设置中添加任意扩展名 |

### 其他设置

- **监控子文件夹** - 是否递归监控子目录
- **新上传时通知** - 检测到新批次时发送系统通知
- **上传完成时通知** - 批次完成时发送系统通知
- **声音选择** - 为开始/完成事件选择不同的提示音
- **未签收提醒** - 定时提醒未签收的批次
- **完成超时(秒)** - 无新文件写入多久后判定上传完成（默认30秒，最小10秒）
- **开机自启动** - 系统启动时自动运行程序

---

## 🔧 FTP 监控场景

本工具特别适合监控 FTP/SFTP 服务器的上传目录：

- ✅ 自动过滤 FTP 客户端产生的临时文件
- ✅ 支持上传完成后重命名的场景
- ✅ 大文件长时间上传不会误判完成
- ✅ 可配置超时适应不同网络环境

**注意**: 监控目录需要是本地挂载的路径，网络映射驱动器可能不支持实时文件监控。

---

## 🛠️ 从源码构建

### 环境要求

- [Go](https://golang.org/) >= 1.21
- GCC (CGO 需要)
  - Windows: MinGW-w64
  - macOS: Xcode Command Line Tools
  - Linux: `gcc`, `libgl1-mesa-dev`, `xorg-dev`

### 构建命令

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

## 📝 更新日志

### v2.2.0 (2025-01-30)
- ✨ **声音选择** - 可从系统声音中选择提示音
- ✨ **独立声音** - 开始上传和上传完成可设置不同声音
- ✨ **未签收提醒** - 定时提醒等待签收的批次
- ✨ **程序图标** - 任务栏/Dock显示应用图标
- 🐛 **修复** - Windows上播放声音不再弹出黑窗

### v2.1.1 (2025-01-14)
- ✨ **Tab栏平铺布局** - 三个标签平均分布
- ✨ **开机自启动** - 支持 Windows/macOS/Linux
- 👍 **弹窗优化** - 文件夹选择和设置对话框更大更舒适
- 🐛 **修复** - 关于页面版本号显示

### v2.0.0
- 🎉 全新 UI 设计
- 📁 智能归批功能
- 🔔 系统通知支持
- 🔄 FTP 上传场景优化

---

## 📄 许可证

MIT License

---

<p align="center">
  Made with 💜 by Fidrua
</p>
