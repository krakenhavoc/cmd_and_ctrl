package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xander's Pact — Sorcery {4}{B}{B} (Streets of New Capenna Commander):
//
//	"Casualty 2 (As you cast this spell, you may sacrifice a creature with
//	 power 2 or greater. When you do, copy this spell.)
//	 Each opponent exiles the top card of their library. You may cast spells
//	 from among those cards this turn. If you cast a spell this way, pay life
//	 equal to that spell's mana value rather than pay its mana cost."
//
// Casualty is Ob Nixilis's optional cost (CR 702.153); the engine makes the
// copy. Each opponent, in turn order from the controller's left (CR 101.4),
// exiles their own top card as an impulse exile (Inside Information's: the
// library read and the player granted to are separate), and the permission
// the controller gets is cast-only (a land among them is not played), lasts
// for the turn, and carries Bolas's Citadel's price (LifeEqualToManaValue,
// CR 118.9, 119.4): a player without the life cannot claim it. The copy
// makes the same exiles and its own permission, so a card may be cast under
// either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "3df11e5c-19aa-441c-a889-6b40eb3fdabf",
		Name:          "Xander's Pact",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Casualty(2)},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, opp := range effects.OpponentsInTurnOrder(ctx.Game, item.Controller) {
				if _, err := ctx.Game.ExileTopWithPermissionForEffect(opp, item.Controller, 1, game.CastPermission{
					CastOnly:             true,
					AltCostKey:           XandersPactAltCostKey,
					LifeEqualToManaValue: true,
					Label:                "Pay life equal to its mana value (Xander's Pact)",
				}); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// xandersPactAltCostKey is the offer the cast claims. An on-disk identity:
// never renamed, never reused.
const XandersPactAltCostKey = "xanders_pact"
