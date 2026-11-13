package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boulderbranch Golem — Artifact Creature — Golem {7}, 6/5 (The Brothers'
// War, #5656): "Prototype {3}{G} — 3/3", "When this creature enters, you
// gain life equal to its power."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// The trigger reads the Golem's power as it resolves (CR 608.2h): 3 for
// a prototyped Golem, 6 for one cast normally, counters and continuous
// effects included, and its last-known power if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82caf3e2-da42-4851-b4db-546e2403c4bf",
		Name:         "Boulderbranch Golem",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Boulderbranch Golem — gain life equal to its power",
				Do(GainLifeEqualToSubject{Stat: SubjectPower})),
		},
	})
}
