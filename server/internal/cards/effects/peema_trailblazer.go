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
			Exhaust: true,
			Cost:    PayEnergy(6),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
				return DrawCards{N: b42GreatestPowerControlledBy(g, item.Controller)}.Apply(ctx)
			},
		}},
	})
}
