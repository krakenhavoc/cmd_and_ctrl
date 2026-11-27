package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bramble Elemental — Creature — Elemental {3}{G}{G}, 4/4 (Planechase 2012, S82
// OLD, #5569):
//
//	"Whenever an Aura becomes attached to this creature, create two 1/1 green
//	 Saproling creature tokens."
//
// Brood Keeper's attach watch (CR 303.4f): an Aura that becomes attached to this
// creature, whoever controls the Aura, once per attachment; re-attaching to the
// host it is already on does nothing (CR 701.3b). The tokens are made by the
// Elemental's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac5c0078-e2f1-4ac1-8748-575d69e09cf8",
		Name:         "Bramble Elemental",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttach, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventAttach && ev.Label == "Aura" && !ev.AlreadyAttached &&
					source != nil && ev.Target != uuid.Nil && ev.Target == source.InstanceID
			}, "Bramble Elemental — create two 1/1 green Saproling creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 2})),
		},
	})
}
