package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vohar, Vodalian Desecrator — Legendary Creature — Phyrexian Merfolk
// Wizard {U}{B}, 1/2 (Dominaria United, #5487):
//
//	"{T}: Draw a card, then discard a card. If you discarded an instant or
//	 sorcery card this way, each opponent loses 1 life and you gain 1 life.
//	 {2}, Sacrifice Vohar: You may cast target instant or sorcery card from
//	 your graveyard this turn. If that spell would be put into your
//	 graveyard, exile it instead. Activate only as a sorcery."
//
// The loot is connive's draw then a chosen discard, whose registered
// continuation reads the card that really left the hand (so a discard
// that is empty-handed pays nothing): an instant or sorcery card drains
// each opponent for 1 and gains 1. The second ability sacrifices Vohar as a
// cost and, as it resolves, grants a permission over the target card for
// the rest of the turn at its mana cost, normal timing (a sorcery still
// waits for a main phase with an empty stack), with the spell exiled
// instead of going to the graveyard (Emet-Selch's ExileInsteadOfGraveyard).
// A target that left the graveyard in response is not granted anything.
//
// No simplification.
func init() {
	const name = "Vohar, Vodalian Desecrator"
	effects.Register(effects.Spec{
		OracleID:     "4380df08-7be4-48ca-9783-d644fd2ea27d",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Draw a card, then discard a card. If you discarded an instant or sorcery card this way, each opponent loses 1 life and you gain 1 life.",
				Cost:    effects.TapCost(),
				Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if err := g.DrawNForEffect(item.Controller, 1); err != nil {
						return err
					}
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						Question: name + " — discard a card",
						Asked:    game.NewAskedItem(item),
					}, voharDiscarded, nil)
				},
			},
			{
				Label:        "{2}, Sacrifice Vohar: You may cast target instant or sorcery card from your graveyard this turn. If that spell would be put into your graveyard, exile it instead. Activate only as a sorcery.",
				Cost:         effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				SorcerySpeed: true,
				Targets: effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					z := g.FindCardZoneForEffect(id)
					c, found := g.LookupCardForEffect(id)
					if !found || z == nil || z.Kind != game.ZoneGraveyard || c.Owner != item.Controller {
						return nil
					}
					g.GrantCastPermissionToCardsForEffect(game.CastPermission{
						Player:                  item.Controller,
						Zone:                    game.ZoneGraveyard,
						Duration:                g.UntilEndOfTurnDuration(),
						AltCostKey:              game.AltCostKeyGrantedCost,
						CastOnly:                true,
						CastsLeft:               1,
						ExileInsteadOfGraveyard: true,
						Source:                  item.SourceCardID,
						SourceName:              name,
						Label:                   name + " — cast " + c.Name + " from your graveyard",
					}, []game.Card{c})
					return nil
				},
			},
		},
	})
}

// voharDiscarded is the loot's "if you discarded an instant or sorcery card
// this way": each opponent loses 1 life and you gain 1 life. An on-disk
// key: never renamed, never reused.
var voharDiscarded = game.RegisterDiscardThen("choose-cards/vohar-vodalian-desecrator-discarded",
	func(g *game.Game, r game.DiscardsDone) error {
		if !voharDiscardedAnInstantOrSorcery(g, r.Player, r.Discarded.By(r.Player)) || r.Item == nil {
			return nil
		}
		ctx := effects.NewContext(g, r.Item)
		if err := (effects.EachOpponentLosesLife{N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.GainLife{Amount: 1}.Apply(ctx)
	})

func voharDiscardedAnInstantOrSorcery(g *game.Game, _ uuid.UUID, ids []uuid.UUID) bool {
	for _, id := range ids {
		if c, ok := g.LookupCardForEffect(id); ok && (c.IsInstant() || c.IsSorcery()) {
			return true
		}
	}
	return false
}
