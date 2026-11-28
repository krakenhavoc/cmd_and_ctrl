package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inner Sanctum — {1}{W}{W} Enchantment (Weatherlight, S82 OLD batch 772, #5629):
//
//	"Cumulative upkeep—Pay 2 life.
//	 Prevent all damage that would be dealt to creatures you control."
//
// The cumulative upkeep is Morinfen's life payment, two life for each age counter
// (CR 702.24a, 119.4), paid all at once or the enchantment is sacrificed. The
// prevention is Blessed Sanctuary's, for every kind of damage and every source,
// to the creatures its controller controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2ae36d31-08b7-481a-bfa1-a5c65921dee0",
		Name:         "Inner Sanctum",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToCreaturesYouControl,
				Label: "Inner Sanctum — prevent all damage that would be dealt to creatures you control",
			}),
		},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Inner Sanctum — cumulative upkeep: pay 2 life", effects.LifePayment(2), "", ""),
		},
	})
}
