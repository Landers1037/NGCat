package ngcat

import (
	"encoding/binary"
	"unsafe"
)

// SetInt32 设置int32类型值
func (ng *NGCache) SetInt32(key string, value int32, expireSeconds int) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(value))
	return ng.setWithPersist(key, buf, expireSeconds)
}

// GetInt32 获取int32类型值
func (ng *NGCache) GetInt32(key string) (int32, error) {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return 0, err
	}
	if len(data) != 4 {
		return 0, ErrInvalidType
	}
	return int32(binary.LittleEndian.Uint32(data)), nil
}

// SetInt64 设置int64类型值
func (ng *NGCache) SetInt64(key string, value int64, expireSeconds int) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return ng.setWithPersist(key, buf, expireSeconds)
}

// GetInt64 获取int64类型值
func (ng *NGCache) GetInt64(key string) (int64, error) {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, ErrInvalidType
	}
	return int64(binary.LittleEndian.Uint64(data)), nil
}

// SetBool 设置bool类型值
func (ng *NGCache) SetBool(key string, value bool, expireSeconds int) error {
	var buf []byte
	if value {
		buf = []byte{1}
	} else {
		buf = []byte{0}
	}
	return ng.setWithPersist(key, buf, expireSeconds)
}

// GetBool 获取bool类型值
func (ng *NGCache) GetBool(key string) (bool, error) {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return false, err
	}
	if len(data) != 1 {
		return false, ErrInvalidType
	}
	return data[0] == 1, nil
}

// SetFloat32 设置float32类型值
func (ng *NGCache) SetFloat32(key string, value float32, expireSeconds int) error {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, *(*uint32)(unsafe.Pointer(&value)))
	return ng.setWithPersist(key, buf, expireSeconds)
}

// GetFloat32 获取float32类型值
func (ng *NGCache) GetFloat32(key string) (float32, error) {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return 0, err
	}
	if len(data) != 4 {
		return 0, ErrInvalidType
	}
	uintVal := binary.LittleEndian.Uint32(data)
	return *(*float32)(unsafe.Pointer(&uintVal)), nil
}

// SetFloat64 设置float64类型值
func (ng *NGCache) SetFloat64(key string, value float64, expireSeconds int) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, *(*uint64)(unsafe.Pointer(&value)))
	return ng.setWithPersist(key, buf, expireSeconds)
}

// GetFloat64 获取float64类型值
func (ng *NGCache) GetFloat64(key string) (float64, error) {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return 0, err
	}
	if len(data) != 8 {
		return 0, ErrInvalidType
	}
	uintVal := binary.LittleEndian.Uint64(data)
	return *(*float64)(unsafe.Pointer(&uintVal)), nil
}

// SetBytes 设置字节数组值
func (ng *NGCache) SetBytes(key string, value []byte, expireSeconds int) error {
	return ng.setWithPersist(key, value, expireSeconds)
}

// GetBytes 获取字节数组值
func (ng *NGCache) GetBytes(key string) ([]byte, error) {
	return ng.getWithPersist(key)
}

// SetString 设置字符串值
func (ng *NGCache) SetString(key string, value string, expireSeconds int) error {
	return ng.setWithPersist(key, []byte(value), expireSeconds)
}

// GetString 获取字符串值
func (ng *NGCache) GetString(key string) (string, error) {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// setWithPersist 内部设置方法，支持持久化
func (ng *NGCache) setWithPersist(key string, value []byte, expireSeconds int) error {
	// 如果是永久缓存（expireSeconds <= 0），存储到持久化数据中
	if expireSeconds <= 0 {
		ng.persistDataMutex.Lock()
		ng.persistData[key] = make([]byte, len(value))
		copy(ng.persistData[key], value)
		ng.persistDataMutex.Unlock()
	}

	// 同时存储到freecache中
	return ng.cache.Set([]byte(key), value, expireSeconds)
}

// getWithPersist 内部获取方法，支持持久化
func (ng *NGCache) getWithPersist(key string) ([]byte, error) {
	// 首先尝试从freecache获取
	value, err := ng.cache.Get([]byte(key))
	if err == nil {
		return value, nil
	}

	// 如果freecache中没有，尝试从持久化数据获取
	ng.persistDataMutex.RLock()
	persistValue, exists := ng.persistData[key]
	ng.persistDataMutex.RUnlock()

	if exists {
		// 将持久化数据重新加载到freecache中（永久缓存）
		ng.cache.Set([]byte(key), persistValue, 0)
		return persistValue, nil
	}

	return nil, ErrKeyNotFound
}
