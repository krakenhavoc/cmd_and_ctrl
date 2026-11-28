package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orim's Prayer — {1}{W}{W} Enchantment (Tempest, S82 OLD batch 769, #5628):
//
//	"Whenever one or more creatures attack you, you gain 1 life for each
//	 attacking creature."
//
// One trigger per attack declaration (OncePerBatch, CR 603.2c) whose attackers
// include one declared against YOU; an attack at a planeswalker or battle you
// control is not an attack on you (CR 508.3e). The life is counted as the ability
// resolves: one for each creature that is attacking then, as the text says, so a
// creature that has left combat is not counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f387b1c6-25a6-4fff-90d3-fff027b247a5",
		Name:         "Orim's Prayer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller && ev.Target == source.Controller
			}, "Orim's Prayer — you gain 1 life for each attacking creature", effects.S82OldB769GainLifePerAttacker)),
		},
	})
}
