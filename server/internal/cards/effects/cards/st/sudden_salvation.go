package cardsst

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sudden Salvation — Instant {2}{W}{W} (Innistrad: Crimson Vow):
//
//	"Choose up to three target permanent cards in graveyards that were put
//	 there from the battlefield this turn. Return them to the battlefield
//	 tapped under their owners' control. You draw a card for each opponent
//	 who controls one or more of those permanents."
//
// The targets are Brought Back's predicate over every graveyard. All of
// them enter as one simultaneous entry under their owners' control (CR
// 603.6a); the tail then counts the opponents of the caster who control at
// least one of the cards that actually entered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5b0f351-5536-41ec-aa09-1e5344a0effb",
		Name:         "Sudden Salvation",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("up to three target permanent cards in graveyards that were put there from the battlefield this turn",
			effects.Permanent(), effects.PutThereFromTheBattlefieldThisTurn).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFromGraveyardTogether{
				Targets:  effects.LegalTargetCardIDs(ctx),
				Tapped:   true,
				ThenTail: suddenSalvationDrawTail,
			}.Apply(ctx)
		},
	})
}

var suddenSalvationDrawTail = effects.TogetherTail("sudden-salvation/draw",
	func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
		in := make(map[uuid.UUID]bool, len(entered))
		for _, id := range entered {
			in[id] = true
		}
		controllers := map[uuid.UUID]bool{}
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if in[c.InstanceID] {
				controllers[c.Controller] = true
			}
		}
		n := 0
		for _, opp := range ctx.Opponents() {
			if controllers[opp] {
				n++
			}
		}
		return effects.DrawCards{N: n}.Apply(ctx)
	})
