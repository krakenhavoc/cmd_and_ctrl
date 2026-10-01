package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sea Serpent — Creature — Serpent {5}{U}, 5/5:
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
		OracleID:     "c16495fc-784d-4bac-9a68-ed437008df73",
		Name:         "Sea Serpent",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Sea Serpent — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
