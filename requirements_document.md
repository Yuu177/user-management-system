# 用户管理系统需求文档

## 内容要求

实现一个用户管理系统，用户可以登录、拉取和编辑他们的 profiles。

用户可以通过在 Web 页面输入username 和 password 登录，backend 系统负责校验用户身份。成功登录后，页面需要展示用户的相关信息，否则页面展示相关错误。

成功登录后，用户可以编辑以下内容：

1. 上传 profile picture

2. 修改 nickname

用户信息包括：

1. username（不可更改）

2. nickname

3. profile picture

需要提前将初始用户数据插入数据库用于测试。确保测试数据库中包含 10,000,000 条用户账号信息。

## 设计要求

- 分别实现 HTTP server 和 TCP server，主要的功能逻辑放在 TCP server 实现
- Backend 鉴权逻辑需要在 TCP server 实现
- 用户账号信息必须存储在 MySQL 数据库。通过 MySQL Go client 连接数据库
- 使用基于 Auth/Session Token 的鉴权机制，Token 存储在 Redis，避免使用 JWT 等加密的形式
- TCP server 需要提供 RPC API，RPC 机制希望自己设计实现
- Web server 不允许直连 MySQL、Redis。所有 HTTP 请求只处理 API 和用户输入，具体的功能逻辑和数据库操作，需要通过 RPC 请求 TCP server 完成
- 尽可能使用 Go 标准库
- 安全性
- 鲁棒性
- 性能

## 验收标准

正确性：

- 必须完整实现相关 API，不能有明显 BUG
- 实现细节必须满足设计要求

安全性：

- 不能有安全问题（如：sessionid 不能轻易被破解和模拟）

鲁棒性：

- 服务不能因为客户端请求 crash

性能：

- 数据库必须有 10,000,000 条用户账号信息
- 必须确保返回结果是正确的
- 每个请求都要包含 RPC 调用以及 MySQL 或 Redis 访问
- 200 并发（固定用户）情况下，HTTP API QPS 大于 3000
  - 200 个 client（200 条 TCP 连接），每个 client 模拟一个用户（因此需要 200 个不同的固定用户账号）
- 200 并发（随机用户）情况下，HTTP API QPS 大于 1000
  - 200 个 client（200 条 TCP 连接），每个 client 每次随机选取一个用户，发起请求（如果涉及到鉴权，可以使用一个测试用的 token）
- 2000 并发（固定用户）情况下，HTTP API QPS 大于 1500
- 2000 并发（随机用户）情况下，HTTP API QPS 大于 800

代码规范：

- 通过 golint
- 通过 go vet
- 尽可能遵循 [Effective Go](https://golang.org/doc/effective_go.html)

代码质量：

- 易读
- 依赖清晰
- 尽量解耦
- 尽可能覆盖单元测试