package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperial Recovery Unit — Artifact — Vehicle {2}{W}, 3/4 (Kamigawa: Neon
// Dynasty, S82):
//
//	"Whenever this Vehicle attacks, return target creature or Vehicle card
//	 with mana value 2 or less from your graveyard to your hand.
//	 Crew 2"
//
// Mirkwood Elk's graveyard-to-hand trigger with Back on Track's creature or
// Vehicle card filter and a mana value cap. A trigger whose one target has
// gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	trigger := effects.WheneverThisAttacks("Imperial Recovery Unit — return target creature or Vehicle card with mana value 2 or less from your graveyard to your hand",
		effects.ReturnTargetGraveyardCardToHand)
	trigger.Targets = effects.TargetCardInGraveyard("target creature or Vehicle card with mana value 2 or less in your graveyard",
		effects.YouOwn(), effects.Or(effects.Creature(), effects.HasSubtype("Vehicle")), effects.ManaValueLE(2))
	effects.Register(effects.Spec{
		OracleID:     "18ad0dd1-2d0e-46b4-8ab1-ece07180267f",
		Name:         "Imperial Recovery Unit",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Imperial Recovery Unit"),
		}},
	})
}
