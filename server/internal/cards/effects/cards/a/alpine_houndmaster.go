package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpine Houndmaster — Creature — Human Warrior {R}{W}, 2/2 (Core Set 2021,
// #5511):
//
//	"When this creature enters, you may search your library for a card named
//	 Alpine Watchdog and/or a card named Igneous Cur, reveal them, put them
//	 into your hand, then shuffle.
//	 Whenever this creature attacks, it gets +X/+0 until end of turn, where X
//	 is the number of other attacking creatures."
//
// The search takes up to two cards, one per name (the different-names set
// rule refuses two of the same name), and shuffles even when nothing is
// found (CR 701.24d). The pump counts the other attackers as the ability
// resolves, so an attacker removed from combat in between is not counted.
//
// No simplification.
func init() {
	const label = "Alpine Houndmaster — you may search your library for a card named Alpine Watchdog and/or a card named Igneous Cur"
	effects.Register(effects.Spec{
		OracleID:     "88939cb3-5f61-4857-95cb-9446c6c1f1aa",
		Name:         "Alpine Houndmaster",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.Name == "Alpine Watchdog" || c.Name == "Igneous Cur"
					},
					Dest:     game.ZoneHand,
					Limit:    2,
					Optional: true,
					Reveal:   true,
					Shuffle:  true,
					Rule:     effects.SearchRuleDifferentNames,
					Reason:   "Alpine Houndmaster — a card named Alpine Watchdog and/or a card named Igneous Cur",
				}.Apply(effects.NewContext(g, item))
			}),
			effects.WheneverThisAttacks("Alpine Houndmaster — +X/+0 until end of turn, X the number of other attacking creatures",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := len(effects.S82OtherAttackers(g, ctx.Source()))
					if x == 0 || !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: x, Label: "Alpine Houndmaster — +X/+0"}.Apply(ctx)
				}),
		},
	})
}
