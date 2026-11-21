package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "2c9bf40b-ddcb-4f46-a81e-56b91dfc784f",
		Name:         "Poison the Waters",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("All creatures get -1/-1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{Match: effects.Creature(), Power: -1, Toughness: -1, Label: "Poison the Waters"}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}}),
			effects.ModeDoing("Target player reveals their hand. You choose an artifact or creature card from it. That player discards that card.",
				effects.TargetPlayer("target player"),
				effects.ModeTargetRevealsYouChooseDiscard(effects.Or(effects.Artifact(), effects.Creature()), "artifact or creature card")),
		),
	})
}
