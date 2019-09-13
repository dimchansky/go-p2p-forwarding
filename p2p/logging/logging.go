package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/ipfs/go-log"
	"github.com/whyrusleeping/go-logging"
)

// Logging environment variables
const (
	envLogLevel   = "P2P_LOGLVL"
	envLogModules = "P2P_LOGMOD"
)

var (
	lvl             = logging.ERROR
	explicitModules = false
)

// Logger retrieves an event logger by name
func Logger(system string) log.EventLogger {
	l := log.Logger(system)
	if !explicitModules {
		// if no explicit modules set for logging, set global level
		logging.SetLevel(lvl, system)
	}
	return l
}

func init() {
	setupLogging()
}

func setupLogging() {
	if logEnv := os.Getenv(envLogLevel); logEnv != "" {
		var err error
		lvl, err = logging.LogLevel(logEnv)
		if err != nil {
			fmt.Println("error setting log levels", err)
		}

		if modulesEnv := os.Getenv(envLogModules); modulesEnv != "" {
			modules := strings.Split(modulesEnv, ",")
			explicitModules = true
			for _, module := range modules {
				logging.SetLevel(lvl, module)
			}
		}
	}
}
