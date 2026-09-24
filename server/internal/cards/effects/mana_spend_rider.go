package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_spend_rider.go — the catalog side of #1547: what a mana
// ability's mana does when it is spent. The engine half — the rider on
// the token, where it fires, and what reads it — is
// game/mana_spend_rider.go; ADR 0040's 2026-09-24 amendment is the
// design.
//
// One constructor per printed shape, and the filter is always the
// printed clause, word for word, in the restriction vocabulary:
//
//	"that spell can't be countered"                 SpentSpellCantBeCountered
//	"if that mana is spent on a creature spell,
//	 it gains haste"                                SpentCreatureGainsHaste
//	"that creature enters with an additional
//	 +1/+1 counter on it"                           SpentEntersWithCounters
//	"when that mana is spent to cast …, <effect>"   WhenManaSpent
//
// A rider's filter decides whether it FIRES, never whether the mana
// may pay — that is the ability's Restrictions, which some rider cards
// have (Cavern of Souls) and some do not (Hall of the Bandit Lord). A
// card with both states the clause twice, once for each job, because
// the printed card does: Cavern's mana pays only for a chosen-type
// creature spell, AND that spell can't be countered.

// SpentSpellCantBeCountered is "and that spell can't be countered" /
// "if that mana is spent on an instant or sorcery spell, that spell
// can't be countered" (Cavern of Souls, Delighted Halfling, Boseiju).
// `when` is the spend filter; pass game.ManaRestrictCast at least — the
// clause is about a SPELL, and an ability can't be countered this way.
func SpentSpellCantBeCountered(when ...string) game.ManaSpendRider {
	return game.ManaSpendRider{Kind: game.ManaRiderCantBeCountered, When: when}
}

// SpentCreatureGainsHaste is "if that mana is spent on a creature spell,
// it gains haste" (Hall of the Bandit Lord): the permanent the spell
// becomes has haste for as long as it remains that object. The creature
// half of the filter is built in.
func SpentCreatureGainsHaste() game.ManaSpendRider {
	return game.ManaSpendRider{
		Kind: game.ManaRiderHaste,
		When: []string{ManaRestrictCast, ManaRestrictType("Creature")},
	}
}

// SpentEntersWithCounters is "if this mana is spent to cast a creature
// spell, that creature enters with an additional <n> <kind> counter(s)
// on it" (Biophagus). A CR 614.1c entry clause, seeded with the card's
// own, so a counter doubler sees it.
func SpentEntersWithCounters(kind string, n int, when ...string) game.ManaSpendRider {
	return game.ManaSpendRider{
		Kind:        game.ManaRiderEntersWithCounters,
		CounterKind: kind,
		Counters:    n,
		When:        when,
	}
}

// WhenManaSpent is "when that mana is spent to cast …, <effect>"
// (Pyromancer's Goggles, Scaled Nurturer, Path of Ancestry): a
// triggered ability that goes on the stack above the spell the mana
// paid for.
//
// `key` names the trigger in the engine's registry and must be unique
// in the catalog — the card's name is the convention. The token carries
// the key, not the closure, which is what lets the rider survive the
// snapshot. This constructor registers `t` under it, so it must run
// exactly once, from the card's init.
//
// The effect reads "that spell" off ctx.PayloadCards()[0]; it does not
// target it.
func WhenManaSpent(key string, t game.ManaSpendTrigger, when ...string) game.ManaSpendRider {
	game.RegisterManaSpendTrigger(key, t)
	return game.ManaSpendRider{Kind: game.ManaRiderTrigger, Trigger: key, When: when}
}

// validateManaSpendRider is Register's check on one declared rider: a
// rider that names a kind the engine does not read, a counter rider with
// no counters, or a trigger nobody registered would all ship a card that
// silently does less than it claims.
func validateManaSpendRider(r game.ManaSpendRider) error {
	switch r.Kind {
	case game.ManaRiderCantBeCountered, game.ManaRiderHaste:
	case game.ManaRiderEntersWithCounters:
		if r.CounterKind == "" || r.Counters <= 0 {
			return fmt.Errorf("an enters-with-counters spend rider needs a counter kind and a positive count")
		}
	case game.ManaRiderTrigger:
		if _, ok := game.ManaSpendTriggerFor(r.Trigger); !ok {
			return fmt.Errorf("spend rider names trigger %q, which is not registered — build it with WhenManaSpent", r.Trigger)
		}
	default:
		return fmt.Errorf("unknown spend rider kind %q", r.Kind)
	}
	if r.Production != uuid.Nil || r.Applied {
		return fmt.Errorf("a declared spend rider carries Production / Applied, which the engine stamps")
	}
	return nil
}

// sharesCreatureTypeWithYourCommander is Path of Ancestry's filter: "a
// creature spell that shares a creature type with your commander".
// "Your commander" is any commander the player OWNS, wherever it is —
// the command zone, the battlefield, the stack (it may be the spell
// being paid for), a hand, a graveyard or exile — and with partners
// either one will do.
func sharesCreatureTypeWithYourCommander(g *game.Game, controller uuid.UUID, spell game.Card) bool {
	if !spell.IsCreature() {
		return false
	}
	found := false
	visit := func(zone *game.Zone) {
		if zone == nil || found {
			return
		}
		for i := range zone.Cards {
			c := zone.Cards[i]
			if c.IsCommander && c.Owner == controller && game.SharesCreatureType(&spell, &c) {
				found = true
				return
			}
		}
	}
	visit(g.Battlefield)
	visit(g.Stack)
	visit(g.Exile)
	if p := g.PlayerByIDForEffect(controller); p != nil {
		visit(p.Command)
		visit(p.Hand)
		visit(p.Graveyard)
	}
	return found
}
