package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vesperlark — Creature — Elemental {2}{W}, 2/1 (Modern Horizons):
//
//	"Flying
//	 When this creature leaves the battlefield, return target creature card
//	 with power 1 or less from your graveyard to the battlefield.
//	 Evoke {1}{W}"
//
// Flying is a printed keyword and evoke is the shared alternative cost
// (CR 702.74a: the sacrifice is a triggered ability, so an evoked Vesperlark
// enters, is sacrificed and its leaves-the-battlefield trigger fires). The
// trigger targets as it goes on the stack (CR 603.3d), and a card with power
// 1 or less is judged by the card's power in the graveyard. Vesperlark's own
// card is in the graveyard by the time the trigger is put on the stack, so
// a 2-power Vesperlark never targets itself, and a target that has left the
// graveyard by resolution is illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "b801a701-15ce-4792-a3d7-d632e7bd9f72",
		Name:             "Vesperlark",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisLeaves("Vesperlark — return target creature card with power 1 or less from your graveyard to the battlefield",
				effects.ReturnFirstLegalGraveyardTargetToBattlefield),
				effects.TargetCardInGraveyard("target creature card with power 1 or less in your graveyard", effects.Creature(), effects.YouOwn(), effects.PowerLE(1))),
		},
	})
}
