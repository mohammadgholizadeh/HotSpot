package domain

import (
	"fmt"
	"regexp"

	"go.uber.org/zap"
)

var mobileRegex = regexp.MustCompile(`^(0)?(\d{10})$`)

type MobileNumber string

func (m MobileNumber) Validate() error {
	if !mobileRegex.MatchString(string(m)) {
		zap.L().Warn("Mobile number format is invalid",
			zap.String("entity", "mobile_number"),
			zap.String("value", string(m)),
		)

		return fmt.Errorf("mobile number format is invalid")
	}

	return nil
}
