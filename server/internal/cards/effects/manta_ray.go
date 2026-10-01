package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manta Ray — Creature — Fish {1}{U}{U}, 3/3:
//
//	"This creature can't attack unless defending player controls an
//	 Island.
//	 This creature can't be blocked except by blue creatures.
//	 When you control no Islands, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control an Island.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Islands: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// "Can't be blocked except by blue creatures" is an ordinary block rule
// (CR 509.1b) on the Ray itself.
//
// No simplification.
func init() {
	q := QuerySubtype("Island")
	Register(Spec{
		OracleID:     "d5129531-e4b6-454e-9c67-dae925c8f2ee",
		Name:         "Manta Ray",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Manta Ray — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnSelf(), OfColor("U"), "blue creatures"),
		},
	})
}
