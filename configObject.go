package conf

import (
	"fmt"
	"strconv"
	"time"
)

//ConfigObject 配置对象
type ConfigObject struct {
	data map[string]Result
	//标记是否存在
	isExistence bool
	//配置来源
	source Source
	//fileName 配置文件标志
	fileName string
}

//Result 配置数据解析结果
type Result struct {
	dataType confType
	value    interface{}
	//标记是否存在
	isExistence bool
}

//Get 获取一个配置结果
func (c *ConfigObject) Get(key string) *Result {
	if key == "" {
		return new(Result)
	}
	r, ok := c.data[key]
	if ok {
		return &r
	}
	return new(Result)

}

// All 获取全部配置
func (c *ConfigObject) All() map[string]Result {
	return c.data
}

// Exists 判断是否存在此配置对象
func (c *ConfigObject) Exists() bool {
	return c.isExistence
}

// Default 当配置不存在时以此方法设置的默认值返回
func (r *Result) Default(defaultValue interface{}) *Result {
	if !r.isExistence {
		rTmp := genResult(defaultValue)
		return &rTmp
	}
	return r
}

// Exists 判断是否存在此配置值
func (r *Result) Exists() bool {
	return r.isExistence
}

// Value 返回原解析配置值而不进行任何转化
func (r *Result) Value() interface{} {
	return r.value
}

// String 以字符串返回配置值
func (r *Result) String() string {
	return fmt.Sprintf("%v", r.value)
}

// Slice 以切片返回配置值在toml中 类似 k=[1,2]这样配置会以切片返回(xdiamond 配置中心不支持这种配置)，除此之外返回包含配置值的切片
func (r *Result) Slice() []interface{} {
	if r.dataType == Array {
		v, ok := r.value.([]interface{})
		if !ok {
			return make([]interface{}, 0, 0)
		}
		return v
	}
	v := make([]interface{}, 0)
	return append(v, r.value)
}

// SliceMap 断言返回类似以下配置
/**
[[crm.slave]]
	addr = "localhost:6379"
    password = ""
    db = 0
[[crm.slave]]
	addr = "localhost:6379"
    password = ""
    db = 0
**/
func (r *Result) SliceMap() []map[string]interface{} {
	v, ok := r.value.([]map[string]interface{})
	if !ok {
		return make([]map[string]interface{}, 0, 0)
	}
	return v
}

// Time 以时间格式返回配置值，时间格式依照toml以RFC3339因特网标准时间为准
func (r *Result) Time() time.Time {
	if r.dataType == Time {
		v, ok := r.value.(time.Time)
		if ok {
			return v
		}
	}
	if r.dataType == String {
		v, ok := r.value.(string)
		if ok {
			t, err := time.Parse(time.RFC3339, v)
			if err == nil {
				return t
			}
		}
	}
	return time.Unix(0, 0)
}

// ToDateTime 尝试以Y-m-d h:i:s的格式返回时间配置值字符串
func (r *Result) ToDateTime() string {
	timeLayout := "2006-01-02 15:04:05"
	return r.Time().Format(timeLayout)
}

// Bool 以布尔型返回配置值
func (r *Result) Bool() bool {
	switch r.dataType {
	case String:
		v, ok := r.value.(string)
		if !ok {
			return false
		}
		n, err := strconv.ParseBool(v)
		if err != nil {
			return false
		}
		return n
	case Int, Uint, Float:
		v, ok := r.value.(float64)
		if !ok {
			return false
		}
		return v != 0
	case Bool:
		v, ok := r.value.(bool)
		if !ok {
			return false
		}
		return v
	case Array:
		v, ok := r.value.([]interface{})
		if !ok {
			return false
		}
		return len(v) > 0
	case Time:
		v, ok := r.value.(time.Time)
		if !ok {
			return false
		}
		return v.Unix() > 0
	case Undefined:
		return !(r.value == nil)
	default:
		return false
	}
}

// Float 以浮点类型返回配置值
func (r *Result) Float() float64 {
	switch r.dataType {
	case String:
		v, ok := r.value.(string)
		if !ok {
			return float64(0)
		}
		n, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return float64(0)
		}
		return n
	case Int, Uint, Float:
		v, ok := r.value.(float64)
		if !ok {
			return float64(0)
		}
		return float64(v)
	case Bool:
		v, ok := r.value.(bool)
		if !ok {
			return float64(0)
		}
		if v {
			return float64(1)
		}
	default:
		return float64(0)
	}
	return float64(0)
}

// Int 以int64类型返回配置值
func (r *Result) Int() int64 {
	switch r.dataType {
	case String:
		v, ok := r.value.(string)
		if !ok {
			return int64(0)
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return int64(0)
		}
		return n
	case Int, Uint, Float:
		v, ok := r.value.(int64)
		if !ok {
			return int64(0)
		}
		return int64(v)
	case Bool:
		v, ok := r.value.(bool)
		if !ok {
			return int64(0)
		}
		if v {
			return int64(1)
		}
	default:
		return int64(0)
	}
	return int64(0)
}

// Uint 以uint64返回配置值，注意：负数转无符号数得到的值可能不是你预期的
func (r *Result) Uint() uint64 {
	return uint64(r.Int())
}
