package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merchant of Truth — Creature — Angel Detective {2}{W}{W}, 2/5:
//
//	"Flying
//	 Whenever a nontoken creature you control dies, investigate.
//	 Clues you control have exalted."
//
// Merchant triggers for itself and for every nontoken creature that
// dies with it. The exalted grant is one instance per Merchant
// (CR 113.2c): exalted is cumulative, so two Merchants give a Clue two
// instances.
func init() {
	grant := KeywordGrant(cmdrB1CluesYouControl, game.KeywordExalted)
	grant.Label = "Clues you control have exalted."
	Register(Spec{
		OracleID:        "84275aab-f718-4896-b61c-63bf49fb483b",
		Name:            "Merchant of Truth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := diedCreature(ev, g)
				return ok && !IsToken(dead) && leftUnderControlOf(ev, dead) == source.Controller
			}, "Merchant of Truth — investigate", Do(cmdrB1Investigate())),
		},
	})
}
