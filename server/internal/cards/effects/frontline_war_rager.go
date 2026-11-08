package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frontline War-Rager — Creature — Kavu Soldier {2}{R}, 2/3 (Edge of
// Eternities, #4461):
//
//	"At the beginning of your end step, if you control two or more tapped
//	 creatures, put a +1/+1 counter on this creature."
//
// Vaultguard Trooper's end-step trigger with a +1/+1 counter on the
// War-Rager in place of the discard (CounterOnThis, which puts nothing
// on a War-Rager that has left the battlefield or come back as a new
// object, CR 400.7). The "if" is an intervening if (CR 603.4): checked
// as the end step begins and again as the ability resolves. The
// War-Rager itself counts toward the two tapped creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "610d006a-3ceb-4136-b92b-a500c74d06b0",
		Name:         "Frontline War-Rager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control two or more tapped creatures",
				Test: func(ctx *Context) bool { return b24TappedCreaturesControlled(ctx.Game, ctx.Controller()) >= 2 },
			}, On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24TappedCreaturesControlled(g, source.Controller) >= 2
			}), "Frontline War-Rager — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
