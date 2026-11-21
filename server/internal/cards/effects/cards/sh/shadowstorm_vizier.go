package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadowstorm Vizier — Creature — Human Cleric {U}{B}, 1/3
// (Amonkhet, #5535):
//
//	"Flying
//	 Whenever you cycle or discard a card, this creature gets +1/+1 until end of turn."
//
// "Cycle or discard" is one event, the discard: cycling's cost discards
// the card (CR 702.29d), so watching the cycle as well would count one
// cycle twice (see Curator of Mysteries).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3e5b559-660c-4e14-9209-29efc07d3563",
		Name:            "Shadowstorm Vizier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.DiscardedByYou(ev, source)
			}, "Shadowstorm Vizier — it gets +1/+1 until end of turn", effects.ThisGetsUntilEndOfTurn(1, 1, "Shadowstorm Vizier — +1/+1")),
		},
	})
}
