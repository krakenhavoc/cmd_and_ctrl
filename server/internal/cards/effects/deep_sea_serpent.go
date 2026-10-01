package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deep-Sea Serpent — Creature — Serpent {4}{U}{U}, 5/5:
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
		OracleID:     "7fbb98cc-585c-4184-97f5-9b3d3ebdb1e5",
		Name:         "Deep-Sea Serpent",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
	})
}
