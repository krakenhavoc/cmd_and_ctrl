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
// the same CR 615.1 turn-scoped replacement, with one extra clause
// requiring the damage's target to be this player.
//
// Non-combat damage is untouched: a Lightning Bolt aimed at the
// controller after this resolves still lands, and so does a Pestilence
// activation.
//
// # Declared simplification (weaker than printed): no populate
//
// Populate (CR 701.32) creates a token that is a copy of a creature
// token its controller chooses. The engine can copy a permanent into
// a token, but there is no populate keyword and no prompt that offers
// "choose a creature token you control" as the thing being copied, so
// the second sentence is not modelled. The card ships as the
// protection half only, for two mana.
//
// That takes something away and adds nothing.
func init() {
	Register(Spec{
		OracleID:     "fde7645a-5f02-4d5f-b38c-8390f325899e",
		Name:         "Druid's Deliverance",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Populate isn't available, so no token copy is made — the spell is the damage prevention only.",
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b38PreventAllCombatDamageToPlayerThisTurn(ctx, item.Controller,
				"Druid's Deliverance — prevent combat damage to you")
		},
	})
}
