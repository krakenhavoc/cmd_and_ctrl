package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rootwater Matriarch — Creature — Merfolk {2}{U}{U}, 2/3 (Tempest, S82 OLD b752,
// #5628):
//
//	"{T}: Gain control of target creature for as long as that creature is
//	 enchanted."
//
// A control change (layer 2, GainControl) whose duration is the new ForAsLongAs
// condition WhilePinnedIsEnchanted, pinned to the creature as it is when the ability
// resolves (CR 611.2b, 400.7): the effect begins only if the creature is enchanted
// then, and once the last Aura leaves it is over for good even if the creature is
// enchanted again. Any Aura counts, whoever controls it. The Matriarch leaving
// does not end it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d260526-db57-486d-b0bc-d523139a0148",
		Name:         "Rootwater Matriarch",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Gain control of target creature for as long as that creature is enchanted.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				d, ok := effects.OldB752DurationWhileItIsEnchanted(ctx, id)
				if !ok {
					return nil
				}
				return effects.GainControl{
					Target:     id,
					Controller: item.Controller,
					Duration:   d,
					Label:      "Rootwater Matriarch — gain control for as long as it is enchanted",
				}.Apply(ctx)
			},
		}},
	})
}
