package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empyreal Voyager — Creature — Vedalken Scout {1}{G}{U}, 2/3:
//
//	"Flying, trample
//	 Whenever this creature deals combat damage to a player, you get
//	 that many {E} (energy counters)."
//
// ADR 0129 PR 1. "That many" is the damage dealt, read off the event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7112d60-1154-4770-9ec4-c7038157be8a",
		Name:            "Empyreal Voyager",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Empyreal Voyager — you get that many {E}",
				func(g *game.Game, item *game.StackItem) error {
					return GetEnergy{N: item.Trigger.Event.Amount}.Apply(NewContext(g, item))
				}),
		},
	})
}
