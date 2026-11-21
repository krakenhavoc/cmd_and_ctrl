package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrangler of the Damned — Creature — Human Soldier {3}{W}{U}, 1/4
// (Outlaws of Thunder Junction, #4468):
//
//	"Flash
//	 At the beginning of your end step, if you haven't cast a spell from
//	 your hand this turn, create a 2/2 white Spirit creature token with
//	 flying."
//
// The token is an intervening "if" (CR 603.4) over the turn's cast tally
// (AtYourEndStepIfNoSpellFromHand, #5235): checked as the end step
// begins and again as the ability resolves. Flashing the Wrangler in
// from the hand is itself a spell cast from the hand, so it makes no
// Spirit the turn it is cast that way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d8fded4-0393-43c2-82a1-383f29de65ce",
		Name:            "Wrangler of the Damned",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfNoSpellFromHand("Wrangler of the Damned — create a 2/2 white Spirit with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 white Spirit with flying"), N: 1})),
		},
	})
}
