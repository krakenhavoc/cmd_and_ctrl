package bot

// Tests for ADR 0095's 2026-09-25 amendment: a pasted list, through the
// paste modal, the Request these cards button and the Moxfield hint.

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

// recordedCall is one REST call the bot made to Discord.
type recordedCall struct {
	Method string
	Path   string
	Body   map[string]any
}

// discordRecorder stands in for Discord's REST API: every call is
// recorded and answered 200 "{}".
type discordRecorder struct {
	mu    sync.Mutex
	calls []recordedCall
}

func (r *discordRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	var body map[string]any
	if req.Body != nil {
		raw, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(raw, &body)
	}
	r.mu.Lock()
	r.calls = append(r.calls, recordedCall{Method: req.Method, Path: req.URL.Path, Body: body})
	r.mu.Unlock()
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString("{}")),
		Request:    req,
	}, nil
}

func (r *discordRecorder) snapshot() []recordedCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedCall(nil), r.calls...)
}

// callback returns the interaction callback (the first response).
func (r *discordRecorder) callback(t *testing.T) recordedCall {
	t.Helper()
	for _, c := range r.snapshot() {
		if strings.HasSuffix(c.Path, "/callback") {
			return c
		}
	}
	t.Fatalf("no interaction callback in %+v", r.snapshot())
	return recordedCall{}
}

// edit returns the last edit of the original response.
func (r *discordRecorder) edit(t *testing.T) recordedCall {
	t.Helper()
	calls := r.snapshot()
	for i := len(calls) - 1; i >= 0; i-- {
		if calls[i].Method == http.MethodPatch && strings.HasSuffix(calls[i].Path, "/messages/@original") {
			return calls[i]
		}
	}
	t.Fatalf("no response edit in %+v", calls)
	return recordedCall{}
}

func recordingSession(t *testing.T) (*discordgo.Session, *discordRecorder) {
	t.Helper()
	s, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	rec := &discordRecorder{}
	s.Client = &http.Client{Transport: rec, Timeout: 2 * time.Second}
	return s, rec
}

func member(id, name string) *discordgo.Member {
	return &discordgo.Member{User: &discordgo.User{ID: id, GlobalName: name}}
}

func commandInteraction(guild, name string, opts ...*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i1", AppID: "app", Token: "tok", GuildID: guild, Member: member("424242", "Alice"),
		Type: discordgo.InteractionApplicationCommand,
		Data: discordgo.ApplicationCommandInteractionData{Name: name, Options: opts},
	}}
}

func modalSubmit(guild, customID, text string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i2", AppID: "app", Token: "tok", GuildID: guild, Member: member("424242", "Alice"),
		Type: discordgo.InteractionModalSubmit,
		Data: discordgo.ModalSubmitInteractionData{
			CustomID: customID,
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{Components: []discordgo.MessageComponent{
					&discordgo.TextInput{CustomID: deckPasteInputID, Value: text},
				}},
			},
		},
	}}
}

const pastedList = "1 Atraxa, Praetors' Voice *CMDR*\n1 Sol Ring\n1 Doubling Season"

func listReport() DeckCoverageReport {
	return DeckCoverageReport{
		Source: "text", DeckKey: "list:0123456789abcdef",
		Counts: map[DeckCoverageBucket]int{BucketManual: 1, BucketAutomated: 1, BucketNoEffect: 1},
		Cards:  []DeckCoverageCard{{Name: "Doubling Season", Bucket: BucketManual, Count: 1}},
	}
}

// --- the modal ---

