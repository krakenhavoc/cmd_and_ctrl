package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tanazir Quandrix — Legendary Creature — Elder Dragon {3}{G}{U}, 4/4
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying, trample
//	 When Tanazir Quandrix enters, double the number of +1/+1 counters on
//	 target creature you control.
//	 Whenever Tanazir Quandrix attacks, you may have the base power and
//	 toughness of other creatures you control become equal to Tanazir
//	 Quandrix's power and toughness until end of turn."
//
// The enter trigger is Primordial Hydra's doubling on a target: the creature
// is chosen as the ability goes on the stack, the count is read as it
// resolves, and the counters are placed through AddCounter, so a counter
// doubler applies to the doubling. A target that has left, or has no +1/+1
// counters, gets nothing.
//
// The attack trigger is a "you may" (CR 603.5). Tanazir's power and toughness
// are read as the ability resolves, counters and pumps included, as he last
// existed if he has gone (CR 608.2h); the creatures affected are the other
// creatures its controller has then (CR 611.2c), and the effect sets their BASE
// values (layer 7b), so their counters and pumps still apply on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73c44cf5-a601-4625-9550-e722fba2ec0b",
		Name:            "Tanazir Quandrix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Tanazir Quandrix — double the number of +1/+1 counters on target creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if id, ok := b16FirstLegalTargetCard(ctx); ok {
						return b23DoublePlusOneCountersOn(ctx, id)
					}
					return nil
				}), TargetCreature("target creature you control", YouControl())),
			Optional(WheneverThisAttacks("Tanazir Quandrix — other creatures you control have base power and toughness equal to Tanazir Quandrix's",
				socB14OtherCreaturesBecomeTanazirsSize),
				"Tanazir Quandrix — have the base power and toughness of your other creatures become equal to his?"),
		},
	})
}
