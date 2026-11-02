package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectacular Spider-Man — Legendary Creature — Spider Human Hero {1}{W}, 3/2
// (Marvel's Spider-Man, #4460):
//
//	"Flash
//	 {1}: Spectacular Spider-Man gains flying until end of turn.
//	 {1}, Sacrifice Spectacular Spider-Man: Creatures you control gain
//	 hexproof and indestructible until end of turn."
//
// The flying grant is Endling's self grant. The sacrifice is a cost, so he is
// already gone when the ability resolves and is not among the creatures that
// gain the keywords; the set is fixed as it resolves (CR 611.2c), exactly as
// Bulwark Ox's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eed5e0cc-8f88-4779-bf10-6a823c72df7f",
		Name:            "Spectacular Spider-Man",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}: Spectacular Spider-Man gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{1}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("flying", "Spectacular Spider-Man — flying until end of turn"),
			},
			{
				Label:   "{1}, Sacrifice Spectacular Spider-Man: Creatures you control gain hexproof and indestructible until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerSurvive, Saves: &game.Saves{Matches: game.SweepCreatures, Partial: true}},
				Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{
						Match:    And(Creature(), YouControl()),
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Spectacular Spider-Man — hexproof and indestructible until end of turn",
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
