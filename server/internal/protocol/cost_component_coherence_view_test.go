package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cost_component_coherence_view_test.go — #1213, the ability-cost
// twin of #1024's cast-surface coherence test
// (cast_surface_coherence_view_test.go).
//
// The three components this issue built are each read by three
// consumers — the engine's validator, the protocol view's picker and
// the legal-move enumerator — off ONE walk apiece. That is the #544
// invariant, and the failure it exists to prevent is silent: the
// client offers a payment the engine refuses, or the bot is never
// offered one it would accept, and nothing in either half looks wrong
// on its own.
//
// So the property is asserted the way #1024 asserts its own: build a
// fixture board, read what the VIEW says can pay, read what the
// ENUMERATOR actually pays with, and compare the two lists.

// --- fixtures --------------------------------------------------------

func landCostClause() *game.TargetSpec {
	return &game.TargetSpec{
		Mode: "permanent", Label: "a land you control", Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsLand() },
		Min:    1, Max: 1,
	}
}

func artifactCostClause(label string, min, max int) *game.TargetSpec {
	return &game.TargetSpec{
		Mode: "permanent", Label: label, Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsArtifact() },
		Min:    min, Max: max,
	}
}

func seatReturnCostSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Return Cost Source",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000000e1",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Return a land you control to its owner's hand: mark",
			Cost: game.AbilityCost{ReturnToHand: &game.ReturnToHandCost{
				Count:  1,
				Filter: landCostClause(),
				Label:  "a land you control",
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	return id
}

func seatLandFor(g *game.Game, controller uuid.UUID, name string, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Land",
		Owner: controller, Controller: controller, Keywords: keywords,
	})
	return id
}

// seatHandCardFor puts a plain instant in a seat's hand, known to its
// owner, so a discard cost has something to pay with.
func seatHandCardFor(_ *game.Game, p *game.Player, name string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Instant"
	c.ManaCost = "{1}{R}"
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	p.Hand.PushTop(c)
	return c.InstanceID
}

func seatArtifactFor(g *game.Game, controller uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Artifact",
		Owner: controller, Controller: controller,
	})
	return id
}

// --- the return component -------------------------------------------

// The shape and the options: the clause's label, exactly the lands the
// seat controls, and a hexproof one among them — paying a cost does
// not target (CR 601.2h), so the picker must not hide it.
func TestActivatedAbilityViewCarriesReturnCostAndOptions(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatReturnCostSource(g, owner)
	plain := seatLandFor(g, owner, "Plain Land")
	hexproof := seatLandFor(g, owner, "Hexproof Land", "hexproof")
	seatLandFor(g, g.Seats[1].ID, "Their Land") // an opponent's: cannot pay
	seatArtifactFor(g, owner, "Not A Land")     // wrong type: cannot pay
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.ReturnLabel != "a land you control" {
		t.Errorf("return_label = %q, want the clause as printed", ab.ReturnLabel)
	}
	if ab.ReturnOptions == nil {
		t.Fatal("return_options is absent on an ability that prints the clause")
	}
	if ab.ReturnOptions.Min != 1 || ab.ReturnOptions.Max != 1 {
		t.Errorf("return_options bounds = %d/%d, want 1/1 — the clause's count", ab.ReturnOptions.Min, ab.ReturnOptions.Max)
	}
	got := append([]string(nil), ab.ReturnOptions.Cards...)
	sort.Strings(got)
	want := []string{plain.String(), hexproof.String()}
	sort.Strings(want)
	if !sameStrings(got, want) {
		t.Errorf("return_options = %v, want my two lands (hexproof included)", ab.ReturnOptions.Cards)
	}
}

// With nothing the clause admits, the component ships no options —
// which is the client's "this row cannot be paid" and CR 118.3's hard
// refusal read through the same walk.
func TestReturnCostViewShipsNoOptionsWhenNothingCanPay(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatReturnCostSource(g, owner)
	seatLandFor(g, g.Seats[1].ID, "Their Land")
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.ReturnOptions != nil && len(ab.ReturnOptions.Cards) != 0 {
		t.Errorf("return_options = %+v, want none", ab.ReturnOptions)
	}
}

// --- the mana-ability discard component ------------------------------

func seatDiscardManaSource(g *game.Game, owner uuid.UUID, dc *game.DiscardCost) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Discard Familiar",
		TypeLine:   "Creature — Imp",
		OracleID:   "00000000-0000-0000-0000-0000000000e2",
		Owner:      owner,
		Controller: owner,
		Power:      3, Toughness: 2,
		ManaAbilities: []game.ManaAbilityShape{{
			DiscardCards: dc,
			Produced:     "{B}",
			Label:        "Discard a card: Add {B}",
		}},
	})
	return id
}

