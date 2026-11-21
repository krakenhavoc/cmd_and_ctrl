package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goldwardens' Gambit — Sorcery {6}{R}{R} (Phyrexia: All Will Be One,
// Commander):
//
//	"Affinity for Equipment (This spell costs {1} less to cast for each
//	 Equipment you control.)
//	 Create five 2/2 red Rebel creature tokens. They gain haste until end of
//	 turn. For each of those tokens, you may attach an Equipment you control
//	 to it."
//
// #5479, S82. Five Rebels in one creation (Ajani's Chosen's keyed tail
// told the new tokens), each given haste until end of turn, then one
// "you may attach an Equipment you control to it" for each token in turn,
// Beatrix's pick-and-attach shape with a pick of none or one. Each token is
// offered the Equipment that could legally go on it (CR 701.3b) when its
// turn comes, so one that an earlier pick moved is offered again and may be
// moved onto a later token, as the sentence is read one token at a time. A
// token that has left the battlefield by then is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf113674-f809-4b97-9e8d-e22385446ddb",
		Name:         "Goldwardens' Gambit",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Equipment", effects.HasSubtype("Equipment")),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: item.Controller,
				Source:     item.SourceCardID,
				Groups: []game.TokenGroup{{
					Template: ctx.TextToken(effects.TokenCard("2/2 red Rebel")),
					Count:    5,
				}},
			}, ctx.ItemTail(goldwardensGambitAfterTokens))
		},
	})
}

// goldwardensGambitAfterTokens is the clauses after the creation: the
// tokens gain haste, then the attach offers begin. An on-disk key.
var goldwardensGambitAfterTokens = effects.CardTail("goldwardens-gambit/after-tokens", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		if err := (effects.GrantKeywordUntilEOT{
			Target:   id,
			Keywords: []string{"haste"},
			Label:    "Goldwardens' Gambit — haste until end of turn",
		}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
	}
	return goldwardensGambitOffer(ctx.Game, ctx.Controller(), ctx.Source(), o.IDs)
})

// goldwardensGambitOffer asks about the first token of `tokens` still on
// the battlefield that has an Equipment it could take; the rest ride the
// prompt's carry.
func goldwardensGambitOffer(g *game.Game, you, source uuid.UUID, tokens []uuid.UUID) error {
	for i, token := range tokens {
		if !g.Battlefield.Contains(token) {
			continue
		}
		var candidates []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller != you || !c.HasSubtype("Equipment") || c.IsAttachedTo(token) {
				continue
			}
			if g.AttachLegalForEffect(c.InstanceID, token) {
				candidates = append(candidates, c.InstanceID)
			}
		}
		if len(candidates) == 0 {
			continue
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  you,
			Source:   source,
			Question: "Goldwardens' Gambit — you may attach an Equipment you control to this Rebel",
			Cards:    candidates,
			Min:      0,
			Max:      1,
			ThenKey:  goldwardensGambitAttach,
			LiveZone: game.ZoneBattlefield,
			Carry:    append([]uuid.UUID{token}, tokens[i+1:]...),
		})
		return nil
	}
	return nil
}

// goldwardensGambitAttach attaches the pick, if any, to the token at the
// head of the carry, then offers the next token. An on-disk key.
var goldwardensGambitAttach game.ChooseCardsThen

// The registration is in init, not in the var: the body offers the next
// token, which names this key, and a package-level initializer cannot.
func init() {
	goldwardensGambitAttach = game.RegisterChooseCardsThen("choose-cards/goldwardens-gambit-attach", goldwardensGambitAttachBody)
}

func goldwardensGambitAttachBody(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 {
		return nil
	}
	token := r.Carry[0]
	if g.Battlefield.Contains(token) {
		for _, id := range r.Picked {
			if !g.Battlefield.Contains(id) {
				continue
			}
			if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: token}); err != nil {
				return err
			}
		}
	}
	return goldwardensGambitOffer(g, r.Chooser, r.Source, r.Carry[1:])
}
