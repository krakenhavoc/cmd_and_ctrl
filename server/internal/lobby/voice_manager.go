package lobby

// voice_manager.go: ADR 0159 §2–§5, the part of a table's voice room
// that talks to Discord. It creates a table's room when the table wants
// one, lets the seated server members in and out, removes the room after
// the game, when the table is left idle, or when the host turns it off,
// and on boot adopts the rooms that survived a restart and deletes the
// ones no table claims.
//
// It is a reconciler. The lobby says only "table X changed" (notify);
// the worker reads what the table wants now (Lobby.voiceTable), compares
// it with what it knows exists, and makes the Discord calls that close
// the gap. A missed or repeated notify costs nothing.
//
// Locking. The worker never holds l.mu across a Discord call: it reads
// the table under l.mu, lets go, calls Discord, then reports back under
// l.mu. A table's voiceRoomState is touched only on the worker goroutine
// (reconcile, sweep, boot), so it needs no lock. vm.mu guards only the
// maps other goroutines reach (dirty, refreshes, rooms, botID) and is
// never held while calling into the lobby, so the order is always l.mu
// before vm.mu (notify, from inside the lobby) and never the reverse.

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The owner's numbers (ADR 0159 §2, §4).
const (
	// voiceEndGrace is how long a room outlives its game.
	voiceEndGrace = 10 * time.Minute
	// voiceIdleAfter is how long a table can go without a game action
	// or a seat change before its room is removed, if nobody is in it.
	voiceIdleAfter = 2 * time.Hour
	// voiceSweepEvery is how often the worker checks grace and idleness.
	voiceSweepEvery = time.Minute
	// voiceMaxAttempts is how many transient failures in a row a table
	// tolerates before it shows "unavailable".
	voiceMaxAttempts = 3
	// voiceCallTimeout bounds one Discord call.
	voiceCallTimeout = 10 * time.Second
	// voicePollEvery is how often the worker reads who is connected to
	// each live room (#3311). The server holds no gateway connection, so
	// this is a REST read per granted seat; a change is pushed only when
	// the answer differs from the last one.
	voicePollEvery = 20 * time.Second
	// voiceMaxRetryWait caps a 429's wait.
	voiceMaxRetryWait = 30 * time.Second
)

// voiceRoomCap is the most live rooms at once: under Discord's
// 50-channel category limit, with room for the owner's own. A var only
// so a test can lower it.
var voiceRoomCap = 45

// voiceAPI is the Discord half the manager uses: discord.Voice, or a
// fake in tests.
type voiceAPI interface {
	BotUserID(ctx context.Context, client *http.Client) (string, error)
	CreateRoom(ctx context.Context, client *http.Client, name, botUserID string, members []string) (string, error)
	GrantMember(ctx context.Context, client *http.Client, channelID, userID string) error
	RevokeMember(ctx context.Context, client *http.Client, channelID, userID string) error
	DeleteRoom(ctx context.Context, client *http.Client, channelID string) error
	Rooms(ctx context.Context, client *http.Client) ([]discord.Room, error)
	IsMember(ctx context.Context, client *http.Client, userID string) (bool, error)
	ConnectedChannel(ctx context.Context, client *http.Client, userID string) (string, error)
	VoiceStateOf(ctx context.Context, client *http.Client, userID string) (discord.VoiceState, error)
	MoveTo(ctx context.Context, client *http.Client, channelID, userID string) error
	DisconnectFrom(ctx context.Context, client *http.Client, channelID, userID string) (bool, error)
	RoomURL(channelID string) string
}

// voiceTable is what one table wants, read under l.mu.
type voiceTable struct {
	name     string
	enabled  bool
	archived bool
	ended    bool
	endedAt  time.Time
	// members are the seated Discord users (not bots, not agents).
	members map[string]bool
	// channelID is the persisted room, games.voice_channel_id.
	channelID string
	// seq is the room's sequence number, and pushSeq the one the
	// manager's own last push left it at, so a push is not activity.
	seq, pushSeq uint64
}

// voiceReport is what the manager tells the lobby about a table: the
// room's status, its link, and each Discord member's access.
type voiceReport struct {
	status string
	url    string
	access map[string]string // Discord id → protocol.VoiceAccess*
	// connected is each Discord id connected to the room, to
	// protocol.VoiceConnLive or VoiceConnMuted (#3311).
	connected map[string]string
}

// connection is Discord id d's connection to the room, "" for none. Safe
// on a nil report.
func (r *voiceReport) connection(d string) string {
	if r == nil {
		return ""
	}
	return r.connected[d]
}

