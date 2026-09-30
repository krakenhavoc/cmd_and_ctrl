package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

// --- buildDeckRequestCustomID / resolveDeckRequestSource round trip ---

func TestDeckRequestCustomID_ShortLinkRoundTrips(t *testing.T) {
	store := newDeckSourceStore()
	now := time.Now()
	link := "https://moxfield.com/decks/AbC123"

	id := buildDeckRequestCustomID(DeckSource{URL: link}, store, now)
	if len(id) > discordCustomIDMax {
		t.Fatalf("custom id exceeds Discord's cap: %d chars", len(id))
	}
	if !strings.HasPrefix(id, deckRequestCustomIDPrefix+"link:") {
		t.Errorf("short link should be embedded directly, got %q", id)
	}

	got, err := resolveDeckRequestSource(id, store, now)
	if err != nil {
		t.Fatalf("resolveDeckRequestSource: %v", err)
	}
	if got != (DeckSource{URL: link}) {
		t.Errorf("round trip: got %+v, want %q", got, link)
	}
}

func TestDeckRequestCustomID_LongLinkUsesTokenStore(t *testing.T) {
	store := newDeckSourceStore()
	now := time.Now()
	// A link long enough that url.QueryEscape(link) plus the prefix
	// blows past Discord's 100-char cap.
	link := "https://archidekt.com/decks/1234567?" + strings.Repeat("query=value&", 10)

	id := buildDeckRequestCustomID(DeckSource{URL: link}, store, now)
	if len(id) > discordCustomIDMax {
		t.Fatalf("custom id exceeds Discord's cap: %d chars (%q)", len(id), id)
	}
	if !strings.HasPrefix(id, deckRequestCustomIDPrefix+"tok:") {
		t.Errorf("long link should fall back to a token, got %q", id)
	}

	got, err := resolveDeckRequestSource(id, store, now)
	if err != nil {
		t.Fatalf("resolveDeckRequestSource: %v", err)
	}
	if got != (DeckSource{URL: link}) {
		t.Errorf("round trip via token: got %+v, want %q", got, link)
	}
}

func TestDeckRequestCustomID_TokenExpires(t *testing.T) {
	store := newDeckSourceStore()
	now := time.Now()
	token := "tok-1"
	store.put(token, DeckSource{URL: "https://moxfield.com/decks/x"}, now)

	_, err := resolveDeckRequestSource(deckRequestCustomIDPrefix+"tok:"+token, store, now.Add(deckSourceTTL+time.Second))
	if !errors.Is(err, errDeckSourceExpired) {
		t.Errorf("want errDeckSourceExpired, got %v", err)
	}
}

func TestDeckRequestCustomID_UnknownToken(t *testing.T) {
	store := newDeckSourceStore()
	_, err := resolveDeckRequestSource(deckRequestCustomIDPrefix+"tok:missing", store, time.Now())
	if !errors.Is(err, errDeckSourceExpired) {
		t.Errorf("want errDeckSourceExpired for an unknown token, got %v", err)
	}
}

func TestDeckRequestCustomID_Malformed(t *testing.T) {
	store := newDeckSourceStore()
	cases := []string{
		deckRequestCustomIDPrefix + "noseparator",
		deckRequestCustomIDPrefix + "weird:",
		deckRequestCustomIDPrefix + "weird:x",
	}
	for _, id := range cases {
		if _, err := resolveDeckRequestSource(id, store, time.Now()); err == nil {
			t.Errorf("resolveDeckRequestSource(%q): want an error", id)
		}
	}
}

func TestDeckSourceStore_PutSweepsExpired(t *testing.T) {
	s := newDeckSourceStore()
	base := time.Now()
	s.put("old", DeckSource{URL: "link-a"}, base)
	// A later put, past "old"'s TTL, should sweep it — same pattern
	// endConfirmations.put already relies on.
	later := base.Add(deckSourceTTL + time.Second)
	s.put("new", DeckSource{Text: "1 Sol Ring"}, later)

	if _, found := s.get("old", later); found {
		t.Error("expired entry should have been swept on the next put")
	}
	if _, found := s.get("new", later); !found {
		t.Error("new entry should still be present")
	}
}

// --- deckCheckContent / deckCheckWantsButton ---

func sampleReport() DeckCoverageReport {
	return DeckCoverageReport{
		DeckName: "Needy Deck",
		Counts: map[DeckCoverageBucket]int{
			BucketManual: 2, BucketUnreviewed: 1, BucketCaveats: 6, BucketAutomated: 40, BucketNoEffect: 29,
		},
		Cards: []DeckCoverageCard{
			{Name: "Doubling Season", Bucket: BucketManual},
			{Name: "Genesis Wave", Bucket: BucketManual},
			{Name: "Forest", Bucket: BucketNoEffect},
		},
	}
}

