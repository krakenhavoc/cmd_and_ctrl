package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gallia, the Merrymaker — Legendary Creature — Satyr {1}{R}, 2/1:
//
//	"Haste
//	 Each other creature you control with a +1/+1 counter on it has
//	 haste.
//	 {1}{R}, {T}: Put a +1/+1 counter on target creature that entered
//	 this turn."
//
// The haste grant is a layer 6 static over the controller's other
// creatures, gated on a +1/+1 counter, so it follows the counter on and
// off. The activation reads the per-object entry tally: a creature that
// entered during another player's turn does not count, and the target
// may be any player's creature (the text says "target creature"). The
// tap symbol makes Gallia summoning-sick like any creature's, haste
// excepted.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "87f39199-3e4b-44fa-8406-62019eb43c10",
		Name:            "Gallia, the Merrymaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller && target.IsCreature() &&
					target.Counters[game.CounterPlusOne] > 0
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, a := range c.Abilities {
					if a == "haste" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "haste")
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, {T}: Put a +1/+1 counter on target creature that entered this turn.",
			Cost:    Plus(ManaCost("{1}{R}"), TapCost()),
			Targets: TargetCreature("target creature that entered this turn", rfCreatureBEnteredThisTurn()),
			Effect:  plusOneCounterOnChosenTargets,
		}},
	})
}
