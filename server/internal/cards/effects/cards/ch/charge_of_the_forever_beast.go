package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charge of the Forever-Beast — Sorcery {2}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"As an additional cost to cast this spell, reveal a creature card from your
//	 hand.
//	 Charge of the Forever-Beast deals damage to target creature or planeswalker
//	 equal to the revealed card's power."
//
// The reveal is a cost (CR 601.2h), so a hand with no creature card can't pay
// it and the spell can't be cast. The card stays in hand; its power is the
// printed one it was revealed with, which is why the payment records it (CR
// 608.2b: the damage reads the revealed card, not what the card becomes). A
// creature card whose power is a star or negative deals no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "937f0c60-003e-446d-8889-059f7b4a2dd6",
		Name:           "Charge of the Forever-Beast",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.S82OldB116RevealACreatureCardCost(),
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve:      effects.S82OldB116DamageEqualToTheChosenPower,
	})
}