func TestDeckPasteModal(t *testing.T) {
	resp := deckPasteModal(deckPasteCheckModalID)
	if resp.Type != discordgo.InteractionResponseModal || resp.Data.CustomID != deckPasteCheckModalID {
		t.Fatalf("response = %+v", resp)
	}
	if resp.Data.Title != "Paste your decklist" || utf8.RuneCountInString(resp.Data.Title) > 45 {
		t.Errorf("title = %q", resp.Data.Title)
	}
	if len(resp.Data.Components) != 1 {
		t.Fatalf("components = %+v", resp.Data.Components)
	}
	row := resp.Data.Components[0].(discordgo.ActionsRow)
	in := row.Components[0].(discordgo.TextInput)
	if in.CustomID != deckPasteInputID || in.Style != discordgo.TextInputParagraph || !in.Required || in.MaxLength != 4000 {
		t.Errorf("input = %+v", in)
	}
	// Discord's caps: a label is at most 45 characters, a placeholder 100.
	if n := utf8.RuneCountInString(in.Label); n > 45 || !strings.Contains(in.Label, "Moxfield") || !strings.Contains(in.Label, "Copy plain text") {
		t.Errorf("label = %q (%d chars)", in.Label, n)
	}
	if n := utf8.RuneCountInString(in.Placeholder); n > 100 || !strings.Contains(in.Placeholder, "1 Card Name") {
		t.Errorf("placeholder = %q (%d chars)", in.Placeholder, n)
	}
	if _, err := json.Marshal(resp); err != nil {
		t.Errorf("marshal: %v", err)
	}
}

// A submission as Discord sends it: rows and inputs decode as
// pointers, and modalTextValue reads them.
func TestModalTextValue_FromDiscordJSON(t *testing.T) {
	raw := `{"id":"1","application_id":"app","type":5,"token":"t","guild_id":"g1",
		"data":{"custom_id":"c2-deck-paste:req","components":[{"type":1,"components":[
			{"type":4,"custom_id":"decklist","value":"1 Sol Ring\n1 Arcane Signet"}]}]}}`
	var i discordgo.Interaction
	if err := json.Unmarshal([]byte(raw), &i); err != nil {
		t.Fatal(err)
	}
	data := i.ModalSubmitData()
	if data.CustomID != deckPasteReqModalID {
		t.Errorf("custom id = %q", data.CustomID)
	}
	if got := modalTextValue(data, deckPasteInputID); got != "1 Sol Ring\n1 Arcane Signet" {
		t.Errorf("value = %q", got)
	}
	if got := modalTextValue(data, "other"); got != "" {
		t.Errorf("unknown input = %q", got)
	}
}

// --- the button carries the list ---

func TestDeckRequestCustomID_PastedListRoundTripsThroughTheStore(t *testing.T) {
	store := newDeckSourceStore()
	now := time.Now()
	src := DeckSource{Text: strings.Repeat("1 Sol Ring\n", 100)}
	id := buildDeckRequestCustomID(src, store, now)
	if len(id) > discordCustomIDMax || !strings.HasPrefix(id, deckRequestCustomIDPrefix+"tok:") {
		t.Fatalf("custom id = %q", id)
	}
	got, err := resolveDeckRequestSource(id, store, now)
	if err != nil || got != src {
		t.Fatalf("round trip: %+v, %v", got, err)
	}
	// Even a list short enough to fit goes through the store: the
	// custom id only ever embeds a link.
	short := DeckSource{Text: "1 Sol Ring"}
	if id := buildDeckRequestCustomID(short, store, now); !strings.HasPrefix(id, deckRequestCustomIDPrefix+"tok:") {
		t.Errorf("short list custom id = %q", id)
	}
	if _, err := resolveDeckRequestSource(id, store, now.Add(deckSourceTTL+time.Second)); err != errDeckSourceExpired {
		t.Errorf("after the TTL: %v", err)
	}
}

func TestDeckCheckContent_PastedList(t *testing.T) {
	content := deckCheckContent(listReport(), "https://cmd.labxp.io/", DeckSource{Text: pastedList})
	if !strings.HasPrefix(content, "**Your pasted list**") {
		t.Errorf("content = %q", content)
	}
	if !strings.Contains(content, "https://cmd.labxp.io/#/deck-check") || strings.Contains(content, "?url=") {
		t.Errorf("a pasted check links the page plainly: %q", content)
	}
}

