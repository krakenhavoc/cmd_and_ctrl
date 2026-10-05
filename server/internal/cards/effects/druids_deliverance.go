package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Druid's Deliverance — Instant {1}{G} (EDHREC rank 3959):
//
//	"Prevent all combat damage that would be dealt to you this turn.
//	 Populate. (Create a token that's a copy of a creature token you
//	 control.)"
//
// A Fog that leaves a creature behind. In a token deck the populate
// half is the reason it is played over the one-mana Fogs: surviving
// the swing and copying a 4/4 Angel in the same card is a two-for-one.
//
// It is in the batch because the first clause is NOT the Fog the
// catalog already has, and the distinction is the whole card.
//
// # "To you", not "to everything"
//
// Fog prevents all combat damage the turn deals, anywhere: the
// caster's creatures survive combat, the other players at the table
// survive their own combats, and nothing connects. Druid's Deliverance
// prevents only the damage dealt TO ITS CONTROLLER. Your blockers
// still die, your attackers still trade, and an opponent swinging at
// a THIRD player still gets there — which at a four-player table is
// most of the combat damage on the turn.
//
// Reusing the table-wide shield would therefore have made the card
// substantially STRONGER than printed, which is the one direction a
// simplification may not go (#259). So the shield is player-scoped:
// PreventAllCombatDamageThisTurn with Player set, the same CR 615.1
// replacement with one extra clause requiring the damage's target to
// be this player.
//
// Non-combat damage is untouched: a Lightning Bolt aimed at the
// controller after this resolves still lands, and so does a Pestilence
// activation.
//
// # Populate
//
// The second sentence is the Populate primitive (populate.go): choose
// one of your creature tokens and copy it, nothing at all when you
// control none (CR 701.36b). It runs after the shield is up, which is
// the order the card prints; the shield only concerns damage dealt to
// you, so the order cannot be told apart in play.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fde7645a-5f02-4d5f-b38c-8390f325899e",
		Name:         "Druid's Deliverance",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (PreventAllCombatDamageThisTurn{
				Player: item.Controller,
				Label:  "Druid's Deliverance — prevent combat damage to you",
			}).Apply(ctx); err != nil {
				return err
			}
			return Populate{}.Apply(ctx)
		},
	})
}
