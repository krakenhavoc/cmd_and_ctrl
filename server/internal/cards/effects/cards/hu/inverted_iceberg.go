package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inverted Iceberg // Iceberg Titan — a transforming artifact (#2124,
// ADR 0137):
//
//	Inverted Iceberg — Artifact {1}{U}
//	  "When this artifact enters, mill a card, then draw a card.
//	   Craft with artifact {4}{U}{U}"
//	Iceberg Titan — Artifact Creature — Golem, 6/6
//	  "Whenever this creature attacks, you may tap or untap target
//	   artifact or creature."
//
// The mill and the draw run in printed order in one resolution. The
// Titan's attack trigger is Janjeet Sentry's resolution-time tap or
// untap pick behind a "you may".
//
// No simplification.
const InvertedIcebergOracleID = "3feedc63-173e-4713-8e96-8f1c9576055f"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvertedIcebergOracleID,
		Name:         "Inverted Iceberg",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Inverted Iceberg — mill a card, then draw a card", effects.Do(effects.MillCards{N: 1}, effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			effects.Craft("Craft with artifact {4}{U}{U}", "{4}{U}{U}", effects.CraftWith("artifact")),
		},
	})

	effects.Register(effects.Spec{
		OracleID:     InvertedIcebergOracleID + "#1",
		Name:         "Iceberg Titan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WheneverThisAttacks("Iceberg Titan — tap or untap target artifact or creature", effects.B38TapOrUntapTarget("Iceberg Titan")),
					effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature()))),
				"Iceberg Titan — tap or untap target artifact or creature?"),
		},
	})
}
