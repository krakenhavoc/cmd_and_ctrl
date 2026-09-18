package game

import (
	"time"

	"github.com/google/uuid"
)

// StartingLife is the Commander format's starting life total.
const StartingLife = 40

// CommanderDamageLethal is the single-commander damage total that wins
// the game via the commander damage rule (CR 903.10a).
const CommanderDamageLethal = 21

// MaxLifeHistoryEntries caps the per-player life-change log so a long
// game doesn't grow snapshots unboundedly. Older entries roll off
// the front when the cap is exceeded; the most recent
// MaxLifeHistoryEntries are always retained.
const MaxLifeHistoryEntries = 50

// DefaultMaxHandSize is the per-player hand-size cap enforced at
// cleanup (CR 402.2 — "the maximum hand size is normally seven").
// Cards like Reliquary Tower / Thought Vessel / Spellbook /
// Library of Leng / Null Profusion / Venser's Journal modify the
// per-player MaxHandSize via S14+ catalog cards plugged into the
// S16 layer system.
const DefaultMaxHandSize = 7

// NoMaxHandSize is the sentinel for "no maximum hand size" (the
// Reliquary Tower / Thought Vessel effect). Encoded as -1 so the
// integer field stays small and the discard-prompt logic can short-
// circuit on a single comparison.
const NoMaxHandSize = -1

// LifeChange is a single entry in a Player's life-change log. Delta
// is the change applied (positive for gain, negative for loss);
// NewTotal is the resulting life total after the change. At is the
// server timestamp of the mutation.
//
// S08 keeps the schema deliberately small. A future extension could
// carry a Reason ("commander damage", "burn spell", "manual") once
// the rules graft surfaces those distinctions; right now every
// change_life action is "manual" by definition.
//
// Seq is a per-player counter, stamped by ChangeLife, that starts at
// 1 and only ever goes up (#703). It is what the client keys the
// life-change popup on: entry COUNT is not usable for that, because
// the log is trimmed at MaxLifeHistoryEntries and stops growing, and
// At is not usable either, because it reaches the wire as RFC3339
// seconds, so two changes in the same second are indistinguishable.
// A watermark over Seq shows every delta a frame carries, not just
// the last, and keeps working past the cap.
//
// Derived from the newest surviving entry rather than from a separate
// Player counter, so it survives a snapshot restore and rewinds with
// an undo exactly like the log it lives on.
type LifeChange struct {
	Delta    int
	NewTotal int
	At       time.Time
	Seq      uint64
}

