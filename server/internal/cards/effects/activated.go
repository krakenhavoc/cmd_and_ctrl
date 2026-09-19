package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// activated.go — S21 sub-PR 2: cost constructors for
// Spec.Activated, so a card file reads like its oracle text:
//
//	Cost: SacrificeACreature(),      // "Sacrifice a creature:"
//	Cost: TapCost(),                 // "{T}:"
//	Cost: SacrificeThis(),           // "Sacrifice this permanent:"
//	Cost: ManaCost("{1}{B}"),        // "{1}{B}:"
//
// Costs compose with Plus for the multi-part case ("{2}, {T}:").

// TapCost is "{T}" — the source must be untapped, and a creature
// source must be free of summoning sickness (CR 302.6).
func TapCost() game.AbilityCost { return game.AbilityCost{Tap: true} }

// SacrificeThis sacrifices the source as the cost.
func SacrificeThis() game.AbilityCost { return game.AbilityCost{SacrificeSelf: true} }

// SacrificeACreature is the sac-outlet cost: "Sacrifice a
// creature". The source itself qualifies when it's a creature —
// Carrion Feeder really can eat itself.
func SacrificeACreature() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a creature", Creature())}
}

// SacrificeAPermanent is "Sacrifice a permanent" (Altar of
// Dementia's broader cousin).
func SacrificeAPermanent() game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec("a permanent")}
}

// SacrificeN is "Sacrifice N <permanents>" — "Sacrifice two
// artifacts" (Sai, Master Thopterist) is
//
//	SacrificeN(2, "two artifacts", Artifact())
//
// and "Sacrifice three Foods" (Samwise Gamgee) is
// SacrificeN(3, "three Foods", HasSubtype("Food")). The label is the
// clause as printed, without the verb; the client shows it in the
// picker.
//
// The count lives on the clause (Min == Max == n, #747), not in a
// separate field, so Plus cannot drop it. The activator names exactly
// n distinct permanents they control, and all of them leave as one
// simultaneous exit. "Sacrifice two OTHER creatures" adds a predicate
// that excludes the source (Priest of Forgotten Gods). A mana ability
// takes SacrificeN(...).SacrificeOther, the way it takes
// SacrificeACreature().SacrificeOther.
//
// Variable counts ("Sacrifice X Treasures", "one or more") have no
// shape: Register refuses a clause whose Min and Max differ.
func SacrificeN(n int, label string, preds ...CardPredicate) game.AbilityCost {
	return game.AbilityCost{SacrificeOther: sacrificeSpec(label, preds...).WithCount(n, n)}
}

// ManaCost is a printed mana component, "{1}{B}". An {X} in the
// string is a real variable cost — the activator announces a value
// for it at CR 602.2b and the effect reads it back with ctx.X() —
// so "{X}" and "{X}{X}" (Treasure Vault) both go here and nothing
// else needs declaring.
func ManaCost(cost string) game.AbilityCost { return game.AbilityCost{Mana: cost} }

// MinX is the floor the printed text puts on X: "X can't be 0"
// (Helm of Obedience) is MinX(1). Compose it onto the mana
// component — Plus(ManaCost("{X}"), TapCost(), MinX(1)).
//
// A floor is not the same as an unaffordable X. The engine refuses
// an announcement below it outright, and the enumerator never offers
// the ability at all when the activator cannot reach the floor —
// which is the difference between "Helm does nothing for {0}" and
// "Helm is not activatable right now", and only the second is the
// card.
func MinX(n int) game.AbilityCost { return game.AbilityCost{MinX: n} }

// PayLife is a life component (CR 119.4).
func PayLife(n int) game.AbilityCost { return game.AbilityCost{Life: n} }

