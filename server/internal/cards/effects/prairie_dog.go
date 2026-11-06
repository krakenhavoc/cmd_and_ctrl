package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prairie Dog — Creature — Squirrel {1}{W}, 2/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Lifelink
//	 At the beginning of your end step, if you haven't cast a spell from
//	 your hand this turn, put a +1/+1 counter on this creature.
//	 {4}{W}: Until end of turn, if you would put one or more +1/+1
//	 counters on a creature you control, put that many plus one +1/+1
//	 counters on it instead."
//
// The end-step counter is an intervening "if" (CR 603.4) over the turn's
// cast tally (AtYourEndStepIfNoSpellFromHand, #5235). The activated
// ability makes a turn-scoped replacement record (CR 614.1a,
// game.ModOneMorePlusOneCounter, ADR 0041 P8): it lasts the turn whether
// or not the Dog stays, applies only to +1/+1 counters its controller
// puts on a creature they control (not counters an opponent puts), and
// two activations make two records, so two more counters (CR 616: each
// applies once to the event). Activated before the end step, it adds one
// to the Dog's own end-step counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "70987216-f84d-43b9-8d75-bb21ff37747b",
		Name:            "Prairie Dog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfNoSpellFromHand("Prairie Dog — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{W}: Until end of turn, if you would put one or more +1/+1 counters on a creature you control, put that many plus one +1/+1 counters on it instead.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{4}{W}"),
			Effect: Do(Instruction(func(ctx *Context) error {
				ctx.Game.OneMorePlusOneCounterThisTurnForEffect(ctx.Item.SourceCardID, ctx.Controller(),
					"Prairie Dog — one more +1/+1 counter")
				return nil
			})),
		}},
	})
}
