package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auxiliary Boosters — Artifact — Equipment {4}{W} (Edge of Eternities,
// #4461):
//
//	"When this Equipment enters, create a 2/2 colorless Robot artifact
//	 creature token and attach this Equipment to it.
//	 Equipped creature gets +1/+2 and has flying.
//	 Equip {3} ({3}: Attach to target creature you control. Equip only
//	 as a sorcery.)"
//
// Dwarven Shortsword's entry trigger (createTokenAndAttachSource): the
// Robot is read through the Boosters' text and the attach rides the
// token's creation, so if the Boosters have left in response the Robot
// is made and nothing is attached (CR 701.3b). The bonus is
// Falcon's Wing Harness's pump and flying grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffa3b20b-d77b-43d5-86ac-6a591df1877f",
		Name:         "Auxiliary Boosters",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttached(1, 2),
			GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Auxiliary Boosters — create a 2/2 colorless Robot artifact creature token and attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "2/2 colorless Robot artifact", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
