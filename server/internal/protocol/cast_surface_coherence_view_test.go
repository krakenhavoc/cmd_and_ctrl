package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cast_surface_coherence_view_test.go — #1015 and #1012. The view and
// the bot enumerator used to answer "which prices may this cast claim"
// with two different functions, and the view's answer was wrong in
// three ways at once:
//
//   - it marked a cast surface on a card whose only zone-bound offer
//     was unpayable, so an escape card in a graveyard too small to pay
//     for it rendered a button the announce path refused with
//     ErrCastCostRequired (#1015);
//   - it stamped the GRANTED offer and then overwrote the whole slice
//     with the printed set, so a card that both prints and is granted
//     a price showed only the printed one (#1012);
//   - it said nothing at all about whether the PRINTED cost was
//     claimable from the zone, leaving the client to infer it from the
//     shape of the offer list (#1012).
//
// All three are now one read of game.CastOffersForLocked, which is the
// list the enumerator walks and the list CastSpell validates against.

// withCastableZones stubs the per-card cast-zone declaration for a set
// of oracle IDs, chaining to whatever the catalog already answered.
func withCastableZones(t *testing.T, zones map[string][]game.ZoneKind) {
	t.Helper()
	prev := game.CatalogCastableZones
	game.CatalogCastableZones = func(id string) []game.ZoneKind {
		if z, ok := zones[id]; ok {
			return z
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastableZones = prev })
}

// withAltCostsByOracle is withAltCosts for more than one oracle ID.
func withAltCostsByOracle(t *testing.T, costs map[string][]game.AlternativeCost) {
	t.Helper()
	prev := game.CatalogAlternativeCosts
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if c, ok := costs[id]; ok {
			return c
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogAlternativeCosts = prev })
}

// escapeCost is the offer #1015 is about: claimable only from the
// graveyard, and only when `n` OTHER cards are sitting under it
// (CR 702.138a).
func escapeCost(mana string, n int) game.AlternativeCost {
	return game.AlternativeCost{
		Key: "escape", Label: "Escape — " + mana + ", Exile other cards from your graveyard",
		ManaCost: mana, FromZone: game.ZoneGraveyard,
		PayLabel: "other cards in your graveyard",
		ExileFromGraveyard: &game.TargetSpec{
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Min:   n, Max: n,
		},
	}
}

// graveyardCard seeds one known card into a seat's graveyard.
func graveyardCard(p *game.Player, name, oracle string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{2}{R}"
	c.OracleID = oracle
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	id := c.InstanceID
	p.Graveyard.PushTop(c)
	return id
}

// cardInSeatZone finds one card in one projected zone.
func cardInSeatZone(t *testing.T, z ZoneView, id uuid.UUID) *CardView {
	t.Helper()
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id.String() {
			return &z.Cards[i]
		}
	}
	t.Fatalf("card %s missing from the %s view", id, z.Kind)
	return nil
}

func keysOf(offers []AlternativeCostView) []string {
	out := make([]string, 0, len(offers))
	for _, o := range offers {
		out = append(out, o.Key)
	}
	sort.Strings(out)
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- #1015 ----------------------------------------------------------

// The canonical case. A card whose ONLY path out of the graveyard is
// escape is not a cast surface while the graveyard is too small to pay
// for it, and is one the moment it is big enough.
func TestCastableHereClearedWhenTheOnlyZoneBoundOfferIsUnpayable(t *testing.T) {
	const oracle = "test-view-escape"
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: {escapeCost("{R}", 2)}})

	spell := graveyardCard(me, "Test Uro", oracle)
	graveyardCard(me, "Filler A", "")

	// One other card under it: escape wants two, so there is no
	// payable price and therefore no cast at all.
	c := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, spell)
	if c.CastableHere {
		t.Errorf("castable_here is set on a card whose only price is an escape cost it cannot pay")
	}
	if len(c.AlternativeCosts) != 0 {
		t.Errorf("offers = %v, want none", keysOf(c.AlternativeCosts))
	}

	// The second other card makes the offer payable, and the card a
	// cast surface — at the escape cost and at nothing else.
	graveyardCard(me, "Filler B", "")
	c = cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, spell)
	if !c.CastableHere {
		t.Errorf("castable_here is clear on a card whose escape cost is payable")
	}
	if got := keysOf(c.AlternativeCosts); !sameStrings(got, []string{"escape"}) {
		t.Errorf("offers = %v, want [escape]", got)
	}
	if !c.AlternativeCostRequired {
		t.Errorf("alternative_cost_required is clear — the printed {2}{R} is not claimable from the graveyard")
	}
}

