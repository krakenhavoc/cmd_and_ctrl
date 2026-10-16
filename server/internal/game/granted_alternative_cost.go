package game

import (
	"fmt"
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
	// (Primal Prayers) are read; Life is derived (LifeEqualToManaValue)
	// rather than declared, and the card components other than emerge's
	// stay unused (Register refuses them). An
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

	// MaxManaValueFunc is a cap that follows the board: Omnipresence's
	// "mana value less than or equal to the number of creatures you
	// control" (#3503). It is asked with the game and the CASTER every
	// time the offer is derived, so the cap is whatever the count is at
	// that moment, never a number frozen when the permanent arrived. It
	// takes the Game, so it must be a pure read of it: the caller holds
	// g.mu. Nil means this static has no computed cap; with MaxManaValue
	// also set, both apply. Needs MaxManaValueText (Register panics
	// without it), the words the ability row shows for it.
	MaxManaValueFunc func(g *Game, caster uuid.UUID) int

	// MaxManaValueText words MaxManaValueFunc for the ability row:
	// "the number of creatures you control".
	MaxManaValueText string

	// PricedAtManaCost is Herigast, Erupting Nullkite's "The emerge
	// cost is equal to its mana cost" (ADR 0135 §4 and PR 6, CR
	// 702.119a): the offer's ManaCost is the mana cost of the spell
	// being cast, read off the face being cast as each offer is
	// derived, so a creature with {X} in its mana cost keeps that {X}
	// in its emerge cost and the caster chooses X as for its mana cost
	// (CR 107.3). Offer.ManaCost must be empty; the derived offer's
	// label gets the price appended. A spell with no mana cost (CR
	// 118.6), or one whose cost does not parse, gets no offer: there is
	// no mana cost for the price to equal.
	PricedAtManaCost bool

	// ActiveWhen gates the static on a designation of its source: a
	// Room door (CR 709.5, #3174). A locked half has no rules text, so
	// Access Maze and Warped Space offer nothing until their door is
	// unlocked. Zero is always on. effects.Room stamps it.
	ActiveWhen Designation

	// OnceEachTurn is "once each turn, you may pay ..." (Warped Space)
	// and "once during each of your turns" (Access Maze), #3174. The
	// limit belongs to the SOURCE OBJECT: a claim is recorded in
	// TurnTally.GrantedAltCostsClaimed under ObjectTallyKey(source,
	// epoch, offer key) as the spell becomes cast, so two such Rooms
	// give two claims, a Room that left and came back is a new object
	// with a fresh one (CR 400.7), and a cast that fails records
	// nothing. Flushed with the rest of the tally as a turn begins.
	OnceEachTurn bool

	// YourTurnOnly is "during each of your turns" (Access Maze): the
	// offer is made only while its controller is the active player.
	YourTurnOnly bool

	// LifeEqualToManaValue is Access Maze's "by paying life equal to its
	// mana value rather than paying its mana cost" (#3174), Bolas's
	// Citadel's price on a granted offer: the derived offer charges
	// AlternativeCost.Life equal to the mana value of the face being
	// cast, and no mana. {X} counts as 0 (CR 107.3b: the spell is cast
	// without paying its mana cost), and a spell with no mana cost (CR
	// 118.6) has mana value 0, so it costs no life. Offer.ManaCost, Life
	// and Energy must be empty.
	LifeEqualToManaValue bool
}

// gatedOn reports whether this static is switched on for `src` right
// now and still has its once-each-turn claim for `caster`: the door
// gate, the "your turn" clause and the per-object limit. Caller must
// hold g.mu.
func (gr GrantedAlternativeCost) gatedOn(g *Game, caster uuid.UUID, src *Card) bool {
	if gr.ActiveWhen.IsGate() && !gr.ActiveWhen.Active(*src) {
		return false
	}
	if gr.YourTurnOnly && g.activePlayerIDLocked() != caster {
		return false
	}
	return !gr.OnceEachTurn || g.TurnTally.GrantedAltCostsClaimed[gr.claimKey(src)] == 0
}

// claimKey is the TurnTally.GrantedAltCostsClaimed key of this
// static's claim on `src`, the object it is printed on.
func (gr GrantedAlternativeCost) claimKey(src *Card) string {
	return ObjectTallyKey(src.InstanceID, src.ObjectEpoch, gr.Offer.Key)
}

// noteGrantedOfferClaimedLocked records a once-each-turn granted offer
// as used, as the spell paid with it becomes cast (#3174). A no-op for
// an offer with no such limit. Caller must hold g.mu in write mode.
func (g *Game) noteGrantedOfferClaimedLocked(alt *AlternativeCost) {
	if alt == nil || alt.onceEachTurnKey == "" {
		return
	}
	if g.TurnTally.GrantedAltCostsClaimed == nil {
		g.TurnTally.GrantedAltCostsClaimed = map[string]int{}
	}
	g.TurnTally.GrantedAltCostsClaimed[alt.onceEachTurnKey]++
}

