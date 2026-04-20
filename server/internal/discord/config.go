// Package discord implements the OAuth + identity-fetch half of
// the S12.5 Discord integration. Slash-command bot and Rich
// Presence live in separate tracks; this package is everything a
// user hits when they click "Sign in with Discord" on an invite
// link.
//
// Dependency-free apart from the Go standard library — we talk to
// Discord with plain net/http and encoding/json rather than pull
// in a third-party SDK, because the surface we touch is small
// (OAuth exchange + /users/@me) and the project already follows
// a "stdlib-first" convention.
package discord

import (
	"os"
	"strings"
)

// Config carries the values the server needs to drive the OAuth
// dance. All three are required for the integration to be live;
// any empty value disables the feature (the /auth/discord/start
// handler returns 503 and the client hides the sign-in button via
// /auth/discord/config).
//
// ClientID + ClientSecret come from the Discord Developer Portal
// (https://discord.com/developers/applications → General
// Information + OAuth2). RedirectURI must exactly match one of
// the callback URLs registered in the app's OAuth2 settings;
// mismatch is Discord's most common debug paper-cut, so the
// callback handler surfaces the redirect mismatch error code
// verbatim rather than collapsing every 4xx to "invalid code".
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Enabled reports whether the config has the three values needed
// to actually run the OAuth flow. Handlers short-circuit on false
// with a 503 so a misconfigured prod deploy surfaces a clear
// error rather than hitting Discord with empty credentials.
func (c Config) Enabled() bool {
	return c.ClientID != "" && c.ClientSecret != "" && c.RedirectURI != ""
}

// ConfigFromEnv reads the three CMDCTRL_DISCORD_* env vars.
// Missing / blank values yield an unfilled Config (Enabled() ==
// false); the caller is expected to log this and continue rather
// than fail boot, so a dev stack without a Discord app still
// starts fine.
//
// The same systemd EnvironmentFile that feeds CMDCTRL_ADMIN_TOKEN
// on prod (/etc/cmd_and_ctrl/env per the S12 deploy) is the
// intended source. Dev is expected to export them in the shell
// before `make dev`.
func ConfigFromEnv() Config {
	return Config{
		ClientID:     strings.TrimSpace(os.Getenv("CMDCTRL_DISCORD_CLIENT_ID")),
		ClientSecret: strings.TrimSpace(os.Getenv("CMDCTRL_DISCORD_CLIENT_SECRET")),
		RedirectURI:  strings.TrimSpace(os.Getenv("CMDCTRL_DISCORD_REDIRECT_URI")),
	}
}
