package bot

// Tests for access.go (ADR 0156's 2026-10-16 amendment, #3325): the
// access request a member's first message in #account-requests gets, and
// the Approve and Decline buttons only a mod can press.

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	testAccessGuild   = "100000000000000001"
	testAccessChannel = "100000000000000002"
	testPlayerRole    = "100000000000000003"
	testModLog        = "100000000000000004"
	testMember        = "100000000000000005"
	testMod           = "100000000000000006"
)

func accessConfig() Config {
	return Config{
		GuildIDs:               []string{testAccessGuild},
		AccessRequestChannelID: testAccessChannel,
		PlayerRoleID:           testPlayerRole,
		ModLogChannelID:        testModLog,
	}
}

// accessSender records the messages HandleAccessRequestMessage sends.
type accessSender struct {
	mu   sync.Mutex
	sent []struct {
		channel string
		msg     *discordgo.MessageSend
	}
	err error
}

func (f *accessSender) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return nil, f.err
	}
	f.sent = append(f.sent, struct {
		channel string
		msg     *discordgo.MessageSend
	}{channelID, data})
	return &discordgo.Message{ID: "1"}, nil
}

func (f *accessSender) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

func accessMessage(authorID string, roles ...string) *discordgo.MessageCreate {
	return &discordgo.MessageCreate{Message: &discordgo.Message{
		ID: "m1", ChannelID: testAccessChannel, GuildID: testAccessGuild,
		Type:   discordgo.MessageTypeDefault,
		Author: &discordgo.User{ID: authorID},
		Member: &discordgo.Member{Roles: roles},
	}}
}

func TestAccessRequestPostedOncePerMemberWithButtons(t *testing.T) {
	h := NewHandler(accessConfig(), nil, discardLogger())
	now := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	h.now = func() time.Time { return now }
	f := &accessSender{}

	h.HandleAccessRequestMessage(f, accessMessage(testMember))
	if f.count() != 1 {
		t.Fatalf("sent %d, want one request", f.count())
	}
	sent := f.sent[0]
	if sent.channel != testAccessChannel || sent.msg.Reference == nil || sent.msg.Reference.MessageID != "m1" {
		t.Errorf("request is not a reply in the channel: %+v", sent)
	}
	if am := sent.msg.AllowedMentions; am == nil || len(am.Parse) != 0 || len(am.Users) != 0 || am.RepliedUser {
		t.Errorf("allowed_mentions = %+v, want nobody pinged", am)
	}
	if !strings.Contains(sent.msg.Embeds[0].Description, "<@"+testMember+">") {
		t.Errorf("description = %q", sent.msg.Embeds[0].Description)
	}
	row := sent.msg.Components[0].(discordgo.ActionsRow)
	approve, decline := row.Components[0].(discordgo.Button), row.Components[1].(discordgo.Button)
	if approve.Label != "Approve" || approve.CustomID != "c2-access:approve:"+testMember ||
		decline.Label != "Decline" || decline.CustomID != "c2-access:decline:"+testMember {
		t.Errorf("buttons = %+v / %+v", approve, decline)
	}
	if len(approve.CustomID) > 100 {
		t.Errorf("custom id over Discord's 100-character cap: %d", len(approve.CustomID))
	}

	// The same member keeps chatting: nothing more that day.
	h.HandleAccessRequestMessage(f, accessMessage(testMember))
	if f.count() != 1 {
		t.Errorf("a second message posted another request")
	}
	now = now.Add(accessRequestCooldown)
	h.HandleAccessRequestMessage(f, accessMessage(testMember))
	if f.count() != 2 {
		t.Errorf("a message a day later posted nothing")
	}
}

