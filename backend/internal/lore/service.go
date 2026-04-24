package lore

import "context"

type Cache interface {
	Get(ctx context.Context, championID string) (string, bool, error)
	Put(ctx context.Context, championID, lore string) error
}

type Service struct {
	geminiURL string
	apiKey    string
	cache     Cache
}

func New(geminiURL, apiKey string, cache Cache) *Service {
	return &Service{geminiURL: geminiURL, apiKey: apiKey, cache: cache}
}

// Generate returns lore text for a champion. On any error, it returns ("", nil)
// so callers can treat lore as "not available" without breaking the response.
func (s *Service) Generate(ctx context.Context, championID, championName string) (string, error) {
	if s.apiKey == "" {
		return "", nil
	}
	if cached, ok, _ := s.cache.Get(ctx, championID); ok {
		return cached, nil
	}
	text, err := s.callGemini(ctx, championName)
	if err != nil || text == "" {
		return "", nil
	}
	_ = s.cache.Put(ctx, championID, text)
	return text, nil
}
