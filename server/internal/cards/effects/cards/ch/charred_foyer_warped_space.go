package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charred Foyer // Warped Space — Enchantment — Room (ADR 0103):
//
//	Charred Foyer {3}{R}: "At the beginning of your upkeep, exile the top
//	 card of your library. You may play it this turn."
//	Warped Space {4}{R}{R}: "Once each turn, you may pay {0} rather than
//	 pay the mana cost for a spell you cast from exile."
//
// Charred Foyer is the impulse-exile primitive on an upkeep trigger.
// Warped Space is a granted alternative cost behind its door (#3174,
// ADR 0118 §3): free for a spell cast from exile, once each turn per
// Warped Space, spent as the spell becomes cast.
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "9f50ec9f-78e6-403f-8f41-84d603fee0eb",
		Name:         "Charred Foyer // Warped Space",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Charred Foyer — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
		}},
		Right: effects.Door{GrantedAlternativeCosts: []game.GrantedAlternativeCost{effects.PayZeroForASpellFromExileOnceEachTurn()}},
	}))
}
