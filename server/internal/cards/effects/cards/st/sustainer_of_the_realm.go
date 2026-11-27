package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sustainer of the Realm — Creature — Angel {2}{W}{W}, 2/3 (Urza's Legacy, S82 OLD b715, #5623):
//
//	"Flying
//	Whenever this creature blocks, it gets +0/+2 until end of turn."
//
// Giant Badger's block trigger at +0/+2: it triggers once however many
// attackers it blocks (CR 509.3a, SelfBlocksOnce), and the pump is its own,
// ending in the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c99d52d-f909-4494-aeee-937ae779fd4a",
		Name:            "Sustainer of the Realm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Sustainer of the Realm — gets +0/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(0, 2, "Sustainer of the Realm — +0/+2")),
		},
	})
}
