package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiritual Focus — Enchantment {1}{W}:
//
//	"Whenever a spell or ability an opponent controls causes you to
//	 discard a card, you gain 2 life and you may draw a card."
//
// The battlefield reader of ADR 0153 §4's predicate, DiscardedByAnOpponent
// without "this card": any card its controller discards because a
// resolving spell or ability an opponent controls told them to (the stack
// item's controller, CR 113.8). Your own Careful Study, a discard paid as
// a cost and the cleanup step's discard do not trigger it. One trigger
// per card, as printed: a two-card Mind Rot gains 4 life and offers two
// draws.
//
// The life is gained at once and only the draw is a "may".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da350f46-6d6e-4257-8f4f-77769c442d0c",
		Name:         "Spiritual Focus",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, DiscardedByAnOpponent(false),
				"Spiritual Focus — gain 2 life and you may draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GainLife{Amount: 2}).Apply(ctx); err != nil {
						return err
					}
					return MayChoice{
						Question: "Spiritual Focus — draw a card?",
						OnYes:    func(ctx *Context) error { return DrawCards{N: 1}.Apply(ctx) },
					}.Apply(ctx)
				}),
		},
	})
}
