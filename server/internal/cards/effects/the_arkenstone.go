package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Arkenstone // Seek the Heart — an adventure card (CR 715):
//
//	The Arkenstone — Legendary Artifact {5}
//	  "Creatures you control get +1/+1.
//	   At the beginning of your end step, draw a card."
//	Seek the Heart — Sorcery — Adventure {2}{W}
//	  "Search your library for a legendary creature card, reveal it, put
//	   it into your hand, then shuffle. (Then exile this card. You may
//	   cast the artifact later from exile.)"
//
// The anthem is Glorious Anthem's Layer 7c static, so it counts creatures
// that arrive later and stops with the artifact. The adventure lifecycle
// (exile the card, cast the artifact later) is the engine's
// (game/adventure.go); Seek the Heart is The Seriema's tutor.
//
// No simplification. The Hobbit (#4454).
func init() {
	const oracle = "09f09d0a-6054-4545-b81e-0295773cc5f0"
	Register(Spec{
		OracleID:     oracle,
		Name:         "The Arkenstone",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("The Arkenstone — draw a card", Do(DrawCards{N: 1})),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Seek the Heart",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: MatchLegendaryCreature,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Seek the Heart — search for a legendary creature card",
			}.Apply(ctx)
		},
	})
}
