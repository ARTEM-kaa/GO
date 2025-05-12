package service

import (
	"context"
	"time"

	"github.com/ARTEM-kaa/GO/internal/proxyproto"
	"github.com/ARTEM-kaa/GO/internal/userdb"
	"github.com/ARTEM-kaa/GO/services/permissions-service/internal/config"
	"github.com/Nerzal/gocloak/v13"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	proxyproto.UnimplementedCentrifugoProxyServer
	conn        *pgxpool.Pool
	storage     *userdb.Queries
	config      *config.Config
	kcClient    *gocloak.GoCloak
	token       *gocloak.JWT
	tokenExpiry time.Time
}

func New(config *config.Config) (*Service, error) {
	connCfg, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), connCfg)
	if err != nil {
		return nil, err
	}

	kcClient := gocloak.NewClient(config.KeyCloakURL)

	return &Service{
		conn:     conn,
		storage:  userdb.New(conn),
		config:   config,
		kcClient: kcClient,
	}, nil
}

func (s *Service) getKeycloakToken(ctx context.Context) (string, error) {
	if s.token == nil || time.Now().After(s.tokenExpiry) {
		token, err := s.kcClient.LoginClient(
			ctx,
			s.config.KeyCloakClient,
			s.config.KeyCloakSecret,
			s.config.KeyCloakRealm,
		)
		if err != nil {
			return "", err
		}
		s.token = token
		s.tokenExpiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	}
	return s.token.AccessToken, nil
}

func (s *Service) GetKeycloakUser(ctx context.Context, userID string) (*gocloak.User, error) {
	token, err := s.getKeycloakToken(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.kcClient.GetUserByID(ctx, token, s.config.KeyCloakRealm, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
