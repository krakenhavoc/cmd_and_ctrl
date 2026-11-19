package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Headsplitter — Artifact — Equipment {1}{R} (Assassin's Creed, #5948):
//
//	"When this Equipment enters, create a 1/1 black Assassin creature token
//	 with menace, then attach this Equipment to it.
//	 Equipped creature gets +1/+0.
//	 Equip {2}"
//
// Job select's shape over an Assassin token with menace: the token is made
// first and the Equipment attached to it once it has entered
// (createTokenAndAttachSource), so a Headsplitter that left or came back in
// response attaches nothing and the Assassin stays unequipped (CR 701.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8333c82f-aae6-470f-a9db-63ab24f9bb46",
		Name:         "Headsplitter",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Static:       []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Headsplitter — create a 1/1 black Assassin creature token with menace, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return createTokenAndAttachSource(g, item, "1/1 black Assassin with menace", false)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
