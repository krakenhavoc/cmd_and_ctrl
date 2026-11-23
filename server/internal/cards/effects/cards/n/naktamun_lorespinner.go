package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Naktamun Lorespinner // Wheel of Fortune — Creature — Jackal Wizard
// {2}{R}, 3/3 // Sorcery {2}{R} (preparation card, CR 722; Secrets of
// Strixhaven Commander, #4834):
//
//	"At the beginning of your upkeep, if a player has one or fewer cards
//	 in hand, this creature becomes prepared. (While it's prepared, you
//	 may cast a copy of its spell. Doing so unprepares it.)"
//
//	Wheel of Fortune — "Each player discards their hand, then draws
//	 seven cards."
//
// The intervening if (CR 603.4) is any player at the table, its
// controller included, and is checked as the upkeep begins and again as
// the trigger resolves (the ruling). The spell is Wheel of Fortune's own
// body (wheelOfFortuneResolve).
//
// No simplification.
func init() {
	const id = "c78783e5-868d-4a8b-a4f8-95a92853cf0a"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Naktamun Lorespinner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
				return effects.ByYou(ev, source, c, g) && effects.SocB9APlayerHasAtMostOneCardInHand(g)
			}, "Naktamun Lorespinner — becomes prepared",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SocB9APlayerHasAtMostOneCardInHand(g) {
						return nil
					}
					return effects.SocB4ThisBecomesPrepared(g, item)
				}),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Wheel of Fortune",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.WheelOfFortuneResolve,
	})
}
