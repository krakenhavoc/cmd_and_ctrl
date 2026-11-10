package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Homicide Investigator — Creature — Human Detective {1}{B}, 2/2
// (Murders at Karlov Manor, #4469):
//
//	"Whenever one or more nontoken creatures you control die,
//	 investigate. This ability triggers only once each turn. (Create a
//	 Clue token. It's an artifact with "{2}, Sacrifice this token: Draw
//	 a card.")"
//
// The death is The Skullspore Nexus's batch trigger (CR 603.2c,
// skullsporeNontokenCreatureYouControlDied: a nontoken creature you
// controlled as it died), so a wrath is one Clue. "Only once each turn"
// is Baron Bertram Graywater's TriggersOnlyOnceEachTurn over the batch.
// It is a leaves-the-battlefield ability, so it looks back (CR 603.10a)
// and triggers on the Investigator's own death (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc4d5602-47cb-47c8-8a43-3b840e12b79c",
		Name:         "Homicide Investigator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventLTB, skullsporeNontokenCreatureYouControlDied,
				"Homicide Investigator — investigate (nontoken creatures you control died)",
				Do(CreateToken{Template: ClueToken(), N: 1})))),
		},
	})
}
