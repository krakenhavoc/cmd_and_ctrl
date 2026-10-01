package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slipstream Serpent — Creature — Serpent {7}{U}, 6/6:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 When you control no Islands, sacrifice this creature.
//	 Morph {5}{U}"
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control an Island.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Islands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// Morph is the catalog's ordinary morph offer. Face down it has none of
// these abilities (CR 708.2a), so a face-down Serpent neither checks
// Islands nor is restricted.
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:     "aa1152cb-255f-43fa-81f5-430304ce4d98",
		Name:         "Slipstream Serpent",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Slipstream Serpent — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		AlternativeCosts: []game.AlternativeCost{Morph("{5}{U}")},
	})
}
