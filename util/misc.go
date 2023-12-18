package util

import (
	"math/rand"
)

func RandFloatingInt64(v int64) int64 {
	randNum := rand.Float64()*0.2 + 0.9
	return int64(float64(v) * randNum)
}

func GetTTL(ttlInstance interface{}) int64 {
	if ttlInstance == nil {
		return 0
	}
	switch ttlInstance.(type) {
	case int:
		return int64(ttlInstance.(int))
	case int8:
		return int64(ttlInstance.(int8))
	case int16:
		return int64(ttlInstance.(int16))
	case int32:
		return int64(ttlInstance.(int32))
	case int64:
		return ttlInstance.(int64)
	default:
		return 0
	}
}
