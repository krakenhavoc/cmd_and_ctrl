package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orgg — Creature — Orgg {3}{R}{R}, 6/6 (Fallen Empires, S82 OLD b854, #5637):
//
//	"Trample
//	 This creature can't attack if defending player controls an untapped
//	 creature with power 3 or greater.
//	 This creature can't block creatures with power 3 or greater."
//
// Mogg Jailer's attack restriction at power 3 or greater, asked of each
// target's defending player at declaration (CR 508.1c), and Zurgo Bellstriker's
// block rule at 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6ea2b1b-1403-4ccd-a23c-48e0e81ceb91",
		Name:            "Orgg",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.CantAttackIfDefendingPlayerControls(game.PermanentQuery{Types: []string{"creature"}, Untapped: true}.WithPowerAtLeast(3)),
		},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(effects.OnSelf(), effects.OnMatching(effects.And(effects.Creature(), effects.PowerGE(3))),
				"it can't block creatures with power 3 or greater"),
		},
	})
}
