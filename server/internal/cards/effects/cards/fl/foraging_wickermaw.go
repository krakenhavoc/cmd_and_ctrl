package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foraging Wickermaw — Artifact Creature — Scarecrow {2}, 1/3 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature enters, surveil 1.
//	 {1}: Add one mana of any color. This creature becomes that color
//	 until end of turn. Activate only once each turn."
//
// The enters trigger is the footprinted Surveil primitive (CR 701.25a).
// The filter is Barrels of Blasting Jelly's once-a-turn mana ability with
// the engine's "becomes that color" rider declared as data
// (ManaAbility.SourceBecomesProducedColor): the colour is often picked
// after the ability's Rider would run, so the engine applies it where
// the colour becomes known — a layer 5 colour-setting effect (CR
// 613.1e) on the Wickermaw for the turn, replacing its colourlessness.
// A mana ability resolves at once (CR 605.3b), so nothing can respond
// between the mana and the colour.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "11ceb2db-021e-422d-976e-7bfe3b67a8aa",
		Name:         "Foraging Wickermaw",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Foraging Wickermaw — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:                       effects.ManaAbilityCost{Mana: "{1}"},
			Produced:                   "{W|U|B|R|G}",
			Label:                      "{1}: Add one mana of any color. This creature becomes that color until end of turn. Activate only once each turn.",
			OncePerTurn:                true,
			SourceBecomesProducedColor: true,
			Answers:                    game.AnswerValue,
		}},
	})
}
