package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fleeting Spirit — Creature — Spirit {1}{W}, 3/1 (Innistrad: Crimson Vow,
// #5495):
//
//	"{W}, Exile three cards from your graveyard: This creature gains first
//	 strike until end of turn.
//	 Discard a card: Exile this creature. Return it to the battlefield
//	 under its owner's control at the beginning of the next end step."
//
// The first ability's cost is {W} and three other cards from your graveyard
// the activator names as it is announced (ExileFromGraveyard, CR 602.2b), so
// it can't be paid with fewer, and the grant is this creature's until end
// of turn. The second is Stenn's blink (exileThisThenScheduleReturn) with a
// discard as its only cost: it exiles only the permanent the ability came
// from and it comes back as a new object at the next end step (CR 400.7,
// CR 603.7).
//
// No simplification.
func init() {
	const name = "Fleeting Spirit"
	Register(Spec{
		OracleID:     "bc79eb3c-33e2-4820-a59c-92a6271fd725",
		Name:         name,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}, Exile three cards from your graveyard: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    Plus(ManaCost("{W}"), ExileFromGraveyard(3, "three cards from your graveyard", nil)),
				Effect:  thisGainsKeywordUntilEndOfTurn("first strike", name+" — first strike until end of turn"),
			},
			{
				Label:   "Discard a card: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    DiscardACard(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return exileThisThenScheduleReturn(NewContext(g, item), item, name+" — return Fleeting Spirit")
				},
			},
		},
	})
}
