package outputs

import (
	"errors"
	"fmt"
)

type CreatorFunc func(map[string]interface{}) (Output, error)

var drivers map[string]CreatorFunc

func init() {
	drivers = make(map[string]CreatorFunc)
}

func RegisterDriver(name string, maker CreatorFunc) {
	fmt.Println("debug: registering driver", name)
	drivers[name] = maker
}

func LoadDriver(name string, conf map[string]interface{}) (Output, error) {
	f, ok := drivers[name]
	if !ok {
		fmt.Println("debug: failed to load driver", name)
		return nil, errors.New(`driver doesn't exist or wasn't loaded`)
	}

	fmt.Println("debug: loading driver", name)
	return f(conf)
}
