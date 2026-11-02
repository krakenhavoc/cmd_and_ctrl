package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steely Resolve — Enchantment {1}{G}:
//
//	"As this enchantment enters, choose a creature type.
//	 Creatures of the chosen type have shroud. (They can't be the targets
//	 of spells or abilities.)"
//
// Adaptive Automaton's choice (ChooseCreatureTypeAsEnters) read by a
// tribal keyword grant over EVERY creature, yours and the opponents', the
// source's controller included. The grant is a layer-6 static (CR 613.1f)
// that is never locked in, so a creature of the type that enters later has
// shroud at once. Shroud stops the controller's own spells and abilities
// too (CR 702.18a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48c127f0-2857-4c36-97bc-1291b6fe4a82",
		Name:         "Steely Resolve",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Steely Resolve"),
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Chosen: true}, "shroud"),
		},
	})
}
