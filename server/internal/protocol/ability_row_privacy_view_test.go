package protocol

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // the three catalog proof cards
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ability_row_privacy_view_test.go — #1369. A battlefield permanent's
// `activated_abilities` and `mana_abilities` rows are built once, with
// the permanent's CONTROLLER as "you", and a permanent is known to the
// whole table — so until #1369 every row reached every viewer intact.
// Three fields on those rows are read out of the controller's HAND:
//
//   - `activated_abilities[i].discard_cost_options` — Fauna Shaman's
//     "Discard a creature card";
//   - `mana_abilities[i].discard_cost_options` — Skirge Familiar's
//     "Discard a card";
//   - `mana_abilities[i].exile_cost_options` — Cadaverous Bloom's
//     "Exile a card from your hand".
//
// For a FILTERED clause the list is hidden information twice over: its
// length is how many creature cards the controller holds, and its IDs
// are a handle on specific cards a viewer may have seen elsewhere (CR
// 401.2, CR 402.3). They now reach the controller's frame alone, on
// the per-seat carrier CardView.abilityOffers.

// rowScope is how far one field of an ability row travels.
type rowScope int

const (
	// rowPublic is the printed ability, or a read of the public
	// battlefield: every viewer of the permanent gets it.
	rowPublic rowScope = iota
	// rowHiddenZone is read out of the controller's hand or library
	// and reaches the controller's frame alone (#1369).
	rowHiddenZone
	// rowHiddenUnlessGraveyard is an exile-cost option list: hidden
	// when `exile_cost_zone` names the hand (or names nothing), public
	// when it names the graveyard, which every viewer may read
	// (#1297, #1369). The guard's filled row carries a zone of "x",
	// so it is checked as hidden there, and the graveyard half has an
	// assertion of its own.
	rowHiddenUnlessGraveyard
)

// activatedRowScopes places every field of ActivatedAbilityView,
// embedded CounterCostView included. A field missing from this table
// fails TestAbilityRowScopesPlaceEveryField rather than defaulting to
// anything: "does this read a zone the viewer cannot see" is the
// question a new field has to answer, and the answer belongs here.
var activatedRowScopes = map[string]rowScope{
	// The printed ability and its printed cost.
	"Index": rowPublic, "Label": rowPublic, "TapCost": rowPublic,
	"SacrificeSelf": rowPublic, "ManaCost": rowPublic, "LifeCost": rowPublic,
	"SorcerySpeed": rowPublic, "LoyaltyCost": rowPublic, "DiscardSelf": rowPublic,
	"ExileSelf": rowPublic, "CrewCost": rowPublic, "DemandsX": rowPublic,
	"MinX": rowPublic, "XSlots": rowPublic, "PhyrexianSymbols": rowPublic,
	"TargetMode": rowPublic, "Modes": rowPublic,
	// The printed discard clause: the count in "Discard two cards"
	// and the clause's words. Public — the issue's own line.
	"DiscardCostN": rowPublic, "DiscardCostLabel": rowPublic,
	"SacrificeLabel": rowPublic, "ReturnLabel": rowPublic, "TapOthersLabel": rowPublic,
	// Verdicts over public state: whose turn it is, the stack, the
	// battlefield, the ability's own activation record.
	"ChargedManaCost": rowPublic, "ConditionUnmet": rowPublic, "Exhausted": rowPublic,
	"CantActivate": rowPublic, "TimingClosed": rowPublic,
	// Option lists read off the BATTLEFIELD, which every viewer can
	// count for themselves.
	"SacrificeOptions": rowPublic, "CrewOptions": rowPublic, "ReturnOptions": rowPublic,
	"Waterbend": rowPublic, "TapOthersOptions": rowPublic,
	// Target sets: an activated ability never targets a card in a
	// hand (game.zonesOfKindLocked's hidden-zone caution).
	"LegalTargets": rowPublic, "Clauses": rowPublic,
	// CounterCostView: every counter lives on a permanent.
	"CounterCostN": rowPublic, "CounterCostKind": rowPublic, "CounterCostSelf": rowPublic,
	"CounterCostLabel": rowPublic, "CounterCostAmong": rowPublic,
	"CounterCostVariable": rowPublic, "CounterCostMax": rowPublic,
	"CounterCostOptions": rowPublic, "CounterCostAdd": rowPublic,
	"CounterCostAddKind": rowPublic, "CounterAddBlocked": rowPublic,
	// #1297's exile-N-cards clause: its printed count and words, and
	// the pile it names.
	"ExileCostN": rowPublic, "ExileCostLabel": rowPublic, "ExileCostZone": rowPublic,
	// #1369: the cards in the controller's HAND that could pay.
	"DiscardCostOptions": rowHiddenZone,
	"ExileCostOptions":   rowHiddenUnlessGraveyard,
}