func TestDeckCheckContent_NamesCountsAndLink(t *testing.T) {
	report := sampleReport()
	content := deckCheckContent(report, "https://cmd.labxp.io", DeckSource{URL: "https://moxfield.com/decks/AbC123"})

	if !strings.Contains(content, "Needy Deck") {
		t.Errorf("content missing deck name: %q", content)
	}
	if !strings.Contains(content, "2 need manual play") || !strings.Contains(content, "1 are unreviewed") {
		t.Errorf("content missing bucket counts: %q", content)
	}
	if !strings.Contains(content, "Doubling Season") || !strings.Contains(content, "Genesis Wave") {
		t.Errorf("content missing manual card names: %q", content)
	}
	if !strings.Contains(content, "https://cmd.labxp.io/#/deck-check?url=") {
		t.Errorf("content missing full-report link: %q", content)
	}
	if !strings.Contains(content, "https%3A%2F%2Fmoxfield.com%2Fdecks%2FAbC123") {
		t.Errorf("full-report link should url-encode the deck link: %q", content)
	}
}

func TestDeckCheckContent_EmptyDeckNameFallsBack(t *testing.T) {
	report := DeckCoverageReport{Counts: map[DeckCoverageBucket]int{}}
	content := deckCheckContent(report, "https://cmd.labxp.io", DeckSource{URL: "https://moxfield.com/decks/x"})
	if !strings.Contains(content, "This deck") {
		t.Errorf("want a fallback name, got %q", content)
	}
}

func TestManualCardNamesLine_TruncatesAtMax(t *testing.T) {
	var cards []DeckCoverageCard
	for i := 0; i < 20; i++ {
		cards = append(cards, DeckCoverageCard{Name: "Card", Bucket: BucketManual})
	}
	report := DeckCoverageReport{Cards: cards}

	line := manualCardNamesLine(report, 15)
	if !strings.Contains(line, "…and 5 more") {
		t.Errorf("want a truncation suffix, got %q", line)
	}
	// Exactly 15 names before the suffix.
	if got := strings.Count(line, "Card"); got != 15 {
		t.Errorf("want 15 shown names, counted %d occurrences in %q", got, line)
	}
}

func TestManualCardNamesLine_Empty(t *testing.T) {
	report := DeckCoverageReport{Cards: []DeckCoverageCard{{Name: "Forest", Bucket: BucketNoEffect}}}
	if got := manualCardNamesLine(report, 15); got != "" {
		t.Errorf("want empty string with no manual cards, got %q", got)
	}
}

func TestDeckCheckWantsButton(t *testing.T) {
	cases := []struct {
		name string
		want bool
		rep  DeckCoverageReport
	}{
		{"manual only", true, DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketManual: 1}}},
		{"unreviewed only", true, DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketUnreviewed: 1}}},
		{"neither", false, DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketAutomated: 40, BucketCaveats: 6, BucketNoEffect: 29}}},
		{"nil counts", false, DeckCoverageReport{}},
	}
	for _, c := range cases {
		if got := deckCheckWantsButton(c.rep); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestDeckCheckReply_IncludesButtonWhenSomethingToRequest(t *testing.T) {
	store := newDeckSourceStore()
	now := time.Now()
	report := sampleReport() // has manual cards
	link := "https://moxfield.com/decks/AbC123"

	content, components := deckCheckReply(report, "https://cmd.labxp.io", DeckSource{URL: link}, store, now)
	if content == "" {
		t.Fatal("want non-empty content")
	}
	if len(components) != 1 {
		t.Fatalf("want one action row with the button, got %d", len(components))
	}
	row, ok := components[0].(discordgo.ActionsRow)
	if !ok || len(row.Components) != 1 {
		t.Fatalf("want a row with one button, got %+v", components[0])
	}
	btn, ok := row.Components[0].(discordgo.Button)
	if !ok || btn.Label != "Request these cards" {
		t.Fatalf("want the Request these cards button, got %+v", row.Components[0])
	}
	// The button's custom ID must resolve back to the checked link.
	got, err := resolveDeckRequestSource(btn.CustomID, store, now)
	if err != nil {
		t.Fatalf("resolveDeckRequestSource: %v", err)
	}
	if got != (DeckSource{URL: link}) {
		t.Errorf("button custom id resolves to %+v, want %q", got, link)
	}
}

func TestDeckCheckReply_NoButtonWhenNothingToRequest(t *testing.T) {
	report := DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketAutomated: 40, BucketCaveats: 6, BucketNoEffect: 29}}
	_, components := deckCheckReply(report, "https://cmd.labxp.io", DeckSource{URL: "https://moxfield.com/decks/x"}, newDeckSourceStore(), time.Now())
	if components != nil {
		t.Errorf("want no button when nothing needs requesting, got %+v", components)
	}
}

