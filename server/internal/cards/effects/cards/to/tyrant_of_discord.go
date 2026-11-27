package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tyrant of Discord — Creature — Elemental {4}{R}{R}{R}, 7/7 (Avacyn Restored,
// S82 OLD, #5570):
//
//	"When this creature enters, target opponent chooses a permanent they
//	 control at random and sacrifices it. If a nonland permanent is
//	 sacrificed this way, repeat this process."
//
// The pick is the game's keyed random draw over everything the opponent
// controls (so an undo rewinds it); the process repeats only after a nonland
// permanent has actually been sacrificed, and ends when a land is, when the
// sacrifice is refused (a permanent that can't be sacrificed) or when they
// control nothing. A target gone by resolution means nothing happens
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce1bf2c4-93d5-41e3-a450-6b95fad7dd21",
		Name:         "Tyrant of Discord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Tyrant of Discord — target opponent sacrifices a random permanent, and repeats while a nonland permanent is sacrificed",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok || t.Kind != game.TargetPlayer {
							return nil
						}
						return effects.S82OldB437RandomSacrificeLoop(ctx, t.ID)
					}),
				effects.TargetPlayer("target opponent", effects.Opponent()),
			),
		},
	})
}
