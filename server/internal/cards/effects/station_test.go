package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// station_test.go is the station half of ADR 0071 (#759): charge
// counters as a threshold, and the Spacecraft that becomes a creature
// when it crosses one. The Class and Case half is
// designations_test.go, and the engine contract behind both is
// game/designations_test.go.

const theSeriemaOracle = "a6bcec1f-f515-4e63-9e84-8eb04cc582ff"

// TestSeriemaBecomesACreatureAtSevenChargeCounters is the station
// threshold end to end, across three layers, through the one gate.
func TestSeriemaBecomesACreatureAtSevenChargeCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "The Seriema",
		TypeLine:   "Legendary Artifact — Spacecraft",
		OracleID:   theSeriemaOracle,
		Power:      5,
		Toughness:  5,
		Owner:      me.ID,
		Controller: me.ID,
	})

	before := layeredCard(t, g, id)
	if before.IsCreature() {
		t.Error("a Spacecraft with no charge counters is not a creature")
	}
	if game.HasKeyword(&before, "flying") {
		t.Error("the 7+ flying line must not be live below the threshold")
	}

	if err := g.AddCounter(id, game.CounterCharge, 6); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if six := layeredCard(t, g, id); six.IsCreature() {
		t.Error("six charge counters is below 7+")
	}

	if err := g.AddCounter(id, game.CounterCharge, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	at7 := layeredCard(t, g, id)
	if !at7.IsCreature() {
		t.Error("at 7+ the Spacecraft is an artifact creature (CR 721.2b)")
	}
	if !at7.IsArtifact() {
		t.Error("'in addition to its other types' — it is still an artifact")
	}
	if !game.HasKeyword(&at7, "flying") {
		t.Error("at 7+ it has flying")
	}
	if at7.CurrentPower() != 5 || at7.CurrentToughness() != 5 {
		t.Errorf("base P/T at 7+ = %d/%d, want 5/5", at7.CurrentPower(), at7.CurrentToughness())
	}

	// CR 721.2a is an "as long as": take a counter off and it stops.
	if err := g.AddCounter(id, game.CounterCharge, -1); err != nil {
		t.Fatalf("AddCounter(-1): %v", err)
	}
	after := layeredCard(t, g, id)
	if after.IsCreature() || game.HasKeyword(&after, "flying") {
		t.Error("below the threshold again, the 7+ abilities go off")
	}
}

// TestSeriemaGrantsIndestructibleToOtherTappedLegends — the line
// printed ABOVE the threshold bar, which has no gate and is on from
// the moment the Spacecraft lands.
func TestSeriemaGrantsIndestructibleToOtherTappedLegends(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "The Seriema",
		TypeLine:   "Legendary Artifact — Spacecraft",
		OracleID:   theSeriemaOracle,
		Power:      5, Toughness: 5,
		Owner: me.ID, Controller: me.ID,
	})
	legend := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Some Legend",
		TypeLine:   "Legendary Creature — Human",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	untapped := layeredCard(t, g, legend)
	if game.HasKeyword(&untapped, "indestructible") {
		t.Error("an UNTAPPED legend gets nothing")
	}
	if err := g.TapCard(legend, true); err != nil {
		t.Fatalf("TapCard: %v", err)
	}
	tapped := layeredCard(t, g, legend)
	if !game.HasKeyword(&tapped, "indestructible") {
		t.Error("a tapped legendary creature you control has indestructible")
	}
}

// TestSeriemaIsCompleteWithItsStationAbility — #759 gave The Seriema
// the station ability it was caveated for, so the card is whole: one
// activated ability, the station one, with #758's tap-another cost
// and sorcery timing. station_ability_test.go plays it.
func TestSeriemaIsCompleteWithItsStationAbility(t *testing.T) {
	spec, ok := Lookup(theSeriemaOracle)
	if !ok {
		t.Fatal("The Seriema is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("want CompletenessFull with no caveats, got %s with %d", spec.Completeness, len(spec.Caveats))
	}
	if len(spec.Activated) != 1 {
		t.Fatalf("The Seriema declares %d activated abilities, want 1 (station)", len(spec.Activated))
	}
	st := spec.Activated[0]
	if st.Label != StationLabel || !st.SorcerySpeed || st.ActiveWhen.IsGate() {
		t.Errorf("station = %q sorcery=%v gated=%v — want the keyword, sorcery speed, and no threshold (CR 721.4)",
			st.Label, st.SorcerySpeed, st.ActiveWhen.IsGate())
	}
	tc := st.Cost.TapOthers
	if tc.Empty() || tc.Count != 1 || !tc.ExcludeSource {
		t.Errorf("station's cost = %+v, want one ANOTHER creature", tc)
	}
}

// TestEveryDesignationKindIsProvenByACard lives here rather than with
// the Class and Case tests because charge counters are the LAST of
// the three built kinds to acquire a card, so this is the first place
// the sweep can pass. It keeps specDesignations honest: that walk is
// what the boot check and the door-gate sweep both use, so a slot it
// forgot would be a slot nothing covers.
func TestEveryDesignationKindIsProvenByACard(t *testing.T) {
	kinds := map[game.DesignationKind]bool{}
	for _, spec := range All() {
		for _, d := range specDesignations(spec) {
			if d.IsGate() {
				kinds[d.Kind] = true
			}
		}
	}
	for _, want := range []game.DesignationKind{
		game.DesignationClassLevel,
		game.DesignationCaseSolved,
		game.DesignationChargeCounters,
	} {
		if !kinds[want] {
			t.Errorf("no registered card gates on designation kind %d — the catalog side of ADR 0071 is not proven", want)
		}
	}
}
