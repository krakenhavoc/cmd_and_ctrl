package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghostly Keybearer — Creature — Spirit {3}{U}, 3/3:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, unlock a
//	 locked door of up to one target Room you control."
//
// The trigger's "up to one target" is a zero-to-one clause, so it goes
// on the stack with no Room to unlock when you control none. The door
// is chosen by UnlockADoor (rooms.go): with one locked door there is
// nothing to ask, with two the controller picks, and a fully unlocked
// Room is a legal target that simply does nothing (CR 709.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec86a45f-046c-4aab-9b82-422c6f39142e",
		Name:            "Ghostly Keybearer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Ghostly Keybearer — unlock a locked door of up to one target Room you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return effects.UnlockADoor{Room: t.ID, Player: item.Controller}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetPermanent("up to one target Room you control", effects.HasSubtype("Room"), effects.YouControl()).WithCount(0, 1)),
		},
	})
}
