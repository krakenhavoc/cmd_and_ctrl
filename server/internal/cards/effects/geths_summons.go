package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geth's Summons — Sorcery {2}{B}{B}:
//
//	"Return up to one target creature card from your graveyard to the
//	 battlefield.
//	 Corrupted — For each opponent who has three or more poison
//	 counters as you cast this spell, put up to one target creature
//	 card from that player's graveyard onto the battlefield under your
//	 control."
//
// The first clause is ordinary. The corrupted clauses are one per
// possible opponent, bound to that seat (perPoisonedOpponentGraveyard-
// Clauses, #5860): a clause offers that opponent's graveyard only if
// they had three or more poison counters as the spell was cast. That
// is judged live while the targets are announced (CR 601.2c) and fixed
// on the item as the cast completes (CR 601.2i), so an opponent who
// loses their poison in response keeps their target and one who gains
// it gets none; the CR 608.2b re-check asks only whether the card is
// still in that player's graveyard.
//
// The two sentences run in order. The corrupted cards enter together
// (ReturnFromGraveyardTogether), so each sees the others enter.
//
// No simplification.
func init() {
	corrupted := perPoisonedOpponentGraveyardClauses(3, "creature card", Creature())
	Register(Spec{
		OracleID:     "c345be5e-243d-44c8-a7c7-280bdb235223",
		Name:         "Geth's Summons",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("up to one target creature card from your graveyard", Creature(), YouOwn()).WithCount(0, 1),
			corrupted...,
		),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return returnClauseCardFromGraveyard(ctx, 0, game.ZoneBattlefield)
			}),
			Instruction(func(ctx *Context) error {
				var ids []uuid.UUID
				for _, t := range ctx.LegalTargets() {
					if t.Slot > 0 && t.Kind == game.TargetCard {
						ids = append(ids, t.ID)
					}
				}
				if len(ids) == 0 {
					return nil
				}
				return ReturnFromGraveyardTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
			}),
		),
	})
}
