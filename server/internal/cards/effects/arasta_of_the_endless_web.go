package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arasta of the Endless Web — Legendary Enchantment Creature — Spider
// {2}{G}{G}, 3/5 (EDHREC rank 1015):
//
//	"Reach
//	 Whenever an opponent casts an instant or sorcery spell, create a
//	 1/2 green Spider creature token with reach."
//
// The anti-spellslinger enchantress body: every opposing cantrip is a
// Spider. Guttersnipe's cast condition with the actor test flipped —
// an OPPONENT's instant or sorcery — and the trigger goes on the stack
// above the spell that caused it, so the Spider is in play before the
// spell resolves, as in paper. Reach on the token is real (token
// Keywords feed the layer engine).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "695eea46-1535-48c5-bbb6-0b8379e77bfc",
		Name:            "Arasta of the Endless Web",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b09OpponentCastInstantOrSorcery(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Arasta of the Endless Web — create a 1/2 Spider with reach",
					func(g *game.Game, item *game.StackItem) error {
						return CreateToken{Controller: item.Controller, Template: b09GreenSpiderToken(), N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
