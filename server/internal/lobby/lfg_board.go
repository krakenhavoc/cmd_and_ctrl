package lobby

// lfg_board.go: the lobby's half of a table's #lfg post (ADR 0161) that
// does not talk to Discord. The host's choice (games.lfg), PUT
// /games/{id}/lfg, the post's status on the meta, and the calls the LFG
// manager (lfg_manager.go) makes to read a table and report back.

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The post's status on GameMeta.LFGStatus, for the table settings.
const (
	// LFGStatusPending: the table is opted in and the manager has not
	// posted yet, or the table cannot be posted right now (a guest
	// host, a game that started).
	LFGStatusPending = "pending"
	// LFGStatusPosted: the post is up in #lfg.
	LFGStatusPosted = "posted"
	// LFGStatusUnavailable: Discord refused the post, or did not answer
	// three times. The table plays as it would without it; the host's
	// toggle tries again.
	LFGStatusUnavailable = "unavailable"
)

var (
	// ErrLFGOff: this server has no #lfg board (ADR 0161 §7). 503 at
	// the HTTP layer, naming what is missing.
	ErrLFGOff = errors.New("lobby: the #lfg board is off on this server")
	// ErrLFGUnavailable: this table cannot be posted. It is the
	// tutorial's practice table, its game has started or ended, or it
	// has been archived. Turning the post off is always allowed.
	ErrLFGUnavailable = errors.New("lobby: only a table waiting in the lobby can be posted to #lfg")
	// ErrLFGFull: every seat is taken, so there is nothing to advertise.
	ErrLFGFull = errors.New("lobby: every seat at this table is taken")
	// ErrLFGGuestHost: the post names the host, and a guest host has no
	// Discord account to name (ADR 0161 §1).
	ErrLFGGuestHost = errors.New("lobby: the host must be signed in with Discord to post to #lfg")
)

// SetLFGManager attaches the #lfg manager, which turns the board on.
// Called once at boot, BEFORE RestoreFromDisk; Start the manager after.
func (l *Lobby) SetLFGManager(m *LFGManager) {
	m.lobby = l
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lfg = m
}

// LFGEnabled reports whether this server has an #lfg board.
func (l *Lobby) LFGEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lfg != nil
}

// SetLFG records the host's choice for a table (ADR 0161 §1). The
// caller has checked the host; this checks the table. Turning it on
// needs a lobby table with an empty seat, a host signed in with
// Discord, and an invite link this process holds (or a post that
// already carries one). Setting the value it already has is a no-op
// that still answers the meta.
func (l *Lobby) SetLFG(id uuid.UUID, enabled bool) (GameMeta, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lfg == nil {
		return GameMeta{}, ErrLFGOff
	}
	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.practice {
		return GameMeta{}, ErrLFGUnavailable
	}
	l.syncStateLocked(entry)
	if enabled {
		switch {
		case entry.meta.State != string(game.StateLobby) || entry.meta.Archived():
			return GameMeta{}, ErrLFGUnavailable
		case len(entry.meta.Players) >= game.MaxPlayers:
			return GameMeta{}, ErrLFGFull
		case l.lfgHostLocked(entry) == "":
			return GameMeta{}, ErrLFGGuestHost
		case entry.meta.InviteToken == "" && entry.lfgMessageID == "":
			// The post is a copy of the invite link, and a table restored
			// after a restart has only its hash (invite_dm.go has the
			// same gap). Refused rather than rotated, for the same
			// reason: rotating revokes the link the table already shared.
			return GameMeta{}, ErrInviteTokenUnavailable
		}
	}
	if entry.meta.LFG != enabled {
		entry.meta.LFG = enabled
		entry.meta.LFGStatus = ""
		if enabled {
			entry.meta.LFGStatus = LFGStatusPending
		}
		// The host's choice is the manager's reset: a fresh "on" tries
		// again after an unavailable post.
		l.lfg.toggled(id)
		l.persistGameLocked(entry)
	}
	return copyMeta(entry.meta), nil
}

