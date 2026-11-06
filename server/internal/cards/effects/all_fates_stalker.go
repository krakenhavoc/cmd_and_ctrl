package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// All-Fates Stalker — Creature — Drix Assassin {3}{W}, 2/3 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, exile up to one target non-Assassin
//	 creature until this creature leaves the battlefield.
//	 Warp {1}{W} (You may cast this card from your hand for its warp
//	 cost. Exile this creature at the beginning of the next end step,
//	 then you may cast it from exile on a later turn.)"
//
// Earth Kingdom Jailer's entry exile (exileChosenTargetUntilThisLeaves,
// CR 610.3) over any creature that is not an Assassin, whoever controls
// it; "up to one" lets the Stalker enter with no target. If the Stalker
// has left before the trigger resolves, nothing is exiled (CR 610.3c).
//
// Warp is the shared constructor (Bygone Colossus). The two halves meet
// on purpose: a warped Stalker is exiled at the next end step, and that
// exit returns the creature it took, so the warp line is a one-turn
// removal that can come back on a later turn to take something again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "da9ec010-8b42-4927-b602-4e96bced4e99",
		Name:             "All-Fates Stalker",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("All-Fates Stalker — exile up to one target non-Assassin creature until this creature leaves the battlefield",
				exileChosenTargetUntilThisLeaves("All-Fates Stalker — the exiled creature returns when All-Fates Stalker leaves the battlefield")),
				TargetCreature("up to one target non-Assassin creature", Not(OfSubtype("Assassin"))).WithCount(0, 1)),
		},
	})
}