// The exception #1015's checklist names: the bit must NOT be cleared
// when the card has a path out of the zone the unpayable offer is not.
// "A grant charging the printed cost" — an impulse-shaped permission
// over a graveyard card — is that path, and rule 4 of
// validateCastPathLocked is why: a permission that names no claimable
// offer charges the printed cost, so a cast that claims nothing is
// accepted and the unpayable offer beside it takes nothing away.
func TestCastableHereSurvivesAnUnpayableOfferWhenThePrintedCostIsClaimable(t *testing.T) {
	const oracle = "test-view-priced-but-not-declared"
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	// The card prices the graveyard and does NOT declare it, so the
	// permission is the whole reason the cast is legal.
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: {escapeCost("{B}", 5)}})
	spell := graveyardCard(me, "Test Priced Spell", oracle)

	c := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, spell)
	if c.CastableHere {
		t.Errorf("castable_here is set on a graveyard card nothing opens")
	}

	if !g.GrantCastPermissionOverCardForEffect(spell, game.CastPermission{
		Player: me.ID, Zone: game.ZoneGraveyard, Scope: game.ScopeCards,
	}) {
		t.Fatalf("GrantCastPermissionOverCardForEffect: card not found")
	}
	c = cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, spell)
	if !c.CastableHere {
		t.Errorf("castable_here is clear on a card a permission opens at its printed cost")
	}
	if c.AlternativeCostRequired {
		t.Errorf("alternative_cost_required is set — the grant charges the printed cost")
	}
	if len(c.AlternativeCosts) != 0 {
		t.Errorf("offers = %v, want none — the escape cost is still unpayable", keysOf(c.AlternativeCosts))
	}
}

// Gravecrawler: the card opens the graveyard itself and prices
// nothing, so the printed cost is the only price and the bit is set
// without any offer behind it.
func TestCastableHereOnACardThatOpensItsZoneAtNoPrice(t *testing.T) {
	const oracle = "test-view-gravecrawler"
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})
	spell := graveyardCard(me, "Test Crawler", oracle)

	c := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, spell)
	if !c.CastableHere {
		t.Errorf("castable_here is clear on a card whose own text opens the graveyard for free")
	}
	if c.AlternativeCostRequired {
		t.Errorf("alternative_cost_required is set on a card that prices nothing")
	}
	if len(c.AlternativeCosts) != 0 {
		t.Errorf("offers = %v, want none", keysOf(c.AlternativeCosts))
	}
}

// And rule 3 in the other direction, because the old view got this
// wrong too: a card that declares AND prices its zone owes that price
// even under a permission that would charge the printed cost, so an
// unpayable bound offer leaves no cast at all.
func TestCastableHereClearedWhenTheCardsOwnPriceIsOwedAndUnpayable(t *testing.T) {
	const oracle = "test-view-declared-and-priced"
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: {escapeCost("{B}", 5)}})
	spell := graveyardCard(me, "Test Declared Escape", oracle)
	if !g.GrantCastPermissionOverCardForEffect(spell, game.CastPermission{
		Player: me.ID, Zone: game.ZoneGraveyard, Scope: game.ScopeCards,
	}) {
		t.Fatalf("GrantCastPermissionOverCardForEffect: card not found")
	}

	c := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, spell)
	if c.CastableHere {
		t.Errorf("castable_here is set: the card's own graveyard price is owed (rule 3) and cannot be paid")
	}
}

// --- #1012, point 1: the granted offer is merged, not overwritten ----

func TestGrantedOfferSurvivesAlongsideThePrintedSet(t *testing.T) {
	const oracle = "test-view-flashback-under-breach"
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: {
		{Key: "flashback", Label: "Flashback {3}{R}", ManaCost: "{3}{R}", FromZone: game.ZoneGraveyard},
	}})
	spell := graveyardCard(me, "Test Deep Analysis", oracle)
	graveyardCard(me, "Filler A", "")

	// Underworld Breach's shape: a graveyard permission that prices
	// the cast at escape, over a card that already prints its own
	// graveyard price.
	if !g.GrantCastPermissionOverCardForEffect(spell, game.CastPermission{
		Player: me.ID, Zone: game.ZoneGraveyard, Scope: game.ScopeCards,
		AltCostKey: "escape", ExileOtherFromGraveyard: 1,
	}) {
		t.Fatalf("GrantCastPermissionOverCardForEffect: card not found")
	}

	c := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, spell)
	if got := keysOf(c.AlternativeCosts); !sameStrings(got, []string{"escape", "flashback"}) {
		t.Errorf("offers = %v, want both the printed flashback and the granted escape", got)
	}
	if !c.CastableHere {
		t.Errorf("castable_here is clear on a card with two payable prices")
	}
	if !c.AlternativeCostRequired {
		t.Errorf("alternative_cost_required is clear — neither price is the printed one")
	}
}

