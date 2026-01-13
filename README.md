# FidruaWatch

<p align="center">
  <img src="docs/logo.png" width="180" alt="FidruaWatch Logo">
</p>

<p align="center">
  <strong>专业的批量视频上传监控工具</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-00d9ff" alt="Platform">
  <img src="https://img.shields.io/badge/version-1.2.0-00ff88" alt="Version">
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
- 🎨 **科技界面** - 蓝绿主题高科技风格

---

## 📸 截图

<table>
  <tr>
    <td align="center">
      <img src="docs/screenshot-monitor.png" width="380" alt="监控界面">
      <br><strong>监控界面</strong>
    </td>
    <td align="center">
      <img src="docs/screenshot-settings.png" width="380" alt="设置界面">
      <br><strong>设置界面</strong>
    </td>
  </tr>
</table>

---

## 📥 下载

前往 [Releases](https://github.com/donma033x/FidruaWatch/releases) 页面下载最新版本：

| 平台 | 文件 |
|------|------|
| 🪟 Windows | `.msi` 或 `.exe` |
| 🍎 macOS (Intel) | `_x64.dmg` |
| 🍎 macOS (Apple Silicon) | `_aarch64.dmg` |
| 🐧 Linux | `.AppImage` 或 `.deb` |

---

## 🚀 使用方法

1. **选择监控目录** - 点击选择要监控的视频上传文件夹
2. **开始监控** - 点击"开始监控"按钮
3. **开始上传通知** - 检测到新视频文件时会提醒
4. **完成通知** - 30秒内无新文件变动则提醒"上传完成"
5. **签收确认** - 点击签收确认已处理的批次

---

## 🎬 支持的视频格式

默认监控以下格式（可在设置中自定义）：

```
.mp4  .avi  .mkv  .mov  .wmv  .flv  .webm  .m4v  .mpeg  .mpg  .3gp  .ts
```

---

## 🛠️ 开发

### 环境要求

- [Node.js](https://nodejs.org/) >= 18
- [Rust](https://www.rust-lang.org/) >= 1.70
- [Tauri CLI](https://tauri.app/)

### 本地开发

```bash
# 安装依赖
npm install

# 开发模式
npm run tauri dev

# 构建
npm run tauri build
```

### 技术栈

- **前端**: HTML / CSS / JavaScript
- **后端**: Rust + Tauri 2.0
- **文件监控**: notify-rs

---

## 📄 许可证

MIT License

---

<p align="center">
  Made with 💙 by Fidrua
</p>
