package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fear of Abduction — Enchantment Creature — Nightmare {4}{W}{W}, 5/5
// (Duskmourn):
//
//	"As an additional cost to cast this spell, exile a creature you control.
//	 Flying
//	 When this creature enters, exile target creature an opponent controls.
//	 When this creature leaves the battlefield, put each card exiled with it
//	 into its owner's hand."
//
// Two exiles are "exiled with it". The cost's creature is exiled with the
// spell on the stack, linked to the spell object, and the permanent the
// spell becomes inherits the link (CR 607.2q, relinkCostExilesLocked). The
// enters trigger's exile is linked to the permanent (ExileTarget.Linked,
// CR 607.2a). The leave trigger reads every card still linked to the object
// that left, as last-known information, and returns each to its OWNER's hand
// (the creature you exiled, and the opponent's). A card that has since left
// exile has lost the link; a Fear that left before the enters trigger
// resolved still exiles its target (that card is then linked to an object
// that is gone and stays exiled), as the rulings say. A countered Fear never
// became a permanent, so the cost's creature stays exiled.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8f5495d7-ac44-47cd-aa39-48a35034b390",
		Name:            "Fear of Abduction",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AdditionalCost:  effects.ExileCreatureYouControlCost(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Fear of Abduction — exile target creature an opponent controls", effects.FearOfAbductionExile),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			),
			effects.WhenThisLeaves("Fear of Abduction — put each card exiled with it into its owner's hand",
				effects.B36ReturnCardsExiledWithToOwnersHands),
		},
	})
}
