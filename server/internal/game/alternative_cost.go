package game

import "github.com/google/uuid"

// alternative_cost.go — S22: a cost paid *instead of* a spell's mana
// cost (CR 118.9). The fourth kind of cost the engine models, after
// a spell's mana cost (S15), an activated ability's cost (S21
// sub-PR 2) and an additional cost (S21 sub-PR 5).
//
// The distinction from AdditionalCost is the whole point of the
// file. An additional cost is paid ALONGSIDE the mana cost; an
// alternative cost is paid INSTEAD of it. "Overload {6}{U}" does not
// mean "{1}{U} and also {6}{U}" — it means the spell costs {6}{U}
// and nothing else. Additional costs survive the swap, because CR
// 601.2f is evaluated independently of the cost chosen at 601.2b, so
// a card that charged both would still charge both.
//
// Three things an alternative cost may carry beyond the price,
// because the keywords that grant one rarely stop there:
//
//   - Overload rewrites the targeting clause out of existence
//     ("change 'target' to 'each'"), so paying it must CLEAR the
//     spell's TargetSpec rather than widen it — otherwise the
//     announce gate would still demand a target the spell no longer
//     has. That is ClearsTargets.
//   - Cleave rewrites the clause into a different one ("remove the
//     words in square brackets"), so paying it SWAPS IN another
//     TargetSpec. That is Targets.
//   - Evoke attaches a triggered ability to the permanent's entry
//     ("it's sacrificed when it enters"). That is SacrificeOnEntry.
//
// Modelled as a slice on the card rather than a single struct: a
// card can offer more than one (spree, and the modal-cost cards),
// and growing a struct into a slice later would churn every
// signature.
//
// Modelled as components rather than a parsed cost string, for the
// same reason AbilityCost and AdditionalCost are: the shapes are
// few, and a cost mini-language has to be maintained against a
// handful of cards.

