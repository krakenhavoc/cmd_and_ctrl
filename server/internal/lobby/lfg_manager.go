package lobby

// lfg_manager.go: ADR 0161 §3–§6, the part of a table's #lfg post that
// talks to Discord. It posts a table the host opted in, edits the post
// as seats fill and free, the invite rotates or the settings change,
// deletes it when the game starts, the host withdraws it or the table
// goes, and on boot keeps the posts of tables still waiting and deletes
// the rest.
//
// It is a reconciler with the voice manager's shape (ADR 0159 §4). The
// lobby says only "table X changed" (notify); the worker reads what the
// table wants now (Lobby.lfgTable), compares it with the post it last
// sent, and posts, edits or deletes. A missed or repeated notify costs
// nothing.
//
// Locking is the voice manager's. The worker never holds l.mu across a
// Discord call: it reads the table under l.mu, lets go, calls Discord,
// then reports back under l.mu. A table's lfgPostState is touched only
// on the worker goroutine (reconcile, boot), so it needs no lock. m.mu
// guards only what other goroutines reach (dirty, resets) and is never
// held while calling into the lobby, so the order is always l.mu before
// m.mu and never the reverse.

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// The ADR's numbers (§3, §4, §5). Constants, not settings.
const (
	// lfgEditEvery coalesces a post's edits to one per ten seconds,
	// under Discord's per-channel rate limit.
	lfgEditEvery = 10 * time.Second
	// lfgPingCooldown is how long a host waits between two pings of the
	// LFG role. A post inside it goes up without the ping.
	lfgPingCooldown = 30 * time.Minute
	// lfgMaxAttempts is how many transient failures in a row a post
	// tolerates before it shows "unavailable".
	lfgMaxAttempts = 3
	// lfgCallTimeout bounds one Discord call.
	lfgCallTimeout = 10 * time.Second
	// lfgMaxRetryWait caps a 429's wait.
	lfgMaxRetryWait = 30 * time.Second
	// lfgSweepEvery is how often the worker reads every opted-in table
	// again. Table settings change over the game socket, which tells
	// the lobby nothing; a sweep that finds no change makes no call.
	lfgSweepEvery = 30 * time.Second
)

// lfgAPI is the Discord half the manager uses: discord.LFG, or a fake
// in tests.
type lfgAPI interface {
	Post(ctx context.Context, client *http.Client, p discord.LFGPost, ping bool) (string, error)
	Edit(ctx context.Context, client *http.Client, messageID string, p discord.LFGPost) error
	Delete(ctx context.Context, client *http.Client, messageID string) error
}

// lfgTable is what one table wants, read under l.mu.
type lfgTable struct {
	// enabled is the host's choice (games.lfg).
	enabled bool
	// eligible is whether the table can be posted now: in the lobby,
	// not archived, its host signed in with Discord.
	eligible bool
	hostID   string
	// inviteToken is the player invite's plaintext, "" when this
	// process does not hold it (lost to a restart).
	inviteToken string
	// messageID is the persisted post, games.lfg_message_id.
	messageID string
	// post is the message, less its join URL, which the manager builds.
	post discord.LFGPost
}

// lfgPostState is the manager's own record of one table's post.
type lfgPostState struct {
	messageID string
	// sent is the post as last sent; nil when not known (adopted at
	// boot), which forces one edit.
	sent *discord.LFGPost
	// joinURL is the link the post's button carries, when known. After
	// a restart the lobby has no invite plaintext, but the button still
	// carries a working link; an edit then leaves it as it is.
	joinURL string
	// lastEdit and editDue coalesce edits (lfgEditEvery).
	lastEdit, editDue time.Time
	// lost is set when Discord said the post was deleted: the repost
	// does not ping (ADR 0161 §6).
	lost bool
	// unavailable is set after a failure that will not fix itself, or
	// lfgMaxAttempts transient ones. The host's toggle clears it.
	unavailable bool
	attempts    int
	// forbiddenLogged keeps a 403's log line to once per table.
	forbiddenLogged bool
}

// LFGManager runs the #lfg posts of one lobby. Construct it with
// NewLFGManager, attach it with Lobby.SetLFGManager before
// RestoreFromDisk, and Start it after.
type LFGManager struct {
	api    lfgAPI
	client *http.Client
	log    *slog.Logger
	now    func() time.Time
	// base is the public origin the join link is built against
	// (CMDCTRL_PUBLIC_BASE_URL), the invite DM's.
	base  string
	lobby *Lobby

	// posts and pinged are the worker's alone.
	posts map[uuid.UUID]*lfgPostState
	// pinged is when each host's last post pinged the role, by Discord
	// id (lfgPingCooldown). In memory: a restart forgets it, which at
	// worst allows one more ping.
	pinged map[string]time.Time

	mu    sync.Mutex
	dirty map[uuid.UUID]bool
	// resets are the tables whose host just changed their choice: the
	// worker forgets an unavailable post and tries again.
	resets map[uuid.UUID]bool
	kick   chan struct{}
}

