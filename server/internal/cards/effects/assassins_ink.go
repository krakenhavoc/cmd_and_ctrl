package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assassin's Ink — Instant {2}{B}{B} (Kamigawa: Neon Dynasty, #5493):
//
//	"This spell costs {1} less to cast if you control an artifact and {1} less
//	 to cast if you control an enchantment.
//	 Destroy target creature or planeswalker."
//
// Two self cost modifiers (CR 601.2f), one per sentence half, each a Cactarantula
// discount; both apply when both are met. A permanent that is both an artifact
// and an enchantment satisfies both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6d95d26-48d8-4a97-9a42-f492a61300c9",
		Name:         "Assassin's Ink",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you control an artifact.", YouControlA(Artifact())),
			CostsLess(1, "This spell costs {1} less to cast if you control an enchantment.", YouControlA(Enchantment())),
		},
		Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