// AlternativeCost is one "you may cast this spell for X rather than
// its mana cost" option.
type AlternativeCost struct {
	// Key is the stable wire identifier the caster names to claim
	// this cost — "overload", "evoke", "cleave". Rides
	// CastSpellParams.AlternativeCost, lands on StackItem.AltCost,
	// and is read back at resolution so a card whose text changes
	// with the cost can branch on it. Required, and unique among a
	// card's alternative costs.
	Key string

	// Label is the clause as printed ("Overload {4}{R}"), shown in
	// the client's picker so the prompt reads like the card rather
	// than like a schema.
	Label string

	// ManaCost is what the caster pays instead of the printed cost,
	// in the same Scryfall brace notation Card.ManaCost uses. Empty
	// means free ("without paying its mana cost") — ParseCost reads
	// an empty string as the zero cost, which is exactly right.
	//
	// Commander tax still applies on top: CR 903.8 adds {2} per
	// prior cast to whatever the cost is, alternative or not. So
	// effectiveCostLocked layers the tax after the swap, not before.
	ManaCost string

	// Targets replaces the card's target clause when this cost is
	// paid. Cleave's "remove the words in square brackets" turns
	// "counter target spell that wasn't cast from its owner's hand"
	// into "counter target spell" — a different, wider clause, not
	// an absent one. Nil leaves the printed clause alone.
	Targets *TargetSpec

	// ClearsTargets removes the target clause outright — overload's
	// "change 'target' in its text to 'each'". A cast that claims
	// such a cost and still sends targets is rejected rather than
	// ignored: the spell has none, and a client that thinks
	// otherwise is confused about which cost it is paying.
	ClearsTargets bool

	// Condition gates the OFFER — "IF YOU CONTROL A COMMANDER, you
	// may cast this spell without paying its mana cost" (Fierce
	// Guardianship), "IF YOU CONTROL A SWAMP, you may pay 4 life
	// rather than pay this spell's mana cost" (Snuff Out).
	//
	// Checked at announce and again in the view, so an offer the
	// caster cannot take is neither shown nor accepted. Nil means
	// unconditional, which is what overload, evoke and cleave are.
	//
	// Read-only, under g.mu. Added in S28.
	Condition func(g *Game, controller uuid.UUID) bool

	// Life is a "pay N life" component of the alternative cost (CR
	// 119.4) — Force of Will's 1 life, Snuff Out's 4.
	//
	// A COST, not a drawback: it is validated before anything is
	// paid, so a player below N life cannot claim the offer at all —
	// and #695 made that true of the OFFER as well as the payment.
	// AlternativeCostPayableLocked is the predicate; lifePayableBy is
	// the line, and since #1200 it also carries CR 119.8 — a player
	// whose life total can't change cannot pay any of it. (CR 119.4
	// lets a player pay life down to exactly zero, and the
	// state-based action kills them afterwards — that is a legal, if
	// unwise, Force of Will. The stale 118.4 citation here was the
	// #693 renumbering tail.)
	Life int

	// Energy is a "pay N {E}" component of the alternative cost (CR
	// 107.14, CR 118.9): Nissa, Worldsoul Speaker's eight, Primal
	// Prayers' one, and Amped Raptor's "an amount of {E} equal to its
	// mana value" (ADR 0129 §5, PR 4). Removed from the caster through
	// payEnergyLocked, the one path that pays energy, with the cast's
	// other non-mana costs.
	//
	// A COST, like Life: AlternativeCostPayableLocked withholds the
	// offer from a caster short of it (CR 118.3) and the announce
	// validator refuses the claim. Never waived: Cast anyway and the
	// permissive posture decide only the mana (ADR 0129 §4).
	Energy int

	// ExileFromHand is "exile a blue card from your hand" (Force of
	// Will) or "exile a white card from your hand" (Solitude's evoke
	// cost), as a spec matched against the caster's hand. The caster
	// names the card in CastSpellParams.AltCostIDs.
	//
	// The spell being cast is never a legal choice: CR 601.2a moves
	// it to the stack before costs are paid, so it is no longer in
	// hand. Force of Will cannot pitch itself.
	//
	// The count is the spec's Min, one when unset — the same reading
	// ExileFromGraveyard's N has. Commandeer's "exile two blue cards
	// from your hand" is a spec of Min 2 (#1745); every other pitch
	// names one card.
	ExileFromHand *TargetSpec

	// ReturnToHand is "return an Island you control to its owner's
	// hand" (Daze), matched against the caster's permanents. Also
	// named in CastSpellParams.AltCostIDs.
	ReturnToHand *TargetSpec

	// ExileFromGraveyard is escape's "Exile N other cards from your
	// graveyard" (CR 702.138a) — the half of the escape cost that is
	// not mana, and the reason escape is priced rather than merely
	// permitted. Added in S29.
	//
	// The COUNT is the spec's Min (which equals its Max): "exile five
	// other cards" is one five-slot payment, not five one-slot ones,
	// and the caster names all five in CastSpellParams.AltCostIDs.
	// The spell being cast is never among them — CR 601.2a has
	// already moved it to the stack, which is precisely what the
	// printed word "other" means, so nothing has to special-case it
	// beyond the castID guard every card component already carries.
	//
	// Matched against the CASTER's graveyard only ("your graveyard"),
	// the same way ExileFromHand is matched against their hand. That
	// it happens to be the zone the cast also comes out of is a
	// coincidence of escape rather than a rule: FromZone is the
	// place, this is the price, and the two are still independent.
	ExileFromGraveyard *TargetSpec

	// Sacrifice is a "sacrifice N <permanents>" component of the
	// price — Dread Return's "Flashback—Sacrifice three creatures",
	// Fireblast's "you may sacrifice two Mountains rather than pay this
	// spell's mana cost" (#1727). Added after the four card components
	// above, and the first of them that pays with a SACRIFICE: a
	// ReturnToHand with a graveyard destination would emit no
	// EventSacrifice, so Blood Artist would never see the three
	// creatures Dread Return ate.
	//
	// The clause is the SAME shape the additional cost's sacrifice and
	// the activated ability's SacrificeOther carry — a TargetSpec over
	// permanents whose count is its Min == Max, built by effects'
	// sacrificeSpec — and it is validated by the same
	// validateSacrificeCostLocked those two use: exactly N, all
	// distinct, each on the battlefield under the CASTER's control (CR
	// 701.21a) and each matching the clause. Matched, never targeted
	// (CR 601.2h), so hexproof does not hide your own creature from
	// your own flashback cost.
	//
	// The caster names the permanents in CastSpellParams.AltCostIDs,
	// not SacrificeIDs: they pay THIS offer, which vanishes when the
	// caster declines it, and SacrificeIDs is the additional cost's
	// list — the one PaidCost.Sacrificed and a per-sacrifice discount
	// (CostQuery.Sacrificing) count. Paid with the spell already on the
	// stack and as one simultaneous exit (payCostSacrificesLocked), so
	// every dies trigger lands ABOVE the spell, and a countered Dread
	// Return leaves the three creatures dead.
	//
	// A fixed count only: effects.Register refuses "sacrifice X" and
	// "any number" here, because no printed alternative cost has a
	// variable sacrifice that the announce path could price.
	Sacrifice *TargetSpec

	// DiscardFromHand is retrace's "discarding a land card in addition
	// to paying its other costs" (CR 702.81a, #2528): a card from the
	// caster's HAND, named in CastSpellParams.AltCostIDs and DISCARDED
	// rather than exiled, so a discard payoff (Waste Not, Mary Read and
	// Anne Bonny, Marauding Mako) sees it and a graveyard card (Wrenn
	// and Six's own land, Crucible) gets it. The spec's Min is the count,
	// one when unset; no printed cost discards more.
	//
	// It is an ALTERNATIVE-cost field carrying an ADDITIONAL cost's
	// rule, and the one place the model bends (ADR 0066, 2026-10-07
	// amendment). Retrace is not an alternative cost (CR 702.81a): the
	// mana cost is still paid. The offer is built to say so — ManaCost
	// is the card's printed cost, never empty — and claiming it is what
	// opens the graveyard, so the discard is owed exactly when the cast
	// is a retrace cast and never on the hand cast of the same card.
	// What the model gives up is combining retrace with another
	// alternative cost on one cast, which no card in the catalog can do.
	//
	// The spell being cast is never a legal discard: CR 601.2a has
	// already moved it to the stack, and it is not in hand anyway.
	//
	// ADR 0135 §2 (#2412) also uses it for a true alternative cost:
	// Snag's "You may discard a Forest card rather than pay this spell's
	// mana cost" is an offer with an empty ManaCost and this component
	// (effects.DiscardInstead). Foil's "an Island card and another card"
	// is two cards under a set rule (TargetSpec.EachOf), which the
	// validator and the payability check read from the hand.
	DiscardFromHand *TargetSpec

	// PayLabel is the picker's prompt copy for the card component —
	// ExileFromHand, ReturnToHand, ExileFromGraveyard, DiscardFromHand or Sacrifice ("a
	// blue card", "an Island you control", "three creatures"). Empty
	// falls back to Label.
	PayLabel string

	// SacrificeOnEntry is evoke's "it's sacrificed when it enters".
	// Modelled as what CR 702.74a says it is — a triggered ability —
	// rather than as an immediate sacrifice inside the resolution.
	// The difference is observable and is the entire reason to evoke
	// a Slithermuse: the sacrifice uses the stack, so opponents get
	// a window, and the creature's own leaves-the-battlefield
	// trigger goes on the stack above nothing and draws the cards.
	SacrificeOnEntry bool

	// FromZone binds this offer to one cast source zone (S29). The
	// zero value — the overwhelming majority — means "from hand",
	// which is where overload, evoke and cleave are paid.
	//
	// Flashback and escape set ZoneGraveyard, and that single field
	// is what makes them alternative costs rather than a new kind of
	// thing: "cast this from your graveyard for {2}{R}" is a price
	// plus a place. The binding cuts BOTH ways and both halves
	// matter. A cast out of the graveyard may not claim overload,
	// and a cast out of hand may not claim flashback — the second
	// being the one that would hand the player a cheaper Faithless
	// Looting for free.
	//
	// A card that declares a bound offer must also list the zone in
	// Spec.CastableZones; Register panics otherwise, because an
	// offer bound to a zone the card cannot be cast from is
	// unclaimable and the card file meant one or the other.
	FromZone ZoneKind

	// ExileOnLeavingStack is flashback's "if the flashback cost was
	// paid, exile this card instead of putting it anywhere else any
	// time it would leave the stack" (CR 702.34a).
	//
	// It is the half of flashback that keeps it from being infinite,
	// and it is a REPLACEMENT rather than an exile bolted onto the
	// resolution — the difference is observable on every path out of
	// the stack that isn't a resolution. A flashed-back spell that
	// fizzles is exiled. A flashed-back spell answered by Hinder is
	// exiled rather than shuffled into its owner's library, because
	// CR 614 replaces the counter's destination too.
	//
	// Escape does NOT set this: an escaped Uro exiles itself through
	// its own printed text, and an escaped Kroxa does not exile at
	// all. Flashback is the keyword that carries the clause.
	ExileOnLeavingStack bool

	// WarpExile is warp's "exile this permanent at the beginning of
	// the next end step, then you may cast it from exile on a later
	// turn" (CR 702.185a).
	//
	// The sibling of SacrificeOnEntry, and modelled the same way:
	// the cost attaches a clause to the permanent's ENTRY, and the
	// clause uses the ordinary machinery rather than a bespoke one.
	// Evoke queues a triggered ability; warp schedules a CR 603.7
	// delayed trigger, and the grant it leaves behind is the same
	// CastPermission airbend uses — CR 611.2b's "for as long as it
	// remains exiled" (Duration.WhileInZone) with a NotBeforeSeq
	// floor for the "on a later turn" clause.
	//
	// A warped creature is therefore a two-for-one paid in tempo:
	// the cheap body now, the real body later. Nothing about the
	// second cast is special — it is an ordinary cast from exile,
	// for the printed mana cost, through the same grant the impulse
	// button already renders.
	WarpExile bool

	// EntersWithCounterName / EntersWithCounterCount is "this
	// creature escapes with a +1/+1 counter on it" (CR 702.138c) —
	// the clause most escape creatures print directly under the
	// cost, and the reason an escaped Voracious Typhon is a 7/7
	// rather than the 4/4 in the corner. Added in S29.
	//
	// It hangs off the COST for the same reason SacrificeOnEntry and
	// WarpExile do: it happens only when that cost was paid, and the
	// resolving StackItem is the last place that fact is reachable.
	// A card-level declaration would have to be re-checked against
	// the cost anyway, and would fire on a Typhon reanimated out of
	// the graveyard, which never escaped anything.
	//
	// The counter is a NAME rather than a bare number because
	// "escapes with a flying counter" is printed too (Tizerus
	// Charger), even though "+1/+1" is the case that matters.
	//
	// Applied through the CR 614 entry pipeline — the same
	// ReplacementEvent.EntersWithCounters map the card-printed clause
	// writes (Hangarback Walker's X, Etched Oracle's sunburst; see
	// entry_counters.go) — rather than stapled on after the permanent
	// lands. So a creature escaping under Doubling Season
	// gets twice the counters (CR 616), and an ETB trigger already
	// sees them.
	EntersWithCounterName  string
	EntersWithCounterCount int

	// FaceDown is CR 708.4: paying this cost casts the card FACE
	// DOWN. Morph's "you may cast this card as a 2/2 face-down
	// creature spell for {3}" (CR 702.37b), megamorph's (CR 702.37b)
	// and disguise's (CR 702.168a) — nil for every other alternative
	// cost in the game.
	//
	// The sixth clause a keyword staples to a price, and the one that
	// changes what the SPELL IS rather than what it costs or what it
	// targets. `alt.FaceDown != nil` is the predicate CastSpell
	// branches on, once, and everything after that line reads the
	// CR 708.2 object because CatalogKey has gone silent for it
	// (ADR 0069 decision 4, ADR 0082 decision 2).
	//
	// It carries the price of turning the permanent back up as well,
	// and that is deliberate — see FaceDownCast.
	FaceDown *FaceDownCast

	// RequiresGrant makes this printed offer claimable only while a
	// live CastPermission for the card OBJECT carries the same Key —
	// miracle's "when you reveal this card this way, you may cast it
	// by paying [cost]" (CR 702.94a, #1665). The offer is the PRICE;
	// the permission the miracle trigger's resolution grants is the
	// RIGHT to pay it, and without it a Terminus in a hand is not
	// castable for {W}.
	//
	// Checked by validateCastPathLocked, which CastSpell and
	// CastOffersForLocked (the view's stamp and the bot's offer list)
	// both call, so an offer this gates is neither shown, enumerated
	// nor accepted. See miracle.go.
	RequiresGrant bool

	// CastsFace is the face this offer casts — disturb's "you may cast
	// this card TRANSFORMED from your graveyard" (CR 702.146a, ADR 0107
	// §4). Zero, the overwhelming majority, casts the face the caster
	// chose, exactly as before.
	//
	// The seventh clause a keyword staples to a price, and the second,
	// after FaceDown, that changes what the SPELL IS. The offer is
	// printed on the FRONT face, so it is read off the card as it sits
	// in its zone (CR 712.8a: a double-faced card in a graveyard has
	// only its front face's characteristics), and claiming it puts the
	// card on the stack back face up (CR 712.11a) with only the back
	// face's characteristics (CR 712.8c). CR 712.11d is the rule that
	// joins the two halves: the front face's ability "is also
	// considered when evaluating that spell to determine if it can be
	// cast", so the zone and the price come from the front face
	// (castOfferByKey, castPathKey) and every other announce gate — the
	// target clause, the timing, the cast restrictions — reads the back.
	//
	// faceForCastLocked turns the claim into the face, and refuses it
	// on a card that is not a `transform` card or has no such face.
	// effects.Register refuses it on a back-face entry, since CR
	// 702.146a puts disturb on the front face.
	CastsFace int

	// Granted marks an offer a battlefield static applies to the spell
	// rather than one the card prints (ADR 0118 §3, #2163: Jodah,
	// Archmage Eternal's {W}{U}{B}{R}{G}, Omniscience's free cast).
	// Set only on the copies grantedAlternativeCostsLocked derives, and
	// read by validateCastPathLocked, which lets a granted offer be
	// claimed only where the printed mana cost could be paid (CR
	// 118.9a). Never serialised: an offer is catalog data built per
	// query, and a claim reaches the stack as its key alone.
	Granted bool

	// AsThoughFlash is "if you cast a spell this way, you may cast it as
	// though it had flash" (Primal Prayers, ADR 0129 §5). CR 601.3c: an
	// effect that lets a spell be cast as though it had flash only if an
	// alternative cost is paid lets its caster BEGIN to cast it at
	// instant speed. So a claim of this offer opens the instant-speed
	// window and a cast of the same card for its printed cost does not.
	// Read by CastTimingForOfferOpenLocked, which CastSpell, the bot
	// enumerator and the view's per-offer timing stamp all call. A
	// per-player restriction (Teferi, Time Raveler) still closes it, as
	// for any other flash grant (CR 101.2).
	AsThoughFlash bool

	// Purpose is what the spell does when cast for this cost, where
	// that differs from the card's own (ADR 0126 §6): overload turns
	// Cyclonic Rift into a bounce sweep. Zero for a cost that leaves the
	// spell's effect alone; the card's Purpose then applies. Catalog
	// data for the view; the engine never reads it.
	Purpose Purpose
}

