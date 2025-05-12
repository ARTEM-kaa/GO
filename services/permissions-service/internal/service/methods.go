package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/ARTEM-kaa/GO/internal/proxyproto"
	"github.com/ARTEM-kaa/GO/internal/userdb"
	"github.com/Nerzal/gocloak/v13"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	CentrifugoInternalServerError = 100
	CentrifugoUnauthorized        = 101
	CentrifugoPermissionDenied    = 103
	CentrifugoBadRequest          = 107
)

func (s *Service) fetchKeycloakUser(ctx context.Context, userId uuid.UUID) (userdb.User, error) {
	kcUser, err := s.GetKeycloakUser(ctx, userId.String())
	if err != nil {
		return userdb.User{}, err
	}

	user := userdb.User{
		ID:         pgtype.UUID{Valid: true, Bytes: userId},
		Username:   *kcUser.Username,
		GivenName:  *kcUser.FirstName,
		FamilyName: *kcUser.LastName,
		Enabled:    *kcUser.Enabled,
	}

	err = s.storage.CreateUser(ctx, UserToCreateUserParams(user))
	if err != nil {
		return userdb.User{}, err
	}

	return user, nil
}

func (s *Service) Subscribe(ctx context.Context, request *proxyproto.SubscribeRequest) (*proxyproto.SubscribeResponse, error) {
	userId, err := uuid.Parse(request.User)
	if err != nil {
		return RespondSubscribeError(CentrifugoBadRequest, "invalid user id")
	}

	user, err := s.storage.GetUserByID(ctx, pgtype.UUID{Bytes: userId, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		user, err = s.fetchKeycloakUser(ctx, userId)
		if err != nil {
			log.Println(err)
			if errors.Is(err, gocloak.APIError{Code: http.StatusNotFound}) {
				return RespondSubscribeError(CentrifugoUnauthorized, "unknown user")
			}
			return RespondSubscribeError(CentrifugoInternalServerError, "internal server error")
		}
	} else if err != nil {
		log.Println(err)
		return RespondSubscribeError(CentrifugoInternalServerError, "internal server error")
	}

	count, err := s.storage.UserCanSubscribe(ctx, userdb.UserCanSubscribeParams{
		ID:      user.ID,
		Channel: request.Channel,
	})

	if count == 0 {
		return RespondSubscribeError(CentrifugoPermissionDenied, "permission denied")
	}

	return &proxyproto.SubscribeResponse{}, nil
}

func (s *Service) Publish(ctx context.Context, request *proxyproto.PublishRequest) (*proxyproto.PublishResponse, error) {
	userId, err := uuid.Parse(request.User)
	if err != nil {
		return RespondPublishError(CentrifugoBadRequest, "invalid user id")
	}

	user, err := s.storage.GetUserByID(ctx, pgtype.UUID{Bytes: userId, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		user, err = s.fetchKeycloakUser(ctx, userId)
		if err != nil {
			log.Println(err)
			if errors.Is(err, gocloak.APIError{Code: http.StatusNotFound}) {
				return RespondPublishError(CentrifugoUnauthorized, "unknown user")
			}
			return RespondPublishError(CentrifugoInternalServerError, "internal server error")
		}
	} else if err != nil {
		log.Println(err)
		return RespondPublishError(CentrifugoInternalServerError, "internal server error")
	}

	count, err := s.storage.UserCanPublish(ctx, userdb.UserCanPublishParams{
		ID:      user.ID,
		Channel: request.Channel,
	})

	if count == 0 {
		return RespondPublishError(CentrifugoPermissionDenied, "permission denied")
	}

	return &proxyproto.PublishResponse{}, nil
}
