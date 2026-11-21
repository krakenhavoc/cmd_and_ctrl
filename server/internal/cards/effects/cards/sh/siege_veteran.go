package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "4d4fc571-a859-4460-a3c7-3f76a671ab1b",
		Name:         "Siege Veteran",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtBeginningOfYourCombat("Siege Veteran — a +1/+1 counter on target creature you control", effects.PlusOneCounterOnTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
			effects.On(game.EventLTB, effects.BroB05NontokenSoldierYouControlDied,
				"Siege Veteran — a 1/1 Soldier artifact creature",
				effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.BroB05SoldierToken), N: 1})),
		},
	})
}
