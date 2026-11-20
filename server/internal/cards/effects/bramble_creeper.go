package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bramble Creeper — Creature — Elemental {4}{G}, 0/3 (Magic 2010, #5585):
//
//	"Whenever this creature attacks, it gets +5/+0 until end of turn."
//
// Charging Bandits' attack trigger with +5/+0: it uses the stack, and the
// pump ends at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33416979-2a3b-4184-818c-6a4e9ccb4580",
		Name:         "Bramble Creeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Bramble Creeper — it gets +5/+0 until end of turn",
				thisGetsUntilEndOfTurn(5, 0, "Bramble Creeper — +5/+0")),
		},
	})
}
