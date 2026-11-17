package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Junktown — Land (Fallout, #5469):
//
//	"{T}: Add {C}.
//	 {4}{R}, {T}, Sacrifice this land: Create three Junk tokens.
//	 (They're artifacts with "{T}, Sacrifice this token: Exile the top
//	 card of your library. You may play that card this turn. Activate
//	 only as a sorcery.")"
//
// A colorless land's mana ability and a sacrifice ability making three
// of the printed Junk tokens (JunkToken, CR 111.10t) in one creation.
// The sacrifice is part of the cost, so the land is gone whatever the
// answer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c470a802-931f-4b63-92fd-ef9cf2e796dd",
		Name:         "Junktown",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}, {T}, Sacrifice this land: Create three Junk tokens.",
			Cost:    Plus(ManaCost("{4}{R}"), TapCost(), SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  Do(CreateToken{Template: JunkToken(), N: 3}),
		}},
	})
}