// manaRowScopes is activatedRowScopes for ManaAbilityView.
var manaRowScopes = map[string]rowScope{
	"Index": rowPublic, "Label": rowPublic, "TapCost": rowPublic,
	"SacrificeCost": rowPublic, "ExileSelf": rowPublic, "LifeCost": rowPublic,
	"ManaCost": rowPublic, "Produced": rowPublic, "Restrictions": rowPublic,
	"SacrificeLabel": rowPublic, "SacrificeOptions": rowPublic,
	"ChargedManaCost": rowPublic, "ConditionUnmet": rowPublic, "Exhausted": rowPublic,
	"CantActivate": rowPublic, "AddsNoMana": rowPublic,
	"DiscardCostN": rowPublic, "DiscardCostLabel": rowPublic,
	"ExileCostN": rowPublic, "ExileCostLabel": rowPublic, "ExileCostZone": rowPublic,
	"CounterCostN": rowPublic, "CounterCostKind": rowPublic, "CounterCostSelf": rowPublic,
	"CounterCostLabel": rowPublic, "CounterCostAmong": rowPublic,
	"CounterCostVariable": rowPublic, "CounterCostMax": rowPublic,
	"CounterCostOptions": rowPublic, "CounterCostAdd": rowPublic,
	"CounterCostAddKind": rowPublic, "CounterAddBlocked": rowPublic,
	// #1369: the cards in the controller's HAND that could pay.
	"DiscardCostOptions": rowHiddenZone,
	"ExileCostOptions":   rowHiddenUnlessGraveyard,
}

// fillEveryField sets every exported field of v, recursing into
// embedded and pointed-to structs, so the scope check below sees every
// key the wire could carry rather than a struct that was mostly empty.
func fillEveryField(t *testing.T, v reflect.Value) {
	t.Helper()
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				fillEveryField(t, v.Field(i))
			}
		}
	case reflect.Ptr:
		p := reflect.New(v.Type().Elem())
		fillEveryField(t, p.Elem())
		v.Set(p)
	case reflect.Slice:
		s := reflect.MakeSlice(v.Type(), 1, 1)
		fillEveryField(t, s.Index(0))
		v.Set(s)
	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		k := reflect.New(v.Type().Key()).Elem()
		e := reflect.New(v.Type().Elem()).Elem()
		fillEveryField(t, k)
		fillEveryField(t, e)
		m.SetMapIndex(k, e)
		v.Set(m)
	case reflect.String:
		v.SetString("x")
	case reflect.Bool:
		v.SetBool(true)
	case reflect.Int, reflect.Int64, reflect.Int32:
		v.SetInt(1)
	default:
		t.Fatalf("fillEveryField: no filler for %s", v.Type())
	}
}

// flatFields lists the exported fields of a struct type, flattening
// embedded structs the way encoding/json does, as name → index path.
func flatFields(typ reflect.Type, prefix []int, out map[string][]int) {
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if !f.IsExported() {
			continue
		}
		path := append(append([]int(nil), prefix...), i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			flatFields(f.Type, path, out)
			continue
		}
		out[f.Name] = path
	}
}