// lfgHostLocked is the Discord id the post names as host: the seated
// host's, or the named host still waiting to sit down (ADR 0075 §2.1).
// "" when the host is a guest, or there is none. Callers hold l.mu.
func (l *Lobby) lfgHostLocked(entry *gameEntry) string {
	if entry.room == nil {
		return ""
	}
	host := entry.room.HostPlayerID()
	if host == uuid.Nil {
		return entry.meta.HostDiscordID
	}
	for _, p := range entry.meta.Players {
		if p.PlayerID == host && !p.IsBot && !p.IsAgent {
			return p.DiscordID
		}
	}
	return ""
}

// lfgChangedLocked is called wherever a table's seats, lifecycle,
// settings or invite are written. The manager reads the table again and
// edits or removes the post if that changed it. Callers hold l.mu.
func (l *Lobby) lfgChangedLocked(entry *gameEntry) {
	if l.lfg != nil && !entry.practice {
		l.lfg.notify(entry.meta.ID)
	}
}

// lfgTable is what table id wants of its post, for the LFG manager. ok
// is false when the table is gone or is the practice table.
func (l *Lobby) lfgTable(id uuid.UUID) (lfgTable, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.games[id]
	if !ok || entry.practice || entry.room == nil {
		return lfgTable{}, false
	}
	host := l.lfgHostLocked(entry)
	t := lfgTable{
		enabled: entry.meta.LFG,
		eligible: entry.room.Game.CurrentState() == game.StateLobby &&
			!entry.meta.Archived() && host != "",
		hostID:      host,
		inviteToken: entry.meta.InviteToken,
		messageID:   entry.lfgMessageID,
		post: discord.LFGPost{
			Table:     entry.meta.Name,
			HostID:    host,
			Seated:    len(entry.meta.Players),
			Seats:     game.MaxPlayers,
			Voice:     entry.meta.VoiceRoom && l.voiceOn,
			Settings:  lfgSettings(entry.room.Game.TableSettingsSnapshot()),
			CreatedAt: entry.meta.CreatedAt,
		},
	}
	return t, true
}

// lfgSettings are the table settings a post shows: the ones that change
// how the game plays, only when they are not the default.
func lfgSettings(s game.TableSettings) []discord.LFGField {
	def := game.DefaultTableSettings()
	var out []discord.LFGField
	if s.StartingLife != def.StartingLife {
		out = append(out, discord.LFGField{Name: "Starting life", Value: strconv.Itoa(s.StartingLife)})
	}
	if s.CommanderDamage != def.CommanderDamage {
		out = append(out, discord.LFGField{Name: "Commander damage", Value: strconv.Itoa(s.CommanderDamage)})
	}
	if s.UndoLimit != def.UndoLimit {
		v := fmt.Sprintf("%d per turn", s.UndoLimit)
		switch s.UndoLimit {
		case game.UndoUnlimited:
			v = "Unlimited"
		case 0:
			v = "None"
		}
		out = append(out, discord.LFGField{Name: "Undos", Value: v})
	}
	if s.AllowSpawn != def.AllowSpawn {
		out = append(out, discord.LFGField{Name: "Card spawning", Value: "On"})
	}
	return out
}

// setLFGMessage records table id's live post (games.lfg_message_id). A
// table that is not resident (deleted, or never restored) has its row
// written directly; a row that is gone is nothing to clear.
func (l *Lobby) setLFGMessage(id uuid.UUID, messageID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	ctx, cancel := storeCtx()
	defer cancel()
	entry, ok := l.games[id]
	if ok {
		if entry.lfgMessageID == messageID {
			return
		}
		entry.lfgMessageID = messageID
		if entry.practice {
			return
		}
		if err := l.store.UpdateGame(ctx, gameRecordLocked(entry)); err != nil {
			slog.Default().Warn("persist lfg message failed", "game_id", id, "err", err)
		}
		return
	}
	rec, _, err := l.store.LoadGame(ctx, id)
	if err != nil {
		if !errors.Is(err, ErrStoreNotFound) {
			slog.Default().Warn("load game to persist lfg message failed", "game_id", id, "err", err)
		}
		return
	}
	if rec.LFGMessageID == messageID {
		return
	}
	rec.LFGMessageID = messageID
	if err := l.store.UpdateGame(ctx, rec); err != nil {
		slog.Default().Warn("persist lfg message failed", "game_id", id, "err", err)
	}
}

// reportLFG takes the manager's status for table id onto its meta.
func (l *Lobby) reportLFG(id uuid.UUID, status string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if entry, ok := l.games[id]; ok {
		entry.meta.LFGStatus = status
	}
}

