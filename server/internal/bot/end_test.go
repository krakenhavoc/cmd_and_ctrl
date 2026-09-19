package bot

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

func discordUser(id string) *discordgo.User { return &discordgo.User{ID: id} }

func guildInteraction(userID string, roles []string) *discordgo.Interaction {
	return &discordgo.Interaction{
		Member: &discordgo.Member{User: discordUser(userID), Roles: roles},
	}
}

func dmInteraction(userID string) *discordgo.Interaction {
	return &discordgo.Interaction{User: discordUser(userID)}
}

// --- Config.IsAdmin ---

func TestIsAdmin_AllowlistedUser(t *testing.T) {
	cfg := Config{AdminUserIDs: []string{"u1", "u2"}}
	if !cfg.IsAdmin(guildInteraction("u1", nil)) {
		t.Error("u1 should be admin")
	}
	if cfg.IsAdmin(guildInteraction("u3", nil)) {
		t.Error("u3 should NOT be admin")
	}
}

func TestIsAdmin_RoleHolder(t *testing.T) {
	cfg := Config{AdminRoleIDs: []string{"r1", "r2"}}
	if !cfg.IsAdmin(guildInteraction("someone", []string{"other", "r2"})) {
		t.Error("member with r2 should be admin")
	}
	if cfg.IsAdmin(guildInteraction("someone", []string{"other"})) {
		t.Error("member without an admin role should NOT be admin")
	}
}

func TestIsAdmin_Neither(t *testing.T) {
	cfg := Config{AdminUserIDs: []string{"u1"}, AdminRoleIDs: []string{"r1"}}
	if cfg.IsAdmin(guildInteraction("u2", []string{"r2"})) {
		t.Error("caller on neither list should NOT be admin")
	}
}

func TestIsAdmin_EmptyConfig(t *testing.T) {
	cfg := Config{}
	if cfg.IsAdmin(guildInteraction("anyone", []string{"anything"})) {
		t.Error("empty admin config must refuse everyone, never fail open")
	}
	if cfg.HasAdmins() {
		t.Error("HasAdmins() should be false for an empty config")
	}
}

func TestIsAdmin_DMInvoker(t *testing.T) {
	cfg := Config{AdminUserIDs: []string{"u1"}}
	if !cfg.IsAdmin(dmInteraction("u1")) {
		t.Error("DM invoker on the allow-list should be admin")
	}
}

func TestAdminRefusalMessage_NamesTheVars(t *testing.T) {
	msg := adminRefusalMessage(Config{})
	if !strings.Contains(msg, "CMDCTRL_DISCORD_ADMIN_USER_IDS") || !strings.Contains(msg, "CMDCTRL_DISCORD_ADMIN_ROLE_IDS") {
		t.Errorf("empty-config refusal must name both env vars: %q", msg)
	}
	msg2 := adminRefusalMessage(Config{AdminUserIDs: []string{"u1"}})
	if strings.Contains(msg2, "No admins are configured") {
		t.Errorf("a configured-but-not-you refusal should differ from the empty-config one: %q", msg2)
	}
}

// --- matchGameByName / resolveGame ---

func TestMatchGameByName(t *testing.T) {
	games := []lobby.GameMeta{
		{ID: uuid.New(), Name: "Friday Night Magic"},
		{ID: uuid.New(), Name: "friday-commander"},
		{ID: uuid.New(), Name: "Saturday Draft"},
	}
	if _, err := matchGameByName(games, "friday"); !errors.Is(err, errGameAmbiguous) {
		t.Errorf("want errGameAmbiguous for a two-way prefix match, got %v", err)
	}
	got, err := matchGameByName(games, "Saturday")
	if err != nil {
		t.Fatalf("Saturday: %v", err)
	}
	if got.Name != "Saturday Draft" {
		t.Errorf("got %q", got.Name)
	}
	if _, err := matchGameByName(games, "nope"); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("want ErrGameNotFound, got %v", err)
	}
}

