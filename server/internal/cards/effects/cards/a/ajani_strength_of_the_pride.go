package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ajani, Strength of the Pride — Legendary Planeswalker {2}{W}{W}, loyalty
// 5 (Jumpstart 2022, #5948):
//
//	"+1: You gain life equal to the number of creatures you control plus
//	 the number of planeswalkers you control.
//	 −2: Create a 2/2 white Cat Soldier creature token named Ajani's
//	 Pridemate with "Whenever you gain life, put a +1/+1 counter on this
//	 token."
//	 0: If you have at least 15 life more than your starting life total,
//	 exile Ajani and each artifact and creature your opponents control."
//
// Both counts are taken as the +1 resolves, Ajani himself among the
// planeswalkers. The 0 checks the life total as it resolves against the
// table's starting life (ADR 0075); when it holds, Ajani and every
// artifact and creature the opponents control are exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c3974ec7-7af0-4b65-baf6-4c1c4fe8a5c4",
		Name:            "Ajani, Strength of the Pride",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: You gain life equal to the number of creatures you control plus the number of planeswalkers you control.",
				Cost:  effects.LoyaltyCost(1),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					n := len(effects.MatchingBattlefield(ctx, effects.And(effects.YouControl(), effects.Or(effects.Creature(), effects.Planeswalker()))))
					return effects.GainLife{Amount: n}.Apply(ctx)
				})),
			},
			{
				Label:   "−2: Create a 2/2 white Cat Soldier creature token named Ajani's Pridemate with \"Whenever you gain life, put a +1/+1 counter on this token.\"",
				Cost:    effects.LoyaltyCost(-2),
				Purpose: game.Purpose{Tokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.AjanisPridemateToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "0: If you have at least 15 life more than your starting life total, exile Ajani and each artifact and creature your opponents control.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile, OpponentsOnly: true, Partial: true}},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.YouHaveAtLeastLifeAboveStarting(g, ctx.Controller(), 15) {
						return nil
					}
					if err := (effects.ExileAllMatching{Match: effects.And(effects.OpponentControls(), effects.Or(effects.Artifact(), effects.Creature()))}).Apply(ctx); err != nil {
						return err
					}
					if !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.ExileTarget{Target: item.SourceCardID}.Apply(ctx)
				},
			},
		},
	})
}
