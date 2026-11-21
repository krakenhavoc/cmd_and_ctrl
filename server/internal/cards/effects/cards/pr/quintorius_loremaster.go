package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quintorius, Loremaster — Legendary Creature — Elephant Cleric
// {3}{R}{W}, 3/5 (Secrets of Strixhaven Commander, #4834):
//
//	"Vigilance
//	 At the beginning of your end step, exile target noncreature,
//	 nonland card from your graveyard. Create a 3/2 red and white Spirit
//	 creature token.
//	 {1}{R}{W}, {T}, Sacrifice a Spirit: Choose target card exiled with
//	 Quintorius. You may cast that card this turn without paying its
//	 mana cost. If that spell would be put into a graveyard, put it on
//	 the bottom of its owner's library instead."
//
// The end-step trigger targets; with no legal target it is never put on the
// stack and makes no Spirit, and a target that is illegal on resolution
// makes the whole ability do nothing (the 2023-04-14 rulings). The exile is
// linked to Quintorius (CR 607.2a), so the activated ability, which targets
// "a card exiled with Quintorius", sees only those cards, and a Quintorius
// that left and came back is a new object with none. The token follows the
// exile as its own clause, whether or not the card reached exile.
//
// The cast is a per-card permission for the rest of the turn at the card's
// own timing (the ruling: a sorcery is cast only in your main phase with the
// stack empty), priced at {0} as an alternative cost (so no other
// alternative cost may be chosen and mandatory additional costs are still
// paid). The last clause is the engine's BottomInsteadOfGraveyard: it
// replaces a move of the spell to its owner's graveyard however it gets
// there (resolving, countered or fizzled), and a spell bounced or put into a
// library goes there instead. Vigilance is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c3d38129-b955-4fb5-8486-095f918935b9",
		Name:            "Quintorius, Loremaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourEndStep("Quintorius, Loremaster — exile target noncreature, nonland card from your graveyard, then create a 3/2 red and white Spirit creature token",
					quintoriusLoremasterExile),
				effects.TargetCardInGraveyard("target noncreature, nonland card from your graveyard",
					effects.YouOwn(), effects.Noncreature(), effects.Nonland())),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{R}{W}, {T}, Sacrifice a Spirit: Choose target card exiled with Quintorius. You may cast that card this turn without paying its mana cost. If that spell would be put into a graveyard, put it on the bottom of its owner's library instead.",
			Cost: effects.Plus(effects.ManaCost("{1}{R}{W}"), effects.TapCost(),
				effects.SacrificeN(1, "a Spirit", effects.HasSubtype("Spirit"))),
			Targets: effects.ExiledWithThis(effects.TargetCardInExile("target card exiled with Quintorius, Loremaster")),
			Effect:  quintoriusLoremasterCast,
		}},
	})
}

// quintoriusLoremasterExile is the end-step trigger: the target is exiled
// linked to Quintorius, then the Spirit is made as the next clause.
func quintoriusLoremasterExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.ExileTarget{Target: t.ID, Linked: true, ThenTail: ctx.ItemTail(quintoriusLoremasterSpirit)}.Apply(ctx)
	}
	return nil
}

// quintoriusLoremasterSpirit is "Create a 3/2 red and white Spirit creature
// token", the clause after the exile, as a keyed tail. An on-disk key.
var quintoriusLoremasterSpirit = effects.CardTail("quintorius-loremaster/spirit", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("3/2 red and white Spirit"), N: 1}.Apply(ctx)
})

// quintoriusLoremasterCast grants the free cast of the still-legal target.
func quintoriusLoremasterCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:                   ctx.Controller(),
			Zone:                     game.ZoneExile,
			Duration:                 g.UntilEndOfTurnDuration(),
			Cost:                     "{0}",
			AltCostKey:               game.AltCostKeyGrantedCost,
			CastOnly:                 true,
			BottomInsteadOfGraveyard: true,
			Source:                   ctx.Source(),
			SourceName:               "Quintorius, Loremaster",
			Label:                    "Quintorius, Loremaster — cast " + c.Name + " without paying its mana cost",
		}, []game.Card{c})
		return nil
	}
	return nil
}
