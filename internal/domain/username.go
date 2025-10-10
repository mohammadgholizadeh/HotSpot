package domain

import (
	"fmt"

	"go.uber.org/zap"
)

type Username string

func (u Username) Validate() error {
	if u == "" {
		zap.L().Warn("Username is empty",
			zap.String("entity", "username"),
		)

		return fmt.Errorf("username cannot be empty")
	}

	return nil
}
