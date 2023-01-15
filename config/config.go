package config

import (
	"math/big"
	"os"
	"strconv"
)

// GetString returns a setting in string.
func GetString(key string) string {
	val, exists := os.LookupEnv(key)
	if !exists {
		panic(key)
	}

	return val
}

// GetBool returns a setting in bool.
func GetBool(key string) bool {
	var val bool
	var err error
	if val, err = strconv.ParseBool(GetString(key)); err != nil {
		panic(err)
	}

	return val
}

// GetInt returns a setting in integer.
func GetInt(key string) int {
	val := int(GetInt64(key))
	return val
}

// GetBigInt returns a setting in bigInt.
func GetBigInt(key string) *big.Int {
	str := GetString(key)
	n := new(big.Int)
	n, ok := n.SetString(str, 10)
	if !ok {
		panic("config invalid" + key)
	}
	return n
}

// GetUint returns a setting in unsigned integer.
func GetUint(key string) uint {
	val := uint(GetUint64(key))
	return val
}

// GetInt64 returns a setting in 64-bit signed integer.
func GetInt64(key string) int64 {
	// Parse int64 value from environment variable.
	var val int64
	var err error
	if val, err = strconv.ParseInt(GetString(key), 0, 64); err != nil {
		panic(err)
	}

	return val
}

// GetUint64 returns a setting in 64-bit unsigned integer.
func GetUint64(key string) uint64 {
	// Parse uint64 value from environment variable.
	var val uint64
	var err error
	if val, err = strconv.ParseUint(GetString(key), 0, 64); err != nil {
		panic(err)
	}

	return val
}
