package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drana's Emissary — Creature — Vampire Cleric Ally {1}{W}{B}, 2/2 (Battle
// for Zendikar, #5547):
//
//	"Flying
//	 At the beginning of your upkeep, each opponent loses 1 life and you gain
//	 1 life."
//
// Not a rally card: the drain is every upkeep, not on an Ally entering.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef49b53d-e5b6-423a-88f0-51a7b2021c46",
		Name:            "Drana's Emissary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Drana's Emissary — each opponent loses 1 life and you gain 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
