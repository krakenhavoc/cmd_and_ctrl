package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turntimber Ranger — Creature — Elf Scout Ranger Ally {3}{G}{G}, 2/2 (Zendikar,
// S82 OLD b496):
//
//	"Whenever this creature or another Ally you control enters, you may create a
//	 2/2 green Wolf creature token. If you do, put a +1/+1 counter on this
//	 creature."
//
// A rally trigger (one per entering Ally), optional; the counter goes on the
// Ranger only as the Wolf is made, and only while it is still the same permanent
// (CounterOnThis).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "391fdf1e-e489-410e-9f33-b674fd98d0fc",
		Name:         "Turntimber Ranger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.S82OldB313Rally("Turntimber Ranger — create a 2/2 green Wolf; if you do, put a +1/+1 counter on this creature",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 green Wolf"), N: 1}, effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				"Turntimber Ranger — create a 2/2 green Wolf creature token?"),
		},
	})
}