func TestResolveGame_ByID(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	fs.getMeta = lobby.GameMeta{ID: id, Name: "friday"}
	h := NewHandler(Config{}, c, discardLogger())

	meta, err := h.resolveGame(context.Background(), id.String())
	if err != nil {
		t.Fatalf("resolveGame: %v", err)
	}
	if meta.Name != "friday" {
		t.Errorf("got %+v", meta)
	}
}

func TestResolveGame_ByID_Unknown(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.getStatus = http.StatusNotFound
	h := NewHandler(Config{}, c, discardLogger())

	_, err := h.resolveGame(context.Background(), uuid.New().String())
	if !errors.Is(err, ErrGameNotFound) {
		t.Errorf("want ErrGameNotFound, got %v", err)
	}
}

func TestResolveGame_ByNamePrefix(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	fs.listMeta = []lobby.GameMeta{{ID: id, Name: "Friday Commander"}}
	h := NewHandler(Config{}, c, discardLogger())

	meta, err := h.resolveGame(context.Background(), "friday")
	if err != nil {
		t.Fatalf("resolveGame: %v", err)
	}
	if meta.ID != id {
		t.Errorf("got %+v", meta)
	}
}

func TestResolveGame_UnknownName(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.listMeta = nil
	h := NewHandler(Config{}, c, discardLogger())

	_, err := h.resolveGame(context.Background(), "no-such-table")
	if !errors.Is(err, ErrGameNotFound) {
		t.Errorf("want ErrGameNotFound, got %v", err)
	}
}

// --- gameChoices (autocomplete) ---

func TestGameChoices(t *testing.T) {
	games := []lobby.GameMeta{
		{ID: uuid.New(), Name: "Friday Night Magic", State: "lobby", Players: make([]lobby.SeatInfo, 2)},
		{ID: uuid.New(), Name: "Saturday Draft", State: "active", Players: make([]lobby.SeatInfo, 1)},
	}
	choices := gameChoices(games, "friday")
	if len(choices) != 1 {
		t.Fatalf("want 1 choice, got %d: %+v", len(choices), choices)
	}
	if choices[0].Value != games[0].ID.String() {
		t.Errorf("choice value should be the game id, got %v", choices[0].Value)
	}
	if !strings.Contains(choices[0].Name, "Friday Night Magic") || !strings.Contains(choices[0].Name, "lobby") {
		t.Errorf("choice label: got %q", choices[0].Name)
	}

	all := gameChoices(games, "")
	if len(all) != 2 {
		t.Errorf("empty query should return every game, got %d", len(all))
	}
}

func TestGameChoices_CapsAt25(t *testing.T) {
	games := make([]lobby.GameMeta, 30)
	for i := range games {
		games[i] = lobby.GameMeta{ID: uuid.New(), Name: "game"}
	}
	choices := gameChoices(games, "")
	if len(choices) != 25 {
		t.Errorf("want 25 choices (Discord's cap), got %d", len(choices))
	}
}

// --- evaluateComponentClick: confirm / cancel / timeout / wrong user ---

func newTestHandler(client *ServerClient, cfg Config, now time.Time) *Handler {
	h := NewHandler(cfg, client, discardLogger())
	h.now = func() time.Time { return now }
	return h
}

func TestEvaluateComponentClick_Confirm(t *testing.T) {
	now := time.Now()
	h := newTestHandler(nil, Config{}, now)
	gid := uuid.New()
	token := "tok1"
	h.confirmations.put(token, pendingEnd{GameID: gid, GameName: "friday", InvokerID: "u1", ExpiresAt: now.Add(confirmTTL)}, now)

	outcome, pending, gotToken := h.evaluateComponentClick(ccEndCustomIDPrefix+"confirm:"+token, "u1", now)
	if outcome != outcomeConfirm {
		t.Fatalf("want outcomeConfirm, got %v", outcome)
	}
	if pending.GameID != gid || gotToken != token {
		t.Errorf("pending/token mismatch: %+v %q", pending, gotToken)
	}
}

