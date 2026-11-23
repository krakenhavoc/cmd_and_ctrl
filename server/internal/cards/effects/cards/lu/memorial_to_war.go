package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memorial to War — Land (Dominaria, #5530):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 {4}{R}, {T}, Sacrifice this land: Destroy target land."
//
// Memorial to Genius's frame with Strip Mine's destroy: it enters
// tapped, taps for {R}, and sacrifices itself to destroy a land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f98db69c-b330-4560-ac53-10857674466b",
		Name:         "Memorial to War",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}, {T}, Sacrifice this land: Destroy target land.",
			Cost:    effects.Plus(effects.ManaCost("{4}{R}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
