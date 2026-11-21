package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tithing Blade // Consuming Sepulcher — a transforming artifact (#2124,
// ADR 0137):
//
//	Tithing Blade — Artifact {1}{B}
//	  "When this artifact enters, each opponent sacrifices a creature of
//	   their choice.
//	   Craft with creature {4}{B}"
//	Consuming Sepulcher — Artifact
//	  "At the beginning of your upkeep, each opponent loses 1 life and
//	   you gain 1 life."
//
// The enters trigger is the EachPlayerSacrifices fan-out (each opponent
// picks their own creature; nothing targets). Craft is the keyword
// constructor: {4}{B}, exile this artifact, and exile a creature you
// control or a creature card from your graveyard, at sorcery speed; the
// card comes back on its back face as a new object under its owner's
// control. Requested for the Mishra deck (#2033).
//
// No simplification.
const TithingBladeOracleID = "883a4180-9ede-4249-a4b8-3a29c998fb63"

func init() {
	effects.Register(effects.Spec{
		OracleID:     TithingBladeOracleID,
		Name:         "Tithing Blade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerEdict(effects.WhenThisEnters("Tithing Blade — each opponent sacrifices a creature",
				effects.Do(effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Creature(), Label: "a creature"}))),
		},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with creature {4}{B}", "{4}{B}", effects.CraftWith("creature")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     TithingBladeOracleID + "#1",
		Name:         "Consuming Sepulcher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Consuming Sepulcher — each opponent loses 1 life and you gain 1 life", effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1})),
		},
	})
}
