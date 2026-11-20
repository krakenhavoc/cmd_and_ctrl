package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Void Snare — Sorcery {U} (Magic 2015, #5556):
//
//	"Return target nonland permanent to its owner's hand."
//
// Disperse at sorcery speed (bounceTheTargetToHand); a target that left
// in response is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06e2ce8e-7341-4448-8625-cb1574eb179d",
		Name:         "Void Snare",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return bounceTheTargetToHand(ctx.Game, item)
		},
	})
}