// CastFaceOf returns the card as a cast claiming this offer puts it on
// the stack: turned to CastsFace for a disturb offer (CR 712.11a), and
// unchanged for every other offer and for nil.
//
// The bot enumerator and the view stamp read the card through this so
// that the target clause, the timing and the cast gate they judge are
// the back face's — the same face CastSpell materialises before it
// asks them. Nil-safe.
func (a *AlternativeCost) CastFaceOf(c Card) Card {
	if a == nil || a.CastsFace == 0 || c.ActiveFace == a.CastsFace {
		return c
	}
	c.SetFace(a.CastsFace)
	return c
}

// cardComponent returns the card-shaped half of this cost: the spec
// candidates are matched against, the zone they are named out of,
// and how many the caster must name. (nil, "", 0) for a cost whose
// only components are mana and life.
//
// One accessor rather than three call sites' worth of if-ladders,
// because escape made the count vary: every pre-S29 component named
// exactly one card, and "exile five other cards from your graveyard"
// is the first that does not. Nil-safe.
func (a *AlternativeCost) cardComponent() (*TargetSpec, ZoneKind, int) {
	if a == nil {
		return nil, "", 0
	}
	switch {
	case a.ExileFromHand != nil:
		n := a.ExileFromHand.Min
		if n < 1 {
			n = 1
		}
		return a.ExileFromHand, ZoneHand, n
	case a.DiscardFromHand != nil:
		n := a.DiscardFromHand.Min
		if n < 1 {
			n = 1
		}
		return a.DiscardFromHand, ZoneHand, n
	case a.ReturnToHand != nil:
		return a.ReturnToHand, ZoneBattlefield, 1
	case a.ExileFromGraveyard != nil:
		n := a.ExileFromGraveyard.Min
		if n < 1 {
			n = 1
		}
		return a.ExileFromGraveyard, ZoneGraveyard, n
	case a.Sacrifice != nil:
		// #1727: the count every other sacrifice cost reads, off the
		// same clause shape (#747) — Register holds it to a fixed
		// count, so this is the printed N.
		return a.Sacrifice, ZoneBattlefield, SacrificeCostCount(a.Sacrifice)
	}
	return nil, "", 0
}

