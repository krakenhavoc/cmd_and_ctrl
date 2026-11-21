package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nahiri, Forged in Fury — Legendary Creature — Kor Artificer {4}{R}{W},
// 5/4 (March of the Machine: The Aftermath):
//
//	"Affinity for Equipment
//	 Whenever an equipped creature you control attacks, exile the top
//	 card of your library. You may play that card this turn. You may cast
//	 Equipment spells this way without paying their mana costs."
//
// Affinity is the self cost modifier Amanda Grayson reads, counting the
// Equipment its caster controls. The trigger is Swordsman, Sharp
// Scoundrel's: one per equipped attacker, "equipped" judged as the attack
// is declared. The exile is the impulse exile (ExileTopWithPermission's
// plain permission: play, lands included, until end of turn) and the
// exiled card also gets a SECOND permission over the same card: a free
// (cost {0}, CR 118.9) cast narrowed to Equipment spells, claimed as the
// granted-cost alternative. A land or a non-Equipment spell is played at
// its printed cost under the first; an Equipment spell may be cast under
// either, so the caster chooses to pay or not to pay, as the text says
// ("may"). The second permission is a continuation of the exile, so it is
// stamped over the card only once it has landed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87966759-206a-4a2d-bf78-86f727c29972",
		Name:         "Nahiri, Forged in Fury",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Equipment", effects.HasSubtype("Equipment")),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.EquippedCreatureYouControlAttacks, "Nahiri, Forged in Fury — exile the top card of your library; you may play it this turn, and cast Equipment spells this way without paying their mana costs",
				nahiriForgedInFuryExile),
		},
	})
}

func nahiriForgedInFuryExile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	you := ctx.Controller()
	perm := effects.ExileTopWithPermission{From: you, GrantTo: you, N: 1}.Permission()
	return g.ExileTopWithPermissionKeyedForEffect(you, you, 1, perm,
		game.KeyedTail{Ref: nahiriFreeEquipment, Player: you, Source: ctx.Source()})
}

// nahiriFreeEquipment stamps the free Equipment cast over the exiled card
// (an on-disk key; never renamed, never reused).
var nahiriFreeEquipment = effects.CardTail("nahiri-forged-in-fury/free-equipment", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	var cards []game.Card
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Cost:       "{0}",
		AltCostKey: game.AltCostKeyGrantedCost,
		CastOnly:   true,
		Filter:     game.PermissionFilter{Subtype: "Equipment"},
		Source:     ctx.Source(),
		SourceName: "Nahiri, Forged in Fury",
		Label:      "Cast an Equipment spell exiled with Nahiri, Forged in Fury without paying its mana cost",
	}, cards)
	return nil
})
