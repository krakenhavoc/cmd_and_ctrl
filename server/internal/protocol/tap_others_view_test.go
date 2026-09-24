package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// tap_others_view_test.go — #759: the tap-another cost component
// (#758's TapOthersCost) on the wire, for station (CR 702.184a).
//
// The same #544 property cost_component_coherence_view_test.go asserts
// for the return component: the view's picker, the enumerator's
// payments and the engine's validator read ONE walk
// (game.TapOthersOptionsForEffect), so the client never offers a
// creature the server refuses and the bots are never denied one it
// would accept.

func creatureCostClause() *game.TargetSpec {
	return &game.TargetSpec{
		Mode: "permanent", Label: "another untapped creature you control", Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.IsCreature() },
		Min:    1, Max: 1,
	}
}

// seatTapOthersSource seats a creature whose one activated ability
// costs "tap another untapped creature you control" — plus the {T}
// symbol when alsoTap, which is the CR 118.3 case where the source is
// spent by its own half of the cost.
func seatTapOthersSource(g *game.Game, owner uuid.UUID, alsoTap bool) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Tap Others Source",
		TypeLine:   "Creature — Construct",
		OracleID:   "00000000-0000-0000-0000-0000000007f9",
		Power:      1, Toughness: 1,
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Tap another untapped creature you control: mark",
			Cost: game.AbilityCost{Tap: alsoTap, TapOthers: &game.TapOthersCost{
				Count:         1,
				Filter:        creatureCostClause(),
				ExcludeSource: !alsoTap,
				Label:         "another untapped creature you control",
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	return id
}

// seatTapOthersManaSource is Springleaf Drum's mana-ability spelling
// of the same component. Its clause does not print "another", but the
// source is still unavailable to that half because {T} spends it
// first (CR 118.3).
func seatTapOthersManaSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Springleaf Drum",
		TypeLine:   "Artifact",
		Owner:      owner,
		Controller: owner,
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost: true,
			TapOthers: &game.TapOthersCost{
				Count:  1,
				Filter: creatureCostClause(),
				Label:  "an untapped creature you control",
			},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
	return id
}

func seatTapBearFor(g *game.Game, controller uuid.UUID, name string, tapped bool, keywords ...string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Tapped: tapped,
		Owner: controller, Controller: controller, Keywords: keywords,
	})
	return id
}

// The shape and the options: the clause's label, the untapped
// creatures the seat controls other than the source, and a hexproof
// one among them — paying a cost does not target (CR 601.2h).
func TestActivatedAbilityViewCarriesTapOthersCostAndOptions(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatTapOthersSource(g, owner, false)
	plain := seatTapBearFor(g, owner, "Plain Bear", false)
	hexproof := seatTapBearFor(g, owner, "Hexproof Bear", false, "hexproof")
	seatTapBearFor(g, owner, "Tired Bear", true)          // tapped: cannot pay
	seatTapBearFor(g, g.Seats[1].ID, "Their Bear", false) // an opponent's: cannot pay
	seatArtifactFor(g, owner, "Not A Creature")           // wrong type: cannot pay
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.TapOthersLabel != "another untapped creature you control" {
		t.Errorf("tap_others_label = %q, want the clause as printed", ab.TapOthersLabel)
	}
	if ab.TapOthersOptions == nil {
		t.Fatal("tap_others_options is absent on an ability that prints the clause")
	}
	if ab.TapOthersOptions.Min != 1 || ab.TapOthersOptions.Max != 1 {
		t.Errorf("tap_others_options bounds = %d/%d, want 1/1", ab.TapOthersOptions.Min, ab.TapOthersOptions.Max)
	}
	got := append([]string(nil), ab.TapOthersOptions.Cards...)
	sort.Strings(got)
	want := []string{plain.String(), hexproof.String()}
	sort.Strings(want)
	if !sameStrings(got, want) {
		t.Errorf("tap_others_options = %v, want my two untapped bears (hexproof included, source excluded)", ab.TapOthersOptions.Cards)
	}

	// The JSON names are the contract the client reads.
	raw, err := json.Marshal(ab)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := m["tap_others_options"]; !ok {
		t.Errorf("wire JSON has no tap_others_options: %s", raw)
	}
	if _, ok := m["tap_others_label"]; !ok {
		t.Errorf("wire JSON has no tap_others_label: %s", raw)
	}
}

// CR 118.3 where {T} and the clause meet: a source that the {T} half
// already spends is not offered for the other half, even though the
// clause itself (no "another") would admit it.
func TestTapOthersViewDropsASourceTheTapSymbolSpends(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatTapOthersSource(g, owner, true)
	bear := seatTapBearFor(g, owner, "Plain Bear", false)
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if ab.TapOthersOptions == nil || !sameStrings(ab.TapOthersOptions.Cards, []string{bear.String()}) {
		t.Errorf("tap_others_options = %+v, want only the bear — the source is spent by {T}", ab.TapOthersOptions)
	}
}