// NewLFGManager builds a manager over the configured #lfg calls (ADR
// 0161 slice 2). base is the origin join links are built against.
func NewLFGManager(f discord.LFG, base string, log *slog.Logger) *LFGManager {
	return newLFGManager(f, &http.Client{Timeout: lfgCallTimeout}, base, log, time.Now)
}

func newLFGManager(api lfgAPI, client *http.Client, base string, log *slog.Logger, now func() time.Time) *LFGManager {
	if log == nil {
		log = slog.Default()
	}
	return &LFGManager{
		api:    api,
		client: client,
		log:    log,
		now:    now,
		base:   base,
		posts:  map[uuid.UUID]*lfgPostState{},
		pinged: map[string]time.Time{},
		dirty:  map[uuid.UUID]bool{},
		resets: map[uuid.UUID]bool{},
		kick:   make(chan struct{}, 1),
	}
}

// notify marks a table for reconciling and wakes the worker. Never
// blocks; safe under l.mu.
func (m *LFGManager) notify(id uuid.UUID) {
	m.mu.Lock()
	m.dirty[id] = true
	m.mu.Unlock()
	select {
	case m.kick <- struct{}{}:
	default:
	}
}

// toggled is the host's choice changing: the next reconcile starts the
// table's post afresh. Safe under l.mu; SetLFG's persist notifies.
func (m *LFGManager) toggled(id uuid.UUID) {
	m.mu.Lock()
	m.resets[id] = true
	m.mu.Unlock()
}

// Start runs the boot reconcile and then the worker until ctx ends.
func (m *LFGManager) Start(ctx context.Context) {
	go func() {
		m.boot(ctx)
		m.run(ctx)
	}()
}

func (m *LFGManager) run(ctx context.Context) {
	ticker := time.NewTicker(lfgSweepEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.kick:
			m.drain(ctx)
		case <-ticker.C:
			for _, id := range m.lobby.lfgTableIDs() {
				m.notify(id)
			}
			m.drain(ctx)
		}
	}
}

// drain reconciles every table marked dirty, in a stable order.
func (m *LFGManager) drain(ctx context.Context) {
	for {
		m.mu.Lock()
		ids := make([]uuid.UUID, 0, len(m.dirty))
		for id := range m.dirty {
			ids = append(ids, id)
		}
		m.dirty = map[uuid.UUID]bool{}
		m.mu.Unlock()
		if len(ids) == 0 {
			return
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
		for _, id := range ids {
			if ctx.Err() != nil {
				return
			}
			m.reconcile(ctx, id)
		}
	}
}

// state returns the manager's record of a table, creating it. Worker
// only.
func (m *LFGManager) state(id uuid.UUID) *lfgPostState {
	st, ok := m.posts[id]
	if !ok {
		st = &lfgPostState{}
		m.posts[id] = st
	}
	return st
}

func (m *LFGManager) call(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, lfgCallTimeout)
}

// reconcile closes the gap between what table id wants and the post
// that exists. Only the worker calls it.
func (m *LFGManager) reconcile(ctx context.Context, id uuid.UUID) {
	t, ok := m.lobby.lfgTable(id)
	st := m.state(id)
	now := m.now()

	m.mu.Lock()
	reset := m.resets[id]
	delete(m.resets, id)
	m.mu.Unlock()
	if reset || !ok || !t.enabled {
		st.unavailable, st.attempts, st.lost = false, 0, false
	}

	if !ok || !t.enabled || !t.eligible {
		if st.messageID != "" {
			cctx, cancel := m.call(ctx)
			err := m.api.Delete(cctx, m.client, st.messageID)
			cancel()
			if err != nil {
				m.failed(id, st, "delete", err)
				return
			}
			st.messageID, st.sent, st.joinURL = "", nil, ""
			m.lobby.setLFGMessage(id, "")
		}
		if ok {
			m.publish(id, t, st)
		}
		if !ok || !t.enabled {
			// Nothing left to remember: every lobby commit notifies,
			// and most tables are never posted.
			delete(m.posts, id)
		}
		return
	}
	if st.unavailable {
		m.publish(id, t, st)
		return
	}

	post := t.post
	if t.inviteToken != "" {
		post.JoinURL = inviteURL(m.base, id, t.inviteToken)
	} else {
		post.JoinURL = st.joinURL
	}

	switch {
	case st.messageID == "":
		if post.JoinURL == "" {
			// Nothing to post: the link was lost to a restart and the
			// old post went with it. SetLFG refuses a fresh "on" in
			// this state; this is the post deleted on Discord since.
			m.log.Warn("lfg post cannot be made: this process does not hold the table's invite link", "game_id", id)
			st.unavailable = true
			m.publish(id, t, st)
			return
		}
		ping := !st.lost && m.mayPing(t.hostID, now)
		cctx, cancel := m.call(ctx)
		msg, err := m.api.Post(cctx, m.client, post, ping)
		cancel()
		if err != nil {
			m.failed(id, st, "post", err)
			return
		}
		if ping {
			m.pinged[t.hostID] = now
		}
		st.messageID, st.joinURL, st.lost = msg, post.JoinURL, false
		st.sent = &post
		st.lastEdit, st.editDue = now, time.Time{}
		m.lobby.setLFGMessage(id, msg)
	case st.sent == nil || !reflect.DeepEqual(*st.sent, post):
		if wait := st.lastEdit.Add(lfgEditEvery).Sub(now); wait > 0 {
			// Coalesce: one timer per post, and whatever the table looks
			// like when it fires is what the edit sends.
			if st.editDue.IsZero() || !st.editDue.After(now) {
				st.editDue = now.Add(wait)
				time.AfterFunc(wait, func() { m.notify(id) })
			}
			m.publish(id, t, st)
			return
		}
		cctx, cancel := m.call(ctx)
		err := m.api.Edit(cctx, m.client, st.messageID, post)
		cancel()
		if errors.Is(err, discord.ErrLFGMessageGone) {
			// Deleted on Discord. Forget it; the next pass posts afresh,
			// without the ping.
			st.messageID, st.sent, st.lost = "", nil, true
			m.lobby.setLFGMessage(id, "")
			m.notify(id)
			return
		}
		if err != nil {
			m.failed(id, st, "edit", err)
			return
		}
		st.sent = &post
		if post.JoinURL != "" {
			st.joinURL = post.JoinURL
		}
		st.lastEdit, st.editDue = now, time.Time{}
	}
	st.attempts = 0
	m.publish(id, t, st)
}

