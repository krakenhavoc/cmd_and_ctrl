package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trail of the Mage-Rings — Plane — Vryn (Planechase Anthology, #4870):
//
//	"Instant and sorcery spells have rebound. (The spell's controller
//	 exiles the spell as it resolves if they cast it from their hand. At
//	 the beginning of that player's next upkeep, they may cast that card
//	 from exile without paying its mana cost.)
//	 Whenever chaos ensues, you may search your library for an instant or
//	 sorcery card, reveal it, put it into your hand, then shuffle."
//
// A plane (ADR 0216 §6). The static is Cast Through Time's layer-6 grant
// over spells (spell_keywords.go, whose planar walk reads a face-up
// plane's statics, CR 901.7), but for every player's instants and
// sorceries; rebound itself is the engine's (CR 702.88: only a spell cast
// from its owner's hand rebounds, and the delayed cast outlives the plane,
// the last ruling). The chaos ability is an optional tutor for the planar
// controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2c95c7a-a04c-46ef-a879-6eb08e444093",
		Name:         "Trail of the Mage-Rings",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:         "Instant and sorcery spells have rebound.",
			Layer:         game.Layer6Ability,
			AffectsSpells: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return Or(Instant(), Sorcery())(g, source.Controller, game.ThroughTextOf(source, *target))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, game.KeywordRebound)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Trail of the Mage-Rings — you may search your library for an instant or sorcery card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.IsInstant() || c.IsSorcery() },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Trail of the Mage-Rings — an instant or sorcery card",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