// lfgMessages maps every resident table's persisted post to its message
// id, and lfgTableIDs lists the resident tables opted in: the boot
// reconcile's two lists.
func (l *Lobby) lfgMessages() map[uuid.UUID]string {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := map[uuid.UUID]string{}
	for id, e := range l.games {
		if e.lfgMessageID != "" {
			out[id] = e.lfgMessageID
		}
	}
	return out
}

func (l *Lobby) lfgTableIDs() []uuid.UUID {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []uuid.UUID
	for id, e := range l.games {
		if e.meta.LFG && !e.practice {
			out = append(out, id)
		}
	}
	return out
}

// lfgOrphans lists the posts of games rows the restore pass did not
// bring back (ADR 0161 §6): closed by closeUnresidentRows, kept for a
// roll-forward, or ended long ago. Every one is deleted at boot. Read
// from the store without l.mu, which no store call needs.
func (l *Lobby) lfgOrphans(log *slog.Logger) map[uuid.UUID]string {
	ctx, cancel := storeCtx()
	ids, err := l.store.GameIDs(ctx)
	cancel()
	if err != nil {
		log.Warn("lfg: listing stored games at boot failed; their posts stay until the next restart", "err", err)
		return nil
	}
	l.mu.Lock()
	var unresident []uuid.UUID
	for _, id := range ids {
		if _, ok := l.games[id]; !ok {
			unresident = append(unresident, id)
		}
	}
	l.mu.Unlock()
	out := map[uuid.UUID]string{}
	for _, id := range unresident {
		ctx, cancel := storeCtx()
		rec, _, err := l.store.LoadGame(ctx, id)
		cancel()
		if err != nil {
			log.Warn("lfg: loading a stored game at boot failed", "game_id", id, "err", err)
			continue
		}
		if rec.LFGMessageID != "" {
			out[id] = rec.LFGMessageID
		}
	}
	return out
}

// lfgRequest is the body of PUT /games/{id}/lfg. Enabled is a pointer
// so an empty body is a 400, not a silent "off".
type lfgRequest struct {
	Enabled *bool `json:"enabled"`
}

// putLFG handles PUT /games/{id}/lfg: the host (or an admin) posts the
// table to #lfg or withdraws it (ADR 0161 §1). It answers the table's
// meta, like PUT /games/{id}/voice.
func putLFG(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	var body lfgRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if body.Enabled == nil {
		return httpError(http.StatusBadRequest, "enabled is required")
	}
	if !c.Lobby.LFGEnabled() {
		return errLFGOffHTTP
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !CanManageTable(p, meta, c.isAdmin(p)) {
		return ErrNotTableManager
	}
	meta, err = c.Lobby.SetLFG(id, *body.Enabled)
	if err := lfgHTTPError(err); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, redactMetaFor(p, c.isAdmin(p), id, meta))
}

// lfgHTTPError maps SetLFG's refusals onto statuses.
func lfgHTTPError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrLFGOff):
		return errLFGOffHTTP
	case errors.Is(err, ErrLFGUnavailable):
		return httpError(http.StatusConflict,
			"only a table waiting in the lobby can be posted to #lfg: this one is a practice table, has started or ended, or is archived")
	case errors.Is(err, ErrLFGFull):
		return httpError(http.StatusConflict, "every seat at this table is taken")
	case errors.Is(err, ErrLFGGuestHost):
		return httpError(http.StatusConflict, "sign in with Discord to post to #lfg: the post names the host")
	case errors.Is(err, ErrInviteTokenUnavailable):
		return httpError(http.StatusConflict,
			"this server no longer holds this table's invite link (it was lost to a restart), and the post is a copy of it; "+
				"mint a replacement with POST /games/{id}/invites/rotate first — note that doing so "+
				"stops the old link working for everyone who already has it")
	default:
		return err
	}
}

// errLFGOffHTTP is the 503 the route answers on a server without a
// board, naming what turns it on (ADR 0161 §7).
var errLFGOffHTTP = httpError(http.StatusServiceUnavailable,
	"the #lfg board is off on this server: "+discord.LFGChannelEnv+", "+
		discord.CommunityGuildEnv+" and "+discord.BotTokenEnv+" must all be set")
