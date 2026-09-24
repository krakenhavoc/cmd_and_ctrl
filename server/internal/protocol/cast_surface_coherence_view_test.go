package protocol

import (
	"encoding/json"
	"fmt"
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

// withTargetSpecsByOracle stubs the target clause of one or more
// oracle IDs, chaining to whatever the catalog already answered — the
// same shape withAltCostsByOracle has, and chaining because the
// fixture's own Lightning Bolt has a real clause this must not blank.
func withTargetSpecsByOracle(t *testing.T, specs map[string]func() *game.TargetSpec) {
	t.Helper()
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if s, ok := specs[id]; ok {
			return s()
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })
}

// targetACreature is the narrowest clause that is non-empty on the
// coherence fixture's board: one creature, anywhere on the
// battlefield. Narrowed per seat by nothing here — which is the
// point. The assertion #1166 adds is about WHOSE answer reaches
// whose frame, not about how small the answer is.
func targetACreature() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  "creature",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// withModeSpecsByOracle stubs the modal clause of one or more oracle
// IDs, chaining like the two stubs above (#1172).
func withModeSpecsByOracle(t *testing.T, specs map[string]func() *game.ModeSpec) {
	t.Helper()
	prev := game.CatalogModeSpec
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if s, ok := specs[id]; ok {
			return s()
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogModeSpec = prev })
}

