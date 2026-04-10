package game

import "github.com/google/uuid"

// StartingLife is the Commander format's starting life total.
const StartingLife = 40

// CommanderDamageLethal is the single-commander damage total that wins
// the game via the commander damage rule (CR 903.10a).
const CommanderDamageLethal = 21

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
// for loss) and returns the new total.
func (p *Player) ChangeLife(delta int) int {
	p.Life += delta
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