// TestAbilityRowScopesPlaceEveryField is the guard on the public
// strip, in the shape TestHandPublicCastSurfaceIsAnAllowlist takes for
// the cast surface: every field placed, every hidden-zone field gone
// from the public row, every public field still there. It fails three
// ways — a field nobody placed, a hidden field the strip kept (the
// leak), and a public field it cleared (a regression every viewer
// would see).
func TestAbilityRowScopesPlaceEveryField(t *testing.T) {
	check := func(t *testing.T, name string, table map[string]rowScope, full reflect.Value, public reflect.Value, private bool) {
		t.Helper()
		fields := map[string][]int{}
		flatFields(full.Type(), nil, fields)
		for f := range fields {
			if _, ok := table[f]; !ok {
				t.Fatalf("%s.%s is not placed: say whether it reads a zone a viewer other than "+
					"the controller cannot see (#1369)", name, f)
			}
		}
		for f := range table {
			if _, ok := fields[f]; !ok {
				t.Errorf("%s placed %s, which is not a field of the view any more", name, f)
			}
		}
		for f, path := range fields {
			if full.FieldByIndex(path).IsZero() {
				t.Fatalf("%s.%s was left zero by the filler; the check below would be vacuous", name, f)
			}
			kept := !public.FieldByIndex(path).IsZero()
			if want := table[f] == rowPublic; kept != want {
				t.Errorf("%s.%s: kept on the public row = %v, want %v", name, f, kept, want)
			}
		}
		if !private {
			t.Errorf("%s: the strip cleared a hidden list and did not report it, so nothing "+
				"would be filed for the controller", name)
		}
	}

	var act ActivatedAbilityView
	fillEveryField(t, reflect.ValueOf(&act).Elem())
	pubAct, private := publicActivatedAbilityRow(act)
	check(t, "ActivatedAbilityView", activatedRowScopes, reflect.ValueOf(act), reflect.ValueOf(pubAct), private)
	// The strip works on a copy: the controller's row is the input.
	if act.DiscardCostOptions == nil {
		t.Error("the public strip reached through into the controller's own row")
	}

	var mana ManaAbilityView
	fillEveryField(t, reflect.ValueOf(&mana).Elem())
	pubMana, private := publicManaAbilityRow(mana)
	check(t, "ManaAbilityView", manaRowScopes, reflect.ValueOf(mana), reflect.ValueOf(pubMana), private)

	// rowHiddenUnlessGraveyard's other half: a GRAVEYARD exile list is
	// a pile every viewer may read, so it survives the strip on both
	// views and, alone, files nothing per seat. A row stamped with NO
	// zone is read as the hand — the direction that does not leak.
	grave := string(game.ZoneGraveyard)
	if pub, p := publicActivatedAbilityRow(ActivatedAbilityView{ExileCostN: 2, ExileCostOptions: []string{"a", "b"}, ExileCostZone: grave}); p || len(pub.ExileCostOptions) != 2 {
		t.Errorf("activated graveyard exile list: kept=%v private=%v, want kept and not private", pub.ExileCostOptions, p)
	}
	if pub, p := publicManaAbilityRow(ManaAbilityView{ExileCostN: 1, ExileCostOptions: []string{"a"}, ExileCostZone: grave}); p || len(pub.ExileCostOptions) != 1 {
		t.Errorf("mana graveyard exile list: kept=%v private=%v, want kept and not private", pub.ExileCostOptions, p)
	}
	if pub, p := publicActivatedAbilityRow(ActivatedAbilityView{ExileCostN: 1, ExileCostOptions: []string{"a"}}); !p || pub.ExileCostOptions != nil {
		t.Errorf("an exile list with no zone survived the strip: %v private=%v", pub.ExileCostOptions, p)
	}

	// And a row with no hidden list reports none, which is what keeps
	// fileAbilityOffers from allocating a per-seat copy for every
	// permanent on the board.
	if _, p := publicActivatedAbilityRow(ActivatedAbilityView{Label: "{T}: Draw", DiscardCostN: 1}); p {
		t.Error("a row with no options reported a private list")
	}
	if _, p := publicManaAbilityRow(ManaAbilityView{Label: "Add {G}", ExileCostN: 1}); p {
		t.Error("a mana row with no options reported a private list")
	}
}

// --- end to end --------------------------------------------------------

// frameCard returns permanent `id` as `viewer`'s frame carries it —
// "" is a spectator. The card is made known to every seat first, which
// is what a battlefield entry does in a real game and a fixture's
// PushTop does not; without it the per-card redaction would clear the
// rows for every viewer and the assertions below would be vacuous.
func frameCard(t *testing.T, g *game.Game, viewer string, id uuid.UUID) CardView {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != id {
				continue
			}
			for _, p := range g.Seats {
				c.AddKnower(p.ID)
			}
		}
	})
	for _, c := range ViewOfGameFor(g, viewer).Battlefield.Cards {
		if c.InstanceID == id.String() {
			if !c.KnownByYou {
				t.Fatalf("permanent %s is not known to viewer %q; the privacy check would be vacuous", id, viewer)
			}
			return c
		}
	}
	t.Fatalf("permanent %s missing from viewer %q's battlefield", id, viewer)
	return CardView{}
}

