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
// Both halves of the activated ability's gap have since closed: #660
// gave AbilityCost a discard component (DiscardCardsMatching) and
// #743 gave activated abilities a Condition slot, so the only piece
// missing was the "only if this creature is attacking" predicate
// itself — SourceIsAttacking, added alongside this card. Discarding
// to pay the ability's cost is an ordinary discard, so it also fires
// the "whenever you discard a card" trigger above, same as paper.
func init() {
	Register(Spec{
		OracleID:        "64ad5657-78e9-4f34-8877-18c4f51fff9a",
		Name:            "Glint-Horn Buccaneer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Glint-Horn Buccaneer — 1 damage to each opponent", func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:     "{1}{R}, Discard a card: Draw a card. Activate only if this creature is attacking.",
			Cost:      Plus(ManaCost("{1}{R}"), DiscardACard()),
			Condition: SourceIsAttacking(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
