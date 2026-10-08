package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const invertedIcebergOracleID = "3feedc63-173e-4713-8e96-8f1c9576055f"

func init() {
	Register(Spec{
		OracleID:     invertedIcebergOracleID,
		Name:         "Inverted Iceberg",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Inverted Iceberg — mill a card, then draw a card", Do(MillCards{N: 1}, DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{
			Craft("Craft with artifact {4}{U}{U}", "{4}{U}{U}", CraftWith("artifact")),
		},
	})

	Register(Spec{
		OracleID:     invertedIcebergOracleID + "#1",
		Name:         "Iceberg Titan",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					WheneverThisAttacks("Iceberg Titan — tap or untap target artifact or creature", b38TapOrUntapTarget("Iceberg Titan")),
					TargetPermanent("target artifact or creature", Or(Artifact(), Creature()))),
				"Iceberg Titan — tap or untap target artifact or creature?"),
		},
	})
}
