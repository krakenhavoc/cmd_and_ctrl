package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deft Dismissal — Instant {3}{W} (Aether Revolt, S82 older sets, #5536):
//
//	"Deft Dismissal deals 3 damage divided as you choose among one, two, or
//	 three target attacking or blocking creatures."
//
// Flameshot's division (CR 601.2d: each chosen creature gets at least 1; one
// that has left takes nothing and its share is lost, CR 608.2b), limited to
// attacking or blocking creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cbe863a-6654-4f54-b4ec-e1040e830174",
		Name:         "Deft Dismissal",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("one, two, or three target attacking or blocking creatures", effects.AttackingOrBlocking()).
			WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
