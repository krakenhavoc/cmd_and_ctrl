package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ffa3b20b-d77b-43d5-86ac-6a591df1877f",
		Name:         "Auxiliary Boosters",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 2),
			effects.GrantToAttached("flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Auxiliary Boosters — create a 2/2 colorless Robot artifact creature token and attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "2/2 colorless Robot artifact", true)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
