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
	// paid, so a player below N life cannot claim the offer at all.
	// (CR 118.4 lets a player pay life down to exactly zero, and the
	// state-based action kills them afterwards — that is a legal, if
	// unwise, Force of Will.)
	Life int

	// ExileFromHand is "exile a blue card from your hand" (Force of
	// Will) or "exile a white card from your hand" (Solitude's evoke
	// cost), as a spec matched against the caster's hand. The caster
	// names the card in CastSpellParams.AltCostIDs.
	//
	// The spell being cast is never a legal choice: CR 601.2a moves
	// it to the stack before costs are paid, so it is no longer in
	// hand. Force of Will cannot pitch itself.
	ExileFromHand *TargetSpec

	// ReturnToHand is "return an Island you control to its owner's
	// hand" (Daze), matched against the caster's permanents. Also
	// named in CastSpellParams.AltCostIDs.
	ReturnToHand *TargetSpec

	// ExileFromGraveyard is escape's "Exile N other cards from your
	// graveyard" (CR 702.144a) — the half of the escape cost that is
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

	// PayLabel is the picker's prompt copy for the ExileFromHand /
	// ReturnToHand component ("a blue card", "an Island you
	// control"). Empty falls back to Label.
	PayLabel string

	// SacrificeOnEntry is evoke's "it's sacrificed when it enters".
	// Modelled as what CR 702.74b says it is — a triggered ability —
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
	// turn" (CR 702.183a).
	//
	// The sibling of SacrificeOnEntry, and modelled the same way:
	// the cost attaches a clause to the permanent's ENTRY, and the
	// clause uses the ordinary machinery rather than a bespoke one.
	// Evoke queues a triggered ability; warp schedules a CR 603.7
	// delayed trigger, and the grant it leaves behind is the same
	// ExilePlayPermission airbend uses — unbounded (the window is
	// "for as long as it remains exiled") with a NotBeforeTurn floor
	// for the "on a later turn" clause.
	//
	// A warped creature is therefore a two-for-one paid in tempo:
	// the cheap body now, the real body later. Nothing about the
	// second cast is special — it is an ordinary cast from exile,
	// for the printed mana cost, through the same grant the impulse
	// button already renders.
	WarpExile bool

	// EntersWithCounterName / EntersWithCounterCount is "this
	// creature escapes with a +1/+1 counter on it" (CR 702.144c) —
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
	// ReplacementEvent.EntersWithCounters map Hangarback Walker's own
	// self-replacement writes — rather than stapled on after the
	// permanent lands. So a creature escaping under Doubling Season
	// gets twice the counters (CR 616), and an ETB trigger already
	// sees them.
	EntersWithCounterName  string
	EntersWithCounterCount int
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
		return a.ExileFromHand, ZoneHand, 1
	case a.ReturnToHand != nil:
		return a.ReturnToHand, ZoneBattlefield, 1
	case a.ExileFromGraveyard != nil:
		n := a.ExileFromGraveyard.Min
		if n < 1 {
			n = 1
		}
		return a.ExileFromGraveyard, ZoneGraveyard, n
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

// PaysCards reports whether this cost has a component the caster
// must name a card for. Nil-safe.
func (a *AlternativeCost) PaysCards() bool {
	spec, _, _ := a.cardComponent()
	return spec != nil
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
	if alt.Life > 0 && p.Life < alt.Life {
		return ErrInvalidParam
	}
	spec, zone, want := alt.cardComponent()
	if spec == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
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
		// Not a target — a cost is not targeted (CR 601.2h), so
		// hexproof and shroud do not apply and the spec is matched
		// directly against the card rather than through the
		// targeting gate.
		c, ok := g.cardInZoneLocked(g.zoneForAltCostLocked(p, zone), id)
		if !ok {
			return ErrCardNotFound
		}
		if zone == ZoneBattlefield && c.Controller != playerID {
			return ErrInvalidParam
		}
		if spec.CardOK != nil && !spec.CardOK(g, playerID, c, zone) {
			return ErrInvalidParam
		}
	}
	return nil
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
func (g *Game) payAlternativeCostLocked(playerID uuid.UUID, alt *AlternativeCost, ids []uuid.UUID) error {
	if alt == nil {
		return nil
	}
	if alt.Life > 0 {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, playerID, -alt.Life); err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		return nil
	}
	switch {
	case alt.ExileFromHand != nil:
		return g.ExileCardForEffect(ids[0])
	case alt.ReturnToHand != nil:
		return g.BounceToHandForEffect(ids[0])
	case alt.ExileFromGraveyard != nil:
		// Escape's exiles are a cost, so they happen with the spell
		// already on the stack — which is what makes an escaped Uro
		// visible to anything watching the graveyard shrink, and
		// what makes the cards stay exiled when the spell is
		// countered.
		for _, id := range ids {
			if err := g.ExileCardForEffect(id); err != nil {
				return err
			}
		}
	}
	return nil
}

