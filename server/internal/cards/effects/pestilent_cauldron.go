package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     pestilentCauldronOracleID,
		Name:         "Pestilent Cauldron",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   `{T}, Discard a card: Create a 1/1 black and green Pest creature token with "When this token dies, you gain 1 life."`,
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    Plus(TapCost(), DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: PestToken(), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{1}, {T}: Each opponent mills cards equal to the amount of life you gained this turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					n := b15LifeGainedThisTurn(g, item.Controller)
					if n <= 0 {
						return nil
					}
					for _, opp := range ctx.Opponents() {
						if err := (MillCards{Player: opp, N: n}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "{4}, {T}: Exile four target cards from a single graveyard. Draw a card.",
				Cost:    Plus(ManaCost("{4}"), TapCost()),
				Targets: TargetCardInGraveyard("four target cards from a single graveyard").WithCount(4, 4).AllShare(FromASingleGraveyard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return exileTargetCardsThen(NewContext(g, item), func(ctx *Context, _ []uuid.UUID, _ map[uuid.UUID]bool) error {
						return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
					})
				},
			},
		},
	})
	Register(Spec{
		OracleID:     pestilentCauldronOracleID + "#1",
		Name:         "Restorative Burst",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("up to two target creature, land, and/or planeswalker cards from your graveyard",
			YouOwn(), Or(Creature(), Land(), Planeswalker())).WithCount(0, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := returnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			for _, player := range apnapPlayers(ctx.Game) {
				if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), player, 4); err != nil {
					return err
				}
			}
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}

const pestilentCauldronOracleID = "78780f37-f215-4ae1-adce-6255b0146873"
