package ngcat

import (
	"errors"
	"sync"
	"time"

	"github.com/coocood/freecache"
)

// PersistFormat 持久化格式类型
type PersistFormat int

const (
	// FormatJSON JSON格式持久化
	FormatJSON PersistFormat = iota
	// FormatBinary 自定义二进制格式持久化
	FormatBinary
)

// PersistConfig 持久化配置
type PersistConfig struct {
	// Enabled 是否启用持久化
	Enabled bool
	// FilePath 持久化文件路径
	FilePath string
	// FileName 持久化文件名
	FileName string
	// Format 持久化格式
	Format PersistFormat
	// Interval 持久化间隔时间
	Interval time.Duration
}

// NGCache 扩展缓存库
type NGCache struct {
	// cache freecache实例
	cache *freecache.Cache
	// persistConfig 持久化配置
	persistConfig *PersistConfig
	// persistMutex 持久化操作互斥锁
	persistMutex sync.RWMutex
	// stopChan 停止持久化的通道
	stopChan chan struct{}
	// persistData 永久缓存数据（Expire=0的数据）
	persistData map[string][]byte
	// persistDataMutex 永久数据互斥锁
	persistDataMutex sync.RWMutex
}

// NewNGCache 创建新的扩展缓存实例
func NewNGCache(size int, config *PersistConfig) *NGCache {
	ng := &NGCache{
		cache:         freecache.NewCache(size),
		persistConfig: config,
		stopChan:      make(chan struct{}),
		persistData:   make(map[string][]byte),
	}

	// 如果启用持久化，先加载数据，然后启动持久化协程
	if config != nil && config.Enabled {
		// 加载持久化数据
		ng.loadFromPersist()
		// 启动持久化协程
		go ng.persistRoutine()
	}

	return ng
}

// Close 关闭缓存并执行最后一次持久化
func (ng *NGCache) Close() error {
	if ng.persistConfig != nil && ng.persistConfig.Enabled {
		close(ng.stopChan)
		return ng.saveToPersist()
	}
	return nil
}

// SetPermanent 设置永久缓存（expire=0）
func (ng *NGCache) SetPermanent(key []byte, value []byte) error {
	// 设置到freecache（永久缓存）
	err := ng.cache.Set(key, value, 0)
	if err != nil {
		return err
	}

	// 如果启用持久化，同时保存到持久化数据
	if ng.persistConfig != nil && ng.persistConfig.Enabled {
		ng.persistDataMutex.Lock()
		ng.persistData[string(key)] = value
		ng.persistDataMutex.Unlock()
	}

	return nil
}

// GetPermanent 获取永久缓存
func (ng *NGCache) GetPermanent(key []byte) ([]byte, error) {
	// 首先尝试从freecache获取
	value, err := ng.cache.Get(key)
	if err == nil {
		return value, nil
	}

	// 如果freecache中没有，尝试从持久化数据获取
	if ng.persistConfig != nil && ng.persistConfig.Enabled {
		ng.persistDataMutex.RLock()
		value, exists := ng.persistData[string(key)]
		ng.persistDataMutex.RUnlock()
		if exists {
			// 重新加载到freecache
			ng.cache.Set(key, value, 0)
			return value, nil
		}
	}

	return nil, err
}

// 常见错误定义
var (
	ErrKeyNotFound = errors.New("key not found")
	ErrInvalidType = errors.New("invalid type")
)
