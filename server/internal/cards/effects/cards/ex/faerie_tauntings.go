package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Tauntings — Kindred Enchantment — Faerie {2}{B} (Lorwyn):
//
//	"Whenever you cast a spell during an opponent's turn, you may have each
//	 opponent lose 1 life."
//
// Voracious Tome-Skimmer's cast condition: the caster is the Tauntings'
// controller and is not the active player. The loss is a "you may"
// (CR 603.5) and hits every opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c5d1a62-3b83-4812-82bb-d95b39b990b4",
		Name:         "Faerie Tauntings",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && !effects.IsActivePlayer(g, source.Controller)
			}, "Faerie Tauntings — you may have each opponent lose 1 life",
				effects.Do(effects.EachOpponentLosesLife{N: 1})),
				"Faerie Tauntings — have each opponent lose 1 life?"),
		},
	})
}