// controllerFrameCard is frameCard for the permanent's own controller:
// the frame the client picker actually reads its options from.
func controllerFrameCard(t *testing.T, g *game.Game, id uuid.UUID) CardView {
	t.Helper()
	var controller uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			controller = c.Controller
		}
	}
	if controller == uuid.Nil {
		t.Fatalf("permanent %s missing from the battlefield", id)
	}
	return frameCard(t, g, controller.String(), id)
}

// handCardFor seats a hand card of the given type line.
func handCardFor(p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.KnownBy = map[uuid.UUID]bool{p.ID: true}
	p.Hand.PushTop(c)
	return c.InstanceID
}

func seatFaunaShaman(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Fauna Shaman", TypeLine: "Creature — Elf Shaman",
		OracleID: "00000000-0000-0000-0000-000000001369",
		Owner:    owner, Controller: owner, Power: 2, Toughness: 2,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{G}, {T}, Discard a creature card: Search your library for a creature card",
			Cost: game.AbilityCost{
				Mana: "{G}", Tap: true,
				DiscardCards: &game.DiscardCost{
					N: 1, Label: "a creature card",
					Match: func(c game.Card) bool { return c.IsCreature() },
				},
			},
		}},
	})
	return id
}

// privacyFixture is one table: seat 0 controls Fauna Shaman (the
// filtered activated clause), Skirge Familiar (a mana discard) and
// Cadaverous Bloom (a mana exile from hand), and holds two creature
// cards and an instant. Seat 1 is the opponent.
type privacyFixture struct {
	g                     *game.Game
	me, opp               *game.Player
	shaman, skirge, bloom uuid.UUID
	creatures, hand       []string
}

func newPrivacyFixture(t *testing.T) privacyFixture {
	t.Helper()
	g := buildActiveGame(t)
	f := privacyFixture{g: g, me: g.Seats[0], opp: g.Seats[1]}
	g.WithWriteLock(func() {
		f.me.Hand.Cards = nil
		f.shaman = seatFaunaShaman(g, f.me.ID)
		f.skirge = seatDiscardManaSource(g, f.me.ID, &game.DiscardCost{N: 1, Label: "a card"})
		f.bloom = seatExileManaSource(g, f.me.ID, &game.ExileCost{N: 1, Label: "a card"})
		a := handCardFor(f.me, "Hidden Bear", "Creature — Bear")
		b := handCardFor(f.me, "Hidden Elk", "Creature — Elk")
		c := handCardFor(f.me, "Hidden Bolt", "Instant")
		f.creatures = sortedIDs(a, b)
		f.hand = sortedIDs(a, b, c)
	})
	g.BumpLayerVersionForTest()
	return f
}

func sortedIDs(ids ...uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	sort.Strings(out)
	return out
}

