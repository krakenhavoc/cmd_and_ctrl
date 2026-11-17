package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seer of Stolen Sight — Creature — Phyrexian Warlock {2}{B}, 2/3 (March
// of the Machine):
//
//	"Menace
//	 Whenever one or more artifacts and/or creatures you control are put
//	 into a graveyard from the battlefield, surveil 1."
//
// A batched trigger (CR 603.2c): it fires once however many go at the same
// time. Types and controller are read as the permanent last existed on the
// battlefield (CR 603.10a), so a token creature, a Treasure and the Seer
// itself all count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2bdccf1-857b-48f8-a9aa-8ef999ef0632",
		Name:            "Seer of Stolen Sight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, s82MomB11ArtifactOrCreatureYouControlDied,
				"Seer of Stolen Sight — surveil 1", Do(Surveil{N: 1}))),
		},
	})
}
