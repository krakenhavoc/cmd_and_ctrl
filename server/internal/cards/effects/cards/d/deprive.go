package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deprive — Instant {U}{U} (Rise of the Eldrazi, #5581):
//
//	"As an additional cost to cast this spell, return a land you control to
//	 its owner's hand.
//	 Counter target spell."
//
// The additional cost is Fear of Isolation's return-to-hand clause narrowed
// to lands: the caster names the land as the spell is cast, and with no land
// to return the spell can't be cast. A land tapped for this spell's own mana
// may pay it (CR 601.2g comes before CR 601.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "1065eceb-9150-4ea4-bd6d-acaabda14e00",
		Name:           "Deprive",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.ReturnPermanentToHandCost("a land you control", effects.Land()),
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		Targets:        effects.TargetSpell("target spell"),
		OnResolve:      effects.CounterTheTargetSpell,
	})
}