// --- #1012, point 2: the wire says when the printed cost is out ------

func TestPrintedCostClaimableFromHandAndNotFromTheGraveyard(t *testing.T) {
	const oracle = "test-view-two-zones"
	g := buildActiveGame(t)
	me := g.Seats[0]
	me.Graveyard.Cards = nil
	withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: {
		{Key: "overload", Label: "Overload {4}{R}", ManaCost: "{4}{R}", ClearsTargets: true},
		{Key: "flashback", Label: "Flashback {2}{R}", ManaCost: "{2}{R}", FromZone: game.ZoneGraveyard},
	}})
	inYard := graveyardCard(me, "Test Looting", oracle)

	inHand := game.NewCard("Test Looting", me.ID)
	inHand.TypeLine = "Sorcery"
	inHand.ManaCost = "{2}{R}"
	inHand.OracleID = oracle
	inHand.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(inHand)

	v := ViewOfGameFor(g, me.ID.String())
	yard := cardInSeatZone(t, v.Seats[0].Graveyard, inYard)
	if !yard.AlternativeCostRequired {
		t.Errorf("graveyard copy: alternative_cost_required is clear, but the printed {2}{R} is not claimable from there")
	}
	if got := keysOf(yard.AlternativeCosts); !sameStrings(got, []string{"flashback"}) {
		t.Errorf("graveyard offers = %v, want [flashback]", got)
	}

	hand := cardInSeatZone(t, v.Seats[0].Hand, inHand.InstanceID)
	if hand.AlternativeCostRequired {
		t.Errorf("hand copy: alternative_cost_required is set, but every hand cast may pay the printed cost (CR 601.1)")
	}
	if got := keysOf(hand.AlternativeCosts); !sameStrings(got, []string{"overload"}) {
		t.Errorf("hand offers = %v, want [overload]", got)
	}
}

// Per viewer: the exile stamps are the grant HOLDER's, and the new
// flag travels with the offer list it qualifies.
func TestPrintedCostFlagIsStrippedForABystander(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-exiled-priced-grant"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})

	spell := game.NewCard("Test Exiled Spell", me.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{3}{R}"
	spell.OracleID = oracle
	id := exileWithGrant(t, g, spell, game.CastPermission{
		Player: me.ID, Zone: game.ZoneExile, Scope: game.ScopeCards,
		AltCostKey: "escape", ExileOtherFromGraveyard: 0,
	})

	holder := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Exile, id)
	if len(holder.AlternativeCosts) != 1 || holder.AlternativeCosts[0].Key != "escape" {
		t.Fatalf("holder offers = %v, want [escape]", keysOf(holder.AlternativeCosts))
	}
	if !holder.AlternativeCostRequired {
		t.Errorf("holder: alternative_cost_required is clear, but the grant prices the cast")
	}

	bystander := cardInSeatZone(t, ViewOfGameFor(g, opp.ID.String()).Exile, id)
	if len(bystander.AlternativeCosts) != 0 {
		t.Errorf("bystander offers = %v, want none", keysOf(bystander.AlternativeCosts))
	}
	if bystander.AlternativeCostRequired {
		t.Errorf("bystander carries alternative_cost_required; it qualifies an offer list they do not get")
	}
}

// --- the coherence test ---------------------------------------------

type castPrices struct {
	printed bool
	keys    map[string]bool
}

