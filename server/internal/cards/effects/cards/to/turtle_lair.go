package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turtle Lair — Land (Teenage Mutant Ninja Turtles, #4457):
//
//	"{T}: Add {C}.
//	 {T}: Add one mana of any color. Spend this mana only to cast a Ninja
//	 or Turtle spell.
//	 {3}, {T}: Target Ninja or Turtle can't be blocked this turn."
//
// The coloured mana is Fíli and Kíli's restriction (CR 106.6): casting
// only, never an activation, and only a spell with the Ninja or the
// Turtle subtype (ManaRestrictSubtype's `|` alternation). The third
// ability is Access Tunnel's (targetCantBeBlockedThisTurn) with a
// subtype target: any permanent that is a Ninja or a Turtle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb002bbc-08df-4bf0-bea3-46494ad261b6",
		Name:         "Turtle Lair",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.ColorlessTapRow(),
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Label:        "{T}: Add one mana of any color. Spend this mana only to cast a Ninja or Turtle spell.",
				Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictSubtype("Ninja|Turtle")},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target Ninja or Turtle can't be blocked this turn.",
			Purpose: effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target Ninja or Turtle", effects.Or(effects.HasSubtype("Ninja"), effects.HasSubtype("Turtle"))),
			Effect:  effects.TargetCantBeBlockedThisTurn("Turtle Lair"),
		}},
	})
}
