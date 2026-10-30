package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Angel Armor — {3}{W}{W} Artifact — Equipment (Marvel Super Heroes
// Commander, #4455):
//
//	"When this Equipment enters, create a 1/1 white Soldier creature token,
//	 then attach this Equipment to it.
//	 Equipped creature gets +3/+3 and has flying and vigilance.
//	 Equip {3}"
//
// Dwarven Shortsword's enter trigger over this card's own Soldier: the
// attach rides the token's creation, and is skipped if the Equipment has
// left in response (CR 701.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cdcfc66d-3d3a-4311-b864-937e7b624d61",
		Name:         "Midnight Angel Armor",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 3), GrantToAttached("flying", "vigilance")},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Midnight Angel Armor — create a 1/1 white Soldier creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "1/1 white Soldier", true)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}
