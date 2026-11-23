package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Livaan, Cultist of Tiamat — Legendary Creature — Dragon Shaman {2}{R}, 1/3
// (Commander Legends: Battle for Baldur's Gate):
//
//	"Whenever you cast a noncreature spell, target creature gets +X/+0 until end of
//	 turn, where X is that spell's mana value.
//	 Choose a Background (You can have a Background as a second commander.)"
//
// A cast trigger over a noncreature spell with a target creature chosen as it goes
// on the stack (CR 603.3d). X is the mana value of the spell as it was cast, X in
// its cost included (CR 202.3e), carried on the cast event; the pump is pinned to the
// target (CR 611.2c) and the target is re-checked as the trigger resolves (CR 608.2b).
// Choose a Background is a deck-building rule the deck validator applies
// (deck/partner.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee565d5c-7a3a-481c-91aa-b982cf384220",
		Name:         "Livaan, Cultist of Tiamat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.Noncreature(), "Livaan, Cultist of Tiamat — target creature gets +X/+0 until end of turn, where X is that spell's mana value",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok || item.Trigger == nil {
							return nil
						}
						return effects.BoostUntilEOT{Target: id, Power: item.Trigger.Event.Amount, Label: "Livaan, Cultist of Tiamat — +X/+0"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
		},
	})
}
