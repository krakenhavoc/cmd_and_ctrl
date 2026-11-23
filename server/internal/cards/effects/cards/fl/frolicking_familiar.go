package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frolicking Familiar // Blow Off Steam — Creature — Otter Wizard {2}{U},
// 2/2 // Instant — Adventure {R} (Wilds of Eldraine, #4471):
//
//	Frolicking Familiar — "Flying
//	 Whenever you cast an instant or sorcery spell, this creature gets
//	 +1/+1 until end of turn."
//	Blow Off Steam — "Blow Off Steam deals 1 damage to any target. (Then
//	 exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The cast trigger is prowess without the noncreature reach:
// only an instant or sorcery spell, and the pump goes on this creature
// as it resolves (it does nothing to a Familiar that left and came back,
// CR 400.7). Blow Off Steam is Lightning Bolt's body at 1 damage; the
// Adventure lifecycle is the engine's (CR 715.3d).
//
// No simplification.
const FrolickingFamiliarOracleID = "bbb6867f-7186-41de-9828-21be469178c1"

func init() {
	effects.Register(effects.Spec{
		OracleID:        FrolickingFamiliarOracleID,
		Name:            "Frolicking Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Frolicking Familiar — gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Frolicking Familiar — +1/+1")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     FrolickingFamiliarOracleID + "#1",
		Name:         "Blow Off Steam",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 1)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(1),
	})
}
