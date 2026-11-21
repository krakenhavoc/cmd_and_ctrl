package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elephant Graveyard — Land (Arabian Nights, #5641):
//
//	"{T}: Add {C}.
//	 {T}: Regenerate target Elephant."
//
// Swarmyard's shape with one creature type: a colourless mana ability
// and a regeneration shield on any Elephant, whoever controls it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "8ada7388-fd8b-434c-a17a-bce19cf3e615",
		Name:          "Elephant Graveyard",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Regenerate target Elephant.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target Elephant", effects.HasSubtype("Elephant")),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
