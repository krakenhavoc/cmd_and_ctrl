package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gorilla Pack — Creature — Ape {2}{G}, 3/3:
//
//	"This creature can't attack unless defending player controls a
//	 Forest.
//	 When you control no Forests, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control a Forest.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Forests: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// No simplification.
func init() {
	q := QuerySubtype("Forest")
	Register(Spec{
		OracleID:     "f2c8814b-581b-483b-a7ae-d3d7b962aec1",
		Name:         "Gorilla Pack",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Gorilla Pack — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
