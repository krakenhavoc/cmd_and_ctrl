package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// High Sentinels of Arashin — Creature — Bird Soldier {3}{W}, 3/4 (Khans
// of Tarkir; Jumpstart):
//
//	"Flying
//	 This creature gets +1/+1 for each other creature you control with a
//	 +1/+1 counter on it.
//	 {3}{W}: Put a +1/+1 counter on target creature."
//
// The static is a Layer 7c modifier over the Sentinels alone, counted
// as the layers run: a creature of yours gaining its first counter, or
// losing its last, changes it at once, and the Sentinels' own counters
// never count. The activation puts a counter on any creature, itself
// included, at instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f0e7d147-379e-45b7-bfc4-d926637e060d",
		Name:            "High Sentinels of Arashin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each other creature you control with a +1/+1 counter on it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := jmpB5OtherCreaturesWithACounter(g, source)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{3}{W}: Put a +1/+1 counter on target creature.",
			Cost:    ManaCost("{3}{W}"),
			Targets: TargetCreature("target creature"),
			Effect:  PlusOneCounterOnTheTarget,
		}},
	})
}
