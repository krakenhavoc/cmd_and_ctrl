package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Michelangelo, the Heart — Legendary Creature — Mutant Ninja Turtle
// {1}{G}, 2/1 (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Trample
//	 Raid (the Fridge) — At the beginning of your second main phase, if
//	 you attacked this turn, put a +1/+1 counter on target creature and
//	 create a Food token.
//	 Partner—Character select"
//
// "Raid (the Fridge)" is an ability word (CR 207.2c). The trigger is the
// turn's second main phase, not an added third (AtYourSecondMainPhase's
// YourSecondMainPhaseBegan), with an intervening "if" (CR 603.4): it
// triggers only when Michelangelo's controller declared an attacker this
// turn, and checks again as it resolves. It targets any creature; with
// the target gone by resolution the ability does nothing at all, and no
// Food is made (CR 608.2b, the card's ruling). The Food waits for any
// prompt the counter asks (a CR 616 order). Partner—Character select is
// a deck-construction rule the deck importer reads (CR 702.124i).
//
// No simplification.
func init() {
	const label = "Michelangelo, the Heart — put a +1/+1 counter on target creature and create a Food token"
	Register(Spec{
		OracleID:        "f11e914c-16bc-4836-bc14-bc1b5de9cd87",
		Name:            "Michelangelo, the Heart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventStepBegan, AllOf(YourSecondMainPhaseBegan, tmcB6YouAttackedThisTurn), label,
				Do(If{Cond: tmcB6CondYouAttacked{}, Then: []Applier{
					Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
					CreateToken{Template: FoodToken(), N: 1},
				}})),
				TargetCreature("target creature")),
		},
	})
}
