package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ebon Drake — Creature — Drake {2}{B}, 3/3 (Fifth Dawn, #5604):
//
//	"Flying
//	 Whenever a player casts a spell, you lose 1 life."
//
// Managorger Hydra's "a player casts a spell" condition, any player's cast,
// the Drake's controller's own included. The life is lost, not dealt as damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db4242ab-fdf3-487a-bc77-42bd2ef973c3",
		Name:            "Ebon Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.CardID != uuid.Nil
			}, "Ebon Drake — you lose 1 life", effects.Do(effects.GainLife{Amount: -1})),
		},
	})
}
