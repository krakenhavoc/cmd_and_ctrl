package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cost_modifier.go — S28: constructors for Spec.CostModifiers, the
// "spells cost {N} more / {N} less to cast" static (CR 601.2f). The
// engine half lives in game/cost_modifier.go; this file is the
// vocabulary a card file writes in.
//
// The shape to copy:
//
//	CostModifiers: []game.CostModifier{
//	    CostsMore(1, "Noncreature spells cost {1} more to cast.", NoncreatureSpell()),
//	    CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
//	        YourSpell(), InstantOrSorcerySpell()),
//	},
//
// Predicates are ANDed, so a clause reads left to right in the order
// the oracle text says it: "instant and sorcery spells YOU cast" is
// YourSpell() plus InstantOrSorcerySpell(). An empty predicate list
// is "spells", full stop — Sphere of Resistance.
//
// Every constructor takes the printed clause as its label rather
// than generating one. A generated "costs {1} more" is right until
// two modifiers on the board disagree about which spells they touch,
// and then the event log says the same thing twice about different
// cards. Quoting the oracle text costs one string and makes the log
// readable.

// CostPredicate narrows a cost modifier to the spells its card
// actually names. Mirrors CardPredicate in targets.go, one type
// over: the subject is a cast rather than a permanent, so it reads a
// game.CostQuery instead of a game.Card.
//
// READ-ONLY and under the cast path's write lock — the same contract
// every predicate in the catalog has.
type CostPredicate func(q game.CostQuery) bool

// allOf ANDs a predicate list into the single hook the engine wants.
// Nil for an empty list, which the engine reads as "applies to every
// spell" — one fewer closure on the hot path for the commonest case.
func allOf(preds []CostPredicate) func(game.CostQuery) bool {
	if len(preds) == 0 {
		return nil
	}
	return func(q game.CostQuery) bool {
		for _, p := range preds {
			if p == nil || !p(q) {
				return false
			}
		}
		return true
	}
}

// CostsMore is "<matching> spells cost {n} more to cast" — Sphere of
// Resistance, Thalia, Thorn of Amethyst, Aura of Silence.
func CostsMore(n int, label string, when ...CostPredicate) game.CostModifier {
	return CostsMoreEach(func(game.CostQuery) int { return n }, label, when...)
}

// CostsMoreEach is CostsMore with an amount computed per cast —
// Damping Sphere's "{1} more for each other spell that player has
// cast this turn".
func CostsMoreEach(amount func(q game.CostQuery) int, label string, when ...CostPredicate) game.CostModifier {
	return game.CostModifier{
		Kind:      game.CostIncrease,
		Label:     label,
		AppliesTo: allOf(when),
		Amount:    amount,
	}
}

// CostsLess is "<matching> spells cost {n} less to cast" — Goblin
// Electromancer, Heartless Summoning.
//
// The reduction spends against GENERIC mana only and stops at zero;
// that rule lives in the engine (game.reduceGeneric) so no card file
// can get it wrong. A card that wants to reduce a coloured
// requirement is not this — no card does, and CR 601.2f is why.
func CostsLess(n int, label string, when ...CostPredicate) game.CostModifier {
	return CostsLessEach(func(game.CostQuery) int { return n }, label, when...)
}

// CostsLessEach is CostsLess with an amount computed per cast —
// Animar's "{1} less for each +1/+1 counter on Animar".
func CostsLessEach(amount func(q game.CostQuery) int, label string, when ...CostPredicate) game.CostModifier {
	return game.CostModifier{
		Kind:      game.CostReduction,
		Label:     label,
		AppliesTo: allOf(when),
		Amount:    amount,
	}
}

// CostsAtLeast is a cost-SETTING effect: "each spell that would cost
// less than `n` mana to cast costs `n` mana to cast" (Trinisphere).
// The shortfall is made up in generic mana, so a {1}{B} spell under
// a three-floor costs {2}{B} rather than {3}.
//
// Applied after every increase and every reduction on the board,
// which is the only reading under which "would cost less than three"
// means anything.
func CostsAtLeast(n int, label string, when ...CostPredicate) game.CostModifier {
	return game.CostModifier{
		Kind:      game.CostFloor,
		Label:     label,
		AppliesTo: allOf(when),
		Amount:    func(game.CostQuery) int { return n },
	}
}

// --- abilities, not spells (#1184) -------------------------------

