package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Poison the Waters — Sorcery {1}{B}:
//
//	"Choose one —
//	 • All creatures get -1/-1 until end of turn.
//	 • Target player reveals their hand. You choose an artifact or
//	   creature card from it. That player discards that card."
//
// The first bullet shrinks every creature on the battlefield as the
// spell resolves (CR 611.2c: one that enters later is not affected).
// The second is the revealed-hand pick (ADR 0116) filtered to artifact
// or creature cards; a hand with neither is revealed and nothing is
// discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2c9bf40b-ddcb-4f46-a81e-56b91dfc784f",
		Name:         "Poison the Waters",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("All creatures get -1/-1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{Match: Creature(), Power: -1, Toughness: -1, Label: "Poison the Waters"}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}}),
			ModeDoing("Target player reveals their hand. You choose an artifact or creature card from it. That player discards that card.",
				TargetPlayer("target player"),
				ModeTargetRevealsYouChooseDiscard(Or(Artifact(), Creature()), "artifact or creature card")),
		),
	})
}
