package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"shorten-service/internal/config"
	"shorten-service/internal/models"
	"shorten-service/internal/repositories"
	"time"
)

const (
	codeLength      = 7
	maxCodeAttempts = 100
)

var (
	ErrorInvalidTargetURL = errors.New("target URL must be an absolute http or https URL")
	ErrorCodeExhausted    = errors.New("could not allocate a unique code")
	ErrorLinkNotFound     = errors.New("link not found")
)

type LinkRepository interface {
	Create(ctx context.Context, code string, targetURL string, expiresAt *time.Time) (*models.Link, error)
	GetByCode(ctx context.Context, code string) (*models.Link, error)
}

type LinkService struct {
	repository LinkRepository
	cfg        *config.Config
}

func NewLinkService(repository LinkRepository, cfg *config.Config) *LinkService {
	return &LinkService{repository: repository, cfg: cfg}
}

// CreateLink creates a link.
func (s *LinkService) CreateLink(ctx context.Context, req models.CreateLinkRequest) (*models.LinkResponse, error) {

	// valide fields
	err := validateTargetURL(req.URL)
	if err != nil {
		return nil, err
	}

	// the generated code could already be used
	// so generate in a loop until it succeeds
	for attempt := 0; attempt < maxCodeAttempts; attempt++ {
		code, err := generateCode(codeLength)
		if err != nil {
			return nil, fmt.Errorf("failed generating code: %w", err)
		}

		link, err := s.repository.Create(ctx, code, req.URL, req.ExpiresAt)
		// success then return link
		if err == nil {
			return link.ToLinkResponse(s.cfg.ShortBaseURL), nil
			// code already used then contionue
		} else if errors.Is(err, repositories.ErrorCodeUsed) {
			continue
		}

		return nil, err
	}

	return nil, ErrorCodeExhausted
}

// GetLink returns a link by its code.
func (s *LinkService) GetLink(ctx context.Context, code string) (*models.LinkResponse, error) {
	link, err := s.repository.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repositories.ErrorLinkNotFound) {
			return nil, ErrorLinkNotFound
		}
		return nil, err
	}

	return link.ToLinkResponse(s.cfg.ShortBaseURL), nil
}

// validateTargetURL checks if url is valid
func validateTargetURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ErrorInvalidTargetURL
	}
	if parsed.Host == "" {
		return ErrorInvalidTargetURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrorInvalidTargetURL
	}
	return nil
}

// generateCode returns a random base62 string of the given length.
func generateCode(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	codeAlphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// 256 is not a multiple of 62, so the first 8 letters of the alphabet are
	// very slightly more likely. Irrelevant at this scale — it does not
	// meaningfully affect collision probability.
	for i, b := range buf {
		buf[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}

	return string(buf), nil
}