// voiceRoomState is the manager's own record of one table.
type voiceRoomState struct {
	channelID string
	// granted are the Discord ids holding an overwrite on the room.
	granted map[string]bool
	// notMember are the Discord ids known not to be in the server.
	notMember map[string]bool
	// idle is set when the room was removed for idleness; the table
	// waits for a reopen (RefreshVoice, the host's toggle, a seat
	// change).
	idle bool
	// unavailable is set after a failure that will not fix itself, or
	// voiceMaxAttempts transient ones.
	unavailable bool
	full        bool
	attempts    int
	lastSeq     uint64
	lastActive  time.Time
	lastMembers map[string]bool
	// connected is who was last seen connected to the room (#3311):
	// Discord id → protocol.VoiceConn*.
	connected map[string]string
}

// VoiceManager runs the voice rooms of one lobby. Construct it with
// NewVoiceManager, attach it with Lobby.SetVoiceManager before
// RestoreFromDisk, and Start it after.
type VoiceManager struct {
	api    voiceAPI
	client *http.Client
	log    *slog.Logger
	now    func() time.Time
	lobby  *Lobby

	mu    sync.Mutex
	rooms map[uuid.UUID]*voiceRoomState
	dirty map[uuid.UUID]bool
	// refreshes are POST /games/{id}/voice/refresh asks waiting for the
	// worker: the Discord ids to check again, per table.
	refreshes map[uuid.UUID][]string
	botID     string
	kick      chan struct{}
	// pullIn moves a granted player who is in another voice channel into
	// the table's room (#3311). Set before Start.
	pullIn bool
}

// NewVoiceManager builds a manager over the configured Discord voice
// calls (ADR 0159 slice 2).
func NewVoiceManager(v discord.Voice, log *slog.Logger) *VoiceManager {
	vm := newVoiceManager(v, &http.Client{Timeout: voiceCallTimeout}, log, time.Now)
	vm.pullIn = v.PullIn
	return vm
}

func newVoiceManager(api voiceAPI, client *http.Client, log *slog.Logger, now func() time.Time) *VoiceManager {
	if log == nil {
		log = slog.Default()
	}
	return &VoiceManager{
		api:       api,
		client:    client,
		log:       log,
		now:       now,
		rooms:     map[uuid.UUID]*voiceRoomState{},
		dirty:     map[uuid.UUID]bool{},
		refreshes: map[uuid.UUID][]string{},
		kick:      make(chan struct{}, 1),
	}
}

// notify marks a table for reconciling and wakes the worker. Never
// blocks; safe under l.mu.
func (vm *VoiceManager) notify(id uuid.UUID) {
	vm.mu.Lock()
	vm.dirty[id] = true
	vm.mu.Unlock()
	select {
	case vm.kick <- struct{}{}:
	default:
	}
}

// Start runs the boot reconcile and then the worker until ctx ends.
func (vm *VoiceManager) Start(ctx context.Context) {
	go func() {
		vm.boot(ctx)
		vm.run(ctx)
	}()
}

func (vm *VoiceManager) run(ctx context.Context) {
	ticker := time.NewTicker(voiceSweepEvery)
	defer ticker.Stop()
	poll := time.NewTicker(voicePollEvery)
	defer poll.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-vm.kick:
			vm.drain(ctx)
		case <-ticker.C:
			vm.sweep(ctx)
			vm.drain(ctx)
		case <-poll.C:
			vm.pollConnections(ctx)
		}
	}
}

// drain reconciles every table marked dirty, in a stable order.
func (vm *VoiceManager) drain(ctx context.Context) {
	for {
		vm.mu.Lock()
		ids := make([]uuid.UUID, 0, len(vm.dirty))
		for id := range vm.dirty {
			ids = append(ids, id)
		}
		vm.dirty = map[uuid.UUID]bool{}
		vm.mu.Unlock()
		if len(ids) == 0 {
			return
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
		for _, id := range ids {
			if ctx.Err() != nil {
				return
			}
			vm.reconcile(ctx, id)
		}
	}
}

// state returns the manager's record of a table, creating it.
func (vm *VoiceManager) state(id uuid.UUID) *voiceRoomState {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	st, ok := vm.rooms[id]
	if !ok {
		st = &voiceRoomState{granted: map[string]bool{}, notMember: map[string]bool{}, lastActive: vm.now()}
		vm.rooms[id] = st
	}
	return st
}

// liveRooms counts the rooms that exist. Worker only.
func (vm *VoiceManager) liveRooms() int {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	n := 0
	for _, st := range vm.rooms {
		if st.channelID != "" {
			n++
		}
	}
	return n
}

// call bounds one Discord call, under the worker's context so a shutdown
// abandons it.
func (vm *VoiceManager) call(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, voiceCallTimeout)
}