// The mana-ability view carries the SAME three discard fields the
// activated view does, under the same wire names, so the client's
// picker is one component for both ability kinds.
func TestManaAbilityViewCarriesDiscardCostAndOptions(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	// The opening hand the fixture deals is not the subject here.
	me.Hand.Cards = nil
	src := seatDiscardManaSource(g, me.ID, &game.DiscardCost{N: 1, Label: "a card"})
	a := seatHandCardFor(g, me, "Pitch A")
	b := seatHandCardFor(g, me, "Pitch B")
	g.BumpLayerVersionForTest()

	ma := vehicleView(t, g, src).ManaAbilities[0]
	if ma.DiscardCostN != 1 || ma.DiscardCostLabel != "a card" {
		t.Errorf("discard cost shape: n=%d label=%q", ma.DiscardCostN, ma.DiscardCostLabel)
	}
	got := append([]string(nil), ma.DiscardCostOptions...)
	sort.Strings(got)
	want := []string{a.String(), b.String()}
	sort.Strings(want)
	if !sameStrings(got, want) {
		t.Errorf("discard_cost_options = %v, want the seat's whole hand", ma.DiscardCostOptions)
	}
}

// --- the variable sacrifice count ------------------------------------

func seatVariableSacSource(g *game.Game, owner uuid.UUID, spec *game.TargetSpec) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Variable Outlet",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000000e3",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Sacrifice one or more artifacts: mark",
			Cost:   game.AbilityCost{SacrificeOther: spec},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	return id
}

// An OPEN count ships its real bounds: a floor of one and max 0, which
// is LegalTargetsView's "no ceiling". A fixed clause still ships N/N.
func TestSacrificeOptionsShipTheVariableBounds(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	open := seatVariableSacSource(g, owner, artifactCostClause("one or more artifacts", 1, 0))
	fixed := seatVariableSacSource(g, owner, artifactCostClause("two artifacts", 2, 2))
	seatArtifactFor(g, owner, "Rock A")
	seatArtifactFor(g, owner, "Rock B")
	g.BumpLayerVersionForTest()

	o := vehicleView(t, g, open).ActivatedAbilities[0].SacrificeOptions
	if o == nil || o.Min != 1 || o.Max != 0 {
		t.Errorf("open clause bounds = %+v, want min 1 and no ceiling", o)
	}
	f := vehicleView(t, g, fixed).ActivatedAbilities[0].SacrificeOptions
	if f == nil || f.Min != 2 || f.Max != 2 {
		t.Errorf("fixed clause bounds = %+v, want 2/2", f)
	}
}

// --- the coherence property -------------------------------------------

// enumeratedAbilityPayments reads the seat's own move list back into
// the shape the view stamps: for each ability source, every permanent
// the enumerator would RETURN, every card it would DISCARD to a mana
// ability, and every sacrifice COUNT it offers.
type abilityPayments struct {
	returns   map[string]bool
	discards  map[string]bool
	sacCounts map[int]bool
}

func enumeratedAbilityPayments(t *testing.T, g *game.Game, seat uuid.UUID) map[string]*abilityPayments {
	t.Helper()
	out := map[string]*abilityPayments{}
	at := func(src string) *abilityPayments {
		if out[src] == nil {
			out[src] = &abilityPayments{
				returns:   map[string]bool{},
				discards:  map[string]bool{},
				sacCounts: map[int]bool{},
			}
		}
		return out[src]
	}
	for _, m := range legal.EnumerateLocked(g, seat, legal.Options{}) {
		switch m.Type {
		case legal.TypeActivateAbility:
			var p struct {
				SourceCardID string   `json:"source_card_id"`
				SacrificeIDs []string `json:"sacrifice_ids"`
				ReturnIDs    []string `json:"return_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("unmarshal activate params: %v", err)
			}
			e := at(p.SourceCardID)
			for _, id := range p.ReturnIDs {
				e.returns[id] = true
			}
			if len(p.SacrificeIDs) > 0 {
				e.sacCounts[len(p.SacrificeIDs)] = true
			}
		case legal.TypeActivateManaAbility:
			var p struct {
				CardID     string   `json:"card_id"`
				DiscardIDs []string `json:"discard_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("unmarshal mana params: %v", err)
			}
			e := at(p.CardID)
			for _, id := range p.DiscardIDs {
				e.discards[id] = true
			}
		}
	}
	return out
}

