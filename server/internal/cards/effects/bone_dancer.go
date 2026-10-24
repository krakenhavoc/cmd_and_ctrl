package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bone Dancer — Creature — Zombie {1}{B}{B}, 2/2:
//
//	"Whenever this creature attacks and isn't blocked, you may put the top
//	 creature card of defending player's graveyard onto the battlefield
//	 under your control. If you do, this creature assigns no combat damage
//	 this turn."
//
// #3531: no target; the card is found as the ability resolves, so a card
// milled in response is the one that comes back. The graveyard keeps its
// order (CR 404.2; the top is the last card), and the topmost CREATURE card
// is taken, which need not be the top card. A defending player with no
// creature card in the graveyard gives nothing to do, so nothing is asked
// and the Dancer's damage is not stopped.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "13b5c9ef-d0e3-453d-ae5e-8de4f310cc38",
		Name:         "Bone Dancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenAttacksAndIsNotBlockedEffect("Bone Dancer — you may put the top creature card of defending player's graveyard onto the battlefield under your control",
				func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
					ctx := NewContext(g, item)
					if topCreatureCardOfGraveyard(g, defender) == uuid.Nil {
						return nil
					}
					return MayChoice{
						Question: "Bone Dancer — put the top creature card of defending player's graveyard onto the battlefield under your control? (it then assigns no combat damage this turn)",
						Then:     boneDancerReanimate,
						Carry:    []uuid.UUID{defender},
					}.Apply(ctx)
				}),
		},
	})
}

// topCreatureCardOfGraveyard is the topmost creature card of a player's
// graveyard, or uuid.Nil.
func topCreatureCardOfGraveyard(g *game.Game, player uuid.UUID) uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return uuid.Nil
	}
	for i := len(p.Graveyard.Cards) - 1; i >= 0; i-- {
		if p.Graveyard.Cards[i].IsCreature() {
			return p.Graveyard.Cards[i].InstanceID
		}
	}
	return uuid.Nil
}

// boneDancerReanimate is the "yes": the top creature card of the defending
// player's graveyard comes onto the battlefield, and Bone Dancer assigns no
// combat damage this turn. Keyed (#4291).
var boneDancerReanimate = MayYes("confirm/bone-dancer-reanimate", func(ctx *Context, r game.ConfirmAnswer) error {
	card := topCreatureCardOfGraveyard(ctx.Game, carryAt(r, 0))
	if card == uuid.Nil {
		return nil
	}
	if err := (ReturnFromGraveyard{Target: card, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
		return err
	}
	assignsNoCombatDamageThisTurn(ctx, "Bone Dancer — assigns no combat damage this turn")
	return nil
})
