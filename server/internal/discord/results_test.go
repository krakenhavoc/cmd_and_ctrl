package discord

// Tests for results.go (ADR 0174 §2, §5, §6), against the httptest.Server
// voice_test.go stands up for Discord's REST API.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

const testResultsChannel = "666666666666666666"

func newResultsStub(t *testing.T) (*voiceStub, Results) {
	t.Helper()
	stub, _ := newVoiceStub(t)
	return stub, Results{Token: testToken, GuildID: testGuild, ChannelID: testResultsChannel}
}

func sampleResults() ResultsPost {
	return ResultsPost{
		Table: "  Friday pod  ",
		Seats: []ResultsSeat{
			{Name: "Alice", Commanders: []string{"Atraxa, Praetors' Voice"}, Place: 1, Result: "win"},
			{Name: "Gus", Commanders: []string{"Tymna the Weaver", "Thrasios, Triton Hero"}, Place: 2, Result: "loss", EliminatedTurn: 9},
			{Name: "Claude", Kind: "AI agent", Commanders: []string{"Korvold, Fae-Cursed King"}, Place: 3, Result: "loss", Conceded: true, EliminatedTurn: 6},
			{Name: "Bot Brian", Kind: "bot", Place: 4, Result: "loss", EliminatedTurn: 4},
		},
		Turns:    11,
		Duration: 72 * time.Minute,
		EndedAt:  time.Date(2026, 10, 14, 21, 30, 0, 0, time.UTC),
	}
}

func TestResultsFromEnv(t *testing.T) {
	cases := []struct {
		name        string
		vals        map[string]string
		wantErr     string
		wantEnabled bool
		wantMissing bool
	}{
		{name: "unset is off", vals: map[string]string{}},
		{name: "guild and token without a channel is off",
			vals: map[string]string{CommunityGuildEnv: testGuild, BotTokenEnv: testToken}},
		{name: "channel, guild and token is on",
			vals:        map[string]string{ResultsChannelEnv: " " + testResultsChannel + " ", CommunityGuildEnv: testGuild, BotTokenEnv: testToken},
			wantEnabled: true},
		{name: "no token is off but boots",
			vals:        map[string]string{ResultsChannelEnv: testResultsChannel, CommunityGuildEnv: testGuild},
			wantMissing: true},
		{name: "the channel without the guild fails",
			vals:    map[string]string{ResultsChannelEnv: testResultsChannel, BotTokenEnv: testToken},
			wantErr: CommunityGuildEnv},
		{name: "a malformed channel fails",
			vals:    map[string]string{ResultsChannelEnv: "#game-results", CommunityGuildEnv: testGuild},
			wantErr: ResultsChannelEnv},
		{name: "a malformed guild fails",
			vals:    map[string]string{ResultsChannelEnv: testResultsChannel, CommunityGuildEnv: "c2"},
			wantErr: CommunityGuildEnv},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ResultsFromEnv(env(tc.vals))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want one naming %s", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if r.Enabled() != tc.wantEnabled || r.MissingToken() != tc.wantMissing {
				t.Errorf("Enabled %v MissingToken %v, want %v %v", r.Enabled(), r.MissingToken(), tc.wantEnabled, tc.wantMissing)
			}
			if tc.wantEnabled && r.ChannelID != testResultsChannel {
				t.Errorf("channel not trimmed: %q", r.ChannelID)
			}
		})
	}
}

func TestResultsOffNeverCallsDiscord(t *testing.T) {
	stub, _ := newResultsStub(t)
	off := Results{GuildID: testGuild, ChannelID: testResultsChannel} // no token
	if err := off.Post(context.Background(), nil, sampleResults()); !errors.Is(err, ErrResultsNotConfigured) {
		t.Errorf("Post err = %v", err)
	}
	if n := len(stub.seen()); n != 0 {
		t.Errorf("%d calls reached Discord with the feed off", n)
	}
}

