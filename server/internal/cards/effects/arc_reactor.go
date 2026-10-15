package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arc Reactor — Artifact {5}:
//
//	"Improvise ... This artifact enters tapped.
//	 {T}: Add {C}{C}{C}."
//
// Improvise is Spec.TapCost; the rest is Basalt Monolith's mana ability
// behind SelfEntersTapped. No simplification.
func init() {
	Register(Spec{
		OracleID:     "55f7b005-31ea-4865-a8f3-460d15bed795",
		Name:         "Arc Reactor",
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}",
			Label:    "Add {C}{C}{C}",
		}},
	})
}
