package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lake-town Toymaker — Creature — Human Artificer {3}{W}, 3/4:
//
//	"At the beginning of combat on your turn, if you've drawn two or
//	 more cards this turn, another target creature you control gets +3/+0
//	 and gains first strike until end of turn."
//
// An intervening "if" (CR 603.4): the trigger is put on the stack only if
// the turn's draw tally already holds two, and is checked again as it
// resolves (CondCardsDrawn; the tally only rises, so a held condition
// reads nothing another trigger could change). "Another" excludes the
// Toymaker itself; the target is chosen as the trigger goes on the stack.
//
// No simplification. The Hobbit (#4454).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f208cd3-7673-4380-9cb8-5b6f09966ebd",
		Name:         "Lake-town Toymaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.InterveningIf(effects.CondCardsDrawn{AtLeast: 2},
					effects.On(game.EventStepBegan, effects.AllOf(effects.StepBegan(game.StepBeginCombat, true), youHaveDrawnTwoOrMoreCards),
						"Lake-town Toymaker — another target creature you control gets +3/+0 and first strike until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							ctx := effects.NewContext(g, item)
							id, ok := effects.B16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							if err := (effects.BoostUntilEOT{Target: id, Power: 3, Label: "Lake-town Toymaker — +3/+0"}).Apply(ctx); err != nil {
								return err
							}
							return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Lake-town Toymaker"}.Apply(ctx)
						})),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "first strike"))),
		},
	})
}

// youHaveDrawnTwoOrMoreCards is the trigger-time half of the intervening
// "if you've drawn two or more cards this turn" (CR 603.4): the ability is
// not put on the stack, and no target is asked, unless it already holds.
func youHaveDrawnTwoOrMoreCards(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.TurnTallyFor(source.Controller).CardsDrawn >= 2
}
