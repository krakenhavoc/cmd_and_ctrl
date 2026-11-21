package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Mist — Enchantment {3}{R} (Innistrad Remastered):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gains double strike until end of turn."
func init() {
	t := effects.AtBeginningOfYourCombat("Blood Mist — target creature you control gains double strike until end of turn",
		effects.EbTargetCreatureGainsUntilEOT("double strike", "Blood Mist — double strike until end of turn"))
	t.Targets = effects.TargetCreature("target creature you control", effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     "22188333-69b6-4685-b96b-b8a102ad5531",
		Name:         "Blood Mist",
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