func TestEvaluateComponentClick_Cancel(t *testing.T) {
	now := time.Now()
	h := newTestHandler(nil, Config{}, now)
	token := "tok2"
	h.confirmations.put(token, pendingEnd{GameID: uuid.New(), GameName: "friday", InvokerID: "u1", ExpiresAt: now.Add(confirmTTL)}, now)

	outcome, _, _ := h.evaluateComponentClick(ccEndCustomIDPrefix+"cancel:"+token, "u1", now)
	if outcome != outcomeCancel {
		t.Fatalf("want outcomeCancel, got %v", outcome)
	}
}

func TestEvaluateComponentClick_Timeout(t *testing.T) {
	now := time.Now()
	h := newTestHandler(nil, Config{}, now)
	token := "tok3"
	h.confirmations.put(token, pendingEnd{GameID: uuid.New(), GameName: "friday", InvokerID: "u1", ExpiresAt: now.Add(confirmTTL)}, now)

	later := now.Add(confirmTTL + time.Second)
	outcome, pending, _ := h.evaluateComponentClick(ccEndCustomIDPrefix+"confirm:"+token, "u1", later)
	if outcome != outcomeExpired {
		t.Fatalf("want outcomeExpired, got %v", outcome)
	}
	if pending.GameName != "friday" {
		t.Errorf("expired outcome should still carry the game name for the message: %+v", pending)
	}
}

func TestEvaluateComponentClick_WrongUser(t *testing.T) {
	now := time.Now()
	h := newTestHandler(nil, Config{}, now)
	token := "tok4"
	h.confirmations.put(token, pendingEnd{GameID: uuid.New(), GameName: "friday", InvokerID: "u1", ExpiresAt: now.Add(confirmTTL)}, now)

	outcome, _, _ := h.evaluateComponentClick(ccEndCustomIDPrefix+"confirm:"+token, "u2", now)
	if outcome != outcomeWrongUser {
		t.Fatalf("want outcomeWrongUser, got %v", outcome)
	}
	// The wrong user's click must not consume the token — the real
	// invoker can still confirm afterwards.
	if _, found := h.confirmations.get(token); !found {
		t.Error("wrong-user click must not delete the pending confirmation")
	}
}

func TestEvaluateComponentClick_NotFound(t *testing.T) {
	h := newTestHandler(nil, Config{}, time.Now())
	outcome, _, _ := h.evaluateComponentClick(ccEndCustomIDPrefix+"confirm:missing", "u1", time.Now())
	if outcome != outcomeNotFound {
		t.Fatalf("want outcomeNotFound, got %v", outcome)
	}
}

func TestEvaluateComponentClick_MalformedCustomID(t *testing.T) {
	h := newTestHandler(nil, Config{}, time.Now())
	outcome, _, _ := h.evaluateComponentClick(ccEndCustomIDPrefix+"noaction", "u1", time.Now())
	if outcome != outcomeInvalidCustomID {
		t.Fatalf("want outcomeInvalidCustomID, got %v", outcome)
	}
}

// --- endConfirmations sweep ---

func TestEndConfirmations_PutSweepsExpired(t *testing.T) {
	e := newEndConfirmations()
	base := time.Now()
	e.put("old", pendingEnd{ExpiresAt: base.Add(1 * time.Second)}, base)

	// A later put, past "old"'s expiry, should sweep it out even
	// though nobody ever clicked its button.
	later := base.Add(2 * time.Second)
	e.put("new", pendingEnd{ExpiresAt: later.Add(confirmTTL)}, later)

	if _, found := e.get("old"); found {
		t.Error("expired entry should have been swept on the next put")
	}
	if _, found := e.get("new"); !found {
		t.Error("new entry should still be present")
	}
}

// --- endConfirmResponse / message builders ---

