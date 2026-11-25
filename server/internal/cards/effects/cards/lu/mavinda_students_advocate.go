package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mavinda, Students' Advocate — Legendary Creature — Bird Advisor {2}{W},
// 2/3 (Strixhaven, #6157):
//
//	"Flying
//	 {0}: You may cast target instant or sorcery card from your graveyard
//	 this turn. If that spell doesn't target a creature you control, it
//	 costs {8} more to cast this way. If that spell would be put into your
//	 graveyard, exile it instead. Activate only once each turn."
//
// Vohar, Vodalian Desecrator's grant (a permission over the target card for
// the rest of the turn, at its mana cost and normal timing, the spell exiled
// instead of going to the graveyard) with the surcharge that reads the cast
// spell's targets (CastPermission.CostMoreUnlessTargetsYourCreature, CR
// 601.2c before 601.2f): {8} more unless one of its targets is a creature
// you control as the spell is cast. A target that left the graveyard in
// response is granted nothing. The once-each-turn limit counts activations
// (CR 602.1b).
//
// No simplification.
func init() {
	const name = "Mavinda, Students' Advocate"
	const label = "{0}: You may cast target instant or sorcery card from your graveyard this turn. If that spell doesn't target a creature you control, it costs {8} more to cast this way. If that spell would be put into your graveyard, exile it instead. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "1c248187-0be3-4a03-a817-14f68f638ae7",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Cost:      effects.ManaCost("{0}"),
			Condition: effects.OncePerTurnActivation(label),
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
					Player:                            item.Controller,
					Zone:                              game.ZoneGraveyard,
					Duration:                          g.UntilEndOfTurnDuration(),
					AltCostKey:                        game.AltCostKeyGrantedCost,
					CastOnly:                          true,
					CastsLeft:                         1,
					ExileInsteadOfGraveyard:           true,
					CostMoreUnlessTargetsYourCreature: 8,
					Source:                            item.SourceCardID,
					SourceName:                        name,
					Label:                             name + " — cast " + c.Name + " from your graveyard",
				}, []game.Card{c})
				return nil
			},
		}},
	})
}