// Clears reports whether paying this cost deletes the spell's target
// clause. Nil-safe, so the cast path can ask without a guard.
func (a *AlternativeCost) Clears() bool {
	return a != nil && a.ClearsTargets
}

// CatalogAlternativeCosts is the catalog hook the effects package
// wires at init, mirroring CatalogAdditionalCost and CatalogModeSpec.
// Nil, or a nil return, means the card offers no alternative cost —
// which is nearly every card.
var CatalogAlternativeCosts func(oracleID string) []AlternativeCost

// AlternativeCostsFor returns the alternative costs a card offers,
// or nil.
func AlternativeCostsFor(oracleID string) []AlternativeCost {
	if CatalogAlternativeCosts == nil || oracleID == "" {
		return nil
	}
	return CatalogAlternativeCosts(oracleID)
}

// AlternativeCostByKey returns the named alternative cost for a card,
// or nil when the card offers none by that name. The empty key is
// how "I am paying the printed mana cost" is spelled on the wire, so
// it always answers nil.
func AlternativeCostByKey(oracleID, key string) *AlternativeCost {
	if key == "" {
		return nil
	}
	for _, ac := range AlternativeCostsFor(oracleID) {
		if ac.Key == key {
			out := ac
			return &out
		}
	}
	return nil
}

// validateAlternativeCost resolves the caster's claim at announce
// (CR 601.2b — the alternative cost is chosen before targets, and
// everything downstream is judged under it).
//
// A key the card doesn't offer is a rejection rather than a
// fall-back to the printed cost: silently charging full price for a
// cast the player thought was an overload is the worst of the
// available failures. A cost that deletes the target clause arriving
// with targets is rejected for the same reason — the client is
// confused about which cost it is paying.
//
// Returns (nil, nil) for the ordinary "paying the mana cost" case.
func validateAlternativeCost(oracleID, key string, targets []TargetRef) (*AlternativeCost, error) {
	if key == "" {
		return nil, nil
	}
	alt := AlternativeCostByKey(oracleID, key)
	if alt == nil {
		return nil, ErrInvalidParam
	}
	if alt.ClearsTargets && len(targets) > 0 {
		return nil, ErrInvalidParam
	}
	return alt, nil
}

// resolveAlternativeCostLocked is validateAlternativeCost widened by
// ADR 0066: the claim is judged against the card's own offers FIRST
// and against the offer a granted permission synthesises second.
//
// The order is the decision, not an implementation detail. A card
// that both prints and is granted the same key keeps its PRINTED
// cost: Deep Analysis flashed back under Past in Flames costs {1}{U}
// and 3 life, not {3}{U}. A grant that overrode the printed offer
// would make the card cheaper than it is, and the one direction a
// sandbox must never err in is the player's favour (#259).
//
// ADR 0118 §3: and the offers the caster's permanents grant to every
// spell they cast out of `zone` (Jodah, Omniscience) last of all,
// which is the order CastOffersForLocked lists them in. Their keys are
// namespaced, so none can shadow a printed or permission key.
//
// Caller must hold g.mu.
func (g *Game) resolveAlternativeCostLocked(caster uuid.UUID, card Card, zone ZoneKind, grant *CastPermission, key string, targets []TargetRef) (*AlternativeCost, error) {
	if key == "" {
		return nil, nil
	}
	if alt, err := validateAlternativeCost(castOfferKey(card), key, targets); err == nil {
		return alt, nil
	}
	// ADR 0107 §4, CR 712.11d: a disturb claim on a card the cast has
	// already turned to its back face is judged by the offer its FRONT
	// face prints. The back face's own entry has never heard of it.
	if alt := frontFaceCastOffer(card, key); alt != nil && !(alt.ClearsTargets && len(targets) > 0) {
		return alt, nil
	}
	granted := grant.AlternativeCostFor(card)
	if granted == nil || granted.Key != key {
		granted = g.grantedAlternativeCostByKeyLocked(caster, card, zone, key)
	}
	if granted == nil {
		return nil, ErrInvalidParam
	}
	if granted.ClearsTargets && len(targets) > 0 {
		return nil, ErrInvalidParam
	}
	return granted, nil
}

