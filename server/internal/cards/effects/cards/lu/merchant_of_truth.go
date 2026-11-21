package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	grant := effects.KeywordGrant(effects.CmdrB1CluesYouControl, game.KeywordExalted)
	grant.Label = "Clues you control have exalted."
	effects.Register(effects.Spec{
		OracleID:        "84275aab-f718-4896-b61c-63bf49fb483b",
		Name:            "Merchant of Truth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{grant},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !effects.IsToken(dead) && effects.LeftUnderControlOf(ev, dead) == source.Controller
			}, "Merchant of Truth — investigate", effects.Do(effects.CmdrB1Investigate())),
		},
	})
}
