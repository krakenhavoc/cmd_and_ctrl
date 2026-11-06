package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystalline Entity — Artifact Creature — Elemental {8}, 7/7 (Star
// Trek, #4973):
//
//	"Flying
//	 When this creature enters, if you cast it, destroy all nonartifact
//	 creatures."
//
// Flying is the printed keyword. "If you cast it" is Cresting
// Mosasaurus's CR 603.4 intervening if (crestingMosasaurusYouCastIt): the
// permanent came from the stack and its cast record names the player who
// controls it now, so an Entity put onto the battlefield any other way
// destroys nothing. The sweep is one simultaneous destruction of every
// creature that is not an artifact (the Entity, an artifact, survives);
// indestructible and regeneration apply (CR 701.8). Its Purpose is a
// partial destroy sweep (ADR 0126 §6), Phyrexian Scriptures' declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ca6f43d8-1498-4a13-b1cc-2d135b24a8b6",
		Name:            "Crystalline Entity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, crestingMosasaurusYouCastIt),
				"Crystalline Entity — destroy all nonartifact creatures",
				Do(DestroyAllMatching{Match: And(Creature(), Not(Artifact()))})),
		},
	})
}
