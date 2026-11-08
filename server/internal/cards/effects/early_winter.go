package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Early Winter — Instant {4}{B} (Bloomburrow, #4467):
//
//	"Choose one —
//	 • Exile target creature.
//	 • Target opponent exiles an enchantment they control."
//
// The first bullet exiles its target if it is still legal (CR 608.2b).
// The second is Doomfall's exile edict over enchantments
// (msc17ExileEdict): the TARGETED opponent chooses one of the
// enchantments they control and exiles it, so hexproof and protection on
// the enchantment don't matter. An opponent with no enchantment is not
// asked, and one that leaves before the answer is withdrawn from the
// pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbf2653e-9cc4-442f-bc73-6e2ba6e067cb",
		Name:         "Early Winter",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Exile target creature.", TargetCreature("target creature"), ExileTheModesTarget),
			ModeDoing("Target opponent exiles an enchantment they control.",
				TargetPlayer("target opponent", Opponent()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return msc17ExileEdict(ctx, t.ID, "Early Winter — exile an enchantment you control", Enchantment(), doomfallExile)
				}),
		),
	})
}
