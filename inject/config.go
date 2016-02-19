package glimmer

import (
	"github.com/miguel-branco/goconfig"
)

var (
	config goconfig.ConfigFile

	port string
)

// this init function should run before the runtime.init function
func init() {
	var err error

	config, err = goconfig.ReadConfigFile("glimmer_config.cfg")
	if err != nil {
		panic(err)
	}

	if port, err = config.GetInt64("default", "port"); err != nil {
		panic(err)
	}
}
