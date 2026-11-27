package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the Maw of Hell — {4}{R}{R} Sorcery (Innistrad, S82 OLD b447, #5572):
//
//	"Destroy target land. Into the Maw of Hell deals 13 damage to target
//	 creature."
//
// Two target clauses, each judged again as the spell resolves (CR 608.2b): if
// only one is still legal the other half still happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f0e8fd5-9db9-4507-a42f-cb99080b89e2",
		Name:         "Into the Maw of Hell",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetPermanent("target land", effects.Land()),
			effects.TargetCreature("target creature"),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			damage := ctx.ItemTail(intoTheMawOfHellDamage)
			if land, found := ctx.ClauseTarget(0); found && land.Kind == game.TargetCard && ctx.IsTargetLegal(land) {
				return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{land.ID}, damage)
			}
			return damage.Run(ctx.Game, nil)
		},
	})
}

// intoTheMawOfHellDamage is the second sentence, run once the land's destruction
// has settled (a replacement may pause it): 13 damage to the creature if it is
// still a legal target. An on-disk key: never renamed, never reused.
var intoTheMawOfHellDamage = effects.CardTail("into-the-maw-of-hell/damage", func(ctx *effects.Context, _ game.TailOutcome) error {
	if t, found := ctx.ClauseTarget(1); found && t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
		return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 13}.Apply(ctx)
	}
	return nil
})
