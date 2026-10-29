package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stone-Giant of High Pass — Creature — Giant {5}{R}{R}, 7/7 (The Hobbit):
//
//	"Whenever this creature enters or attacks, create a 3/1 colorless Wall
//	 artifact creature token with defender named Stone Boulder.
//	 {2}{R}, Sacrifice an artifact: This creature deals 4 damage to any
//	 target."
//
// One ability for both events (WhenThisEntersOrAttacks), creating the named
// token from its own table row. The damage ability's cost is paid at announce
// with the sacrifice as a cost (Ravenous Squirrel's sacrificeSpec over
// artifacts), so a Stone Boulder, a Treasure or any other artifact pays for
// it, and the damage comes from the Giant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfaa8b7b-7bfc-4660-bbc7-a717e05df6ef",
		Name:         "Stone-Giant of High Pass",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Stone-Giant of High Pass — create a 3/1 Wall artifact creature token named Stone Boulder",
				Do(CreateToken{Template: TokenCard("3/1 colorless Wall artifact with defender named Stone Boulder"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{R}, Sacrifice an artifact: This creature deals 4 damage to any target.",
			Cost: Plus(ManaCost("{2}{R}"), game.AbilityCost{
				SacrificeOther: sacrificeSpec("an artifact", Artifact()),
			}),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 4)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageToFirstTarget(4)(item, NewContext(g, item))
			},
		}},
	})
}