func TestAccessRequestSkipsWhoShouldNotGetOne(t *testing.T) {
	cfg := accessConfig()
	cfg.AdminRoleIDs = []string{"900"}
	other := accessMessage(testMember)
	other.ChannelID = "999"
	bot := accessMessage(testMember)
	bot.Author.Bot = true
	join := accessMessage(testMember)
	join.Type = discordgo.MessageTypeGuildMemberJoin
	hook := accessMessage(testMember)
	hook.WebhookID = "1"
	wrongGuild := accessMessage(testMember)
	wrongGuild.GuildID = "998"

	for name, m := range map[string]*discordgo.MessageCreate{
		"a member who has the role": accessMessage(testMember, testPlayerRole),
		"a mod on the admin role":   accessMessage(testMember, "900"),
		"another channel":           other,
		"a bot":                     bot,
		"a join notice":             join,
		"a webhook":                 hook,
		"a guild not allowed":       wrongGuild,
	} {
		h := NewHandler(cfg, nil, discardLogger())
		f := &accessSender{}
		h.HandleAccessRequestMessage(f, m)
		if f.count() != 0 {
			t.Errorf("%s: a request was posted", name)
		}
	}

	off := accessConfig()
	off.AccessRequestChannelID = ""
	h := NewHandler(off, nil, discardLogger())
	f := &accessSender{}
	h.HandleAccessRequestMessage(f, accessMessage(testMember))
	if f.count() != 0 {
		t.Error("posted with the feature off")
	}
}

func TestAccessRequestFailedPostIsTriedAgain(t *testing.T) {
	h := NewHandler(accessConfig(), nil, discardLogger())
	f := &accessSender{err: io.ErrUnexpectedEOF}
	h.HandleAccessRequestMessage(f, accessMessage(testMember))
	f.err = nil
	h.HandleAccessRequestMessage(f, accessMessage(testMember))
	if f.count() != 1 {
		t.Errorf("sent %d after a failure and a retry, want 1", f.count())
	}
}

func TestParseAccessCustomID(t *testing.T) {
	for id, want := range map[string]bool{
		"c2-access:approve:" + testMember: true,
		"c2-access:decline:" + testMember: true,
		"c2-access:ban:" + testMember:     false,
		"c2-access:approve:":              false,
		"c2-access:approve:abc":           false,
		"c2-access:approve":               false,
		"c2-end:confirm:x":                false,
	} {
		if _, _, ok := parseAccessCustomID(id); ok != want {
			t.Errorf("%q: ok %v, want %v", id, ok, want)
		}
	}
}

// --- the buttons, through Dispatch against a scripted Discord ---

// scriptedDiscord stands in for Discord's REST API: every call is
// recorded, with its audit-log reason, and answered from a table keyed
// by "METHOD path-suffix", else 200 "{}".
type scriptedDiscord struct {
	mu      sync.Mutex
	calls   []scriptedCall
	answers map[string]scriptedAnswer
}

type scriptedCall struct {
	Method, Path, Reason string
	Body                 map[string]any
}

type scriptedAnswer struct {
	status int
	body   string
}

func (d *scriptedDiscord) RoundTrip(req *http.Request) (*http.Response, error) {
	var body map[string]any
	if req.Body != nil {
		raw, _ := io.ReadAll(req.Body)
		_ = json.Unmarshal(raw, &body)
	}
	d.mu.Lock()
	d.calls = append(d.calls, scriptedCall{Method: req.Method, Path: req.URL.Path, Reason: req.Header.Get("X-Audit-Log-Reason"), Body: body})
	a := scriptedAnswer{status: http.StatusOK, body: "{}"}
	for k, v := range d.answers {
		method, suffix, _ := strings.Cut(k, " ")
		if req.Method == method && strings.HasSuffix(req.URL.Path, suffix) {
			a = v
		}
	}
	d.mu.Unlock()
	return &http.Response{
		StatusCode: a.status,
		Status:     http.StatusText(a.status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(a.body)),
		Request:    req,
	}, nil
}

func (d *scriptedDiscord) snapshot() []scriptedCall {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]scriptedCall(nil), d.calls...)
}

func (d *scriptedDiscord) find(method, suffix string) (scriptedCall, bool) {
	for _, c := range d.snapshot() {
		if c.Method == method && strings.HasSuffix(c.Path, suffix) {
			return c, true
		}
	}
	return scriptedCall{}, false
}

