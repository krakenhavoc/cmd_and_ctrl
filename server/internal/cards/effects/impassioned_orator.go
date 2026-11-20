package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impassioned Orator — Creature — Human Cleric {1}{W}, 2/2 (Ravnica
// Allegiance, #5521):
//
//	"Whenever another creature you control enters, you gain 1 life."
//
// Hinterland Sanctifier's trigger: another creature entering under your
// control, a token included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7107a2d-2dcf-42e9-9ea9-d0fc0d6d2ec6",
		Name:         "Impassioned Orator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			CreatureEnterPaying(WheneverAnotherCreatureEntersUnderYourControl("Impassioned Orator — you gain 1 life", Do(GainLife{Amount: 1})),
				game.CreatureEnterPayoff{LifeGain: 1}),
		},
	})
}
