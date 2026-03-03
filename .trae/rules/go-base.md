**Go 开发规范**

## 1. 目录结构

```
project/
├─ cmd/            // 程序入口
├─ internal/       // 业务代码
├─ pkg/            // 可复用工具（可选）
├─ go.mod
```

规则：

* `main.go` 只做启动与依赖注入
* 业务逻辑放 `internal`

## 2. 命名

### 包名

* 小写
* 单词
* 不使用下划线

```
user
auth
config
```

### 导出规则

* 大写开头 → 对外可见
* 小写 → 私有

## 3. 错误处理

必须检查：

```
if err != nil {
    return err
}
```

禁止忽略：

```
_ = fn()
```

需要上下文信息时：

```
return fmt.Errorf("load config: %w", err)
```

## 4. Context

* 作为第一个参数
* 不传 `nil`
* 不存入结构体

```
func (s *Service) Do(ctx context.Context) error
```

## 5. 并发

* goroutine 必须可退出
* 使用：

  * `context`
  * `WaitGroup`
  * `channel`

## 6. 日志

* 不使用 `fmt.Println` 作为运行日志
* 统一日志库

## 7. 格式与检查

必须执行：

```
go fmt ./...
go vet ./...
```

建议：

```
golangci-lint run
```

## 8. 依赖

仅使用：

```
go mod tidy
```

禁止提交无用依赖。
