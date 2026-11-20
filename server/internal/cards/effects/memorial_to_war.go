package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f98db69c-b330-4560-ac53-10857674466b",
		Name:         "Memorial to War",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}, {T}, Sacrifice this land: Destroy target land.",
			Cost:    Plus(ManaCost("{4}{R}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target land", Land()),
			Effect:  destroyChosenPermanent,
		}},
	})
}