// TestHandCostOptionsReachTheControllerAlone is #1369 per field: the
// controller's frame carries each hand-sourced list, and an opponent's
// and a spectator's carry the row, its printed count and its label,
// and no list.
func TestHandCostOptionsReachTheControllerAlone(t *testing.T) {
	f := newPrivacyFixture(t)

	type field struct {
		name string
		src  uuid.UUID
		read func(CardView) (n int, label string, opts []string)
		want []string
	}
	fields := []field{
		{
			name: "activated_abilities[0].discard_cost_options (Fauna Shaman)", src: f.shaman,
			read: func(c CardView) (int, string, []string) {
				a := c.ActivatedAbilities[0]
				return a.DiscardCostN, a.DiscardCostLabel, a.DiscardCostOptions
			},
			// The FILTERED clause: two creature cards of three.
			want: f.creatures,
		},
		{
			name: "mana_abilities[0].discard_cost_options (Skirge Familiar)", src: f.skirge,
			read: func(c CardView) (int, string, []string) {
				m := c.ManaAbilities[0]
				return m.DiscardCostN, m.DiscardCostLabel, m.DiscardCostOptions
			},
			want: f.hand,
		},
		{
			name: "mana_abilities[0].exile_cost_options (Cadaverous Bloom)", src: f.bloom,
			read: func(c CardView) (int, string, []string) {
				m := c.ManaAbilities[0]
				return m.ExileCostN, m.ExileCostLabel, m.ExileCostOptions
			},
			want: f.hand,
		},
	}
	for _, fd := range fields {
		t.Run(fd.name, func(t *testing.T) {
			mine := frameCard(t, f.g, f.me.ID.String(), fd.src)
			n, label, opts := fd.read(mine)
			if n != 1 || label == "" {
				t.Errorf("controller: n=%d label=%q, want the printed clause", n, label)
			}
			if got := sortedCopy(opts); !sameStrings(got, fd.want) {
				t.Errorf("controller: options = %v, want %v", got, fd.want)
			}

			for _, viewer := range []struct{ who, id string }{
				{"opponent", f.opp.ID.String()},
				{"spectator", ""},
			} {
				theirs := frameCard(t, f.g, viewer.id, fd.src)
				n, label, opts := fd.read(theirs)
				if len(opts) != 0 {
					t.Errorf("%s: options = %v, want none — a count and IDs out of another player's hand (#1369)",
						viewer.who, opts)
				}
				// The row itself and its printed half stay: the ability
				// is public, and "Discard a creature card" is on the card.
				if n != 1 || label == "" {
					t.Errorf("%s: n=%d label=%q, want the printed clause kept", viewer.who, n, label)
				}
			}
		})
	}
}

