package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// less than {R}{R}. The tokens are encore's shape (encore.go): ONE
// creation, one group of two per opponent tagged with that opponent
// (#4867, game.TokenGroup.Tag), so all of them are created at the same time
// and enter together (CR 608.2f, CR 603.6a), and each pair carries the
// requirement to attack THAT opponent (game.AttackPlayerIfAbleMod, until
// end of turn) and haste until end of turn. The creation goes through the
// keyed creation door, so a creation that pauses on a CR 616 ordering
// question is a restore point and the requirement and haste land on
// whatever it made. The
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
			var groups []game.TokenGroup
			for _, opp := range ctx.Opponents() {
				groups = append(groups, game.TokenGroup{
					Template: ctx.TextToken(TokenCard("3/3 blue and red Elemental with flying")), Count: 2, Tag: opp,
				})
			}
			if len(groups) == 0 {
				return nil
			}
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: item.Controller,
				Source:     item.SourceCardID,
				Groups:     groups,
			}, ctx.ItemTail(furygaleFlockingBindTaggedTokens))
		},
	})
}

// The "furygale-flocking/attack-that-opponent-and-haste" tail is "that
// attack that opponent this turn if able. They gain haste until end of
// turn." over the tokens one creation made; the opponent is the tail's
// Carry. An on-disk key: never renamed, never reused. Since #4867 the
// spell makes one creation for every opponent
// (furygaleFlockingBindTaggedTokens) and no longer schedules this tail; it
// stays registered (by CardTail at init, hence the blank identifier) for a
// restore point written while a creation per opponent was paused.
var _ = CardTail("furygale-flocking/attack-that-opponent-and-haste", func(ctx *Context, o game.TailOutcome) error {
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

// furygaleFlockingBindTaggedTokens is the attack-that-opponent tail over
// the one creation for every opponent (#4867): each token attacks the opponent
// its group was tagged with (game.TokenTags). A token a replacement added
// (an untagged group) is not one of the Elementals the text binds. An
// on-disk key: never renamed, never reused.
var furygaleFlockingBindTaggedTokens = CardTail("furygale-flocking/bind-tagged-tokens", func(ctx *Context, o game.TailOutcome) error {
	tags := game.TokenTags(o, 0)
	for _, tok := range o.IDs {
		opp, ok := tags[tok]
		if !ok {
			continue
		}
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
