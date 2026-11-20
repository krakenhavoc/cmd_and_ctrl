package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2a4bf27f-cf53-4a10-b909-be2ca696649d",
		Name:         "Wrath of Leknif",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return DestroyAllMatching{Match: Creature(), CantBeRegenerated: true}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return gapB34UntapUpToYourLands(ctx, 4, "Wrath of Leknif — untap up to four lands you control")
			}),
		),
	})
}
