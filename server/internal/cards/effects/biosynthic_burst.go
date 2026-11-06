package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Biosynthic Burst — Instant {1}{G} (Edge of Eternities, #4461):
//
//	"Put a +1/+1 counter on target creature you control. It gains
//	 reach, trample, and indestructible until end of turn. Untap it.
//	 (Damage and effects that say "destroy" don't destroy it.)"
//
// Battlefield Promotion's steps (plusOneCounterThenKeywordsOnTheTarget,
// ResolveSteps, CR 608.2c) with three keywords and an untap after them:
// the counter first, then the keywords until end of turn, then the
// untap, each step waiting for any prompt the one before it asked and
// each skipped when the creature is no longer a legal target
// (CR 608.2b). Indestructible keeps it on the battlefield through a
// Wrath, so the spell is declared as survive as well as pump and a
// combat grant.
//
// No simplification.
func init() {
	steps := append(plusOneCounterThenKeywordsOnTheTarget("Biosynthic Burst — reach, trample and indestructible until end of turn",
		"reach", "trample", "indestructible"),
		Instruction(func(ctx *Context) error {
			if id := firstLegalCardTarget(ctx); id != uuid.Nil {
				return UntapTarget{Target: id}.Apply(ctx)
			}
			return nil
		}))
	Register(Spec{
		OracleID:     "3abd2342-f783-4871-ac78-1a54b35722c2",
		Name:         "Biosynthic Burst",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		Purpose: game.Purpose{
			Answers: game.AnswerPump | game.AnswerSurvive | game.AnswerCombatGrant,
			Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "reach", "trample", "indestructible")),
		},
		OnResolve: ResolveSteps(steps...),
	})
}
