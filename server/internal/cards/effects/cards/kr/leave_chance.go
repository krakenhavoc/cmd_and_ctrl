package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leave // Chance — Instant // Sorcery with aftermath (Hour of Devastation):
//
//	Leave {1}{W}: "Return any number of target permanents you own to your hand."
//	Chance {3}{R}: "Aftermath (Cast this spell only from your graveyard. Then
//	 exile it.) Discard any number of cards, then draw that many cards."
//
// Chance asks for the discard first and draws as many cards as were really
// discarded (Fervent Mastery's discard-then-draw shape).
func init() {
	const oracle = "dff5d16f-3a41-4873-8188-076e77f18c75"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Leave",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("any number of target permanents you own", effects.YouOwn()).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.NeoB20BounceTargetedPermanents(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Chance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			me := ctx.Controller()
			n := effects.B14HandSize(ctx.Game, me)
			if n == 0 {
				return nil
			}
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   me,
				Source:   ctx.Source(),
				N:        n,
				UpTo:     true,
				Question: "Chance — discard any number of cards, then draw that many",
				Asked:    game.NewAskedItem(ctx.Item),
			}, leaveChanceDiscarded, nil)
		},
	})
}

var leaveChanceDiscarded = game.RegisterDiscardThen("choose-cards/leave-chance-discarded", func(g *game.Game, r game.DiscardsDone) error {
	if n := r.Discarded.Count(); n > 0 {
		return g.DrawNForEffect(r.Player, n)
	}
	return nil
})
