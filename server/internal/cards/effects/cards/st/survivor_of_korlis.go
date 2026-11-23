package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Survivor of Korlis — Creature — Human Soldier {W}, 1/1 (The Brothers' War):
//
//	"First strike
//	 {1}{W}, Exile this card from your graveyard: Scry 2."
//
// Campus Crier's graveyard activation: exiling the card is part of the cost
// (CR 118.3), so the scry reads only its controller once the source has gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca7a25f4-c280-48b8-9e04-b15f4c13e569",
		Name:            "Survivor of Korlis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Exile this card from your graveyard: Scry 2.",
			Purpose: game.Purpose{Scry: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.Do(effects.Scry{N: 2}),
		}},
	})
}
