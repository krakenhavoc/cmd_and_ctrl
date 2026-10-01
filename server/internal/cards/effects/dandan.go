package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dandân — Creature — Fish {U}{U}, 4/1:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control an Island.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Islands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:     "88929373-b2c8-4a81-a809-fed87fd5b0d7",
		Name:         "Dandân",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Dandân — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
