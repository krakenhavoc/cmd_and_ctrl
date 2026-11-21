package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Architect of the Untamed — Creature — Elf Artificer Druid {2}{G}, 2/3:
//
//	"Landfall — Whenever a land you control enters, you get {E} (an
//	 energy counter).
//	 Pay eight {E}: Create a 6/6 colorless Beast artifact creature
//	 token."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be6cab52-b3ae-46a2-a581-2468a49bff25",
		Name:         "Architect of the Untamed",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LandfallPaying("Architect of the Untamed — you get {E}", game.LandfallPayoff{Energy: 1}, effects.EbYouGetEnergy(1)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay eight {E}: Create a 6/6 colorless Beast artifact creature token.",
			Cost:    effects.PayEnergy(8),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Effect: effects.Do(effects.CreateToken{N: 1, Template: game.Card{
				Name:      "Beast",
				TypeLine:  "Token Artifact Creature — Beast",
				Power:     6,
				Toughness: 6,
			}}),
		}},
	})
}
