package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imaryll, Elfhame Elite — Legendary Creature — Elf Warrior {2}{G}{G}, 3/3
// (Game Night: Free-for-All, #5485):
//
//	"Whenever Imaryll attacks, it gets +X/+X until end of turn, where X is
//	 the number of other Elves you control.
//	 You may have Imaryll assign its combat damage as though it weren't
//	 blocked."
//
// The attack trigger is Herald of Amity's: X is counted once, as the
// trigger resolves, over the Elf permanents the trigger's controller
// controls other than Imaryll, and the bonus is then fixed. "Elf" is a
// word of the source's text (CR 612.2), and the subtype is read after the
// layers, so a changeling counts.
//
// "As though it weren't blocked" is Wolverine, Claws Out's
// Characteristic.AssignsAsThoughUnblocked, an option over the CR 510.1c
// blocked assignment: a layer 6 static on Imaryll sets it on Imaryll
// alone, and the combat damage step asks her controller, for a blocked
// Imaryll, whether to divide her damage among her blockers or put all of
// it on what she is attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef6eb92f-93e6-40e4-a7a6-8e826bb6a5e0",
		Name:         "Imaryll, Elfhame Elite",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "You may have Imaryll assign its combat damage as though it weren't blocked.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Imaryll, Elfhame Elite — it gets +X/+X until end of turn, where X is the number of other Elves you control",
				imaryllPump),
		},
	})
}

// imaryllPump counts the other Elves the trigger's controller controls
// and pumps Imaryll by that many.
func imaryllPump(g *game.Game, item *game.StackItem) error {
	elf := g.TextReaderForEffect(item).TextSubtype("Elf")
	x := countControlled(g, item.Controller, func(c game.Card) bool {
		return c.InstanceID != item.SourceCardID && c.HasSubtype(elf)
	})
	return thisGetsUntilEndOfTurn(x, x, fmt.Sprintf("Imaryll, Elfhame Elite — +%d/+%d until end of turn", x, x))(g, item)
}
