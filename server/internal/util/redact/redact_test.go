package redact

import (
	"net/url"
	"strings"
	"testing"
)

// token is a realistic session token: 32 random bytes, base64url, the
// shape internal/util/token mints.
const token = "Zk3q9_Rb-2xVn8LmPq4tYw7cHs1dJe6uKo0aBf5gTiA"

const invite = "iNv1te-T0ken_abcdefghijklmnop"

func TestSecretsRedacts(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{
			"ws URL keeps game and player",
			"connected to wss://cmd.example/ws?game=g-1&token=" + token + "&player=p-9",
			"connected to wss://cmd.example/ws?game=g-1&token=REDACTED&player=p-9",
		},
		{"token last", "/games/1/replay?token=" + token, "/games/1/replay?token=REDACTED"},
		{
			// auth.HMACAuthenticator's shape (#517): the dots must not
			// end the value, or the signed payload and signature leak.
			"signed session token",
			"/ws?token=v1.eyJyIjoicGxheWVyIn0.Zk3q9_Rb-2xVn8LmPq4tYw7cHs1dJe6uKo0aBf5gTiA&game=g-1",
			"/ws?token=REDACTED&game=g-1",
		},
		{
			"signed session bearer",
			"Authorization: Bearer v1.eyJyIjoicGxheWVyIn0.Zk3q9_Rb-2xVn8LmPq4tYw7cHs1dJe6uKo0aBf5gTiA",
			"Authorization: Bearer REDACTED",
		},
		{"token before fragment", "/ws?token=" + token + "#frag", "/ws?token=REDACTED#frag"},
		{
			"invite t in a hash route",
			"https://h/#/games/abc/join?t=" + invite + "&spectator=1",
			"https://h/#/games/abc/join?t=REDACTED&spectator=1",
		},
		{"reclaim t", "#/games/abc/reclaim?t=" + invite, "#/games/abc/reclaim?t=REDACTED"},
		{"html-escaped &amp;t", "join?game=1&amp;t=" + invite, "join?game=1&amp;t=REDACTED"},
		{
			"oauth tokens",
			"cb?access_token=" + token + "&refresh_token=r1&id_token=i2&client_secret=s3&scope=identify",
			"cb?access_token=REDACTED&refresh_token=REDACTED&id_token=REDACTED&client_secret=REDACTED&scope=identify",
		},
		{"bare token in prose", "failed with token=" + token + " after retry", "failed with token=REDACTED after retry"},
		{"case-insensitive key", "?Token=" + token + "&ACCESS_TOKEN=x", "?Token=REDACTED&ACCESS_TOKEN=REDACTED"},
		{"value with escapes", "?token=ab%2Bcd%2Fef&game=1", "?token=REDACTED&game=1"},
		{
			"url-encoded nested URL keeps trailing params",
			"https://h/login?next=" + url.QueryEscape("/ws?token="+token+"&game=g-1"),
			"https://h/login?next=%2Fws%3Ftoken%3DREDACTED%26game%3Dg-1",
		},
		{"double-encoded equals", "token%253D" + token, "token%253DREDACTED"},
		{"double-encoded equals, upper-case key and hex", "TOKEN%253d" + token, "TOKEN%253dREDACTED"},
		{
			"double-encoded nested URL keeps trailing params",
			"https://h/r?u=" + url.QueryEscape(url.QueryEscape("/ws?token="+token+"&game=g-1")),
			"https://h/r?u=%252Fws%253Ftoken%253DREDACTED%2526game%253Dg-1",
		},
		{
			"double-encoded invite t",
			"next=" + url.QueryEscape(url.QueryEscape("#/games/abc/join?t="+invite+"&x=1")),
			"next=%2523%252Fgames%252Fabc%252Fjoin%253Ft%253DREDACTED%2526x%253D1",
		},
		{"authorization bearer", "Authorization: Bearer " + token, "Authorization: Bearer REDACTED"},
		{"encoded bearer", "Authorization%3A%20Bearer%20" + token, "Authorization%3A%20Bearer%20REDACTED"},
		{"json token", `{"token":"` + token + `","gameID":"g"}`, `{"token":"REDACTED","gameID":"g"}`},
		{"escaped json token", `"{\"token\": \"` + token + `\"}"`, `"{\"token\": \"REDACTED\"}"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Secrets(tc.in)
			if got != tc.want {
				t.Fatalf("Secrets(%q)\n got: %q\nwant: %q", tc.in, got, tc.want)
			}
			if strings.Contains(got, token) || strings.Contains(got, invite) {
				t.Fatalf("credential survived: %q", got)
			}
			if again := Secrets(got); again != got {
				t.Fatalf("not idempotent: %q -> %q", got, again)
			}
		})
	}
}

func TestSecretsEncodedInvite(t *testing.T) {
	got := Secrets("redirect=" + url.QueryEscape("#/games/abc/join?t="+invite))
	if strings.Contains(got, invite) || !strings.Contains(got, "t%3DREDACTED") {
		t.Fatalf("encoded invite not redacted: %q", got)
	}
}

// TestSecretsLeavesOrdinaryTextAlone pins the false positives that would
// make the redactor cost triage information.
func TestSecretsLeavesOrdinaryTextAlone(t *testing.T) {
	for _, s := range []string{
		"",
		"server error code=bad_request message=not your priority",
		"action create_token id=1234abcd",
		"snapshot seq=12 turn=3 step=draw",
		`chat from=Ada text="at t=3 I attacked"`,
		"card art failed to load: id=abc face=0 size=small",
		"wss://cmd.example/ws?game=g-1&player=p-9",
		`{"token":true,"name":"Bearer of the Heavens","type_line":"Token Creature"}`,
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
	} {
		if got := Secrets(s); got != s {
			t.Errorf("Secrets(%q) = %q, want unchanged", s, got)
		}
	}
}
