package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exorcise — Sorcery {1}{W} (Duskmourn: House of Horror, #4466):
//
//	"Exile target artifact, enchantment, or creature with power 4 or
//	 greater."
//
// "With power 4 or greater" binds to the creature only: any artifact or
// enchantment is a legal pick, and a creature is one when its current
// power is 4 or more. The power is judged at announce and again at
// resolution (CR 608.2b), so a creature shrunk below 4 in response is
// no longer a legal target and the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "58916455-844c-4d97-908f-8205bdc117c0",
		Name:         "Exorcise",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target artifact, enchantment, or creature with power 4 or greater",
			Or(Artifact(), Enchantment(), And(Creature(), PowerGE(4)))),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
