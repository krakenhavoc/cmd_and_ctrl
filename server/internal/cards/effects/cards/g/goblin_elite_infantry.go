package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Elite Infantry — Creature — Goblin Warrior {1}{R}, 2/2 (Mirage, S82
// OLD b797, #5633):
//
//	"Whenever this creature blocks or becomes blocked, it gets -1/-1 until
//	 end of turn."
//
// One printed trigger, watching two events (the block declaration and the
// attacker becoming blocked), each for the source itself. It becomes
// blocked once however many creatures block it (CR 509.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6a8861c-b6fe-45a6-86e4-9160c36362ff",
		Name:         "Goblin Elite Infantry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				"Goblin Elite Infantry — it gets -1/-1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(-1, -1, "Goblin Elite Infantry — -1/-1 until end of turn")),
		},
	})
}
