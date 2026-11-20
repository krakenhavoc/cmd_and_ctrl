package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Healer of the Pride — Creature — Cat Cleric {3}{W}, 2/3 (Magic 2013, #5568):
//
//	"Whenever another creature you control enters, you gain 2 life."
//
// Dazzling Angel's trigger at two life: a creature entering under your
// control other than the Healer itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69b9fa68-b409-4de1-9a40-5262386a0180",
		Name:         "Healer of the Pride",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			CreatureEnterPaying(WheneverAnotherCreatureEntersUnderYourControl("Healer of the Pride — you gain 2 life", Do(GainLife{Amount: 2})),
				game.CreatureEnterPayoff{LifeGain: 2}),
		},
	})
}