// alternativeCostString is the "pay" half: the mana cost a cast
// actually owes. The swap is total — nothing adds the printed cost
// back, which is the difference between this file and
// additional_cost.go.
func alternativeCostString(card Card, key string) string {
	if alt := AlternativeCostByKey(CatalogKey(card), key); alt != nil {
		return alt.ManaCost
	}
	return card.ManaCost
}

// altCostExilesFromStack reports whether the cost this spell was
// cast for replaces every stack-exit destination with exile — CR
// 702.34a's flashback clause. Reads the key off the StackItem, so a
// spell cast for its printed cost always answers false even on a
// card that offers flashback.
func altCostExilesFromStack(card Card, altCostKey string) bool {
	alt := AlternativeCostByKey(CatalogKey(card), altCostKey)
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
// when it enters" (CR 702.74b) and warp's "exile this at the
// beginning of the next end step, then you may cast it from exile on
// a later turn" (CR 702.183a).
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
	g.queueHarvestedTriggerLocked(&StackItem{
		Kind:         StackItemTriggered,
		Controller:   item.Controller,
		Owner:        item.Controller,
		SourceCardID: card.InstanceID,
		Label:        card.Name + " — " + alt.Label + ", sacrifice it",
		Effect: func(g *Game, it *StackItem) error {
			// The permanent may already have left — the trigger sat
			// on the stack and anyone could answer it. Nothing to
			// sacrifice is not an error; the ability simply does as
			// much as it can (CR 608.2c).
			if g.controllerOfBattlefieldCardLocked(it.SourceCardID) == uuid.Nil {
				return nil
			}
			return g.sacrificePermanentLocked(it.SourceCardID)
		},
	})
}

// applyAltCostEntryCountersLocked folds "this creature escapes with
// a +1/+1 counter on it" (CR 702.144c) into the permanent's ENTRY
// event, before the CR 614 pipeline runs.
//
// The sibling of queueAltCostEntryTriggerLocked and the reason both
// exist: an alternative cost can attach a clause to the permanent's
// entry, and the clause should use whichever existing machinery
// matches its timing. Evoke's sacrifice is a triggered ability, warp's
// exile is a delayed trigger, and "escapes with counters" is a
// replacement — so it rides the same EntersWithCounters map Hangarback
// Walker writes, rather than an AddCounter after the permanent lands.
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
// exile on a later turn" (CR 702.183a).
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
	notBefore := g.Turn.Number + 1
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: card.InstanceID,
		Label:        card.Name + " — " + alt.Label + ", exile it",
		At:           StepEnd,
		Cards:        []uuid.UUID{card.InstanceID},
		Effect: func(g *Game, it *StackItem) error {
			for _, t := range it.Targets {
				if t.Kind != TargetCard {
					continue
				}
				// The permanent may have died, been exiled by
				// something else, or bounced since the warp. Nothing
				// to exile is not an error — CR 608.2c — and the
				// grant simply never lands.
				if g.controllerOfBattlefieldCardLocked(t.ID) == uuid.Nil {
					continue
				}
				if err := g.ExileCardWithPermissionForEffect(t.ID, ExilePlayPermission{
					// Zero Player means "the card's owner", which is
					// what warp says: YOU cast it later, and the
					// warping player owns the card.
					WhileExiled:   true,
					NotBeforeTurn: notBefore,
				}); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
