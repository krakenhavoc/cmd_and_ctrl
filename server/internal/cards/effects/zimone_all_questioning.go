package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zimone, All-Questioning — Legendary Creature — Human Wizard {1}{G}{U}, 1/1
// (Secrets of Strixhaven Commander, #4834):
//
//	"At the beginning of your end step, if a land entered the battlefield
//	 under your control this turn and you control a prime number of lands,
//	 create Primo, the Indivisible, a legendary 0/0 green and blue Fractal
//	 creature token, then put that many +1/+1 counters on it. (2, 3, 5, 7,
//	 11, 13, 17, 19, 23, 29, and 31 are prime numbers.)"
//
// An intervening "if" (CR 603.4): both halves are checked as the end step
// begins and again as the ability resolves (InterveningIf). "A land entered
// the battlefield under your control this turn" is the turn tally (bumped at
// the entry itself, so a land that has since died or been sacrificed still
// counts). "That many" is the number of lands as the ability resolves
// (CR 608.2h). Primo is legendary, so a second one is put into the graveyard
// by the legend rule (CR 704.5j); the token row is the printed token card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c044dd1-8938-4d62-a7b2-e2298de08f70",
		Name:         "Zimone, All-Questioning",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "a land entered under your control this turn and you control a prime number of lands",
				Test: func(ctx *Context) bool {
					return socB15LandEnteredAndPrimeLands(ctx.Game, ctx.Controller())
				},
			}, On(game.EventBeginEndStep,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor == source.Controller && socB15LandEnteredAndPrimeLands(g, source.Controller)
				},
				"Zimone, All-Questioning — create Primo, the Indivisible with a +1/+1 counter for each land you control",
				socB15PrimoWithLandCounters())),
		},
	})
}