// --- checkDeck (the createInviteGame-style server call) ---

func TestCheckDeck_Success(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageBody = DeckCoverageReport{DeckName: "Needy Deck", Counts: map[DeckCoverageBucket]int{BucketManual: 2}}
	h := NewHandler(Config{}, c, discardLogger())

	report, err := h.checkDeck(context.Background(), DeckSource{URL: "https://moxfield.com/decks/AbC123"})
	if err != nil {
		t.Fatalf("checkDeck: %v", err)
	}
	if report.DeckName != "Needy Deck" {
		t.Errorf("report: got %+v", report)
	}
	if fs.gotDeckCoverageURL != "https://moxfield.com/decks/AbC123" {
		t.Errorf("server saw url %q", fs.gotDeckCoverageURL)
	}
}

func TestDeckCoverageErrorMessage(t *testing.T) {
	apiErr := &DeckAPIError{StatusCode: http.StatusUnprocessableEntity, Message: "That deck is private."}
	if got := deckCoverageErrorMessage(apiErr); got != "That deck is private." {
		t.Errorf("want the server's sentence, got %q", got)
	}
	if got := deckCoverageErrorMessage(ErrServerUnreachable); !strings.Contains(got, "not reachable") {
		t.Errorf("got %q", got)
	}
	if got := deckCoverageErrorMessage(ErrUnauthorized); !strings.Contains(got, "not authorized") {
		t.Errorf("got %q", got)
	}
	if got := deckCoverageErrorMessage(errors.New("boom")); !strings.Contains(got, "went wrong") {
		t.Errorf("got %q", got)
	}
}

// --- requestDeck / deckRequestOutcomeMessage / errors ---

func TestRequestDeck_Success(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckRequestBody = DeckRequestResult{Status: DeckRequestFiled, IssueURL: "https://x/issues/1"}
	h := NewHandler(Config{}, c, discardLogger())

	result, err := h.requestDeck(context.Background(), DeckSource{URL: "https://moxfield.com/decks/x"}, DeckRequester{DiscordID: "1", DisplayName: "Alice"})
	if err != nil {
		t.Fatalf("requestDeck: %v", err)
	}
	if result.Status != DeckRequestFiled {
		t.Errorf("result: got %+v", result)
	}
	if fs.gotDeckRequest.Requester.DisplayName != "Alice" {
		t.Errorf("requester not forwarded: got %+v", fs.gotDeckRequest)
	}
}

func TestDeckRequestOutcomeMessage(t *testing.T) {
	filed := DeckRequestResult{
		Status: DeckRequestFiled, IssueURL: "https://x/issues/7",
		Report: &DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketManual: 5, BucketUnreviewed: 2}},
	}
	content, visible := deckRequestOutcomeMessage(filed, DeckSource{URL: "https://moxfield.com/decks/x"})
	if !visible {
		t.Error("filed should be channel-visible")
	}
	if !strings.Contains(content, "https://x/issues/7") || !strings.Contains(content, "7 cards to add") {
		t.Errorf("filed content: got %q", content)
	}

	joined := DeckRequestResult{Status: DeckRequestJoined, IssueURL: "https://x/issues/7"}
	content, visible = deckRequestOutcomeMessage(joined, DeckSource{URL: "link"})
	if !visible {
		t.Error("joined (fresh comment) should be channel-visible")
	}
	if !strings.Contains(content, "https://x/issues/7") {
		t.Errorf("joined content: got %q", content)
	}

	alreadyAsked := DeckRequestResult{Status: DeckRequestJoined, IssueURL: "https://x/issues/7", AlreadyRequested: true}
	content, visible = deckRequestOutcomeMessage(alreadyAsked, DeckSource{URL: "https://moxfield.com/decks/x"})
	if visible {
		t.Error("already_requested should be ephemeral")
	}
	if !strings.Contains(content, "already asked") {
		t.Errorf("already_requested content: got %q", content)
	}

	nothing := DeckRequestResult{Status: DeckRequestNothingToAdd, Report: &DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketAutomated: 40}}}
	content, visible = deckRequestOutcomeMessage(nothing, DeckSource{URL: "link"})
	if visible {
		t.Error("nothing_to_add should be ephemeral")
	}
	if !strings.Contains(content, "already in the engine") || !strings.Contains(content, "40 automated") {
		t.Errorf("nothing_to_add content: got %q", content)
	}

	limited := DeckRequestResult{Status: DeckRequestRateLimited, RetryAfter: 3661}
	content, visible = deckRequestOutcomeMessage(limited, DeckSource{URL: "link"})
	if visible {
		t.Error("rate_limited should be ephemeral")
	}
	if !strings.Contains(content, "1h 1m") {
		t.Errorf("rate_limited content should carry the wait: got %q", content)
	}
}

