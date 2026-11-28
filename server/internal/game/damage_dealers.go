package game

import "github.com/google/uuid"

// damage_dealers.go — "each creature that dealt damage to this creature this
// turn" (Giant Albatross, S82 OLD b827, #5635): the readers of
// TurnTally.DamagedObjects (dealt_damage_by.go) answer "did X damage Y", and
// this one lists the Xs for a given Y.

// DamageDealersToPermanentThisTurnForEffect lists the source objects that
// dealt damage to the permanent `victimID` this turn, each once, in the order
// each first dealt it. The victim is read as the permanent object it is now or
// most recently was this turn (PermanentRefForEffect), so a dies trigger
// finds the creature that has just died, and a card that left and came back
// is a new object nothing has damaged (CR 400.7). Combat and noncombat damage
// both count, prevented damage never does.
//
// Caller must hold g.mu.
func (g *Game) DamageDealersToPermanentThisTurnForEffect(victimID uuid.UUID) []ObjectRef {
	if len(g.TurnTally.DamagedObjects) == 0 {
		return nil
	}
	victim, ok := g.PermanentRefForEffect(victimID)
	if !ok {
		return nil
	}
	var out []ObjectRef
	seen := map[ObjectRef]bool{}
	for _, rec := range g.TurnTally.DamagedObjects {
		if rec.Victim != victim || seen[rec.Source] {
			continue
		}
		seen[rec.Source] = true
		out = append(out, rec.Source)
	}
	return out
}
