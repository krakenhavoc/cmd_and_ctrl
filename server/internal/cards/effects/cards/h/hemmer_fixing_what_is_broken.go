package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hemmer, Fixing What Is Broken — Legendary Creature — Aenar Artificer
// {W}{U}, 2/3 (Star Trek, #4973):
//
//	"During your turn, Spacecraft you control have hexproof.
//	 Whenever Hemmer becomes tapped, draw a card, then discard a card."
//
// The hexproof is a layer-6 grant (CR 613.1f) to every Spacecraft its
// controller controls, creature or not, read through the layered view
// (so a permanent that became a Spacecraft counts), whose predicate reads
// the turn: on an opponent's turn the Spacecraft can be targeted again.
// "Becomes tapped" is ThisBecameTapped over the tap and the attack events
// (#4534), so a vigilance attack does not loot. The loot is one draw,
// then one discard chosen as it resolves.
//
// No simplification.
func init() {
	hexproof := effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.HasSubtype("Spacecraft") && effects.IsActivePlayer(g, source.Controller)
	}, effects.KeywordHexproof)
	hexproof.Label = "During your turn, Spacecraft you control have hexproof."
	effects.Register(effects.Spec{
		OracleID:     "83d8e98d-ab97-40b1-820a-2d193a5ed261",
		Name:         "Hemmer, Fixing What Is Broken",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{hexproof},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Hemmer, Fixing What Is Broken — draw a card, then discard a card", effects.B35LootOne),
				game.Purpose{Draws: 1, Discards: 1}),
		},
	})
}
