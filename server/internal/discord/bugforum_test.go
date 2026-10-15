package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const (
	testBugForum = "444444444444444444"
	testThreadID = "555555555555555555"
)

func newBugForumStub(t *testing.T) (*voiceStub, BugForum) {
	t.Helper()
	stub, _ := newVoiceStub(t)
	return stub, BugForum{Token: testToken, ChannelID: testBugForum}
}

func sampleBugThread() BugThread {
	return BugThread{
		IssueNumber: 812,
		IssueURL:    "https://github.com/o/r/issues/812",
		Title:       "Stack panel clips\nthe last item",
		Kind:        "bug",
		Description: "It cuts off.\n@everyone <@&99> **bold** [link](https://evil.example)",
		ReporterID:  "501",
		Images:      []string{"https://cmd.example/bugreport/abc/img1.png", "https://cmd.example/bugreport/abc/img2.png", "javascript:alert(1)"},
	}
}

func TestBugForumFromEnv(t *testing.T) {
	cases := []struct {
		name        string
		vals        map[string]string
		wantErr     string
		wantEnabled bool
		wantMissing bool
	}{
		{name: "unset is off", vals: map[string]string{BotTokenEnv: testToken}},
		{name: "forum and token is on", vals: map[string]string{BugForumEnv: testBugForum, BotTokenEnv: " " + testToken + " "}, wantEnabled: true},
		{name: "no token is off but boots", vals: map[string]string{BugForumEnv: testBugForum}, wantMissing: true},
		{name: "malformed forum fails", vals: map[string]string{BugForumEnv: "bug-reports", BotTokenEnv: testToken}, wantErr: BugForumEnv},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f, err := BugForumFromEnv(env(tc.vals))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want one naming %s", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if f.Enabled() != tc.wantEnabled || f.MissingToken() != tc.wantMissing {
				t.Errorf("Enabled/MissingToken = %v/%v, want %v/%v", f.Enabled(), f.MissingToken(), tc.wantEnabled, tc.wantMissing)
			}
		})
	}
}

