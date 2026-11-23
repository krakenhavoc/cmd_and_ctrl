package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "5802aa98-a9b7-4b0d-9d7c-c3938701870d",
		Name:            "Jem Lightfoote, Sky Explorer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfNoSpellFromHand("Jem Lightfoote, Sky Explorer — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
