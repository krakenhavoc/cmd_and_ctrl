package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regna, the Redeemer — Legendary Creature — Angel {5}{W}, 4/4
// (Battlebond, #5529):
//
//	"Partner with Krav, the Unredeemed (When this creature enters, target
//	 player may put Krav into their hand from their library, then
//	 shuffle.)
//	 Flying
//	 At the beginning of each end step, if your team gained life this
//	 turn, create two 1/1 white Warrior creature tokens."
//
// Partner with is the shared CR 702.124j pair: the commander pairing is
// read off the oracle text by internal/deck, and the entry search is
// PartnerWith. Flying rides PrintedKeywords.
//
// "Your team" is "you" (CR 102.4). Crested Sunmare's end step: each end
// step, not only the controller's, with an intervening "if" (CR 603.4) on
// the controller's tally of life gained this turn, checked as the end step
// begins and again as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0e14f59f-021a-4ad3-a26b-8f79e680c426",
		Name:            "Regna, the Redeemer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Regna, the Redeemer", "Krav, the Unredeemed"),
			effects.On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B33EndStepAndYouGainedLifeThisTurn(ev, source, g)
			}, "Regna, the Redeemer — create two 1/1 white Warrior creature tokens",
				effects.Do(effects.If{
					Cond: effects.CondLifeGained{},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("1/1 white Warrior"), N: 2}},
				})),
		},
	})
}
