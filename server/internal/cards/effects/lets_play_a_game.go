package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "dd893746-e8bd-49fa-a1ce-755d5bd4f513",
		Name:         "Let's Play a Game",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Creatures your opponents control get -1/-1 until end of turn.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return BoostUntilEOT{
						Match:     And(Creature(), OpponentControls()),
						Power:     -1,
						Toughness: -1,
						Label:     "Let's Play a Game — -1/-1 until end of turn",
					}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1, OpponentsOnly: true}}),
			ModeDoing("Each opponent discards two cards.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					for _, opp := range ctx.Opponents() {
						ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player: opp,
							Source: item.SourceCardID,
							N:      2,
						})
					}
					return nil
				}),
			ModeDoing("Each opponent loses 3 life and you gain 3 life.", nil,
				func(item *game.StackItem, ctx *Context, _ int) error {
					if err := eachOpponentLosesLife(ctx.Game, item, 3); err != nil {
						return err
					}
					return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
				}),
		).AnyNumberIf(DeliriumForModes),
	})
}
