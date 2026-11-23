package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scout for Survivors — Sorcery {2}{W} (Edge of Eternities, #4461):
//
//	"Return up to three target creature cards with total mana value 3 or
//	 less from your graveyard to the battlefield. Put a +1/+1 counter on
//	 each of them."
//
// The targets are one clause under the total set rule
// (TotalManaValueAtMost): their mana values summed may not exceed 3 as
// the spell is cast (CR 601.2c), and the ones still legal are summed
// again as it resolves (CR 608.2b). The return is Smile at Death's body:
// every card still in the graveyard enters together (CR 603.6a), and
// each one that arrived then gets a +1/+1 counter, since the counter is
// the sentence after the return and not an "enters with".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a367c9c-9a7f-4587-aa05-c0fba46ed838",
		Name:         "Scout for Survivors",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("up to three target creature cards with total mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Creature()).WithCount(0, 3).WithTotal(effects.TotalManaValueAtMost(3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B35ReturnChosenToBattlefieldWithCounter(ctx.Game, item)
		},
	})
}
