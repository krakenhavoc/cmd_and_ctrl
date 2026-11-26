package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Captain — Creature — Angel Ally {3}{R}{W}, 4/3 (Battle for
// Zendikar):
//
//	"Flying
//	 Whenever this creature attacks, it gets +1/+1 until end of turn for each
//	 other attacking Ally."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ffc56a2e-f3bd-493b-9b7b-5f04fad69382",
		Name:            "Angelic Captain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Angelic Captain — it gets +1/+1 for each other attacking Ally",
				effects.S82OldB312PlusPerOtherAttacking("Angelic Captain", "Ally")),
		},
	})
}
