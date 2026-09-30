//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type accountTestPromptSettingRepo struct {
	value string
	err   error
}

func (r *accountTestPromptSettingRepo) GetValue(context.Context, string) (string, error) {
	return r.value, r.err
}

func (r *accountTestPromptSettingRepo) Get(context.Context, string) (*Setting, error) {
	return nil, r.err
}

func (r *accountTestPromptSettingRepo) Set(context.Context, string, string) error { return nil }

func (r *accountTestPromptSettingRepo) GetMultiple(context.Context, []string) (map[string]string, error) {
	return nil, r.err
}

func (r *accountTestPromptSettingRepo) SetMultiple(context.Context, map[string]string) error {
	return nil
}

func (r *accountTestPromptSettingRepo) GetAll(context.Context) (map[string]string, error) {
	return map[string]string{SettingKeyAccountTestPrompt: r.value}, r.err
}

func (r *accountTestPromptSettingRepo) Delete(context.Context, string) error { return nil }

func TestAccountTestPrompt_DefaultsAndPersists(t *testing.T) {
	t.Parallel()

	svc := NewSettingService(&settingUpdateRepoStub{}, &config.Config{})
	parsed := svc.parseSettings(map[string]string{
		SettingKeyAccountTestPrompt: "  ciallo  ",
	})
	require.Equal(t, "ciallo", parsed.AccountTestPrompt)
	require.Equal(t, defaultAccountTextTestPrompt, svc.parseSettings(map[string]string{}).AccountTestPrompt)
	require.Equal(t, defaultAccountTextTestPrompt, svc.parseSettings(map[string]string{
		SettingKeyAccountTestPrompt: "   ",
	}).AccountTestPrompt)

	repo := &settingUpdateRepoStub{}
	svc = NewSettingService(repo, &config.Config{})
	updates, err := svc.buildSystemSettingsUpdates(context.Background(), &SystemSettings{AccountTestPrompt: " ciallo "})
	require.NoError(t, err)
	require.Equal(t, "ciallo", updates[SettingKeyAccountTestPrompt])
}

func TestSettingService_GetAccountTestPromptFallsBackWhenMissingOrBlank(t *testing.T) {
	t.Parallel()

	for name, repo := range map[string]*accountTestPromptSettingRepo{
		"missing": {err: ErrSettingNotFound},
		"blank":   {value: "  "},
		"custom":  {value: " ciallo "},
	} {
		t.Run(name, func(t *testing.T) {
			svc := NewSettingService(repo, &config.Config{})
			got, err := svc.GetAccountTestPrompt(context.Background())
			require.NoError(t, err)
			want := defaultAccountTextTestPrompt
			if name == "custom" {
				want = "ciallo"
			}
			require.Equal(t, want, got)
		})
	}
}

func TestAccountTextTestPayloadsUseConfiguredPrompt(t *testing.T) {
	t.Parallel()

	claudePayload, err := createTestPayload("claude-3-5-sonnet", " ciallo ")
	require.NoError(t, err)
	claudeJSON, err := json.Marshal(claudePayload)
	require.NoError(t, err)
	require.Equal(t, "ciallo", gjson.GetBytes(claudeJSON, "messages.0.content.0.text").String())

	openAIPayload := createOpenAITestPayload("gpt-5.4", false, " ciallo ")
	openAIJSON, err := json.Marshal(openAIPayload)
	require.NoError(t, err)
	require.Equal(t, "ciallo", gjson.GetBytes(openAIJSON, "input.0.content.0.text").String())

	chatPayload := createOpenAIChatCompletionsTestPayload("gpt-5.4", " ciallo ")
	chatJSON, err := json.Marshal(chatPayload)
	require.NoError(t, err)
	require.Equal(t, "ciallo", gjson.GetBytes(chatJSON, "messages.0.content").String())

	geminiPayload := createGeminiTestPayload("gemini-2.5-flash", " ciallo ")
	require.Equal(t, "ciallo", gjson.GetBytes(geminiPayload, "contents.0.parts.0.text").String())
}

func TestAccountMediaAndBackgroundGrokPromptsKeepTheirDefaults(t *testing.T) {
	t.Parallel()

	imagePayload := createGeminiTestPayload("gemini-2.5-flash-image", defaultGeminiImageTestPrompt)
	require.Equal(t, defaultGeminiImageTestPrompt, gjson.GetBytes(imagePayload, "contents.0.parts.0.text").String())

	backgroundProbe, err := buildGrokQuotaProbeBody("grok-4.5")
	require.NoError(t, err)
	require.Equal(t, grokQuotaProbeInput, gjson.GetBytes(backgroundProbe, "input").String())

	configuredProbe, err := buildGrokQuotaProbeBody("grok-4.5", "ciallo")
	require.NoError(t, err)
	require.Equal(t, "ciallo", gjson.GetBytes(configuredProbe, "input").String())
}
