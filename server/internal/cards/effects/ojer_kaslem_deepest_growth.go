package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ojer Kaslem, Deepest Growth — Legendary Creature — God {3}{G}{G},
// 6/5, the front face of a transforming card whose back is Temple of
// Cultivation:
//
//	"Trample
//	 Whenever Ojer Kaslem deals combat damage to a player, reveal that
//	 many cards from the top of your library. You may put a creature
//	 card and/or a land card from among them onto the battlefield. Put
//	 the rest on the bottom of your library in a random order.
//	 When Ojer Kaslem dies, return it to the battlefield tapped and
//	 transformed under its owner's control."
//
// The reveal is a RevealTopOfLibraryForEffect of the damage dealt
// (read off the trigger's event, fixed when the damage was dealt), and
// the pick is PutFromLibraryOntoBattlefield over creature-or-land with
// a Validate rule that the picks can be told apart: at most one
// creature and at most one land. "A creature card and/or a land
// card" is a rule about the SET, which is what Validate is for, and it
// runs in the legal-move enumerator too, so a bot is only offered sets
// the resolver accepts. The dies trigger is
// ReturnFromGraveyard{Transformed: true} (#1900, ADR 0079 amendment of
// 2026-10-08).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        ojerKaslemOracleID,
		Name:            "Ojer Kaslem, Deepest Growth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Source == source.InstanceID && combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Ojer Kaslem — reveal that many cards from the top of your library", ojerKaslemReveal),
			ojerDiesReturnTransformed("Ojer Kaslem", nil),
		},
	})
}

func ojerKaslemReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), item.Trigger.Event.Amount,
		"Ojer Kaslem — revealed from the top of the library")
	return PutFromLibraryOntoBattlefield{
		Player:   item.Controller,
		Cards:    revealed,
		Match:    Or(Creature(), Land()),
		Max:      2,
		Optional: true,
		Validate: ojerKaslemCreatureAndOrLand,
		Label:    "Ojer Kaslem — you may put a creature card and/or a land card onto the battlefield",
		Then:     PutRestOnBottomInRandomOrder,
	}.Apply(ctx)
}

// ojerKaslemCreatureAndOrLand is "a creature card and/or a land card":
// one card of either kind, or two cards that are a creature and a land
// between them. Two creatures, or two lands, are not a legal pick. A
// card that is both (a creature land) can fill either slot.
func ojerKaslemCreatureAndOrLand(picks []game.Card) bool {
	switch len(picks) {
	case 1:
		return picks[0].IsCreature() || picks[0].IsLand()
	case 2:
		a, b := picks[0], picks[1]
		return (a.IsCreature() && b.IsLand()) || (a.IsLand() && b.IsCreature())
	}
	return false
}

// Temple of Cultivation — Land, the back face of Ojer Kaslem:
//
//	"(Transforms from Ojer Kaslem, Deepest Growth.)
//	 {T}: Add {G}.
//	 {2}{G}, {T}: Transform this land. Activate only if you control ten
//	 or more permanents and only as a sorcery."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     ojerKaslemOracleID + "#1",
		Name:         "Temple of Cultivation",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "{T}: Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}{G}, {T}: Transform this land. Activate only if you control ten or more permanents and only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{G}"), TapCost()),
			SorcerySpeed: true,
			Condition:    controlsTenPermanents,
			Effect:       Do(TransformThis{}),
		}},
	})
}

const ojerKaslemOracleID = "eda11077-b2ce-408b-b982-def2da8fe599"

// controlsTenPermanents is "Activate only if you control ten or more
// permanents".
func controlsTenPermanents(g *game.Game, controller, _ uuid.UUID) bool {
	return countControlled(g, controller, func(game.Card) bool { return true }) >= 10
}
