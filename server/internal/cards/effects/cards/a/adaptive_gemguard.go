package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adaptive Gemguard — Artifact Creature — Gnome {3}{W}, 3/3 (The Lost
// Caverns of Ixalan, #4470):
//
//	"Tap two untapped artifacts and/or creatures you control: Put a +1/+1
//	 counter on this creature. Activate only as a sorcery."
//
// Sunshot Militia's cost (TapOthersCost, ADR 0071 §2): any two untapped
// artifacts or creatures you control, the Gemguard itself included since
// the cost does not say "other", and summoning sickness does not stop a
// creature from being tapped for it because the cost has no {T}
// (CR 302.6). The counter goes on the Gemguard only while it is still
// this object on the battlefield.
//
// No simplification.
func init() {
	const tapTwo = "two untapped artifacts and/or creatures you control"
	effects.Register(effects.Spec{
		OracleID:     "2fdc0d15-ecb4-4a6d-be77-7a3fca188c9c",
		Name:         "Adaptive Gemguard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Tap two untapped artifacts and/or creatures you control: Put a +1/+1 counter on this creature. Activate only as a sorcery.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: effects.TargetPermanent(tapTwo, effects.Or(effects.Artifact(), effects.Creature())),
				Label:  tapTwo,
			}},
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}
