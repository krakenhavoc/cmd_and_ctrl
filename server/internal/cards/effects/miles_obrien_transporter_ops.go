package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Miles O'Brien, Transporter Ops — Legendary Creature — Human Artificer
// {2}{W}, 3/3 (Star Trek, #4973):
//
//	"When Miles O'Brien enters, return up to one other target creature or
//	 artifact you control to its owner's hand.
//	 At the beginning of each end step, if a permanent was returned to
//	 your hand from the battlefield this turn, draw a card."
//
// The enters trigger is Rescue, Pepper Potts' "up to one other" clause
// (Another(..).WithCount(0, 1)) with bounceEachLegalTarget's body. The
// end-step trigger fires in every player's end step and is an
// intervening "if" (CR 603.4), checked as it triggers and again as it
// resolves, against the per-player turn tally of permanents put into
// your hand from the battlefield (PlayerTurnTally.PermanentsReturnedToHand,
// keyed on the hand's owner). Any route counts: Miles' own bounce, an
// opponent's bounce of your permanent, a return as a cost; a permanent
// you control but don't own goes to its owner's hand, not yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "31e3f39f-8b18-4316-984f-4d57d2d480ca",
		Name:         "Miles O'Brien, Transporter Ops",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Miles O'Brien — return up to one other target creature or artifact you control to its owner's hand",
				bounceEachLegalTarget),
				Another(TargetPermanent("up to one other target creature or artifact you control", Or(Creature(), Artifact()), YouControl()).WithCount(0, 1))),
			InterveningIf(condPermanentReturnedToYourHand{},
				On(game.EventBeginEndStep, permanentReturnedToYourHand,
					"Miles O'Brien — draw a card if a permanent was returned to your hand from the battlefield this turn",
					drawOneCardEffect)),
		},
	})
}
