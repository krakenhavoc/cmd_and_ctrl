package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch's Mist — Enchantment {2}{B} (Future Sight, #5594):
//
//	"{2}{B}, {T}: Destroy target creature that was dealt damage this
//	 turn."
//
// Fatal Blow's target clause (DealtDamageThisTurn) on a repeatable
// enchantment, without Fatal Blow's "can't be regenerated": a
// regeneration shield saves the target. The {T} is the enchantment's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4327225e-140f-45a4-a56f-dcf472f6c062",
		Name:         "Witch's Mist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, {T}: Destroy target creature that was dealt damage this turn.",
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
			},
		}},
	})
}
