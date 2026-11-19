package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kuldotha Rebirth — Sorcery {R}:
//
//	"As an additional cost to cast this spell, sacrifice an artifact.
//	 Create three 1/1 red Goblin creature tokens."
//
// Hordeling Outburst's three Goblins behind Deadly Dispute's sacrifice cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "f3e28778-0149-4dbd-badb-ba5aeacacd76",
		Name:           "Kuldotha Rebirth",
		Purpose:        game.Purpose{Tokens: 3, CreatureTokens: 3},
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact", Artifact()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: RedGoblinToken(), N: 3}.Apply(ctx)
		},
	})
}
