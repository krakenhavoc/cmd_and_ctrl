package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vacuumelt — Sorcery {2}{U} (Duel Decks: Izzet vs. Golgari):
//
//	"Replicate {2}{U} (When you cast this spell, copy it for each time you paid
//	 its replicate cost. You may choose new targets for the copies.)
//	 Return target creature to its owner's hand."
//
// Shattering Spree's replicate (CR 702.56a) over a bounce: each copy returns
// its own target, if it is still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "eb421a1b-147f-422b-bb8d-483a182a4096",
		Name:          "Vacuumelt",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.ReplicateMana("{2}{U}", 20)},
		Triggered:     []game.TriggeredAbility{effects.Replicate()},
		OnResolve:     effects.BounceTheTarget,
	})
}
