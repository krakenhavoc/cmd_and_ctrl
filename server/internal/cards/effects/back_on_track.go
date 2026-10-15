package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Back on Track — Sorcery {4}{B}:
//
//	"Return target creature or Vehicle card from your graveyard to the
//	 battlefield. Create a 1/1 colorless Pilot creature token with
//	 "This token saddles Mounts and crews Vehicles as though its power
//	 were 2 greater.""
//
// One target, so a spell whose only target is gone at resolution does
// nothing at all, token included (CR 608.2b); the engine fizzles it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aaf46eaf-2e01-46ac-b79d-6b70739d11c2",
		Name:         "Back on Track",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target creature or Vehicle card in your graveyard",
			Or(Creature(), HasSubtype("Vehicle")), YouOwn()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if _, ok := reanimateSingleTarget(ctx, item.Controller); !ok {
				return nil
			}
			return CreateToken{Template: PilotToken(), N: 1}.Apply(ctx)
		},
	})
}
