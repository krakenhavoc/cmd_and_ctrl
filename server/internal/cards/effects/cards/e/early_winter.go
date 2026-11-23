package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "fbf2653e-9cc4-442f-bc73-6e2ba6e067cb",
		Name:         "Early Winter",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Exile target creature.", effects.TargetCreature("target creature"), effects.ExileTheModesTarget),
			effects.ModeDoing("Target opponent exiles an enchantment they control.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.Msc17ExileEdict(ctx, t.ID, "Early Winter — exile an enchantment you control", effects.Enchantment(), effects.DoomfallExile)
				}),
		),
	})
}
