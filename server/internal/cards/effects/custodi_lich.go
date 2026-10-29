package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Custodi Lich — Creature — Zombie Cleric {3}{B}{B}, 4/2:
//
//	"When this creature enters, you become the monarch.
//	 Whenever you become the monarch, target player sacrifices a
//	 creature of their choice."
//
// The first catalog card to watch EventMonarchChanged (#1722). Its own
// ETB crowns you, which triggers the edict — the card's printed
// synergy — and so does every later becoming: taking the crown back in
// combat, a Court entering while an opponent held it, a CR 725.4
// hand-on. It does NOT trigger when you are already the monarch and an
// effect tells you to become it: that is no change of holder (CR
// 725.3), and the engine emits nothing.
//
// The target is a PLAYER; the creature is that player's choice, so
// hexproof on their creatures is irrelevant and a player with no
// creature is still a legal target (the edict does nothing).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d95ee14-9ff0-4d7f-be56-361a500ce36f",
		Name:         "Custodi Lich",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Custodi Lich"),
			Targeting(WheneverYouBecomeTheMonarch("Custodi Lich — target player sacrifices a creature",
				targetPlayerSacrificesACreature("Custodi Lich — sacrifice a creature")),
				TargetPlayer("target player")),
		},
	})
}
