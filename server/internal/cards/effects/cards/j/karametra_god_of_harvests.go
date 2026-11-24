package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karametra, God of Harvests — Legendary Enchantment Creature — God
// {3}{G}{W}, 6/7 (Secret Lair Drop):
//
//	"Indestructible
//	 As long as your devotion to green and white is less than seven,
//	 Karametra isn't a creature.
//	 Whenever you cast a creature spell, you may search your library for a
//	 Forest or Plains card, put it onto the battlefield tapped, then shuffle."
//
// The God clause is the shared Theros static (CR 700.5). The trigger is an
// optional one (CR 603.5): the "may" is asked as it resolves, and a yes
// searches for a card with the Forest or Plains subtype (basic or not), which
// may also come up empty. The land enters tapped, and the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "984e763c-5a83-4231-8ec4-1be76846479a",
		Name:            "Karametra, God of Harvests",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Purpose:         game.Purpose{Tutors: 1, Lands: 1},
		Static:          []game.StaticAbility{effects.GodUnlessDevotionToColors(7, "G", "W")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.YouCastACreatureSpell,
				"Karametra, God of Harvests — you may search for a Forest or Plains card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return c.IsLand() && (c.HasSubtypeWord("Forest") || c.HasSubtypeWord("Plains"))
						},
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Reason:        "Karametra, God of Harvests — a Forest or Plains card",
					}.Apply(effects.NewContext(g, item))
				}), "Karametra, God of Harvests — search for a Forest or Plains card?"),
		},
	})
}
