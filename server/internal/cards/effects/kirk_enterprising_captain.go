package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kirk, Enterprising Captain — Legendary Creature — Human Officer {3}{R},
// 4/3 (Star Trek Commander, #4974):
//
//	"Menace
//	 Whenever Kirk attacks while creatures you control have total power
//	 8 or greater, exile the top card of your library. You may play that
//	 card this turn."
//
// Menace is the printed keyword. "While" is part of the trigger event,
// not an intervening "if" (CR 603.4 is the "if" form): the total power
// of the creatures Kirk's controller controls — current power, counters
// and pumps included, Kirk himself counted — is read as he is declared
// as an attacker, and the trigger is not rechecked as it resolves. It
// exiles the top card of its controller's library with permission to
// play it this turn.
//
// No simplification.
func init() {
	powerAtLeastEight := creaturesYouControlTotalPowerAtLeast(8)
	Register(Spec{
		OracleID:        "6a70e571-6932-45d8-9f14-ede4a561cc49",
		Name:            "Kirk, Enterprising Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return Self(ev, source, ch, g) && powerAtLeastEight(g, source.Controller)
			}, "Kirk, Enterprising Captain — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := b12ImpulseExileForTurn(g, item, 1)
					return err
				}),
		},
	})
}
