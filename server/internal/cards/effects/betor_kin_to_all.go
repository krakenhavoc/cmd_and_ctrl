package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Betor, Kin to All — Legendary Creature — Spirit Dragon
// {2}{W}{B}{G}, 5/7 (EDHREC rank 3221):
//
//	"Flying
//	 At the beginning of your end step, if creatures you control
//	 have total toughness 10 or greater, draw a card. Then if
//	 creatures you control have total toughness 20 or greater, untap
//	 each creature you control. Then if creatures you control have
//	 total toughness 40 or greater, each opponent loses half their
//	 life, rounded up."
//
// The toughness commander. Flying rides PrintedKeywords. One
// end-step trigger with an intervening-if at 10 (CR 603.4 — checked
// when the end step begins and again on resolution), then the 20
// and 40 thresholds re-read in turn as each clause resolves. Total
// toughness is the sum of current toughness — layers and counters
// both — over the creatures the controller controls, the Dragon
// itself included. The untap is every tapped creature the controller
// controls; the life loss is half of each opponent's life as it
// stands, rounded up, as a life change rather than damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "826a0db6-195a-4697-9bd6-f544b279e3cf",
		Name:            "Betor, Kin to All",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30EndStepAndTotalToughnessAtLeast(ev, source, g, 10)
			}, "Betor, Kin to All — draw at 10 toughness, untap at 20, halve opponents' life at 40", b30BetorEndStep),
		},
	})
}
