package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scorpion, Seething Striker — Legendary Creature — Scorpion Human Villain
// {3}{B}, 3/3 (Marvel's Spider-Man):
//
//	"Deathtouch
//	 At the beginning of your end step, if a creature died this turn,
//	 target creature you control connives."
//
// The "if" is an intervening if (CR 603.4) on the table-wide tally of
// creatures that died this turn, any controller, tokens included (Zombie
// Ogre's condition): checked as the end step begins and again as the
// ability resolves. The target is chosen as the trigger goes on the stack
// and may be Scorpion himself. ConniveEach draws, discards, and puts the
// counter on the target for a nonland discard (CR 701.50a).
//
// No simplification.
func init() {
	const label = "Scorpion, Seething Striker — target creature you control connives"
	Register(Spec{
		OracleID:        "c357230a-afde-4c2a-b390-9f6928599236",
		Name:            "Scorpion, Seething Striker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condACreatureDiedThisTurn,
				Targeting(
					AtYourEndStep(label, conniveFirstLegalTarget),
					TargetCreature("target creature you control", YouControl()))),
		},
	})
}
