package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "8ada7388-fd8b-434c-a17a-bce19cf3e615",
		Name:          "Elephant Graveyard",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Regenerate target Elephant.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    TapCost(),
			Targets: TargetCreature("target Elephant", HasSubtype("Elephant")),
			Effect:  regenerateTheTargetPermanent,
		}},
	})
}
