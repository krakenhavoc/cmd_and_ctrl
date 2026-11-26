package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warden of the Beyond — Creature — Human Wizard {2}{W}, 2/2 (Magic 2015,
// #5556):
//
//	"Vigilance
//	 This creature gets +2/+2 as long as an opponent owns a card in exile."
//
// A layer 7c static read on every recompute (CR 611.3a): any card in exile
// whose owner is not the Warden's controller counts, face up or face down.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6e9d5e5e-d571-4a1a-be8c-6672fd6b582f",
		Name:            "Warden of the Beyond",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Gets +2/+2 as long as an opponent owns a card in exile.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.WhileYouHave(effects.S82OldB357OpponentOwnsExiledCard, effects.SelfCreature),
			Apply:     effects.PumpBy(2, 2),
		}},
	})
}
