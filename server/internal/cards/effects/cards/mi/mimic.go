package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mimic — Artifact — Treasure {2} (Adventures in the Forgotten Realms):
//
//	"{T}, Sacrifice this artifact: Add one mana of any color.
//	 {2}: This artifact becomes a Shapeshifter artifact creature with base
//	 power and toughness 3/3 until end of turn."
//
// The mana ability is a Treasure's (Buried Treasure's). The animation is Firdoch
// Core's BecomeCreatureUntilEOT: it adds Artifact and Creature and the
// Shapeshifter creature type, and the Mimic keeps its Treasure subtype, because
// "becomes a [creature type] artifact creature" retains every prior type but
// creature types (CR 205.1b). Mimic is a nontoken Treasure, and mana from it
// is still Treasure mana: the mana source records the subtype, not token-ness.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba8156cd-2e78-4763-b61d-8bc85b2f3657",
		Name:         "Mimic",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This artifact becomes a Shapeshifter artifact creature with base power and toughness 3/3 until end of turn.",
			Cost:    effects.ManaCost("{2}"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.BecomeCreatureUntilEOT{
					Subtypes: []string{"Shapeshifter"},
					SetPower: 3, SetToughness: 3,
					Label: "Mimic — a 3/3 Shapeshifter artifact creature",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
