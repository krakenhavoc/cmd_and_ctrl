package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Godhunter Octopus — Creature — Octopus {5}{U}, 5/5:
//
//	"This creature can't attack unless defending player controls an
//	 enchantment or an enchanted permanent."
//
// ADR 0107 §2's restriction (#1879, CR 508.1c) with two queries, either of
// which lets it attack: an enchantment, or a permanent with an Aura
// attached. An Aura one player controls on another player's creature makes
// that creature's controller a legal defender, because they control an
// enchanted permanent. An Equipment does not enchant. The defending player
// is worked out per target (CR 508.5, 508.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "930b48b8-dbd1-4109-9eaa-7d7e9a04fa5b",
		Name:         "Godhunter Octopus",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(
				game.PermanentQuery{Types: []string{"enchantment"}},
				game.PermanentQuery{Enchanted: true},
			),
		},
	})
}
