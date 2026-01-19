# 广播站点歌系统

这是一个基于 Gin 框架开发的广播站点歌系统，允许用户提交点歌请求，管理员可以管理这些请求。

## 功能特性

- 用户友好的点歌界面
- 实时显示歌曲列表及状态
- 管理员后台控制系统
- 歌曲状态管理（待处理、播放中、已播放）
- 响应式设计，支持移动设备

## 技术栈

- Go (Golang)
- Gin Web Framework
- HTML/CSS/JavaScript
- Bootstrap 5
- RESTful API

## 项目结构

```
SmartRadio/
├── main.go                 # 主程序入口
├── go.mod                  # Go 模块定义
├── go.sum                  # Go 模块校验和
├── templates/              # HTML 模板文件
│   ├── index.html          # 用户点歌页面
│   └── admin.html          # 管理员控制页面
└── static/                 # 静态资源
    └── css/
        └── style.css       # 自定义样式
```

## API 接口

### 获取所有歌曲
- `GET /api/songs` - 获取所有歌曲
- `GET /api/songs?status=pending` - 按状态过滤歌曲

### 添加歌曲
- `POST /api/songs` - 添加新的点歌请求
- 请求体: `{ "title": "歌曲名", "artist": "歌手", "requester": "点歌人" }`

### 更新歌曲状态
- `PUT /api/songs/:id/status` - 更新歌曲状态
- 请求体: `{ "status": "pending|playing|played" }`

### 删除歌曲
- `DELETE /api/songs/:id` - 删除指定歌曲

## 使用方法

1. 启动服务器：
   ```bash
   go run main.go
   ```

2. 访问以下页面：
   - 用户点歌页面：http://localhost:8080
   - 管理员控制台：http://localhost:8080/admin

## 环境要求

- Go 1.18 或更高版本
- 支持现代浏览器（Chrome, Firefox, Safari, Edge）

## 开发说明

本系统使用内存存储数据，重启后数据会丢失。在生产环境中，建议集成数据库（如 MySQL, PostgreSQL 或 MongoDB）。

## 许可证

MIT License