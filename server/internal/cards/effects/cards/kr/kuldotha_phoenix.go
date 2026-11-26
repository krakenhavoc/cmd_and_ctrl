package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kuldotha Phoenix — Creature — Phoenix {2}{R}{R}{R}, 4/4 (Scars of Mirrodin,
// S82 OLD batch 469, #5577):
//
//	"Flying, haste
//	 Metalcraft — {4}: Return this card from your graveyard to the
//	 battlefield. Activate only during your upkeep and only if you control
//	 three or more artifacts."
//
// Cauldron Familiar's graveyard return (ReturnThisFromGraveyardToBattlefield)
// with both gates checked as it is activated (CR 602.5b): your own upkeep and
// metalcraft. A card that has left the graveyard by resolution is a new object
// and nothing happens (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "884260cf-b3f3-43a4-82b6-4f6a0b19cc75",
		Name:            "Kuldotha Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:     "Metalcraft — {4}: Return this card from your graveyard to the battlefield. Activate only during your upkeep and only if you control three or more artifacts.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{4}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.AllConditions(effects.DuringYourUpkeep(), effects.OldB469MetalcraftActivation()),
			Effect:    effects.ReturnThisFromGraveyardToBattlefield,
		}},
	})
}