// TestNoHandCardIDReachesAnotherFrame is the same property asked of the
// whole battlefield zone at once, as the wire carries it: not one hand
// card's instance ID appears anywhere in an opponent's or a spectator's
// serialised battlefield, and every one appears in the controller's.
// It is the check a field added to the rows tomorrow cannot slip past
// without also being in the scope tables above.
func TestNoHandCardIDReachesAnotherFrame(t *testing.T) {
	f := newPrivacyFixture(t)
	// Make every permanent known to both seats, as a real entry would.
	for _, id := range []uuid.UUID{f.shaman, f.skirge, f.bloom} {
		frameCard(t, f.g, "", id)
	}
	zoneJSON := func(viewer string) string {
		b, err := json.Marshal(ViewOfGameFor(f.g, viewer).Battlefield)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	mine := zoneJSON(f.me.ID.String())
	for _, id := range f.hand {
		if !strings.Contains(mine, id) {
			t.Errorf("the controller's battlefield does not carry hand card %s; the check below is vacuous", id)
		}
	}
	for _, viewer := range []struct{ who, id string }{{"opponent", f.opp.ID.String()}, {"spectator", ""}} {
		theirs := zoneJSON(viewer.id)
		for _, id := range f.hand {
			if strings.Contains(theirs, id) {
				t.Errorf("%s's battlefield carries hand card %s (#1369)", viewer.who, id)
			}
		}
	}
}

// TestAbilityOffersPromotionIsPerFrame pins the carrier's two
// lifecycle properties: filtering for one viewer does not change what
// the next viewer gets (in either order), and the unfiltered view —
// what the crash-recovery dump marshals — carries only the public rows.
func TestAbilityOffersPromotionIsPerFrame(t *testing.T) {
	f := newPrivacyFixture(t)
	frameCard(t, f.g, "", f.shaman)
	v := ViewOfGame(f.g)
	find := func(gv GameView) CardView {
		for _, c := range gv.Battlefield.Cards {
			if c.InstanceID == f.shaman.String() {
				return c
			}
		}
		t.Fatal("Fauna Shaman missing")
		return CardView{}
	}
	if opts := find(v).ActivatedAbilities[0].DiscardCostOptions; len(opts) != 0 {
		t.Errorf("the unfiltered view's exported row carries the hand list %v", opts)
	}
	for _, order := range [][]string{
		{f.opp.ID.String(), f.me.ID.String(), f.opp.ID.String(), ""},
		{f.me.ID.String(), "", f.me.ID.String()},
	} {
		for _, viewer := range order {
			got := find(FilterViewFor(v, viewer)).ActivatedAbilities[0].DiscardCostOptions
			if viewer == f.me.ID.String() {
				if !sameStrings(sortedCopy(got), f.creatures) {
					t.Errorf("controller after %v: options = %v, want %v", order, got, f.creatures)
				}
			} else if len(got) != 0 {
				t.Errorf("viewer %q after %v: options = %v, want none", viewer, order, got)
			}
		}
	}
}

// TestControllerFrameAndEnumeratorAgreeOnHandCosts is the controller's
// half: the frame their client reads its picker from offers every card
// the enumerator — the bot's move list and the client's timing lookup
// — would pay with, for all three hand-sourced components. Moving the
// list onto a per-seat carrier must not have moved it away from the
// one seat that needs it.
func TestControllerFrameAndEnumeratorAgreeOnHandCosts(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	var shaman, skirge, bloom uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		shaman = seatFaunaShaman(g, me.ID)
		skirge = seatDiscardManaSource(g, me.ID, &game.DiscardCost{N: 1, Label: "a card"})
		bloom = seatExileManaSource(g, me.ID, &game.ExileCost{N: 1, Label: "a card"})
		handCardFor(me, "Pay Bear", "Creature — Bear")
		handCardFor(me, "Pay Bolt", "Instant")
	})
	g.BumpLayerVersionForTest()

	type paid struct{ discards, exiles map[string]bool }
	moves := map[string]*paid{}
	at := func(src string) *paid {
		if moves[src] == nil {
			moves[src] = &paid{discards: map[string]bool{}, exiles: map[string]bool{}}
		}
		return moves[src]
	}
	{
		for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
			var p struct {
				SourceCardID string   `json:"source_card_id"`
				CardID       string   `json:"card_id"`
				DiscardIDs   []string `json:"discard_ids"`
				ExileIDs     []string `json:"exile_ids"`
			}
			if m.Type != legal.TypeActivateAbility && m.Type != legal.TypeActivateManaAbility {
				continue
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			e := at(p.SourceCardID + p.CardID)
			for _, id := range p.DiscardIDs {
				e.discards[id] = true
			}
			for _, id := range p.ExileIDs {
				e.exiles[id] = true
			}
		}
	}

	for _, tc := range []struct {
		name  string
		src   uuid.UUID
		pays  func(*paid) map[string]bool
		shown func(CardView) []string
	}{
		{"Fauna Shaman discard", shaman, func(p *paid) map[string]bool { return p.discards },
			func(c CardView) []string { return c.ActivatedAbilities[0].DiscardCostOptions }},
		{"Skirge Familiar discard", skirge, func(p *paid) map[string]bool { return p.discards },
			func(c CardView) []string { return c.ManaAbilities[0].DiscardCostOptions }},
		{"Cadaverous Bloom exile", bloom, func(p *paid) map[string]bool { return p.exiles },
			func(c CardView) []string { return c.ManaAbilities[0].ExileCostOptions }},
	} {
		offered := moves[tc.src.String()]
		if offered == nil || len(tc.pays(offered)) == 0 {
			t.Errorf("%s: the enumerator offered no payment; the containment check is vacuous", tc.name)
			continue
		}
		shown := map[string]bool{}
		for _, id := range tc.shown(controllerFrameCard(t, g, tc.src)) {
			shown[id] = true
		}
		for id := range tc.pays(offered) {
			if !shown[id] {
				t.Errorf("%s: the enumerator pays with %s, which the controller's frame does not offer", tc.name, id)
			}
		}
	}
}

