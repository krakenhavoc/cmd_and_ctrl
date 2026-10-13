package lobby

// voice_room.go: the lobby's half of a table's Discord voice room (ADR
// 0159) that does not talk to Discord. The host's choice
// (games.voice_room), PUT /games/{id}/voice and POST
// /games/{id}/voice/refresh, the voice state the room stamps on the game
// view, and the calls the voice manager (voice_manager.go) makes to read
// a table and report back.

import (
	"errors"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

var (
	// ErrVoiceRoomsOff: this server offers no voice rooms (ADR 0159 §7).
	// 503 at the HTTP layer, naming what is missing.
	ErrVoiceRoomsOff = errors.New("lobby: voice rooms are off on this server")
	// ErrVoiceRoomUnavailable: this table cannot have one. It is the
	// tutorial's practice table, or it has ended or been archived and
	// the request turns voice on. Turning it off is always allowed.
	ErrVoiceRoomUnavailable = errors.New("lobby: this table cannot have a voice room")
)

// SetVoiceRoomsEnabled turns table voice rooms on or off for this
// server. Called once at boot, BEFORE RestoreFromDisk, so a restored
// table gets its stamp back.
func (l *Lobby) SetVoiceRoomsEnabled(on bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.voiceOn = on
}

// SetVoiceManager attaches the voice manager and turns voice rooms on.
// Called once at boot, BEFORE RestoreFromDisk; Start the manager after.
func (l *Lobby) SetVoiceManager(vm *VoiceManager) {
	vm.lobby = l
	l.mu.Lock()
	defer l.mu.Unlock()
	l.voice = vm
	l.voiceOn = true
}

// VoiceRoomsEnabled reports whether this server offers voice rooms.
func (l *Lobby) VoiceRoomsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.voiceOn
}

// SetVoiceRoom records the host's choice for a table (ADR 0159 §1). The
// caller has checked the host; this checks the table. Setting the value
// it already has is a no-op that still answers the meta. A change goes
// through the room, so the table sees it at once.
func (l *Lobby) SetVoiceRoom(id uuid.UUID, enabled bool) (GameMeta, error) {
	l.mu.Lock()
	if !l.voiceOn {
		l.mu.Unlock()
		return GameMeta{}, ErrVoiceRoomsOff
	}
	entry, ok := l.games[id]
	if !ok {
		l.mu.Unlock()
		return GameMeta{}, ErrGameNotFound
	}
	if entry.practice {
		l.mu.Unlock()
		return GameMeta{}, ErrVoiceRoomUnavailable
	}
	l.syncStateLocked(entry)
	if enabled && (entry.meta.State == string(game.StateEnded) || entry.meta.Archived()) {
		l.mu.Unlock()
		return GameMeta{}, ErrVoiceRoomUnavailable
	}
	if entry.meta.VoiceRoom == enabled {
		meta := copyMeta(entry.meta)
		l.mu.Unlock()
		return meta, nil
	}
	entry.meta.VoiceRoom = enabled
	// The manager's last report was about the old choice; until it
	// reports on the new one, the table shows pending or off.
	entry.voiceReport = nil
	// applyLocked restamps the voice state inside the commit, so its
	// capture carries the change.
	broadcast, err := l.applyLocked(id, entry, func() error { return nil })
	l.persistGameLocked(entry)
	meta := copyMeta(entry.meta)
	l.mu.Unlock()
	if err == nil {
		broadcast()
	}
	return meta, nil
}

// voiceViewFor is the voice state the room stamps for entry: nil when
// the feature is off or the table is the practice table, and otherwise
// the host's choice, the room's status and every seat's access.
//
// Until the voice manager reports a room, a table that asked for one is
// "pending" and every Discord seat's access is "pending" too. A seat
// with no Discord id (a guest, a bot, an agent) has nothing to grant.
//
// The seats are the engine's, not meta.Players: applyLocked restamps
// inside a join's own commit, before the lobby's meta catches up.
func (l *Lobby) voiceViewFor(entry *gameEntry) *protocol.VoiceView {
	if !l.voiceOn || entry.practice || entry.room == nil {
		return nil
	}
	seats := entry.room.Game.Seats
	v := &protocol.VoiceView{
		Enabled: entry.meta.VoiceRoom,
		Status:  protocol.VoiceStatusOff,
		Access:  make(map[string]string, len(seats)),
	}
	rep := entry.voiceReport
	if entry.meta.VoiceRoom {
		v.Status = protocol.VoiceStatusPending
		if rep != nil && rep.status != "" {
			v.Status = rep.status
			v.URL = rep.url
		}
	}
	for _, p := range seats {
		access := protocol.VoiceAccessNoDiscord
		if voiceMember(p.DiscordID, p.IsBot, p.Agent) {
			access = protocol.VoiceAccessPending
			if rep != nil && rep.access[p.DiscordID] != "" {
				access = rep.access[p.DiscordID]
			}
		}
		v.Access[p.ID.String()] = access
	}
	return v
}

