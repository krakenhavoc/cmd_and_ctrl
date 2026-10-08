package lobby

// practice.go is the tutorial's practice table (ADR 0076 §2.2, #1078):
// a private game with one human seat and one `random`-tier bot, both
// on a fixed deck, created in one call and never listed.
//
// # What makes it different from a table made with POST /games
//
//   - It is created already STARTED, with the opening roll open (ADR
//     0125 §5.2): both seats roll, ties roll again, and the winner
//     chooses who goes first, exactly as at any table (ADR 0121, CR
//     103.1), so the player meets the roll they will meet at their
//     first real table. The tutorial's steps are ordered by the
//     player's first turn (ADR 0076 §2.1), so a bot that wins the roll
//     hands the first turn to the human (aiseat.SeatSpec.FirstTurnTo,
//     set in botStartLocked for this table only). CR 103.1 lets the
//     winner choose any player, so that is a legal choice, made through
//     the ordinary choose_starting_player move, not a rigged roll. A
//     human who wins chooses for themselves.
//   - It has no invite, so nobody else can join or watch it, and it is
//     not in GET /games, to the admin included.
//   - It is never written to the store and its room writes nothing to
//     disk (ws.RoomManager.CreateEphemeral). There is no resume
//     (§2.2): a restart, a Leave, or a new practice table ends it, and
//     a half-finished practice game is never brought back. Keeping no
//     rows also keeps it out of "My games" and out of any count of
//     games a person has finished — a tutorial is not a game played.
//   - It is reaped. Leaving abandons the table (POST
//     /games/{id}/practice/leave), but a closed tab cannot be relied on
//     to say so, so a table whose room has seen no commit for
//     PracticeLimits.Idle, or which has lived PracticeLimits.MaxAge, is
//     deleted and its bot runner stopped.
//   - One per person, and a ceiling on all of them. Opening a second
//     practice table abandons the caller's first; past
//     PracticeLimits.MaxTables the create is refused. Any session may
//     open one (that is what a tutorial is for), so these two rules
//     are what bound what a session can make the server do.

import (
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var (
	// ErrNotPracticeTable is returned by LeavePractice for a real
	// table: the practice route must never be a way to delete one.
	ErrNotPracticeTable = errors.New("lobby: not a practice table")
	// ErrPracticeTablesFull is returned by CreatePractice past
	// PracticeLimits.MaxTables.
	ErrPracticeTablesFull = errors.New("lobby: too many practice tables are open")
)

// PracticeLimits bounds the practice tables. A zero field means its
// default.
type PracticeLimits struct {
	// MaxTables is how many practice tables may be open at once,
	// across every caller. Default 8 — the whole intended user base
	// (AGENTS.md §2) taking the tutorial at the same moment.
	MaxTables int
	// Idle is how long a practice table may go without a single
	// commit before it is reaped. Default 20 minutes: the human holds
	// priority whenever the tutorial is waiting on them (it forces
	// autoPassPriority off), so a table nobody is at stops committing
	// entirely, while a player reading a coach card is a few seconds
	// from their next click.
	Idle time.Duration
	// MaxAge is the longest a practice table lives regardless. Default
	// two hours: the tutorial is five minutes, and a player who keeps
	// playing the practice game afterwards gets a generous evening.
	MaxAge time.Duration
}

const (
	defaultPracticeTables = 8
	defaultPracticeIdle   = 20 * time.Minute
	defaultPracticeMaxAge = 2 * time.Hour
)

func (p PracticeLimits) withDefaults() PracticeLimits {
	if p.MaxTables <= 0 {
		p.MaxTables = defaultPracticeTables
	}
	if p.Idle <= 0 {
		p.Idle = defaultPracticeIdle
	}
	if p.MaxAge <= 0 {
		p.MaxAge = defaultPracticeMaxAge
	}
	return p
}

// SetPracticeLimits replaces the practice-table bounds. Tests shrink
// them; production keeps the defaults.
func (l *Lobby) SetPracticeLimits(p PracticeLimits) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.practiceLimits = p
}

// PracticeTableName is the display name every practice table carries.
const PracticeTableName = "Practice game"

