package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suppression Field — Enchantment {1}{W} (Ravnica: City of Guilds):
//
//	"Activated abilities cost {2} more to activate unless they're mana abilities."
//
// An activation-scoped cost increase (CR 601.2f, 602.2b) over every activated
// ability on the table, whoever controls it and from whatever zone it works
// (cycling, a graveyard activation, a loyalty ability), except a CR 605 mana
// ability, which the query marks. It reaches a cost with no mana in it ({T}
// alone, a sacrifice), which then owes the {2}. Two Fields tax {4}.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63439f19-01e5-4ae1-9841-90493d62d0c8",
		Name:         "Suppression Field",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.ActivationCostsMore(2, "Activated abilities cost {2} more to activate unless they're mana abilities.",
				func(q game.CostQuery) bool { return q.Ability != nil && !q.Ability.Mana }),
		},
	})
}
