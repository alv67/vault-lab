package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/amelamela/vault-lab/internal/auth"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/price"
	"github.com/amelamela/vault-lab/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailExists        = errors.New("email already registered")
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
)

type Service struct {
	repos   *repository.Repository
	jwtAuth *auth.JWTAuth
}

func New(repos *repository.Repository, jwtAuth *auth.JWTAuth) *Service {
	return &Service{repos: repos, jwtAuth: jwtAuth}
}

func (s *Service) Register(ctx context.Context, email, name, password string) (*model.User, error) {
	existing, _ := s.repos.User.FindByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailExists
	}
	return s.repos.User.Create(ctx, email, name, password)
}

func (s *Service) Login(ctx context.Context, email, password string) (*model.User, string, string, error) {
	user, err := s.repos.User.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	accessToken, err := s.jwtAuth.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := s.jwtAuth.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, tokenString string) (string, string, error) {
	claims, err := s.jwtAuth.ValidateToken(tokenString)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}
	if claims.TokenType != "refresh" {
		return "", "", ErrInvalidCredentials
	}

	user, err := s.repos.User.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", "", ErrNotFound
	}

	accessToken, err := s.jwtAuth.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.jwtAuth.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, claims *auth.Claims) (*model.User, error) {
	return s.repos.User.FindByID(ctx, claims.UserID)
}

func (s *Service) CreateAsset(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	return s.repos.Asset.Create(ctx, asset)
}

func (s *Service) GetAsset(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	return s.repos.Asset.FindByID(ctx, id)
}

func (s *Service) SearchAssets(ctx context.Context, query string) ([]*model.Asset, error) {
	return s.repos.Asset.Search(ctx, query)
}

func (s *Service) LookupAsset(ctx context.Context, query string) ([]price.AssetLookup, error) {
	return price.LookupAsset(ctx, query)
}

func (s *Service) ListAssets(ctx context.Context) ([]*model.Asset, error) {
	return s.repos.Asset.List(ctx)
}

func (s *Service) CreatePortfolio(ctx context.Context, userID uuid.UUID, name, description, currency string) (*model.Portfolio, error) {
	p := &model.Portfolio{
		UserID:      userID,
		Name:        name,
		Description: description,
		Currency:    currency,
	}
	return s.repos.Portfolio.Create(ctx, p)
}

func (s *Service) ListPortfolios(ctx context.Context, userID uuid.UUID) ([]*model.Portfolio, error) {
	return s.repos.Portfolio.FindByUser(ctx, userID)
}

func (s *Service) GetPortfolio(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Portfolio, error) {
	p, err := s.repos.Portfolio.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.canAccessPortfolio(ctx, p, userID) {
		return nil, ErrForbidden
	}
	return p, nil
}

func (s *Service) UpdatePortfolio(ctx context.Context, p *model.Portfolio) error {
	return s.repos.Portfolio.Update(ctx, p)
}

func (s *Service) DeletePortfolio(ctx context.Context, id uuid.UUID) error {
	return s.repos.Portfolio.Delete(ctx, id)
}

func (s *Service) AddTransaction(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	return s.repos.Transaction.Create(ctx, tx)
}

func (s *Service) ListTransactions(ctx context.Context, portfolioID uuid.UUID) ([]model.TransactionWithAsset, error) {
	return s.repos.Transaction.FindByPortfolio(ctx, portfolioID)
}

func (s *Service) GetPortfolioSummary(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioSummary, error) {
	return s.repos.Portfolio.GetSummary(ctx, portfolioID)
}

func (s *Service) GetPortfolioAllocation(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetAllocation, error) {
	return s.repos.Portfolio.GetAllocation(ctx, portfolioID)
}

func (s *Service) GetPortfolioPerformance(ctx context.Context, portfolioID uuid.UUID) ([]*model.PortfolioPerformance, error) {
	return s.repos.Portfolio.GetPerformance(ctx, portfolioID)
}

func (s *Service) GetPortfolioROI(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetROI, error) {
	return s.repos.Portfolio.GetROI(ctx, portfolioID)
}

func (s *Service) GetPrices(ctx context.Context, assetID uuid.UUID) ([]*model.Price, error) {
	return s.repos.Price.FindByAsset(ctx, assetID)
}

func (s *Service) canAccessPortfolio(ctx context.Context, portfolio *model.Portfolio, userID uuid.UUID) bool {
	if portfolio.UserID == userID {
		return true
	}
	return false // TODO: check portfolio_shares
}
