package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Docent of Perfection // Final Iteration — {3}{U}{U} Creature — Insect
// Horror 5/4 // Creature — Eldrazi Insect 6/5 (Innistrad Remastered, S82
// GAP, #5948):
//
//	Front: "Flying
//	        Whenever you cast an instant or sorcery spell, create a 1/1 blue
//	        Human Wizard creature token. Then if you control three or more
//	        Wizards, transform this creature."
//	Back:  "Flying
//	        Wizards you control get +2/+1 and have flying.
//	        Whenever you cast an instant or sorcery spell, create a 1/1 blue
//	        Human Wizard creature token."
//
// The token is made first and counted afterwards (it is one of the three
// Wizards), as a separate instruction rather than an intervening "if". The
// back face's lord covers every Wizard you control, the tokens included (a
// layer 7c anthem and a layer 6 flying grant), and repeats the token
// trigger without the transform.
//
// No simplification.
func init() {
	const oracle = "85e797de-da27-4cd5-8fa4-4aef948988b2"
	wizards := TribeFilter{Tribes: []string{"Wizard"}, YoursOnly: true}
	Register(Spec{
		OracleID:        oracle,
		Name:            "Docent of Perfection",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Docent of Perfection — create a 1/1 blue Human Wizard creature token; then if you control three or more Wizards, transform this creature",
				Do(
					CreateToken{Template: TokenCard("1/1 blue Human Wizard"), N: 1},
					Instruction(func(ctx *Context) error {
						if len(MatchingBattlefield(ctx, And(Creature(), YouControl(), HasSubtype("Wizard")))) < 3 {
							return nil
						}
						return TransformThis{}.Apply(ctx)
					}),
				)),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Final Iteration",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			TribalAnthem(wizards, 2, 1),
			TribalKeywordGrant(wizards, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Final Iteration — create a 1/1 blue Human Wizard creature token",
				Do(CreateToken{Template: TokenCard("1/1 blue Human Wizard"), N: 1})),
		},
	})
}
