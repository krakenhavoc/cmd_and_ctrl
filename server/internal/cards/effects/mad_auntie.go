package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mad Auntie — Creature — Goblin Shaman {2}{B}, 2/2 (Lorwyn, #5593):
//
//	"Other Goblin creatures you control get +1/+1.
//	 {T}: Regenerate another target Goblin."
//
// A Goblin lord (layer 7c, CR 613.4c) and a regeneration row. "Another
// target Goblin" is any Goblin permanent but Mad Auntie herself, anyone's;
// the shield replaces its next destruction this turn (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b582b696-17fa-4eea-8124-293f9bd3a208",
		Name:         "Mad Auntie",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{
			lrwB1RegenerateTargetRow("{T}: Regenerate another target Goblin.", TapCost(),
				Another(TargetPermanent("another target Goblin", HasSubtype("Goblin")))),
		},
	})
}
