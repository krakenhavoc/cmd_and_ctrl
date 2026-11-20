package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stir the Grave — Sorcery {X}{B} (Betrayers of Kamigawa, #5602):
//
//	"Return target creature card with mana value X or less from your
//	 graveyard to the battlefield."
//
// The target clause is bound to the announced X (WithManaValueAtMostX,
// Here Comes a New Hero's binding), so a creature card of greater mana
// value is refused at announce and re-checked as the spell resolves (CR
// 608.2b). X = 0 still returns a creature card with mana value 0, and a
// card with {X} in its own cost counts that X as 0 (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "296dcc57-5651-4cb1-b89c-fcdd3f5311bd",
		Name:         "Stir the Grave",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target creature card with mana value X or less from your graveyard",
			Creature(), YouOwn()).WithManaValueAtMostX(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return returnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
