package main

import (
	"fmt"
	"runtime"
	"time"

	"ngcat"
)

// BenchmarkResult 基准测试结果
type BenchmarkResult struct {
	Operation string        // 操作名称
	Count     int           // 操作次数
	Duration  time.Duration // 总耗时
	AvgTime   time.Duration // 平均耗时
	OpsPerSec int64         // 每秒操作数
}

// runBenchmark 运行基准测试
func runBenchmark(name string, count int, fn func()) BenchmarkResult {
	// 强制GC，确保测试环境一致
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	start := time.Now()
	fn()
	duration := time.Since(start)

	avgTime := duration / time.Duration(count)
	opsPerSec := int64(float64(count) / duration.Seconds())

	return BenchmarkResult{
		Operation: name,
		Count:     count,
		Duration:  duration,
		AvgTime:   avgTime,
		OpsPerSec: opsPerSec,
	}
}

// printResult 打印测试结果
func printResult(result BenchmarkResult) {
	fmt.Printf("%-25s %10d ops %12v %10v/op %12d ops/sec\n",
		result.Operation,
		result.Count,
		result.Duration,
		result.AvgTime,
		result.OpsPerSec)
}

func main() {
	fmt.Println("=== NGCache 性能基准测试 ===")
	fmt.Printf("Go版本: %s\n", runtime.Version())
	fmt.Printf("操作系统: %s\n", runtime.GOOS)
	fmt.Printf("架构: %s\n", runtime.GOARCH)
	fmt.Printf("CPU核心数: %d\n\n", runtime.NumCPU())

	// 创建缓存实例
	cache := ngcat.NewNGCache(100*1024*1024, nil) // 100MB缓存
	defer cache.Close()

	fmt.Printf("%-25s %10s %12s %12s %12s\n", "操作", "次数", "总耗时", "平均耗时", "ops/sec")
	fmt.Println("--------------------------------------------------------------------------------")

	// 1. 字符串操作基准测试
	const stringTestCount = 100000
	testData := make([]string, stringTestCount)
	for i := 0; i < stringTestCount; i++ {
		testData[i] = fmt.Sprintf("test_value_%d", i)
	}

	// 字符串设置测试
	result := runBenchmark("SetString", stringTestCount, func() {
		for i := 0; i < stringTestCount; i++ {
			key := fmt.Sprintf("str_key_%d", i)
			cache.SetString(key, testData[i])
		}
	})
	printResult(result)

	// 字符串获取测试
	result = runBenchmark("GetString", stringTestCount, func() {
		for i := 0; i < stringTestCount; i++ {
			key := fmt.Sprintf("str_key_%d", i)
			cache.GetString(key)
		}
	})
	printResult(result)

	// 2. 整数操作基准测试
	const intTestCount = 100000

	// Int64设置测试
	result = runBenchmark("SetInt64", intTestCount, func() {
		for i := 0; i < intTestCount; i++ {
			key := fmt.Sprintf("int_key_%d", i)
			cache.SetInt64(key, int64(i))
		}
	})
	printResult(result)

	// Int64获取测试
	result = runBenchmark("GetInt64", intTestCount, func() {
		for i := 0; i < intTestCount; i++ {
			key := fmt.Sprintf("int_key_%d", i)
			cache.GetInt64(key)
		}
	})
	printResult(result)

	// 3. 浮点数操作基准测试
	const floatTestCount = 50000

	// Float64设置测试
	result = runBenchmark("SetFloat64", floatTestCount, func() {
		for i := 0; i < floatTestCount; i++ {
			key := fmt.Sprintf("float_key_%d", i)
			cache.SetFloat64(key, float64(i)*3.14159)
		}
	})
	printResult(result)

	// Float64获取测试
	result = runBenchmark("GetFloat64", floatTestCount, func() {
		for i := 0; i < floatTestCount; i++ {
			key := fmt.Sprintf("float_key_%d", i)
			cache.GetFloat64(key)
		}
	})
	printResult(result)

	// 4. 结构体序列化基准测试
	type TestStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Age  int    `json:"age"`
		Data []byte `json:"data"`
	}

	const structTestCount = 10000
	testStructs := make([]TestStruct, structTestCount)
	for i := 0; i < structTestCount; i++ {
		testStructs[i] = TestStruct{
			ID:   i,
			Name: fmt.Sprintf("用户_%d", i),
			Age:  20 + i%50,
			Data: make([]byte, 100), // 100字节数据
		}
	}

	// JSON序列化设置测试
	result = runBenchmark("SetJSON", structTestCount, func() {
		for i := 0; i < structTestCount; i++ {
			key := fmt.Sprintf("json_key_%d", i)
			cache.SetJSON(key, testStructs[i])
		}
	})
	printResult(result)

	// JSON序列化获取测试
	result = runBenchmark("GetJSON", structTestCount, func() {
		for i := 0; i < structTestCount; i++ {
			key := fmt.Sprintf("json_key_%d", i)
			var s TestStruct
			cache.GetJSON(key, &s)
		}
	})
	printResult(result)

	// Gob序列化设置测试
	result = runBenchmark("SetAny", structTestCount, func() {
		for i := 0; i < structTestCount; i++ {
			key := fmt.Sprintf("gob_key_%d", i)
			cache.SetAny(key, testStructs[i])
		}
	})
	printResult(result)

	// Gob序列化获取测试
	result = runBenchmark("GetAny", structTestCount, func() {
		for i := 0; i < structTestCount; i++ {
			key := fmt.Sprintf("gob_key_%d", i)
			var s TestStruct
			cache.GetAny(key, &s)
		}
	})
	printResult(result)

	// 5. 永久缓存基准测试
	const permanentTestCount = 50000

	// 永久缓存设置测试
	result = runBenchmark("SetPermanent", permanentTestCount, func() {
		for i := 0; i < permanentTestCount; i++ {
			key := fmt.Sprintf("perm_key_%d", i)
			value := fmt.Sprintf("permanent_value_%d", i)
			cache.SetPermanent([]byte(key), []byte(value))
		}
	})
	printResult(result)

	// 永久缓存获取测试
	result = runBenchmark("GetPermanent", permanentTestCount, func() {
		for i := 0; i < permanentTestCount; i++ {
			key := fmt.Sprintf("perm_key_%d", i)
			cache.GetPermanent([]byte(key))
		}
	})
	printResult(result)

	fmt.Println("\n=== 内存使用情况 ===")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("分配的内存: %.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Printf("系统内存: %.2f MB\n", float64(m.Sys)/1024/1024)
	fmt.Printf("GC次数: %d\n", m.NumGC)
	fmt.Printf("GC暂停时间: %v\n", time.Duration(m.PauseTotalNs))

	fmt.Println("\n=== 基准测试完成 ===")
}
