package config

import (
	"errors"
	"fmt"
)

// ValidateInstructorConfig validates the instructor sub config
func ValidateInstructorConfig(cfg *InstructorConfig) error {
	if cfg == nil {
		return errors.New("missing instructor config")
	}

	if len(cfg.Workers) == 0 {
		return errors.New("no worker targets configured")
	}

	for _, v := range cfg.Workers {
		if v.Alias == "" {
			return fmt.Errorf("invalid alias '%s'", v.Alias)
		}

		// TODO: Add ip address regex match
		if v.Adress == "" {
			return fmt.Errorf("invalid adress '%s'", v.Adress)
		}

		// TODO: Add Port Range check
		if v.Port == 0 {
			return fmt.Errorf("invalid port '%d'", v.Port)
		}
	}

	return nil
}
