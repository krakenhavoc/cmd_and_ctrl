package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Indebted Samurai — Creature — Human Samurai {3}{W}, 2/3 (Betrayers of
// Kamigawa, #5602):
//
//	"Bushido 1 (Whenever this creature blocks or becomes blocked, it gets
//	 +1/+1 until end of turn.)
//	 Whenever a Samurai you control dies, you may put a +1/+1 counter on
//	 this creature."
//
// Bushido is the shared trigger (CR 702.45a). "A Samurai you control
// dies" reads the creature's last-known subtypes and controller (CR
// 603.10a), the Elemental dies trigger's shape, so a changeling counts
// and "Samurai" reads through a text change. When Indebted Samurai
// itself dies it triggers too, and the counter has nowhere to go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fee392d9-a961-4f41-81ba-ab1b87c931fa",
		Name:         "Indebted Samurai",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Bushido(1, "Indebted Samurai"),
			effects.Optional(effects.On(game.EventLTB, effects.BokB1SamuraiYouControlDied,
				"Indebted Samurai — put a +1/+1 counter on this creature", effects.PutACounterOnThis("+1/+1")),
				"Indebted Samurai — put a +1/+1 counter on it?"),
		},
	})
}
