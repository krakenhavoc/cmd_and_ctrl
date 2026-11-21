package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nita, Forum Conciliator — Legendary Creature — Human Advisor {1}{W}{B}, 2/3
// (Secrets of Strixhaven, #4456):
//
//	"Whenever you cast a spell you don't own, put a +1/+1 counter on each
//	 creature you control.
//	 {2}, Sacrifice another creature: Exile target instant or sorcery card
//	 from an opponent's graveyard. You may cast it this turn, and mana of any
//	 type can be spent to cast that spell. If that spell would be put into a
//	 graveyard, exile it instead. Activate only as a sorcery."
//
// The cast trigger reads the spell's owner (a spell cast from an opponent's
// graveyard or exile, or one Nita's own ability lets you cast, is "a spell you
// don't own"); the counters go on the creatures you control as the ability
// resolves, a set fixed then (Archangel of Thune's body), Nita included if she
// is still there. The activated ability sacrifices another creature as a
// cost, targets the opponent's card, exiles it, and as a second clause grants
// a cast permission for the rest of the turn at the card's own timing and
// printed cost, payable with mana of any type (CR 609.4b), with the
// engine's "exile instead of the graveyard" riding the permission (claimed
// under the granted-cost key with no price of its own, as Emet-Selch's is, so
// the cast is made at the printed mana cost and the replacement travels with
// the spell). The cast
// follows the exile only if the card actually reached exile.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f589e5ee-399c-4613-b7a2-9ab2866cc830",
		Name:         "Nita, Forum Conciliator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.SocB30SpellYouDontOwn(),
				"Nita, Forum Conciliator — a +1/+1 counter on each creature you control",
				effects.B11PutCounterOnEachCreatureYouControl),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Sacrifice another creature: Exile target instant or sorcery card from an opponent's graveyard. You may cast it this turn, and mana of any type can be spent to cast that spell. If that spell would be put into a graveyard, exile it instead. Activate only as a sorcery.",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Targets: effects.TargetCardInGraveyard("target instant or sorcery card from an opponent's graveyard",
				effects.Or(effects.Instant(), effects.Sorcery()), effects.Not(effects.YouOwn())),
			SorcerySpeed: true,
			Effect:       nitaExileAndOfferCast,
		}},
	})
}

// nitaExileAndOfferCast exiles the still-legal target; the cast permission is
// the clause after the exile.
func nitaExileAndOfferCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(nitaAfterExile)}.Apply(ctx)
	}
	return nil
}

// nitaAfterExile is "You may cast it this turn, and mana of any type can be
// spent to cast that spell. If that spell would be put into a graveyard,
// exile it instead", as a keyed tail. An on-disk key.
var nitaAfterExile = effects.CardTail("nita-forum-conciliator/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:                  ctx.Controller(),
			Zone:                    game.ZoneExile,
			Duration:                ctx.Game.UntilEndOfTurnDuration(),
			AltCostKey:              game.AltCostKeyGrantedCost,
			CastOnly:                true,
			AnyType:                 true,
			ExileInsteadOfGraveyard: true,
			Source:                  ctx.Source(),
			SourceName:              "Nita, Forum Conciliator",
			Label:                   "Nita, Forum Conciliator — cast " + c.Name + ", spending mana of any type",
		}, []game.Card{c})
	}
	return nil
})
