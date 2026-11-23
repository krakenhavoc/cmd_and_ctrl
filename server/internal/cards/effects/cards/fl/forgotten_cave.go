package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forgotten Cave — Land (EDHREC top 1,000, #3654):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 Cycling {R} ({R}, Discard this card: Draw a card.)"
//
// The red member of the Onslaught cycling-land cycle: Polluted Mire's
// shape with a different colour and cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "394c6de5-7957-4a0b-a6b9-ee0c707cd022",
		Name:         "Forgotten Cave",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{R}")},
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
}
