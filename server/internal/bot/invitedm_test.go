package bot

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

func inviteDMData(user, game string) discordgo.ApplicationCommandInteractionData {
	opts := []*discordgo.ApplicationCommandInteractionDataOption{
		{Name: "user", Type: discordgo.ApplicationCommandOptionUser, Value: user},
	}
	if game != "" {
		opts = append(opts, &discordgo.ApplicationCommandInteractionDataOption{
			Name: "game", Type: discordgo.ApplicationCommandOptionString, Value: game,
		})
	}
	return discordgo.ApplicationCommandInteractionData{Name: CmdInviteDM, Options: opts}
}

func inviteDMInteraction(invoker string) *discordgo.InteractionCreate {
	return &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionApplicationCommand,
		Member:  &discordgo.Member{User: discordUser(invoker)},
	}}
}

func TestInviteDM_Client_SendsDiscordIDAndDecodes(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	res, err := c.InviteDM(context.Background(), id, "123456789012345678")
	if err != nil {
		t.Fatalf("InviteDM: %v", err)
	}
	if !res.Sent || res.DisplayName != "Bob" {
		t.Errorf("result: %+v", res)
	}
	if fs.gotDMDiscordID != "123456789012345678" || fs.gotDMGameID != id.String() {
		t.Errorf("server saw game %q discord_id %q", fs.gotDMGameID, fs.gotDMDiscordID)
	}
	if got := fs.gotAuthHeaders[len(fs.gotAuthHeaders)-1]; got != "Bearer session-token" {
		t.Errorf("must use the admin session, got %q", got)
	}
}

func TestInviteDM_Client_ErrorStatuses(t *testing.T) {
	for _, status := range []int{400, 403, 404, 409, 422, 424, 429, 503} {
		fs, c := newFakeServer(t)
		fs.dmStatus = status
		fs.dmBody = map[string]string{"error": "boom"}
		_, err := c.InviteDM(context.Background(), uuid.New(), "1")
		var de *InviteDMError
		if !errors.As(err, &de) || de.StatusCode != status || de.Message != "boom" {
			t.Errorf("status %d: got %v", status, err)
		}
	}
}

func TestInviteDM_Client_Unauthorized(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.requireBearer = "never-matches"
	_, err := c.InviteDM(context.Background(), uuid.New(), "1")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("want ErrUnauthorized, got %v", err)
	}
}

func TestInviteDMErrorMessage(t *testing.T) {
	cases := map[int]string{
		503: "not set up",
		422: "would not deliver",
		429: "Too many",
		409: "Rotate",
		404: "could not find",
	}
	for status, want := range cases {
		got := inviteDMErrorMessage(&InviteDMError{StatusCode: status})
		if !strings.Contains(got, want) {
			t.Errorf("%d: %q lacks %q", status, got, want)
		}
	}
	// The 503 message has to name the variable an operator can act on.
	if got := inviteDMErrorMessage(&InviteDMError{StatusCode: 503}); !strings.Contains(got, "CMDCTRL_DISCORD_BOT_TOKEN") {
		t.Errorf("503 message should name the variable: %q", got)
	}
	if got := inviteDMErrorMessage(ErrServerUnreachable); !strings.Contains(got, "not reachable") {
		t.Errorf("transport error: %q", got)
	}
}

func TestInviteDMDefinition(t *testing.T) {
	var def *discordgo.ApplicationCommand
	for _, d := range commandDefinitions() {
		if d.Name == CmdInviteDM {
			def = d
		}
	}
	if def == nil {
		t.Fatal("missing /c2-invite-dm")
	}
	if len(def.Options) != 3 {
		t.Fatalf("options: %+v", def.Options)
	}
	u, g, n := def.Options[0], def.Options[1], def.Options[2]
	if u.Name != "user" || u.Type != discordgo.ApplicationCommandOptionUser || !u.Required {
		t.Errorf("user option: %+v", u)
	}
	if g.Name != "game" || g.Required || !g.Autocomplete {
		t.Errorf("game option: %+v", g)
	}
	if n.Name != "name" || n.Required {
		t.Errorf("name option: %+v", n)
	}
}

func TestCommandTimeout_InviteDMIsDeferredBudget(t *testing.T) {
	if commandTimeout(CmdInviteDM) != deckInteractionTimeout {
		t.Error("/c2-invite-dm defers and needs the long budget")
	}
}

