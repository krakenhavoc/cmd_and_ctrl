package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Airdrop Aeronauts — Creature — Dwarf Scout {3}{W}{W}, 4/3:
//
//	"Flying
//	 Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, you gain 5 life."
//
// Revolt is an intervening if (revolt.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "009a399b-c78e-475b-8bc3-7db2afc9d676",
		Name:            "Airdrop Aeronauts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersIfRevolt("Airdrop Aeronauts — you gain 5 life", Do(GainLife{Amount: 5})),
		},
	})
}
