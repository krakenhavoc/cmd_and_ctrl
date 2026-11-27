package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disturbing Plot — Sorcery {1}{B} (Shadowmoor, S82 older sets):
//
//	"Return target creature card from a graveyard to its owner's hand.
//	 Conspire (As you cast this spell, you may tap two untapped creatures you
//	 control that share a color with it. When you do, copy it and you may choose a new target for the copy.)"
//
// Endbringer's Revel's target (any graveyard) with Argivian Find's return to the owner's hand.
// Conspire is the engine's optional cost and cast trigger (conspire.go, CR
// 702.78), so the card declares the cost only. The conspire copy is a copy of the spell and may choose a new target (CR 707.10c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "3efecb6c-7c21-4b3b-9612-e3916f91c680",
		Name:          "Disturbing Plot",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
		OptionalCosts: []game.AdditionalCost{effects.Conspire()},
		OnResolve:     effects.ReturnTargetFromGraveyardToHand,
	})
}
