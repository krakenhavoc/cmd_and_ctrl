package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethereal Whiskergill — Creature — Elemental {3}{U}, 4/3:
//
//	"Flying
//	 This creature can't attack unless defending player controls an
//	 Island."
//
// The restriction is ADR 0107 §2's (#1879, CR 508.1c). The defending
// player is worked out per target (CR 508.5, 508.5a), so in Commander this
// may attack any opponent who controls an Island, that opponent's
// planeswalkers and the battles they protect, and no one else. The card
// shows whom it can't attack and why.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d3e6dc0-2aed-4afa-bfa9-59d04ade9bee",
		Name:            "Ethereal Whiskergill",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
	})
}