// PaysCards reports whether this cost has a component the caster
// must name a card for. Nil-safe.
func (a *AlternativeCost) PaysCards() bool {
	spec, _, _ := a.cardComponent()
	return spec != nil
}

// CardPaymentCount is how many cards the caster must name in
// CastSpellParams.AltCostIDs to pay this offer's card component: one
// for a pitch or a bounce, escape's N, zero for an offer whose only
// components are mana and life.
//
// Exported because the payment is exactly `want` and not "at least"
// (see validateAlternativeCostPaymentLocked), so a caller BUILDING an
// announcement — the bot enumerator, a client picker — has to know
// the number rather than discover it from a rejection. Nil-safe.
func (a *AlternativeCost) CardPaymentCount() int {
	_, _, n := a.cardComponent()
	return n
}

// Available reports whether a player may claim this offer right now
// — its Condition, and nothing else. The payment components are
// checked separately, at announce, because "you control no Swamp" is
// a reason to hide the offer while "you named the wrong card" is a
// reason to reject a cast.
//
// Nil-safe and nil-Condition-safe: an offer with no condition is
// always available. Caller must hold g.mu.
func (a *AlternativeCost) Available(g *Game, controller uuid.UUID) bool {
	if a == nil {
		return false
	}
	return a.Condition == nil || a.Condition(g, controller)
}

// lifePayableBy is CR 119.4 and CR 119.8 on this offer's life
// component: may `p` pay it right now?
//
// One line, and it is a FUNCTION rather than a comparison spelled out
// at each reader because it had been spelled out at only one of them
// (#695): announce refused the cast, the view showed the offer anyway,
// and a Force of Will at 0 life was a button that could only fail. It
// forwards to g.CanPayLifeLocked (life_lock.go), the one predicate
// every life-cost validator in the engine reads, so a player whose
// life total can't change (#1200) is refused Snuff Out's "pay 4 life"
// here as well as at the payment.
//
// Nil-safe on both sides. An offer with no life component is payable
// by anybody. Caller must hold g.mu.
func (g *Game) lifePayableBy(a *AlternativeCost, p *Player) bool {
	if a == nil {
		return false
	}
	return g.CanPayLifeLocked(p, a.Life)
}

// AlternativeCostPayableLocked reports whether a player could pay
// EVERY component of this offer if they claimed it right now — the
// one predicate behind "is this offer on the table" (#695).
//
// Four questions, in the order announce asks them:
//
//  1. the offer's own Condition (Available) — "if you control a
//     Swamp", "if you control a commander";
//  2. the life component (CR 119.4), which is why this function
//     exists: the view used to ask only (1) and show Snuff Out's
//     "pay 4 life" to a player at 3;
//  3. the energy component (CR 107.14, CR 118.3, ADR 0129 §5) —
//     Nissa, Worldsoul Speaker's eight {E};
//  4. the card component (CR 601.2b) — whether the caster's hand,
//     graveyard or battlefield holds as many cards matching the
//     clause as the cost demands. Force of Will with no other blue
//     card in hand, an escape cost with two cards left in the
//     graveyard.
//
// MANA IS DELIBERATELY NOT ASKED. CR 601.2g lets the caster activate
// mana abilities after the cost is settled, so "you cannot afford it
// yet" is not a reason to withhold the offer — it is what the
// auto-tapper and the strict-mana gate are for. Every other component
// is settled by the board at the moment the offer is read.
//
// `castID` is the spell being cast, which is never a legal payment:
// CR 601.2a has already moved it to the stack by the time the cost is
// paid, and it is what the printed "other" in escape's clause means.
//
// The three readers are the view's offer stamp, the bot enumerator
// and CastSpell's own validator, so a shown offer, an enumerated move
// and an accepted cast cannot disagree. Caller must hold g.mu.
func (g *Game) AlternativeCostPayableLocked(playerID, castID uuid.UUID, alt *AlternativeCost) bool {
	if alt == nil {
		return false
	}
	if !alt.Available(g, playerID) {
		return false
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return false
	}
	if !g.lifePayableBy(alt, p) {
		return false
	}
	// ADR 0129 §5: the energy component (CR 107.14, CR 118.3), through
	// the predicate the activation and the payment both read.
	if EnergyShortfall(p, alt.Energy) != nil {
		return false
	}
	spec, zone, want := alt.cardComponent()
	if spec == nil {
		return true
	}
	setRule := len(SacrificeSetKinds(spec)) > 0
	var pool []uuid.UUID
	have := 0
	for _, c := range g.zoneForAltCostLocked(p, zone).Cards {
		if c.InstanceID == castID {
			continue
		}
		if g.altCostCardOKLocked(p, spec, zone, c) {
			if setRule {
				pool = append(pool, c.InstanceID)
				continue
			}
			have++
			if have >= want {
				return true
			}
		}
	}
	// ADR 0135 §2: with a set rule, enough cards is not enough — Foil
	// held beside two non-Island cards can't be paid, so it isn't
	// offered (#695).
	return setRule && g.CostSetPaymentForEffect(spec, zone, pool) != nil
}

// altCostCardOKLocked is the per-card half of an alternative cost's
// card component: may THIS card pay it?
//
// The one predicate, shared by the payability check above and by
// validateAlternativeCostPaymentLocked's walk of the named IDs, so
// the offer the client is shown and the payment the engine accepts
// are decided by the same three lines.
//
// Not a target — a cost is not targeted (CR 601.2h), so hexproof and
// shroud do not apply and the spec is matched directly against the
// card rather than through the targeting gate. Caller must hold g.mu.
//
// #1727: a PERMANENT is matched by specMatchLocked(…, false), the
// non-targeting match every other battlefield cost component reads —
// validateSacrificeCostLocked, the return / tap-others / counter
// costs. For Daze's Island that is the CardOK call it always was; for
// the sacrifice component it makes the offer the view stamps and the
// candidates the enumerator pays from the same rule the announce
// validator then judges the named permanents by.
func (g *Game) altCostCardOKLocked(p *Player, spec *TargetSpec, zone ZoneKind, c Card) bool {
	if zone == ZoneBattlefield {
		// CR 701.21a for the sacrifice, "an Island YOU CONTROL" for the
		// bounce: a cost is paid with your own permanents.
		if c.Controller != p.ID {
			return false
		}
		return g.specMatchLocked(SourceChooser(p.ID), spec, TargetRef{Kind: TargetCard, ID: c.InstanceID}, false)
	}
	return spec.CardOK == nil || spec.CardOK(g, p.ID, c, zone)
}

