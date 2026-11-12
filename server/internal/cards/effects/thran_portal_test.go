package effects

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const thranPortalOracle = "926ce6a2-7bdd-4380-ac65-bc902ba0c284"

// playThranPortal plays a Thran Portal from the active player's hand
// and names `landType` for it.
func playThranPortal(t *testing.T, g *game.Game, landType string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Thran Portal", TypeLine: "Land — Gate",
		OracleID: thranPortalOracle, Owner: me.ID, Controller: me.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play Thran Portal: %v", err)
	}
	pick := latestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no basic land type prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == landType {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("%s is not offered: %+v", landType, pick.PickOptions)
	}
	answerOptionPick(t, g, me.ID, idx)
	return id
}

// thranPortalRows is the land's mana abilities as (produced, life cost).
func thranPortalRows(t *testing.T, g *game.Game, id uuid.UUID) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, ab := range game.ManaAbilitiesForCard(layeredCard(t, g, id)) {
		out[ab.Produced] = ab.LifeCost
	}
	return out
}

func TestThranPortalIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(thranPortalOracle)
	if !ok || spec.Name != "Thran Portal" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Thran Portal: registered %v, %+v", ok, spec)
	}
}

// With two or fewer other lands it enters untapped, is the chosen type
// as well as a Gate, and taps for that type's mana at 1 life.
func TestThranPortalEntersUntappedIsTheChosenTypeAndTapsForOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedPermanentFor(g, me.ID, "Wastes A", "Land")
	seedPermanentFor(g, me.ID, "Wastes B", "Land")
	id := playThranPortal(t, g, "Island")

	card, ok := battlefieldCard(g, id)
	if !ok || card.Tapped {
		t.Fatalf("on the battlefield %v, tapped %v; want untapped with two other lands", ok, card.Tapped)
	}
	subtypes := effectiveSubtypes(t, g, id)
	sort.Strings(subtypes)
	if !reflect.DeepEqual(subtypes, []string{"Gate", "Island"}) {
		t.Errorf("subtypes %v, want [Gate Island]", subtypes)
	}
	if got := thranPortalRows(t, g, id); !reflect.DeepEqual(got, map[string]int{"{U}": 1}) {
		t.Fatalf("mana abilities %v, want {U} for 1 life", got)
	}

	life := me.Life
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d: the mana ability costs 1 life", me.Life, life-1)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "U" {
		t.Errorf("pool %+v, want one {U}", me.ManaPool)
	}
	if card, _ := battlefieldCard(g, id); !card.Tapped {
		t.Error("the Portal did not tap")
	}
}

// The third other land makes it enter tapped.
func TestThranPortalEntersTappedWithThreeOtherLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, n := range []string{"Wastes A", "Wastes B", "Wastes C"} {
		seedPermanentFor(g, me.ID, n, "Land")
	}
	// An opponent's lands are not "other lands you control".
	seedPermanentFor(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Wastes", "Land")
	id := playThranPortal(t, g, "Forest")
	if card, ok := battlefieldCard(g, id); !ok || !card.Tapped {
		t.Fatalf("on the battlefield %v, tapped %v; want tapped with three other lands", ok, card.Tapped)
	}
}

// CR 119.4: the life is a cost. A player who cannot pay it cannot tap
// the land for mana, and nothing is paid or tapped.
func TestThranPortalCannotBeTappedByAPlayerWhoCannotPayTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playThranPortal(t, g, "Swamp")
	g.WithWriteLock(func() { me.Life = 0 })
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("ActivateManaAbility at 0 life: %v, want ErrInvalidParam", err)
	}
	if card, _ := battlefieldCard(g, id); card.Tapped || len(me.ManaPool) != 0 || me.Life != 0 {
		t.Errorf("a refused activation tapped %v, pool %+v, life %d", card.Tapped, me.ManaPool, me.Life)
	}
}

// "Mana abilities of this land" covers one another effect grants it
// (Chromatic Lantern), and no other land's.
func TestThranPortalChargesAGrantedManaAbilityAndNoOtherLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := seedPermanentFor(g, me.ID, "Forest", "Basic Land — Forest")
	id := playThranPortal(t, g, "Plains")
	pushCatalogPermanent(g, me.ID, "Chromatic Lantern", "Artifact", gaChromaticLanternOracle, false)

	rows := game.ManaAbilitiesForCard(layeredCard(t, g, id))
	if len(rows) != 2 {
		t.Fatalf("%d mana abilities, want the Plains one and the Lantern's: %+v", len(rows), rows)
	}
	for _, ab := range rows {
		if ab.LifeCost != 1 {
			t.Errorf("%q costs %d life, want 1", ab.Label, ab.LifeCost)
		}
	}
	for _, ab := range game.ManaAbilitiesForCard(layeredCard(t, g, forest)) {
		if ab.LifeCost != 0 {
			t.Errorf("the Forest's %q costs %d life, want 0", ab.Label, ab.LifeCost)
		}
	}
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, id, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// CR 305.7: under Magus of the Moon the Portal is a Mountain with none
// of its own abilities, the surcharge among them.
func TestThranPortalUnderMagusOfTheMoonIsAFreeMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playThranPortal(t, g, "Island")
	pushCatalogPermanent(g, me.ID, "Magus of the Moon", "Creature — Human Wizard", ltMagusOfTheMoonOracle, false)
	if got := thranPortalRows(t, g, id); !reflect.DeepEqual(got, map[string]int{"{R}": 0}) {
		t.Fatalf("mana abilities %v, want {R} for no life", got)
	}
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// The auto-tapper prefers a free land, and when it books the Portal it
// says the payment costs a life and pays it.
func TestThranPortalIsAutoTappedLastAndPaysItsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := seedPermanentFor(g, me.ID, "Island", "Basic Land — Island")
	id := playThranPortal(t, g, "Island")

	plan, ok := previewFor(t, g, me.ID, "{U}")
	if !ok || !planHas(plan, island) || planHas(plan, id) {
		t.Fatalf("plan for {U}: %+v (ok %v), want the Island alone", plan, ok)
	}
	plan, ok = previewFor(t, g, me.ID, "{U}{U}")
	if !ok || !planHas(plan, id) {
		t.Fatalf("plan for {U}{U}: %+v (ok %v), want both lands", plan, ok)
	}
	for _, e := range plan {
		want := 0
		if e.CardID == id {
			want = 1
		}
		if e.Life != want {
			t.Errorf("%s: plan says %d life, want %d", e.Name, e.Life, want)
		}
	}

	life := me.Life
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Counterspell", TypeLine: "Instant", ManaCost: "{U}{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{AutoTap: true}); err != nil {
		t.Fatalf("cast with auto-tap: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d: one activation of the Portal", me.Life, life-1)
	}
}

// The choice is on the permanent, so a restored game has the same land:
// the chosen type, its mana ability and the life it costs.
func TestThranPortalSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	id := playThranPortal(t, g, "Mountain")
	r := restoreRoundTrip(t, g, true)
	me := r.Seats[r.Turn.ActiveSeat]
	if got := thranPortalRows(t, r, id); !reflect.DeepEqual(got, map[string]int{"{R}": 1}) {
		t.Fatalf("restored mana abilities %v, want {R} for 1 life", got)
	}
	life := me.Life
	if err := r.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility after restore: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}
