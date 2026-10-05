package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "22a78443-db21-4656-b38a-e3e3186fd94b",
		Name:         "Collective Brutality",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseOneOrMore(
			ModeDoing("Target opponent reveals their hand. You choose an instant or sorcery card from it. That player discards that card.",
				TargetPlayer("target opponent", Opponent()),
				ModeTargetRevealsYouChooseDiscard(Or(Instant(), Sorcery()), "instant or sorcery card")),
			ModeDoing("Target creature gets -2/-2 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-2, -2, "Collective Brutality")),
			ModeDoing("Target opponent loses 2 life and you gain 2 life.",
				TargetPlayer("target opponent", Opponent()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					if err := (GainLife{Player: t.ID, Amount: -2}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
				}),
		), EscalateDiscard(1)),
	})
}
