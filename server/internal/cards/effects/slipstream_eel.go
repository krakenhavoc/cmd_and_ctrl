package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slipstream Eel — Creature — Fish Beast {5}{U}{U}, 6/6:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)"
//
// The restriction is ADR 0107 §2's (#1879, CR 508.1c), with the defending
// player worked out per target (CR 508.5, 508.5a). Cycling is the
// catalog's ordinary Cycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "74e6dd0f-2866-4d45-a214-8b09b837bc02",
		Name:         "Slipstream Eel",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(QuerySubtype("Island")),
		},
		Activated: []ActivatedAbility{Cycling("{1}{U}")},
	})
}