// PracticeHuman is the person a practice table is opened for.
type PracticeHuman struct {
	// Name is the seat label. A Discord identity's display name wins
	// over it, as on every other seat.
	Name     string
	Identity DiscordIdentity
	// UserID is the caller's users row, or uuid.Nil. It is carried on
	// the seat (and on the session the route mints) but never written
	// anywhere: the table has no rows.
	UserID uuid.UUID
	// Owner is who the table belongs to for the one-per-person rule.
	// "" opts out of it. See practiceOwner in practice_http.go.
	Owner    string
	DeckName string
	Cards    []game.Card
}

// PracticeBot is the practice table's bot seat.
type PracticeBot struct {
	Name     string
	Tier     string
	DeckID   string
	DeckName string
	Cards    []game.Card
}

// CreatePractice opens a practice table: it seats the human at seat 0
// and the bot at seat 1, starts the game with the opening roll open,
// and launches the bot's runner, which hands the first turn to the
// human if the bot wins the roll. It returns the table and the human's
// player ID.
//
// Any practice table the same Owner already has is abandoned first —
// there is no resume, so a second one replaces the first rather than
// joining it.
func (l *Lobby) CreatePractice(human PracticeHuman, bot PracticeBot) (GameMeta, uuid.UUID, error) {
	if human.Identity.Populated() {
		human.Name = human.Identity.DisplayName()
	}
	human.Name = trimToLimit(human.Name, 40)
	bot.Name = trimToLimit(bot.Name, 40)
	if human.Name == "" || bot.Name == "" {
		return GameMeta{}, uuid.Nil, ErrEmptyName
	}
	if len(human.Cards) == 0 || len(bot.Cards) == 0 {
		return GameMeta{}, uuid.Nil, ErrDeckNotUploaded
	}

	// Everything that must not run under l.mu — stopping a replaced
	// table's runner, closing its sockets, and starting this one's
	// runner — is collected here and run on the way out.
	var after []func()
	defer func() {
		for _, f := range after {
			f()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()
	limits := l.practiceLimits.withDefaults()

	// One per person: the caller's previous table goes first, so it
	// does not count against the ceiling it is about to free.
	open := 0
	for id, e := range l.games {
		if !e.practice {
			continue
		}
		if human.Owner != "" && e.practiceOwner == human.Owner {
			after = append(after, l.dropPracticeLocked(id, "replaced"))
			continue
		}
		open++
	}
	if open >= limits.MaxTables {
		return GameMeta{}, uuid.Nil, ErrPracticeTablesFull
	}

	g := game.NewGame()
	room := l.mgr.CreateEphemeral(g)
	var humanP, botP *game.Player
	_, _, err := room.ApplyExternal(func() error {
		var err error
		if humanP, err = g.AddPlayer(human.Name, human.Cards); err != nil {
			return err
		}
		// AddPlayer installs the deck; ReplaceDeck marks it as a real
		// one (DeckImported), as every other seated deck is.
		if err = g.ReplaceDeck(humanP.ID, human.Cards); err != nil {
			return err
		}
		if human.Identity.Populated() {
			if err = g.SetDiscordIdentity(humanP.ID, human.Identity.ID, human.Identity.AvatarHash, human.Identity.DisplayName()); err != nil {
				return err
			}
		}
		if botP, err = g.AddPlayer(bot.Name, bot.Cards); err != nil {
			return err
		}
		if err = g.SetBot(botP.ID, bot.Tier, bot.DeckID); err != nil {
			return err
		}
		if err = g.ReplaceDeck(botP.ID, bot.Cards); err != nil {
			return err
		}
		room.SetHost(humanP.ID)
		l.bindPlaymat(room, humanP.ID, human.UserID)
		// The opening roll, as at every table (ADR 0125 §5.2). Nothing
		// is shuffled or dealt until the winner chooses. See the file
		// comment.
		return g.StartWithOpeningRoll(nil)
	})
	if err != nil {
		l.mgr.Delete(g.ID)
		return GameMeta{}, uuid.Nil, err
	}

	humanSeat := SeatInfo{
		PlayerID:     humanP.ID,
		Name:         humanP.Name,
		Seat:         humanP.Seat,
		DeckName:     human.DeckName,
		DeckUploaded: true,
	}
	if human.Identity.Populated() {
		humanSeat.DiscordID = human.Identity.ID
		humanSeat.DiscordAvatarHash = human.Identity.AvatarHash
		humanSeat.DisplayName = human.Identity.DisplayName()
	}
	if human.UserID != uuid.Nil {
		humanSeat.UserID = human.UserID.String()
	}
	meta := GameMeta{
		ID:        g.ID,
		Name:      PracticeTableName,
		CreatedAt: g.CreatedAt,
		Players: []SeatInfo{humanSeat, {
			PlayerID:     botP.ID,
			Name:         botP.Name,
			Seat:         botP.Seat,
			DeckName:     bot.DeckName,
			DeckUploaded: true,
			IsBot:        true,
			BotTier:      bot.Tier,
			BotDeck:      bot.DeckID,
		}},
		State:    string(game.StateLobby),
		Practice: true,
	}
	entry := &gameEntry{
		meta:          meta,
		room:          room,
		stop:          make(chan struct{}),
		startedKnown:  true,
		practice:      true,
		practiceOwner: human.Owner,
		practiceHuman: humanP.ID,
	}
	l.games[g.ID] = entry
	l.syncStateLocked(entry)
	l.syncHostLocked(entry)
	l.startWatchLocked(entry)
	go l.watchPractice(entry, limits)
	if start := l.botStartLocked(entry); start != nil {
		after = append(after, start)
	}
	return copyMeta(entry.meta), humanP.ID, nil
}

// LeavePractice abandons a practice table: it is deleted, its bot
// runner stopped and its sockets closed. playerID must be the table's
// human seat. (An admin who wants one gone uses DELETE /games/{id},
// like any other table.)
//
// ErrGameNotFound for a table that is already gone, which the route
// treats as success — leaving is idempotent, and the tab-close path
// and the next page load may both ask. ErrNotPracticeTable for a real
// table, and ErrPlayerNotInGame for any other caller.
func (l *Lobby) LeavePractice(id, playerID uuid.UUID) error {
	l.mu.Lock()
	entry, ok := l.games[id]
	switch {
	case !ok:
		l.mu.Unlock()
		return ErrGameNotFound
	case !entry.practice:
		l.mu.Unlock()
		return ErrNotPracticeTable
	case playerID == uuid.Nil || playerID != entry.practiceHuman:
		l.mu.Unlock()
		return ErrPlayerNotInGame
	}
	after := l.dropPracticeLocked(id, "left")
	l.mu.Unlock()
	after()
	return nil
}

// dropPracticeLocked removes a practice table from the registry and
// the room manager, and returns the half that must run after l.mu is
// released: stopping its bot runner (which waits for the move in
// flight) and closing its sockets. Callers hold l.mu.
func (l *Lobby) dropPracticeLocked(id uuid.UUID, why string) func() {
	l.dropEntryLocked(id)
	l.mgr.Delete(id)
	l.dropReclaimsLocked(id)
	bots, evictor := l.bots, l.evictor
	slog.Default().Info("practice table closed", "game_id", id, "reason", why)
	return func() {
		if bots != nil {
			bots.StopBots(id)
		}
		if evictor != nil {
			evictor.EvictGame(id)
		}
	}
}

// watchPractice reaps entry when its room goes PracticeLimits.Idle
// without a commit, or once it has lived PracticeLimits.MaxAge. It
// rides Room.Subscribe, the same wake watchEnd and the bot runners
// use, and exits when the entry leaves the registry (stop is closed).
func (l *Lobby) watchPractice(entry *gameEntry, limits PracticeLimits) {
	wake, unsubscribe := entry.room.Subscribe()
	defer unsubscribe()
	idle := time.NewTimer(limits.Idle)
	defer idle.Stop()
	age := time.NewTimer(limits.MaxAge)
	defer age.Stop()
	for {
		var why string
		select {
		case <-entry.stop:
			return
		case <-wake:
			// Go 1.23+ timer channels are unbuffered: after Reset, a
			// receive never sees a value from the old deadline, so the
			// Stop-and-drain this once needed is gone (ADR 0122 §1.1).
			idle.Reset(limits.Idle)
			continue
		case <-idle.C:
			why = "idle"
		case <-age.C:
			why = "max_age"
		}
		l.mu.Lock()
		if l.games[entry.meta.ID] != entry {
			l.mu.Unlock()
			return
		}
		after := l.dropPracticeLocked(entry.meta.ID, why)
		l.mu.Unlock()
		after()
		return
	}
}

// PracticeOwner returns the owner key of the live practice table id,
// and false when id is not one. A session seated at a practice table
// that opens another belongs to the same person, so the route hands
// the new table this owner and the old one is replaced.
func (l *Lobby) PracticeOwner(id uuid.UUID) (string, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.games[id]
	if !ok || !e.practice {
		return "", false
	}
	return e.practiceOwner, true
}
