package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Witch Hunt — Enchantment {4}{R}:
//
//	"Players can't gain life.
//	 At the beginning of your upkeep, this enchantment deals 4 damage to
//	 you.
//	 At the beginning of your end step, target opponent chosen at random
//	 gains control of this enchantment."
//
// "Players can't gain life" is ADR 0107 §5's battlefield static
// (CR 119.7). The upkeep damage is ordinary damage from the
// enchantment to its controller.
//
// The end-step target is ChosenAtRandom (#1815): the engine draws it
// uniformly from the controller's opponents as the trigger goes on the
// stack, through the keyed random stream, so players can respond
// knowing who it is (the 2013-10-17 rulings). You are never a legal
// target, and with no opponent left the trigger is removed and you keep
// it. The control change is indefinite (CR 611.2a); an enchantment that
// left and came back is a new object and is not handed over (CR 400.7),
// and if the new controller leaves the game the effect ends with them
// (CR 800.4a). Under its new controller "your upkeep" and "your end
// step" are theirs, so it keeps moving round the table.
//
// No simplification.
func init() {
	endStep := AtYourEndStep("Witch Hunt — target opponent chosen at random gains control of this enchantment",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer || sourceIsNewObject(g, item) {
				return nil
			}
			g.GainControlForEffect(item.SourceCardID, item.SourceCardID, t.ID, game.IndefiniteDuration(),
				"Witch Hunt — target opponent chosen at random gains control of this enchantment")
			return nil
		})
	endStep.Targets = ChosenAtRandom(TargetPlayer("target opponent chosen at random", Opponent()))
	Register(Spec{
		OracleID:     "e86bd38f-7804-449d-af29-21e96a56ab30",
		Name:         "Witch Hunt",
		Completeness: CompletenessFull,
		CantGainLife: PlayersCantGainLife(),
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Witch Hunt — this enchantment deals 4 damage to you", func(g *game.Game, item *game.StackItem) error {
				return DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 4}.Apply(NewContext(g, item))
			}),
			endStep,
		},
	})
}