func TestEndConfirmResponse_NamesTheTable(t *testing.T) {
	meta := lobby.GameMeta{
		ID:        uuid.New(),
		Name:      "Friday Commander",
		CreatedAt: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC),
		Players:   make([]lobby.SeatInfo, 3),
	}
	resp := endConfirmResponse(meta, "tok")
	if resp.Data.Flags&discordgo.MessageFlagsEphemeral == 0 {
		t.Error("confirmation should be ephemeral")
	}
	if !strings.Contains(resp.Data.Content, "Friday Commander") || !strings.Contains(resp.Data.Content, "3 players") {
		t.Errorf("content missing name/seat count: %q", resp.Data.Content)
	}
	if !strings.Contains(resp.Data.Content, "2026-09-19") {
		t.Errorf("content missing created time: %q", resp.Data.Content)
	}
	if len(resp.Data.Components) != 1 {
		t.Fatalf("want one action row, got %d", len(resp.Data.Components))
	}
	row, ok := resp.Data.Components[0].(discordgo.ActionsRow)
	if !ok || len(row.Components) != 2 {
		t.Fatalf("want a row with Confirm+Cancel buttons, got %+v", resp.Data.Components[0])
	}
}

func TestResolveGameErrorMessage(t *testing.T) {
	cases := map[error]string{
		ErrGameNotFound:      "No game matches",
		errGameAmbiguous:     "More than one",
		ErrServerUnreachable: "not reachable",
		ErrUnauthorized:      "not authorized",
	}
	for err, want := range cases {
		if got := resolveGameErrorMessage(err); !strings.Contains(got, want) {
			t.Errorf("resolveGameErrorMessage(%v): got %q, want substring %q", err, got, want)
		}
	}
}

func TestArchiveErrorMessage(t *testing.T) {
	cases := map[error]string{
		ErrGameNotFound:      "gone",
		ErrServerUnreachable: "not reachable",
		ErrUnauthorized:      "not authorized",
	}
	for err, want := range cases {
		if got := archiveErrorMessage(err); !strings.Contains(got, want) {
			t.Errorf("archiveErrorMessage(%v): got %q, want substring %q", err, got, want)
		}
	}
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

// --- Dispatch-level wiring: handleEnd / dispatchComponent end to
// end, driven with a nil Session the way TestDispatch_UnknownCommand
// already does elsewhere in this package. The Session call at the
// very end of each path panics on the nil pointer; recover() lets
// the test observe everything that happened before that point
// (authorization decision, confirmation stored, archive HTTP call
// made) without needing a fake discordgo.Session.

func dispatchRecover(t *testing.T, h *Handler, i *discordgo.InteractionCreate) {
	t.Helper()
	defer func() { _ = recover() }()
	h.Dispatch(nil, i)
}

func TestHandleEnd_Unauthorized_NoConfirmationCreated(t *testing.T) {
	_, c := newFakeServer(t)
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger()) // no admins configured

	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionApplicationCommand,
		Member:  &discordgo.Member{User: discordUser("u1")},
		Data: discordgo.ApplicationCommandInteractionData{
			Name: CmdEnd,
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "game", Type: discordgo.ApplicationCommandOptionString, Value: "friday"},
			},
		},
	}}
	dispatchRecover(t, h, i)

	if len(h.confirmations.pending) != 0 {
		t.Error("an unauthorized /cc-end must not create a pending confirmation")
	}
}

func TestHandleEnd_Authorized_CreatesConfirmation(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	fs.listMeta = []lobby.GameMeta{{ID: id, Name: "friday-commander", State: "lobby"}}
	h := NewHandler(Config{GuildIDs: []string{"g1"}, AdminUserIDs: []string{"u1"}}, c, discardLogger())

	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionApplicationCommand,
		Member:  &discordgo.Member{User: discordUser("u1")},
		Data: discordgo.ApplicationCommandInteractionData{
			Name: CmdEnd,
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "game", Type: discordgo.ApplicationCommandOptionString, Value: "friday"},
			},
		},
	}}
	dispatchRecover(t, h, i)

	if len(h.confirmations.pending) != 1 {
		t.Fatalf("want exactly one pending confirmation, got %d", len(h.confirmations.pending))
	}
	for _, p := range h.confirmations.pending {
		if p.GameID != id || p.InvokerID != "u1" {
			t.Errorf("pending confirmation: got %+v", p)
		}
	}
}

