package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rogue Kavu — Creature — Kavu {1}{R}, 1/1 (Invasion, S82 OLD #5616):
//
//	"Whenever this creature attacks alone, it gets +2/+0 until end of turn."
//
// "Alone" is game.AttackedAlone, read when the trigger is harvested on the
// Kavu's own attack (CR 506.5), as Bob, Reluctant HYDRA Agent reads it. The
// pump is until end of turn, on the Kavu itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17df95b2-49ef-43ca-b3c8-66a270f457a7",
		Name:         "Rogue Kavu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
				return effects.ThisAttacked(ev, source, c, g) && game.AttackedAlone(g)
			}, "Rogue Kavu — it gets +2/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Rogue Kavu — +2/+0 until end of turn")),
		},
	})
}
