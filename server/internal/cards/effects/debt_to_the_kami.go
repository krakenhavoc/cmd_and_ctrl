package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	edict := func(text, label string, pred CardPredicate) game.ModeOption {
		return ModeWithPurpose(ModeDoing(text,
			TargetPlayer("target opponent", Opponent()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return msc17ExileEdict(ctx, t.ID, label, pred, doomfallExile)
			}), game.Purpose{Answers: game.AnswerRemove})
	}
	Register(Spec{
		OracleID:     "c6c1e576-7197-4067-8cd1-aa535188e1b8",
		Name:         "Debt to the Kami",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			edict("Target opponent exiles a creature they control.",
				"Debt to the Kami — exile a creature you control", Creature()),
			edict("Target opponent exiles an enchantment they control.",
				"Debt to the Kami — exile an enchantment you control", Enchantment()),
		),
	})
}
