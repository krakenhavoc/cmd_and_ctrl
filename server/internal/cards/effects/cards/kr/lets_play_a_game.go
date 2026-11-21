package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Let's Play a Game — {3}{B} Sorcery:
//
//	"Delirium — Choose one. If there are four or more card types among
//	 cards in your graveyard, choose one or more instead.
//	 • Creatures your opponents control get -1/-1 until end of turn.
//	 • Each opponent discards two cards.
//	 • Each opponent loses 3 life and you gain 3 life."
//
// Prophetic Titan's delirium condition (#1655) on a "choose one or
// more instead" count: the maximum becomes every bullet and the
// minimum stays one, so AnyNumberIf(DeliriumForModes). Read "as you
// cast" — the graveyard at announce, not at resolution.
//
// Nothing targets. The -1/-1 is a one-shot set chosen at resolution
// (CR 611.2c); each discard is the opponent's own choice; the gain is
// a flat 3 whatever the opponents actually lost. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd893746-e8bd-49fa-a1ce-755d5bd4f513",
		Name:         "Let's Play a Game",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Creatures your opponents control get -1/-1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.BoostUntilEOT{
						Match:     effects.And(effects.Creature(), effects.OpponentControls()),
						Power:     -1,
						Toughness: -1,
						Label:     "Let's Play a Game — -1/-1 until end of turn",
					}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
			effects.ModeDoing("Each opponent discards two cards.", nil,
				func(item *game.StackItem, ctx *effects.Context, _ int) error {
					for _, opp := range ctx.Opponents() {
						ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player: opp,
							Source: item.SourceCardID,
							N:      2,
						})
					}
					return nil
				}),
			// The gain waits for a prompt the loss asks (a CR 616 order; CR 608.2c, #4155).
			effects.ModeSteps("Each opponent loses 3 life and you gain 3 life.", nil,
				effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentLosesLifeNow(ctx.Game, ctx.Item, 3) }),
				effects.GainLife{Amount: 3}),
		).AnyNumberIf(effects.DeliriumForModes),
	})
}
