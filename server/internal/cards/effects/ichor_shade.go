package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ichor Shade — Creature — Phyrexian Shade {2}{B}, 2/3 (March of the
// Machine):
//
//	"At the beginning of your end step, if an artifact or creature was
//	 put into a graveyard from the battlefield this turn, put a +1/+1
//	 counter on this creature."
//
// An intervening "if" (CR 603.4): checked as the step begins and again as
// the ability resolves. The fact is the turn's, whoever controlled the
// permanent and tokens included.
//
// No simplification.
func init() {
	const label = "Ichor Shade — put a +1/+1 counter on this creature"
	Register(Spec{
		OracleID:     "083ea8b0-ff2a-406e-90f1-668f33443747",
		Name:         "Ichor Shade",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou,
				func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return s82MomB11ArtifactOrCreatureDiedThisTurn(g)
				}), label,
				Do(If{
					Cond: CondBoard{
						Name: "an artifact or creature was put into a graveyard from the battlefield this turn",
						Test: func(ctx *Context) bool { return s82MomB11ArtifactOrCreatureDiedThisTurn(ctx.Game) },
					},
					Then: []Applier{CounterOnThis{Kind: game.CounterPlusOne, N: 1}},
				})),
		},
	})
}
