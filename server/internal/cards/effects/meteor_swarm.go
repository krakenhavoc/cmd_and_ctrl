package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meteor Swarm — Sorcery {X}{R}{R}{R}:
//
//	"Meteor Swarm deals 8 damage divided as you choose among X target
//	 creatures and/or planeswalkers."
//
// Exactly X targets (the announced X, CR 107.3a), and the fixed 8
// divided among them with at least 1 each (CR 601.2d). A target that
// left or stopped qualifying takes nothing and its share is not moved
// (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "3e7bd846-7295-4590-a419-c97cd2526bc1",
		Name:         "Meteor Swarm",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      neoB15XTargets().Dividing(Divide(8)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DealDividedDamage(ctx)
		},
	})
}