// chooseOneTargetedMode is the modal clause #1172 is about: a "choose
// one" whose bullets TARGET, so each option carries a legal set
// computed for the seat the stamp was built for. The second bullet has
// two clauses, which is what puts a nested `clauses` on the wire
// beside the nested `legal_targets`.
func chooseOneTargetedMode() *game.ModeSpec {
	twoClauses := targetACreature()
	twoClauses.Label = "target creature"
	second := targetACreature()
	second.Label = "a second target creature"
	twoClauses.Rest = []game.TargetClause{*second}
	return &game.ModeSpec{
		Prompt: "Choose one —",
		Min:    1, Max: 1,
		Options: []game.ModeOption{
			{Label: "Destroy target creature.", Targets: targetACreature()},
			{Label: "Tap target creature and a second target creature.", Targets: twoClauses},
			{Label: "Draw a card."},
		},
	}
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

// castWindowOpen puts a buildActiveGame fixture in the window a
// SORCERY can actually be cast in: precombat main, empty stack, and
// seat 0 — the seat every fixture below casts as — active.
//
// buildActiveGame stops at the untap step, which did not matter while
// `castable_here` answered "is this zone a cast surface at a price you
// can pay". Since #1195 it also answers CR 307.1, out of the same
// game.CastTimingOpenLocked the announce path and the bot enumerator
// read — so a fixture that seeds a sorcery in a graveyard and asserts
// the bit has to be somewhere the cast is open. That the seven tests
// touched by this needed it is the divergence the issue is about:
// each of them marked a cast surface the announce path would have
// refused with ErrSorcerySpeedRequired.
func castWindowOpen(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
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

// multiFaceGraveyardCard seeds one MULTI-FACE card into a seat's
// graveyard, front face up (CR 712.8a, MoveCard), each half's catalog
// entry keyed the way ADR 0034 keys them: the bare oracle ID for the
// front and "<oracle_id>#1" for the back.
//
// A modal DFC, because its halves are independently castable
// (CR 712.11b) — which is what makes "the BACK face opens the
// graveyard" something a cast can act on. Built from a test spec
// rather than seeded from the catalog because no catalog card
// declares it yet: #1171 is latent, and a fixture is how a latent
// divergence gets pinned before the first printing makes it live.
func multiFaceGraveyardCard(p *game.Player, oracle string, faces []game.Face) uuid.UUID {
	c := game.NewCard(faces[0].Name, p.ID)
	c.OracleID = oracle
	c.Layout = game.LayoutModalDFC
	c.Faces = faces
	c.SetFace(0)
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	id := c.InstanceID
	p.Graveyard.PushTop(c)
	return id
}

// cardInSeatZone finds one card in one projected zone.
func cardInSeatZone(t *testing.T, z ZoneView, id uuid.UUID) *CardView {
	t.Helper()
	if c := findInZone(z, id); c != nil {
		return c
	}
	t.Fatalf("card %s missing from the %s view", id, z.Kind)
	return nil
}

// findInZone is cardInSeatZone for a card that may legitimately be
// absent — a pile a given viewer may not see into.
func findInZone(z ZoneView, id uuid.UUID) *CardView {
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id.String() {
			return &z.Cards[i]
		}
	}
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
	castWindowOpen(t, g)
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
	castWindowOpen(t, g)
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
	castWindowOpen(t, g)
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
	castWindowOpen(t, g)
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
	castWindowOpen(t, g)
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
	castWindowOpen(t, g)
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
		t.Errorf("hand copy: alternative_cost_required is set, but every hand cast may pay the printed cost (CR 601.2)")
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
//
// ACROSS THE CASTABLE FACES as well as the card (#1171, #992). The
// card's own block describes the face that is UP, and a pile is
// front-up (CR 712.8a) — so a card whose BACK face alone opens the
// zone has a front that is no cast surface and a back that is, and
// the union over the blocks is what the client's face picker reads.
// It is the shape of the other side of this comparison too: the
// enumerator walks game.CastableFacesUnder and files every face's
// moves under the one instance.
func viewPrices(c *CardView, zone string) castPrices {
	out := castPrices{keys: map[string]bool{}}
	add := func(s CastSurfaceView) {
		if zone != "hand" && zone != "command" && !s.CastableHere {
			return
		}
		if !s.AlternativeCostRequired {
			out.printed = true
		}
		for _, o := range s.AlternativeCosts {
			out.keys[o.Key] = true
		}
	}
	add(c.CastSurfaceView)
	for _, f := range c.Faces {
		add(f.CastSurfaceView)
	}
	return out
}

// seededCast is one row of the coherence fixture: a card, the zone it
// sits in and WHOSE zone that is. The third field is what #1035 added
// — a cast surface is no longer always a pile of the seat being
// enumerated, because a permission over another seat's graveyard or
// library top is one too.
type seededCast struct {
	label string
	seat  uuid.UUID
	zone  string
}

// zoneViewOf picks one seat's projected pile out of a filtered view.
func zoneViewOf(t *testing.T, v GameView, seat uuid.UUID, zone string) ZoneView {
	t.Helper()
	p := seatViewOf(t, v, seat)
	switch zone {
	case "hand":
		return p.Hand
	case "library":
		return p.Library
	default:
		return p.Graveyard
	}
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
		oracleRevealed  = "test-coherence-revealed-in-hand"
		// #1171: a MULTI-FACE card, and everything it declares it
		// declares on the BACK — the half the pile is not showing.
		oracleBackFace = "test-coherence-back-face-flashback"
		// #1172: a MODAL card in a graveyard. `modes` is public there
		// — it is the card's printed text — and until #1172 the legal
		// sets inside it were the pile OWNER's answer on every frame.
		oracleModal = "test-coherence-modal-flashback"
	)
	g := busyTable(t, 1)
	me := g.Seats[g.Turn.ActiveSeat]
	withCastableZones(t, map[string][]game.ZoneKind{
		oracleFlashback: {game.ZoneGraveyard},
		oracleEscape:    {game.ZoneGraveyard},
		// #1171: THE BACK FACE's entry and not the card's. The front
		// half declares nothing, so a reader that asks the bare
		// oracle ID — which is face 0's catalog key — is told this
		// card does not open the graveyard at all.
		oracleBackFace + "#1": {game.ZoneGraveyard},
		oracleModal:           {game.ZoneGraveyard},
	})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		oracleFlashback: {
			{Key: "overload", Label: "Overload {1}{R}", ManaCost: "{1}{R}", ClearsTargets: true},
			{Key: "flashback", Label: "Flashback {1}{R}", ManaCost: "{1}{R}", FromZone: game.ZoneGraveyard},
		},
		oracleEscape: {escapeCost("{R}", 2)},
		// #1171: priced on the back face too, so the row below pins
		// the PRICE LIST of a face the pile is not showing and not
		// merely the bit above it.
		oracleBackFace + "#1": {
			{Key: "flashback", Label: "Flashback {1}{R}", ManaCost: "{1}{R}", FromZone: game.ZoneGraveyard},
		},
		// #1172: an offer with a target clause of its own AND a
		// card-shaped pay clause, so `alternative_costs[i]` carries
		// both nested lists on a PUBLIC pile.
		oracleModal: {escapeCost("{R}", 2)},
	})
	// #1166: the one seeded card with a target clause, so the
	// per-viewer half of the assertion below is answered by a stamp
	// rather than by a field nothing filled.
	//
	// #1172 adds the modal card's, which is also what gives its
	// escape offer a `legal_targets` to carry: an offer that leaves
	// the spell's clause alone is projected with the card's own.
	withTargetSpecsByOracle(t, map[string]func() *game.TargetSpec{
		oracleRevealed: targetACreature,
		oracleModal:    targetACreature,
	})
	// #1172: the modal clause itself, on the graveyard card and on
	// the modal DFC's BACK face — the per-card and the per-face halves
	// of the same nested split.
	withModeSpecsByOracle(t, map[string]func() *game.ModeSpec{
		oracleModal:           chooseOneTargetedMode,
		oracleBackFace + "#1": chooseOneTargetedMode,
	})

	// instance ID + "@" + zone -> which row it is, for the failure
	// message and for finding the pile again in the projection.
	seeded := map[string]seededCast{}

	// The graveyard: a flashback card (one price, and it is not the
	// printed one), an escape card the yard can pay for, and a plain
	// card that is no cast surface at all.
	fb := graveyardCard(me, "Coherence Flashback", oracleFlashback)
	seeded[fb.String()+"@graveyard"] = seededCast{"flashback card in the graveyard", me.ID, "graveyard"}
	esc := graveyardCard(me, "Coherence Escape", oracleEscape)
	seeded[esc.String()+"@graveyard"] = seededCast{"escape card in the graveyard", me.ID, "graveyard"}
	plain := graveyardCard(me, "Coherence Plain", "")
	seeded[plain.String()+"@graveyard"] = seededCast{"plain card in the graveyard", me.ID, "graveyard"}
	// #1172: a MODAL card in the same pile. `modes` is public there,
	// and it carries a legal set per targeted bullet plus a nested
	// `clauses` on the two-clause one; its escape offer carries a
	// `legal_targets` and a `pay_options`. Four nested lists on one
	// card, all of them the pile owner's answer, all of them public
	// until this issue.
	modal := graveyardCard(me, "Coherence Modal", oracleModal)
	seeded[modal.String()+"@graveyard"] = seededCast{"modal card in the graveyard", me.ID, "graveyard"}
	// #1055: known to the whole table, so the per-viewer assertion at
	// the bottom is answered by the STAMP and not by the redaction a
	// non-knower would get anyway.
	for _, id := range []uuid.UUID{fb, esc, plain, modal} {
		knownToEveryone(g, me.Graveyard, id)
	}
	graveyardCard(me, "Coherence Filler A", "")
	graveyardCard(me, "Coherence Filler B", "")

	// #1171: a modal DFC whose BACK face alone opens — and prices —
	// the graveyard. The row this fixture could not have: every other
	// seeded card is single-faced, so "which zones does this card
	// open" was never asked of a card with more than one answer.
	//
	// The enumerator has always asked it of every castable face
	// (game.CardCastableFromAnyFace); the view asked it of the bare
	// oracle ID, which resolves to face 0's entry. So this card was a
	// legal move for a bot and a card with NO announce stamps at all
	// on the wire — no `castable_here`, no price list, no target
	// clause, and a zone browser with no button behind a cast
	// CastSpell would have accepted. Silent, because the two answers
	// were never compared for a multi-face card until this row.
	mdfc := multiFaceGraveyardCard(me, oracleBackFace, []game.Face{
		{Name: "Coherence Front", TypeLine: "Creature — Bear", ManaCost: "{1}{R}", Power: 2, Toughness: 2},
		{Name: "Coherence Back", TypeLine: "Instant", ManaCost: "{1}{R}"},
	})
	knownToEveryone(g, me.Graveyard, mdfc)
	seeded[mdfc.String()+"@graveyard"] = seededCast{"modal DFC whose back face prices the graveyard", me.ID, "graveyard"}

	// And the same flashback card in hand, where the printed cost IS
	// claimable and only the unbound offer shows.
	inHand := game.NewCard("Coherence Flashback", me.ID)
	inHand.TypeLine = "Instant"
	inHand.ManaCost = "{R}"
	inHand.OracleID = oracleFlashback
	inHand.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(inHand)
	seeded[inHand.InstanceID.String()+"@hand"] = seededCast{"flashback card in hand", me.ID, "hand"}

	// #1166: and a REVEALED card in the same hand, with a target
	// clause. A Thoughtseize or a Telepathy makes another seat a
	// knower, keepKnownInHandZone keeps the card on their frame, and
	// until #1166 their copy carried the HAND OWNER's legal target
	// set — the #891 shape #1055 took off the graveyard, one zone
	// over. Seeded here rather than in a parallel test because this
	// is already the fixture that asks "whose answer is this", and a
	// hand is the last cast surface that was not asked.
	revealed := game.NewCard("Coherence Revealed", me.ID)
	revealed.TypeLine = "Instant"
	revealed.ManaCost = "{R}"
	revealed.OracleID = oracleRevealed
	me.Hand.PushTop(revealed)
	knownToEveryone(g, me.Hand, revealed.InstanceID)
	seeded[revealed.InstanceID.String()+"@hand"] = seededCast{"revealed targeted card in hand", me.ID, "hand"}

	// #1035: ANOTHER seat's library top, under a Xanathar-shaped grant
	// this seat holds. It is the same property on a pile that is not
	// this seat's — and the row that would have been silently vacuous
	// before, because the permission opened nothing at all.
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	them.Library.Cards = nil
	buried := libraryTopCard(them, "Coherence Buried", "")
	theirTop := libraryTopCard(them, "Coherence Their Top", "")
	xanatharOver(t, g, me, them)
	seeded[theirTop.String()+"@library"] = seededCast{"another seat's library top under a grant", them.ID, "library"}

	moves := enumeratedPrices(t, g, me.ID)
	// #1035 non-vacuity, and the position rule from the other end: the
	// enumerator offers the top card of the named library and nothing
	// under it.
	if got := moves[theirTop.String()+"@library"]; !got.printed {
		t.Fatalf("fixture: the enumerator offers %v (printed=%v) for the foreign library top, want the printed cost",
			got.sorted(), got.printed)
	}
	if got, ok := moves[buried.String()+"@library"]; ok {
		t.Errorf("the enumerator offered a card BELOW another seat's library top: %v", got.sorted())
	}
	// Non-vacuity: two lists that are both empty agree about nothing.
	// The graveyard flashback cast is the row the whole comparison
	// hangs on, so the fixture asserts the enumerator really offers it
	// before the loop below checks that the view says the same.
	if got := moves[fb.String()+"@graveyard"]; got.printed || !got.keys["flashback"] {
		t.Fatalf("fixture: the enumerator offers %v (printed=%v) for the graveyard flashback card, want [flashback] only",
			got.sorted(), got.printed)
	}
	// #1171 non-vacuity: the enumerator really does offer the BACK
	// face's flashback out of the graveyard, so the loop below is
	// comparing the view against a non-empty answer.
	if got := moves[mdfc.String()+"@graveyard"]; got.printed || !got.keys["flashback"] {
		t.Fatalf("fixture: the enumerator offers %v (printed=%v) for the back-face flashback card, want [flashback] only",
			got.sorted(), got.printed)
	}

	v := ViewOfGameFor(g, me.ID.String())

	// #1166, the owner's half: routing the hand through
	// applyCastStampsFor must not cost the hand's owner their own
	// clause. Asserted before the bystander loop, so a fix that
	// simply stopped stamping hands at all fails here rather than
	// passing everything below.
	own := cardInSeatZone(t, zoneViewOf(t, v, me.ID, "hand"), revealed.InstanceID)
	if own.LegalTargets == nil || len(own.LegalTargets.Cards) == 0 {
		t.Errorf("the hand's owner lost their own legal target set: %+v", own.LegalTargets)
	}

	// #1172, the owner's half of the NESTED split, and the
	// non-vacuity for the bystander loop below: the pile's owner keeps
	// every nested list the stamp computed for them, on the card and
	// on the face that carries the modal clause.
	ownModal := cardInSeatZone(t, zoneViewOf(t, v, me.ID, "graveyard"), modal)
	assertNestedTargetsPresent(t, "the graveyard's owner", ownModal.CastSurfaceView)
	ownMDFC := cardInSeatZone(t, zoneViewOf(t, v, me.ID, "graveyard"), mdfc)
	if len(ownMDFC.Faces) != 2 {
		t.Fatalf("the modal DFC ships %d faces, want both halves", len(ownMDFC.Faces))
	}
	if ownMDFC.Faces[1].Modes == nil {
		t.Fatalf("the owner's faces[1] carries no `modes`; the per-face nested assertions are vacuous")
	}
	if lt := ownMDFC.Faces[1].Modes.Options[0].LegalTargets; lt == nil || len(lt.Cards) == 0 {
		t.Errorf("the owner lost their own per-face nested legal target set: %+v", lt)
	}

	for key, row := range seeded {
		id, zone := key[:36], key[37:]
		parsed, err := uuid.Parse(id)
		if err != nil {
			t.Fatalf("%s: bad id %q", row.label, id)
		}
		zv := zoneViewOf(t, v, row.seat, zone)
		got := viewPrices(cardInSeatZone(t, zv, parsed), zone)
		want := moves[key]
		if want.keys == nil {
			want.keys = map[string]bool{}
		}
		if got.printed != want.printed {
			t.Errorf("%s: view says printed-cost-claimable=%v, the enumerator says %v",
				row.label, got.printed, want.printed)
		}
		if !sameStrings(got.sorted(), want.sorted()) {
			t.Errorf("%s: view offers %v, the enumerator offers %v",
				row.label, got.sorted(), want.sorted())
		}
	}

	// #1055: the property is PER VIEWER, and the loop above only ever
	// asked it of the seat the moves were enumerated for. A cast
	// surface is a statement about a player — "may YOU cast this from
	// here" — so the same fixture, projected for a seat that holds
	// none of these casts, must mark none of them.
	//
	// The enumerator cannot be the other half of this comparison: a
	// seat without priority gets no move list at all, so two empty
	// lists would agree about nothing. The assertion is the view's own
	// bit, against the seat it is being built for.
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	vOther := ViewOfGameFor(g, other.ID.String())
	for key, row := range seeded {
		if row.seat == other.ID {
			// A pile this seat owns is their own answer and belongs
			// in the loop above.
			continue
		}
		// #1166: the hand is no longer skipped. An UNREVEALED hand
		// card is dropped from the bystander's frame wholesale by
		// keepKnownInHandZone and falls out at the findInZone check
		// below, which is the stronger answer; a REVEALED one is
		// kept, is known to this seat, and is exactly the copy that
		// used to carry its owner's clause.
		id, err := uuid.Parse(key[:36])
		if err != nil {
			t.Fatalf("%s: bad id %q", row.label, key[:36])
		}
		zv := zoneViewOf(t, vOther, row.seat, row.zone)
		c := findInZone(zv, id)
		if c == nil {
			// A foreign library top this seat may not look at is
			// dropped from their projection wholesale, which is a
			// stronger answer than an unset bit.
			continue
		}
		if !c.KnownByYou {
			t.Errorf("%s: the fixture hid the card from the bystander, so the bit below "+
				"would be cleared by the redaction rather than by the stamp", row.label)
		}
		if c.CastableHere {
			t.Errorf("%s: castable_here is set for a seat that cannot cast it — "+
				"the bit is the VIEWER's answer, not the pile owner's (#1055)", row.label)
		}
		if c.LegalTargets != nil {
			t.Errorf("%s: a bystander got one seat's legal target set: %+v", row.label, c.LegalTargets)
		}
		// #1172: and one level in, where `modes` and
		// `alternative_costs` legitimately survive on a public pile
		// and their legal sets do not.
		assertNoNestedTargets(t, row.label, "", c.CastSurfaceView)
		// #992 / #1171: and per face, because the multi-face row is
		// the one whose answer lives there. The per-viewer split
		// travels down to a face's block exactly as it does to a
		// card's, so a bystander reads neither.
		for i := range c.Faces {
			if c.Faces[i].CastableHere {
				t.Errorf("%s: faces[%d].castable_here is set for a seat that cannot cast it", row.label, i)
			}
			if c.Faces[i].LegalTargets != nil {
				t.Errorf("%s: a bystander got one seat's per-face legal target set: %+v",
					row.label, c.Faces[i].LegalTargets)
			}
			assertNoNestedTargets(t, row.label, fmt.Sprintf("faces[%d].", i), c.Faces[i].CastSurfaceView)
		}
	}
}

