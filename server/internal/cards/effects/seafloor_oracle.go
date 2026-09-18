package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seafloor Oracle — Creature — Merfolk Wizard {2}{U}{U}, 2/3 (EDHREC
// rank 4517):
//
//	"Whenever a Merfolk you control deals combat damage to a player,
//	 draw a card."
//
// The Merfolk deck's Coastal Piracy. A tribe built out of small
// unblockable bodies wants exactly this: every connection is a card,
// and the Oracle itself is a Merfolk, so it draws off its own swing.
//
// One trigger per DAMAGE EVENT, so three Merfolk connecting with
// three different players draw three cards — which is what the card
// says, since each is a separate "deals combat damage to a player".
// First strike and regular damage are separate events too, so a
// double striker draws twice, as printed.
//
// Subtype is read post-layer off the damage source, so a changeling
// or a lorded-into-Merfolk creature counts. A source that is already
// gone by the time the event is harvested does not fire the trigger
// — weaker than printed, never stronger; see
// b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f9f719cb-7778-4109-bbd3-4504ee1b5f49",
		Name:         "Seafloor Oracle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer(ev, source, g, "Merfolk")
			}, "Seafloor Oracle — draw a card", func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
