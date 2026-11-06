package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jem Lightfoote, Sky Explorer — Legendary Creature — Human Scout
// {2}{W}{U}, 3/3 (Outlaws of Thunder Junction, #4468):
//
//	"Flying, vigilance
//	 At the beginning of your end step, if you haven't cast a spell from
//	 your hand this turn, draw a card."
//
// The draw is an intervening "if" (CR 603.4) over the turn's cast tally
// (AtYourEndStepIfNoSpellFromHand, #5235): checked as the end step
// begins and again as the ability resolves, so a spell cast from the
// hand in response turns it off, and one cast from exile (a plotted
// card), a graveyard or the command zone does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5802aa98-a9b7-4b0d-9d7c-c3938701870d",
		Name:            "Jem Lightfoote, Sky Explorer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfNoSpellFromHand("Jem Lightfoote, Sky Explorer — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