func TestUserOptionID(t *testing.T) {
	d := inviteDMData("42", "")
	if got := userOptionID(d.Options, "user"); got != "42" {
		t.Errorf("got %q", got)
	}
	if got := userOptionID(d.Options, "missing"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestRunInviteDM_NewGame_CreatesThenDMs(t *testing.T) {
	fs, c := newFakeServer(t)
	gid := uuid.New()
	fs.createMeta = lobby.GameMeta{ID: gid, Name: "friday", InviteToken: "tok", State: "lobby"}
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())

	msg := h.runInviteDM(context.Background(), inviteDMInteraction("host"), inviteDMData("222", ""))

	if fs.dmCalls != 1 || fs.gotDMGameID != gid.String() || fs.gotDMDiscordID != "222" {
		t.Errorf("dm call: calls=%d game=%q discord=%q", fs.dmCalls, fs.gotDMGameID, fs.gotDMDiscordID)
	}
	if !strings.Contains(msg, "Created") || !strings.Contains(msg, "Bob") {
		t.Errorf("message: %q", msg)
	}
	if strings.Contains(msg, "tok") {
		t.Errorf("success must not echo the invite token: %q", msg)
	}
}

func TestRunInviteDM_NotConfigured_503IsClearAndKeepsTheLink(t *testing.T) {
	fs, c := newFakeServer(t)
	gid := uuid.New()
	fs.createMeta = lobby.GameMeta{ID: gid, Name: "friday", InviteToken: "tok", State: "lobby"}
	fs.dmStatus = http.StatusServiceUnavailable
	fs.dmBody = map[string]string{"error": "CMDCTRL_DISCORD_BOT_TOKEN is not set"}
	h := NewHandler(Config{GuildIDs: []string{"g1"}, ClientBaseURL: "https://cmd.example"}, c, discardLogger())

	msg := h.runInviteDM(context.Background(), inviteDMInteraction("host"), inviteDMData("222", ""))

	if !strings.Contains(msg, "not set up") || !strings.Contains(msg, "CMDCTRL_DISCORD_BOT_TOKEN") {
		t.Errorf("503 should say DMs are not configured: %q", msg)
	}
	if !strings.Contains(msg, "https://cmd.example/#/games/"+gid.String()+"/join?t=tok") {
		t.Errorf("the created game's link should be handed back: %q", msg)
	}
}

func TestRunInviteDM_BotTargetRefused(t *testing.T) {
	fs, c := newFakeServer(t)
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())
	data := inviteDMData("999", "")
	data.Resolved = &discordgo.ApplicationCommandInteractionDataResolved{
		Users: map[string]*discordgo.User{"999": {ID: "999", Bot: true}},
	}
	msg := h.runInviteDM(context.Background(), inviteDMInteraction("host"), data)
	if !strings.Contains(msg, "bot") || fs.dmCalls != 0 || fs.createName != "" {
		t.Errorf("a bot target must be refused before any server call: %q calls=%d", msg, fs.dmCalls)
	}
}

func TestRunInviteDM_ExistingGame_NonCreatorRefused(t *testing.T) {
	fs, c := newFakeServer(t)
	gid := uuid.New()
	fs.listMeta = []lobby.GameMeta{{ID: gid, Name: "friday-commander", State: "lobby"}}
	fs.creatorIsCreator = false
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())

	msg := h.runInviteDM(context.Background(), inviteDMInteraction("rando"), inviteDMData("222", "friday"))

	if fs.dmCalls != 0 {
		t.Error("a non-creator, non-admin must never reach the DM route")
	}
	if !strings.Contains(msg, "creator or a bot admin") {
		t.Errorf("message: %q", msg)
	}
}

func TestRunInviteDM_ExistingGame_AdminAllowed(t *testing.T) {
	fs, c := newFakeServer(t)
	gid := uuid.New()
	fs.listMeta = []lobby.GameMeta{{ID: gid, Name: "friday-commander", State: "lobby"}}
	h := NewHandler(Config{GuildIDs: []string{"g1"}, AdminUserIDs: []string{"admin1"}}, c, discardLogger())

	msg := h.runInviteDM(context.Background(), inviteDMInteraction("admin1"), inviteDMData("222", "friday"))

	if fs.dmCalls != 1 || fs.gotDMGameID != gid.String() {
		t.Errorf("dm call: calls=%d game=%q", fs.dmCalls, fs.gotDMGameID)
	}
	if fs.createName != "" {
		t.Error("naming an existing game must not create another")
	}
	if !strings.Contains(msg, "Sent Bob") {
		t.Errorf("message: %q", msg)
	}
}

func TestRunInviteDM_ExistingGame_CreatorAllowed(t *testing.T) {
	fs, c := newFakeServer(t)
	gid := uuid.New()
	fs.listMeta = []lobby.GameMeta{{ID: gid, Name: "friday-commander", State: "lobby"}}
	fs.creatorIsCreator = true
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())

	_ = h.runInviteDM(context.Background(), inviteDMInteraction("creator"), inviteDMData("222", "friday"))

	if fs.dmCalls != 1 {
		t.Errorf("a creator may DM their table's invite, calls=%d", fs.dmCalls)
	}
}

func TestRunInviteDM_UnknownGame(t *testing.T) {
	fs, c := newFakeServer(t)
	h := NewHandler(Config{GuildIDs: []string{"g1"}, AdminUserIDs: []string{"admin1"}}, c, discardLogger())
	msg := h.runInviteDM(context.Background(), inviteDMInteraction("admin1"), inviteDMData("222", "nope"))
	if fs.dmCalls != 0 || !strings.Contains(msg, "No game matches") {
		t.Errorf("message: %q calls=%d", msg, fs.dmCalls)
	}
}

func TestDispatch_InviteDM_RejectsNonAllowedGuild(t *testing.T) {
	fs, c := newFakeServer(t)
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())
	i := inviteDMInteraction("host")
	i.GuildID = "elsewhere"
	i.Data = inviteDMData("222", "")
	dispatchRecover(t, h, i)
	if fs.dmCalls != 0 || fs.createName != "" {
		t.Error("a non-allowed guild must not reach any server call")
	}
}
