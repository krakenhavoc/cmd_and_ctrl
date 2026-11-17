package effects

import (
	"github.com/google/uuid"

	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_dmu_b02_helpers.go — helpers for Dominaria United batch 2 (#5487).

// dmuB02OpponentCastTwoOrMoreSpells passes while some opponent of the caster
// has cast two or more spells this turn (Ertai's Scorn). The per-turn tally
// is bumped after a spell reaches the stack and a cost is priced before
// that, so the spell being priced is never counted.
func dmuB02OpponentCastTwoOrMoreSpells() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		for _, p := range q.Game.Seats {
			if p == nil || p.Eliminated || p.ID == q.Controller {
				continue
			}
			if q.Game.CastTallyFor(p.ID).Total >= 2 {
				return true
			}
		}
		return false
	}
}

// dmuB02CostsLessMana is "This spell costs <unit> less to cast if <when>"
// for a reduction that is not plain generic mana: each coloured symbol
// removes one mana of its colour, or one generic when the cost has none of
// it left, and the cost never goes below {0} (CR 118.7, CR 601.2f).
func dmuB02CostsLessMana(unit, label string, when ...CostPredicate) game.CostModifier {
	u, err := game.ParseCost(unit)
	if err != nil || u.XSlots > 0 || u.ManaValue() == 0 {
		panic(fmt.Sprintf("effects.dmuB02CostsLessMana: %q is not a fixed mana amount (%v)", unit, err))
	}
	m := CostsLess(1, label, when...)
	m.Unit = &u
	return m
}

// dmuB02TotalPowerAndToughnessAtMost passes when the creature's current
// power plus current toughness is at most n (Cut Down).
func dmuB02TotalPowerAndToughnessAtMost(n int) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.CurrentPower()+c.CurrentToughness() <= n
	}
}
