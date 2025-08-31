package handler

import (
	"errors"
	"regexp"

	"github.com/juanpmarin/broadcaster/internal/protocol"
)

type ChannelIdValidator struct {
	channelIdRegex *regexp.Regexp
}

func NewChannelIdValidator() *ChannelIdValidator {
	return &ChannelIdValidator{
		channelIdRegex: regexp.MustCompile(`^([\w-]+:?)*\w$`),
	}
}

func (v *ChannelIdValidator) Validate(channelId string) error {
	valid := v.channelIdRegex.MatchString(channelId)
	if !valid {
		return protocol.NewError(protocol.ErrorCodeInvalidArgument, errors.New("invalid channelId"))
	}

	return nil
}
