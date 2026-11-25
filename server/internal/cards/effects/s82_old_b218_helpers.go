package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b218_helpers.go — S82 OLD batch 218 (Rivals of Ixalan, #5531).

// RaidPay is "Raid — If you attacked this turn, you may pay <cost> rather than
// pay this spell's mana cost" (Admiral's Order): an alternative cost offered
// only to a player who has declared an attack this turn, checked at announce
// and in the view like every conditional offer (CR 118.9).
func RaidPay(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:      "raid",
		Label:    "Raid — pay " + cost + " rather than this spell's mana cost",
		ManaCost: cost,
		Condition: func(g *game.Game, controller uuid.UUID) bool {
			return B18AttackedThisTurn(g, controller)
		},
	}
}

// RaidCostsLess is "Raid — This spell costs {n} less to cast if you attacked
// this turn" (Siren Reaver), read as the cost is determined (CR 601.2f).
func RaidCostsLess(n int, label string) game.CostModifier {
	return CostsLess(n, label, func(q game.CostQuery) bool {
		return q.Game != nil && B18AttackedThisTurn(q.Game, q.Controller)
	})
}

// OldB218AttackedWithKeywordCreatures is Path of Mettle's trigger condition:
// the declaration that brings the player to at least two attacking creatures
// that have first strike, double strike, vigilance and/or haste (read through
// the layers). It fires once, on the declaration that reaches two
// (B16PlayerAttackedWithAtLeast's shape: the engine emits one EventAttack per
// creature, and the per-turn record of the label declines the rest).
func OldB218AttackedWithKeywordCreatures(ev game.Event, source *game.Card, g *game.Game, label string) bool {
	if ev.Kind != game.EventAttack || ev.Actor == uuid.Nil || ev.Actor != source.Controller {
		return false
	}
	g.RecomputeLayersIfStaleLocked()
	n := 0
	for _, id := range B13AttackingCreaturesYouControl(g, ev.Actor) {
		c, ok := g.LookupCardForEffect(id)
		if ok && OldB218HasFirstStrikeDoubleStrikeVigilanceOrHaste(c) {
			n++
		}
	}
	return n >= 2 && !B11TriggeredThisTurn(g, source.InstanceID, label)
}

// OldB218HasFirstStrikeDoubleStrikeVigilanceOrHaste is "has first strike,
// double strike, vigilance, or haste" on a card read through the layers.
func OldB218HasFirstStrikeDoubleStrikeVigilanceOrHaste(c game.Card) bool {
	for _, kw := range []string{"first strike", "double strike", "vigilance", "haste"} {
		if game.HasKeyword(&c, kw) {
			return true
		}
	}
	return false
}

// OldB218DistinctManaValues counts the different mana values among the cards
// exiled with `source` (the object, CR 607.2a): Azor's Gateway's "cards with
// five or more different mana values".
func OldB218DistinctManaValues(g *game.Game, source game.PermissionCardRef) int {
	seen := map[int]bool{}
	for _, c := range g.ExiledWithForEffect(source) {
		seen[c.ManaValue()] = true
	}
	return len(seen)
}
