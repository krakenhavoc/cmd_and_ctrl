package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ovinomancer — Creature — Human Wizard {2}{U}, 0/1 (Multiverse Gift Box,
// #5632):
//
//	"When this creature enters, sacrifice it unless you return three
//	 basic lands you control to their owner's hand.
//	 {T}, Return this creature to its owner's hand: Destroy target
//	 creature. It can't be regenerated. That creature's controller
//	 creates a 0/1 green Sheep creature token."
//
// The enters trigger is Bull Elephant's (returnLandsOrSacrifice) with
// three basic lands: exactly three or none.
//
// The activated ability returns the Wizard as its cost (ReturnThis, Maze's
// End's), so it can't be used the turn it arrives (CR 302.6) and it is
// back in hand before the ability resolves. The target's controller is
// read before the destroy, as Beast Within's is, and the Sheep is created
// even when the creature survives (an indestructible one), since the
// sentence is not conditional; a target that left in response is illegal
// and the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "98026d30-a29f-4590-968f-b1b36926a275",
		Name:         "Ovinomancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ovinomancerReturn.trigger("Ovinomancer — sacrifice it unless you return three basic lands you control to their owner's hand"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Return this creature to its owner's hand: Destroy target creature. It can't be regenerated. That creature's controller creates a 0/1 green Sheep creature token.",
			Purpose: game.Purpose{Targets: game.ForTargets(RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 0, TokenToughness: 1})), Answers: game.AnswerRemove},
			Cost:    Plus(TapCost(), ReturnThis()),
			Targets: TargetCreature("target creature"),
			Effect:  ovinomancerDestroy,
		}},
	})
}

var ovinomancerReturn = newReturnLandsOrSacrifice(
	"Ovinomancer — return three basic lands you control to their owner's hand (or sacrifice Ovinomancer)", 3,
	"choose-cards/ovinomancer-return-three-basic-lands", "mgb/ovinomancer-after-return",
	func(_ *game.Card, land game.Card) bool { return IsBasicLand(land) })

// ovinomancerDestroy destroys the target with no regeneration, then gives
// its controller a 0/1 green Sheep.
func ovinomancerDestroy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
		return nil
	}
	controller, ok := controllerOfTarget(ctx, t.ID)
	if !ok {
		return nil
	}
	if err := (DestroyTarget{Target: t.ID, CantBeRegenerated: true}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Controller: controller, Template: TokenCard("0/1 green Sheep"), N: 1}.Apply(ctx)
}
