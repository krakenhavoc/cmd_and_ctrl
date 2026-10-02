package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Rock Sled — Creature — Goblin {1}{R}, 3/1:
//
//	"Trample
//	 This creature doesn't untap during your untap step if it attacked
//	 during your last turn.
//	 This creature can't attack unless defending player controls a
//	 Mountain."
//
// Two seams already in the engine and one new one. The attack clause is
// ADR 0107 §2's (CR 508.1c), asked of each target's defending player
// (CR 508.5), so in Commander it may attack only the opponents who control
// a Mountain. The untap clause is an UntapStepRestriction (CR 502.3) that
// reads ADR 0108 §6's record of what this creature attacked with during
// its controller's last turn: only a declaration counts (CR 508.4), only
// this object does (CR 400.7), and only its controller's own turn does, not
// an opponent's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ada3247e-ec5e-499d-bc15-1d9dd80a59ae",
		Name:            "Goblin Rock Sled",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static:          []game.StaticAbility{CantAttackUnlessDefendingPlayerControls(QuerySubtype("Mountain"))},
		UntapStepRestrictions: []game.UntapStepRestriction{
			doesntUntapIfItAttackedDuringYourLastTurn(selfOnly),
		},
	})
}
