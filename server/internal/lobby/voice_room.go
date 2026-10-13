package lobby

// voice_room.go: the lobby's half of a table's Discord voice room (ADR
// 0159) that does not talk to Discord. The host's choice
// (games.voice_room), PUT /games/{id}/voice, and the voice state the
// room stamps on the game view.
//
// The voice manager (ADR 0159 slice 3b) creates and removes rooms and
// reports their state; until it is wired, a table that asks for a room
// shows "pending".

import (
	"errors"
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
	if entry.meta.VoiceRoom {
		v.Status = protocol.VoiceStatusPending
	}
	for _, p := range seats {
		access := protocol.VoiceAccessNoDiscord
		if p.DiscordID != "" && !p.IsBot && !p.Agent {
			access = protocol.VoiceAccessPending
		}
		v.Access[p.ID.String()] = access
	}
	return v
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

// errVoiceRoomsOffHTTP is the 503 a voice route answers on a server
// that offers no rooms, naming what turns them on (ADR 0159 §7).
var errVoiceRoomsOffHTTP = httpError(http.StatusServiceUnavailable,
	"voice rooms are off on this server: "+discord.VoiceCategoryEnv+", "+
		discord.CommunityGuildEnv+" and "+discord.BotTokenEnv+" must all be set")
