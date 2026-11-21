package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nirkana Revenant — Creature — Vampire Shade {4}{B}{B}, 4/4:
//
//	"Whenever you tap a Swamp for mana, add an additional {B}.
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Crypt Ghast's triggered mana ability (a Swamp of yours, by effective
// type line, adds one more {B}, with no stack) and a shade's pump,
// which can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6dff1def-5b94-40c8-a942-dea8743ab47c",
		Name:         "Nirkana Revenant",
		Completeness: effects.CompletenessFull,
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverYouTapALandOfSubtypeForMana(
				"Nirkana Revenant — add an additional {B}", "Swamp", effects.AddsFixedMana("{B}")),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: This creature gets +1/+1 until end of turn",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Nirkana Revenant — +1/+1 until end of turn"),
		}},
	})
}
