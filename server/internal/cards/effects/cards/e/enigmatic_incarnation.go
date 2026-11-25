package cardse

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Enigmatic Incarnation — Enchantment {2}{G}{U} (Theros Beyond Death, S82 OLD,
// #5514):
//
//	"At the beginning of your end step, you may sacrifice another enchantment.
//	 If you do, search your library for a creature card with mana value equal
//	 to 1 plus the sacrificed enchantment's mana value, put that card onto the
//	 battlefield, then shuffle."
//
// Ziatora's end-step "you may sacrifice another X. If you do" (MaySacrificeOne:
// asked as the trigger resolves, over whatever you control then, and never
// the Incarnation itself). The search runs only for a sacrifice that really
// happened, at 1 plus the sacrificed enchantment's mana value as it last
// existed on the battlefield (CR 608.2h): a token, or a copy of nothing, has
// mana value 0, so it finds a 1. Neoform's search follows: the library is
// searched and shuffled even when no card matches (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "915091d5-3f58-4429-8f3a-0fd7aef87afe",
		Name:         "Enigmatic Incarnation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Enigmatic Incarnation — you may sacrifice another enchantment; if you do, search your library for a creature card with mana value 1 plus its mana value",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Enigmatic Incarnation — sacrifice another enchantment to search for a creature?", effects.Enchantment(), true, enigmaticIncarnationSacrificed)
				}),
		},
	})
}

// enigmaticIncarnationSacrificed is the search once an enchantment has really
// gone. An on-disk key.
var enigmaticIncarnationSacrificed = effects.RegisterSacrificedOne("prompt-run/enigmatic-incarnation-sacrificed", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	mv := 0
	if info, ok := ctx.Game.LastKnownPermanentForEffect(sacrificed); ok {
		mv = info.ManaValue
	}
	mv++
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() == mv },
		Dest:      game.ZoneBattlefield,
		Limit:     1,
		Shuffle:   true,
		Reason:    "Enigmatic Incarnation — a creature card with mana value " + strconv.Itoa(mv),
	}.Apply(ctx)
})
