package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collective Brutality — Sorcery {1}{B}:
//
//	"Escalate—Discard a card. (Pay this cost for each mode chosen
//	 beyond the first.)
//	 Choose one or more —
//	 • Target opponent reveals their hand. You choose an instant or
//	   sorcery card from it. That player discards that card.
//	 • Target creature gets -2/-2 until end of turn.
//	 • Target opponent loses 2 life and you gain 2 life."
//
// Escalate (CR 702.120a, #2126): the discard is owed once per mode
// beyond the first and is paid with the spell on the stack, so a
// discard payoff triggers above it and a countered Brutality still
// cost the cards. The first bullet is the revealed-hand pick (ADR 0116)
// filtered to instants and sorceries; a hand with none is revealed and
// nothing is discarded (CR 609.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22a78443-db21-4656-b38a-e3e3186fd94b",
		Name:         "Collective Brutality",
		Completeness: effects.CompletenessFull,
		Modes: effects.Escalating(effects.ChooseOneOrMore(
			effects.ModeDoing("Target opponent reveals their hand. You choose an instant or sorcery card from it. That player discards that card.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.ModeTargetRevealsYouChooseDiscard(effects.Or(effects.Instant(), effects.Sorcery()), "instant or sorcery card")),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -2/-2 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-2, -2, "Collective Brutality")), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
			// The gain waits for a prompt the loss asks (a CR 616 order; CR 608.2c, #4155).
			effects.ModeSteps("Target opponent loses 2 life and you gain 2 life.",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.GainLife{Player: t.ID, Amount: -2}.Apply(ctx)
				}),
				effects.SocB15ModeYouGainLife(2)),
		), effects.EscalateDiscard(1)),
	})
}
