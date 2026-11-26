package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// surge_keyword.go — surge (CR 702.117a, #6255): "You may pay [cost]
// rather than pay this spell's mana cost as you cast this spell if you or
// one of your teammates has cast another spell this turn."
//
// Spectacle's gated alternative cost (alternative_cost.go), with the gate
// reading the caster's per-turn cast tally (Game.CastTallyFor). A spell is
// counted once it has been cast, after it reaches the stack, and the gate
// is read at announce, before the spell being cast is counted, so "another
// spell" comes free: the first spell of a turn can't be surged, the second
// can. A countered spell was still cast and still counts; a copy was not
// cast and does not. A C2 table is Free-for-All (CR 903.2), with no
// teammates (CR 102.3), so "you or one of your teammates" is "you". Checked at announce and in the view, never
// again once the spell is on the stack.
//
// "If this spell's surge cost was paid" is SurgeCostWasPaid on a resolving
// spell; "if its surge cost was paid" on a permanent's trigger is
// CastWithSurge, read off its cast record (CR 400.7d).

// SurgeAltCostKey is the key Surge gives its offer, the value a cast
// record's AltCost holds when the surge cost was paid.
const SurgeAltCostKey = "surge"

// Surge is "Surge <cost>".
//
//	AlternativeCosts: []game.AlternativeCost{Surge("{1}{R}")}, // Goblin Freerunner
func Surge(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:       SurgeAltCostKey,
		Label:     "Surge " + cost,
		ManaCost:  cost,
		Condition: SurgeConditionMet,
	}
}

// SurgeConditionMet is surge's gate: the caster has cast another spell
// this turn.
func SurgeConditionMet(g *game.Game, controller uuid.UUID) bool {
	return g.CastTallyFor(controller).Total > 0
}

// SurgeCostWasPaid is "if this spell's surge cost was paid", read as the
// spell resolves: the alternative cost announced with the cast (CR
// 601.2b).
func SurgeCostWasPaid(ctx *Context) bool {
	return ctx.PaidAltCost(SurgeAltCostKey)
}

// CastWithSurge is "if its surge cost was paid" on a permanent: its cast
// provenance names the surge cost. False for a permanent that wasn't cast,
// or was cast another way.
func CastWithSurge(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.AltCost == SurgeAltCostKey
}