// validateAlternativeCostPaymentLocked checks the non-mana half of a
// claimed alternative cost without paying any of it — the same
// validate-all-then-pay discipline the additional cost follows, so a
// rejected cast never leaves a half-paid cost behind.
//
// `castID` is the spell being cast, which is never a legal pitch: CR
// 601.2a has already moved it to the stack.
//
// An offer whose Condition is false is rejected here rather than
// silently downgraded to the printed cost, for the reason
// validateAlternativeCost gives about unknown keys: charging full
// price for a cast the player thought was free is the worst available
// failure.
//
// Caller must hold g.mu.
func (g *Game) validateAlternativeCostPaymentLocked(playerID, castID uuid.UUID, alt *AlternativeCost, ids []uuid.UUID) error {
	if alt == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if !alt.Available(g, playerID) {
		return ErrInvalidParam
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// CR 119.4 and CR 119.8, through the same one-line predicate the
	// view's offer stamp and the bot enumerator read (#695), so an
	// offer the client can see is one this validator will accept.
	if !g.lifePayableBy(alt, p) {
		return ErrInvalidParam
	}
	// ADR 0129 §5: "Not enough energy (have 2, need 8)", the refusal
	// an activation gives (CR 118.3).
	if err := EnergyShortfall(p, alt.Energy); err != nil {
		return err
	}
	spec, zone, want := alt.cardComponent()
	if spec == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if alt.Sacrifice != nil {
		// #1727: the one sacrifice validator the additional cost, the
		// activated abilities and the mana abilities share — exactly N
		// (SacrificeCountLegal), each named once, each on the
		// battlefield under the caster's control (CR 701.21a) and each
		// matching the clause, with nothing moved on any failure. The
		// spell being cast is on the stack (CR 601.2a), so it cannot be
		// among them; there is no X to announce for a fixed clause.
		_, err := g.validateSacrificeCostLocked(playerID, castID, AbilityCost{SacrificeOther: alt.Sacrifice}, ids, 0)
		return err
	}
	// Exactly `want`, not "at least": escape's five is a price, and
	// a caster who named four has not paid it while one who named
	// six has overpaid by a card the cost never asked for.
	if len(ids) != want {
		return ErrInvalidParam
	}
	// Distinctness is the other half of the count. Without it a
	// five-card escape cost could be paid by naming the same card
	// five times, which is the cheapest possible Uro and not a cost
	// at all.
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if id == castID || seen[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		c, ok := g.cardInZoneLocked(g.zoneForAltCostLocked(p, zone), id)
		if !ok {
			return ErrCardNotFound
		}
		// The same per-card predicate AlternativeCostPayableLocked
		// counts candidates with, so "the client was offered this"
		// and "the engine accepts this" are one rule (#695).
		if !g.altCostCardOKLocked(p, spec, zone, c) {
			return ErrInvalidParam
		}
	}
	// ADR 0135 §2: a set rule over the picks (Foil's "an Island card
	// and another card"). Each card passed the clause's union predicate
	// above; the SET must still fill every entry, one card each. Two
	// non-Island cards are not an Island card and another card.
	if !g.costSetSatisfiedLocked(spec, zone, ids) {
		return ErrInvalidParam
	}
	return nil
}

// AltCostSetPaymentLocked searches `candidates` (in the caller's
// preferred order) for ONE payment of an offer whose card component has
// a set rule (TargetSpec.EachOf, ADR 0135 §2: Foil's "an Island card
// and another card"). ok is false when the component has no set rule,
// and then the caller pays the first N as before; with a rule, a nil
// payment means the candidates cannot fill it. Caller must hold g.mu.
func (g *Game) AltCostSetPaymentLocked(alt *AlternativeCost, candidates []uuid.UUID) (pay []uuid.UUID, ok bool) {
	spec, zone, _ := alt.cardComponent()
	if len(SacrificeSetKinds(spec)) == 0 {
		return nil, false
	}
	return g.CostSetPaymentForEffect(spec, zone, candidates), true
}

// zoneForAltCostLocked picks the zone an alternative cost's card
// component is paid from: the caster's own hand or graveyard, or the
// shared battlefield.
//
// Hand and graveyard are per-player zones, so resolving them through
// the caster's Player is what enforces "your hand" / "your
// graveyard" — an escape cost can no more exile an opponent's
// graveyard than a Force of Will can pitch from one.
func (g *Game) zoneForAltCostLocked(p *Player, kind ZoneKind) *Zone {
	switch kind {
	case ZoneHand:
		return p.Hand
	case ZoneGraveyard:
		return p.Graveyard
	}
	return g.Battlefield
}

// AltCostCandidatesLocked lists the cards `playerID` may name to the
// CARD-shaped half of `alt` right now — Force of Will's blue card in
// hand, Daze's Island, escape's other cards in the graveyard — in
// zone order, with the spell being cast excluded.
//
// It is AlternativeCostPayableLocked's sibling: that one counts the
// candidates and stops at `want`, this one names them, and both walk
// `zoneForAltCostLocked` through the SAME per-card predicate
// (altCostCardOKLocked) the announce validator judges the caster's
// named IDs with. One rule, three readers, so a payment the bot
// enumerates is a payment the engine accepts.
//
// A copy built out of specCandidatesLocked would be close and not
// equal: that walk visits EVERY seat's graveyard and leans on the
// spec's own "you own it" predicate, which is a second statement of
// "your graveyard" rather than the same one.
//
// `castID` is the spell being cast, which is never a legal payment:
// CR 601.2a has already moved it to the stack by the time the cost is
// paid, and that is precisely what escape's printed "other" means.
//
// Nil for an offer whose only components are mana and life, and for
// nil. Caller must hold g.mu.
func (g *Game) AltCostCandidatesLocked(playerID, castID uuid.UUID, alt *AlternativeCost) []uuid.UUID {
	spec, kind, _ := alt.cardComponent()
	if spec == nil {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil
	}
	z := g.zoneForAltCostLocked(p, kind)
	if z == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(z.Cards))
	for i := range z.Cards {
		c := z.Cards[i]
		if c.InstanceID == castID {
			continue
		}
		if !g.altCostCardOKLocked(p, spec, kind, c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	if alt.Sacrifice != nil {
		// #1727: a sacrifice payment is offered in the order every
		// other sacrifice cost is (#747) — tokens first, then the
		// cheapest — so the bot pays what the client's "Choose for me"
		// would, before a policy re-sorts it.
		return g.SacrificePaymentOrderForEffect(out, uuid.Nil)
	}
	return out
}

// payAlternativeCostLocked pays the non-mana components of a claimed
// alternative cost. Call only after
// validateAlternativeCostPaymentLocked has passed and after the spell
// itself has reached the stack (CR 601.2a before 601.2h), so anything
// watching the exile, the life payment or the bounce triggers ABOVE
// the spell and resolves first.
//
// That window is the whole reason this is not folded into the
// resolution: a Daze that returned its Island on resolution would
// hand the opponent a turn of information, and an exiled Force of
// Will pitch that never happened because the spell was countered
// would make Force of Will free.
//
// Caller must hold g.mu.
func (g *Game) payAlternativeCostLocked(playerID, castID uuid.UUID, alt *AlternativeCost, ids []uuid.UUID, answers map[uuid.UUID]bool) error {
	if alt == nil {
		return nil
	}
	// ADR 0129 §5: the energy, through the one path that pays it (CR
	// 107.14), naming the spell as the source so the log reads "pays 8
	// energy (Craterhoof Behemoth)". No replacement window, as for every
	// energy payment.
	if err := g.payEnergyLocked(playerID, alt.Energy, castID); err != nil {
		return err
	}
	// #793: the cost path. Snuff Out's "pay 4 life" is a cost, so it
	// runs the CR 614 window (CR 119.4 — paying life is losing life)
	// but settles without a CR 616 prompt: a free spell whose payment
	// paused would be on the stack with nothing paid for it.
	if alt.Life > 0 {
		if err := g.PayLifeForEffect(uuid.Nil, playerID, alt.Life); err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		return nil
	}
	// #1397: each card's move carries the CR 903.9 answer its owner
	// gave before the cast was paid for (cost_commander_choice.go), so
	// a pitched, returned or escaped commander no longer pauses the
	// payment half way through it. #1420: MustSettleNow extends that
	// indivisible-payment rule to CR 616 ordering prompts; a spell may
	// not sit on the stack while the card paying for it is still in its
	// old zone.
	move := func(id uuid.UUID, dst ZoneKind) error {
		_, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:          id,
			Dst:             dst,
			MustSettleNow:   true,
			commanderAnswer: commanderAnswerFor(answers, id),
		})
		return err
	}
	switch {
	case alt.ExileFromHand != nil:
		// One card for every pitch but Commandeer's two (#1745).
		for _, id := range ids {
			if err := move(id, ZoneExile); err != nil {
				return err
			}
		}
	case alt.DiscardFromHand != nil:
		// Retrace (#2528, CR 702.81a): a discard, paid as a COST, so
		// with the spell already on the stack a discard trigger lands
		// above it, and a countered Flame Jab does not hand the land
		// back. Through the one discard helper's cost path (it may not
		// pause — CR 601.2h) with the CR 903.9 answers the cast
		// collected up front.
		return g.discardCardsLocked(playerID, ids, discardOptions{
			cause:            DiscardCauseCost,
			commanderAnswers: answers,
		})
	case alt.ReturnToHand != nil:
		return move(ids[0], ZoneHand)
	case alt.ExileFromGraveyard != nil:
		// Escape's exiles are a cost, so they happen with the spell
		// already on the stack — which is what makes an escaped Uro
		// visible to anything watching the graveyard shrink, and
		// what makes the cards stay exiled when the spell is
		// countered.
		for _, id := range ids {
			if err := move(id, ZoneExile); err != nil {
				return err
			}
		}
	case alt.Sacrifice != nil:
		// #1727: the additional cost's payer, verbatim. Each permanent
		// is SACRIFICED — EventSacrifice, then its own route to its
		// owner's graveyard — and the N leave as one simultaneous exit,
		// so a Blood Artist among them sees every death (CR 603.10a).
		// With the spell already on the stack, so the dies triggers
		// resolve first; and nothing gives them back if it is
		// countered. The CR 903.9 answers ride along as they do for the
		// moves above (#1397).
		return g.payCostSacrificesLocked(ids, answers)
	}
	return nil
}

