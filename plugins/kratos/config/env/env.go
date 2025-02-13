package env

import (
	"strconv"
	"strings"
	"syscall"

	"github.com/go-cinch/common/copierx"
)

func NewRevolver(options ...func(*Options)) func(map[string]interface{}) error {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	resolver := func(sub map[string]interface{}) error {
		return envResolver(*ops, sub)
	}
	return resolver
}

func envResolver(ops Options, sub map[string]interface{}) error {
	// tip: json has string/int/float64/map[string]interface{}/[]interface{}
	for k, v := range sub {
		key := strings.Join([]string{ops.prefix, k}, ops.separator)
		if ops.prefix == "" {
			key = k
		}
		key = strings.ToUpper(key)
		var found1 bool
		var v1 interface{}
		switch vt := v.(type) {
		case string:
			v1, found1 = syscall.Getenv(key)
		case bool:
			v1, found1 = getBoolEnv(key)
		case int:
			v1, found1 = getIntEnv(key)
		case float64:
			v1, found1 = getFloat64Env(key)
		case map[string]interface{}:
			newOps := ops
			newOps.prefix = key
			if err := envResolver(newOps, vt); err != nil {
				return err
			}
		case []interface{}:
			if len(vt) == 0 {
				continue
			}
			// arr item is map need check env by index
			switch v0 := vt[0].(type) {
			case map[string]interface{}:
				arrKeys := getKeys(v0)
				// check env from 0 to end
				// end: cannot find any val from env by current index
				index := 0
				for {
					var found2 bool
					for _, arrKey := range arrKeys {
						idxKey := strings.Join([]string{key, strconv.Itoa(index), strings.ToUpper(arrKey)}, ops.separator)
						switch v0[arrKey].(type) {
						case string:
							_, found2 = syscall.Getenv(idxKey)
						case bool:
							_, found2 = getBoolEnv(idxKey)
						case int:
							_, found2 = getIntEnv(idxKey)
						case float64:
							_, found2 = getFloat64Env(idxKey)
						}
						if found2 {
							break
						}
					}
					if !found2 {
						break
					}
					if index > 0 {
						// increase vt cap
						// new a copy
						m := make(map[string]interface{}, len(v0))
						copierx.Copy(&m, v0)
						vt = append(vt, m)
					}
					index++
				}
			}
			for i, item := range vt {
				idxKey := strings.Join([]string{key, strconv.Itoa(i)}, ops.separator)
				var found2 bool
				var v2 interface{}
				switch it := item.(type) {
				case string:
					v2, found2 = syscall.Getenv(idxKey)
				case bool:
					v2, found2 = getBoolEnv(idxKey)
				case int:
					v2, found2 = getIntEnv(idxKey)
				case float64:
					v2, found2 = getFloat64Env(idxKey)
				case map[string]interface{}:
					newOps := ops
					newOps.prefix = idxKey
					if err := envResolver(newOps, it); err != nil {
						return err
					}
					continue
				}
				if found2 {
					vt[i] = v2
					if ops.loaded != nil {
						ops.loaded(idxKey, v2)
					}
				}
			}
			sub[k] = vt
			continue
		}
		if found1 {
			sub[k] = v1
			if ops.loaded != nil {
				ops.loaded(key, v1)
			}
		}
	}
	return nil
}

func getBoolEnv(key string) (v bool, ok bool) {
	envV, found := syscall.Getenv(key)
	if found {
		envVV, err := strconv.ParseBool(envV)
		if err != nil {
			return
		}
		v = envVV
		ok = true
		return
	}
	return
}

func getIntEnv(key string) (v int, ok bool) {
	envV, found := syscall.Getenv(key)
	if found {
		envVV, err := strconv.Atoi(envV)
		if err != nil {
			return
		}
		v = envVV
		ok = true
		return
	}
	return
}

func getFloat64Env(key string) (v float64, ok bool) {
	envV, found := syscall.Getenv(key)
	if found {
		envVV, err := strconv.ParseFloat(envV, 64)
		if err != nil {
			return
		}
		v = envVV
		ok = true
		return
	}
	return
}

func getKeys(m map[string]interface{}) []string {
	j := 0
	keys := make([]string, len(m), len(m))
	for k := range m {
		keys[j] = k
		j++
	}
	return keys
}