func TestDeckRequestOutcomeMessage_PastedList(t *testing.T) {
	src := DeckSource{Text: pastedList}
	content, visible := deckRequestOutcomeMessage(DeckRequestResult{
		Status: DeckRequestJoined, AlreadyRequested: true, IssueURL: "https://x/issues/3",
	}, src)
	if visible || !strings.Contains(content, "https://x/issues/3") || strings.Contains(content, "Sol Ring") {
		t.Errorf("already asked: %q", content)
	}
	content, _ = deckRequestOutcomeMessage(DeckRequestResult{Status: "weird"}, src)
	if !strings.Contains(content, "this list") {
		t.Errorf("unexpected status: %q", content)
	}
}

// --- the Moxfield hint ---

func TestDeckErrorMessages_MoxfieldHint(t *testing.T) {
	hinted := &DeckAPIError{
		StatusCode: http.StatusBadGateway, Code: "upstream_blocked", Hint: DeckHintPasteList,
		Message: "Moxfield blocks our server. On Moxfield, open the deck → Export → Copy plain text, then paste the list here instead.",
	}
	got := deckCoverageErrorMessage(hinted)
	if !strings.HasPrefix(got, "Moxfield blocks our server.") || !strings.Contains(got, "Run `/c2-deck-check` with no link") {
		t.Errorf("check: %q", got)
	}
	got = deckRequestErrorMessage(hinted)
	if !strings.HasPrefix(got, "Moxfield blocks our server.") || !strings.Contains(got, "Run `/c2-deck-req` with no link") {
		t.Errorf("request: %q", got)
	}
	plain := &DeckAPIError{StatusCode: http.StatusBadGateway, Code: "upstream_blocked", Message: "Archidekt is blocking us."}
	if got := deckCoverageErrorMessage(plain); got != "Archidekt is blocking us." {
		t.Errorf("unhinted: %q", got)
	}
}

// --- Dispatch, end to end against the fake server ---

func TestDispatch_DeckCommandsWithNoLinkShowTheModal(t *testing.T) {
	fs, c := newFakeServer(t)
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())
	for cmd, want := range map[string]string{CmdDeckCheck: deckPasteCheckModalID, CmdDeckReq: deckPasteReqModalID} {
		sess, rec := recordingSession(t)
		h.Dispatch(sess, commandInteraction("g1", cmd))
		cb := rec.callback(t)
		data, _ := cb.Body["data"].(map[string]any)
		if cb.Body["type"] != float64(discordgo.InteractionResponseModal) || data["custom_id"] != want {
			t.Errorf("%s: callback = %+v", cmd, cb.Body)
		}
		if len(rec.snapshot()) != 1 {
			t.Errorf("%s: a modal is the whole response, got %+v", cmd, rec.snapshot())
		}
	}
	if len(fs.gotAuthHeaders) != 0 {
		t.Errorf("showing the modal called the server: %v", fs.gotAuthHeaders)
	}
}

