package game

import (
	"sort"

	"github.com/google/uuid"
)

// granted_alternative_cost.go — ADR 0118 §3, #2163: a static ability of
// a permanent that offers its controller one more alternative cost
// (CR 118.9) for each spell they cast.
//
//	"You may pay {W}{U}{B}{R}{G} rather than pay the mana cost for
//	 spells you cast." (Jodah, Archmage Eternal; Fist of Suns;
//	 Leyline of Mutation)
//	"You may cast spells from your hand without paying their mana
//	 costs." (Omniscience)
//
// CR 118.9 names both: an alternative cost is "listed in a spell's
// text, or applied to it from another effect", and "You may cast
// [this object] without paying its mana cost" is the second phrasing.
//
// Its own declaration, not a CastPermission (ADR 0118 call 6). A
// permission is the REASON a cast is legal from a zone and may carry
// one price; these statics add a price wherever the printed mana cost
// may already be paid. Nothing is stored: the offers are derived from
// the battlefield on every query, exactly as standing permissions and
// cast-timing statements are, so the offer lasts as long as the source
// is on the battlefield under its controller and two sources compose.
//
// Two hooks read it, and everything else reads those two:
//
//   - CastOffersForLocked lists the offers, after the card's own and
//     the permissions' (announce precedence), so the view's picker, the
//     strip's prices, the drag verdict, the bot enumerator and the
//     commander-return check all see them.
//   - resolveAlternativeCostLocked accepts a claim of one, so CastSpell
//     and effectiveCostLocked (the price, the auto-tap preview and the
//     auto-tapper) charge it.
//
// validateCastPathLocked holds the one rule (CR 118.9a, 601.2b): a
// granted offer may be claimed only where the printed mana cost could
// be paid. See grantedOfferClaimableLocked.
//
// The keys are on-disk identities — a claim lands on StackItem.AltCost
// and is captured with the stack — so, like a token slug, a key is
// never renamed or reused. They are namespaced "granted-…" so a card's
// own offer is never shadowed by CastOffersForLocked's per-key dedupe.
// A restored stack item whose key this binary no longer grants is
// inert: every resolution reader looks the key up with
// AlternativeCostByKey, which answers nil for a key the card does not
// print.

// GrantedAlternativeCost is one "you may pay X rather than pay the mana
// cost for spells you cast" static, as a catalog declaration
// (Spec.GrantedAlternativeCosts).
type GrantedAlternativeCost struct {
	// Offer is the price. Key, Label, ManaCost, Energy (Nissa,
	// Worldsoul Speaker's eight {E}, ADR 0129 §5) and AsThoughFlash
	// (Primal Prayers) are read; Life and the card components stay
	// unused (no printed card grants one, and Register refuses them). An
	// empty ManaCost is free — "without paying its mana cost" — as for
	// any AlternativeCost.
	Offer AlternativeCost

	// Zones are the zones the offer reaches: Omniscience's "from your
	// hand". Nil means every zone a spell is cast from.
	Zones []ZoneKind

	// Spells narrows which spells the offer reaches (ADR 0129 §5):
	// Nissa, Worldsoul Speaker's "permanent spells you cast"
	// (NonLandPermanentOnly), Primal Prayers' "creature spells"
	// (CreatureOnly). Judged against the face being cast. The zero
	// filter is every spell, which is Jodah's and Omniscience's.
	Spells PermissionFilter

	// MaxManaValue caps the mana value of the spell the offer reaches:
	// Primal Prayers' "with mana value 3 or less". Nil means no cap.
	// The mana value is the card's as it is cast for this offer, which
	// replaces the mana cost, so an {X} in it counts as 0 (CR 107.3b,
	// CR 202.3).
	MaxManaValue *int
}

// covers reports whether this static's offer reaches a cast of `card`.
func (gr GrantedAlternativeCost) covers(card Card) bool {
	if !gr.Spells.Matches(card) {
		return false
	}
	return gr.MaxManaValue == nil || card.ManaValue() <= *gr.MaxManaValue
}

// reaches reports whether this static covers a cast out of `zone`.
func (gr GrantedAlternativeCost) reaches(zone ZoneKind) bool {
	if len(gr.Zones) == 0 {
		return true
	}
	for _, z := range gr.Zones {
		if z == zone {
			return true
		}
	}
	return false
}

// CatalogGrantedAlternativeCosts returns the granted-alternative-cost
// statics a battlefield permanent with the given catalog ABILITY key
// has. Installed at init to read CardDef.GrantedAlternativeCosts; a
// game-package test may stub it. Nil, or a nil return, means none.
var CatalogGrantedAlternativeCosts func(abilityKey string) []GrantedAlternativeCost

// grantedAlternativeCostsLocked derives the offers the permanents
// `playerID` controls grant to a cast of `card` out of `zone`: one per
// key, each a fresh copy with Granted set and its label suffixed with
// the source's name. Two sources granting the same key give one offer,
// labelled after the first in battlefield order.
//
// CatalogAbilityKey rather than CatalogKey, for the reason
// standingCastPermissionsLocked gives: these are static abilities, and
// a Jodah that has lost its abilities (Humility) grants nothing.
//
// A land gets none: a land is not a spell, and has no mana cost to
// replace (CR 305.1). Not filtered by claimability or payability:
// CastOffersForLocked and the announce path apply those gates.
//
// Caller must hold g.mu.
func (g *Game) grantedAlternativeCostsLocked(playerID uuid.UUID, card Card, zone ZoneKind) []*AlternativeCost {
	if CatalogGrantedAlternativeCosts == nil || g.Battlefield == nil || playerID == uuid.Nil || card.IsLand() {
		return nil
	}
	var out []*AlternativeCost
	var seen map[string]bool
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller != playerID {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, gr := range CatalogGrantedAlternativeCosts(key) {
			if gr.Offer.Key == "" || !gr.reaches(zone) || !gr.covers(card) || seen[gr.Offer.Key] {
				continue
			}
			if seen == nil {
				seen = make(map[string]bool, 2)
			}
			seen[gr.Offer.Key] = true
			offer := gr.Offer
			offer.Granted = true
			if offer.Label == "" {
				offer.Label = offer.Key
			}
			if src.Name != "" {
				offer.Label += " (" + src.Name + ")"
			}
			out = append(out, &offer)
		}
	}
	return out
}

