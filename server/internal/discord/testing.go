package discord

// SwapEndpointsForTesting overrides the package-level token + user
// endpoint URLs and returns a function that restores the originals.
// Intended for tests in other packages (lobby's HTTP test wires
// /auth/discord/callback against an httptest.Server stub) that
// need to redirect Discord traffic without going over the live
// network. Same package's _test.go files mutate the vars directly.
//
// Two URLs are enough for the OAuth flow we use; the authorize
// endpoint is only used by AuthorizeURL, which doesn't make a
// request — the browser does, and tests inspect the URL string
// rather than visit it.
func SwapEndpointsForTesting(token, user string) func() {
	origToken := tokenEndpoint
	origUser := userEndpoint
	tokenEndpoint = token
	userEndpoint = user
	return func() {
		tokenEndpoint = origToken
		userEndpoint = origUser
	}
}

// SwapBotAPIBaseForTesting overrides the REST root the bot's DM calls
// use and returns a function that restores the original. Intended for
// tests in other packages (lobby's DM-invite test drives the route
// against an httptest.Server standing in for Discord).
func SwapBotAPIBaseForTesting(base string) func() {
	orig := botAPIBase
	botAPIBase = base
	return func() { botAPIBase = orig }
}
