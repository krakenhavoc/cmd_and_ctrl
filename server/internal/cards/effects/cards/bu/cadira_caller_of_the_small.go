package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cadira, Caller of the Small — Legendary Creature — Orc Ranger
// {1}{G}{W}, 3/3 (EDHREC rank 3843):
//
//	"Trample
//	 Whenever Cadira deals combat damage to a player, for each token
//	 you control, create a 1/1 white Rabbit creature token."
//
// The token doubler on a body. Trample rides PrintedKeywords; the
// trigger fires when Cadira herself deals combat damage to a player
// (trample's overflow included), and the count is taken as the
// trigger resolves — every token the controller controls, Treasures
// and Clues included, as printed. The Rabbits enter together, so
// none of them counts itself; next combat, each one does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fcf22321-2f55-41ca-b8e1-4792540ba3ee",
		Name:            "Cadira, Caller of the Small",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B36SelfDealtCombatDamageToPlayer(ev, source, g)
			}, "Cadira, Caller of the Small — a 1/1 white Rabbit for each token you control", effects.B36RabbitsPerTokenYouControl),
		},
	})
}