// The coherence property: every creature the enumerator pays with is
// one the view offers, and for the one-creature clause the enumerator
// offers every one of them (one move each), so the policy — not the
// enumerator — picks the creature.
func TestTapOthersViewAndEnumeratorAgree(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	src := seatTapOthersSource(g, me.ID, false)
	seatTapBearFor(g, me.ID, "Bear A", false)
	seatTapBearFor(g, me.ID, "Bear B", false, "hexproof")
	seatTapBearFor(g, me.ID, "Tired Bear", true)
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seatTapBearFor(g, them.ID, "Their Bear", false)
	g.BumpLayerVersionForTest()

	view := vehicleView(t, g, src).ActivatedAbilities[0]
	if view.TapOthersOptions == nil || len(view.TapOthersOptions.Cards) == 0 {
		t.Fatal("the view offers no tap options for an ability the enumerator can pay")
	}
	shown := map[string]bool{}
	for _, id := range view.TapOthersOptions.Cards {
		shown[id] = true
	}

	paid := map[string]bool{}
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type != legal.TypeActivateAbility {
			continue
		}
		var p struct {
			SourceCardID string   `json:"source_card_id"`
			TapIDs       []string `json:"tap_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal activate params: %v", err)
		}
		if p.SourceCardID != src.String() {
			continue
		}
		if len(p.TapIDs) != 1 {
			t.Errorf("an activation of the tap-another ability names %d creatures, want 1", len(p.TapIDs))
			continue
		}
		paid[p.TapIDs[0]] = true
	}
	if len(paid) == 0 {
		t.Fatal("the enumerator offered no activation of the tap-another ability the view says is payable")
	}
	for id := range paid {
		if !shown[id] {
			t.Errorf("the enumerator taps %s, which the view does not offer", id)
		}
	}
	for id := range shown {
		if !paid[id] {
			t.Errorf("the view offers %s but the enumerator never taps it", id)
		}
	}
}

// The enumerator's half of the same CR 118.3 rule: with {T} and the
// clause on one ability, no offered move names the source for the
// tap-another half.
func TestTapOthersEnumeratorNeverTapsASourceTheTapSymbolSpends(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	src := seatTapOthersSource(g, me.ID, true)
	seatTapBearFor(g, me.ID, "Bear A", false)
	g.BumpLayerVersionForTest()

	offered := 0
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type != legal.TypeActivateAbility {
			continue
		}
		var p struct {
			SourceCardID string   `json:"source_card_id"`
			TapIDs       []string `json:"tap_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal activate params: %v", err)
		}
		if p.SourceCardID != src.String() {
			continue
		}
		offered++
		for _, id := range p.TapIDs {
			if id == src.String() {
				t.Errorf("the enumerator taps the source for the tap-another half: %s", m.Label)
			}
		}
	}
	if offered == 0 {
		t.Fatal("the enumerator offered no activation, though the bear can pay")
	}
}

func TestManaAbilityViewCarriesTapOthersCostAndAgreesWithEnumerator(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	src := seatTapOthersManaSource(g, me.ID)
	plain := seatTapBearFor(g, me.ID, "Plain Bear", false)
	hexproof := seatTapBearFor(g, me.ID, "Hexproof Bear", false, "hexproof")
	seatTapBearFor(g, me.ID, "Tired Bear", true)
	seatTapBearFor(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Bear", false)
	g.BumpLayerVersionForTest()

	rows := vehicleView(t, g, src).ManaAbilities
	if len(rows) != 1 {
		t.Fatalf("mana ability rows = %d, want 1", len(rows))
	}
	row := rows[0]
	if row.TapOthersLabel != "an untapped creature you control" {
		t.Errorf("tap_others_label = %q", row.TapOthersLabel)
	}
	if row.TapOthersOptions == nil || row.TapOthersOptions.Min != 1 || row.TapOthersOptions.Max != 1 {
		t.Fatalf("tap_others_options = %+v, want a fixed one-card choice", row.TapOthersOptions)
	}
	shown := map[string]bool{}
	for _, id := range row.TapOthersOptions.Cards {
		shown[id] = true
	}
	for _, id := range []uuid.UUID{plain, hexproof} {
		if !shown[id.String()] {
			t.Errorf("view omitted payable creature %s", id)
		}
	}

	paid := map[string]bool{}
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type != legal.TypeActivateManaAbility {
			continue
		}
		var p struct {
			CardID string   `json:"card_id"`
			TapIDs []string `json:"tap_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal mana params: %v", err)
		}
		if p.CardID != src.String() {
			continue
		}
		if len(p.TapIDs) != 1 {
			t.Errorf("mana move names %d tap payments, want 1", len(p.TapIDs))
			continue
		}
		paid[p.TapIDs[0]] = true
	}
	if len(paid) == 0 {
		t.Fatal("enumerator offered no payable Springleaf Drum move")
	}
	for id := range paid {
		if !shown[id] {
			t.Errorf("enumerator taps %s, which the view did not offer", id)
		}
	}
	for id := range shown {
		if !paid[id] {
			t.Errorf("view offers %s, but the enumerator did not", id)
		}
	}
}
