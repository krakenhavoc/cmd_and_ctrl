package game

import "github.com/google/uuid"

// life_floor.go — "damage that would reduce your life total to less than 1
// reduces it to 1 instead" (Angel's Grace, #3172).
//
// It is not a prevention and not a replacement of the damage amount: the
// 2021-03-19 ruling keeps the damage whole, so lifelink still gains the full
// amount and a "whenever you're dealt damage" trigger sees all of it. What
// changes is the life that damage costs (CR 120.3a), so it lives in the one
// place that decides that — applyResolvedDamageToPlayerLocked's life-loss
// step — as a cap, beside the "life total can't change" lock.

// GrantLifeFloorForEffect gives `player` the floor for duration d (until
// end of turn when d is zero). Caller must hold g.mu.
func (g *Game) GrantLifeFloorForEffect(player uuid.UUID, label string, source uuid.UUID, d Duration) {
	p := g.playerByIDLocked(player)
	if p == nil {
		return
	}
	if d.IsZero() {
		d = g.UntilEndOfTurnDuration()
	}
	p.Statics = append(p.Statics, PlayerStatic{
		LifeFloorOne: true,
		Source:       source,
		Label:        label,
		Duration:     d,
	})
}

// playerLifeFloorLocked reports whether a live life floor protects p.
func (g *Game) playerLifeFloorLocked(p *Player) bool {
	if p == nil {
		return false
	}
	for _, s := range p.Statics {
		if s.LifeFloorOne && !g.durationExpiredLocked(s.Duration, false) {
			return true
		}
	}
	return false
}

// capLifeLossByFloorLocked is the life that `lifeLoss` of damage costs p
// under a live floor: never more than takes p down to 1. A player already
// below 1 is left alone (the floor reduces damage that would take a total
// TO less than 1; it does not raise one).
func (g *Game) capLifeLossByFloorLocked(p *Player, lifeLoss int) int {
	if lifeLoss <= 0 || p.Life < 1 || !g.playerLifeFloorLocked(p) {
		return lifeLoss
	}
	return min(lifeLoss, p.Life-1)
}
