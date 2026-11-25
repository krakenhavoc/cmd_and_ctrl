package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mutiny — Sorcery {R} (Rivals of Ixalan, S82 OLD batch 217, #5531):
//
//	"Target creature an opponent controls deals damage equal to its power to
//	 another target creature that player controls."
//
// Breaking of the Fellowship's shape without the Ring: the second target is
// related to the first (ControlledByTargetsController), so it must be
// controlled by whoever controls the first, as the spell is cast and again as
// it resolves (CR 608.2b). The first creature's current power is read as it
// resolves and it is the source of the damage (so deathtouch and lifelink are
// its); with either target illegal, nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "647b4cb4-569d-4938-b56c-31a70e68ee41",
		Name:         "Mutiny",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			effects.Distinct(effects.ControlledByTargetsController(effects.TargetCreature("another target creature that player controls"), 0)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			biter, biterOK := ctx.ClauseTarget(0)
			victim, victimOK := ctx.ClauseTarget(1)
			if !biterOK || !victimOK || !ctx.IsTargetLegal(biter) || !ctx.IsTargetLegal(victim) {
				return nil
			}
			return effects.EachDealsDamageEqualToItsPower(ctx, []game.TargetRef{biter}, victim.ID)
		},
	})
}
