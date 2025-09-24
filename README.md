# NGCache - 基于FreeCache的扩展缓存库

NGCache是基于[coocood/freecache](https://github.com/coocood/freecache)开发的扩展缓存库，保持了FreeCache的zero-GC优化特性，同时增加了丰富的类型操作、序列化功能和持久化支持。

## 特性

- ✅ **Zero-GC优化**: 继承FreeCache的zero-GC特性，避免GC压力
- ✅ **丰富的类型支持**: 支持基础类型的直接操作
- ✅ **任意类型序列化**: 支持结构体等复杂类型的序列化存储
- ✅ **持久化功能**: 支持JSON和自定义二进制格式的持久化
- ✅ **永久缓存**: 支持expire=0的永久缓存
- ✅ **高性能**: 保持与FreeCache相同的高性能特性
- ✅ **线程安全**: 支持高并发访问

## 安装

```bash
go mod init your-project
go get github.com/coocood/freecache
```

## 快速开始

```go
package main

import (
    "fmt"
    "time"
    "your-project/ngcat"
)

func main() {
    // 创建持久化配置
    persistConfig := &ngcat.PersistConfig{
        Enabled:  true,
        FilePath: "./cache_data",
        FileName: "ngcache.json",
        Format:   ngcat.FormatJSON,
        Interval: 30 * time.Second,
    }

    // 创建NGCache实例（100MB缓存）
    cache := ngcat.NewNGCache(100*1024*1024, persistConfig)
    defer cache.Close()

    // 基础类型操作
    cache.SetString("username", "张三")
    username, _ := cache.GetString("username")
    fmt.Printf("用户名: %s\n", username)

    // 永久缓存
    cache.SetPermanent([]byte("key"), []byte("永久数据"))
    data, _ := cache.GetPermanent([]byte("key"))
    fmt.Printf("永久数据: %s\n", string(data))
}
```

## API文档

### 初始化

#### NewNGCache

```go
func NewNGCache(cacheSize int, persistConfig *PersistConfig) *NGCache
```

创建新的NGCache实例。

**参数:**
- `cacheSize`: 缓存大小（字节）
- `persistConfig`: 持久化配置，可为nil

**返回:**
- `*NGCache`: NGCache实例

### 基础类型操作

#### 整数类型

```go
// 32位整数
func (ng *NGCache) SetInt32(key string, value int32) error
func (ng *NGCache) GetInt32(key string) (int32, error)

// 64位整数
func (ng *NGCache) SetInt64(key string, value int64) error
func (ng *NGCache) GetInt64(key string) (int64, error)
```

#### 浮点数类型

```go
// 32位浮点数
func (ng *NGCache) SetFloat32(key string, value float32) error
func (ng *NGCache) GetFloat32(key string) (float32, error)

// 64位浮点数
func (ng *NGCache) SetFloat64(key string, value float64) error
func (ng *NGCache) GetFloat64(key string) (float64, error)
```

#### 布尔类型

```go
func (ng *NGCache) SetBool(key string, value bool) error
func (ng *NGCache) GetBool(key string) (bool, error)
```

#### 字符串和字节数组

```go
// 字符串
func (ng *NGCache) SetString(key string, value string) error
func (ng *NGCache) GetString(key string) (string, error)

// 字节数组
func (ng *NGCache) SetBytes(key string, value []byte) error
func (ng *NGCache) GetBytes(key string) ([]byte, error)
```

### 序列化操作

#### 任意类型序列化（Gob）

```go
func (ng *NGCache) SetAny(key string, value interface{}) error
func (ng *NGCache) GetAny(key string, value interface{}) error
```

使用Go的gob包进行序列化，支持大部分Go类型。

**示例:**
```go
type User struct {
    ID   int
    Name string
}

user := User{ID: 1, Name: "张三"}
cache.SetAny("user", user)

var retrievedUser User
err := cache.GetAny("user", &retrievedUser)
```

#### JSON序列化

```go
func (ng *NGCache) SetJSON(key string, value interface{}) error
func (ng *NGCache) GetJSON(key string, value interface{}) error
```

使用JSON格式进行序列化，支持跨语言兼容。

#### 智能结构体序列化

```go
func (ng *NGCache) SetStruct(key string, value interface{}) error
func (ng *NGCache) GetStruct(key string, value interface{}) error
```

自动选择最适合的序列化方式（Gob或JSON）。

### 永久缓存

```go
func (ng *NGCache) SetPermanent(key []byte, value []byte) error
func (ng *NGCache) GetPermanent(key []byte) ([]byte, error)
```

设置和获取永久缓存（expire=0），数据不会过期。

### 持久化配置

```go
type PersistConfig struct {
    Enabled  bool          // 是否启用持久化
    FilePath string        // 持久化文件路径
    FileName string        // 持久化文件名
    Format   PersistFormat // 持久化格式
    Interval time.Duration // 持久化间隔
}

type PersistFormat int

const (
    FormatJSON   PersistFormat = iota // JSON格式
    FormatBinary                      // 二进制格式
)
```

**示例配置:**
```go
persistConfig := &ngcat.PersistConfig{
    Enabled:  true,
    FilePath: "./cache_data",
    FileName: "ngcache.bin",
    Format:   ngcat.FormatBinary,
    Interval: 60 * time.Second,
}
```

### 缓存管理

```go
// 关闭缓存并保存持久化数据
func (ng *NGCache) Close() error
```

## 持久化格式

### JSON格式

JSON格式便于调试和跨语言兼容：

```json
{
  "version": 1,
  "timestamp": 1640995200,
  "entries": [
    {
      "key": "username",
      "value": "5byg5LiJ" // base64编码的值
    }
  ]
}
```

### 二进制格式

自定义二进制格式，空间效率更高：

```
[魔数:4字节][版本:4字节][时间戳:8字节][条目数:4字节]
[键长度:4字节][键数据:N字节][值长度:4字节][值数据:M字节]...
```

- 魔数: 0x4E474341 ("NGCA")
- 版本: 当前为1
- 所有多字节数据使用小端序

## 性能特性

### Zero-GC优化

NGCache继承了FreeCache的zero-GC特性：

- 使用环形缓冲区减少指针数量
- 数据分片存储，降低锁竞争
- 避免频繁的内存分配和回收

### 性能基准

基于FreeCache的性能测试结果：

```
BenchmarkCacheSet-8    10000000    251 ns/op
BenchmarkCacheGet-8    20000000    113 ns/op
```

## 线程安全

NGCache是完全线程安全的：

- 底层FreeCache提供线程安全保证
- 持久化操作使用独立的互斥锁
- 支持高并发读写操作

## 错误处理

```go
// 常见错误
var (
    ErrNotFound     = errors.New("key not found")
    ErrKeyTooLarge  = errors.New("key too large")
    ErrValueTooLarge = errors.New("value too large")
)
```

## 最佳实践

### 1. 缓存大小设置

```go
// 根据可用内存设置合适的缓存大小
cacheSize := 512 * 1024 * 1024 // 512MB
cache := ngcat.NewNGCache(cacheSize, nil)
```

### 2. 持久化配置

```go
// 生产环境推荐使用二进制格式
persistConfig := &ngcat.PersistConfig{
    Enabled:  true,
    FilePath: "/var/cache/app",
    FileName: "cache.bin",
    Format:   ngcat.FormatBinary,
    Interval: 5 * time.Minute,
}
```

### 3. 优雅关闭

```go
func main() {
    cache := ngcat.NewNGCache(cacheSize, persistConfig)
    
    // 注册信号处理
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        cache.Close() // 优雅关闭
        os.Exit(0)
    }()
    
    // 应用逻辑...
}
```

## 许可证

本项目基于MIT许可证开源。

## 贡献

欢迎提交Issue和Pull Request来改进这个项目。

## 致谢

感谢[coocood/freecache](https://github.com/coocood/freecache)项目提供的优秀基础实现。