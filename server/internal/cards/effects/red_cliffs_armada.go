package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Red Cliffs Armada — Creature — Human Soldier {4}{U}, 5/4:
//
//	"This creature can't attack unless defending player controls an
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
		OracleID:     "063884e0-1f5e-4be9-930b-e73895b2fa41",
		Name:         "Red Cliffs Armada",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
	})
}
