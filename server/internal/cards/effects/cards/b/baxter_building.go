package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baxter Building — Land (Marvel Super Heroes, #4455):
//
//	"{T}: Add {C}.
//	 {4}, {T}: Add four mana in any combination of colors.
//	 {4}, {T}: Draw a card. Activate only if you control a creature with
//	 toughness 4 or greater."
//
// The first two are mana abilities (CR 605.1a): the second pays {4} out
// of the pool as it taps, its four slots each any colour. The draw is an
// ordinary activation on the stack with Bonders' Enclave's gate (CR
// 602.1b): the creature's CURRENT toughness, so an anthem or a counter
// counts, read when the ability is activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71bc69a5-7cec-4abd-b97d-13f8e1f9afac",
		Name:         "Baxter Building",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.ColorlessTapRow(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{4}"},
				Produced: effects.AnyCombinationOfColors(4),
				Label:    "{4}, {T}: Add four mana in any combination of colors.",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{4}, {T}: Draw a card. Activate only if you control a creature with toughness 4 or greater.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Condition: effects.ControlsAtLeast(1, mshA1CreatureWithToughnessAtLeast4),
			Effect:    effects.B36DrawOne,
		}},
	})
}

// mshA1CreatureWithToughnessAtLeast4 matches a creature whose current
// toughness is at least 4: Baxter Building's gate.
func mshA1CreatureWithToughnessAtLeast4(c game.Card) bool {
	return c.IsCreature() && c.CurrentToughness() >= 4
}
