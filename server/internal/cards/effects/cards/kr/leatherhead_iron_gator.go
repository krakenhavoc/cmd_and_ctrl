package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leatherhead, Iron Gator — {5}{G}{G} Legendary Creature — Crocodile Mutant
// Rogue 5/5 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Trample, haste
//	 Whenever Leatherhead attacks, put two +1/+1 counters on each
//	 creature you control."
//
// The attack trigger is Avabruck Caretaker's body
// (b11PutCountersOnEachCreatureYouControl): each creature its controller
// controls as it resolves, Leatherhead included while he is still there,
// gets two counters as ONE placement, so a Hardened Scales adds one to each
// creature once, not twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e9679e8-bef1-495c-8251-0d87d971e32a",
		Name:            "Leatherhead, Iron Gator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Leatherhead — put two +1/+1 counters on each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B11PutCountersOnEachCreatureYouControl(g, item, 2)
				}),
		},
	})
}