// ActivationCostsLess is "<matching> abilities cost {n} less to
// activate" — Boom Scholar. The activation twin of CostsLess, and
// the ONLY difference is which announcements it is shown: the
// Activations bit partitions the board's modifiers into the ones
// that price casts and the ones that price activations, so a Sphere
// of Resistance written as "spells" never silently starts taxing
// abilities. Everything else — the CR 601.2f order, the generic
// floor, the negative-amount refusal — is the same pass.
//
// Note the predicates below are ability predicates: an activation's
// q.Card is the SOURCE permanent rather than a spell, so a
// spell-shaped predicate (CreatureSpell, SpellManaValueAtLeast) means
// something different here and should not be reached for.
func ActivationCostsLess(n int, label string, when ...CostPredicate) game.CostModifier {
	m := CostsLessEach(func(game.CostQuery) int { return n }, label, when...)
	m.Activations = true
	return m
}

// AnExhaustAbilityCost — the ability being priced prints the exhaust
// keyword (#1181). Reads the bit off the query rather than the
// ability's label, for the same reason the trigger predicate does:
// the bit is the declaration and the label is prose.
func AnExhaustAbilityCost() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Ability != nil && q.Ability.Exhaust
	}
}

// OfAnotherPermanentYouControl — "abilities of OTHER permanents you
// control". Two clauses in one predicate because the card prints them
// as one phrase: the source of the ability is controlled by the
// modifier's controller, and it is not the modifier's own permanent.
//
// The "other" half is the one that matters on the board: Boom Scholar
// prints an exhaust ability of its own, and discounting that too
// would be a cheaper card than the one in the pack.
func OfAnotherPermanentYouControl() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Card.Controller == q.Source.Controller &&
			q.Card.InstanceID != q.Source.InstanceID
	}
}

// --- special actions, not spells or abilities (#1319) -------------

// SpecialActionCostsLess is "<matching> [special action] costs {n}
// less" — Ranar the Ever-Watchful's "The first card you foretell each
// turn costs {0} to foretell." The special-action twin of
// ActivationCostsLess, one door further: the SpecialActions bit
// partitions the board's modifiers a third way, so a card written for
// spells or for activated abilities can never reach a CR 116.2 action,
// and vice versa.
//
// Note the predicates below read q.SpecialAction, not q.Card as a
// spell or q.Ability as an activation: a spell-shaped or
// ability-shaped predicate means nothing here and should not be
// reached for.
func SpecialActionCostsLess(n int, label string, when ...CostPredicate) game.CostModifier {
	m := CostsLessEach(func(game.CostQuery) int { return n }, label, when...)
	m.SpecialActions = true
	return m
}

// ASpecialActionOfKind passes when the special action being priced is
// exactly `kind` — "the first card you foretell" names foretell and
// not suspend or turn_face_up, even from the same permanent.
func ASpecialActionOfKind(kind game.SpecialActionKind) CostPredicate {
	return func(q game.CostQuery) bool {
		return q.SpecialAction != nil && q.SpecialAction.Kind == kind
	}
}

// TheFirstOneThisTurn passes when the controller taking the special
// action hasn't foretold anything yet this turn — Ranar's "the FIRST
// card you foretell each turn". Reads Game.ForetoldThisTurn, which
// SpecialActionManaCostForEffect's caller (PerformSpecialAction) bumps
// only once the card has actually landed in exile, so a refused or
// paused foretell never counts as the one that used up the discount.
func TheFirstOneThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.ForetoldCountThisTurn(q.Controller) == 0
	}
}

// --- who cast it -------------------------------------------------

// YourSpell passes on a spell cast by the modifier source's own
// controller — the "you" in "spells YOU cast cost {1} less".
func YourSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Controller == q.Source.Controller }
}

// OpponentsSpell passes on a spell cast by anyone else — "artifact
// and enchantment spells your OPPONENTS cast cost {2} more"
// (Aura of Silence).
//
// Anyone-else rather than a seat-by-seat opponent check because the
// game has no teams: every other player at a Commander table is an
// opponent, and a modifier source with no controller (a fixture, a
// token mid-construction) matches nobody rather than everybody.
func OpponentsSpell() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Source.Controller != ZeroUUID && q.Controller != q.Source.Controller
	}
}

// CastFromGraveyardOrExile passes on a spell being cast from a
// graveyard or from exile — "spells your opponents cast from
// graveyards or from exile cost {2} more" (Aven Interrupter). Reads
// the zone the cast is coming FROM, so flashback, escape, a warped
// card's recast, an airbent card and a plotted card are all taxed, and
// the command zone and the hand are not.
func CastFromGraveyardOrExile() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.FromZone == game.ZoneGraveyard || q.FromZone == game.ZoneExile
	}
}

// --- what kind of spell ------------------------------------------

// CreatureSpell passes on a creature spell.
func CreatureSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsCreature() }
}

// NoncreatureSpell passes on anything that is not a creature spell —
// Thalia, Thorn of Amethyst. A land is never a spell and never
// reaches this predicate: playing one is a special action (CR
// 116.2a) and the cast path returns before pricing.
func NoncreatureSpell() CostPredicate {
	return func(q game.CostQuery) bool { return !q.Card.IsCreature() }
}

