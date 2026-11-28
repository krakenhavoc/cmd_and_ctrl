package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodfire Infusion — Enchantment — Aura {2}{R} (Apocalypse, S82 OLD, #5614):
//
//	"Enchant creature you control
//	 {R}, Sacrifice enchanted creature: This Aura deals damage equal to the
//	 sacrificed creature's power to each creature."
//
// The sacrifice is a cost naming only the creature the Aura is attached to
// (SacrificeEnchantedCreature, the new TargetSpec.SourceAttachedTo), so it is
// gone, and the Aura with it, before the ability resolves. The amount is the
// sacrificed creature's power as it last existed on the battlefield, floored at
// zero (CR 107.1b, 608.2h), and the Aura, as last known, deals it to every
// creature, the sacrificed one's controller's included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5bd7c0b-c412-47ed-afb8-1a9ba04be77c",
		Name:         "Bloodfire Infusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Activated: []effects.ActivatedAbility{{
			Label: "{R}, Sacrifice enchanted creature: This Aura deals damage equal to the sacrificed creature's power to each creature.",
			// The amount is the sacrificed creature's power, known only at
			// resolution (AmountIsX).
			Purpose: game.Purpose{
				Answers: game.AnswerRemove | game.AnswerSacOutlet,
				Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true},
			},
			Cost: effects.Plus(effects.ManaCost("{R}"), effects.SacrificeEnchantedCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := ctx.SacrificedPower()
				if n <= 0 {
					return nil
				}
				return effects.DamageEachMatching(ctx, effects.Creature(), n)
			},
		}},
	})
}
