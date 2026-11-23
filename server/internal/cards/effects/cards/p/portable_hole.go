package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Portable Hole — Artifact {W} (AFR):
//
//	"When this artifact enters, exile target nonland permanent an opponent
//	 controls with mana value 2 or less until this artifact leaves the
//	 battlefield."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e57a8cd-f4f0-40b2-84a0-e002751c0d40",
		Name:         "Portable Hole",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: effects.Self,
				Targets: effects.TargetPermanent("target nonland permanent an opponent controls with mana value 2 or less",
					effects.Nonland(), effects.OpponentControls(), effects.ManaValueLE(2)),
				Key:    portableHoleExileLabel,
				Effect: effects.ExileChosenTargetUntilThisLeaves("Portable Hole — the exiled card returns when Portable Hole leaves the battlefield"),
			},
			effects.UntilThisLeavesLegacyReturn("Portable Hole — return the exiled card", portableHoleExileLabel),
		},
	})
}

const portableHoleExileLabel = "Portable Hole — exile target nonland permanent an opponent controls with mana value 2 or less"
