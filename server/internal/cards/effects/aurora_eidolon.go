package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aurora Eidolon — Creature — Spirit {3}{W}, 2/2 (Dissension, #5598):
//
//	"{W}, Sacrifice this creature: Prevent the next 3 damage that would be
//	 dealt to any target this turn.
//	 Whenever you cast a multicolored spell, you may return this card from
//	 your graveyard to your hand."
//
// The shield is Goblin Legionnaire's preventNextDamageToTheTarget.
// The return is disEidolonReturn, a graveyard trigger (InGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fc8a155-4eb8-4b86-8ccc-2569630d3bfa",
		Name:         "Aurora Eidolon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W}, Sacrifice this creature: Prevent the next 3 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    Plus(ManaCost("{W}"), SacrificeThis()),
			Targets: TargetAny(),
			Effect:  preventNextDamageToTheTarget("Aurora Eidolon", 3),
		}},
		Triggered: []game.TriggeredAbility{disEidolonReturn("Aurora Eidolon")},
	})
}
