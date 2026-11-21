package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Annex Sentry — Artifact Creature — Phyrexian Cleric {2}{W}, 1/4 (March of
// the Machine: Phyrexia: All Will Be One):
//
//	"Toxic 1
//	 When this creature enters, exile target artifact or creature an
//	 opponent controls with mana value 3 or less until this creature leaves
//	 the battlefield."
//
// Toxic is the engine's damage tail (CR 702.164). The entry trigger is the
// Oblivion Ring family's (exileChosenTargetUntilThisLeaves) over an artifact
// or creature an opponent controls whose mana value is 3 or less (CR 202.3),
// read at targeting and again as the trigger resolves (CR 608.2b). The card
// returns under its owner's control when the Sentry leaves (CR 610.3c); if
// it has already left when the trigger resolves nothing is exiled
// (CR 610.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "be31fa46-d5ac-4c85-9c19-3f0e2b1f223b",
		Name:            "Annex Sentry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"toxic 1"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Annex Sentry — exile target artifact or creature an opponent controls with mana value 3 or less until this creature leaves the battlefield",
				effects.ExileChosenTargetUntilThisLeaves("Annex Sentry — the exiled card returns when Annex Sentry leaves the battlefield")),
				effects.TargetPermanent("target artifact or creature an opponent controls with mana value 3 or less",
					effects.Or(effects.Artifact(), effects.Creature()), effects.OpponentControls(), effects.ManaValueLE(3))),
		},
	})
}
