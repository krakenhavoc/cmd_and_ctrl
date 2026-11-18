package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Portable Hole — Artifact {W} (AFR):
//
//	"When this artifact enters, exile target nonland permanent an opponent
//	 controls with mana value 2 or less until this artifact leaves the
//	 battlefield."
func init() {
	Register(Spec{
		OracleID:     "5e57a8cd-f4f0-40b2-84a0-e002751c0d40",
		Name:         "Portable Hole",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Targets: TargetPermanent("target nonland permanent an opponent controls with mana value 2 or less",
					Nonland(), OpponentControls(), ManaValueLE(2)),
				Key:    portableHoleExileLabel,
				Effect: exileChosenTargetUntilThisLeaves("Portable Hole — the exiled card returns when Portable Hole leaves the battlefield"),
			},
			UntilThisLeavesLegacyReturn("Portable Hole — return the exiled card", portableHoleExileLabel),
		},
	})
}

const portableHoleExileLabel = "Portable Hole — exile target nonland permanent an opponent controls with mana value 2 or less"
