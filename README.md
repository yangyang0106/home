# Home Decision Calculator

当前目录现在有两套内容：

- `frontend/`：移动端 H5，调用后端 REST API
- `backend/`：Go 服务端骨架，负责权重、房源和评分输出

## 启动后端

```bash
cd /Users/ruigu/Documents/home/backend
go run ./cmd/server
```

如果你要直连本地 MySQL，先设置环境变量：

```bash
export APP_STORE_MODE=mysql
export APP_MYSQL_DSN='root:NewPassword123!@tcp(127.0.0.1:3306)/home_decision?parseTime=true&charset=utf8mb4'
```

## 打开前端

直接打开 [frontend/index.html](/Users/ruigu/Documents/home/frontend/index.html) 即可，或者用任意静态文件服务托管 `frontend/`。

## 说明

- 当前后端支持 `memory` 和 `mysql` 两种模式
- `backend/migrations/001_init.sql` 已经给出 MySQL 表结构
- `backend/internal/store/mysql.go` 已经是真实 MySQL 存储实现