func TestHandleEnd_AlreadyArchived_NoConfirmation(t *testing.T) {
	fs, c := newFakeServer(t)
	archivedAt := time.Now().UTC()
	id := uuid.New()
	fs.getMeta = lobby.GameMeta{ID: id, Name: "friday", ArchivedAt: &archivedAt}
	h := NewHandler(Config{GuildIDs: []string{"g1"}, AdminUserIDs: []string{"u1"}}, c, discardLogger())

	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionApplicationCommand,
		Member:  &discordgo.Member{User: discordUser("u1")},
		Data: discordgo.ApplicationCommandInteractionData{
			Name: CmdEnd,
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{Name: "game", Type: discordgo.ApplicationCommandOptionString, Value: id.String()},
			},
		},
	}}
	dispatchRecover(t, h, i)

	if len(h.confirmations.pending) != 0 {
		t.Error("an already-archived game must not produce a confirmation prompt")
	}
}

func TestDispatchComponent_Confirm_CallsArchive(t *testing.T) {
	fs, c := newFakeServer(t)
	id := uuid.New()
	fs.archiveMeta = lobby.GameMeta{ID: id, Name: "friday"}
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())
	token := "confirm-tok"
	h.confirmations.put(token, pendingEnd{GameID: id, GameName: "friday", InvokerID: "u1", ExpiresAt: h.now().Add(confirmTTL)}, h.now())

	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionMessageComponent,
		Member:  &discordgo.Member{User: discordUser("u1")},
		Data:    discordgo.MessageComponentInteractionData{CustomID: ccEndCustomIDPrefix + "confirm:" + token},
	}}
	dispatchRecover(t, h, i)

	if fs.archiveStatus != http.StatusOK {
		t.Fatalf("test setup: unexpected archiveStatus %d", fs.archiveStatus)
	}
	found := false
	for _, hdr := range fs.gotAuthHeaders {
		if hdr == "Bearer session-token" {
			found = true
		}
	}
	if !found {
		t.Error("confirm click should have called the server (no authorized request seen)")
	}
	if _, stillPending := h.confirmations.get(token); stillPending {
		t.Error("confirm click should consume the token")
	}
}

func TestDispatchComponent_Confirm_ArchiveError(t *testing.T) {
	fs, c := newFakeServer(t)
	fs.archiveStatus = http.StatusNotFound
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())
	token := "confirm-tok-err"
	id := uuid.New()
	h.confirmations.put(token, pendingEnd{GameID: id, GameName: "friday", InvokerID: "u1", ExpiresAt: h.now().Add(confirmTTL)}, h.now())

	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionMessageComponent,
		Member:  &discordgo.Member{User: discordUser("u1")},
		Data:    discordgo.MessageComponentInteractionData{CustomID: ccEndCustomIDPrefix + "confirm:" + token},
	}}
	dispatchRecover(t, h, i)

	if _, stillPending := h.confirmations.get(token); stillPending {
		t.Error("token should be consumed even when the archive call fails")
	}
}

func TestDispatchComponent_Cancel_DoesNotCallArchive(t *testing.T) {
	fs, c := newFakeServer(t)
	h := NewHandler(Config{GuildIDs: []string{"g1"}}, c, discardLogger())
	token := "cancel-tok"
	h.confirmations.put(token, pendingEnd{GameID: uuid.New(), GameName: "friday", InvokerID: "u1", ExpiresAt: h.now().Add(confirmTTL)}, h.now())

	i := &discordgo.InteractionCreate{Interaction: &discordgo.Interaction{
		GuildID: "g1",
		Type:    discordgo.InteractionMessageComponent,
		Member:  &discordgo.Member{User: discordUser("u1")},
		Data:    discordgo.MessageComponentInteractionData{CustomID: ccEndCustomIDPrefix + "cancel:" + token},
	}}
	dispatchRecover(t, h, i)

	if len(fs.gotAuthHeaders) != 0 {
		t.Error("cancel must never call the server")
	}
	if _, stillPending := h.confirmations.get(token); stillPending {
		t.Error("cancel click should consume the token")
	}
}
