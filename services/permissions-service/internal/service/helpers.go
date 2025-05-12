package service

import (
	"github.com/ARTEM-kaa/GO/internal/proxyproto"
	"github.com/ARTEM-kaa/GO/internal/userdb"
)

func ConnectRespondError(code uint32, msg string) (*proxyproto.ConnectResponse, error) {
	return &proxyproto.ConnectResponse{
		Error: &proxyproto.Error{
			Code:    code,
			Message: msg,
		},
	}, nil
}

func RespondSubscribeError(code uint32, msg string) (*proxyproto.SubscribeResponse, error) {
	return &proxyproto.SubscribeResponse{
		Error: &proxyproto.Error{
			Code:    code,
			Message: msg,
		},
	}, nil
}

func RespondPublishError(code uint32, msg string) (*proxyproto.PublishResponse, error) {
	return &proxyproto.PublishResponse{
		Error: &proxyproto.Error{
			Code:    code,
			Message: msg,
		},
	}, nil
}

func UserToCreateUserParams(user userdb.User) userdb.CreateUserParams {
	return userdb.CreateUserParams{
		ID:         user.ID,
		Username:   user.Username,
		GivenName:  user.GivenName,
		FamilyName: user.FamilyName,
		Enabled:    user.Enabled,
	}
}
