# task133-structload — Benzhi 评测说明

## 业务问题

建筑结构荷载规范合规引擎：面向建筑结构设计与审图，对一个建筑的构件（梁/柱/板）做 ASCE 7 规范荷载计算与充分性合规检查。引擎维护 项目 → 楼层 → 构件 → 荷载工况 → 荷载组合 → 充分性检查 的完整闭环，并在进程重启后能从权威输入重放出逐值一致的派生荷载（风压 W、雪载 S、折减活载 L）与检查结果。全部数值用定点整数表示（压力/面载 0.01 kPa、集中力 0.01 kN、弯矩 0.01 kN·m、长度 mm、面积 cm²、风速 0.1 m/s、充分性比 UR 0.01），无浮点对外接口。

主要输入：项目风/雪参数与构件几何/名义承载力、手动荷载工况（D/L/Lr/R/E）、LRFD 荷载组合系数集。
主要输出：派生荷载工况、每个构件×组合的充分性比 UR 与 pass/marginal/fail 状态、合规摘要、只读设计复核报告、调整后重算、重启一致性审计。

## 标准本地命令

```bash
go build ./...                 # 编译全部包
go run . --addr=:8080 --db=structload.db   # 启动 HTTP 服务（前端在 /）
go test ./...                  # 运行全部单元测试
go run . --smoke-test          # 自检：8 个场景执行后自行退出
go run . --migrate-only        # 仅建表后退出
```

服务启动后浏览器访问 `http://localhost:8080/` 即可操作前端（项目→楼层→构件→录工况→派生→组合→运行检查→查看 UR/pass-fail→合规摘要→调整重算）。

## Docker 构建

`build_benzhi_docker.sh <镜像名> <平台>` 仅通过根目录 `benzhi.Dockerfile` 构建。前端为原生 HTML/CSS/JS，经 `//go:embed` 打入 Go 二进制，无需 Node 构建步骤；镜像内无独立前端产物目录。

```bash
# amd64
bash ./build_benzhi_docker.sh go-task-benzhi:amd64 linux/amd64
docker run --rm go-task-benzhi:amd64 go version            # 期望 go1.26.3 linux/amd64
docker run --rm go-task-benzhi:amd64 bash -lc 'cd /app && go build . && ./app --smoke-test --db=/tmp/sl.db'

# arm64
bash ./build_benzhi_docker.sh go-task-benzhi:arm64 linux/arm64
docker run --rm go-task-benzhi:arm64 go version
docker run --rm go-task-benzhi:arm64 bash -lc 'cd /app && go build . && ./app --smoke-test --db=/tmp/sl.db'
```

进入容器交互：`docker run -it go-task-benzhi:amd64`

## 双架构验证

```bash
docker buildx build --platform linux/amd64 --load -t go-task-check:amd64 -f Dockerfile .
docker run --rm go-task-check:amd64 --smoke-test
docker buildx build --platform linux/arm64 --load -t go-task-check:arm64 -f Dockerfile .
docker run --rm go-task-check:arm64 --smoke-test
```

## Smoke-test（同时验证页面与业务 API）

```bash
go run . --smoke-test --db=/tmp/structload-smoke.db
```

该自检覆盖：建项目 → 加楼层构件 → 录手动工况 → 派生风/雪/折减活载 → 建 LRFD 组合 → 运行检查得 UR 与 pass/marginal/fail → 调整承载力后重算 UR 下降 → 重启 LoadAll + ReconcileAll 结果逐值一致 → 合规摘要计数 → 边界（小面积不折减活载、风面 Cp 取值、雪载坡度递减）。全部断言为定点整数精确比较，无浮点容差。

## 技术栈

- Go 1.26.3（`go.mod` 指令 `go 1.26.3`，`GOTOOLCHAIN=local`，`CGO_ENABLED=0`）
- 持久化：SQLite（纯 Go 驱动 `modernc.org/sqlite v1.52.0`，对应 SQLite 3.46.1），WAL + 事务 + ReconcileAll 重算恢复
- 前端：原生 HTML/CSS/JavaScript（无构建、无 Node、无 npm），`//go:embed web` 打入二进制
- 组件版本登记：`env/component-versions.json`（go=1.26.3, sqlite=3.46.1，与 `docs/component_versions.lock.json` 一致）
