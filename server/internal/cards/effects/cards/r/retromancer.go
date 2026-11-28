package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Retromancer — Creature — Lizard Shaman 3/3 {2}{R}{R} (Urza's Saga, S82 OLD b723):
//
//	Whenever this creature becomes the target of a spell or ability, this creature deals 3 damage to that spell or ability's controller.
//
// One trigger per becoming a target (each spell or ability that targets it, once each); the damage goes to the controller of that spell or ability, who is the event's actor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c34949d-0765-455d-9ef4-81bc162b48d9",
		Name:         "Retromancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Retromancer — 3 damage to that spell or ability's controller", effects.S82OldB723DamageThatPlayer(3)),
		},
	})
}
