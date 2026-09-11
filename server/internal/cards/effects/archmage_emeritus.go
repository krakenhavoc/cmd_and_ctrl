package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archmage Emeritus — Creature — Human Wizard {2}{U}{U}, 2/2 (EDHREC
// rank 255):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery
//	 spell, draw a card."
//
// The spellslinger deck's card-draw engine; Storm-Kiln Artist's
// trigger with a draw where the Treasure was. The trigger goes on
// the stack above the spell that caused it and resolves first, so
// the card is in hand before the spell resolves — as in paper.
//
// Sandbox simplification: the "or COPY" half is not implemented —
// the engine has no spell-copy event. Weaker than printed until
// spell copying exists; Storm-Kiln Artist declares the same gap.
func init() {
	Register(Spec{
		OracleID:     "8305d576-21d8-4ce7-8eda-a7cd9793aca5",
		Name:         "Archmage Emeritus",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The draw trigger only fires on instants and sorceries you cast, not on copies of them."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b02CastInstantOrSorcery(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Archmage Emeritus — draw a card (magecraft)",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