// altCostExilesFromStack reports whether the cost this spell was
// cast for replaces every stack-exit destination with exile — CR
// 702.34a's flashback clause. A spell cast for its printed cost
// always answers false, even on a card that offers flashback.
//
// The StackItem is the authority, not the catalog (ADR 0066, CR
// 400.7g): a Snapcaster'd Brainstorm was cast for a flashback cost
// the catalog has never heard of, and by the time it leaves the stack
// the permission that granted it may well be gone. So the announce
// path stamps the fact on the item and this reads it back, falling
// back to the catalog for a printed keyword — which is also what
// makes a game restored from a snapshot written before this field
// behave exactly as it did.
func altCostExilesFromStack(card Card, item *StackItem) bool {
	if item == nil {
		return false
	}
	if item.AltCostExiles {
		return true
	}
	alt := AlternativeCostByKey(CatalogKey(card), item.AltCost)
	return alt != nil && alt.ExileOnLeavingStack
}

// TargetSpecUnderAlternativeCost applies an alternative cost's
// rewrite of the target clause: overload deletes it, cleave swaps
// it, anything else leaves the printed clause alone. Used at
// announce (CR 601.2c), again at resolution (CR 608.2b) so both
// checks judge the spell under the text it was actually cast with,
// and by the view layer so the client's picker shows the legal set
// each offer would produce.
func TargetSpecUnderAlternativeCost(base *TargetSpec, alt *AlternativeCost) *TargetSpec {
	if alt == nil {
		return base
	}
	if alt.ClearsTargets {
		return nil
	}
	if alt.Targets != nil {
		return alt.Targets
	}
	return base
}

