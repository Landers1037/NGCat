package ngcat

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"reflect"
)

// SetAny 设置任意类型值（使用gob序列化）
func (ng *NGCache) SetAny(key string, value interface{}, expireSeconds int) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(value)
	if err != nil {
		return err
	}
	return ng.setWithPersist(key, buf.Bytes(), expireSeconds)
}

// GetAny 获取任意类型值（使用gob反序列化）
func (ng *NGCache) GetAny(key string, value interface{}) error {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(value)
}

// SetJSON 设置任意类型值（使用JSON序列化）
func (ng *NGCache) SetJSON(key string, value interface{}, expireSeconds int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return ng.setWithPersist(key, data, expireSeconds)
}

// GetJSON 获取任意类型值（使用JSON反序列化）
func (ng *NGCache) GetJSON(key string, value interface{}) error {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

// SetStruct 设置结构体（自动选择最优序列化方式）
func (ng *NGCache) SetStruct(key string, value interface{}, expireSeconds int) error {
	// 检查类型是否可以用gob序列化
	if ng.canUseGob(value) {
		return ng.SetAny(key, value, expireSeconds)
	}
	// 否则使用JSON
	return ng.SetJSON(key, value, expireSeconds)
}

// GetStruct 获取结构体（自动选择反序列化方式）
func (ng *NGCache) GetStruct(key string, value interface{}) error {
	data, err := ng.getWithPersist(key)
	if err != nil {
		return err
	}

	// 尝试gob反序列化
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(value)
	if err == nil {
		return nil
	}

	// 如果gob失败，尝试JSON
	return json.Unmarshal(data, value)
}

// canUseGob 检查类型是否可以使用gob序列化
func (ng *NGCache) canUseGob(value interface{}) bool {
	t := reflect.TypeOf(value)
	if t == nil {
		return false
	}

	// 检查是否包含不支持gob的类型
	switch t.Kind() {
	case reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return false
	case reflect.Ptr:
		return ng.canUseGob(reflect.ValueOf(value).Elem().Interface())
	case reflect.Struct:
		// 检查结构体字段
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				return false
			}
			if !ng.canUseGobType(field.Type) {
				return false
			}
		}
		return true
	default:
		return ng.canUseGobType(t)
	}
}

// canUseGobType 检查类型是否支持gob
func (ng *NGCache) canUseGobType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128,
		reflect.String:
		return true
	case reflect.Array, reflect.Slice:
		return ng.canUseGobType(t.Elem())
	case reflect.Map:
		return ng.canUseGobType(t.Key()) && ng.canUseGobType(t.Elem())
	case reflect.Ptr:
		return ng.canUseGobType(t.Elem())
	case reflect.Struct:
		return true // 结构体在上层函数中检查
	default:
		return false
	}
}