func scriptedSession(t *testing.T, answers map[string]scriptedAnswer) (*discordgo.Session, *scriptedDiscord) {
	t.Helper()
	s, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	s.MaxRestRetries = 0
	d := &scriptedDiscord{answers: answers}
	s.Client = &http.Client{Transport: d, Timeout: 2 * time.Second}
	return s, d
}

func accessPress(action string, presser *discordgo.Member) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		ID: "i9", AppID: "app", Token: "tok", GuildID: testAccessGuild, ChannelID: testAccessChannel,
		Type:   discordgo.InteractionMessageComponent,
		Member: presser,
		Data:   discordgo.MessageComponentInteractionData{CustomID: accessCustomID(action, testMember)},
	}}
}

func modMember(perms int64) *discordgo.Member {
	return &discordgo.Member{User: &discordgo.User{ID: testMod, GlobalName: "Mod Mia"}, Permissions: perms}
}

const rolePath = "/guilds/" + testAccessGuild + "/members/" + testMember + "/roles/" + testPlayerRole

func callbackOf(t *testing.T, d *scriptedDiscord) map[string]any {
	t.Helper()
	c, ok := d.find(http.MethodPost, "/callback")
	if !ok {
		t.Fatalf("no interaction callback in %+v", d.snapshot())
	}
	return c.Body
}

func TestAccessApproveGrantsTheRoleAndLogsIt(t *testing.T) {
	h := NewHandler(accessConfig(), nil, discardLogger())
	sess, d := scriptedSession(t, map[string]scriptedAnswer{"PUT " + rolePath: {status: http.StatusNoContent}})

	h.Dispatch(sess, accessPress(accessApprove, modMember(discordgo.PermissionManageRoles)))

	grant, ok := d.find(http.MethodPut, rolePath)
	if !ok {
		t.Fatalf("no role grant in %+v", d.snapshot())
	}
	if !strings.Contains(grant.Reason, "approved by Mod Mia") {
		t.Errorf("audit reason = %q", grant.Reason)
	}
	cb := callbackOf(t, d)
	if cb["type"] != float64(discordgo.InteractionResponseUpdateMessage) {
		t.Fatalf("callback = %+v", cb)
	}
	data := cb["data"].(map[string]any)
	if comps, _ := data["components"].([]any); len(comps) != 0 {
		t.Errorf("buttons left on an answered request: %+v", data["components"])
	}
	desc := data["embeds"].([]any)[0].(map[string]any)["description"].(string)
	if !strings.Contains(desc, "Approved by <@"+testMod+">") {
		t.Errorf("description = %q", desc)
	}
	if am, _ := data["allowed_mentions"].(map[string]any); am == nil {
		t.Errorf("an answered request may ping: %+v", data)
	}

	log, ok := d.find(http.MethodPost, "/channels/"+testModLog+"/messages")
	if !ok {
		t.Fatalf("nothing in #mod-log: %+v", d.snapshot())
	}
	embed := log.Body["embeds"].([]any)[0].(map[string]any)
	if embed["title"] != "Access request approved" {
		t.Errorf("mod log = %+v", embed)
	}
}

func TestAccessDeclineChangesNoRole(t *testing.T) {
	h := NewHandler(accessConfig(), nil, discardLogger())
	sess, d := scriptedSession(t, nil)

	h.Dispatch(sess, accessPress(accessDecline, modMember(discordgo.PermissionAdministrator)))

	if _, granted := d.find(http.MethodPut, rolePath); granted {
		t.Error("Decline granted the role")
	}
	data := callbackOf(t, d)["data"].(map[string]any)
	desc := data["embeds"].([]any)[0].(map[string]any)["description"].(string)
	if !strings.Contains(desc, "Declined by <@"+testMod+">") {
		t.Errorf("description = %q", desc)
	}
	if _, ok := d.find(http.MethodPost, "/channels/"+testModLog+"/messages"); !ok {
		t.Error("a decline is not in #mod-log")
	}
}

