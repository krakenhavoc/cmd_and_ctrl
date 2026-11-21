package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seeker's Folly — Sorcery {2}{B} (Foundations, #4465):
//
//	"Choose one —
//	 • Target opponent discards two cards.
//	 • Creatures your opponents control get -1/-1 until end of turn."
//
// Cerebral Confiscation's discard mode and Let's Play a Game's -1/-1
// mode. The opponent picks the two cards. The -1/-1 reaches the
// creatures opponents control as it resolves, not ones that arrive
// later (CR 611.2c), and is a declared one-sided sweep (ADR 0126).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "223794f2-67a3-4ab5-81c5-046c63307c6c",
		Name:         "Seeker's Folly",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target opponent discards two cards.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.TheModesTargetDiscards(2)),
			effects.ModeWithPurpose(effects.ModeDoing("Creatures your opponents control get -1/-1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.OpponentControls()),
						Power:     -1,
						Toughness: -1,
						Label:     "Seeker's Folly — -1/-1 until end of turn",
					}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
		),
	})
}
