package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glint-Horn Buccaneer — 2/4 Creature — Minotaur Pirate for
// {1}{R}{R}:
//
//	"Haste
//	 Whenever you discard a card, this creature deals 1 damage to
//	 each opponent.
//	 {1}{R}, Discard a card: Draw a card. Activate only if this
//	 creature is attacking."
//
// The discard trigger is the half that matters here and the half the
// engine can express. The activated ability is deferred twice over:
// AbilityCost has no discard component, and there's no "only if this
// creature is attacking" timing predicate. Both are noted in the
// decklist triage rather than half-implemented.
func init() {
	Register(Spec{
		OracleID:        "64ad5657-78e9-4f34-8877-18c4f51fff9a",
		Name:            "Glint-Horn Buccaneer",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The \"{1}{R}, discard a card: draw a card\" ability while attacking cannot be activated; only the discard damage trigger works."},
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Glint-Horn Buccaneer — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
	})
}
