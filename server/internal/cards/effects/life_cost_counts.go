package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// life_cost_counts.go — #1594, ADR 0020 Decision 47: the registered
// counts a computed life cost on an activated ability can name. Each
// is a package-level var so it is registered exactly once, at init,
// under a key a snapshot-restored game finds again; see
// game.LifeCount for why the cost carries a key and not a func.

// LifeEqualToCommanderColors is "Pay life equal to the number of
// colors in your commanders' color identity" (War Room). "Your
// commanders'" is every commander the activator owns — a partner
// pair's combined identity (CR 903.4) — and no commander, or a
// colourless one, is 0.
var LifeEqualToCommanderColors = game.LifeCount("commander-identity-colors",
	func(g *game.Game, activator, _ uuid.UUID) int {
		return g.CommanderIdentityColorCountForEffect(activator)
	})

// LifeHalfYoursRoundedUp is "Pay half your life, rounded up" (Murderous
// Betrayal, Lurking Evil). Read at announce, so the half is of the
// total the activator has then (CR 601.2f–g); at 1 life that is 1, and
// at 0 life it is 0, which CR 119.4 lets any player pay.
var LifeHalfYoursRoundedUp = game.LifeCount("half-your-life-rounded-up",
	func(g *game.Game, activator, _ uuid.UUID) int {
		p := g.PlayerByIDForEffect(activator)
		if p == nil || p.Life <= 0 {
			return 0
		}
		return (p.Life + 1) / 2
	})

// PayLifeCount is a life component whose amount is a registered count
// rather than a printed number — PayLifeCount(LifeEqualToCommanderColors).
// Compose it like PayLife: Plus(ManaCost("{3}"), TapCost(),
// PayLifeCount(LifeEqualToCommanderColors)).
func PayLifeCount(c game.LifeCostCount) game.AbilityCost {
	return game.AbilityCost{LifeFrom: c}
}
