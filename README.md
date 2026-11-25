# allingo

Go 1.26 公共基础设施，使用 Fx 装配应用；业务模块按需引入，不要求使用 GORM、Redis 或事件总线。

## 应用与 HTTP 生命周期

原有 `app.Current().Option(...)` 和 `Run()` 保持可用。`RunContext(ctx) error` 同时等待调用方取消、SIGINT/SIGTERM 或 Fx `Shutdowner`；非零 Fx 退出码和停止错误通过返回值暴露。启动／停止预算使用 `fx.StartTimeout`、`fx.StopTimeout`，缺省各 15 秒。停止使用独立 context，调用方取消不会跳过清理；Fx 停止后沿用异步池与日志各最多五秒的退出等待。一个 App 实例只能运行一次，启动后不能再注册 option。

RunContext 自动注入 `app.RuntimeContext{Context context.Context}`，供长生命周期任务使用。收到调用方取消、系统信号或 Fx Shutdown 后，先取消该运行期 context，再调用任何 OnStop；即使首个 HTTP drain 耗尽整个停止预算，任务和后台 I/O 已经收到取消。启动失败同样撤销运行期 context。仅用 `fx.ValidateApp` 检查消费该类型的图时，测试显式提供 `fx.Supply(app.RuntimeContext{Context: context.Background()})`，生产运行不重复提供。

`common.Module` 仍提供 Gin 和 HTTP 生命周期。可选注入 `*common.HTTPConfig` 指定监听地址、完整 `http.Handler`、四个标准库超时及 TLS 文件；Handler 为 nil 时使用原 Gin engine。

```go
fx.Provide(func() *common.HTTPConfig {
    return &common.HTTPConfig{
        Addr:              ":8080",
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       0,
        WriteTimeout:      0,
        IdleTimeout:       2 * time.Minute,
        // Handler: hostRouter, // 可覆盖默认 Gin，按 Host 分发不同站点
        // TLSCertFile / TLSKeyFile: 两者同时提供才启用 TLS
        // PprofAddr: "127.0.0.1:18080", // 显式配置时空值关闭 pprof
    }
})
```

显式 HTTPConfig 要求非空 Addr，超时零值表示不设置限制。OnStart 验证 TLS 并同步绑定端口；证书或监听失败使启动失败，不留下部分监听。Serve 的意外错误触发 Fx 非零退出。OnStop 停止 pprof 并等待 HTTP 请求；预算耗尽时关闭活动连接、取消请求 context，并返回错误。

不注入 HTTPConfig 时保留旧 `HTTP_PORT`、`HTTP_CERT_FILE`、`HTTP_KEY_FILE`、`RUN_MODE`、`PPROF_PORT` 行为，包括非 release 模式的 pprof；该监听现在也在停止时关闭。新服务宜显式提供配置，避免因缺少 RUN_MODE 意外开放调试端口。HTTP 与任务、存储的 hook 顺序由依赖图决定；仅排列 Module 不能代替显式依赖。

## 原生 PostgreSQL、对象存储与任务

`common/db/pgx`（包名 `pgxdb`）提供原生 pgxpool：宿主向 `pgxdb.Module` 提供 `*pgxpool.Config`，Manager 构造不连接数据库，OnStart 建池并 Ping，失败时释放连接；成功后通过 `GetPool()` 交给基础设施层。OnStop 使用原生 Close，因此依赖它的 HTTP 和任务必须先停止、归还连接。`pgxdb.Open(ctx, config)` 也可用于拥有独立生命周期的可选连接池；不隐式注册 GORM 或执行迁移。

`common/oss` 的 `ClientConfig`、`NewClient` 和可选 Module 提供原生 `*minio.Client`；构造不检查 Bucket、不发网络请求。资源／日志等多个 Bucket 可共用该客户端。`NewManagerWithClient` 绑定具体 Bucket，提供 Context 版本的操作；旧环境变量、Manager 和无 Context 调用保留原用法。Bucket 检查、权限和迁移仍由使用方在合适的启动阶段负责。

`task.NewManagedManager(lifecycle, kv)` 构造不启动 cron，由 OnStart 启动。使用 `AddContext`／`MustAddContext` 注册 `func(context.Context) error`；同名任务最多有一个排队或执行实例。任务 context 独立于短期的 Fx OnStart context，在停止时才取消。`Stop(ctx)` 撤销新任务准入、停止 cron、取消任务 context，并等待实际异步任务退出；预算耗尽返回错误，已受理任务仍继续退出，后续 Stop 可继续等待。

应用应使用 `task.NewManagedManagerWithContext(lifecycle, kv, running.Context)` 绑定上述 RuntimeContext；父 context 取消后，正在执行的任务立即收到取消，Trigger、cron 准入和 AddContext 返回 ErrStopped，不再新增 KV／异步工作。旧 NewManagedManager 仍以 Background 为父 context，保持原调用方式。父 context 不能替代 Stop 的等待和调度器清理。

`Trigger(name)` 通过与 cron 相同的准入、非重入、context 和等待路径执行一次已注册任务，可用于启动后的首次清理。尚未启动、任务名不存在和停止后分别返回 ErrNotStarted、ErrTaskNotFound、ErrStopped；同名任务已在途时不重复排队。

原有 `NewManager(kv)` 仍立即启动，`Add`／`MustAdd` 仍接受无 Context 回调；这种旧回调不能被强制取消。使用托管构造时，应通过 Fx factory 显式依赖数据库 Manager，再注册任务 hook；HTTP 依赖任务，形成启动 `数据库 → 任务 → HTTP`、停止反向的顺序。任务回调必须遵守取消信号，不能假定 Stop 超时会终止任意 Go 函数。

## JSON 日志

`log.SetFormat(log.JSONFormat)` 在启动前启用全局 JSON 单行输出；默认 `log.TextFormat` 保持原有彩色文本。JSON 保留 time、level、msg、logger_name、go_id、error 和顶层结构字段，例如 stage、order_id、action。前述元数据名称为保留字段，普通 Fields 不应占用；error 类型的结构字段按错误文本序列化。无法编码的字段会明确向 stderr 报告格式错误。

JSON 与文本使用同一输出点，每条记录只写一次；已有 Hook 仍在输出后接收完整 Entry，切勿在 Hook 中向相同 stdout 再输出一份 JSON。`SetOutput` 保持逐 logger 行为，队列及日志退出 barrier 不变。环境变量包不再为缺少可选 `.env` 输出提示，其他加载错误仍会报告。

## 验证与发布

生命周期单元测试只创建短期本机回环监听并在测试内关闭；不需要数据库、Redis、MinIO 或开发服务器。原生连接及真实存储需要单独环境验证。

推送 main 后由现有 CI 运行测试，再自动创建版本标签和 Release；不要手工创建发布标签。跨仓库本地联调可临时使用 workspace／replace，正式依赖应在获得发布授权、CI 产生实际标签并核对提交后固定，不提交本机绝对路径替换。
