package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunshot Militia — Creature — Human Soldier {1}{R}, 1/3 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Tap two untapped artifacts and/or creatures you control: This
//	 creature deals 1 damage to each opponent. Activate only as a sorcery."
//
// Worf, Honor-Bound Warrior's TapOthers cost of two: it is not the {T}
// symbol, so the Militia may tap itself as one of the two and a permanent
// that arrived this turn may pay (CR 302.6). The damage is Merciless
// Enforcers' (damageToEachOpponent), from the Militia. Sorcery timing is
// CR 602.5d.
//
// No simplification.
func init() {
	const tapTwo = "two untapped artifacts and/or creatures you control"
	Register(Spec{
		OracleID:     "19afaf35-b420-49e0-bddf-3180bf1dc30f",
		Name:         "Sunshot Militia",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "Tap two untapped artifacts and/or creatures you control: This creature deals 1 damage to each opponent. Activate only as a sorcery.",
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent(tapTwo, Or(Artifact(), Creature())),
				Label:  tapTwo,
			}},
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToEachOpponent(g, item, 1)
			},
		}},
	})
}
