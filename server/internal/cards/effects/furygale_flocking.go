package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Furygale Flocking — Sorcery {8}{R}{R} (Secrets of Strixhaven
// Commander, #4834):
//
//	"This spell costs {1} less to cast for each instant and sorcery card
//	 in your graveyard.
//	 For each opponent, create two 3/3 blue and red Elemental creature
//	 tokens with flying that attack that opponent this turn if able. They
//	 gain haste until end of turn."
//
// The reduction is a self cost modifier over the caster's graveyard
// (Spec.SelfCostModifiers), generic mana only, so the spell never costs
// less than {R}{R}. The tokens are encore's shape (encore.go): one
// creation per opponent, each pair carrying the requirement to attack
// THAT opponent (game.AttackPlayerIfAbleMod, until end of turn) and haste
// until end of turn. Each creation goes through the keyed creation door,
// so a creation that pauses on a CR 616 ordering question is a restore
// point and the requirement and haste land on whatever it made. The
// requirement names a player, so a token is never made to attack a
// planeswalker or battle; one that can't attack that player is free to
// do anything else, as the rulings say.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "943fc88d-c281-46bb-9dc7-1052f789362b",
		Name:         "Furygale Flocking",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 6},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(CountCardsInYourGraveyard(Or(Instant(), Sorcery())),
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, opp := range ctx.Opponents() {
				tail := ctx.ItemTail(furygaleFlockingBindTokens)
				tail.Carry = []uuid.UUID{opp}
				if err := ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
					Controller: item.Controller,
					Source:     item.SourceCardID,
					Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("3/3 blue and red Elemental with flying")), Count: 2}},
				}, tail); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

// furygaleFlockingBindTokens is "that attack that opponent this turn if
// able. They gain haste until end of turn." over the tokens one creation
// made; the opponent is the tail's Carry. An on-disk key: never renamed,
// never reused.
var furygaleFlockingBindTokens = CardTail("furygale-flocking/attack-that-opponent-and-haste", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	opp := o.Tail.Carry[0]
	for _, tok := range o.IDs {
		if err := (ScopedEffectFor{
			Target:   tok,
			Mods:     []game.Mod{game.AttackPlayerIfAbleMod(opp), game.AddKeywordsMod("haste")},
			Duration: DurationUntilEndOfTurn(ctx),
			Label:    "Furygale Flocking — attacks that opponent this turn if able and has haste",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})
