package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rushwood Herbalist — Creature — Human Spellshaper {2}{G}, 2/2 (Mercadian
// Masques, S82 OLD b681, #5619):
//
//	"{G}, {T}, Discard a card: Regenerate target creature."
//
// Black Poplar Shaman's regeneration shield (CR 701.19a) on any creature,
// paid with mana, tap and a discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2e35bcef-505a-4559-a1fe-f42f4543197f",
		Name:         "Rushwood Herbalist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}, Discard a card: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
