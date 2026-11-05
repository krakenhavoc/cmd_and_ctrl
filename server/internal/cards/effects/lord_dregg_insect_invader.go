package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lord Dregg, Insect Invader — {3}{B} Legendary Creature — Insect Warrior
// 3/2 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 Disappear — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, create a 1/1 black
//	 Insect Warrior creature token with flying.
//	 {3}{G}, Sacrifice a token: Draw a card."
//
// Disappear is an ability word (CR 207.2c) over revolt's condition: Hidden
// Stockpile's end-step intervening if (CR 603.4). The activation's cost
// takes any token Lord Dregg's controller controls, a creature or not, and
// the sacrifice itself turns the condition on for that end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2db77e35-a54a-4e9c-8081-c97822515879",
		Name:            "Lord Dregg, Insect Invader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfRevolt("Lord Dregg — create a 1/1 black Insect Warrior creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 black Insect Warrior with flying"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, Sacrifice a token: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost:    Plus(ManaCost("{3}{G}"), b08SacrificeAToken()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
