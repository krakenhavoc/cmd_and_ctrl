package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "46eefc72-2d9e-4389-8ae1-26d9ee472b5c",
		Name:         "Tolaria West",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []effects.ActivatedAbility{effects.Transmute("{1}{U}{U}")},
	})
}
