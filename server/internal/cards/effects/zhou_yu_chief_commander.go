package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zhou Yu, Chief Commander — Legendary Creature — Human Soldier {5}{U}{U}, 8/8:
//
//	"Zhou Yu can't attack unless defending player controls an Island."
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
		OracleID:     "0b4742b7-e769-4354-beaf-6b4d18768ec1",
		Name:         "Zhou Yu, Chief Commander",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
	})
}
