package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider-Man, New Champion — Legendary Creature — Spider Human Hero
// {2}{R}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever you discard one or more cards, Spider-Man deals that much
//	 damage to any target."
//
// "One or more ... that much" is the batch total (Captain Howler's shape,
// game/batch_total.go): the engine emits one EventDiscardCard per card and
// AtBatchEnd: ThatManyEvents() holds the trigger until the discard has
// finished, so it goes on the stack ONCE (one target pick) carrying the
// number of cards. The damage is dealt by Spider-Man himself; should he
// have left, the damage still comes from his last-known information
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4604c8fb-d38f-42aa-935a-eada20f6aafa",
		Name:            "Spider-Man, New Champion",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WithAtBatchEnd(effects.Targeting(effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Spider-Man, New Champion — deals damage equal to the cards discarded to any target",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SourceDealsComputedDamageToEachLegalTarget(func(*effects.Context) int { return effects.SettledEventAmount(item) })(g, item)
				}), effects.TargetAny()), effects.ThatManyEvents()),
		},
	})
}
