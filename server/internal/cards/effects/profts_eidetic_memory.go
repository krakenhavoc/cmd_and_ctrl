package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Proft's Eidetic Memory — Legendary Enchantment {1}{U} (Murders at Karlov
// Manor, #4469):
//
//	"When Proft's Eidetic Memory enters, draw a card.
//	 You have no maximum hand size.
//	 At the beginning of combat on your turn, if you've drawn more than one
//	 card this turn, put X +1/+1 counters on target creature you control,
//	 where X is the number of cards you've drawn this turn minus one."
//
// "Cards you've drawn this turn" is the turn tally's draw count (CardsDrawn,
// CR 121), which counts draws made before the enchantment arrived (the
// ruling) and every one made this turn. The intervening "if" (CR 603.4) is
// checked as combat begins and again as the trigger resolves; with no
// creature to target the trigger is not put on the stack. X is fixed as the
// ability resolves (the ruling), from the tally then.
//
// No simplification.
func init() {
	drawnMoreThanOne := CondBoard{
		Name: "you've drawn more than one card this turn",
		Test: func(ctx *Context) bool { return ctx.Game.TurnTallyFor(ctx.Controller()).CardsDrawn > 1 },
	}
	Register(Spec{
		OracleID:     "b749e9e4-0f2a-4157-8cca-facd6b48157f",
		Name:         "Proft's Eidetic Memory",
		Completeness: CompletenessFull,
		HandSize:     []game.HandSizeStatic{{Players: game.HandSizeYou, Kind: game.HandSizeNoMaximum}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Proft's Eidetic Memory — draw a card", Do(DrawCards{N: 1})),
			Targeting(
				InterveningIf(drawnMoreThanOne,
					On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), proftDrewMoreThanOne),
						"Proft's Eidetic Memory — put X +1/+1 counters on target creature you control, where X is the cards you've drawn this turn minus one",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							t, ok := ctx.ClauseTarget(0)
							if !ok {
								return nil
							}
							x := g.TurnTallyFor(ctx.Controller()).CardsDrawn - 1
							if x <= 0 {
								return nil
							}
							return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
						})),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// proftDrewMoreThanOne is the intervening "if you've drawn more than one
// card this turn" as the trigger condition: asked as combat begins, before
// any target is chosen (CR 603.4).
func proftDrewMoreThanOne(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.TurnTallyFor(source.Controller).CardsDrawn > 1
}
