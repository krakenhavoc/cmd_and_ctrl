package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Morbid Plunder — Sorcery {1}{B}{B} (Mirrodin Besieged,
// #5576):
//
//	"Return up to two target creature cards from your graveyard to your
//	 hand."
//
// Macabre Reconstruction's return without its discount: up to two creature
// cards from your own graveyard, each returned if still there as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bb7009b-aeb0-4545-be4c-c8c5d723bcf5",
		Name:         "Morbid Plunder",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("up to two target creature cards from your graveyard", Creature(), YouOwn()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return returnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