// lifeEqualToManaValue is the life a LifeEqualToManaValue offer charges
// for `card`, and whether it can be priced at all. A spell with {X} in
// its cost is cast for this offer without paying its mana cost, so X is
// 0 (CR 107.3b) and the mana value is the rest of the cost. A spell with
// no mana cost (CR 118.6, Ancestral Vision) may be cast through an
// alternative cost, and its mana value is 0 (CR 202.3), so it costs no
// life. A cost the engine cannot read gets no offer.
func lifeEqualToManaValue(card Card) (int, bool) {
	if HasNoManaCost(card) {
		return 0, true
	}
	if card.ManaCost == "" {
		return 0, false
	}
	pc, err := ParseCost(card.ManaCost)
	if err != nil {
		return 0, false
	}
	return pc.ManaValueWithX(0), true
}

// covers reports whether this static's offer reaches a cast of `card`
// by `caster`. The one reader of both caps, so the cast gate, the
// bot's move list and the view's picker, which all derive their offers
// through grantedAlternativeCostsLocked, cannot disagree about a
// computed one.
//
// Caller must hold g.mu.
func (gr GrantedAlternativeCost) covers(g *Game, caster uuid.UUID, card Card) bool {
	if !gr.Spells.Matches(card) {
		return false
	}
	if gr.MaxManaValue != nil && card.ManaValue() > *gr.MaxManaValue {
		return false
	}
	return gr.MaxManaValueFunc == nil || card.ManaValue() <= gr.MaxManaValueFunc(g, caster)
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
	if playerID == uuid.Nil || card.IsLand() {
		return nil
	}
	// #2709: a standing "the next <kind> spell you cast this turn can be
	// cast without paying its mana cost" promise (Apex Observatory) is
	// the same offer, held on the player rather than on a permanent.
	out := g.nextSpellFreeOffersLocked(playerID, card)
	if CatalogGrantedAlternativeCosts == nil || g.Battlefield == nil {
		return out
	}
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
			if gr.Offer.Key == "" || !gr.reaches(zone) || !gr.covers(g, playerID, card) || seen[gr.Offer.Key] {
				continue
			}
			// #3174: a door, a turn and a once-each-turn claim. Asked
			// before the key is marked seen, so a second source whose
			// claim is unspent still makes the offer when the first's is
			// spent.
			if !gr.gatedOn(g, playerID, src) {
				continue
			}
			if gr.PricedAtManaCost && !pricedAtManaCostOK(card) {
				continue
			}
			life, lifeOK := 0, true
			if gr.LifeEqualToManaValue {
				life, lifeOK = lifeEqualToManaValue(card)
			}
			if !lifeOK {
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
			if gr.PricedAtManaCost {
				// ADR 0135 PR 6: the price is the spell's own mana cost,
				// shown in the label ("Emerge {4}{G}{G}").
				offer.ManaCost = card.ManaCost
				offer.Label += " " + card.ManaCost
			}
			if gr.LifeEqualToManaValue {
				// #3174: Bolas's Citadel's price (CastPermission.
				// AlternativeCostFor), read off the face being cast.
				offer.ManaCost = ""
				offer.Life = life
				offer.Label += fmt.Sprintf(" (%d life)", offer.Life)
			}
			if gr.OnceEachTurn {
				offer.onceEachTurnKey = gr.claimKey(src)
			}
			if src.Name != "" {
				offer.Label += " (" + src.Name + ")"
			}
			out = append(out, &offer)
		}
	}
	return out
}

// GrantedAltCostNextFree is the key of the free offer a live
// NextSpellPromise.WithoutPayingManaCost makes (#2709). An on-disk
// identity, like every granted key: never renamed, never reused.
const GrantedAltCostNextFree = "granted-next-spell-free"

// nextSpellFreeOffersLocked is the one offer `playerID`'s live
// "without paying its mana cost" promises make to a cast of `card`: at
// most one, however many such promises match, since they are the same
// offer. Its label is the first matching promise's text.
//
// Caller must hold g.mu.
func (g *Game) nextSpellFreeOffersLocked(playerID uuid.UUID, card Card) []*AlternativeCost {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil
	}
	for _, s := range p.Statics {
		np := s.NextSpell
		if !np.WithoutPayingManaCost || !g.livePromise(s) || !np.Filter.Matches(card) {
			continue
		}
		label := "Cast it without paying its mana cost"
		if s.Label != "" {
			label += " (" + s.Label + ")"
		}
		return []*AlternativeCost{{Key: GrantedAltCostNextFree, Label: label, Granted: true}}
	}
	return nil
}

// pricedAtManaCostOK reports whether `card` has a mana cost a
// PricedAtManaCost offer can equal: one printed (CR 118.6: an object
// with no mana cost has none to pay) and parseable as a single cost.
func pricedAtManaCostOK(card Card) bool {
	if HasNoManaCost(card) || card.ManaCost == "" {
		return false
	}
	_, err := ParseCost(card.ManaCost)
	return err == nil
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
	if under != nil && (under.NamesPrice() || under.AltCostKey != "") {
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
	return a != nil && a.Condition == nil && a.Life == 0 && a.Energy == 0 && !a.AsThoughFlash && a.XCeiling == 0 && !a.PaysCards() &&
		a.Targets == nil && !a.ClearsTargets && !a.SacrificeOnEntry &&
		!a.ExileOnLeavingStack && !a.WarpExile && !a.Dash && !a.Blitz && a.EntersWithCounterName == "" &&
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
