package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brightling — Creature — Shapeshifter {1}{W}{W}, 3/3 (Battlebond, #5529):
//
//	"{W}: This creature gains vigilance until end of turn.
//	 {W}: This creature gains lifelink until end of turn.
//	 {W}: Return this creature to its owner's hand.
//	 {1}: This creature gets +1/-1 or -1/+1 until end of turn."
//
// Endling's white twin. The two grants are layer-6 keywords on the source
// for the turn; the return is the ability's effect, not a cost, so it can
// be answered, and a Brightling that left and came back is a new object
// and stays (CR 400.7). The +1/-1 or -1/+1 choice is made as the ability
// resolves, a two-way question then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "909ff4a8-8bdd-4478-bd43-ce5175a02d68",
		Name:         "Brightling",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{W}: This creature gains vigilance until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{W}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("vigilance", "Brightling — vigilance until end of turn"),
			},
			{
				Label:   "{W}: This creature gains lifelink until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{W}"),
				Effect:  thisGainsKeywordUntilEndOfTurn("lifelink", "Brightling — lifelink until end of turn"),
			},
			{
				Label:   "{W}: Return this creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{W}"),
				Effect:  returnThisPermanentToOwnersHand,
			},
			{
				Label:   "{1}: This creature gets +1/-1 or -1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return MayChoice{
						Question: "Brightling — +1/-1 or -1/+1 until end of turn?",
						YesLabel: "+1/-1",
						NoLabel:  "-1/+1",
						Then:     brightlingPump,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// brightlingPump is Brightling's choice: +1/-1 on a yes, -1/+1 on a no. An
// on-disk key: never renamed, never reused.
var brightlingPump = MayYesNo("confirm/brightling-pump",
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return thisGetsUntilEndOfTurn(1, -1, "Brightling — +1/-1 until end of turn")(ctx.Game, ctx.Item)
	},
	func(ctx *Context, _ game.ConfirmAnswer) error {
		return thisGetsUntilEndOfTurn(-1, 1, "Brightling — -1/+1 until end of turn")(ctx.Game, ctx.Item)
	})