// voiceMember reports whether a seat can be let into a room: a person
// signed in with Discord, not a bot and not an agent.
func voiceMember(discordID string, isBot, agent bool) bool {
	return discordID != "" && !isBot && !agent
}

// restampVoice sets entry's room to its current voice state, if that
// changed. Callers hold l.mu, or own entry outright (loadEntry, before
// the entry is registered).
func (l *Lobby) restampVoice(entry *gameEntry) {
	if entry.room == nil {
		return
	}
	next := l.voiceViewFor(entry)
	if !reflect.DeepEqual(entry.room.Voice(), next) {
		entry.room.SetVoice(next)
	}
}

// voiceChangedLocked is called wherever a table's seats or lifecycle
// are written (persistSeatsLocked, persistGameLocked). It keeps the
// room's stamp current for the next capture. It pushes nothing: the
// mutations that change what a seat may see go through applyLocked,
// which restamps inside their own commit.
func (l *Lobby) voiceChangedLocked(entry *gameEntry) {
	l.restampVoice(entry)
	if l.voice != nil {
		l.voice.notify(entry.meta.ID)
	}
}

// voiceTable is what table id wants of its room, for the voice manager.
// ok is false when the table is gone or is the practice table.
func (l *Lobby) voiceTable(id uuid.UUID) (voiceTable, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok || entry.practice || entry.room == nil {
		return voiceTable{}, false
	}
	t := voiceTable{
		name:      entry.meta.Name,
		enabled:   entry.meta.VoiceRoom,
		archived:  entry.meta.Archived(),
		ended:     entry.meta.State == string(game.StateEnded),
		members:   map[string]bool{},
		channelID: entry.voiceChannelID,
		seq:       entry.room.Seq(),
		pushSeq:   entry.voicePushSeq,
	}
	if entry.endedAt != nil {
		t.endedAt = *entry.endedAt
	}
	for _, p := range entry.room.Game.Seats {
		if voiceMember(p.DiscordID, p.IsBot, p.Agent) {
			t.members[p.DiscordID] = true
		}
	}
	return t, true
}

// setVoiceChannel records table id's live room (games.voice_channel_id).
func (l *Lobby) setVoiceChannel(id uuid.UUID, channelID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok || entry.voiceChannelID == channelID {
		return
	}
	entry.voiceChannelID = channelID
	ctx, cancel := storeCtx()
	defer cancel()
	if err := l.store.UpdateGame(ctx, gameRecordLocked(entry)); err != nil {
		slog.Default().Warn("persist voice channel failed", "game_id", id, "err", err)
	}
}

// reportVoice takes the voice manager's report on table id and, when it
// changes what the table sees, pushes a fresh view.
func (l *Lobby) reportVoice(id uuid.UUID, rep voiceReport) {
	l.mu.Lock()
	entry, ok := l.games[id]
	if !ok || entry.room == nil {
		l.mu.Unlock()
		return
	}
	entry.voiceReport = &rep
	if reflect.DeepEqual(entry.room.Voice(), l.voiceViewFor(entry)) {
		l.mu.Unlock()
		return
	}
	broadcast, err := l.applyLocked(id, entry, func() error { return nil })
	if err == nil {
		entry.voicePushSeq = entry.room.Seq()
	}
	l.mu.Unlock()
	if err == nil {
		broadcast()
	}
}

// voiceChannels maps every live table's persisted room to the table,
// for the boot reconcile.
func (l *Lobby) voiceChannels() map[string]uuid.UUID {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := map[string]uuid.UUID{}
	for id, e := range l.games {
		if e.voiceChannelID != "" {
			out[e.voiceChannelID] = id
		}
	}
	return out
}