// Player is a seat at the table. Each player owns a set of private
// zones (library, hand, graveyard, command zone) and tracks their own
// life, poison, energy, and commander-damage bookkeeping.
//
// NOTE on commander damage: S25 (#77) rekeyed CommanderDamage from
// the opponent *player ID* to the commander *instance ID*. S02 chose
// the player key and every sprint since carried a note saying it was
// wrong; CR 903.14a is explicitly per-commander ("damage dealt to a
// player by ONE commander"), so a partner pair sharing a seat was
// pooling two commanders' damage into one 21-point clock and killing
// its controller early.
//
// The client had already been written against the correct shape:
// HoverZoomOverlay reads `commander_damage[instance_id]` off each
// seat to draw the per-commander bars, which meant every row read 0
// against the player-keyed map the server was actually sending. The
// rekey is what makes that UX display real numbers.
type Player struct {
	ID   uuid.UUID
	Name string
	Seat int // 0..MaxPlayers-1, assigned at AddPlayer time

	// Life, starting at StartingLife. Can go negative — death-by-life
	// is a state-based action evaluated in S13+ rules work.
	Life int

	// Poison counters (10 = loss). Energy is tracked but has no rules
	// effect until mechanics referencing it are implemented.
	Poison int
	Energy int

	// Private zones.
	Library   *Zone
	Hand      *Zone
	Graveyard *Zone
	Command   *Zone

	// CommanderDamage maps commander INSTANCE ID → total damage that
	// one commander has dealt to this player across the game
	// (CR 903.14a). See the note above the struct.
	//
	// The key is an instance ID and survives zone changes, which is
	// the behaviour the rule wants: a commander that dies, returns to
	// the command zone and is recast keeps accruing toward the same
	// 21. A commander that is exiled and returned as a NEW object
	// starts a fresh clock, which is also correct (CR 400.7) and
	// falls out of the new instance ID for free.
	CommanderDamage map[uuid.UUID]int

	// LifeHistory is the rolling log of every change to Life. Bounded
	// at MaxLifeHistoryEntries; older entries fall off the front. The
	// log is public — life is visible to all opponents in MTG, and the
	// history is a UX affordance, not hidden information.
	LifeHistory []LifeChange

	// Eliminated is set when the player concedes (S08) or, in the
	// future, loses to a state-based action (S13+ rules graft). An
	// eliminated player still occupies their seat for spectating; the
	// game's State transitions to StateEnded once exactly one
	// non-eliminated seat remains.
	Eliminated bool

	// HandKept is true once the player has committed to their opening
	// hand. False during the mulligan-decision window between Start
	// and the first turn-1 action. Mulligan resets it to false (the
	// player must commit again after redrawing); KeepHand sets it to
	// true. Game.MulligansOpen flips false once every seated player
	// has KeptHand. Added in S08.
	HandKept bool

	// MulligansTaken is the count of Mulligan calls this player has
	// made in the current opening-hand window. Reset at Start. Used
	// by the UI to show "mulligans taken: N" and is the seed for a
	// future London-style bottom-N penalty (S13+ if it lands).
	MulligansTaken int

	// DeckImported is true once the player has had a real deck
	// installed via ReplaceDeck. False after AddPlayer (which only
	// installs a 1-card placeholder commander so the seat is real).
	// The in-game deck-import modal in Game.svelte (S08.5 wave 1)
	// uses this to distinguish "fresh invitee with placeholder" from
	// "imported and ready to play" — a card-count check is unreliable
	// because the placeholder also produces non-zero library counts.
	DeckImported bool

	// UndosRemaining is how many undos this player can still spend in
	// the current turn. Refreshed to Game.UndoLimit on entering this
	// player's untap step. Decremented per successful undo. Added in
	// S11 alongside the per-caller undo gate.
	UndosRemaining int

	// Discord identity metadata (S12.5). Populated on AddPlayer when
	// the seat is claimed via the OAuth flow; zero values for seats
	// claimed via the manual-name form. Exposed through PlayerView so
	// the client can render the avatar + display name without holding
	// the lobby SeatInfo.
	DiscordID         string
	DiscordAvatarHash string
	DisplayName       string

	// IsBot marks a seat driven by an aiseat runner rather than a
	// WebSocket client; BotTier names its policy tier ("random",
	// "heuristic", …) and BotDeck the curated deck it was seated
	// with (empty for a raw decklist). Set by Game.SetBot at seat
	// time and exposed through PlayerView so the client can render
	// the BOT chip. Carried in the snapshot, which is what makes a
	// bot seat survive a deploy: the lobby's restore path reads
	// these back and relaunches the runner. Added in S31 sub-PR 4.
	IsBot   bool
	BotTier string
	BotDeck string

	// AttemptedEmptyDraw records the one fact CR 704.5b reads: this
	// player has attempted to draw a card from an empty library since
	// the last state-based-action check. The SBA loop eliminates a
	// player with it set. Set by actuallyDrawCardLocked (every draw
	// path: the draw step, the DrawCard action, draw effects) when
	// PopTop returns ErrZoneEmpty; cleared on elimination. A mill,
	// an "exile the top N" or any other run off the bottom of the
	// library never sets it (CR 701.17b, #767).
	//
	// Named LosesAtNextSBA until ADR 0057 sub-PR 1; the JSON tag
	// keeps the old name so snapshots need no schema bump. One other
	// writer remains until ADR 0057 sub-PR 2 makes an effect loss
	// immediate: LoseTheGameForEffect (the Pact cycle) borrows the
	// flag to defer its loss to the next SBA check. Added in S13.1.
	AttemptedEmptyDraw bool

	// CommanderCasts tracks the per-commander cast count from the
	// command zone for the Commander tax (CR 903.8 — each cast costs
	// {2} more for every previous cast). Keyed by the commander's
	// instance ID so partner-pair players get independent counters
	// (the pre-S13.1 per-opponent CommanderDamage map collapsed
	// partners; this fixes that for the cast-tax half — the damage
	// half is sub-PR-deferred until the partner-pair playtest demands
	// it). Surfaced on the wire so clients can render "+0 / +2 / +4"
	// next to the commander tile. Sandbox: the engine doesn't enforce
	// the mana cost — players track mana on paper / in their head.
	// Added in S13.1.
	CommanderCasts map[uuid.UUID]int

	// Counters is the per-player named counter map (S13.2 — poison,
	// energy, experience, rad, plus any homebrew). Distinct from the
	// per-card Card.Counters map. The legacy single-int Player.Poison
	// and Player.Energy fields are kept in sync via SetPoison /
	// SetEnergy / AddPlayerCounter so the older actions and the SBA
	// loop see consistent values; new code should prefer Counters.
	// Zero-valued entries are removed to keep the map sparse on the
	// wire.
	Counters map[string]int

	// MaxHandSize is the per-player cleanup-step hand-size cap
	// (CR 402.2). Default DefaultMaxHandSize (7); NoMaxHandSize (-1)
	// disables the cap (Reliquary Tower / Thought Vessel). Set via
	// the set_max_hand_size action; effect-catalog work in S14+
	// will write to it via the S16 layer pipeline. Added in S13.4.
	MaxHandSize int

	// LandDropsPerTurn is the player's BASE land-play allowance
	// (CR 305.2). DefaultLandDropsPerTurn (1) on every freshly seated
	// player. Permanents that grant additional land plays are NOT
	// written here — they are derived from the battlefield, so two of
	// them compose and one leaving doesn't strand the other's grant;
	// see land_drops.go. Added in #500.
	//
	// A pre-#500 snapshot has no value for this, and 0 would restore
	// a table where nobody may ever play a land, so restorePlayer
	// treats a non-positive restored value as the default.
	LandDropsPerTurn int

	// ManaPool holds the player's currently-floating mana tokens.
	// Slice (not multiset) to preserve insertion order for the S15
	// auto-tapper preview + S17+ filter-land sub-payment routing.
	// Cleared at every step boundary by the step-change hook
	// (CR 106.4). Added in S15 sub-PR 2.
	ManaPool ManaPool
}

