package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tolaria West — Land (Future Sight, #5594):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 Transmute {1}{U}{U} ({1}{U}{U}, Discard this card: Search your
//	 library for a card with mana value 0, reveal it, put it into your
//	 hand, then shuffle. Transmute only as a sorcery.)"
//
// A tapped blue land with transmute (CR 702.53). A land's mana value is
// 0, so the shared Transmute row (Drift of Phantasms) finds a card with
// mana value 0, lands and zero-cost spells alike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46eefc72-2d9e-4389-8ae1-26d9ee472b5c",
		Name:         "Tolaria West",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{Transmute("{1}{U}{U}")},
	})
}
