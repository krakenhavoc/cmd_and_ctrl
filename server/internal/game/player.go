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

// LifeChange is a single entry in a Player's life-change log. Delta
// is the change applied (positive for gain, negative for loss);
// NewTotal is the resulting life total after the change. At is the
// server timestamp of the mutation.
//
// S08 keeps the schema deliberately small. A future extension could
// carry a Reason ("commander damage", "burn spell", "manual") once
// the rules graft surfaces those distinctions; right now every
// change_life action is "manual" by definition.
type LifeChange struct {
	Delta    int
	NewTotal int
	At       time.Time
}

// Player is a seat at the table. Each player owns a set of private
// zones (library, hand, graveyard, command zone) and tracks their own
// life, poison, energy, and commander-damage bookkeeping.
//
// NOTE on commander damage: S02 keys CommanderDamage by the opponent
// *player ID*, not by commander *instance ID*. This is incorrect for
// partner commanders (two commanders per player), where damage should
// be tracked per commander to evaluate the 21-damage lethal condition.
// Good enough for S02; S10's Commander UX sprint is the natural place
// to broaden this to a per-commander map.
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

	// CommanderDamage maps opposing player ID → damage dealt by that
	// opponent's commander(s). See the package-level note above.
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
}

// newPlayer constructs a player with empty zones and their starting
// life total. The library is populated and shuffled by the Game when
// the game starts.
func newPlayer(name string, seat int) *Player {
	id := uuid.New()
	p := &Player{
		ID:              id,
		Name:            name,
		Seat:            seat,
		Life:            StartingLife,
		CommanderDamage: make(map[uuid.UUID]int),
	}
	p.Library = newZone(ZoneLibrary, id)
	p.Hand = newZone(ZoneHand, id)
	p.Graveyard = newZone(ZoneGraveyard, id)
	p.Command = newZone(ZoneCommand, id)
	return p
}

// ChangeLife mutates life total by delta (positive for gain, negative
// for loss), records a LifeHistory entry stamped at time.Now().UTC(),
// and returns the new total. A delta of 0 is a no-op — no history
// entry is recorded so the log isn't polluted by accidental clicks.
func (p *Player) ChangeLife(delta int) int {
	if delta == 0 {
		return p.Life
	}
	p.Life += delta
	p.LifeHistory = append(p.LifeHistory, LifeChange{
		Delta:    delta,
		NewTotal: p.Life,
		At:       time.Now().UTC(),
	})
	if len(p.LifeHistory) > MaxLifeHistoryEntries {
		// Trim by copying the tail forward so the underlying array
		// doesn't grow without bound across a long game.
		dropped := len(p.LifeHistory) - MaxLifeHistoryEntries
		p.LifeHistory = append(p.LifeHistory[:0], p.LifeHistory[dropped:]...)
	}
	return p.Life
}

// RecordCommanderDamage adds damage dealt to this player by a given
// opponent's commander. Idempotent and cumulative across a game.
func (p *Player) RecordCommanderDamage(fromOpponent uuid.UUID, amount int) int {
	if amount < 0 {
		amount = 0
	}
	p.CommanderDamage[fromOpponent] += amount
	return p.CommanderDamage[fromOpponent]
}

// IsDeadByCommanderDamage reports whether any single opponent's
// commander has dealt 21 or more damage to this player.
func (p *Player) IsDeadByCommanderDamage() bool {
	for _, d := range p.CommanderDamage {
		if d >= CommanderDamageLethal {
			return true
		}
	}
	return false
}
