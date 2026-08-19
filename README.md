# 长租房源与租约管理平台

Go REST API + Vue 3 管理前端，用于管理房源、租客、租约、账单、收款、押金、维修工单、审批提醒和退租结算。

## Docker 启动

直接执行：

```powershell
docker compose up --build
```

默认端口：

- 管理前端：`http://localhost:28080`
- Go API：`http://localhost:28081`
- MySQL：`localhost:23306`
- Redis：`localhost:26379`

容器内部仍通过 `mysql:3306`、`redis:6379` 和 `api:8080` 通信，宿主机端口变化不会影响服务发现。

本地首次启动会创建验收管理员 `admin / Admin123!`。该账号仅用于本地验收，生产环境必须通过环境变量覆盖并立即修改密码。

可复制 `.env.example` 为 `.env` 覆盖默认端口和凭据；不复制也可以直接启动。

## 本地验证

```powershell
go mod tidy
gofmt -w cmd internal
go build ./...
go test ./...
go vet ./...
./scripts/verify-go-size.ps1
cd web
npm install
npm run typecheck
npm run build
```
