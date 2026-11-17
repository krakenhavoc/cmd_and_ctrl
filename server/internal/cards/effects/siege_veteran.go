package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Siege Veteran — Creature — Human Soldier {2}{W}, 2/2 (The Brothers' War,
// #5482):
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on target
//	 creature you control.
//	 Whenever another nontoken Soldier you control dies, create a 1/1
//	 colorless Soldier artifact creature token."
//
// Two triggers. The combat trigger's target is chosen as it goes on the
// stack. The dies trigger reads the dead Soldier as it last existed
// (CR 603.10a): a token, a non-Soldier, or one an opponent controlled does
// not trigger it, and neither does the Veteran's own death.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4d4fc571-a859-4460-a3c7-3f76a671ab1b",
		Name:         "Siege Veteran",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Siege Veteran — a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
			On(game.EventLTB, broB05NontokenSoldierYouControlDied,
				"Siege Veteran — a 1/1 Soldier artifact creature",
				Do(CreateToken{Template: TokenCard(broB05SoldierToken), N: 1})),
		},
	})
}
