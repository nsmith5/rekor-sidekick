package outputs

import "errors"

type CreatorFunc func(map[string]interface{}) (Output, error)

var drivers map[string]CreatorFunc

func init() {
	drivers = make(map[string]CreatorFunc)
}

func RegisterDriver(name string, maker CreatorFunc) {
	drivers[name] = maker
}

func LoadDriver(name string, conf map[string]interface{}) (Output, error) {
	f, ok := drivers[name]
	if !ok {
		return nil, errors.New(`driver doesn't exist or wasn't loaded`)
	}

	return f(conf)
}