// queueAltCostEntryTriggerLocked applies the clauses an alternative
// cost attaches to the permanent's ENTRY: evoke's "it's sacrificed
// when it enters" (CR 702.74a) and warp's "exile this at the
// beginning of the next end step, then you may cast it from exile on
// a later turn" (CR 702.185a).
//
// Called from the resolution path right after the permanent lands
// and its ETB hook fires, which is the last moment the StackItem —
// and so the cost that was paid — is still in hand.
//
// The two clauses use different machinery, and deliberately: evoke's
// sacrifice happens NOW and uses the stack, so it is an ordinary
// triggered ability; warp's exile happens at a later step, so it is
// a CR 603.7 delayed trigger. Neither gets a bespoke loop.
//
// No-op for a spell cast for its mana cost, and for an alternative
// cost that carries neither clause. Caller must hold g.mu.
func (g *Game) queueAltCostEntryTriggerLocked(card Card, item *StackItem) {
	if item == nil || item.AltCost == "" {
		return
	}
	alt := AlternativeCostByKey(CatalogKey(card), item.AltCost)
	if alt == nil {
		return
	}
	if alt.WarpExile {
		g.scheduleWarpExileLocked(card, item, alt)
	}
	if !alt.SacrificeOnEntry {
		return
	}
	label := card.Name + " — " + alt.Label + ", sacrifice it"
	pass := g.newHarvestPassLocked(Event{Kind: EventETB, Actor: item.Controller, CardID: card.InstanceID})
	// ADR 0041 P9 (#1497, tier 4): evoke's sacrifice has no catalog
	// row, so its item is keyed directly.
	//
	// It is the permanent's own enters-the-battlefield ability (CR
	// 702.74a), so a trigger suppressor stops it as it stops any other
	// (#1735). A creature evoked under Torpor Orb stays on the battlefield.
	controller, cardID := item.Controller, card.InstanceID
	g.harvestMatchLocked(&pass, card, card.Effective(), TriggeredAbility{Key: label, Build: func(_ Event, _ *Card, _ Characteristic, _ *Game) *StackItem {
		return &StackItem{
			Kind: StackItemTriggered, Controller: controller, Owner: controller,
			SourceCardID: cardID, Label: label,
			Body:   evokeSacrificeBody.Key(),
			Effect: bodyEffect(evokeSacrificeBody.Key(), EffectParams{}),
		}
	}}, triggerOfPermanent)
}

// evokeSacrificeBody is "evoke/sacrifice" (ADR 0041 P9, #1497, tier 4):
// the permanent may already have left — the trigger sat on the stack
// and anyone could answer it. Nothing to sacrifice is not an error;
// the ability simply does as much as it can (CR 608.2c). Left and COME
// BACK is the same answer (#1432, CR 400.7): an evoked creature
// flickered in response is a new object, cast for nothing in
// particular, and is not sacrificed.
//
// Assigned in init: a var initialiser would be an initialisation
// cycle through the exit primitives, the same shape
// madnessGraveyardBody's comment explains.
var evokeSacrificeBody BodyRef

func init() {
	evokeSacrificeBody = SimpleDelayedBody("evoke/sacrifice", func(g *Game, it *StackItem) error {
		if g.AbilitySourceGoneForEffect(it) {
			return nil
		}
		return g.sacrificePermanentLocked(it.SourceCardID)
	})
}

// applyAltCostEntryCountersLocked folds "this creature escapes with
// a +1/+1 counter on it" (CR 702.138c) into the permanent's ENTRY
// event, before the CR 614 pipeline runs.
//
// The sibling of queueAltCostEntryTriggerLocked and the reason both
// exist: an alternative cost can attach a clause to the permanent's
// entry, and the clause should use whichever existing machinery
// matches its timing. Evoke's sacrifice is a triggered ability, warp's
// exile is a delayed trigger, and "escapes with counters" is a
// replacement — so it rides the same EntersWithCounters map the CARD's
// own printed "enters with X +1/+1 counters" writes one line later
// (applyCastEntryCountersLocked, entry_counters.go), rather than an
// AddCounter after the permanent lands.
//
// The difference is observable in both directions. Under Doubling
// Season a Typhon escaping with three counters gets six, because the
// counters go on through the counter-replacement pipeline; and the
// permanent's own ETB trigger already sees them, because the map is
// drained before EventETB fires.
//
// No-op for a spell cast for its mana cost, and for an alternative
// cost that declares no counters. Caller must hold g.mu.
func (g *Game) applyAltCostEntryCountersLocked(ev *ReplacementEvent, card Card, item *StackItem) {
	if ev == nil || item == nil || item.AltCost == "" {
		return
	}
	alt := AlternativeCostByKey(CatalogKey(card), item.AltCost)
	if alt == nil {
		return
	}
	ev.AddCounterAtETB(alt.EntersWithCounterName, alt.EntersWithCounterCount)
}

// scheduleWarpExileLocked schedules warp's "exile this permanent at
// the beginning of the next end step, then you may cast it from
// exile on a later turn" (CR 702.185a).
//
// The "later turn" floor is computed HERE, at schedule time, rather
// than inside the effect. Both readings give the same answer for a
// creature warped at sorcery speed on its controller's own turn —
// which is every printed warp card — but computing it at schedule
// time is the reading that survives a warped creature with flash:
// the floor is one past the turn the spell was cast, not one past
// whichever turn the end step happened to arrive in.
//
// Caller must hold g.mu.
func (g *Game) scheduleWarpExileLocked(card Card, item *StackItem, alt *AlternativeCost) {
	notBefore := g.Turn.Seq + 1
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: card.InstanceID,
		Label:        card.Name + " — " + alt.Label + ", exile it",
		At:           StepEnd,
		Cards:        []uuid.UUID{card.InstanceID},
		Body:         warpExileBody,
		// The "later turn" floor, computed now (see above).
		Params: EffectParams{Amount: notBefore},
	})
}

// warpExileBody is the delayed trigger's body, a registered key (ADR
// 0041 phase 3, #1497): the floor it used to capture is Params.Amount.
// warpExileBody is assigned in init: a var initialiser would be an
// initialisation cycle through the exit primitives.
var warpExileBody BodyRef

func init() { warpExileBody = DelayedBody("warp/exile", warpExile) }

func warpExile(g *Game, it *StackItem, p EffectParams) error {
	for _, t := range it.Targets {
		if t.Kind != TargetCard {
			continue
		}
		// The permanent may have died, been exiled by something else,
		// or bounced since the warp. Nothing to exile is not an error —
		// CR 608.2c — and the grant simply never lands.
		if g.controllerOfBattlefieldCardLocked(t.ID) == uuid.Nil {
			continue
		}
		if err := g.ExileCardWithPermissionForEffect(t.ID, CastPermission{
			// Zero Player means "the card's owner", which is what
			// warp says: YOU cast it later, and the warping player
			// owns the card.
			Duration:     WhileInZoneDuration(),
			NotBeforeSeq: p.Amount,
			Label:        "Warp — cast it from exile",
		}); err != nil {
			return err
		}
	}
	return nil
}

// sharesAnID reports whether any ID appears in both lists — the CR
// 118.3 check that one object is not named to pay two cost components
// (#1727: an alternative cost's sacrifice and an additional cost's).
func sharesAnID(a, b []uuid.UUID) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	seen := make(map[uuid.UUID]bool, len(a))
	for _, id := range a {
		seen[id] = true
	}
	for _, id := range b {
		if seen[id] {
			return true
		}
	}
	return false
}
