package game

import "github.com/google/uuid"

// life_floor.go — "damage that would reduce your life total to less than 1
// reduces it to 1 instead" (Angel's Grace, #3172; Sustaining Spirit, #5770).
//
// It is not a prevention and not a replacement of the damage amount: the
// 2021-03-19 ruling keeps the damage whole, so lifelink still gains the full
// amount and a "whenever you're dealt damage" trigger sees all of it. What
// changes is the life that damage costs (CR 120.3a), so it lives in the one
// place that decides that — applyResolvedDamageToPlayerLocked's life-loss
// step — as a cap, beside the "life total can't change" lock.
//
// TWO SOURCES, as life_lock.go has, and only one of them is stored:
//
//   - DERIVED — a permanent on the battlefield whose printed static is
//     the floor (Sustaining Spirit, #5770). Declared as
//     effects.Spec.PlayerLifeFloorOne, read through
//     CatalogPlayerLifeFloorOne on every query and written nowhere, so it
//     is its CONTROLLER's while it is on the battlefield, moves with a
//     change of control, ends when it leaves, and two copies compose.
//   - GRANTED — a resolved spell, "until end of turn" (Angel's Grace).
//     Stored as PlayerStatic.LifeFloorOne, because the spell is in a
//     graveyard a moment after it resolves.
//
// ONLY DAMAGE. "Damage that would reduce your life total" — a "lose N
// life" effect or a life payment takes the total below 1 as normal
// (Sustaining Spirit's 2008-10-01 ruling), so the one consumer is the
// damage tail; the CR 614 life window never asks this.

// CatalogPlayerLifeFloorOne reports whether a battlefield permanent with
// this catalog key gives its CONTROLLER the floor — true for Sustaining
// Spirit. Populated at init time from CardDef.PlayerLifeFloorOne (set by
// the cards/effects package from effects.Spec.PlayerLifeFloorOne). A nil
// hook (no catalog wired) means no permanent gives one.
var CatalogPlayerLifeFloorOne func(oracleID string) bool

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

// playerLifeFloorLocked reports whether a live life floor protects p:
// a permanent p controls whose printed static is the floor, first, then
// a stored grant whose duration has not run out.
func (g *Game) playerLifeFloorLocked(p *Player) bool {
	if p == nil {
		return false
	}
	if CatalogPlayerLifeFloorOne != nil && g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller != p.ID {
				continue
			}
			// CatalogAbilityKey, as the life lock reads it: the floor is a
			// static ability of the permanent, so one that has lost its
			// abilities (CR 613.1f) stops giving it.
			key := catalogAbilityKeyOf(c)
			if key != "" && CatalogPlayerLifeFloorOne(key) {
				return true
			}
		}
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
