package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demogorgon's Clutches — Sorcery {2}{B} (Adventures in the Forgotten Realms):
//
//	"Target opponent discards two cards, mills two cards, and loses 2 life."
//
// The opponent chooses what to discard. The mill and the life loss run in the
// discard's continuation (CR 608.2c), so a card discarded can't have been milled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bbab99f-2ff6-4d48-b7c1-1ab0500c9422",
		Name:         "Demogorgon's Clutches",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2, LifeLoss: 2}),
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer || !ctx.IsTargetLegal(t) {
				return nil
			}
			// The mill and the life loss run in the discard's continuation,
			// once the opponent has chosen (CR 608.2c).
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(
				game.DiscardPrompt{Player: t.ID, Source: ctx.Source(), N: 2},
				demogorgonsClutchesAfterDiscard, nil)
		},
	})
}

// demogorgonsClutchesAfterDiscard is "mills two cards, and loses 2 life",
// run once the discard is settled. An on-disk key.
var demogorgonsClutchesAfterDiscard = game.RegisterDiscardThen("choose-cards/demogorgons-clutches-mill-and-drain",
	func(g *game.Game, r game.DiscardsDone) error {
		if err := g.MillNForEffect(r.Player, 2); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(r.Source, r.Player, -2)
	})
