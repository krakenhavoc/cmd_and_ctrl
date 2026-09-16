package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wood Elves — Creature — Elf Scout {2}{G}, 1/1:
//
//	"When Wood Elves enters, search your library for a Forest card,
//	put it onto the battlefield, then shuffle."
//
// The body is incidental; this is Nature's Lore stapled to a
// creature, which is why it's a staple in green ramp — the land
// enters UNTAPPED and the Elf is a sacrifice-outlet body later.
//
// A Triggered ETB rather than OnResolve: the search fires when the
// creature crosses into the battlefield, so it also triggers off
// reanimation or a blink rather than only off casting — and it uses
// the stack (#578), so it can be answered.
func init() {
	Register(Spec{
		OracleID: "8973bd99-20f8-4867-90ef-50392147ee1b",
		Name:     "Wood Elves",
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wood Elves — search for a Forest card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsLandWithSubtype("forest"),
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Wood Elves — a Forest card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