func keysOfSet(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// The two answers are subsets of one list, by construction, and this
// is the test that says so out loud.
//
// The RETURN component is the strict case: the enumerator offers one
// move per candidate, so the set it pays from is exactly the set the
// view shows. The DISCARD and SACRIFICE components are the bounded
// case — the enumerator picks ONE payment out of the options and, for
// a variable count, a capped ladder of counts — so what is asserted
// there is CONTAINMENT: every payment the enumerator offers is one the
// view listed, which is the half a disagreement would break (#544).
func TestAbilityCostViewAndEnumeratorAgree(t *testing.T) {
	// busyTable, not buildActiveGame: the enumerator needs a seat with
	// priority in a main phase and six untapped lands, so its
	// affordability probe never removes a payment the view showed.
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	ret := seatReturnCostSource(g, me.ID)
	mana := seatDiscardManaSource(g, me.ID, &game.DiscardCost{N: 1, Label: "a card"})
	open := seatVariableSacSource(g, me.ID, artifactCostClause("one or more artifacts", 1, 0))
	// The board the payments come out of, beside the fixture's own.
	seatLandFor(g, me.ID, "Coherence Land", "hexproof")
	seatArtifactFor(g, me.ID, "Rock A")
	seatArtifactFor(g, me.ID, "Rock B")
	// An opponent's board, which may never pay for any of them.
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seatLandFor(g, them.ID, "Their Land")
	seatArtifactFor(g, them.ID, "Their Rock")
	g.BumpLayerVersionForTest()

	moves := enumeratedAbilityPayments(t, g, me.ID)

	// --- the return component ---
	view := vehicleView(t, g, ret).ActivatedAbilities[0]
	if view.ReturnOptions == nil || len(view.ReturnOptions.Cards) == 0 {
		t.Fatal("the view offers no return options for an ability the enumerator can pay")
	}
	offered := moves[ret.String()]
	if offered == nil || len(offered.returns) == 0 {
		t.Fatal("the enumerator offered no activation of the return-cost ability the view says is payable")
	}
	shown := map[string]bool{}
	for _, id := range view.ReturnOptions.Cards {
		shown[id] = true
	}
	for id := range offered.returns {
		if !shown[id] {
			t.Errorf("the enumerator pays with %s, which the view does not offer", id)
		}
	}
	// The other direction, bounded by the enumerator's own expansion
	// cap: it must reach at least the first option the view listed,
	// which is the one the client's "Choose for me" would take.
	if !offered.returns[view.ReturnOptions.Cards[0]] {
		t.Errorf("the enumerator never offers the view's first option %s (%v)",
			view.ReturnOptions.Cards[0], keysOfSet(offered.returns))
	}

	// --- the mana-ability discard ---
	manaView := vehicleView(t, g, mana).ManaAbilities[0]
	shownDiscards := map[string]bool{}
	for _, id := range manaView.DiscardCostOptions {
		shownDiscards[id] = true
	}
	manaOffered := moves[mana.String()]
	if manaOffered == nil || len(manaOffered.discards) == 0 {
		t.Fatal("the enumerator offered no mana activation for an ability the view says can be paid")
	}
	for id := range manaOffered.discards {
		if !shownDiscards[id] {
			t.Errorf("the enumerator pitches %s, which the view does not offer", id)
		}
	}

	// --- the variable sacrifice count ---
	openView := vehicleView(t, g, open).ActivatedAbilities[0].SacrificeOptions
	if openView == nil {
		t.Fatal("the view offers no sacrifice options for the open-count ability")
	}
	openOffered := moves[open.String()]
	if openOffered == nil || len(openOffered.sacCounts) == 0 {
		t.Fatal("the enumerator offered no activation of the open-count ability")
	}
	for n := range openOffered.sacCounts {
		if n < openView.Min {
			t.Errorf("the enumerator offers a payment of %d, below the view's floor of %d", n, openView.Min)
		}
		if openView.Max != 0 && n > openView.Max {
			t.Errorf("the enumerator offers a payment of %d, above the view's ceiling of %d", n, openView.Max)
		}
		if n > len(openView.Cards) {
			t.Errorf("the enumerator offers a payment of %d out of %d options", n, len(openView.Cards))
		}
	}
	// The cap is a policy and it is documented in docs/bot.md; what is
	// pinned here is that it IS a cap, so a wide board never turns one
	// variable cost into an arity of the target cross product.
	if len(openOffered.sacCounts) > 3 {
		t.Errorf("%d distinct sacrifice counts offered, the cap is maxEnumeratedVariableCounts (3)", len(openOffered.sacCounts))
	}
}
