package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// untap_others_view_test.go — #5669: "Untap N tapped creatures you
// control" (game.TapOthersCost.Untap, Halo Fountain) on the wire. It
// rides the tap-another component's label, options and `tap_ids`, with
// `tap_others_untap` for the picker's verb.

// seatUntapOthersSource seats a permanent whose one activated ability
// costs "{T}, Untap a tapped creature you control".
func seatUntapOthersSource(g *game.Game, owner uuid.UUID, typeLine string, tapped bool) uuid.UUID {
	id := uuid.New()
	clause := creatureCostClause()
	clause.Label = "a tapped creature you control"
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Untap Others Source",
		TypeLine:   typeLine,
		OracleID:   "00000000-0000-0000-0000-000000005669",
		Power:      1, Toughness: 1,
		Tapped:     tapped,
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "{T}, Untap a tapped creature you control: mark",
			Cost: game.AbilityCost{Tap: true, TapOthers: &game.TapOthersCost{
				Count:  1,
				Filter: clause,
				Label:  "a tapped creature you control",
				Untap:  true,
			}},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	return id
}

// The options are the TAPPED creatures the seat controls, and the row
// says the clause untaps.
func TestActivatedAbilityViewCarriesAnUntapOthersCost(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatUntapOthersSource(g, owner, "Artifact", false)
	tired := seatTapBearFor(g, owner, "Tired Bear", true)
	hexproof := seatTapBearFor(g, owner, "Tired Hexproof Bear", true, "hexproof")
	seatTapBearFor(g, owner, "Plain Bear", false)        // untapped: cannot pay
	seatTapBearFor(g, g.Seats[1].ID, "Their Bear", true) // an opponent's: cannot pay
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, src).ActivatedAbilities[0]
	if !ab.TapOthersUntap || ab.TapOthersFreesSource {
		t.Errorf("tap_others_untap = %v, tap_others_frees_source = %v; want true, false", ab.TapOthersUntap, ab.TapOthersFreesSource)
	}
	if ab.TapOthersLabel != "a tapped creature you control" {
		t.Errorf("tap_others_label = %q, want the clause as printed", ab.TapOthersLabel)
	}
	if ab.TapOthersOptions == nil || ab.TapOthersOptions.Min != 1 || ab.TapOthersOptions.Max != 1 {
		t.Fatalf("tap_others_options = %+v, want bounds 1/1", ab.TapOthersOptions)
	}
	got := append([]string(nil), ab.TapOthersOptions.Cards...)
	sort.Strings(got)
	want := []string{tired.String(), hexproof.String()}
	sort.Strings(want)
	if !sameStrings(got, want) {
		t.Errorf("tap_others_options = %v, want my two tapped bears", ab.TapOthersOptions.Cards)
	}
	raw, err := json.Marshal(ab)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if m["tap_others_untap"] != true {
		t.Errorf("wire JSON has no tap_others_untap: %s", raw)
	}

	// The tap form never carries the flag.
	plain := vehicleView(t, g, seatTapOthersSource(g, owner, false)).ActivatedAbilities[0]
	if plain.TapOthersUntap || plain.TapOthersFreesSource {
		t.Error("a tap-another cost is marked as an untap cost")
	}
}

// CR 601.2h: a creature source pays its own untap clause beside its
// {T}. Untapped it is one of the options; tapped it is too, and the row
// says naming it frees the {T}; with a stun counter it is neither.
func TestUntapOthersViewOffersACreatureSource(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	upright := seatUntapOthersSource(g, owner, "Artifact Creature — Golem", false)
	g.BumpLayerVersionForTest()
	ab := vehicleView(t, g, upright).ActivatedAbilities[0]
	if ab.TapOthersOptions == nil || !sameStrings(ab.TapOthersOptions.Cards, []string{upright.String()}) || ab.TapOthersFreesSource {
		t.Errorf("untapped creature source: options %+v, frees %v; want itself, false", ab.TapOthersOptions, ab.TapOthersFreesSource)
	}

	g = buildActiveGame(t)
	owner = g.Seats[0].ID
	tapped := seatUntapOthersSource(g, owner, "Artifact Creature — Golem", true)
	g.BumpLayerVersionForTest()
	ab = vehicleView(t, g, tapped).ActivatedAbilities[0]
	if ab.TapOthersOptions == nil || !sameStrings(ab.TapOthersOptions.Cards, []string{tapped.String()}) || !ab.TapOthersFreesSource {
		t.Errorf("tapped creature source: options %+v, frees %v; want itself, true", ab.TapOthersOptions, ab.TapOthersFreesSource)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tapped, game.CounterStun, 1) })
	g.BumpLayerVersionForTest()
	ab = vehicleView(t, g, tapped).ActivatedAbilities[0]
	if ab.TapOthersOptions == nil || len(ab.TapOthersOptions.Cards) != 0 || ab.TapOthersFreesSource {
		t.Errorf("stunned tapped creature source: options %+v, frees %v; want none, false", ab.TapOthersOptions, ab.TapOthersFreesSource)
	}
}

// The #544 property for the untap form: the enumerator untaps exactly
// the creatures the view offers, and the engine accepts each move.
func TestUntapOthersViewAndEnumeratorAgree(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	src := seatUntapOthersSource(g, me.ID, "Artifact Creature — Golem", false)
	seatTapBearFor(g, me.ID, "Tired A", true)
	seatTapBearFor(g, me.ID, "Tired B", true)
	seatTapBearFor(g, me.ID, "Plain Bear", false)
	g.BumpLayerVersionForTest()

	view := vehicleView(t, g, src).ActivatedAbilities[0]
	shown := map[string]bool{}
	for _, id := range view.TapOthersOptions.Cards {
		shown[id] = true
	}
	if len(shown) != 3 {
		t.Fatalf("the view offers %v, want the two tapped bears and the source", view.TapOthersOptions.Cards)
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
			t.Errorf("an activation names %d creatures, want 1", len(p.TapIDs))
			continue
		}
		paid[p.TapIDs[0]] = true
	}
	for id := range shown {
		if !paid[id] {
			t.Errorf("the view offers %s but the enumerator never untaps it", id)
		}
	}
	for id := range paid {
		if !shown[id] {
			t.Errorf("the enumerator untaps %s, which the view does not offer", id)
		}
		if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{uuid.MustParse(id)}}); err != nil {
			t.Errorf("the engine refuses the enumerated payment %s: %v", id, err)
		}
		break
	}
}
