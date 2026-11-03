package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Burrog Barrage — Instant {1}{G}, Secrets of Strixhaven (#4456):
//
//	"Target creature you control gets +1/+0 until end of turn if you've cast
//	 another instant or sorcery spell this turn. Then it deals damage equal to
//	 its power to up to one target creature an opponent controls."
//
// Two target clauses read positionally (Bridgeworks Battle's shape): slot 0 is
// the creature you control, slot 1 is "up to one" creature an opponent
// controls. The +1/+0 comes first and stays when the damage cannot happen;
// the damage reads the layered power after the pump, with the creature as the
// source (so deathtouch and lifelink carry). A creature of yours that left in
// response deals nothing; a recipient that left takes nothing (CR 608.2b). The
// condition is the cast tally (socB21CastAnotherInstantOrSorcery).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c47caa45-e7fc-414a-8a5d-59b58f1751f3",
		Name:         "Burrog Barrage",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(BenefitUntilEOT(0, 1, 0)),
		},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1),
		),
		OnResolve: socB21BurrogBarrage,
	})
}
