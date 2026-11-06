package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Belligerent of the Ball — Creature — Ogre Warrior {2}{R}, 3/3 (Wilds
// of Eldraine, #5164):
//
//	"Celebration — At the beginning of combat on your turn, if two or
//	 more nonland permanents entered the battlefield under your control
//	 this turn, target creature you control gets +1/+0 and gains menace
//	 until end of turn."
//
// Celebration is an intervening if (celebration.go, CR 603.4): checked
// as combat begins and again as the ability resolves. The target is
// any creature you control, the Belligerent included; with none, the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Belligerent of the Ball — target creature you control gets +1/+0 and gains menace until end of turn"
	Register(Spec{
		OracleID:     "2be2f6c8-d8df-4e5c-a0c1-7566183bbe81",
		Name:         "Belligerent of the Ball",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombatIfCelebration(label, firstTargetGainsUntilEndOfTurn(1, label, "menace")),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
