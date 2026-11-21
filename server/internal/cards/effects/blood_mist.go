package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Mist — Enchantment {3}{R} (Innistrad Remastered):
//
//	"At the beginning of combat on your turn, target creature you control
//	 gains double strike until end of turn."
func init() {
	t := AtBeginningOfYourCombat("Blood Mist — target creature you control gains double strike until end of turn",
		ebTargetCreatureGainsUntilEOT("double strike", "Blood Mist — double strike until end of turn"))
	t.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "22188333-69b6-4685-b96b-b8a102ad5531",
		Name:         "Blood Mist",
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
