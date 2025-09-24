package ngcat

import (
	"testing"
	"time"
)

func TestPersistence(t *testing.T) {
	nc := NewNGCache(1024*1024, &PersistConfig{
		Enabled:  true,
		FilePath: "",
		FileName: "test.cat",
		Format:   FormatBinary,
		Interval: 5,
	})

	nc.SetString("string", "test", 0)
	nc.SetBool("bool", true, 0)
	nc.SetInt32("int", 1, 0)
	nc.SetFloat32("float", 0.99, 0)

	vstring, _ := nc.GetString("string")
	t.Log(vstring)

	vbool, _ := nc.GetBool("bool")
	t.Log(vbool)

	vint, _ := nc.GetInt32("int")
	t.Log(vint)

	vfloat, _ := nc.GetFloat32("float")
	t.Log(vfloat)

	time.Sleep(time.Duration(1) * time.Second)
	nc.Close()
}

func TestPersistenceLoad(t *testing.T) {
	nc := NewNGCache(1024*1024, &PersistConfig{
		Enabled:  true,
		FilePath: "",
		FileName: "test.cat",
		Format:   FormatBinary,
		Interval: 5,
	})

	// 自动加载持久化数据

	vstring, _ := nc.GetString("string")
	t.Log(vstring)

	vbool, _ := nc.GetBool("bool")
	t.Log(vbool)

	vint, _ := nc.GetInt32("int")
	t.Log(vint)

	vfloat, _ := nc.GetFloat32("float")
	t.Log(vfloat)

	nc.Close()
}
