package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0f208cd3-7673-4380-9cb8-5b6f09966ebd",
		Name:         "Lake-town Toymaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				InterveningIf(CondCardsDrawn{AtLeast: 2},
					On(game.EventStepBegan, AllOf(StepBegan(game.StepBeginCombat, true), youHaveDrawnTwoOrMoreCards),
						"Lake-town Toymaker — another target creature you control gets +3/+0 and first strike until end of turn",
						func(g *game.Game, item *game.StackItem) error {
							ctx := NewContext(g, item)
							id, ok := b16FirstLegalTargetCard(ctx)
							if !ok {
								return nil
							}
							if err := (BoostUntilEOT{Target: id, Power: 3, Label: "Lake-town Toymaker — +3/+0"}).Apply(ctx); err != nil {
								return err
							}
							return GrantKeywordUntilEOT{Target: id, Keywords: []string{"first strike"}, Label: "Lake-town Toymaker"}.Apply(ctx)
						})),
				Another(TargetCreature("another target creature you control", YouControl()))),
				ForTargets(BenefitUntilEOT(0, 3, 0, "first strike"))),
		},
	})
}

// youHaveDrawnTwoOrMoreCards is the trigger-time half of the intervening
// "if you've drawn two or more cards this turn" (CR 603.4): the ability is
// not put on the stack, and no target is asked, unless it already holds.
func youHaveDrawnTwoOrMoreCards(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return g.TurnTallyFor(source.Controller).CardsDrawn >= 2
}
