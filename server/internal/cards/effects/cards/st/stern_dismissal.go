package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stern Dismissal — Instant {U} (Theros Beyond Death, #5514):
//
//	"Return target creature or enchantment an opponent controls to
//	 its owner's hand."
//
// Void Squall's bounce with this card's clause: a creature or an
// enchantment an opponent controls goes to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "319d5cf8-ab72-401e-b04e-c13a7a6c1aac",
		Name:         "Stern Dismissal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target creature or enchantment an opponent controls",
			effects.Or(effects.Creature(), effects.Enchantment()), effects.OpponentControls()),
		OnResolve: effects.BounceTheTarget,
	})
}
