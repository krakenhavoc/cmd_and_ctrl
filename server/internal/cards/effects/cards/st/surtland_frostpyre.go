package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surtland Frostpyre — Land (Kaldheim, #5505):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {2}{U}{U}{R}, {T}, Sacrifice this land: Scry 2. This land deals 2 damage
//	 to each creature. Activate only as a sorcery."
//
// The damage is the sacrificed land's (last known information, CR 608.2h),
// dealt after the scry is answered, to every creature including the
// activator's own.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "965aa666-3919-4053-8584-b773bdd54f0b",
		Name:         "Surtland Frostpyre",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{U}{U}{R}, {T}, Sacrifice this land: Scry 2. This land deals 2 damage to each creature. Activate only as a sorcery.",
			Purpose: game.Purpose{Scry: 2,
				Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
			Cost:         effects.Plus(effects.ManaCost("{2}{U}{U}{R}"), effects.TapCost(), effects.SacrificeThis()),
			SorcerySpeed: true,
			Effect: effects.Do(effects.Scry{N: 2}, effects.Instruction(func(ctx *effects.Context) error {
				return effects.DamageEachMatching(ctx, effects.Creature(), 2)
			})),
		}},
	})
}
