package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drifting Meadow — Land (Urza's Saga):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The white member of Remote Isle's Urza's Saga cycling-land cycle.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6eb2814-0021-4308-ad44-6c8cc59b0d1c",
		Name:         "Drifting Meadow",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
