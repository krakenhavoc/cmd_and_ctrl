package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Matsu-Tribe Birdstalker — Creature — Snake Warrior Archer {2}{G}{G}, 2/2
// (Saviors of Kamigawa, S82 OLD b583):
//
//	"Whenever this creature deals combat damage to a creature, tap that
//	 creature and it doesn't untap during its controller's next untap step.
//	 {G}: This creature gains reach until end of turn."
//
// The trigger is Voracious Cobra's (ThisDealtCombatDamageToACreature, naming
// the damaged creature as the object it was when the damage was dealt, CR
// 510.2), with White Dragon's tap-and-freeze for its effect
// (OldB166TapAndFreezeDamagedCreature). The reach is Moopsy's activation.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4c90ca5c-551c-4da5-9482-62521b1ae17c",
		Name:         "Matsu-Tribe Birdstalker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.B583WhenThisDealsCombatDamageToACreatureTapFreeze("Matsu-Tribe Birdstalker — tap the creature it dealt combat damage to; it doesn't untap during its controller's next untap step"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: This creature gains reach until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{G}"),
			Effect:  effects.ThisGainsKeywordUntilEndOfTurn("reach", "Matsu-Tribe Birdstalker — reach until end of turn"),
		}},
	})
}