// reconcile closes the gap between what table id wants and what exists.
// Only the worker calls it, so a table's state is never reconciled twice
// at once.
func (vm *VoiceManager) reconcile(ctx context.Context, id uuid.UUID) {
	t, ok := vm.lobby.voiceTable(id)
	st := vm.state(id)
	now := vm.now()

	if ok && !sameSet(st.lastMembers, t.members) {
		// A seat change is activity, and a reason to reopen.
		st.lastActive = now
		st.lastMembers = copySet(t.members)
		st.idle = false
	}
	vm.mu.Lock()
	asked := vm.refreshes[id]
	delete(vm.refreshes, id)
	vm.mu.Unlock()
	if asked != nil {
		// Someone at the table asked again: forget who was outside the
		// server, and reopen a room removed for idleness.
		for _, d := range asked {
			delete(st.notMember, d)
		}
		st.idle, st.unavailable, st.attempts = false, false, 0
		st.lastActive = now
	}

	want := ok && t.enabled && !t.archived && !st.idle && len(t.members) > 0 &&
		!(t.ended && !now.Before(t.endedAt.Add(voiceEndGrace)))
	if !ok || !t.enabled {
		// The host's choice is the reset: a fresh "on" starts clean.
		st.idle, st.unavailable, st.full, st.attempts = false, false, false, 0
	}

	if !want {
		if st.channelID != "" {
			if err := vm.deleteRoom(ctx, st); err != nil {
				vm.failed(id, st, "delete", err)
				return
			}
			vm.lobby.setVoiceChannel(id, "")
		}
		if !ok {
			vm.mu.Lock()
			delete(vm.rooms, id)
			vm.mu.Unlock()
			return
		}
		vm.publish(id, t, st)
		return
	}
	if st.unavailable {
		vm.publish(id, t, st)
		return
	}

	// Who is in the server: ask about each member not yet granted and
	// not already known to be outside it.
	var toGrant []string
	for _, d := range sortedKeys(t.members) {
		if st.granted[d] || st.notMember[d] {
			continue
		}
		cctx, cancel := vm.call(ctx)
		member, err := vm.api.IsMember(cctx, vm.client, d)
		cancel()
		if err != nil {
			vm.failed(id, st, "member check", err)
			return
		}
		if member {
			toGrant = append(toGrant, d)
		} else {
			st.notMember[d] = true
		}
	}

	if st.channelID == "" {
		if vm.liveRooms() >= voiceRoomCap {
			st.full = true
			vm.publish(id, t, st)
			return
		}
		st.full = false
		botID, err := vm.ensureBotID(ctx)
		if err != nil {
			vm.failed(id, st, "bot id", err)
			return
		}
		cctx, cancel := vm.call(ctx)
		channel, err := vm.api.CreateRoom(cctx, vm.client, t.name, botID, toGrant)
		cancel()
		if err != nil {
			vm.failed(id, st, "create", err)
			return
		}
		st.channelID = channel
		st.granted = map[string]bool{}
		for _, d := range toGrant {
			st.granted[d] = true
		}
		st.lastActive = now
		vm.lobby.setVoiceChannel(id, channel)
	} else {
		for _, d := range toGrant {
			cctx, cancel := vm.call(ctx)
			err := vm.api.GrantMember(cctx, vm.client, st.channelID, d)
			cancel()
			if errors.Is(err, discord.ErrVoiceNotFound) {
				// The room was deleted on Discord. Forget it; the next
				// pass makes a fresh one.
				vm.forgetRoom(id, st)
				return
			}
			if err != nil {
				vm.failed(id, st, "grant", err)
				return
			}
			st.granted[d] = true
		}
	}

	// Members who left: hide the room from them and take them out of it.
	for _, d := range sortedKeys(st.granted) {
		if t.members[d] {
			continue
		}
		cctx, cancel := vm.call(ctx)
		err := vm.api.RevokeMember(cctx, vm.client, st.channelID, d)
		if err == nil {
			_, err = vm.api.DisconnectFrom(cctx, vm.client, st.channelID, d)
		}
		cancel()
		if err != nil {
			vm.failed(id, st, "revoke", err)
			return
		}
		delete(st.granted, d)
	}
	for d := range st.notMember {
		if !t.members[d] {
			delete(st.notMember, d)
		}
	}
	st.attempts = 0
	vm.publish(id, t, st)
}