// voiceTableIDs lists the live tables that asked for a room.
func (l *Lobby) voiceTableIDs() []uuid.UUID {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []uuid.UUID
	for id, e := range l.games {
		if e.meta.VoiceRoom && !e.practice {
			out = append(out, id)
		}
	}
	return out
}

// RefreshVoice is POST /games/{id}/voice/refresh: the seat playerID asks
// the manager to check its membership again and to reopen an idle room.
// ErrPlayerNotInGame for a seat that is not here; ErrVoiceRoomUnavailable
// for a seat with no Discord id, or a table without voice on.
func (l *Lobby) RefreshVoice(id, playerID uuid.UUID) error {
	l.mu.Lock()
	if l.voice == nil {
		l.mu.Unlock()
		return ErrVoiceRoomsOff
	}
	entry, ok := l.games[id]
	if !ok {
		l.mu.Unlock()
		return ErrGameNotFound
	}
	if !entry.meta.VoiceRoom || entry.practice {
		l.mu.Unlock()
		return ErrVoiceRoomUnavailable
	}
	p := entry.room.Game.PlayerByID(playerID)
	if p == nil {
		l.mu.Unlock()
		return ErrPlayerNotInGame
	}
	if !voiceMember(p.DiscordID, p.IsBot, p.Agent) {
		l.mu.Unlock()
		return ErrVoiceRoomUnavailable
	}
	vm, discordID := l.voice, p.DiscordID
	l.mu.Unlock()
	vm.refresh(id, discordID)
	return nil
}

// voiceRoomRequest is the body of PUT /games/{id}/voice. Enabled is a
// pointer so an empty body is a 400, not a silent "off".
type voiceRoomRequest struct {
	Enabled *bool `json:"enabled"`
}

// putVoiceRoom handles PUT /games/{id}/voice: the host (or an admin)
// turns the table's Discord voice room on or off (ADR 0159 §1). It
// answers the table's meta, like POST /games/{id}/host.
func putVoiceRoom(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	var body voiceRoomRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if body.Enabled == nil {
		return httpError(http.StatusBadRequest, "enabled is required")
	}
	if !c.Lobby.VoiceRoomsEnabled() {
		return errVoiceRoomsOffHTTP
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !CanManageTable(p, meta, c.isAdmin(p)) {
		return ErrNotTableManager
	}
	meta, err = c.Lobby.SetVoiceRoom(id, *body.Enabled)
	switch {
	case errors.Is(err, ErrVoiceRoomUnavailable):
		return httpError(http.StatusConflict,
			"this table cannot have a voice room: it is a practice table, or it has ended or been archived")
	case errors.Is(err, ErrVoiceRoomsOff):
		return errVoiceRoomsOffHTTP
	case err != nil:
		return err
	}
	return writeJSON(w, http.StatusOK, redactMetaFor(p, c.isAdmin(p), id, meta))
}

// refreshVoiceRoom handles POST /games/{id}/voice/refresh: a seated
// player asks to be checked again (they just joined the Discord server)
// or to reopen a room removed for idleness (ADR 0159 §3). 202: the voice
// manager does the rest and the table sees it on the game view.
func refreshVoiceRoom(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	if p.GameID != id || p.PlayerID == uuid.Nil {
		return httpError(http.StatusForbidden, "only a player seated at this table may do that")
	}
	err = c.Lobby.RefreshVoice(id, p.PlayerID)
	switch {
	case errors.Is(err, ErrVoiceRoomsOff):
		return errVoiceRoomsOffHTTP
	case errors.Is(err, ErrVoiceRoomUnavailable):
		return httpError(http.StatusConflict, "this table has no voice room, or this seat did not sign in with Discord")
	case errors.Is(err, ErrPlayerNotInGame):
		return httpError(http.StatusForbidden, "only a player seated at this table may do that")
	case err != nil:
		return err
	}
	w.WriteHeader(http.StatusAccepted)
	return nil
}

// errVoiceRoomsOffHTTP is the 503 a voice route answers on a server
// that offers no rooms, naming what turns them on (ADR 0159 §7).
var errVoiceRoomsOffHTTP = httpError(http.StatusServiceUnavailable,
	"voice rooms are off on this server: "+discord.VoiceCategoryEnv+", "+
		discord.CommunityGuildEnv+" and "+discord.BotTokenEnv+" must all be set")
