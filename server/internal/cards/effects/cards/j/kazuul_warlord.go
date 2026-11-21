package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kazuul Warlord — Creature — Minotaur Warrior Ally {4}{R}, 3/3 (Zendikar, #5583):
//
//	"Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on each Ally creature you control."
//
// The Zendikar Ally trigger (zen_b2_helpers.go) with Earth King's
// Lieutenant's counters: one on each Ally creature its controller
// controls as the ability resolves, the Warlord included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78504edb-6536-4a11-ac1c-7d53684f2226",
		Name:         "Kazuul Warlord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ZenB2AllyTrigger("Kazuul Warlord — you may put a +1/+1 counter on each Ally creature you control",
				"Put a +1/+1 counter on each Ally creature you control?",
				func(g *game.Game, item *game.StackItem) error {
					ally := effects.NewContext(g, item).TextSubtype("Ally")
					return effects.PutAPlusOneCounterOnEachYouControl(effects.And(effects.Creature(), effects.Subtype(ally)))(g, item)
				}),
		},
	})
}
