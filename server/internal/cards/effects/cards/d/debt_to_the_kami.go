package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Debt to the Kami — Instant {2}{B} (Kamigawa: Neon Dynasty, #5493):
//
//	"Choose one —
//	 • Target opponent exiles a creature they control.
//	 • Target opponent exiles an enchantment they control."
//
// An exile edict: the targeted opponent chooses, the permanent isn't
// targeted (hexproof and protection don't help it), and an opponent with
// nothing of the kind isn't asked. Same body as Doomfall's exile edict.
//
// No simplification.
func init() {
	edict := func(text, label string, pred effects.CardPredicate) game.ModeOption {
		return effects.ModeWithPurpose(effects.ModeDoing(text,
			effects.TargetPlayer("target opponent", effects.Opponent()),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return effects.Msc17ExileEdict(ctx, t.ID, label, pred, effects.DoomfallExile)
			}), game.Purpose{Answers: game.AnswerRemove})
	}
	effects.Register(effects.Spec{
		OracleID:     "c6c1e576-7197-4067-8cd1-aa535188e1b8",
		Name:         "Debt to the Kami",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			edict("Target opponent exiles a creature they control.",
				"Debt to the Kami — exile a creature you control", effects.Creature()),
			edict("Target opponent exiles an enchantment they control.",
				"Debt to the Kami — exile an enchantment you control", effects.Enchantment()),
		),
	})
}
