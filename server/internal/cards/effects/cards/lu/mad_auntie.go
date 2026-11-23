package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b582b696-17fa-4eea-8124-293f9bd3a208",
		Name:         "Mad Auntie",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{
			effects.LrwB1RegenerateTargetRow("{T}: Regenerate another target Goblin.", effects.TapCost(),
				effects.Another(effects.TargetPermanent("another target Goblin", effects.HasSubtype("Goblin")))),
		},
	})
}