// assertNestedTargetsPresent is the non-vacuity half of #1172: the
// seat the stamp was built for really is handed all four nested lists,
// so an assertion that a bystander has none is about the strip and not
// about a fixture that never filled them.
func assertNestedTargetsPresent(t *testing.T, who string, s CastSurfaceView) {
	t.Helper()
	if s.Modes == nil || len(s.Modes.Options) < 2 {
		t.Fatalf("%s: `modes` is absent or too short: %+v", who, s.Modes)
	}
	if lt := s.Modes.Options[0].LegalTargets; lt == nil || len(lt.Cards) == 0 {
		t.Errorf("%s: lost modes[0].legal_targets: %+v", who, lt)
	}
	if len(s.Modes.Options[1].Clauses) < 2 {
		t.Errorf("%s: lost modes[1].clauses: %+v", who, s.Modes.Options[1].Clauses)
	}
	if len(s.AlternativeCosts) == 0 {
		t.Fatalf("%s: no offers at all, so the nested offer assertions are vacuous", who)
	}
	if lt := s.AlternativeCosts[0].LegalTargets; lt == nil || len(lt.Cards) == 0 {
		t.Errorf("%s: lost alternative_costs[0].legal_targets: %+v", who, lt)
	}
	if po := s.AlternativeCosts[0].PayOptions; po == nil || len(po.Cards) == 0 {
		t.Errorf("%s: lost alternative_costs[0].pay_options: %+v", who, po)
	}
}

// assertNoNestedTargets is the bystander half: `modes` and
// `alternative_costs` may survive — they are the card's printed text
// and its printed prices — and every legal set inside them must be
// gone (#1172).
func assertNoNestedTargets(t *testing.T, label, path string, s CastSurfaceView) {
	t.Helper()
	if s.Modes != nil {
		for i, o := range s.Modes.Options {
			if o.LegalTargets != nil {
				t.Errorf("%s: a bystander got %smodes[%d].legal_targets — one seat's answer inside a "+
					"public field (#1172): %+v", label, path, i, o.LegalTargets)
			}
			if o.Clauses != nil {
				t.Errorf("%s: a bystander got %smodes[%d].clauses: %+v", label, path, i, o.Clauses)
			}
		}
	}
	for i, o := range s.AlternativeCosts {
		if o.LegalTargets != nil {
			t.Errorf("%s: a bystander got %salternative_costs[%d].legal_targets: %+v",
				label, path, i, o.LegalTargets)
		}
		if o.PayOptions != nil {
			t.Errorf("%s: a bystander got %salternative_costs[%d].pay_options: %+v",
				label, path, i, o.PayOptions)
		}
	}
}
