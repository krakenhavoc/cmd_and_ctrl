package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marina Vendrell — Legendary Creature — Human Warlock {W}{U}{B}{R}{G},
// 3/5:
//
//	"When Marina Vendrell enters, reveal the top seven cards of your
//	 library. Put all enchantment cards from among them into your hand
//	 and the rest on the bottom of your library in a random order.
//	 {T}: Lock or unlock a door of target Room you control. Activate
//	 only as a sorcery."
//
// The enter trigger is Goblin Ringleader's shape over seven cards:
// RevealTopOfLibraryForEffect reveals them to the table, and
// TakeFromLibraryToHand takes every enchantment (no prompt, nothing to
// decide) and sends the rest to the bottom in a random order. The tap
// ability is Keys to the House's second ability (lockOrUnlockTargetRoom),
// and it needs the Marina to have been under your control since the turn
// began, like any {T} ability of a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5aafc6c3-14f2-45b1-bd5b-c27760d791dd",
		Name:         "Marina Vendrell",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Marina Vendrell — reveal the top seven cards; put the enchantments into your hand, the rest on the bottom in a random order",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.TakeFromLibraryToHand{
						Player: item.Controller,
						Cards:  g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 7, "Marina Vendrell — reveal the top seven cards"),
						Match:  effects.Enchantment(),
						All:    true,
						Label:  "Marina Vendrell — enchantment cards into your hand",
						Key:    effects.TakeRestBottomRandom,
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}: Lock or unlock a door of target Room you control. Activate only as a sorcery.",
			Cost:         effects.TapCost(),
			Targets:      effects.TargetPermanent("target Room you control", effects.HasSubtype("Room"), effects.YouControl()),
			SorcerySpeed: true,
			Effect:       effects.LockOrUnlockTargetRoom,
		}},
	})
}