// SpellOfTheSourcesChosenType passes on a spell whose subtypes carry
// the creature type the MODIFIER'S SOURCE named as it entered (CR
// 614.12) — Herald's Horn's "creature spells you cast of the chosen
// type", Urza's Incubator's, Stoneforge Acolyte's.
//
// It reads q.Source.NamedTribe, not a printed list, so two copies
// naming different types each discount their own. A source with no
// type named yet matches nothing: an empty tribe read as "every
// spell" would make the artifact a Semblance Anvil the moment it
// entered, which is the dangerous direction.
//
// The spell's subtypes are printed characteristics (the layer engine
// does not recompute a card on the stack), and HasSubtype answers
// true for every creature type on a changeling, which is the printed
// ruling.
func SpellOfTheSourcesChosenType() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Source.NamedTribe != "" && q.Card.HasSubtype(q.Source.NamedTribe)
	}
}

// InstantOrSorcerySpell passes on an instant or sorcery — Goblin
// Electromancer.
func InstantOrSorcerySpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsInstant() || q.Card.IsSorcery() }
}

// ArtifactOrEnchantmentSpell passes on an artifact or enchantment
// spell — Aura of Silence.
func ArtifactOrEnchantmentSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsArtifact() || q.Card.IsEnchantment() }
}

// ArtifactSpell passes on an artifact spell — Foundry Inspector,
// Semblance Anvil.
func ArtifactSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsArtifact() }
}

// ColoredSpell passes on a spell that IS the given colour — "BLUE
// spells you cast cost {1} less to cast" (The Water Crystal). `color`
// is a single-letter code, "W" / "U" / "B" / "R" / "G", the spelling
// game.Card.HasColor takes.
//
// Colour is read off the spell as it stands at announce, through
// EffectiveColors, so a layer-5 colour change is honoured and a
// multicolour spell is every colour it is: a {U}{R} spell is a blue
// spell and a red one, and each discount that names one applies.
func ColoredSpell(color string) CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.HasColor(color) }
}

// SpellManaValueAtLeast passes when the mana value of the spell
// being cast is at least n. It reads the mana cost, not the price:
// CR 202.3c says a cost modifier changes what a spell costs and never
// its mana value, so a Goblin Electromancer can't drop a spell below
// n. {X} counts as the announced value (CR 202.3e). By the time the
// total cost is locked in, the spell is on the stack with X chosen
// (CR 601.2b, 601.2f).
func SpellManaValueAtLeast(n int) CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.ManaValueWithX(q.XValue) >= n }
}

// --- the source's own state --------------------------------------

// SourceUntapped passes while the permanent contributing the
// modifier is untapped — Trinisphere's "as long as this artifact is
// untapped".
//
// An ordinary predicate rather than engine machinery because the
// modifier pass reads the source's live battlefield state on every
// cast; there is nothing to invalidate and nothing to recompute.
func SourceUntapped() CostPredicate {
	return func(q game.CostQuery) bool { return !q.Source.Tapped }
}

// --- counting ----------------------------------------------------

// CountersOnSource counts the named counters on the modifier's own
// source — Animar's "+1/+1 counter on Animar". Use it as the amount
// hook of CostsLessEach.
func CountersOnSource(kind string) func(q game.CostQuery) int {
	return func(q game.CostQuery) int { return q.Source.Counters[kind] }
}

// OtherSpellsCastThisTurn counts the spells the CASTER has already
// cast this turn — Damping Sphere's "for each OTHER spell that
// player has cast this turn".
//
// "Other" comes free: the per-turn tally is bumped after the spell
// reaches the stack, and cost modifiers are priced before that, so
// the count this reads already excludes the spell being priced.
func OtherSpellsCastThisTurn() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		return q.Game.SpellsCastThisTurn[q.Controller].Total
	}
}

// SpellWithKeyword passes on a spell that has `kw` as it sits on the
// stack — "creature spells WITH FLYING you cast cost {1} less"
// (Warden of Evos Isle), and the discount half of every tribal or
// keyword lord that prices rather than pumps.
//
// Reads the keyword off the spell rather than off a battlefield
// permanent, which is the whole point: game.HasKeyword falls back to
// the card's own Keywords and then to CatalogPrintedKeywords for an
// object that is not on the battlefield, and a spell being priced is
// on the stack. A creature that only GAINS the keyword once it
// resolves is not discounted, which is what CR 601.2f says — the cost
// is locked in from the spell as it exists while it is being cast.
func SpellWithKeyword(kw string) CostPredicate {
	return func(q game.CostQuery) bool { return game.HasKeyword(&q.Card, kw) }
}