// newPlayer constructs a player with empty zones and their starting
// life total. The library is populated and shuffled by the Game when
// the game starts.
func newPlayer(name string, seat int) *Player {
	id := uuid.New()
	p := &Player{
		ID:               id,
		Name:             name,
		Seat:             seat,
		Life:             StartingLife,
		CommanderDamage:  make(map[uuid.UUID]int),
		CommanderCasts:   make(map[uuid.UUID]int),
		MaxHandSize:      DefaultMaxHandSize,
		LandDropsPerTurn: DefaultLandDropsPerTurn,
	}
	p.Library = newZone(ZoneLibrary, id)
	p.Hand = newZone(ZoneHand, id)
	p.Graveyard = newZone(ZoneGraveyard, id)
	p.Command = newZone(ZoneCommand, id)
	return p
}

// ChangeLife mutates life total by delta (positive for gain, negative
// for loss), records a LifeHistory entry stamped at time.Now().UTC()
// and with the next per-player Seq, and returns the new total. A delta
// of 0 is a no-op — no history entry is recorded so the log isn't
// polluted by accidental clicks.
func (p *Player) ChangeLife(delta int) int {
	if delta == 0 {
		return p.Life
	}
	p.Life += delta
	p.LifeHistory = append(p.LifeHistory, LifeChange{
		Delta:    delta,
		NewTotal: p.Life,
		At:       time.Now().UTC(),
		Seq:      p.nextLifeSeq(),
	})
	if len(p.LifeHistory) > MaxLifeHistoryEntries {
		// Trim by copying the tail forward so the underlying array
		// doesn't grow without bound across a long game.
		dropped := len(p.LifeHistory) - MaxLifeHistoryEntries
		p.LifeHistory = append(p.LifeHistory[:0], p.LifeHistory[dropped:]...)
	}
	return p.Life
}

// nextLifeSeq is the Seq the next LifeHistory entry gets: one past the
// newest entry still in the log. Trimming only ever drops from the
// FRONT, so the newest entry is always present and the counter keeps
// climbing after the log stops growing. An empty log starts at 1; zero
// is the un-stamped sentinel, which is also what a pre-#703 snapshot's
// entries decode as (they then all read as "older than anything new",
// which is exactly right).
func (p *Player) nextLifeSeq() uint64 {
	if n := len(p.LifeHistory); n > 0 {
		return p.LifeHistory[n-1].Seq + 1
	}
	return 1
}

// RecordCommanderDamage adds damage dealt to this player by one
// commander, identified by its card INSTANCE ID. Cumulative across
// the game; returns the new running total for that commander.
func (p *Player) RecordCommanderDamage(fromCommander uuid.UUID, amount int) int {
	if amount < 0 {
		amount = 0
	}
	p.CommanderDamage[fromCommander] += amount
	return p.CommanderDamage[fromCommander]
}

// IsDeadByCommanderDamage reports whether any SINGLE commander has
// dealt 21 or more damage to this player (CR 903.14a). Totals are
// never summed across commanders — two partners at 15 apiece is 30
// damage and not a loss.
func (p *Player) IsDeadByCommanderDamage() bool {
	for _, d := range p.CommanderDamage {
		if d >= CommanderDamageLethal {
			return true
		}
	}
	return false
}
