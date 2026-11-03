package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ennis, Debate Moderator — Legendary Creature — Human Cleric {1}{W}, 1/1
// (Secrets of Strixhaven, #4456):
//
//	"When Ennis enters, exile up to one other target creature you control.
//	 Return that card to the battlefield under its owner's control at the
//	 beginning of the next end step.
//	 At the beginning of your end step, if one or more cards were put into
//	 exile this turn, put a +1/+1 counter on Ennis."
//
// The first ability is Roon's delayed blink (exileTargetsThenScheduleReturn)
// over an "up to one other" target: with no target nothing is exiled and
// nothing is scheduled, a target that became illegal stays where it is (CR
// 608.2b), and the card comes back as a new object at the next end step
// (CR 400.7, CR 603.7); a token exiled this way does not return (CR 111.7).
//
// The second is one trigger with an intervening "if" (CR 603.4). Any card
// put into exile by anyone this turn counts, Ennis's own blink included
// (socB24CardsWerePutIntoExileThisTurn reads the turn's zone moves). The
// tally only rises, so the check at the step is the check at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97e1fcdf-04e8-40cf-b04a-8505d9b7cecc",
		Name:         "Ennis, Debate Moderator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Ennis, Debate Moderator — exile up to one other target creature you control; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return exileTargetsThenScheduleReturn(NewContext(g, item),
							"Ennis, Debate Moderator — return the exiled creature")
					}),
				Another(UpToOneTargetCreature("up to one other target creature you control", YouControl())),
			),
			On(game.EventBeginEndStep, AllOf(ByYou, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return socB24CardsWerePutIntoExileThisTurn(g)
			}), "Ennis, Debate Moderator — put a +1/+1 counter on Ennis",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