func TestBugThreadRenderNeverPingsAnyoneButTheReporter(t *testing.T) {
	raw, err := json.Marshal(sampleBugThread().Render())
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Name    string `json:"name"`
		Message struct {
			Content string `json:"content"`
			Embeds  []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
				Image       *struct {
					URL string `json:"url"`
				} `json:"image"`
				Fields []struct{ Name, Value string } `json:"fields"`
			} `json:"embeds"`
			AllowedMentions struct {
				Parse []string `json:"parse"`
				Roles []string `json:"roles"`
				Users []string `json:"users"`
			} `json:"allowed_mentions"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != "#812 Stack panel clips the last item" {
		t.Errorf("name = %q", got.Name)
	}
	am := got.Message.AllowedMentions
	if am.Parse == nil || len(am.Parse) != 0 || len(am.Roles) != 0 || len(am.Users) != 1 || am.Users[0] != "501" {
		t.Errorf("allowed_mentions = %+v, want parse [] and users [501] only", am)
	}
	if !strings.HasPrefix(got.Message.Content, "<@501> ") {
		t.Errorf("content = %q, want it to open by mentioning the reporter", got.Message.Content)
	}
	e := got.Message.Embeds[0]
	if e.URL != "https://github.com/o/r/issues/812" || e.Title != "GitHub issue #812" {
		t.Errorf("embed title/url = %q / %q", e.Title, e.URL)
	}
	// The reporter's text is typed, not markup: every mention and link
	// character is escaped, line breaks kept.
	for _, bad := range []string{"@", "<", "*", "[", "]", ":"} {
		for i := 0; i < len(e.Description); i++ {
			if e.Description[i:i+1] == bad && (i == 0 || e.Description[i-1] != '\\') {
				t.Errorf("description %q carries an unescaped %q at %d", e.Description, bad, i)
			}
		}
	}
	if !strings.Contains(e.Description, "It cuts off.\n") {
		t.Errorf("description lost its line break: %q", e.Description)
	}
	if e.Image == nil || e.Image.URL != "https://cmd.example/bugreport/abc/img1.png" {
		t.Errorf("image = %+v", e.Image)
	}
	var shots string
	for _, f := range e.Fields {
		if f.Name == "Screenshots" {
			shots = f.Value
		}
	}
	if !strings.Contains(shots, "img2.png") || strings.Contains(shots, "javascript:") {
		t.Errorf("screenshots field = %q: want both http(s) links and no other scheme", shots)
	}
}

func TestBugThreadRenderGuestAndLimits(t *testing.T) {
	th := BugThread{IssueNumber: 5, IssueURL: "https://x/5", Title: strings.Repeat("t", 300), Description: strings.Repeat("d", 9000), ReporterID: "not-a-snowflake"}
	body := th.Render()
	if n := len([]rune(body.Name)); n > maxThreadName {
		t.Errorf("thread name is %d runes, Discord's limit is %d", n, maxThreadName)
	}
	if n := len([]rune(body.Message.Embeds[0].Description)); n > maxBugForumDesc {
		t.Errorf("description is %d runes", n)
	}
	if body.Message.AllowedMentions.Users != nil || strings.Contains(*body.Message.Content, "<@") {
		t.Errorf("a reporter id that is not a snowflake must not be mentioned: %+v / %q", body.Message.AllowedMentions, *body.Message.Content)
	}
	if empty := (BugThread{IssueNumber: 1}).Render(); !strings.Contains(empty.Message.Embeds[0].Description, "no description") {
		t.Errorf("empty description = %q", empty.Message.Embeds[0].Description)
	}
}

func TestIssueNumberFromThreadName(t *testing.T) {
	for name, want := range map[string]int{
		"#812 Stack panel":    812,
		"#7":                  7,
		"Stack panel #812":    0,
		"#abc title":          0,
		"#1234567890 title":   0,
		"#12title":            0,
		"":                    0,
		"#0 zero is no issue": 0,
	} {
		if got := IssueNumberFromThreadName(name); got != want {
			t.Errorf("IssueNumberFromThreadName(%q) = %d, want %d", name, got, want)
		}
	}
}

func TestBugForumCreateThread(t *testing.T) {
	stub, f := newBugForumStub(t)
	stub.answer("POST", "/channels/"+testBugForum+"/threads", http.StatusCreated, `{"id":"`+testThreadID+`"}`)
	id, err := f.CreateThread(context.Background(), nil, sampleBugThread())
	if err != nil || id != testThreadID {
		t.Fatalf("CreateThread = %q, %v", id, err)
	}
	calls := stub.seen()
	if len(calls) != 1 || calls[0].Auth != "Bot "+testToken || calls[0].Reason != bugForumAuditReason {
		t.Fatalf("calls = %+v", calls)
	}
	if calls[0].Body["name"] != "#812 Stack panel clips the last item" {
		t.Errorf("name = %v", calls[0].Body["name"])
	}
}

func TestBugForumErrorsAreMapped(t *testing.T) {
	stub, f := newBugForumStub(t)
	path := "/channels/" + testBugForum + "/threads"
	stub.answer("POST", path, http.StatusForbidden, `{"code":50013}`)
	if _, err := f.CreateThread(context.Background(), nil, sampleBugThread()); err != ErrBugForumForbidden {
		t.Errorf("403 = %v, want ErrBugForumForbidden", err)
	}
	stub.answer("POST", path, http.StatusNotFound, `{"code":10003}`)
	if _, err := f.CreateThread(context.Background(), nil, sampleBugThread()); err == nil || !strings.Contains(err.Error(), ErrBugForumNotFound.Error()) {
		t.Errorf("404 = %v, want ErrBugForumNotFound", err)
	}
	stub.answer("POST", path, http.StatusTooManyRequests, `{"retry_after":1.5}`)
	if _, err := f.CreateThread(context.Background(), nil, sampleBugThread()); !Transient(err) {
		t.Errorf("429 = %v, want a transient error", err)
	}
	if _, err := (BugForum{}).CreateThread(context.Background(), nil, sampleBugThread()); err != ErrBugForumNotConfigured {
		t.Errorf("off = %v", err)
	}
	for _, call := range stub.seen() {
		if strings.Contains(call.Path, testToken) {
			t.Errorf("token leaked into a path: %s", call.Path)
		}
	}
}

func TestBugForumActiveThreadsKeepsOnlyTheBotsOwnPosts(t *testing.T) {
	stub, f := newBugForumStub(t)
	stub.answer("GET", "/channels/"+testBugForum, http.StatusOK, `{"guild_id":"`+testGuild+`","type":15}`)
	stub.answer("GET", "/users/@me", http.StatusOK, `{"id":"`+testBot+`"}`)
	stub.answer("GET", "/guilds/"+testGuild+"/threads/active", http.StatusOK, `{"threads":[
		{"id":"1","name":"#812 mine","parent_id":"`+testBugForum+`","owner_id":"`+testBot+`"},
		{"id":"2","name":"#813 by a person","parent_id":"`+testBugForum+`","owner_id":"999"},
		{"id":"3","name":"#814 in another forum","parent_id":"777","owner_id":"`+testBot+`"},
		{"id":"4","name":"renamed by a maintainer","parent_id":"`+testBugForum+`","owner_id":"`+testBot+`"}]}`)
	got, err := f.ActiveThreads(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "1" || got[0].Name != "#812 mine" {
		t.Fatalf("threads = %+v, want only the bot's own numbered post", got)
	}
}

func TestBugForumActiveThreadsReadsAListingLargerThanTheDefaultCap(t *testing.T) {
	stub, f := newBugForumStub(t)
	stub.answer("GET", "/channels/"+testBugForum, http.StatusOK, `{"guild_id":"`+testGuild+`","type":15}`)
	stub.answer("GET", "/users/@me", http.StatusOK, `{"id":"`+testBot+`"}`)
	var b strings.Builder
	b.WriteString(`{"threads":[`)
	for i := 0; i < 400; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{"id":"1","name":"someone else's thread with a name that takes room","parent_id":"777","owner_id":"1","topic":"` + strings.Repeat("x", 80) + `"}`)
	}
	b.WriteString(`,{"id":"9","name":"#9 last","parent_id":"` + testBugForum + `","owner_id":"` + testBot + `"}]}`)
	if b.Len() <= maxDMBody {
		t.Fatalf("fixture is only %d bytes", b.Len())
	}
	stub.answer("GET", "/guilds/"+testGuild+"/threads/active", http.StatusOK, b.String())
	got, err := f.ActiveThreads(context.Background(), nil)
	if err != nil || len(got) != 1 || got[0].ID != "9" {
		t.Fatalf("threads = %+v, %v", got, err)
	}
}

