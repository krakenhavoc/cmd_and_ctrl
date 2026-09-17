package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kuldotha Forgemaster — Artifact Creature — Construct {5}, 3/5
// (EDHREC rank 3068):
//
//	"{T}, Sacrifice three artifacts: Search your library for an
//	 artifact card, put it onto the battlefield, then shuffle."
//
// The artifact deck's Birthing Pod for Blightsteel. The cost is a tap
// plus a sacrifice clause with a count of three (#747, SacrificeN).
// The clause says "artifacts", not "other artifacts", and the
// Forgemaster is itself an artifact, so it may be one of the three —
// tapping and sacrificing the same permanent is one legal payment,
// as on paper. The enumerator and the "Choose for me" button both put
// it last. It is a creature with {T} in its cost, so summoning
// sickness applies (CR 302.6).
//
// The search is the shared library search straight onto the
// battlefield; the searcher picks the artifact.
//
// No simplification. Declared skipped on #387 until #747 gave the cost
// a shape.
func init() {
	Register(Spec{
		OracleID:     "b0f99367-4313-4bb1-a9fd-7711bc4ce40e",
		Name:         "Kuldotha Forgemaster",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}, Sacrifice three artifacts: Search your library for an artifact card, put it onto the battlefield, then shuffle.",
			Cost:  Plus(TapCost(), SacrificeN(3, "three artifacts", Artifact())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: b03IsArtifactCard,
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Source:    item.SourceCardID,
					Reason:    "Kuldotha Forgemaster — an artifact card, onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