func TestAccessButtonsAreForModsOnly(t *testing.T) {
	cfg := accessConfig()
	cfg.AdminUserIDs = []string{"424242"}
	h := NewHandler(cfg, nil, discardLogger())

	for name, presser := range map[string]*discordgo.Member{
		"the member themselves":               {User: &discordgo.User{ID: testMember}},
		"a member who can only send messages": modMember(discordgo.PermissionSendMessages),
	} {
		sess, d := scriptedSession(t, nil)
		h.Dispatch(sess, accessPress(accessApprove, presser))
		if _, granted := d.find(http.MethodPut, rolePath); granted {
			t.Errorf("%s: the role was granted", name)
		}
		cb := callbackOf(t, d)
		data := cb["data"].(map[string]any)
		if data["flags"] != float64(discordgo.MessageFlagsEphemeral) || !strings.Contains(data["content"].(string), "Only a mod") {
			t.Errorf("%s: callback = %+v", name, cb)
		}
	}

	// Someone on the bot's admin list may, with no Discord permission.
	sess, d := scriptedSession(t, nil)
	h.Dispatch(sess, accessPress(accessApprove, &discordgo.Member{User: &discordgo.User{ID: "424242"}}))
	if _, granted := d.find(http.MethodPut, rolePath); !granted {
		t.Error("a configured admin could not approve")
	}
}

func TestAccessApproveWhenDiscordRefuses(t *testing.T) {
	h := NewHandler(accessConfig(), nil, discardLogger())

	sess, d := scriptedSession(t, map[string]scriptedAnswer{
		"PUT " + rolePath: {status: http.StatusForbidden, body: `{"message":"Missing Permissions","code":50013}`},
	})
	h.Dispatch(sess, accessPress(accessApprove, modMember(discordgo.PermissionManageRoles)))
	data := callbackOf(t, d)["data"].(map[string]any)
	if data["content"] != approveForbiddenMessage || data["flags"] != float64(discordgo.MessageFlagsEphemeral) {
		t.Errorf("forbidden: %+v", data)
	}
	if _, logged := d.find(http.MethodPost, "/channels/"+testModLog+"/messages"); logged {
		t.Error("a refused grant was logged as approved")
	}

	sess, d = scriptedSession(t, map[string]scriptedAnswer{
		"PUT " + rolePath: {status: http.StatusNotFound, body: `{"message":"Unknown Member","code":10007}`},
	})
	h.Dispatch(sess, accessPress(accessApprove, modMember(discordgo.PermissionManageRoles)))
	cb := callbackOf(t, d)
	desc := cb["data"].(map[string]any)["embeds"].([]any)[0].(map[string]any)["description"].(string)
	if cb["type"] != float64(discordgo.InteractionResponseUpdateMessage) || !strings.Contains(desc, "left the server") {
		t.Errorf("unknown member: %+v", cb)
	}
}

func TestAccessConfigValidate(t *testing.T) {
	base := Config{BotToken: "t", AppID: "a", GuildIDs: []string{"1"}, AdminToken: "x", ServerBaseURL: "s", ClientBaseURL: "c"}
	if err := base.Validate(); err != nil {
		t.Fatalf("base: %v", err)
	}
	c := base
	c.AccessRequestChannelID = testAccessChannel
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), PlayerRoleEnv) {
		t.Errorf("channel without the role: %v", err)
	}
	c.PlayerRoleID = testPlayerRole
	if err := c.Validate(); err != nil || !c.AccessRequestsEnabled() {
		t.Errorf("channel and role: %v, enabled %v", err, c.AccessRequestsEnabled())
	}
	c.ModLogChannelID = "#mod-log"
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), ModLogChannelEnv) {
		t.Errorf("malformed mod log: %v", err)
	}
	c = base
	c.PlayerRoleID = testPlayerRole
	if err := c.Validate(); err != nil || c.AccessRequestsEnabled() {
		t.Errorf("the role alone is valid and off: %v, enabled %v", err, c.AccessRequestsEnabled())
	}
}