// LoyaltyCost is a planeswalker's loyalty cost — the "+1", "[0]" or
// "−3" printed to the left of the ability. Positive adds counters,
// negative removes them, zero does neither and still spends the
// turn's activation (CR 606.3).
//
// Setting it is the whole declaration: the engine derives sorcery
// speed, once-per-turn, "must be a planeswalker you control" and
// CR 606.6 from the presence of the component, so a card file
// writes the cost and nothing else. ADR 0020 kept loyalty out of
// AbilityCost; ADR 0032 §7 reversed that — see the field comment on
// game.AbilityCost.Loyalty.
func LoyaltyCost(n int) game.AbilityCost { return game.AbilityCost{Loyalty: &n} }

// Plus merges cost components: Plus(ManaCost("{2}"), TapCost()) is
// "{2}, {T}". Later components win for scalar fields, which only
// matters if a caller passes two mana strings (they shouldn't).
func Plus(costs ...game.AbilityCost) game.AbilityCost {
	var out game.AbilityCost
	for _, c := range costs {
		if c.Tap {
			out.Tap = true
		}
		if c.SacrificeSelf {
			out.SacrificeSelf = true
		}
		if c.SacrificeOther != nil {
			out.SacrificeOther = c.SacrificeOther
		}
		if c.Mana != "" {
			out.Mana = c.Mana
		}
		if c.Life != 0 {
			out.Life = c.Life
		}
		if c.Loyalty != nil {
			out.Loyalty = c.Loyalty
		}
		if c.Crew != 0 {
			out.Crew = c.Crew
		}
		// #625: without this a composed "{T}, Remove a +1/+1 counter"
		// silently loses its counter component and the ability becomes
		// free to repeat — stronger than printed, the #259 direction.
		if c.RemoveCounters != nil {
			out.RemoveCounters = c.RemoveCounters
		}
		// #789: the same reasoning, the other direction — a composed
		// "{T}, Put a -1/-1 counter on this creature" that dropped
		// the counter would untap for free.
		if c.AddCounter != nil {
			out.AddCounter = c.AddCounter
		}
		if c.MinX != 0 {
			out.MinX = c.MinX
		}
		// #660: without these two a composed "{2}, Discard this card"
		// silently loses its discard and cycling becomes a free draw
		// — the same failure mode the counter component above had.
		if c.DiscardSelf {
			out.DiscardSelf = true
		}
		if c.DiscardCards != nil {
			out.DiscardCards = c.DiscardCards
		}
	}
	return out
}

// DiscardThis is cycling's "Discard this card" cost component
// (CR 702.29a). It only means anything on an ability that functions
// from the hand, and Register refuses it anywhere else — build the
// ability with Cycling / Typecycling rather than composing this by
// hand.
func DiscardThis() game.AbilityCost { return game.AbilityCost{DiscardSelf: true} }

// DiscardACard is "Discard a card" as a cost — Cryptbreaker's
// "{1}{B}, {T}, Discard a card:". The activator picks from hand at
// announce (CR 602.2b).
func DiscardACard() game.AbilityCost {
	return DiscardN(1, "a card")
}

// DiscardCardsMatching is "Discard a <kind> card" — Fauna Shaman's
// "{G}, {T}, Discard a creature card:", Survival of the Fittest's
// "{G}, Discard a creature card:", Tortured Existence's "{B},
// Discard a creature card:". The label is the clause as printed,
// without the verb; the client shows it in the picker.
func DiscardCardsMatching(n int, label string, match func(game.Card) bool) game.AbilityCost {
	return game.AbilityCost{DiscardCards: &game.DiscardCost{N: n, Label: label, Match: match}}
}

// DiscardN is "Discard N cards" with no restriction on which —
// DiscardN(2, "two cards").
func DiscardN(n int, label string) game.AbilityCost {
	return DiscardCardsMatching(n, label, nil)
}

// sacrificeSpec builds the "what may I sacrifice" clause. It reuses
// TargetSpec because the shape is identical — a predicate over
// battlefield cards — but it is NOT targeting: a sacrifice cost
// doesn't target, so hexproof and "can't be the target of" never
// apply to it. The controller filter lives in the engine's cost
// validation (CR 701.21a).
func sacrificeSpec(label string, preds ...CardPredicate) *game.TargetSpec {
	return TargetPermanent(label, preds...)
}

