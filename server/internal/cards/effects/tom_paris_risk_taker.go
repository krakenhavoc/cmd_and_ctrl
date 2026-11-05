package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tom Paris, Risk-Taker — Legendary Creature — Human Pilot {4}{U}{R}, 4/4
// (Star Trek Commander, #4974):
//
//	"Seek out — Whenever you draw your second card each turn, creatures
//	 you control gain flying, double strike, vigilance, and haste until
//	 end of turn."
//
// "Seek out" is an ability word (CR 207.2c). The trigger is Atlantean
// Cavalry's: it reads the turn's draw tally after the draw, so it fires
// on the second card and no other, on any player's turn. The set is
// fixed as the effect begins (CR 611.2c): a creature that arrives later
// in the turn gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "327458f1-d3c4-4e66-8abd-a6533bef63f7",
		Name:         "Tom Paris, Risk-Taker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41DrewYourSecondCardThisTurn(ev, source, g)
			}, "Tom Paris, Risk-Taker — creatures you control gain flying, double strike, vigilance, and haste until end of turn",
				Do(GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"flying", "double strike", "vigilance", "haste"},
					Label:    "Tom Paris, Risk-Taker — seek out",
				})),
		},
	})
}
