package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, Death Mage — Legendary Planeswalker — Liliana {4}{B}{B}, loyalty 4
// (Core Set 2021, #5511):
//
//	"+1: Return up to one target creature card from your graveyard to your
//	 hand.
//	 −3: Destroy target creature. Its controller loses 2 life.
//	 −7: Target opponent loses 2 life for each creature card in their
//	 graveyard."
//
// The −3 is two sentences: the life loss is not gated on the creature dying
// (an indestructible one still costs its controller 2), and it is read off the
// creature's controller as it was, from last-known information once it has
// left (CR 608.2h). A target that is gone makes the whole ability do nothing
// (CR 608.2b). The −7 counts the creature cards in the opponent's graveyard as
// it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d654dd73-763d-4dfb-a30f-a058fd7f27a6",
		Name:            "Liliana, Death Mage",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Return up to one target creature card from your graveyard to your hand.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCardInGraveyard("up to one target creature card from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 1),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					for _, id := range effects.LegalTargetCardIDs(ctx) {
						return effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
					}
					return nil
				})),
			},
			{
				Label:   "−3: Destroy target creature. Its controller loses 2 life.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: effects.Do(
					effects.Instruction(func(ctx *effects.Context) error { return effects.S82MidDestroyTheTarget(nil, ctx) }),
					effects.Instruction(lilianaDeathMageLoseTwo)),
			},
			{
				Label:   "−7: Target opponent loses 2 life for each creature card in their graveyard.",
				Cost:    effects.LoyaltyCost(-7),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					p := ctx.Game.PlayerByIDForEffect(t.ID)
					if p == nil {
						return nil
					}
					n := 0
					for _, c := range p.Graveyard.Cards {
						if c.IsCreature() {
							n++
						}
					}
					return effects.GainLife{Player: t.ID, Amount: -2 * n}.Apply(ctx)
				})),
			},
		},
	})
}

// lilianaDeathMageLoseTwo is "its controller loses 2 life": the controller of
// the targeted creature, live while it is still on the battlefield and
// last-known once it has left. The ability did nothing if the target was not
// legal as it resolved, which is read off the item's own target.
func lilianaDeathMageLoseTwo(ctx *effects.Context) error {
	ts := ctx.Targets()
	if len(ts) == 0 || ts[0].Kind != game.TargetCard {
		return nil
	}
	id := ts[0].ID
	var controller uuid.UUID
	if effects.OnBattlefield(ctx.Game, id) {
		if !ctx.IsTargetLegal(ts[0]) {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok {
			return nil
		}
		controller = c.Controller
	} else {
		info, ok := ctx.Game.LastKnownPermanentForEffect(id)
		if !ok {
			return nil
		}
		controller = info.Controller
	}
	return effects.GainLife{Player: controller, Amount: -2}.Apply(ctx)
}
