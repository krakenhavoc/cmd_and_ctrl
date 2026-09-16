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
// source must be free of summoning sickness (CR 302.1).
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

// PayLife is a life component (CR 118.8).
func PayLife(n int) game.AbilityCost { return game.AbilityCost{Life: n} }

// LoyaltyCost is a planeswalker's loyalty cost — the "+1", "[0]" or
// "−3" printed to the left of the ability. Positive adds counters,
// negative removes them, zero does neither and still spends the
// turn's activation (CR 606.5).
//
// Setting it is the whole declaration: the engine derives sorcery
// speed, once-per-turn, "must be a planeswalker you control" and
// CR 606.3 from the presence of the component, so a card file
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
		if c.MinX != 0 {
			out.MinX = c.MinX
		}
	}
	return out
}

// sacrificeSpec builds the "what may I sacrifice" clause. It reuses
// TargetSpec because the shape is identical — a predicate over
// battlefield cards — but it is NOT targeting: a sacrifice cost
// doesn't target, so hexproof and "can't be the target of" never
// apply to it. The controller filter lives in the engine's cost
// validation (CR 701.17b).
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
