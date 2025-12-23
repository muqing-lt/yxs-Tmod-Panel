🌲yxs-Tmod Panel

yxs-Tmod Panel 是一个专为 Terraria (tModLoader) 服务器设计的轻量级、现代化网页管理面板。

它旨在解决传统命令行开服的繁琐，提供了一个可视化的仪表盘来监控性能、管理玩家、修改配置以及实时查看日志。后端采用 Go (Gin) 编写，前端采用 Vue 3 (Naive UI)，最终编译为单文件运行，部署极简。

<img width="2471" height="1342" alt="image" src="https://github.com/user-attachments/assets/4919604a-6ea6-4fd4-9ff9-eba36d9065ee" />


✨ 核心功能

📊 实时监控仪表盘

可视化展示 CPU、内存、磁盘占用及在线人数。

实时检测 tModLoader 进程状态。

一键启动、停止、重启服务器（支持优雅关闭）。

💻 Web 终端控制台

基于 WebSocket 的实时日志流，低延迟回显。

支持网页端直接发送服务端指令（如 save, say, dawn 等）。

内置常用指令快捷键。

👥 高级玩家管理

实时列表：自动同步在线玩家信息。

可视化操作：右键/按钮菜单支持 踢出 (Kick)、封禁 (Ban)。

权限集成：深度集成 Hero's Mod，支持网页端一键授予/撤销管理员权限。

动态日志：记录玩家进出、死亡及操作的时间轴。

⚙️ 配置文件编辑器

图形化修改启动参数（端口、最大人数、世界名称、难度等）。

支持直接读写 serverconfig.txt，自动补全绝对路径。

🛠️ 自动化运维

环境自检：自动检测并下载安装 .NET 8.0 运行环境。

Mod 管理：支持一键安装/更新 Hero's Mod。

防卡死：包含进程无响应强制查杀机制。

🏗️ 技术栈

后端: Golang, Gin (Web框架), Gorilla WebSocket, Gopsutil (系统监控)

前端: Vue 3, TypeScript, Vite, Naive UI, Tailwind CSS

架构: 前后端分离开发，最终通过 go:embed 打包为独立二进制文件。

🚀 快速开始

1. 运行环境要求

Linux (推荐 Ubuntu/Debian/CentOS)

或者 Windows (WSL2)

网络畅通 (用于下载 tModLoader 和 .NET)

2. 编译部署

如果你想自己编译本项目：

前端构建:

cd frontend
npm install
npm run build
# 构建产物将生成在 frontend/dist 目录


后端构建:

# 确保 frontend/dist 目录存在
go mod tidy
go build -o terra-panel main.go


3. 运行

直接运行编译好的二进制文件：

chmod +x terra-panel
./terra-panel


访问浏览器：http://localhost:8080

首次启动: 会自动进入初始化设置页面，请创建一个管理员账号。

配置文件: 账号信息将存储在 config.yaml 中。

📂 目录结构说明

.
├── server/              # tModLoader 服务端及存档数据 (自动生成)
│   ├── storage/         # 存档和模组目录
│   ├── dotnet/          # .NET 运行环境
│   └── tModLoaderServer # 服务端核心
├── config.yaml          # 面板配置文件 (用户账号)
├── terra-panel          # 面板主程序


📝 常见问题

Q: 如何让服务器支持 Hero's Mod 权限管理？
A: 在面板的“权限管理”页面，点击“一键安装 Hero's Mod”。安装完成后重启服务器即可。在玩家列表中点击“管理”即可授予管理员权限。

Q: 存档在哪里？
A: 默认位于运行目录下的 ./server/storage/Worlds。你可以在“房间设置”中修改路径。

🤝 贡献

欢迎提交 Issue 或 Pull Request 来改进这个项目！

📄 开源协议

MIT License
