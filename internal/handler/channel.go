package handler

import (
	"errors"
	"regexp"
)

type ChannelIdValidator struct {
	channelIdRegex *regexp.Regexp
}

func NewChannelIdValidator() *ChannelIdValidator {
	return &ChannelIdValidator{
		channelIdRegex: regexp.MustCompile("^([a-zA-Z0-9-]+:?)*[a-zA-Z]$"),
	}
}

func (v *ChannelIdValidator) Validate(channelId string) error {
	valid := v.channelIdRegex.MatchString(channelId)
	if !valid {
		return NewError(ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	return nil
}