func TestRetryAfterWords(t *testing.T) {
	cases := map[int]string{
		0:    "less than a minute",
		-5:   "less than a minute",
		30:   "less than a minute",
		90:   "1m",
		3600: "1h",
		3661: "1h 1m",
		7322: "2h 2m",
	}
	for secs, want := range cases {
		if got := retryAfterWords(secs); got != want {
			t.Errorf("retryAfterWords(%d): got %q, want %q", secs, got, want)
		}
	}
}

func TestDeckRequestErrorMessage(t *testing.T) {
	unavailable := &DeckAPIError{StatusCode: http.StatusServiceUnavailable, Message: "CMDCTRL_GITHUB_TOKEN not set"}
	if got := deckRequestErrorMessage(unavailable); got != "Deck requests aren't set up on this server." {
		t.Errorf("503 should get the fixed player-facing message, got %q", got)
	}

	other := &DeckAPIError{StatusCode: http.StatusBadGateway, Message: "GitHub is unavailable"}
	if got := deckRequestErrorMessage(other); got != "GitHub is unavailable" {
		t.Errorf("non-503 should use the server's sentence, got %q", got)
	}

	if got := deckRequestErrorMessage(ErrServerUnreachable); !strings.Contains(got, "not reachable") {
		t.Errorf("got %q", got)
	}
	if got := deckRequestErrorMessage(ErrUnauthorized); !strings.Contains(got, "not authorized") {
		t.Errorf("got %q", got)
	}
}

func TestDeckSourceResolveErrorMessage(t *testing.T) {
	if got := deckSourceResolveErrorMessage(errDeckSourceExpired); !strings.Contains(got, "expired") {
		t.Errorf("got %q", got)
	}
	if got := deckSourceResolveErrorMessage(errors.New("bad")); !strings.Contains(got, "c2-deck-check") {
		t.Errorf("got %q", got)
	}
}

// --- invokerDisplayName ---

func TestInvokerDisplayName_PrefersNickThenGlobalNameThenUsername(t *testing.T) {
	nick := &discordgo.Interaction{Member: &discordgo.Member{
		Nick: "Nicky",
		User: &discordgo.User{GlobalName: "Global", Username: "raw"},
	}}
	if got := invokerDisplayName(nick); got != "Nicky" {
		t.Errorf("got %q, want Nicky", got)
	}

	global := &discordgo.Interaction{Member: &discordgo.Member{
		User: &discordgo.User{GlobalName: "Global", Username: "raw"},
	}}
	if got := invokerDisplayName(global); got != "Global" {
		t.Errorf("got %q, want Global", got)
	}

	username := &discordgo.Interaction{Member: &discordgo.Member{
		User: &discordgo.User{Username: "raw"},
	}}
	if got := invokerDisplayName(username); got != "raw" {
		t.Errorf("got %q, want raw", got)
	}

	dm := &discordgo.Interaction{User: &discordgo.User{GlobalName: "DM Global", Username: "dmraw"}}
	if got := invokerDisplayName(dm); got != "DM Global" {
		t.Errorf("DM invoker: got %q, want DM Global", got)
	}

	if got := invokerDisplayName(&discordgo.Interaction{}); got != "" {
		t.Errorf("no member or user: got %q, want empty", got)
	}
	if got := invokerDisplayName(nil); got != "" {
		t.Errorf("nil interaction: got %q, want empty", got)
	}
}

func TestDeckRequester_CarriesInvokerID(t *testing.T) {
	i := &discordgo.Interaction{Member: &discordgo.Member{
		User: &discordgo.User{ID: "424242", GlobalName: "Alice"},
	}}
	req := deckRequester(i)
	if req.DiscordID != "424242" || req.DisplayName != "Alice" {
		t.Errorf("got %+v", req)
	}
}

// --- countsSentence / manualPlusUnreviewed ---

func TestManualPlusUnreviewed(t *testing.T) {
	if got := manualPlusUnreviewed(nil); got != 0 {
		t.Errorf("nil report: got %d", got)
	}
	report := &DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketManual: 3, BucketUnreviewed: 2}}
	if got := manualPlusUnreviewed(report); got != 5 {
		t.Errorf("got %d, want 5", got)
	}
}
