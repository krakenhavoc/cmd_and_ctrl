package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "64ad5657-78e9-4f34-8877-18c4f51fff9a",
		Name:            "Glint-Horn Buccaneer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Glint-Horn Buccaneer — 1 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 1})), game.Purpose{DiscardPayoff: &game.DiscardPayoff{Any: true, DamageEachOpponent: 1}}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{R}, Discard a card: Draw a card. Activate only if this creature is attacking.",
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:      effects.Plus(effects.ManaCost("{1}{R}"), effects.DiscardACard()),
			Condition: effects.SourceIsAttacking(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
