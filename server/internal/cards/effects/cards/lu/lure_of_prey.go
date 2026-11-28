package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lure of Prey — Instant {2}{G}{G} (Mirage, S82 OLD b798, #5633):
//
//	"Cast this spell only if an opponent cast a creature spell this turn.
//	 You may put a green creature card from your hand onto the battlefield."
//
// The restriction is a cast condition read as the spell is announced (CR
// 601.3), off the per-turn cast tally of each opponent (every spell they cast
// minus the noncreature ones). The put is Elvish Piper's, narrowed to green
// creature cards: a put, never a cast, with a real "you may".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "3c83ea11-7141-41a2-ba77-1ed2722dfe74",
		Name:               "Lure of Prey",
		Completeness:       effects.CompletenessFull,
		Purpose:            game.Purpose{Answers: game.AnswerMakesBlocker},
		CastCondition:      effects.OldB798OpponentCastACreatureSpellThisTurn,
		CastConditionLabel: "Cast this spell only if an opponent cast a creature spell this turn.",
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutFromHandOntoBattlefield{
				Match:    effects.And(effects.Creature(), effects.OfColor("G")),
				Optional: true,
				Label:    "Lure of Prey — you may put a green creature card from your hand onto the battlefield",
			}.Apply(ctx)
		},
	})
}
