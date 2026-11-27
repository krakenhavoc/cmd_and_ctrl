package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recurring Nightmare — Enchantment {2}{B} (Exodus, S82 OLD batch 745, #5626):
//
//	"Sacrifice a creature, Return this enchantment to its owner's hand:
//	 Return target creature card from your graveyard to the battlefield.
//	 Activate only as a sorcery."
//
// Chthonian Nightmare's row without the energy and the mana value: the
// sacrifice and the return-this cost (#2028) are paid as the ability is
// announced (CR 602.2b), so it resolves with the enchantment already in its
// owner's hand, and the target is checked again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6708b11-1bcd-4208-a967-fe91f2e3313c",
		Name:         "Recurring Nightmare",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "Sacrifice a creature, Return this enchantment to its owner's hand: Return target creature card from your graveyard to the battlefield. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.SacrificeACreature(), effects.ReturnThis()),
			Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			SorcerySpeed: true,
			Purpose:      effects.ForTargets(effects.Reanimates(0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				effects.ReanimateSingleTarget(effects.NewContext(g, item), item.Controller)
				return nil
			},
		}},
	})
}
