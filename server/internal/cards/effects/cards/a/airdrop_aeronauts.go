package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "009a399b-c78e-475b-8bc3-7db2afc9d676",
		Name:            "Airdrop Aeronauts",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersIfRevolt("Airdrop Aeronauts — you gain 5 life", effects.Do(effects.GainLife{Amount: 5})),
		},
	})
}
