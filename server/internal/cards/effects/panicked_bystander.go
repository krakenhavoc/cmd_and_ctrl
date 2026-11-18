package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Panicked Bystander // Cackling Culprit — {1}{W} Creature — Human Peasant
// 2/2 // Creature — Human Rogue 3/5 (Innistrad: Crimson Vow):
//
//	Front: "Whenever this creature or another creature you control dies, you
//	        gain 1 life.
//	        At the beginning of your end step, if you gained 3 or more life
//	        this turn, transform this creature."
//	Back:  "Whenever this creature or another creature you control dies, you
//	        gain 1 life.
//	        {1}{B}: This creature gains deathtouch until end of turn."
//
// The life-gain trigger is the source's own or any creature you control
// dying, so the Bystander's death counts. The end-step "if" is checked as
// the step begins and again as the trigger resolves (CR 603.4).
//
// No simplification.
func init() {
	const oracle = "fc58a086-2743-4465-bf16-4a20701cd3b1"
	gain := func(name string) game.TriggeredAbility {
		return WheneverACreatureYouControlDies(name+" — you gain 1 life", Do(GainLife{Amount: 1}))
	}
	gained3 := func(g *game.Game, source *game.Card) bool { return s82VowB14YouGainedAtLeast(g, source.Controller, 3) }
	Register(Spec{
		OracleID:     oracle,
		Name:         "Panicked Bystander",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			gain("Panicked Bystander"),
			On(game.EventBeginEndStep, s82VowB14YourEndStepIf(gained3),
				"Panicked Bystander — transform this creature", s82VowB14TransformIf(gained3)),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Cackling Culprit",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{gain("Cackling Culprit")},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{B}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Cackling Culprit — deathtouch until end of turn"),
		}},
	})
}