func TestDispatch_PasteModalSubmitChecksTheList(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckCoverageBody = listReport()
	h := NewHandler(Config{GuildIDs: []string{"g1"}, ClientBaseURL: "https://cmd.labxp.io"}, c, discardLogger())
	sess, rec := recordingSession(t)

	h.Dispatch(sess, modalSubmit("g1", deckPasteCheckModalID, "  "+pastedList+"\n"))

	if fs.gotDeckCoverageText != pastedList || fs.gotDeckCoverageURL != "" {
		t.Fatalf("server saw url %q text %q", fs.gotDeckCoverageURL, fs.gotDeckCoverageText)
	}
	if cb := rec.callback(t); cb.Body["type"] != float64(discordgo.InteractionResponseDeferredChannelMessageWithSource) {
		t.Errorf("submit should defer: %+v", cb.Body)
	}
	edit := rec.edit(t)
	content, _ := edit.Body["content"].(string)
	if !strings.Contains(content, "Your pasted list") || !strings.Contains(content, "Doubling Season") {
		t.Errorf("reply = %q", content)
	}

	// The reply's button carries the list: pressing it requests {text}.
	rows, _ := edit.Body["components"].([]any)
	if len(rows) != 1 {
		t.Fatalf("components = %+v", edit.Body["components"])
	}
	btn := rows[0].(map[string]any)["components"].([]any)[0].(map[string]any)
	customID, _ := btn["custom_id"].(string)
	fs.deckRequestBody = DeckRequestResult{Status: DeckRequestFiled, IssueURL: "https://x/issues/5", IssueNumber: 5}
	press := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i3", AppID: "app", Token: "tok", GuildID: "g1", Member: member("515151", "Bob"),
		Type: discordgo.InteractionMessageComponent,
		Data: discordgo.MessageComponentInteractionData{CustomID: customID},
	}}
	sess2, _ := recordingSession(t)
	h.Dispatch(sess2, press)
	if fs.gotDeckRequest.Text != pastedList || fs.gotDeckRequest.URL != "" || fs.gotDeckRequest.Requester.DiscordID != "515151" {
		t.Errorf("button request = %+v", fs.gotDeckRequest)
	}
}

func TestDispatch_PasteModalSubmitRequestsTheList(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.deckRequestBody = DeckRequestResult{Status: DeckRequestFiled, IssueURL: "https://x/issues/8", IssueNumber: 8,
		Report: &DeckCoverageReport{Counts: map[DeckCoverageBucket]int{BucketManual: 2}}}
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())
	sess, rec := recordingSession(t)

	h.Dispatch(sess, modalSubmit("g1", deckPasteReqModalID, pastedList))

	if fs.gotDeckRequest.Text != pastedList || fs.gotDeckRequest.Requester.DiscordID != "424242" ||
		fs.gotDeckRequest.Requester.DisplayName != "Alice" {
		t.Fatalf("server saw %+v", fs.gotDeckRequest)
	}
	// Filed is channel-visible: a follow-up with the issue link.
	var followup string
	for _, call := range rec.snapshot() {
		if call.Method == http.MethodPost && strings.HasSuffix(call.Path, "/webhooks/app/tok") {
			followup, _ = call.Body["content"].(string)
		}
	}
	if !strings.Contains(followup, "https://x/issues/8") {
		t.Errorf("follow-up = %q in %+v", followup, rec.snapshot())
	}
}

func TestDispatch_PasteModalSubmitRefusals(t *testing.T) {
	fs, c := newFakeServer(t)
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())

	for name, tc := range map[string]struct {
		i    *discordgo.InteractionCreate
		want string
	}{
		"another guild": {modalSubmit("g2", deckPasteCheckModalID, pastedList), "not authorized"},
		"empty list":    {modalSubmit("g1", deckPasteReqModalID, "   "), "Paste a decklist"},
		"unknown modal": {modalSubmit("g1", "c2-deck-paste:nope", pastedList), "went wrong"},
	} {
		sess, rec := recordingSession(t)
		h.Dispatch(sess, tc.i)
		cb := rec.callback(t)
		data, _ := cb.Body["data"].(map[string]any)
		content, _ := data["content"].(string)
		flags, _ := data["flags"].(float64)
		if !strings.Contains(content, tc.want) || int(flags)&int(discordgo.MessageFlagsEphemeral) == 0 {
			t.Errorf("%s: callback = %+v", name, cb.Body)
		}
	}
	if fs.gotDeckCoverageText != "" || fs.gotDeckRequest.Text != "" {
		t.Errorf("a refused submission reached the server: %q %+v", fs.gotDeckCoverageText, fs.gotDeckRequest)
	}
}
