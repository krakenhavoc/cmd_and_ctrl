package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auspicious Ancestor — Creature — Human Cleric {3}{W}, 2/3 (Mirage, S82 OLD
// batch 809, #5633):
//
//	"When this creature dies, you gain 3 life.
//	 Whenever a player casts a white spell, you may pay {1}. If you do, you
//	 gain 1 life."
//
// The death trigger is a plain life gain (CR 700.4). The cast trigger watches
// every player's white spells and offers its controller the {1} on
// resolution (Jeweled Torque's may-pay shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82fb9834-de65-41a5-b908-f49668dee3ec",
		Name:         "Auspicious Ancestor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Auspicious Ancestor — you gain 3 life", effects.Do(effects.GainLife{Amount: 3})),
			effects.On(game.EventCast, effects.S82OldB809AnyPlayerCastsAWhiteSpell,
				"Auspicious Ancestor — you may pay {1}; if you do, you gain 1 life",
				effects.S82OldB809MayPayOneToGainOneLife("Auspicious Ancestor")),
		},
	})
}