func TestResultsPostCardAndNoMentions(t *testing.T) {
	stub, r := newResultsStub(t)
	stub.answer("POST", "/channels/"+testResultsChannel+"/messages", http.StatusOK, `{"id":"900"}`)
	if err := r.Post(context.Background(), nil, sampleResults()); err != nil {
		t.Fatal(err)
	}
	c := stub.seen()[0]
	if c.Method != "POST" || c.Auth != "Bot "+testToken {
		t.Errorf("request = %s, auth %q", c.Method, c.Auth)
	}
	if _, has := c.Body["content"]; has {
		t.Errorf("content = %v, want none", c.Body["content"])
	}
	am := c.Body["allowed_mentions"].(map[string]any)
	if parse, ok := am["parse"].([]any); !ok || len(parse) != 0 || am["roles"] != nil || am["users"] != nil {
		t.Errorf("allowed_mentions = %v, want parse: [] and nothing else", am)
	}
	embed := c.Body["embeds"].([]any)[0].(map[string]any)
	if embed["title"] != "Friday pod" || embed["description"] != "**Winner:** Alice" ||
		embed["timestamp"] != "2026-10-14T21:30:00Z" {
		t.Errorf("embed = %v", embed)
	}
	var names, values []string
	for _, raw := range embed["fields"].([]any) {
		m := raw.(map[string]any)
		names = append(names, m["name"].(string))
		values = append(values, m["value"].(string))
	}
	wantNames := []string{"1st · Alice", "2nd · Gus", "3rd · Claude (AI agent)", "4th · Bot Brian (bot)", "Turns", "Duration"}
	if strings.Join(names, "|") != strings.Join(wantNames, "|") {
		t.Errorf("field names = %q, want %q", names, wantNames)
	}
	wantValues := []string{
		"Atraxa, Praetors' Voice",
		"Tymna the Weaver / Thrasios, Triton Hero\nEliminated on turn 9",
		"Korvold, Fae-Cursed King\nConceded on turn 6",
		"No commander\nEliminated on turn 4",
		"11",
		"1h 12m",
	}
	if strings.Join(values, "|") != strings.Join(wantValues, "|") {
		t.Errorf("field values = %q, want %q", values, wantValues)
	}
	pretty, _ := json.MarshalIndent(c.Body, "", "  ")
	t.Logf("example post body:\n%s", pretty)
}

func TestResultsDrawAndHostileNames(t *testing.T) {
	p := ResultsPost{
		Table: "@everyone",
		Seats: []ResultsSeat{
			{Name: "<@123456789> **boss**", Place: 1, Result: "draw", Commanders: []string{"@here"}},
			{Name: "Bob\nspam", Place: 1, Result: "draw"},
		},
		Turns: 8,
	}
	m := p.Render()
	raw, _ := json.Marshal(m)
	if strings.Contains(string(raw), `"content"`) {
		t.Errorf("a results post has a content line: %s", raw)
	}
	e := m.Embeds[0]
	if e.Description != `**Draw** between \<\@123456789\> \*\*boss\*\*, Bob spam` {
		t.Errorf("description = %q", e.Description)
	}
	if e.Fields[0].Name != `1st · \<\@123456789\> \*\*boss\*\*` || e.Fields[0].Value != `\@here` {
		t.Errorf("field 0 = %+v", e.Fields[0])
	}
	// No duration known: the field is left out.
	for _, f := range e.Fields {
		if f.Name == "Duration" {
			t.Errorf("duration field with no duration: %+v", f)
		}
	}
	if len(m.AllowedMentions.Parse) != 0 || m.AllowedMentions.Roles != nil {
		t.Errorf("allowed_mentions = %+v", m.AllowedMentions)
	}
}

func TestResultsErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   error
	}{
		{http.StatusForbidden, `{"message":"Missing Access","code":50001}`, ErrResultsForbidden},
		{http.StatusNotFound, `{"message":"Unknown Channel","code":10003}`, ErrResultsChannelNotFound},
		{http.StatusUnauthorized, `{}`, ErrBotUnauthorized},
		{http.StatusBadGateway, `{}`, ErrVoiceUnavailable},
	}
	for _, tc := range cases {
		stub, r := newResultsStub(t)
		stub.answer("POST", "/channels/"+testResultsChannel+"/messages", tc.status, tc.body)
		if err := r.Post(context.Background(), nil, sampleResults()); !errors.Is(err, tc.want) {
			t.Errorf("status %d: err = %v, want %v", tc.status, err, tc.want)
		}
	}
}
