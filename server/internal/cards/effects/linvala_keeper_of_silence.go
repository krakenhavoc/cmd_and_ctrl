package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Linvala, Keeper of Silence — Legendary Creature — Angel {2}{W}{W},
// 3/4:
//
//	"Flying
//	 Activated abilities of creatures your opponents control can't be
//	 activated."
//
// The asymmetrical Cursed Totem, and the reason the "your opponents"
// half is a constructor of its own rather than a flag: the seat the
// clause measures against is LINVALA'S CONTROLLER, not the player
// trying to activate. A Linvala an opponent steals silences the
// original controller's creatures — the card keeps working, it just
// starts working for them — and a clause written as "not you" on the
// activator's side would have got that backwards.
//
// NO MANA EXEMPTION, like Cursed Totem: the opponents' mana dorks are
// off too, which is most of what the card does at a Commander table.
//
// The "creatures your opponents control" reading is the OBJECT'S
// controller, and the restriction is therefore battlefield-only: a
// card in a hand has no controller (CR 108.4), and a clause about
// creatures somebody controls has nothing to say about a cycling
// ability fired from hand.
//
// Flying is an ordinary printed keyword; the legend rule is the
// engine's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "88dadc31-dfac-41b3-bf2a-65fa89e3c16d",
		Name:            "Linvala, Keeper of Silence",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ActivationRestrictions: []game.ActivationRestriction{
			OpponentsSourcesCantActivate(
				"Linvala, Keeper of Silence — activated abilities of creatures your opponents control can't be activated.",
				Creature(), false),
		},
	})
}
