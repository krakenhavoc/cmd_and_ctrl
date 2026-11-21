package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undercover Operative — Creature — Shapeshifter Rogue {2}{U}{U}, 0/0:
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it enters with a shield counter on it if you
//	 control that creature. (If it would be dealt damage or destroyed,
//	 remove a shield counter from it instead.)"
//
// Clone's entry replacement with an except clause that reads WHAT was
// copied (ReplacementEvent.CopySourceForEffect): the shield counter rides
// the entry (ev.AddCounterAtETB), so it is on the permanent before anything
// sees it enter, and only when the copied creature is controlled by the
// player who controls the entering Operative. Copying an opponent's creature
// enters with no counter. Declining, or having nothing to copy, leaves it a
// 0/0 that dies (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0bd57503-b544-45f8-bcd2-f5b2170010f6",
		Name:         "Undercover Operative",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Undercover Operative",
				effects.AnyCreatureOnBattlefield,
				func(ev *game.ReplacementEvent, _ *game.PrintedValues, g *game.Game, src *game.Card) {
					copied, ok := g.LookupCardForEffect(ev.CopySourceForEffect())
					if !ok {
						return
					}
					controller, _ := effects.CopyChooserAndSelf(ev, src)
					if copied.Controller == controller {
						ev.AddCounterAtETB(game.CounterShield, 1)
					}
				},
			),
		},
	})
}
