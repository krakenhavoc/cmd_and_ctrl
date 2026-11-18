package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// March of Otherworldly Light — Instant {X}{W} (Kamigawa: Neon Dynasty):
//
//	"As an additional cost to cast this spell, you may exile any number
//	 of white cards from your hand. This spell costs {2} less to cast for
//	 each card exiled this way.
//	 Exile target artifact, creature, or enchantment with mana value X or
//	 less."
//
// The cost is March of Swirling Mist's. The target is bound to the
// announced X (WithManaValueAtMostX, re-checked on resolution, CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "ed06e421-480c-40c8-8c3e-1e466e5327b2",
		Name:           "March of Otherworldly Light",
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: ExileAnyNumberFromHandCost("white cards", "W"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerExiledFromHand("{2}", "This spell costs {2} less to cast for each card exiled this way"),
		},
		Targets: TargetPermanent("target artifact, creature, or enchantment with mana value X or less",
			Or(Artifact(), Creature(), Enchantment())).WithManaValueAtMostX(),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ExileTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