// RemoveCountersFromThis is "Remove N <kind> counters from this
// permanent" — Dragon's Hoard's "Remove a gold counter from this
// artifact" is RemoveCountersFromThis("gold", 1). Paid at announce
// (#625), so a response cannot spend the same counter twice.
func RemoveCountersFromThis(kind string, n int) game.AbilityCost {
	return game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{Counter: kind, N: n}}
}

// RemoveCountersFrom is "Remove N <kind> counters from <a permanent
// you control>": Heart of Kiran's "remove a loyalty counter from a
// planeswalker you control" is
//
//	RemoveCountersFrom(game.CounterLoyalty, 1, "a planeswalker you control", Planeswalker())
//
// An empty kind is "a counter" of any kind (Fain, the Broker), and the
// activator names the kind with the permanent; Register refuses an
// empty kind with n > 1, because one kind choice cannot say how a
// "remove two counters" that mixes kinds was paid.
//
// Like a sacrifice clause the permanent is chosen, not targeted, and
// "you control" is the engine's rule rather than a predicate the card
// file has to remember — the label says it for the player.
func RemoveCountersFrom(kind string, n int, label string, preds ...CardPredicate) game.AbilityCost {
	return game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		Counter: kind,
		N:       n,
		From:    TargetPermanent(label, preds...),
	}}
}

// RemoveCountersXFromThis is "Remove X <kind> counters from this
// permanent" and "Remove any number of <kind> counters from this
// permanent" — they are the same cost, because what X can be is
// bounded by what the permanent holds and by nothing else (#789).
// Mage-Ring Network's "Remove any number of storage counters from
// this land" is RemoveCountersXFromThis("storage", 0); a card that
// printed a floor ("X can't be 0") passes 1.
//
// The count is announced at activation, the way MinX announces an
// {X} in a mana component, and reaches the effect through the paid-
// cost record: an activated ability reads ctx.CountersRemoved(), and
// a mana ability's ProducedForPaid is handed the record directly.
func RemoveCountersXFromThis(kind string, min int) game.AbilityCost {
	return game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		Counter:  kind,
		N:        min,
		Variable: true,
	}}
}

// RemoveCountersAmong is "Remove N <kind> counters from among <a
// clause>" — Iron Spider, Stark Upgrade's "from among artifacts you
// control", Hopeful Initiate's "from among creatures you control"
// (#789):
//
//	RemoveCountersAmong("+1/+1", 2, "artifacts you control", Artifact())
//
// The N counters may be split across any number of the permanents
// the clause matches, in whatever amounts the activator names. Like
// every other counter cost the permanents are chosen, not targeted,
// and "you control" is the engine's rule rather than a predicate the
// card file has to remember.
//
// An empty kind is the ANY-KIND among form (#943): Tekuthal, Inquiry
// Dominus' "Remove three counters from among other artifacts,
// creatures, and planeswalkers you control" takes three counters of
// whatever kinds are there, so the activator names a kind per
// permanent as well as a count. It is the same constructor and the
// same component — the kind question is simply asked once per part
// instead of once per payment.
func RemoveCountersAmong(kind string, n int, label string, preds ...CardPredicate) game.AbilityCost {
	return game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
		Counter: kind,
		N:       n,
		From:    TargetPermanent(label, preds...),
		Among:   true,
	}}
}

// AddCounterToThis is "Put a <kind> counter on this permanent" as a
// COST — Devoted Druid's "Put a -1/-1 counter on this creature:
// Untap this creature" (#789).
//
// A cost, not an effect, with everything that follows from CR 121.1:
// nothing doubles it (Doubling Season does not make the Druid's
// untapper cost two counters), nothing replaces it, and an activator
// who cannot have the counter put on cannot activate at all
// (CR 118.3).
func AddCounterToThis(kind string, n int) game.AbilityCost {
	return game.AbilityCost{AddCounter: &game.CounterAddCost{Counter: kind, N: n}}
}
