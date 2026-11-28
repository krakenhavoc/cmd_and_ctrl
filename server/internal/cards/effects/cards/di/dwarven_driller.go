package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dwarven Driller — Creature — Dwarf {3}{R}, 2/2 (Judgment):
//
//	"{T}: Destroy target land unless its controller has this creature deal 2
//	 damage to them."
//
// The land's controller is asked as the ability resolves (Thunder Totem's
// "unless" shape, answered by the one it bears on): a yes has the Driller deal
// 2 damage to them, from its last-known information if it has gone (CR
// 608.2h), and the land stays; a no destroys the land. A land that has left
// or stopped being a legal target does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5967c3d3-2ef1-4496-b51c-23478fa77f09",
		Name:         "Dwarven Driller",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Lands: 1},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target land unless its controller has this creature deal 2 damage to them.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Lands: 1},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPermanent("target land", effects.Land()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				land, ok := g.LookupCardForEffect(id)
				if !ok || land.Controller == uuid.Nil {
					return nil
				}
				return effects.MayChoice{
					Player:   land.Controller,
					Question: "Dwarven Driller — have it deal 2 damage to you? (If not, your land is destroyed.)",
					YesLabel: "Take 2 damage",
					NoLabel:  "Destroy the land",
					Then:     dwarvenDrillerAnswer,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			},
		}},
	})
}

// dwarvenDrillerAnswer is the land's controller's answer: 2 damage from the
// Driller on a yes, the land destroyed on a no. An on-disk key: never
// renamed, never reused.
var dwarvenDrillerAnswer = effects.MayYesNo("confirm/s82-old-b627-dwarven-driller",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		return effects.DealDamage{Source: ctx.Source(), Target: r.Chooser, Amount: 2}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
			return nil
		}
		return effects.DestroyTarget{Target: r.Carry[0]}.Apply(ctx)
	})
