package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cdcfc66d-3d3a-4311-b864-937e7b624d61",
		Name:         "Midnight Angel Armor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3), effects.GrantToAttached("flying", "vigilance")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Midnight Angel Armor — create a 1/1 white Soldier creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenAndAttachSource(g, item, "1/1 white Soldier", true)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
