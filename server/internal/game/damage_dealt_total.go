package game

import "github.com/google/uuid"

// damage_dealt_total.go — "if <this> has dealt N or more damage this turn"
// (Chandra, Fire of Kaladesh): the damage an object has dealt to anything this
// turn, players and permanents together.
//
// The two tallies it reads are the ones the engine already keeps
// (TurnTally.DamagedPlayers, damage_to_player_by.go; TurnTally.DamagedObjects,
// dealt_damage_by.go): each entry is the damage actually dealt after
// replacement and prevention, so prevented damage counts for nothing, and each
// names its source as an object (CR 400.7), so a source that left and came back
// has dealt nothing.

// DamageDealtThisTurnByObjectForEffect is the total damage the object `source`
// has dealt this turn, to players and to permanents, combat or not. The tally
// listener runs ahead of the trigger harvester, so a resolving ability already
// counts the damage it dealt. Caller must hold g.mu.
func (g *Game) DamageDealtThisTurnByObjectForEffect(source ObjectRef) int {
	if source.ID == uuid.Nil {
		return 0
	}
	total := 0
	for _, rec := range g.TurnTally.DamagedPlayers {
		if rec.Source == source {
			total += rec.Amount
		}
	}
	for _, rec := range g.TurnTally.DamagedObjects {
		if rec.Source == source {
			total += rec.Amount
		}
	}
	return total
}
