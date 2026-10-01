package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bog Serpent — Creature — Serpent {5}{B}, 5/5:
//
//	"This creature can't attack unless defending player controls a
//	 Swamp.
//	 When you control no Swamps, sacrifice this creature."
//
// ADR 0107 (#1858, #1879). The attack restriction is §2's (CR 508.1c),
// with the defending player worked out per target (CR 508.5, 508.5a), so
// in Commander it may attack only the opponents who control a Swamp.
// The sacrifice is §1's CR 603.8 state trigger over its own controller's
// Swamps: it triggers as soon as the last one is gone and not again while
// it waits or is on the stack. Both halves read one PermanentQuery.
//
// No simplification.
func init() {
	q := QuerySubtype("Swamp")
	Register(Spec{
		OracleID:     "9a9877b5-9f75-4c83-b11b-f006aecb075b",
		Name:         "Bog Serpent",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CantAttackUnlessDefendingPlayerControls(q),
		},
		Triggered: []game.TriggeredAbility{
			WhenYouControlNo(q, "Bog Serpent — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
