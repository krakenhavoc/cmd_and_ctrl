package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vault Guardsman — Artifact Creature — Human Advisor {5}{W}, 3/4 (Marvel
// Super Heroes Commander, #4455):
//
//	"Convoke
//	 When this creature enters, exile target artifact or creature an
//	 opponent controls until this creature leaves the battlefield."
//
// Convoke is Spec.TapCost (Lethal Scheme). The enter trigger is Banishing
// Light's exile-until-this-leaves (CR 610.3) over an artifact or creature an
// opponent controls: the exiled card returns under its owner's control as
// the Guardsman leaves, and a Guardsman gone before the trigger resolves
// exiles nothing.
//
// No simplification.
func init() {
	const exileLabel = "Vault Guardsman — exile target artifact or creature an opponent controls"
	Register(Spec{
		OracleID:     "0068f8c1-d1a2-4f7b-b39f-963acb2c023b",
		Name:         "Vault Guardsman",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets: TargetPermanent("target artifact or creature an opponent controls",
					And(Or(Artifact(), Creature()), OpponentControls())),
				Key:    exileLabel,
				Effect: exileChosenTargetUntilThisLeaves("Vault Guardsman — the exiled card returns when Vault Guardsman leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Vault Guardsman — return the exiled card", exileLabel),
		},
	})
}
