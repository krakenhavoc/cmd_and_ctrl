package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Third Path Iconoclast — Creature — Human Monk {U}{R}, 2/1 (EDHREC
// rank 1360):
//
//	"Whenever you cast a noncreature spell, create a 1/1 colorless
//	 Soldier artifact creature token."
//
// Young Pyromancer with artifact tokens — the Izzet spells deck's
// two-drop. The condition is Firebrand Archer's (the spell is read
// off the stack, where its type line is intact), and the token is
// an ARTIFACT creature, so it feeds every artifact count, every
// Krark-Clan Ironworks and every "whenever an artifact enters"
// trigger, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f7156897-2b02-4ecd-868d-d4d59244e9ed",
		Name:         "Third Path Iconoclast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(), "Third Path Iconoclast — a 1/1 Soldier artifact creature", Do(CreateToken{Template: b12ColorlessSoldierArtifactToken(), N: 1})),
		},
	})
}
