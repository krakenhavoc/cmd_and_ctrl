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
				// The question waits for a prompt the life gain asks (a CR 616 order; CR 608.2c, #4155).
				Do(
					GainLife{Amount: 2},
					MayChoice{
						Question: "Spiritual Focus — draw a card?",
						Draws:    1,
						Then:     mayDrawCards,
						Word:     "1",
					},
				)),
		},
	})
}
