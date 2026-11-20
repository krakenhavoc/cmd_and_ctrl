package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wasitora, Nekoru Queen — Legendary Creature — Cat Dragon {2}{B}{R}{G},
// 5/4 (Commander 2017, #5533):
//
//	"Flying, trample
//	 Whenever Wasitora deals combat damage to a player, that player
//	 sacrifices a creature of their choice. If the player can't, you create
//	 a 3/3 black, red, and green Cat Dragon creature token with flying."
//
// The damaged player chooses the sacrifice (CR 701.21a) through Invoke
// Despair's keyed sacrifice run, whose continuation reads what they really
// sacrificed: nothing means they couldn't, and then the Queen's controller
// creates the Cat Dragon (c17WasitoraSacrificed). The run finishes even when
// the player controls no creature to be asked about.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13fc4168-fcc2-4011-804f-211d5d86b7dd",
		Name:            "Wasitora, Nekoru Queen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Wasitora, Nekoru Queen — that player sacrifices a creature; if they can't, you create a 3/3 Cat Dragon",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					victim := item.Trigger.Event.Target
					if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
						return nil
					}
					return g.PlayerSacrificesThenKeyedForEffect(item.SourceCardID, victim,
						sacrificeSpec("a creature", Creature()),
						"Wasitora, Nekoru Queen — sacrifice a creature", 1,
						c17WasitoraSacrificed, []uuid.UUID{item.Controller, victim})
				}),
		},
	})
}
