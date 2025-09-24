package main

import (
	"fmt"
	"time"

	"ngcat"
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

	fmt.Println("=== NGCache 扩展缓存库示例 ===")

	// 1. 基础类型操作示例
	fmt.Println("\n1. 基础类型操作:")

	// 整数类型
	cache.SetInt32("age", 25, 0)
	age, _ := cache.GetInt32("age")
	fmt.Printf("年龄: %d\n", age)

	cache.SetInt64("timestamp", time.Now().Unix(), 0)
	timestamp, _ := cache.GetInt64("timestamp")
	fmt.Printf("时间戳: %d\n", timestamp)

	// 浮点数类型
	cache.SetFloat64("price", 99.99, 0)
	price, _ := cache.GetFloat64("price")
	fmt.Printf("价格: %.2f\n", price)

	// 布尔类型
	cache.SetBool("is_active", true, 0)
	isActive, _ := cache.GetBool("is_active")
	fmt.Printf("是否激活: %t\n", isActive)

	// 字符串类型
	cache.SetString("username", "张三", 0)
	username, _ := cache.GetString("username")
	fmt.Printf("用户名: %s\n", username)

	// 2. 任意类型序列化示例
	fmt.Println("\n2. 任意类型序列化:")

	// 结构体示例
	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	user := User{ID: 1, Name: "李四", Age: 30}
	cache.SetAny("user_info", user, 0)

	var retrievedUser User
	err := cache.GetAny("user_info", &retrievedUser)
	if err == nil {
		fmt.Printf("用户信息: %+v\n", retrievedUser)
	}

	// JSON序列化示例
	cache.SetJSON("user_json", user, 0)
	var jsonUser User
	err = cache.GetJSON("user_json", &jsonUser)
	if err == nil {
		fmt.Printf("JSON用户信息: %+v\n", jsonUser)
	}

	// 3. 永久缓存示例
	fmt.Println("\n3. 永久缓存:")

	// 设置永久缓存
	cache.SetPermanent([]byte("permanent_key"), []byte("这是永久缓存的数据"))
	permanentData, err := cache.GetPermanent([]byte("permanent_key"))
	if err == nil {
		fmt.Printf("永久缓存数据: %s\n", string(permanentData))
	}

	// 4. 持久化功能演示
	fmt.Println("\n4. 持久化功能:")
	fmt.Println("数据将自动保存到文件，程序重启后可恢复")

	// 等待一段时间让持久化协程工作
	time.Sleep(2 * time.Second)

	// 5. 性能测试示例
	fmt.Println("\n5. 性能测试:")
	start := time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("test_key_%d", i)
		cache.SetString(key, fmt.Sprintf("test_value_%d", i), 0)
	}
	setDuration := time.Since(start)
	fmt.Printf("设置10000个字符串耗时: %v\n", setDuration)

	start = time.Now()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("test_key_%d", i)
		cache.GetString(key)
	}
	getDuration := time.Since(start)
	fmt.Printf("获取10000个字符串耗时: %v\n", getDuration)

	fmt.Println("\n=== 示例完成 ===")
}