func TestBugForumActiveThreadsRefusesANonForum(t *testing.T) {
	stub, f := newBugForumStub(t)
	stub.answer("GET", "/channels/"+testBugForum, http.StatusOK, `{"guild_id":"`+testGuild+`","type":0}`)
	if _, err := f.ActiveThreads(context.Background(), nil); err != ErrBugForumNotAForum {
		t.Errorf("err = %v, want ErrBugForumNotAForum", err)
	}
}

func TestBugForumNoteAndArchive(t *testing.T) {
	stub, f := newBugForumStub(t)
	if err := f.PostNote(context.Background(), nil, testThreadID, "closed @everyone"); err != nil {
		t.Fatal(err)
	}
	if err := f.Archive(context.Background(), nil, testThreadID); err != nil {
		t.Fatal(err)
	}
	calls := stub.seen()
	if len(calls) != 2 || calls[0].Path != "/channels/"+testThreadID+"/messages" || calls[1].Method != "PATCH" || calls[1].Body["archived"] != true {
		t.Fatalf("calls = %+v", calls)
	}
	am, _ := calls[0].Body["allowed_mentions"].(map[string]any)
	if parse, ok := am["parse"].([]any); !ok || len(parse) != 0 {
		t.Errorf("a note must parse no mentions: %v", calls[0].Body["allowed_mentions"])
	}
	if err := f.Archive(context.Background(), nil, "../../guilds/1"); err == nil {
		t.Error("a thread id that is not a snowflake must be refused before any request")
	}
}
