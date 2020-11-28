package config

import (
	"github.com/spf13/viper"
)

// InstructorWorkerConfig specified a worker destination service
type InstructorWorkerConfig struct {
	// Human readable name for the worker target
	Alias string

	// IP or DNS resolvable hostname of the worker
	Adress string

	// TCP port of the worker
	Port int

	Certificate string

	Secret string
}

type InstructorEndpoint struct {
	Url    string
	Weight int
}

// InstructorConfig represents the configuration structure for
// instructor mode
type InstructorConfig struct {

	// Workers is a list of worker targets a Loago instance in instructor mode
	// should reach out to for requesting load tests
	Workers []*InstructorWorkerConfig

	// Endpoints called by workers.
	Endpoints []*InstructorEndpoint

	// Amount of users to simulate per worker
	Amount int

	// Minimum time to wait in milliseconds for the next request per worker
	MinWait int

	// Maximum time to wait in milliseconds for the next request per worker
	MaxWait int
}

func NewInstructorConfig(v *viper.Viper) (*InstructorConfig, error) {
	var cfg InstructorConfig
	err := v.Unmarshal(&cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