func (vm *VoiceManager) deleteRoom(ctx context.Context, st *voiceRoomState) error {
	cctx, cancel := vm.call(ctx)
	defer cancel()
	if err := vm.api.DeleteRoom(cctx, vm.client, st.channelID); err != nil {
		return err
	}
	st.channelID = ""
	st.granted = map[string]bool{}
	st.connected = nil
	return nil
}

// forgetRoom drops a room Discord no longer has and reconciles again.
func (vm *VoiceManager) forgetRoom(id uuid.UUID, st *voiceRoomState) {
	st.channelID = ""
	st.granted = map[string]bool{}
	st.connected = nil
	vm.lobby.setVoiceChannel(id, "")
	vm.notify(id)
}

func (vm *VoiceManager) ensureBotID(ctx context.Context) (string, error) {
	vm.mu.Lock()
	id := vm.botID
	vm.mu.Unlock()
	if id != "" {
		return id, nil
	}
	cctx, cancel := vm.call(ctx)
	defer cancel()
	id, err := vm.api.BotUserID(cctx, vm.client)
	if err != nil {
		return "", err
	}
	vm.mu.Lock()
	vm.botID = id
	vm.mu.Unlock()
	return id, nil
}

// failed records a failed Discord call. A transient one is retried after
// a wait, up to voiceMaxAttempts in a row; anything else, or one too
// many, shows "unavailable" until the next lifecycle event or toggle.
// The game is never affected.
func (vm *VoiceManager) failed(id uuid.UUID, st *voiceRoomState, step string, err error) {
	st.attempts++
	if discord.Transient(err) && st.attempts < voiceMaxAttempts {
		wait := time.Duration(st.attempts) * 5 * time.Second
		var rl *discord.RateLimitError
		if errors.As(err, &rl) && rl.RetryAfter > wait {
			wait = min(rl.RetryAfter, voiceMaxRetryWait)
		}
		vm.log.Info("voice room call failed; retrying", "game_id", id, "step", step, "err", err, "wait", wait)
		time.AfterFunc(wait, func() { vm.notify(id) })
		return
	}
	vm.log.Warn("voice room unavailable", "game_id", id, "step", step, "err", err)
	st.unavailable = true
	if t, ok := vm.lobby.voiceTable(id); ok {
		vm.publish(id, t, st)
	}
}

// publish tells the lobby the table's room status and each member's
// access, which restamps the game view.
func (vm *VoiceManager) publish(id uuid.UUID, t voiceTable, st *voiceRoomState) {
	rep := voiceReport{access: map[string]string{}, connected: map[string]string{}}
	for d := range t.members {
		switch {
		case st.granted[d]:
			rep.access[d] = protocol.VoiceAccessGranted
			if c := st.connected[d]; c != "" && st.channelID != "" {
				rep.connected[d] = c
			}
		case st.notMember[d]:
			rep.access[d] = protocol.VoiceAccessNotInServer
		default:
			rep.access[d] = protocol.VoiceAccessPending
		}
	}
	switch {
	case !t.enabled:
		rep.status = protocol.VoiceStatusOff
	case st.channelID != "":
		rep.status = protocol.VoiceStatusReady
		rep.url = vm.api.RoomURL(st.channelID)
	case st.unavailable:
		rep.status = protocol.VoiceStatusUnavailable
	case st.full:
		rep.status = protocol.VoiceStatusFull
	case st.idle:
		rep.status = protocol.VoiceStatusIdle
	case t.ended || t.archived:
		rep.status = protocol.VoiceStatusOff
	default:
		rep.status = protocol.VoiceStatusPending
	}
	vm.lobby.reportVoice(id, rep)
}

// sweep handles the clock: rooms past their game's grace, and rooms
// idle for voiceIdleAfter with nobody connected.
func (vm *VoiceManager) sweep(ctx context.Context) {
	vm.mu.Lock()
	ids := make([]uuid.UUID, 0, len(vm.rooms))
	for id, st := range vm.rooms {
		if st.channelID != "" {
			ids = append(ids, id)
		}
	}
	vm.mu.Unlock()
	now := vm.now()
	for _, id := range ids {
		t, ok := vm.lobby.voiceTable(id)
		st := vm.state(id)
		if !ok || t.ended || t.archived || !t.enabled {
			vm.notify(id)
			continue
		}
		if t.seq != st.lastSeq {
			if t.seq != t.pushSeq {
				st.lastActive = now
			}
			st.lastSeq = t.seq
		}
		if now.Sub(st.lastActive) < voiceIdleAfter {
			continue
		}
		if vm.anyoneConnected(ctx, st) {
			st.lastActive = now
			continue
		}
		vm.log.Info("voice room idle; removing it", "game_id", id)
		st.idle = true
		vm.notify(id)
	}
}

