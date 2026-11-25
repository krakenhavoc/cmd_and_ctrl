package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forever Young — Sorcery {1}{B} (Throne of Eldraine, S82 old sets):
//
//	"Put any number of target creature cards from your graveyard on top of your
//	 library.
//	 Draw a card."
//
// Any number of targets, so it can be cast with none. The cards still in the
// graveyard as it resolves are put on top one after another through the shared
// tuck route (a commander's owner is offered the command zone, CR 903.9), in
// the order the targets were chosen, so the caster arranges them by choosing
// them in that order; the last one chosen ends up on top (CR 401.4). The draw
// follows the put (CR 608.2c), so it takes the last card put there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "31f32afa-5b2a-4218-ae7b-642e1cac3911",
		Name:         "Forever Young",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("any number of target creature cards from your graveyard",
			effects.Creature(), effects.YouOwn()).WithCount(0, 0),
		Purpose: game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := ctx.Game.TuckToLibraryForEffect(t.ID, false); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.DrawCards{N: 1},
		),
	})
}
