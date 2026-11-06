package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mai and Zuko — Legendary Creature — Human Noble Ally {1}{U}{B}{R}, 3/5
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Firebending 3
//	 You may cast Ally spells and artifact spells as though they had
//	 flash."
//
// Firebending 3 is the set's attack trigger (firebending(3)): it uses
// the stack and adds {R}{R}{R} that lasts until end of combat. The
// flash permission is two timing statements, one per noun, so "Ally
// spells and artifact spells" is the union it reads as: an Ally of any
// card type, or any artifact. Neither opens anything else.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b86ba0aa-89f3-4924-b3c4-f7ca44468e9b",
		Name:         "Mai and Zuko",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Mai and Zuko — firebending 3", Do(firebending(3))),
		},
		CastTimings: []game.CastTimingRule{
			CastKindAsThoughFlash(game.PermissionFilter{CreatureType: "Ally"},
				"You may cast Ally spells as though they had flash."),
			CastKindAsThoughFlash(game.PermissionFilter{CardType: "artifact"},
				"You may cast artifact spells as though they had flash."),
		},
	})
}
