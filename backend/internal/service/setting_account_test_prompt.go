package service

import (
	"context"
	"errors"
	"strings"
)

func normalizeAccountTestPrompt(prompt string) string {
	if prompt = strings.TrimSpace(prompt); prompt != "" {
		return prompt
	}
	return defaultAccountTextTestPrompt
}

func (s *SettingService) GetAccountTestPrompt(ctx context.Context) (string, error) {
	if s == nil || s.settingRepo == nil {
		return defaultAccountTextTestPrompt, nil
	}
	prompt, err := s.settingRepo.GetValue(ctx, SettingKeyAccountTestPrompt)
	if errors.Is(err, ErrSettingNotFound) {
		return defaultAccountTextTestPrompt, nil
	}
	if err != nil {
		return "", err
	}
	return normalizeAccountTestPrompt(prompt), nil
}
