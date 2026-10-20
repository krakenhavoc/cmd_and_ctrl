package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peema Trailblazer — Creature — Elephant Warrior {2}{G}, 3/3:
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, you get
//	 that many {E} (energy counters).
//	 Exhaust — Pay six {E}: Put two +1/+1 counters on this creature.
//	 Then draw cards equal to the greatest power among creatures you
//	 control. (Activate each exhaust ability only once.)"
//
// "That many" is the damage the triggering event dealt
// (ctx.Trigger().Event.Amount). The exhaust ability (#1181) pays six
// energy (ADR 0129 §2); the draw is counted after the counters land, so
// the Trailblazer's own two counters count, as printed ("Then").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "88600bd4-4dcc-4788-bba4-78b6cf5ad8f8",
		Name:            "Peema Trailblazer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Peema Trailblazer — you get that many {E}", youGetThatManyEnergy),
		},
		Activated: []ActivatedAbility{{
			Label:   "Exhaust — Pay six {E}: Put two +1/+1 counters on this creature. Then draw cards equal to the greatest power among creatures you control.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Exhaust: true,
			Cost:    PayEnergy(6),
			// The draw waits for a prompt the counters ask (a CR 616 order; CR 608.2c,
			// #4155) and counts the greatest power when it runs.
			Effect: Do(
				Instruction(func(ctx *Context) error {
					return AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					return DrawCards{N: b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())}.Apply(ctx)
				}),
			),
		}},
	})
}
