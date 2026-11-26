package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foul Renewal — Instant {3}{B} (Dragons of Tarkir):
//
//	"Return target creature card from your graveyard to your hand. Target
//	 creature gets -X/-X until end of turn, where X is the toughness of the card
//	 returned this way."
//
// X is the toughness the card has in the graveyard (its printed value, CR 400.7)
// and only counts when the card really reached the hand: with the first target
// gone nothing is returned and the second target gets -0/-0. The two targets are
// separate clauses (CR 115.1); the creature may be any creature, the returned
// card is in the graveyard when it is chosen so it is never that one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6beaf310-c623-4ab6-a0d3-f4c8221f255b",
		Name:         "Foul Renewal",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
			effects.TargetCreature("target creature"),
		),
		Purpose:   effects.ForTargets(effects.HarmUntilEOT(1, 0, 0)),
		OnResolve: foulRenewalResolve,
	})
}

func foulRenewalResolve(_ *game.StackItem, ctx *effects.Context) error {
	x := 0
	if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
		if card, found := ctx.Game.LookupCardForEffect(t.ID); found {
			toughness := card.Toughness
			if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			if foulRenewalInHand(ctx.Game, ctx.Controller(), t.ID) {
				x = max(toughness, 0)
			}
		}
	}
	t, ok := ctx.ClauseTarget(1)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return effects.BoostUntilEOT{Target: t.ID, Power: -x, Toughness: -x, Label: "Foul Renewal — -X/-X until end of turn"}.Apply(ctx)
}

func foulRenewalInHand(g *game.Game, player, card uuid.UUID) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Hand.Contains(card)
}