// grantedAlternativeCostByKeyLocked is the claim half: the granted
// offer named `key` for this cast, or nil.
//
// Caller must hold g.mu.
func (g *Game) grantedAlternativeCostByKeyLocked(playerID uuid.UUID, card Card, zone ZoneKind, key string) *AlternativeCost {
	if key == "" {
		return nil
	}
	for _, ac := range g.grantedAlternativeCostsLocked(playerID, card, zone) {
		if ac.Key == key {
			return ac
		}
	}
	return nil
}

// grantedOfferClaimableLocked is validateCastPathLocked's rule for a
// granted offer (CR 118.9a: "Only one alternative cost can be applied
// to any one spell as it's being cast"; CR 601.2b: "A player can't
// apply two alternative methods of casting or two alternative costs to
// a single spell"). A granted offer replaces the MANA COST, so it may
// be claimed exactly where the printed mana cost may be paid:
//
//   - a claim of nothing is allowed out of this zone under this
//     permission, so a flashback card in a graveyard (rule 3), a card
//     Snapcaster gave flashback to and the top of the library under
//     Bolas's Citadel (rule 4) get no granted offer; and
//   - the permission charges no price of its own: no flat Cost
//     (cascade, discover, airbend, a Siege) and no claimable offer.
//     Such a cast is already alternatively costed.
//
// An impulse-exiled card, a Gravecrawler, a Future Sight top card, a
// hand card and a commander pay their printed cost, so they get the
// offer; a commander still pays the tax on top (CR 118.9d, 903.8).
//
// The permission is judged as the cast would use it: narrowed by
// ForClaim, so a live miracle grant over a hand card (which opens one
// claim and nothing else) does not keep the card from paying Jodah's
// price at sorcery speed.
//
// Caller must hold g.mu.
func (g *Game) grantedOfferClaimableLocked(card Card, zone ZoneKind, alt *AlternativeCost, grant *CastPermission) error {
	under := grant.ForClaim(alt)
	if err := g.validateCastPathLocked(card, zone, nil, under); err != nil {
		return err
	}
	if under != nil && (under.Cost != "" || under.AltCostKey != "") {
		return ErrCastCostRequired
	}
	return nil
}

// duplicatesListedPrice reports whether a granted offer only repeats a
// price already in `listed` (ADR 0118 call 3): the same mana cost as
// the printed cost (the nil entry — Sliver Queen under Fist of Suns),
// or as an offer whose only component is mana (a Bringer's own
// {W}{U}{B}{R}{G}). Two identical rows would only ask the caster to
// choose between equals.
//
// A card with no mana cost (CR 118.6, Ancestral Vision) has no printed
// price to repeat, so Omniscience's free offer is never dropped for it.
func duplicatesListedPrice(card Card, listed []*AlternativeCost, granted *AlternativeCost) bool {
	// ADR 0129 §5: an offer that charges more than mana is never a
	// repeat. Primal Prayers' "pay {E}" has no mana at all, and an
	// Ornithopter's printed {0} is not the same price.
	if !granted.onlyMana() {
		return false
	}
	for _, ac := range listed {
		if ac == nil {
			if !HasNoManaCost(card) && sameManaPrice(card.ManaCost, granted.ManaCost) {
				return true
			}
			continue
		}
		if ac.onlyMana() && sameManaPrice(ac.ManaCost, granted.ManaCost) {
			return true
		}
	}
	return false
}

// onlyMana reports whether this offer's whole price is its mana cost:
// no condition, no life, no card to pitch, return, exile or sacrifice,
// and no rider that changes the spell (a target rewrite, a face, an
// entry clause, an exile on leaving the stack).
func (a *AlternativeCost) onlyMana() bool {
	return a != nil && a.Condition == nil && a.Life == 0 && a.Energy == 0 && !a.AsThoughFlash && !a.PaysCards() &&
		a.Targets == nil && !a.ClearsTargets && !a.SacrificeOnEntry &&
		!a.ExileOnLeavingStack && !a.WarpExile && a.EntersWithCounterName == "" &&
		a.FaceDown == nil && !a.RequiresGrant && a.CastsFace == 0
}

// sameManaPrice reports whether two cost strings ask for the same mana,
// symbol for symbol in any order. An unparseable string matches
// nothing.
func sameManaPrice(a, b string) bool {
	pa, err := ParseCost(a)
	if err != nil {
		return false
	}
	pb, err := ParseCost(b)
	if err != nil {
		return false
	}
	if pa.Generic != pb.Generic || pa.XSlots != pb.XSlots || len(pa.Required) != len(pb.Required) {
		return false
	}
	sa, sb := requirementSymbols(pa), requirementSymbols(pb)
	for i := range sa {
		if sa[i] != sb[i] {
			return false
		}
	}
	return true
}

// requirementSymbols is a parsed cost's coloured symbols, sorted.
func requirementSymbols(c ParsedCost) []string {
	out := make([]string, len(c.Required))
	for i, r := range c.Required {
		out[i] = r.String()
	}
	sort.Strings(out)
	return out
}
