package bot

import (
	"strings"
	"testing"
)

func TestConfigFromEnv_Defaults(t *testing.T) {
	t.Setenv("CMDCTRL_DISCORD_BOT_TOKEN", "bot-token")
	t.Setenv("CMDCTRL_DISCORD_APP_ID", "app-id")
	t.Setenv("CMDCTRL_DISCORD_GUILD_IDS", "123,456")
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "admin-token")
	t.Setenv("CMDCTRL_SERVER_BASE_URL", "")
	t.Setenv("CMDCTRL_CLIENT_BASE_URL", "")

	c := ConfigFromEnv()

	if c.ServerBaseURL != defaultServerBaseURL {
		t.Errorf("ServerBaseURL default: got %q, want %q", c.ServerBaseURL, defaultServerBaseURL)
	}
	if c.ClientBaseURL != defaultClientBaseURL {
		t.Errorf("ClientBaseURL default: got %q, want %q", c.ClientBaseURL, defaultClientBaseURL)
	}
	if got, want := c.GuildIDs, []string{"123", "456"}; !stringsEqual(got, want) {
		t.Errorf("GuildIDs: got %v, want %v", got, want)
	}
}

func TestConfigFromEnv_Overrides(t *testing.T) {
	t.Setenv("CMDCTRL_DISCORD_BOT_TOKEN", "bot-token")
	t.Setenv("CMDCTRL_DISCORD_APP_ID", "app-id")
	t.Setenv("CMDCTRL_DISCORD_GUILD_IDS", "789")
	t.Setenv("CMDCTRL_ADMIN_TOKEN", "admin-token")
	t.Setenv("CMDCTRL_SERVER_BASE_URL", "http://example:9000")
	t.Setenv("CMDCTRL_CLIENT_BASE_URL", "https://staging.example.io")

	c := ConfigFromEnv()

	if c.ServerBaseURL != "http://example:9000" {
		t.Errorf("ServerBaseURL override: got %q", c.ServerBaseURL)
	}
	if c.ClientBaseURL != "https://staging.example.io" {
		t.Errorf("ClientBaseURL override: got %q", c.ClientBaseURL)
	}
}

func TestSplitGuildIDs(t *testing.T) {
	cases := map[string][]string{
		"":            nil,
		"   ":         nil,
		"123":         {"123"},
		"123,456":     {"123", "456"},
		" 123 , 456 ": {"123", "456"},
		"123,,456":    {"123", "456"},
		"123 , ,456":  {"123", "456"},
		",,,":         nil,
		"a,b,c":       {"a", "b", "c"},
	}
	for in, want := range cases {
		got := splitGuildIDs(in)
		if !stringsEqual(got, want) {
			t.Errorf("splitGuildIDs(%q): got %v, want %v", in, got, want)
		}
	}
}

func TestConfigDisabled(t *testing.T) {
	if (Config{}).Disabled() != true {
		t.Error("empty config should be Disabled()")
	}
	if (Config{BotToken: "x"}).Disabled() != false {
		t.Error("config with BotToken should not be Disabled()")
	}
}

func TestConfigValidate(t *testing.T) {
	good := Config{
		BotToken:      "bot",
		AppID:         "app",
		GuildIDs:      []string{"g1"},
		ServerBaseURL: "http://x",
		AdminToken:    "admin",
		ClientBaseURL: "https://x",
	}
	if err := good.Validate(); err != nil {
		t.Fatalf("good config: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Config)
		wantIn string
	}{
		{"missing bot token", func(c *Config) { c.BotToken = "" }, "CMDCTRL_DISCORD_BOT_TOKEN"},
		{"missing app id", func(c *Config) { c.AppID = "" }, "CMDCTRL_DISCORD_APP_ID"},
		{"missing guild ids", func(c *Config) { c.GuildIDs = nil }, "CMDCTRL_DISCORD_GUILD_IDS"},
		{"missing admin token", func(c *Config) { c.AdminToken = "" }, "CMDCTRL_ADMIN_TOKEN"},
		{"missing server base url", func(c *Config) { c.ServerBaseURL = "" }, "CMDCTRL_SERVER_BASE_URL"},
		{"missing client base url", func(c *Config) { c.ClientBaseURL = "" }, "CMDCTRL_CLIENT_BASE_URL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := good
			tc.mutate(&c)
			err := c.Validate()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantIn) {
				t.Errorf("error %q does not mention %q", err.Error(), tc.wantIn)
			}
		})
	}
}

func TestGuildAllowed(t *testing.T) {
	c := Config{GuildIDs: []string{"alpha", "beta"}}
	if !c.GuildAllowed("alpha") {
		t.Error("alpha should be allowed")
	}
	if !c.GuildAllowed("beta") {
		t.Error("beta should be allowed")
	}
	if c.GuildAllowed("gamma") {
		t.Error("gamma should NOT be allowed")
	}
	if c.GuildAllowed("") {
		t.Error("empty guild should NOT be allowed")
	}
}

func TestConfigRedacted(t *testing.T) {
	c := Config{BotToken: "1234567", AdminToken: "abcdefghij", AppID: "app"}
	r := c.Redacted()
	if r["bot_token"] != "<7 chars>" {
		t.Errorf("bot_token redaction: %v", r["bot_token"])
	}
	if r["admin_token"] != "<10 chars>" {
		t.Errorf("admin_token redaction: %v", r["admin_token"])
	}
	if r["app_id"] != "app" {
		t.Errorf("app_id should be kept plain: %v", r["app_id"])
	}
}

func stringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
