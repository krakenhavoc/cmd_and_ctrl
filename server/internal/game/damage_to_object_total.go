package game

import "github.com/google/uuid"

// TotalDamageDealtToObjectThisTurnForEffect is the damage dealt this turn to
// the permanent object `victimID`, from any source, combat or not, after
// replacement and prevention (Rushing-Tide Zubera: "if 4 or more damage was
// dealt to it this turn"). The victim is read as the object it is now or, once
// it has left the battlefield, the one it most recently was, so a dies trigger
// still finds the damage (CR 603.10a, 608.2h); a victim that came back is a new
// object (CR 400.7). Caller must hold g.mu.
func (g *Game) TotalDamageDealtToObjectThisTurnForEffect(victimID uuid.UUID) int {
	if len(g.TurnTally.DamagedObjects) == 0 {
		return 0
	}
	victim, ok := g.PermanentRefForEffect(victimID)
	if !ok {
		return 0
	}
	total := 0
	for _, rec := range g.TurnTally.DamagedObjects {
		if rec.Victim == victim {
			total += rec.Amount
		}
	}
	return total
}
