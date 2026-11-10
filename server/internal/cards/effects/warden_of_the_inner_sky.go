package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warden of the Inner Sky — Creature — Human Soldier {W}, 1/2 (The Lost
// Caverns of Ixalan, #4470):
//
//	"As long as this creature has three or more counters on it, it has
//	 flying and vigilance.
//	 Tap three untapped artifacts and/or creatures you control: Put a +1/+1
//	 counter on this creature. Scry 1. Activate only as a sorcery."
//
// The static reads every counter on the Warden, of any kind (a stun counter
// counts too), on each recompute (CR 611.3a, 604.2): layer 6 grants flying
// and vigilance. The cost is the "tap N untapped" component (Heritage
// Druid's): any three untapped artifacts and/or creatures the activator
// controls, the Warden itself and a creature that arrived this turn
// included (the ruling: this is not the {T} symbol, CR 302.6). The
// scry follows the counter as one instruction.
//
// No simplification.
func init() {
	const label = "Tap three untapped artifacts and/or creatures you control: Put a +1/+1 counter on this creature. Scry 1. Activate only as a sorcery."
	Register(Spec{
		OracleID:     "5268f772-5b42-4b5b-9e57-c32c662a42f3",
		Name:         "Warden of the Inner Sky",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "As long as this creature has three or more counters on it, it has flying and vigilance.",
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				total := 0
				for _, n := range source.Counters {
					total += n
				}
				if total >= 3 {
					appendKeywordsTo(c, []string{"flying", "vigilance"})
				}
			},
		}},
		Activated: []ActivatedAbility{{
			Label: label,
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: TargetPermanent("untapped artifacts and/or creatures you control", Or(Artifact(), Creature())),
				Label:  "untapped artifacts and/or creatures you control",
			}},
			SorcerySpeed: true,
			Purpose:      game.Purpose{Scry: 1},
			Effect:       Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}, Scry{N: 1}),
		}},
	})
}
