package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Distant Melody — Sorcery {3}{U}:
//
//	"Choose a creature type. Draw a card for each permanent you
//	 control of that type."
//
// The type is chosen as the spell resolves (#2382); the draw count is
// read when the answer arrives, from the board as it is then. It counts
// PERMANENTS, not creatures, so a Kindred artifact or a changeling
// counts, and a Zombie you control that is not a creature does too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac6c5852-71c7-4f19-8c87-ccb345052862",
		Name:         "Distant Melody",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Distant Melody — choose a creature type",
				ctx.ItemTail(distantMelodyTypeChosen))
			return nil
		},
	})
}

// distantMelodyTypeChosen draws a card for each permanent of the chosen type
// the caster controls. An on-disk key.
var distantMelodyTypeChosen = effects.CreatureTypeChosen("then/distant-melody-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	controller := ctx.Controller()
	return effects.DrawCards{Player: controller, N: effects.PermanentsOfTypeYouControl(ctx.Game, controller, t)}.Apply(ctx)
})
