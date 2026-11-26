package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mine Excavation — Sorcery {1}{W} (Shadowmoor, S82 older sets):
//
//	"Return target artifact or enchantment card from a graveyard to its owner's hand.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
//
// Argivian Find's return over any graveyard.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy is a copy of the spell and may choose a new target (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9bd5dc7a-11b0-48eb-bf52-2956dd3989ab",
		Name:          "Mine Excavation",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCardInGraveyard("target artifact or enchantment card from a graveyard", effects.Or(effects.Artifact(), effects.Enchantment())),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve:     effects.ReturnTargetFromGraveyardToHand,
	})
}
