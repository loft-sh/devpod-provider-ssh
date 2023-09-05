package options

import (
	"fmt"
	"os"
)

var (
	DOCKER_PATH = "DOCKER_PATH"
	AGENT_PATH  = "AGENT_PATH"
	HOST        = "HOST"
	PORT        = "PORT"
	EXTRA_FLAGS = "EXTRA_FLAGS"
)

type Options struct {
	DockerPath string
	AgentPath  string
	User       string
	Host       string
	Port       string
	ExtraFlags string
}

func FromEnv() (*Options, error) {
	retOptions := &Options{}

	var err error

	retOptions.DockerPath, err = fromEnvOrError(DOCKER_PATH)
	if err != nil {
		return nil, err
	}

	retOptions.AgentPath, err = fromEnvOrError(AGENT_PATH)
	if err != nil {
		return nil, err
	}

	retOptions.ExtraFlags = os.Getenv(EXTRA_FLAGS)

	retOptions.Host, err = fromEnvOrError(HOST)
	if err != nil {
		return nil, err
	}

	retOptions.Port, err = fromEnvOrError(PORT)
	if err != nil {
		return nil, err
	}

	return retOptions, nil
}

func fromEnvOrError(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf(
			"couldn't find option %s in environment, please make sure %s is defined",
			name,
			name,
		)
	}

	return val, nil
}
