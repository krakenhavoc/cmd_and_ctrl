package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blind Creeper — Creature — Zombie Beast {1}{B}, 3/3 (Fifth Dawn, #5604):
//
//	"Whenever a player casts a spell, this creature gets -1/-1 until
//	 end of turn."
//
// Managorger Hydra's "a player casts a spell" condition, with the shared
// thisGetsUntilEndOfTurn body for the -1/-1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02d4c3a8-c733-45c1-bca5-51eed47c9230",
		Name:         "Blind Creeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.CardID != uuid.Nil
			}, "Blind Creeper — -1/-1 until end of turn", effects.ThisGetsUntilEndOfTurn(-1, -1, "Blind Creeper — -1/-1 until end of turn")),
		},
	})
}
