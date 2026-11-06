package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cautious Survivor — Creature — Elf Survivor {3}{G}, 4/4 (Duskmourn:
// House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, you gain 2 life."
//
// Defiant Survivor's Survival trigger (survival.go) with a life gain:
// the "if" is an intervening if, asked as the phase begins and again
// as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28eb696f-a42c-4444-8139-dfd62029e020",
		Name:         "Cautious Survivor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Survival("Cautious Survivor — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