func (c castPrices) sorted() []string {
	out := make([]string, 0, len(c.keys))
	for k := range c.keys {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// enumeratedPrices reads the seat's own move list back into the shape
// the view stamps: for each (card, zone), whether a cast at the
// printed cost was offered and which alternative-cost keys were.
func enumeratedPrices(t *testing.T, g *game.Game, seat uuid.UUID) map[string]castPrices {
	t.Helper()
	out := map[string]castPrices{}
	for _, m := range legal.EnumerateLocked(g, seat, legal.Options{}) {
		if m.Type != legal.TypeCastSpell || m.Kind != legal.KindCast {
			continue
		}
		var p struct {
			InstanceID      string `json:"instance_id"`
			FromZone        string `json:"from_zone"`
			AlternativeCost string `json:"alternative_cost"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal move params: %v", err)
		}
		zone := p.FromZone
		if zone == "" {
			zone = "hand"
		}
		key := p.InstanceID + "@" + zone
		e := out[key]
		if e.keys == nil {
			e.keys = map[string]bool{}
		}
		if p.AlternativeCost == "" {
			e.printed = true
		} else {
			e.keys[p.AlternativeCost] = true
		}
		out[key] = e
	}
	return out
}

// viewPrices reads one projected card the way a client does: a
// graveyard card is a cast surface only when `castable_here` says so,
// a hand card always is, and `alternative_cost_required` says whether
// the printed cost is on the menu.
func viewPrices(c *CardView, zone string) castPrices {
	surface := zone == "hand" || zone == "command" || c.CastableHere
	if !surface {
		return castPrices{keys: map[string]bool{}}
	}
	keys := map[string]bool{}
	for _, o := range c.AlternativeCosts {
		keys[o.Key] = true
	}
	return castPrices{printed: !c.AlternativeCostRequired, keys: keys}
}

// The two answers are the same list, by construction, and this is the
// test that says so out loud: for a fixture board, the prices the view
// stamps on each castable card equal the prices the enumerator offers
// for the same card out of the same zone.
//
// MANA is deliberately kept out of the fixture's way: the seats hold
// six untapped lands and every seeded card is cheap, so the
// enumerator's affordability probe never removes a price the view
// showed. That asymmetry is the rule (CR 601.2g — the caster taps
// AFTER choosing the cost), not a disagreement.
func TestViewAndEnumeratorOfferTheSamePrices(t *testing.T) {
	const (
		oracleFlashback = "test-coherence-flashback"
		oracleEscape    = "test-coherence-escape"
	)
	g := busyTable(t, 1)
	me := g.Seats[g.Turn.ActiveSeat]
	withCastableZones(t, map[string][]game.ZoneKind{
		oracleFlashback: {game.ZoneGraveyard},
		oracleEscape:    {game.ZoneGraveyard},
	})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		oracleFlashback: {
			{Key: "overload", Label: "Overload {1}{R}", ManaCost: "{1}{R}", ClearsTargets: true},
			{Key: "flashback", Label: "Flashback {1}{R}", ManaCost: "{1}{R}", FromZone: game.ZoneGraveyard},
		},
		oracleEscape: {escapeCost("{R}", 2)},
	})

	// instance ID + "@" + zone -> what the row is, for the failure
	// message.
	seeded := map[string]string{}

	// The graveyard: a flashback card (one price, and it is not the
	// printed one), an escape card the yard can pay for, and a plain
	// card that is no cast surface at all.
	fb := graveyardCard(me, "Coherence Flashback", oracleFlashback)
	seeded[fb.String()+"@graveyard"] = "flashback card in the graveyard"
	esc := graveyardCard(me, "Coherence Escape", oracleEscape)
	seeded[esc.String()+"@graveyard"] = "escape card in the graveyard"
	plain := graveyardCard(me, "Coherence Plain", "")
	seeded[plain.String()+"@graveyard"] = "plain card in the graveyard"
	graveyardCard(me, "Coherence Filler A", "")
	graveyardCard(me, "Coherence Filler B", "")

	// And the same flashback card in hand, where the printed cost IS
	// claimable and only the unbound offer shows.
	inHand := game.NewCard("Coherence Flashback", me.ID)
	inHand.TypeLine = "Instant"
	inHand.ManaCost = "{R}"
	inHand.OracleID = oracleFlashback
	inHand.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(inHand)
	seeded[inHand.InstanceID.String()+"@hand"] = "flashback card in hand"

	moves := enumeratedPrices(t, g, me.ID)
	// Non-vacuity: two lists that are both empty agree about nothing.
	// The graveyard flashback cast is the row the whole comparison
	// hangs on, so the fixture asserts the enumerator really offers it
	// before the loop below checks that the view says the same.
	if got := moves[fb.String()+"@graveyard"]; got.printed || !got.keys["flashback"] {
		t.Fatalf("fixture: the enumerator offers %v (printed=%v) for the graveyard flashback card, want [flashback] only",
			got.sorted(), got.printed)
	}

	v := ViewOfGameFor(g, me.ID.String())
	var seat *PlayerView
	for i := range v.Seats {
		if v.Seats[i].ID == me.ID.String() {
			seat = &v.Seats[i]
		}
	}
	if seat == nil {
		t.Fatalf("the active seat is missing from its own view")
	}

	for key, label := range seeded {
		id, zone := key[:36], key[37:]
		zv := seat.Graveyard
		if zone == "hand" {
			zv = seat.Hand
		}
		parsed, err := uuid.Parse(id)
		if err != nil {
			t.Fatalf("%s: bad id %q", label, id)
		}
		got := viewPrices(cardInSeatZone(t, zv, parsed), zone)
		want := moves[key]
		if want.keys == nil {
			want.keys = map[string]bool{}
		}
		if got.printed != want.printed {
			t.Errorf("%s: view says printed-cost-claimable=%v, the enumerator says %v",
				label, got.printed, want.printed)
		}
		if !sameStrings(got.sorted(), want.sorted()) {
			t.Errorf("%s: view offers %v, the enumerator offers %v",
				label, got.sorted(), want.sorted())
		}
	}
}
