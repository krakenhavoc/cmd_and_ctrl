package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Timeless Voyager — Legendary Planeswalker — Teferi {4}{U}{U},
// loyalty 4 (Core Set 2021, #5511):
//
//	"+1: Draw a card.
//	 −3: Put target creature on top of its owner's library.
//	 −8: Each creature target opponent controls phases out. Until the end of
//	 your next turn, they can't phase in."
//
// The −8 phases out every creature the target opponent controls at once, with
// what is attached to them (CR 702.26g), then bars exactly those creatures
// from phasing in until the end of your next turn
// (Game.CantPhaseInThroughYourNextTurnForEffect): they miss that opponent's
// next untap step and phase in on the ordinary clock afterwards (CR 502.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7240037f-ddea-425e-9824-866e2bb0dc84",
		Name:            "Teferi, Timeless Voyager",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:  "+1: Draw a card.",
				Cost:   effects.LoyaltyCost(1),
				Effect: effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "−3: Put target creature on top of its owner's library.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.PutTargetOnTopOfOwnersLibrary(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "−8: Each creature target opponent controls phases out. Until the end of your next turn, they can't phase in.",
				Cost:    effects.LoyaltyCost(-8),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect:  effects.Do(effects.Instruction(teferiTimelessVoyagerPhaseOut)),
			},
		},
	})
}

// teferiTimelessVoyagerPhaseOut phases out the creatures the target
// opponent controls and bars those creatures from phasing in until the end
// of the activator's next turn.
func teferiTimelessVoyagerPhaseOut(ctx *effects.Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == t.ID {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	if err := (effects.PhaseOut{Targets: ids}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.CantPhaseInThroughYourNextTurnForEffect(ctx.Controller(), ids...)
	return nil
}
