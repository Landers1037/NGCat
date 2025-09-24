package ngcat

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// PersistEntry 持久化条目
type PersistEntry struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// PersistData 持久化数据结构
type PersistData struct {
	Version   int             `json:"version"`
	Timestamp int64           `json:"timestamp"`
	Entries   []PersistEntry  `json:"entries"`
}

// 二进制格式常量
const (
	// BinaryMagic 二进制文件魔数
	BinaryMagic = 0x4E474341 // "NGCA"
	// BinaryVersion 二进制格式版本
	BinaryVersion = 1
)

// persistRoutine 持久化协程
func (ng *NGCache) persistRoutine() {
	if ng.persistConfig == nil || !ng.persistConfig.Enabled {
		return
	}

	ticker := time.NewTicker(ng.persistConfig.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ng.saveToPersist()
		case <-ng.stopChan:
			return
		}
	}
}

// saveToPersist 保存到持久化文件
func (ng *NGCache) saveToPersist() error {
	if ng.persistConfig == nil || !ng.persistConfig.Enabled {
		return nil
	}

	ng.persistMutex.Lock()
	defer ng.persistMutex.Unlock()

	// 确保目录存在
	dir := ng.persistConfig.FilePath
	if dir == "" {
		dir = "."
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("创建持久化目录失败: %v", err)
	}

	// 构建完整文件路径
	filePath := filepath.Join(dir, ng.persistConfig.FileName)

	// 收集持久化数据
	ng.persistDataMutex.RLock()
	entries := make([]PersistEntry, 0, len(ng.persistData))
	for key, value := range ng.persistData {
		entries = append(entries, PersistEntry{
			Key:   key,
			Value: value,
		})
	}
	ng.persistDataMutex.RUnlock()

	persistData := PersistData{
		Version:   1,
		Timestamp: time.Now().Unix(),
		Entries:   entries,
	}

	// 根据格式保存
	switch ng.persistConfig.Format {
	case FormatJSON:
		return ng.saveToJSON(filePath, &persistData)
	case FormatBinary:
		return ng.saveToBinary(filePath, &persistData)
	default:
		return fmt.Errorf("不支持的持久化格式: %d", ng.persistConfig.Format)
	}
}

// saveToJSON 保存为JSON格式
func (ng *NGCache) saveToJSON(filePath string, data *PersistData) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建JSON文件失败: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// saveToBinary 保存为二进制格式
func (ng *NGCache) saveToBinary(filePath string, data *PersistData) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建二进制文件失败: %v", err)
	}
	defer file.Close()

	// 写入魔数
	err = binary.Write(file, binary.LittleEndian, uint32(BinaryMagic))
	if err != nil {
		return err
	}

	// 写入版本
	err = binary.Write(file, binary.LittleEndian, uint32(BinaryVersion))
	if err != nil {
		return err
	}

	// 写入时间戳
	err = binary.Write(file, binary.LittleEndian, data.Timestamp)
	if err != nil {
		return err
	}

	// 写入条目数量
	err = binary.Write(file, binary.LittleEndian, uint32(len(data.Entries)))
	if err != nil {
		return err
	}

	// 写入每个条目
	for _, entry := range data.Entries {
		// 写入键长度和键
		err = binary.Write(file, binary.LittleEndian, uint32(len(entry.Key)))
		if err != nil {
			return err
		}
		_, err = file.Write([]byte(entry.Key))
		if err != nil {
			return err
		}

		// 写入值长度和值
		err = binary.Write(file, binary.LittleEndian, uint32(len(entry.Value)))
		if err != nil {
			return err
		}
		_, err = file.Write(entry.Value)
		if err != nil {
			return err
		}
	}

	return nil
}

// loadFromPersist 从持久化文件加载
func (ng *NGCache) loadFromPersist() error {
	if ng.persistConfig == nil || !ng.persistConfig.Enabled {
		return nil
	}

	ng.persistMutex.Lock()
	defer ng.persistMutex.Unlock()

	// 构建完整文件路径
	dir := ng.persistConfig.FilePath
	if dir == "" {
		dir = "."
	}
	filePath := filepath.Join(dir, ng.persistConfig.FileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，不是错误
	}

	// 根据格式加载
	switch ng.persistConfig.Format {
	case FormatJSON:
		return ng.loadFromJSON(filePath)
	case FormatBinary:
		return ng.loadFromBinary(filePath)
	default:
		return fmt.Errorf("不支持的持久化格式: %d", ng.persistConfig.Format)
	}
}

// loadFromJSON 从JSON格式加载
func (ng *NGCache) loadFromJSON(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开JSON文件失败: %v", err)
	}
	defer file.Close()

	var data PersistData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return fmt.Errorf("解析JSON文件失败: %v", err)
	}

	// 加载数据到内存
	ng.persistDataMutex.Lock()
	for _, entry := range data.Entries {
		ng.persistData[entry.Key] = entry.Value
		// 同时加载到freecache（永久缓存）
		ng.cache.Set([]byte(entry.Key), entry.Value, 0)
	}
	ng.persistDataMutex.Unlock()

	return nil
}

// loadFromBinary 从二进制格式加载
func (ng *NGCache) loadFromBinary(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开二进制文件失败: %v", err)
	}
	defer file.Close()

	// 读取魔数
	var magic uint32
	err = binary.Read(file, binary.LittleEndian, &magic)
	if err != nil {
		return fmt.Errorf("读取魔数失败: %v", err)
	}
	if magic != BinaryMagic {
		return fmt.Errorf("无效的二进制文件魔数: 0x%X", magic)
	}

	// 读取版本
	var version uint32
	err = binary.Read(file, binary.LittleEndian, &version)
	if err != nil {
		return fmt.Errorf("读取版本失败: %v", err)
	}
	if version != BinaryVersion {
		return fmt.Errorf("不支持的二进制文件版本: %d", version)
	}

	// 读取时间戳
	var timestamp int64
	err = binary.Read(file, binary.LittleEndian, &timestamp)
	if err != nil {
		return fmt.Errorf("读取时间戳失败: %v", err)
	}

	// 读取条目数量
	var entryCount uint32
	err = binary.Read(file, binary.LittleEndian, &entryCount)
	if err != nil {
		return fmt.Errorf("读取条目数量失败: %v", err)
	}

	// 读取每个条目
	ng.persistDataMutex.Lock()
	for i := uint32(0); i < entryCount; i++ {
		// 读取键长度
		var keyLen uint32
		err = binary.Read(file, binary.LittleEndian, &keyLen)
		if err != nil {
			ng.persistDataMutex.Unlock()
			return fmt.Errorf("读取键长度失败: %v", err)
		}

		// 读取键
		keyBytes := make([]byte, keyLen)
		_, err = io.ReadFull(file, keyBytes)
		if err != nil {
			ng.persistDataMutex.Unlock()
			return fmt.Errorf("读取键失败: %v", err)
		}

		// 读取值长度
		var valueLen uint32
		err = binary.Read(file, binary.LittleEndian, &valueLen)
		if err != nil {
			ng.persistDataMutex.Unlock()
			return fmt.Errorf("读取值长度失败: %v", err)
		}

		// 读取值
		valueBytes := make([]byte, valueLen)
		_, err = io.ReadFull(file, valueBytes)
		if err != nil {
			ng.persistDataMutex.Unlock()
			return fmt.Errorf("读取值失败: %v", err)
		}

		// 存储到内存
		key := string(keyBytes)
		ng.persistData[key] = valueBytes
		// 同时加载到freecache（永久缓存）
		ng.cache.Set(keyBytes, valueBytes, 0)
	}
	ng.persistDataMutex.Unlock()

	return nil
}