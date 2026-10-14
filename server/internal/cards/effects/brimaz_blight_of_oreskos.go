package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brimaz, Blight of Oreskos — Legendary Creature — Phyrexian Cat {2}{W}{B},
// 3/4 (March of the Machine):
//
//	"Whenever you cast a Phyrexian creature or artifact creature spell,
//	 incubate X, where X is that spell's mana value.
//	 At the beginning of each end step, if a Phyrexian died under your
//	 control this turn, proliferate."
//
// The first ability is complete. The second is NOT implemented: the
// engine tallies "a creature died this turn" table-wide
// (TurnTally.CreaturesDied) but not per subtype or per controller, so an
// intervening-if on "a Phyrexian died under your control" has nothing to
// read. Shipping it as "any creature died" would be stronger than
// printed, so the card ships without the proliferate. The seam is a
// per-controller, per-subtype death tally beside EnteredSubtypes.
func init() {
	Register(Spec{
		OracleID:     "a6e7cbfc-6d5a-4a67-9dad-d512d3c97bda",
		Name:         "Brimaz, Blight of Oreskos",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The end-step proliferate isn't implemented — only the incubate on casting a Phyrexian creature or artifact creature spell happens.",
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Creature(), Or(HasSubtype("Phyrexian"), Artifact())),
				"Brimaz, Blight of Oreskos — incubate X, where X is that spell's mana value",
				incubateThatSpellsManaValue),
		},
	})
}
