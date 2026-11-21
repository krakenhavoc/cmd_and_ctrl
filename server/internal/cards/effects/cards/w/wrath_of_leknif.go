package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrath of Leknif — Sorcery {1}{W}{W}{U} (Mystery Booster 2, #5948):
//
//	"Destroy all creatures. They can't be regenerated. Untap up to four lands
//	 you control."
//
// Wrath of God's sweep (the creatures die as one event and ignore
// regeneration shields, CR 701.19c), then the controller chooses up to four
// of their own tapped lands to untap. The untap waits for any prompt the sweep
// asks (a commander's zone choice; CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a4bf27f-cf53-4a10-b909-be2ca696649d",
		Name:         "Wrath of Leknif",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DestroyAllMatching{Match: effects.Creature(), CantBeRegenerated: true}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GapB34UntapUpToYourLands(ctx, 4, "Wrath of Leknif — untap up to four lands you control")
			}),
		),
	})
}
