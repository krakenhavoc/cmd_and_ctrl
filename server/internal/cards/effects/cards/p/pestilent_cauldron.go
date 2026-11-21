package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pestilent Cauldron // Restorative Burst — a modal double-faced card,
// oracle 78780f37-f215-4ae1-adce-6255b0146873.
//
// Front, Artifact {2}{B}:
//
//	"{T}, Discard a card: Create a 1/1 black and green Pest creature
//	 token with "When this token dies, you gain 1 life."
//	 {1}, {T}: Each opponent mills cards equal to the amount of life you
//	 gained this turn.
//	 {4}, {T}: Exile four target cards from a single graveyard. Draw a
//	 card."
//
// Back, Sorcery {3}{G}{G} (key "<oracle>#1"):
//
//	"Return up to two target creature, land, and/or planeswalker cards
//	 from your graveyard to your hand. Each player gains 4 life. Exile
//	 Restorative Burst."
//
// #1807, ADR 0106 §5. The third ability is the family's one EXACT
// count: four targets, all from one graveyard. The engine offers it
// only when some one graveyard holds four cards (the 2021-04-16
// ruling: "you must be able to target four cards in a single
// graveyard"), because the activation check counts the largest
// graveyard rather than all of them.
//
// The second ability counts all the life gained this turn and ignores
// any lost (the 2021-04-16 ruling); that is TurnTally.LifeGained.
//
// Restorative Burst with no targets still gives everyone 4 life and
// exiles itself; with targets that are all gone on resolution it does
// not resolve and goes to the graveyard (the 2021-04-16 ruling, the
// engine's fizzle).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     PestilentCauldronOracleID,
		Name:         "Pestilent Cauldron",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   `{T}, Discard a card: Create a 1/1 black and green Pest creature token with "When this token dies, you gain 1 life."`,
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.PestToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{1}, {T}: Each opponent mills cards equal to the amount of life you gained this turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					n := effects.B15LifeGainedThisTurn(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return effects.MillEachPlayerInTurn(ctx, ctx.Opponents(), effects.ConstantMill(n), game.KeyedTail{})
				},
			},
			{
				Label:   "{4}, {T}: Exile four target cards from a single graveyard. Draw a card.",
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
				Targets: effects.TargetCardInGraveyard("four target cards from a single graveyard").WithCount(4, 4).AllShare(effects.FromASingleGraveyard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTargetCardsThen(effects.NewContext(g, item), pestilentCauldronDraw)
				},
			},
		},
	})
	effects.Register(effects.Spec{
		OracleID:     PestilentCauldronOracleID + "#1",
		Name:         "Restorative Burst",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("up to two target creature, land, and/or planeswalker cards from your graveyard",
			effects.YouOwn(), effects.Or(effects.Creature(), effects.Land(), effects.Planeswalker())).WithCount(0, 2),
		// The life and the exile wait for the prompts before them (a return's replacement
		// offer, a CR 616 order on the life; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.Instruction(func(ctx *effects.Context) error {
				for _, player := range effects.ApnapPlayers(ctx.Game) {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, 4); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}

const PestilentCauldronOracleID = "78780f37-f215-4ae1-adce-6255b0146873"

// pestilentCauldronDraw is "draw a card" after the exile (#4253 slice 9).
var pestilentCauldronDraw = effects.GraveyardExileTail("pestilent-cauldron/draw-a-card", func(ctx *effects.Context, _ []uuid.UUID, _ map[uuid.UUID]bool) error {
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
