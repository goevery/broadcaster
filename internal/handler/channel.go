package handler

import (
	"errors"
	"regexp"

	"github.com/goevery/broadcaster/internal/ierr"
)

type ChannelValidator struct {
	channelRegex *regexp.Regexp
}

func NewChannelValidator() *ChannelValidator {
	return &ChannelValidator{
		channelRegex: regexp.MustCompile(`^([\w-]+:?)*\w$`),
	}
}

func (v *ChannelValidator) Validate(channel string) error {
	valid := v.channelRegex.MatchString(channel)
	if !valid {
		return ierr.New(ierr.ErrorCodeInvalidArgument, errors.New("invalid channel"))
	}

	return nil
}
