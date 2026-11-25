package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aviary Mechanic — Creature — Dwarf Artificer {1}{W}, 2/2 (Kaladesh, #5539):
//
//	"When this creature enters, you may return another permanent you control
//	 to its owner's hand."
//
// Ambrosia Whiteheart's resolution-time pick (nothing targets, so the pick is
// made as the trigger resolves), over every other permanent its controller
// controls rather than creatures only. Choosing none is the "may" declined;
// the Mechanic itself is never offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b47d37ee-2bb2-4ca8-b245-94d2cc4b6684",
		Name:         "Aviary Mechanic",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Aviary Mechanic — you may return another permanent you control to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.S82OneB04ReturnAnotherPermanent(g, item,
						"Aviary Mechanic — you may return another permanent you control to its owner's hand", aviaryMechanicReturn)
				}),
		},
	})
}

// aviaryMechanicReturn returns the permanent the controller picked. An
// on-disk key.
var aviaryMechanicReturn = effects.RegisterPermanentsPick("prompt-run/aviary-mechanic-return", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			return effects.BounceToHand{Target: id}.Apply(ctx)
		}
		return nil
	},
})
