package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm King's Thunder — Instant {X}{R}{R}{R} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"When you next cast an instant or sorcery spell this turn, copy that spell
//	 X times. You may choose new targets for the copies."
//
// Rimefire Torque's "when you next cast" delayed trigger (CR 603.7), with X
// captured as the spell resolves and carried as the body's Amount: the copies
// are made X at a time through CopySpell (CR 707.10, not cast), each with the
// new-targets offer (CR 707.10c). The copied spell is read from last-known
// information if it has already gone (CR 608.2h). X of 0 makes no copy; if no
// instant or sorcery is cast this turn the trigger ends with the turn.
// stormKingsThunderCopies is an on-disk body key: never renamed, never reused.
var stormKingsThunderCopies = game.DelayedBody("storm-kings-thunder/copy-x-times",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		cast := ctx.PayloadCards()
		if len(cast) == 0 || p.Amount <= 0 {
			return nil
		}
		return effects.CopySpell{
			StackID:          cast[0],
			Controller:       item.Controller,
			Count:            p.Amount,
			ChooseNewTargets: true,
			FromLastKnown:    true,
		}.Apply(ctx)
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5c5f462-10fa-422a-90da-07f46257f22f",
		Name:         "Storm King's Thunder",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			d := effects.WhenYouNextCast(
				"Storm King's Thunder — copy that spell X times",
				game.CastFilter{Types: []string{"Instant", "Sorcery"}},
				stormKingsThunderCopies)
			d.Params = game.EffectParams{Amount: ctx.X()}
			return d.Apply(ctx)
		},
	})
}
