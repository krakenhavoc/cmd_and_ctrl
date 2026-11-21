package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shrewd Storyteller — Creature — Human Survivor {1}{G}{W}, 3/3
// (Duskmourn: House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, put a +1/+1 counter on target creature."
//
// Defiant Survivor's Survival trigger (survival.go), its "if" asked as
// the phase begins and again as it resolves (CR 603.4), with a target
// creature chosen as it goes on the stack: any creature, the Storyteller
// itself or an opponent's included. Amok's counter body puts the counter
// on the target if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec238819-81f6-4904-922d-298a17cbac7e",
		Name:         "Shrewd Storyteller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Survival("Shrewd Storyteller — put a +1/+1 counter on target creature",
				effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
