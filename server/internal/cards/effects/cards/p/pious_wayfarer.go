package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pious Wayfarer — Creature — Human Scout {W}, 1/2 (Jumpstart 2022):
//
//	"Constellation — Whenever an enchantment you control enters, target
//	 creature gets +1/+1 until end of turn."
//
// Constellation without the "or this creature" half (Captivating Unicorn's
// condition: the Wayfarer is not an enchantment). The target is chosen as the
// trigger goes on the stack; one that left or stopped being legal by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8790c7be-c14a-4293-bc4a-3d0efe42ed79",
		Name:         "Pious Wayfarer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.IsEnchantment()
				}, "Pious Wayfarer — target creature gets +1/+1 until end of turn (constellation)",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetCard {
								return effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1,
									Label: "Pious Wayfarer — +1/+1 until end of turn"}.Apply(ctx)
							}
						}
						return nil
					}),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
