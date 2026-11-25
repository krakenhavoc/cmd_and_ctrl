package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dovin, Architect of Law — Legendary Planeswalker — Dovin {4}{W}{U} (Ravnica
// Allegiance), loyalty 5:
//
//	"+1: You gain 2 life and draw a card.
//	 −1: Tap target creature. It doesn't untap during its controller's next untap
//	 step.
//	 −9: Tap all permanents target opponent controls. That player skips their
//	 next untap step."
//
// The −1 is Rooftop Nuisance's tap-and-freeze. The −9 taps every untapped
// permanent the chosen opponent controls, then records that player's skipped
// untap step (Game.SkipNextStepForEffect, CR 614.10): the step is skipped
// whole, so even permanents they would untap for another reason stay tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a3386efe-d6c5-4853-a4db-61460da3364a",
		Name:            "Dovin, Architect of Law",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: You gain 2 life and draw a card.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1, LifeGain: 2},
				Effect:  effects.Do(effects.GainLife{Amount: 2}, effects.DrawCards{N: 1}),
			},
			{
				Label:   "−1: Tap target creature. It doesn't untap during its controller's next untap step.",
				Cost:    effects.LoyaltyCost(-1),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Dovin, Architect of Law — it doesn't untap during its controller's next untap step"}.Apply(ctx)
				},
			},
			{
				Label:   "−9: Tap all permanents target opponent controls. That player skips their next untap step.",
				Cost:    effects.LoyaltyCost(-9),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect:  dovinArchitectOfLawUltimate,
			},
		},
	})
}

func dovinArchitectOfLawUltimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := effects.TargetedPlayer(ctx)
	if player == uuid.Nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && !c.Tapped {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (effects.TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	g.SkipNextStepForEffect(player, game.StepUntap, item.SourceCardID, "Dovin, Architect of Law — that player skips their next untap step")
	return nil
}
