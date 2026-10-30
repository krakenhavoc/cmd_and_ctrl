package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wondrous Revival — Sorcery {4}{W}{W} (Marvel Super Heroes Commander,
// #4455):
//
//	"Return up to three target Hero creature cards from your graveyard to
//	 the battlefield."
//
// Reveillark's return (b24ReturnGraveyardTargetsToBattlefield): the
// still-legal targets enter together. Hero is a creature subtype here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19cab267-3c51-4322-b1af-22941bbbdadf",
		Name:         "Wondrous Revival",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to three target Hero creature cards from your graveyard",
			YouOwn(), Creature(), Subtype("Hero")).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b24ReturnGraveyardTargetsToBattlefield(ctx, 3)
		},
	})
}
