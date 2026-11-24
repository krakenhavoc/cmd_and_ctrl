package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Great Hall of Starnheim — Land (Kaldheim, #5505):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 {W}{W}{B}, {T}, Sacrifice this land and a creature you control: Create a
//	 4/4 white Angel Warrior creature token with flying and vigilance.
//	 Activate only as a sorcery."
//
// Both sacrifices are costs paid together: the land (this permanent) and a
// creature that is not the land (a permanent can't be sacrificed twice, so
// an animated Great Hall can't be both).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75d8ef26-b745-448b-94b9-3c64668ce171",
		Name:         "Great Hall of Starnheim",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{W}{W}{B}, {T}, Sacrifice this land and a creature you control: Create a 4/4 white Angel Warrior creature token with flying and vigilance. Activate only as a sorcery.",
			Cost: effects.Plus(effects.ManaCost("{W}{W}{B}"), effects.TapCost(), effects.SacrificeThis(),
				effects.SacrificeAnotherN(1, "a creature you control", effects.Creature())),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 white Angel Warrior with flying and vigilance"), N: 1}),
		}},
	})
}
