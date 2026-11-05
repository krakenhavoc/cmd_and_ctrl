package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	hexproof := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.HasSubtype("Spacecraft") && isActivePlayer(g, source.Controller)
	}, KeywordHexproof)
	hexproof.Label = "During your turn, Spacecraft you control have hexproof."
	Register(Spec{
		OracleID:     "83d8e98d-ab97-40b1-820a-2d193a5ed261",
		Name:         "Hemmer, Fixing What Is Broken",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{hexproof},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Hemmer, Fixing What Is Broken — draw a card, then discard a card", b35LootOne),
				game.Purpose{Draws: 1, Discards: 1}),
		},
	})
}