// TestCatalogHandCostCardsKeepTheirListsForTheController is the same
// property on the printed cards the issue is about, through the catalog
// rather than a fixture shape: Fauna Shaman, Skirge Familiar, Cadaverous
// Bloom and (#1297) Holistic Wisdom, seated for seat 0 over a hand of
// one creature card and one instant, viewed by seat 0, seat 1 and a
// spectator. Grim Lavamancer is the control: its exile list is read off
// seat 0's GRAVEYARD, a public pile, and reaches every viewer.
func TestCatalogHandCostCardsKeepTheirListsForTheController(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seat := func(name, typeLine, oracle string) uuid.UUID {
		c := game.NewCard(name, me.ID)
		c.TypeLine = typeLine
		c.OracleID = oracle
		c.Power, c.Toughness = 2, 2
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	var shaman, skirge, bloom, wisdom, lavamancer uuid.UUID
	var creature, instant, graveA, graveB uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		shaman = seat("Fauna Shaman", "Creature — Elf Shaman", "35b8fa77-4e85-418b-b335-cd1af127075c")
		skirge = seat("Skirge Familiar", "Creature — Phyrexian Imp", "ba95f24d-42da-48ce-bcf1-1b7c4b3c45b5")
		bloom = seat("Cadaverous Bloom", "Enchantment", "fbb0f73b-5e30-4632-99c1-e49582e41f8d")
		wisdom = seat("Holistic Wisdom", "Enchantment", "7e108285-52da-473c-accd-d48e646a49c0")
		lavamancer = seat("Grim Lavamancer", "Creature — Human Wizard", "37445e06-88a1-4e2e-a432-383736c9b977")
		me.Graveyard.Cards = nil
		for _, id := range []*uuid.UUID{&graveA, &graveB} {
			c := game.NewCard("Graveyard Card", me.ID)
			c.TypeLine = "Sorcery"
			me.Graveyard.PushTop(c)
			*id = c.InstanceID
		}
		creature = handCardFor(me, "Hidden Bear", "Creature — Bear")
		instant = handCardFor(me, "Hidden Bolt", "Instant")
	})
	g.BumpLayerVersionForTest()

	// Each card's hand list, read off one frame. The row is found by
	// its cost component rather than by index, so a catalog entry that
	// grows a second ability does not break the test.
	lists := func(c CardView) map[string][]string {
		out := map[string][]string{}
		for _, a := range c.ActivatedAbilities {
			if a.DiscardCostN > 0 {
				out["activated discard"] = sortedCopy(a.DiscardCostOptions)
			}
			if a.ExileCostN > 0 {
				out["activated exile "+a.ExileCostZone] = sortedCopy(a.ExileCostOptions)
			}
		}
		for _, m := range c.ManaAbilities {
			if m.DiscardCostN > 0 {
				out["mana discard"] = sortedCopy(m.DiscardCostOptions)
			}
			if m.ExileCostN > 0 {
				out["mana exile"] = sortedCopy(m.ExileCostOptions)
			}
		}
		return out
	}
	for _, tc := range []struct {
		card string
		id   uuid.UUID
		key  string
		want []string
		// public: the list is read off a public pile and every viewer
		// gets it (Grim Lavamancer's graveyard).
		public bool
	}{
		{"Fauna Shaman", shaman, "activated discard", sortedIDs(creature), false},
		{"Skirge Familiar", skirge, "mana discard", sortedIDs(creature, instant), false},
		{"Cadaverous Bloom", bloom, "mana exile", sortedIDs(creature, instant), false},
		{"Holistic Wisdom", wisdom, "activated exile hand", sortedIDs(creature, instant), false},
		{"Grim Lavamancer", lavamancer, "activated exile graveyard", sortedIDs(graveA, graveB), true},
	} {
		mine := lists(frameCard(t, g, me.ID.String(), tc.id))
		got, ok := mine[tc.key]
		if !ok {
			t.Fatalf("%s: the controller's frame has no %s row; the catalog entry changed shape", tc.card, tc.key)
		}
		if !sameStrings(got, tc.want) {
			t.Errorf("%s: controller's options = %v, want %v", tc.card, got, tc.want)
		}
		for _, viewer := range []struct{ who, id string }{{"opponent", opp.ID.String()}, {"spectator", ""}} {
			theirs := lists(frameCard(t, g, viewer.id, tc.id))
			got, ok := theirs[tc.key]
			if !ok {
				t.Errorf("%s: %s lost the %s row itself; only its hand list should go", tc.card, viewer.who, tc.key)
			}
			if tc.public {
				if !sameStrings(got, tc.want) {
					t.Errorf("%s: %s sees %v, want the public graveyard list %v", tc.card, viewer.who, got, tc.want)
				}
				continue
			}
			if len(got) != 0 {
				t.Errorf("%s: %s sees %v out of the controller's hand (#1369)", tc.card, viewer.who, got)
			}
		}
	}
}
