package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b7fb0141-4fbd-4269-a2e9-92f75ec95271",
		Name:         "Lyla, Holographic Assistant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventDrawCard, ByYou,
				"Lyla, Holographic Assistant — put a +1/+1 counter on target creature",
				PlusOneCounterOnTheTarget),
				TargetCreature("target creature")),
		},
	})
}
