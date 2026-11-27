package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hammer of Bogardan — Sorcery {1}{R}{R} (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"Hammer of Bogardan deals 3 damage to any target.
//	 {2}{R}{R}{R}: Return this card from your graveyard to your hand. Activate
//	 only during your upkeep."
//
// Lightning Bolt's damage for three, and Bestial Bloodline's graveyard return
// restricted to your own upkeep (Augur il-Vec's condition).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3089bdd7-87e6-4d0e-ad98-c32e6f861c30",
		Name:         "Hammer of Bogardan",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(3),
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}{R}{R}{R}: Return this card from your graveyard to your hand. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.ManaCost("{2}{R}{R}{R}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.DuringYourUpkeep(),
			Effect:    effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