// pollConnections reads who is connected to each live room and publishes
// the tables where that changed (#3311). With pull-in on, a granted
// player found in a different voice channel is moved into the room; a
// player already in the room is never touched. Worker only.
func (vm *VoiceManager) pollConnections(ctx context.Context) {
	vm.mu.Lock()
	ids := make([]uuid.UUID, 0, len(vm.rooms))
	for id, st := range vm.rooms {
		if st.channelID != "" {
			ids = append(ids, id)
		}
	}
	vm.mu.Unlock()
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}
		t, ok := vm.lobby.voiceTable(id)
		st := vm.state(id)
		if !ok || !t.enabled || st.channelID == "" {
			continue
		}
		next := map[string]string{}
		for _, d := range sortedKeys(st.granted) {
			if !t.members[d] {
				continue
			}
			cctx, cancel := vm.call(ctx)
			vs, err := vm.api.VoiceStateOf(cctx, vm.client, d)
			cancel()
			if err != nil {
				// Keep what was last seen for this seat.
				if c := st.connected[d]; c != "" {
					next[d] = c
				}
				continue
			}
			if vs.ChannelID != "" && vs.ChannelID != st.channelID && vm.pullIn {
				mctx, mcancel := vm.call(ctx)
				merr := vm.api.MoveTo(mctx, vm.client, st.channelID, d)
				mcancel()
				if merr != nil {
					vm.log.Info("voice pull-in failed", "game_id", id, "err", merr)
				} else {
					vs.ChannelID = st.channelID
				}
			}
			if vs.ChannelID != st.channelID {
				continue
			}
			if vs.Muted {
				next[d] = protocol.VoiceConnMuted
			} else {
				next[d] = protocol.VoiceConnLive
			}
		}
		if !mapsEqual(st.connected, next) {
			st.connected = next
			vm.publish(id, t, st)
		}
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// anyoneConnected reports whether a granted member is in the room. A
// failed check counts as "yes": never remove a room someone may be in.
func (vm *VoiceManager) anyoneConnected(ctx context.Context, st *voiceRoomState) bool {
	for _, d := range sortedKeys(st.granted) {
		cctx, cancel := vm.call(ctx)
		ch, err := vm.api.ConnectedChannel(cctx, vm.client, d)
		cancel()
		if err != nil || ch == st.channelID {
			return true
		}
	}
	return false
}

// boot is ADR 0159 §5's reconcile after a restart: keep the rooms a
// table still claims, delete every other channel in the category, and
// reconcile every voice table.
func (vm *VoiceManager) boot(ctx context.Context) {
	claimed := vm.lobby.voiceChannels()
	cctx, cancel := vm.call(ctx)
	rooms, err := vm.api.Rooms(cctx, vm.client)
	cancel()
	if err != nil {
		vm.log.Warn("voice rooms: listing the category at boot failed; stray rooms stay until the next restart", "err", err)
	} else {
		exists := map[string]bool{}
		for _, r := range rooms {
			exists[r.ID] = true
			if _, ok := claimed[r.ID]; ok {
				continue
			}
			dctx, dcancel := vm.call(ctx)
			if derr := vm.api.DeleteRoom(dctx, vm.client, r.ID); derr != nil {
				vm.log.Warn("voice rooms: deleting a stray room failed", "channel_id", r.ID, "err", derr)
			}
			dcancel()
		}
		for channel, id := range claimed {
			if !exists[channel] {
				vm.lobby.setVoiceChannel(id, "")
			}
		}
	}
	for channel, id := range claimed {
		if err != nil || channelExists(rooms, channel) {
			st := vm.state(id)
			st.channelID = channel
			st.granted = map[string]bool{}
		}
	}
	for _, id := range vm.lobby.voiceTableIDs() {
		vm.notify(id)
	}
	vm.drain(ctx)
}

// refresh is POST /games/{id}/voice/refresh's half here. The worker
// forgets that discordID was outside the server, reopens an idle or
// unavailable room, and reconciles. Safe from any goroutine.
func (vm *VoiceManager) refresh(id uuid.UUID, discordID string) {
	vm.mu.Lock()
	vm.refreshes[id] = append(vm.refreshes[id], discordID)
	vm.mu.Unlock()
	vm.notify(id)
}

func channelExists(rooms []discord.Room, id string) bool {
	for _, r := range rooms {
		if r.ID == id {
			return true
		}
	}
	return false
}

func sameSet(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}

func copySet(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k := range m {
		out[k] = true
	}
	return out
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
