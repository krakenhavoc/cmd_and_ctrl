package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firemane Angel — Creature — Angel {3}{R}{W}{W}, 4/3 (Ravnica: City of Guilds):
//
//	"Flying, first strike
//	 At the beginning of your upkeep, if Firemane Angel is in your graveyard or
//	 on the battlefield, you may gain 1 life.
//	 {6}{R}{R}{W}{W}: Return this card from your graveyard to the battlefield.
//	 Activate only during your upkeep."
//
// The upkeep trigger watches from both zones its text names (CR 113.6,
// Zones {battlefield, graveyard}), so it is the intervening "if" of CR 603.4
// by construction: the card is in one of the two or the ability never
// triggers, and the "may" is asked as the ability resolves. The return is a
// graveyard activation (Reassembling Skeleton's path) gated on the
// controller's own upkeep (DuringYourUpkeep), and it brings the card back
// untapped, as printed.
//
// No simplification.
func init() {
	upkeep := effects.Optional(effects.AtYourUpkeep("Firemane Angel — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		"Firemane Angel — gain 1 life?")
	upkeep.Zones = []game.ZoneKind{game.ZoneBattlefield, game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:        "837e4452-3002-41e3-890b-3d54b469e7b9",
		Name:            "Firemane Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered:       []game.TriggeredAbility{upkeep},
		Activated: []effects.ActivatedAbility{{
			Label:     "{6}{R}{R}{W}{W}: Return this card from your graveyard to the battlefield. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{6}{R}{R}{W}{W}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.DuringYourUpkeep(),
			Effect:    effects.ReturnThisFromYourGraveyardToTheBattlefield,
		}},
	})
}
