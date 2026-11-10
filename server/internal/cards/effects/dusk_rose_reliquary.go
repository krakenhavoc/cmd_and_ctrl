package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dusk Rose Reliquary — Artifact {W} (The Lost Caverns of Ixalan, #4470):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or
//	 creature.
//	 Ward {2}
//	 When this artifact enters, exile target artifact or creature an
//	 opponent controls until this artifact leaves the battlefield."
//
// The sacrifice is Anchor to Reality's additional cost (CR 601.2b, 601.2h:
// paid as the spell is cast, so a spell with nothing to sacrifice cannot be
// cast). Ward is the table's ward trigger (CR 702.21a). The enters
// trigger is All-Fates Stalker's exile "until this leaves the
// battlefield" (CR 610.3): the card returns under its owner's control
// right after the Reliquary leaves (CR 610.3c), and nothing is exiled if
// the Reliquary has already left by the time the trigger resolves
// (CR 610.3b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "dca7102e-c020-4fdc-9b59-f186e3e615a5",
		Name:           "Dusk Rose Reliquary",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Dusk Rose Reliquary — ward {2}"),
			Targeting(WhenThisEnters("Dusk Rose Reliquary — exile target artifact or creature an opponent controls until this artifact leaves the battlefield",
				exileChosenTargetUntilThisLeaves("Dusk Rose Reliquary — the exiled card returns when Dusk Rose Reliquary leaves the battlefield")),
				TargetPermanent("target artifact or creature an opponent controls", Or(Artifact(), Creature()), OpponentControls())),
		},
	})
}