// mayPing reports whether host's new post may ping the role: once per
// lfgPingCooldown. Worker only.
func (m *LFGManager) mayPing(host string, now time.Time) bool {
	last, ok := m.pinged[host]
	return !ok || !now.Before(last.Add(lfgPingCooldown))
}

// failed records a failed Discord call. A transient one is retried
// after a wait, up to lfgMaxAttempts in a row; anything else, or one too
// many, shows "unavailable" until the host's toggle. The table is never
// affected.
func (m *LFGManager) failed(id uuid.UUID, st *lfgPostState, step string, err error) {
	st.attempts++
	if discord.Transient(err) && st.attempts < lfgMaxAttempts {
		wait := time.Duration(st.attempts) * 5 * time.Second
		var rl *discord.RateLimitError
		if errors.As(err, &rl) && rl.RetryAfter > wait {
			wait = min(rl.RetryAfter, lfgMaxRetryWait)
		}
		m.log.Info("lfg post call failed; retrying", "game_id", id, "step", step, "err", err, "wait", wait)
		time.AfterFunc(wait, func() { m.notify(id) })
		return
	}
	switch {
	case errors.Is(err, discord.ErrLFGForbidden):
		if !st.forbiddenLogged {
			st.forbiddenLogged = true
			m.log.Warn("lfg post unavailable: the bot cannot post in the channel", "game_id", id, "step", step,
				"channel_var", discord.LFGChannelEnv, "err", err)
		}
	default:
		m.log.Warn("lfg post unavailable", "game_id", id, "step", step, "err", err)
	}
	st.unavailable = true
	if t, ok := m.lobby.lfgTable(id); ok {
		m.publish(id, t, st)
	}
}

// publish tells the lobby the post's status.
func (m *LFGManager) publish(id uuid.UUID, t lfgTable, st *lfgPostState) {
	status := ""
	switch {
	case !t.enabled:
	case st.unavailable:
		status = LFGStatusUnavailable
	case st.messageID != "":
		status = LFGStatusPosted
	default:
		status = LFGStatusPending
	}
	m.lobby.reportLFG(id, status)
}

// boot is ADR 0161 §6's reconcile after a restart. A resident table's
// post is adopted and re-rendered by the normal pass, which deletes it
// if the table no longer wants it. A post whose games row did not come
// back is deleted and its column cleared. Nothing else in the channel
// is touched: #lfg is shared, and the manager lists no messages.
func (m *LFGManager) boot(ctx context.Context) {
	for id, msg := range m.lobby.lfgMessages() {
		st := m.state(id)
		st.messageID, st.sent = msg, nil
		m.notify(id)
	}
	for _, id := range m.lobby.lfgTableIDs() {
		m.notify(id)
	}
	for id, msg := range m.lobby.lfgOrphans(m.log) {
		cctx, cancel := m.call(ctx)
		err := m.api.Delete(cctx, m.client, msg)
		cancel()
		if err != nil {
			m.log.Warn("lfg: deleting the post of an unrestored table failed; it stays until the next restart",
				"game_id", id, "err", err)
			continue
		}
		m.lobby.setLFGMessage(id, "")
	}
	m.drain(ctx)
}
