package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Uldaros Theorix — Legendary Creature — Elder Sphinx {3}{U}{B}{B},
// 5/5:
//
//	"Flying
//	 When Uldaros Theorix enters, if you cast him, exile up to one
//	 target nonland card of each card type from your graveyard. Copy
//	 those cards. You may cast any number of spells with total mana
//	 value 6 or less from among the copies without paying their mana
//	 costs. (Permanent spells cast this way become tokens.)"
//
// "If you cast him" is an intervening if (CR 603.4) read off the
// permanent's cast provenance (CR 400.7d), as Geological Appraiser's is:
// a reanimated or flickered Theorix triggers nothing.
//
// # One target clause per card type
//
// "Up to one target nonland card of each card type" is eight target
// clauses in one statement (#764's multi-clause shape, as Betor's two
// are): artifact, battle, creature, enchantment, instant, kindred,
// planeswalker and sorcery, each "up to one" and each Distinct from the
// earlier ones. Distinct is what makes a card with two types fill ONE
// slot: an artifact creature card picked for the artifact slot is not
// offered again for the creature slot (CR 115.3 — one object can't be
// chosen twice for one "target"), and which slot it fills is the
// player's choice by the order they answer. A clause with nothing legal
// is skipped by the walk and one the player leaves empty is a legal
// answer, so a graveyard of two spells costs two answers. Land is not a
// slot: the printed text says nonland. The pick is the engine's
// ordinary trigger target prompt, so the client and the bot need nothing
// new; only the number of prompts is new (one per slot that has a
// candidate, at most eight).
//
// # The budget
//
// The exile goes through the CR 614 window and only the cards that
// really reached exile are copied. Every copy is then offered through
// ONE permission carrying a running total of 6
// (game.CastPermission.ManaValueBudget, #2017; ADR 0066's 2026-10-14
// amendment): each cast spends its spell's mana value, X at 0 as the
// ruling of 2026-08-21 says a free cast has it, and a copy whose mana
// value is more than what is left is refused by the one gate the cast
// path, the client's offers and the bot share. A mana value 2 and a mana
// value 4 spell fit, a 3, a 3 and a 1 do not. The order is the
// caster's: any set with a total of 6 or less can be cast in some order
// without ever going over, so the running total and the printed
// "total mana value 6 or less" accept the same casts.
//
// As with every "you may cast it" a resolution offers (ADR 0066, ADR
// 0091), the casts are grants taken with ordinary cast_spell actions
// after the trigger resolves, and the window closes when the caster next
// passes priority; copies not cast cease to exist (CR 704.5e). A
// permanent copy becomes a token as it resolves (CR 707.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "020d561b-2e5d-439e-9e6a-318d23e73542",
		Name:            "Uldaros Theorix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, uldarosTheorixWasCast),
					"Uldaros Theorix — exile up to one nonland card of each type from your graveyard, copy them, and cast copies with total mana value 6 or less",
					uldarosTheorixResolve),
				Clauses(
					uldarosSlot("artifact"),
					Distinct(uldarosSlot("battle")),
					Distinct(uldarosSlot("creature")),
					Distinct(uldarosSlot("enchantment")),
					Distinct(uldarosSlot("instant")),
					Distinct(uldarosSlot("kindred")),
					Distinct(uldarosSlot("planeswalker")),
					Distinct(uldarosSlot("sorcery")),
				),
			),
		},
	})
}

// uldarosTheorixSlots is how many card-type clauses the trigger has; the
// resolution reads each clause's pick by its position.
const uldarosTheorixSlots = 8

// uldarosTheorixBudget is "total mana value 6 or less".
const uldarosTheorixBudget = 6

// uldarosTheorixWasCast is "if you cast him".
func uldarosTheorixWasCast(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source != nil && source.Provenance.FromZone != ""
}

// uldarosSlot is one "up to one target <type> card from your graveyard"
// clause: a nonland card of that type that the caster owns. The type is
// read through the card's effective types, so a card with several
// qualifies for each of its types and the player picks the slot.
func uldarosSlot(cardType string) *game.TargetSpec {
	return TargetCardInGraveyard("up to one target nonland "+cardType+" card from your graveyard",
		YouOwn(), Nonland(), ofCardType(cardType)).WithCount(0, 1)
}

// ofCardType passes for a card that has the lower-case card type.
func ofCardType(cardType string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasCardType(cardType) }
}

// uldarosTheorixResolve exiles the still-legal picks, then copies what
// really reached exile and offers a budgeted cast of the copies.
func uldarosTheorixResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for slot := 0; slot < uldarosTheorixSlots; slot++ {
		if t, ok := ctx.ClauseTarget(slot); ok && t.Kind == game.TargetCard {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		budget := uldarosTheorixBudget
		return CopyAndCast{Cards: exiled, Free: true, TotalManaValueAtMost: &budget}.Apply(NewContext(g, item))
	})
}
