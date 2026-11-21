package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lyla, Holographic Assistant — Legendary Artifact Creature — Illusion
// Advisor {3}{U}, 2/2 (Marvel's Spider-Man, #4460):
//
//	"Whenever you draw a card, put a +1/+1 counter on target creature."
//
// One trigger per card drawn, each choosing its own target (any creature,
// not only yours) as it goes on the stack and re-checking it on
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b7fb0141-4fbd-4269-a2e9-92f75ec95271",
		Name:         "Lyla, Holographic Assistant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventDrawCard, effects.ByYou,
				"Lyla, Holographic Assistant — put a +1/+1 counter on target creature",
				effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
