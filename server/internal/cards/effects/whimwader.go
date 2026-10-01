package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whimwader — Creature — Elemental {4}{U}, 6/4:
//
//	"This creature can't attack unless defending player controls a blue
//	 permanent."
//
// ADR 0107 §2's restriction (#1879, CR 508.1c), over the defending
// player's permanents' current colours. The defending player is worked out
// per target (CR 508.5, 508.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21d9ce2c-eb6a-4f43-a79b-0b99b3dc4a00",
		Name:         "Whimwader",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(game.PermanentQuery{Colors: []string{"U"}}),
		},
	})
}
