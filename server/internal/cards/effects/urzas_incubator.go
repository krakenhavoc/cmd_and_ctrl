package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Urza's Incubator — Artifact {3} (EDHREC rank 316):
//
//	"As this artifact enters, choose a creature type.
//	 Creature spells of the chosen type cost {2} less to cast."
//
// The CR 614.12 chosen-type prompt (ChooseCreatureTypeAsEnters) paired
// with the cost-modification group's own named example —
// SpellOfTheSourcesChosenType (cost_modifier.go) says in its doc
// comment that it is built for exactly this card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b380c04b-0bce-4328-8f0a-1425e48bee72",
		Name:         "Urza's Incubator",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Urza's Incubator"),
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Creature spells of the chosen type cost {2} less to cast.",
				CreatureSpell(), SpellOfTheSourcesChosenType()),
		},
	})
}
