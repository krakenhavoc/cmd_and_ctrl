package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farid, Enterprising Salvager — Legendary Creature — Human Soldier {2}{R},
// 3/3 (The Brothers' War Commander, S82):
//
//	"Whenever a nontoken artifact you control is put into a graveyard from
//	 the battlefield, create a colorless artifact token named Scrap.
//	 {1}{R}, Sacrifice an artifact: Choose one —
//	 • Put a +1/+1 counter on Farid. It gains menace until end of turn.
//	 • Goad target creature.
//	 • Discard a card, then draw a card."
//
// The trigger reads the artifact as it last existed and under whom
// (CR 603.10a), so an animated artifact or a Vehicle that was a creature
// still counts and a token does not. The sacrifice itself is a cost, so a
// nontoken artifact sacrificed to the ability triggers Farid. A modal
// activated ability is one ability per bullet with the same cost
// (Goblin Cratermaker's shape): the mode is chosen at activation.
//
// No simplification.
func init() {
	cost := effects.Plus(effects.ManaCost("{1}{R}"), effects.SacrificeN(1, "an artifact", effects.Artifact()))
	effects.Register(effects.Spec{
		OracleID:     "83cb9ab7-6a8f-40b5-b076-6fa669ff15f2",
		Name:         "Farid, Enterprising Salvager",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, controller, ok := effects.S82BroB08ArtifactLeft(ev, g)
				return ok && controller == source.Controller && !c.IsToken()
			}, "Farid — create a Scrap token", effects.Do(effects.CreateToken{Template: effects.TokenCard("Scrap artifact"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{R}, Sacrifice an artifact: Put a +1/+1 counter on Farid. It gains menace until end of turn.",
				Cost:    cost,
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					if !effects.OnBattlefield(ctx.Game, ctx.Source()) {
						return nil
					}
					if err := (effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"menace"}, Label: "Farid — menace"}.Apply(ctx)
				})),
			},
			{
				Label:   "{1}{R}, Sacrifice an artifact: Goad target creature.",
				Cost:    cost,
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.GoadTarget,
			},
			{
				Label:   "{1}{R}, Sacrifice an artifact: Discard a card, then draw a card.",
				Cost:    cost,
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Effect:  effects.S82BroB08Rummage,
			},
		},
	})
}
