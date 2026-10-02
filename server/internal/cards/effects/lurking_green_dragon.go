package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lurking Green Dragon — Creature — Dragon {3}{G}, 4/4:
//
//	"Flying
//	 This creature can't attack unless defending player controls a
//	 creature with flying."
//
// ADR 0107 §2's restriction (#1879, CR 508.1c). "A creature with flying"
// is read off the defending player's creatures as they are now, so a
// granted flying counts and a lost one doesn't. The defending player is
// worked out per target (CR 508.5, 508.5a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "262e1cf6-61c8-444d-a71b-050eebcaf932",
		Name:            "Lurking Green Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(game.PermanentQuery{Types: []string{"creature"}, Keyword: "flying"}),
		},
	})
}
