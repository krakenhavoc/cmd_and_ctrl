package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const frolickingFamiliarOracleID = "bbb6867f-7186-41de-9828-21be469178c1"

func init() {
	Register(Spec{
		OracleID:        frolickingFamiliarOracleID,
		Name:            "Frolicking Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Frolicking Familiar — gets +1/+1 until end of turn",
				thisGetsUntilEndOfTurn(1, 1, "Frolicking Familiar — +1/+1")),
		},
	})
	Register(Spec{
		OracleID:     frolickingFamiliarOracleID + "#1",
		Name:         "Blow Off Steam",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 1)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(1),
	})
}
