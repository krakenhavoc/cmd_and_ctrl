package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dire Mimic — Artifact — Treasure {2} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"Flash
//	 {T}, Sacrifice this artifact: Add one mana of any color.
//	 {3}: This artifact becomes a Shapeshifter artifact creature with base
//	 power and toughness 5/5 until end of turn."
//
// Mimic with flash and a 5/5 for {3}. It is a Treasure by subtype, so it
// carries the Treasure's own mana ability, written out as printed on the
// card. As a creature it keeps being an artifact Treasure (CR 205.1b: it
// gains the creature type and keeps its others).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b657b3df-9287-44cf-a1cd-a80ead71a553",
		Name:            "Dire Mimic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Sacrifice this artifact: Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: This artifact becomes a Shapeshifter artifact creature with base power and toughness 5/5 until end of turn.",
			Cost:    effects.ManaCost("{3}"),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.BecomesShapeshifterUntilEOT(5, 5, "Dire Mimic — a 5/5 Shapeshifter artifact creature"),
		}},
	})
}
