package effects

import (
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// producible_mana_test.go — CR 106.7, "the type of mana a permanent
// could produce", once it reads a land's CHOSEN colour (#782).
//
// Two bugs, opposite directions, one function. For an imported
// chosen-colour land the old reader unioned Scryfall's
// `produced_mana`, which lists all five colours for every one of
// them, so an Exotic Orchard facing an opponent's Thriving Isle that
// chose red could tap for white, black or green — stronger than
// printed, the direction #259 refuses. For a catalog-built card with
// no Scryfall data it returned nothing at all, because the whole mana
// ability of a Thriving Isle is a ProducedFunc and every ProducedFunc
// was skipped.

const (
	thrivingIsleOracle   = "69fc70b8-b143-4662-ac95-e2743037239d"
	unchartedHavenOracle = "d23c3613-bc5e-4fc5-939c-62a090c53a79"
)

// allFiveColors is what Scryfall stamps on every chosen-colour land —
// checked in the dump for Thriving Isle, Sea Gate and Uncharted Haven.
var allFiveColors = []string{"W", "U", "B", "R", "G"}

// seedChosenColorLand puts a catalog chosen-colour land on the
// battlefield the way deck import does — with Scryfall's all-five
// `produced_mana` — and answers its "choose a color" prompt when
// `color` is non-empty. An empty `color` leaves the land unchosen,
// which is a real state: it enters tapped and nobody can act before
// the prompt is answered, but CR 106.7 still has an answer for it.
func seedChosenColorLand(t *testing.T, g *game.Game, owner uuid.UUID, name, oracleID, color string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID:   id,
		Name:         name,
		TypeLine:     "Land",
		OracleID:     oracleID,
		ProducedMana: allFiveColors,
		Owner:        owner,
		Controller:   owner,
	})
	if color != "" {
		g.WithWriteLock(func() { g.QueueColorChoiceForEffect(owner, id, name, nil) })
		answerColor(t, g, owner, color)
	}
	return id
}

// producible is CR 106.7's answer for one battlefield card, read the
// way the derivations read it.
func producible(g *game.Game, id uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out = g.ProducibleManaLocked(c)
				return
			}
		}
	})
	return out
}

// pickOptions is the colour list an activated mana ability offers.
func pickOptions(t *testing.T, g *game.Game, controller uuid.UUID, source uuid.UUID) []string {
	t.Helper()
	if err := g.ActivateManaAbility(controller, source, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := manaPickFor(g, controller); pick != nil {
		return pick.ColorOptions
	}
	// A single-option slot drops straight into the pool rather than
	// queuing a pick.
	var out []string
	for _, tok := range playerOf(g, controller).ManaPool {
		out = append(out, tok.Color)
	}
	return out
}

func playerOf(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// --- the chosen colour is the answer ------------------------------

// TestProducibleReadsTheChosenColour — CR 106.7 for a chosen-colour
// land, with Scryfall's all-five array sitting on the card. The array
// must not win.
func TestProducibleReadsTheChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, oracle, chosen string
		want                 []string
	}{
		{"Thriving Isle", thrivingIsleOracle, "R", []string{"U", "R"}},
		{"Thriving Isle, unchosen", thrivingIsleOracle, "", []string{"U"}},
		{"Uncharted Haven", unchartedHavenOracle, "G", []string{"G"}},
		{"Uncharted Haven, unchosen", unchartedHavenOracle, "", nil},
	} {
		name, oracle := tc.name, tc.oracle
		if i := strings.Index(name, ","); i >= 0 {
			name = name[:i]
		}
		id := seedChosenColorLand(t, g, me.ID, name, oracle, tc.chosen)
		got := producible(g, id)
		sort.Strings(got)
		want := append([]string(nil), tc.want...)
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: could produce %v, want %v", tc.name, got, tc.want)
		}
	}
}

// TestExoticOrchardReadsAnOpponentsChosenColour — the headline. The
// printed card allows blue or red off a Thriving Isle that chose red;
// before #782 it allowed all five.
func TestExoticOrchardReadsAnOpponentsChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	orchard := seedPermanentWithOracle(g, me.ID, "Exotic Orchard", "Land", exoticOrchardOracle)
	seedChosenColorLand(t, g, them.ID, "Thriving Isle", thrivingIsleOracle, "R")

	got := append([]string(nil), pickOptions(t, g, me.ID, orchard)...)
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"R", "U"}) {
		t.Errorf("Exotic Orchard offers %v, want exactly {U, R} — the Isle's printed blue and its chosen red", got)
	}
}

// TestExoticOrchardSeesNothingFromAnUnchosenHaven — CR 106.7 on a
// land with no colour chosen yet: it would produce nothing, so the
// Orchard derives nothing from it. Scryfall's five colours must not
// leak through.
func TestExoticOrchardSeesNothingFromAnUnchosenHaven(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	orchard := seedPermanentWithOracle(g, me.ID, "Exotic Orchard", "Land", exoticOrchardOracle)
	seedChosenColorLand(t, g, them.ID, "Uncharted Haven", unchartedHavenOracle, "")

	if err := g.ActivateManaAbility(me.ID, orchard, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := manaPickFor(g, me.ID); pick != nil {
		t.Errorf("a colour pick was queued with options %v; an unchosen Haven could produce nothing", pick.ColorOptions)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// TestExoticOrchardStillReadsAPlainBasic — the ordinary case is
// unchanged, which is what makes the two above a narrowing rather
// than a break.
func TestExoticOrchardStillReadsAPlainBasic(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	orchard := seedPermanentWithOracle(g, me.ID, "Exotic Orchard", "Land", exoticOrchardOracle)
	seedManaLand(g, them.ID, "Forest", "Basic Land — Forest", "G")

	got := pickOptions(t, g, me.ID, orchard)
	if len(got) != 1 || got[0] != "G" {
		t.Errorf("Exotic Orchard offers %v off an opposing Forest, want {G}", got)
	}
}

// --- the commander-identity tri-state (#875) ----------------------

// TestReflectingPoolReadsCommandTowerThroughTheIdentity — Command
// Tower "could produce" exactly the identity's colours, because that
// is exactly what a tap would offer (manaPickOptions, CR 903.4f).
func TestReflectingPoolReadsCommandTowerThroughTheIdentity(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveImportedCommander(g, me, "Two-Colour Commander", "{U}{B}", []string{"U", "B"})
	pool := seedPermanentWithOracle(g, me.ID, "Reflecting Pool", "Land", reflectingPoolOracle)
	seedPermanentWithOracle(g, me.ID, "Command Tower", "Land", commandTowerOracle)

	got := append([]string(nil), pickOptions(t, g, me.ID, pool)...)
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"B", "U"}) {
		t.Errorf("Reflecting Pool offers %v beside a Command Tower under a UB commander, want exactly {U, B}", got)
	}
}

// TestReflectingPoolSeesNothingFromATowerWithNoIdentity — CR 903.4f:
// a colourless commander leaves the Tower adding nothing, so there is
// nothing for the Pool to copy either.
func TestReflectingPoolSeesNothingFromATowerWithNoIdentity(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	riderGiveImportedCommander(g, me, "Colourless Commander", "{8}", nil)
	pool := seedPermanentWithOracle(g, me.ID, "Reflecting Pool", "Land", reflectingPoolOracle)
	seedPermanentWithOracle(g, me.ID, "Command Tower", "Land", commandTowerOracle)

	if err := g.ActivateManaAbility(me.ID, pool, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := manaPickFor(g, me.ID); pick != nil {
		t.Errorf("a colour pick was queued with options %v; a Tower that adds no mana could produce nothing", pick.ColorOptions)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// --- the fallback, and the guard ----------------------------------

// TestProducibleFallsBackToScryfallOnlyWithoutACatalogAnswer — an
// imported land the catalog has never heard of, with no basic land
// type, has nothing BUT `produced_mana`.
func TestProducibleFallsBackToScryfallOnlyWithoutACatalogAnswer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	triome := seedManaLand(g, me.ID, "Raugrin Triome", "Land — Triome", "U", "R", "W")
	got := producible(g, triome)
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"R", "U", "W"}) {
		t.Errorf("an imported land with no spec could produce %v, want its produced_mana {W,U,R}", got)
	}

	// And the fallback must NOT fire beside a catalog answer: the
	// Thriving Isle above carries all five in `produced_mana` and
	// answers with two.
	isle := seedChosenColorLand(t, g, me.ID, "Thriving Isle", thrivingIsleOracle, "B")
	got = producible(g, isle)
	sort.Strings(got)
	if !reflect.DeepEqual(got, []string{"B", "U"}) {
		t.Errorf("a catalog land could produce %v, want {U,B} — Scryfall's five overrode the catalog answer", got)
	}
}

// TestProducibleReadsATokenLandsProducedFunc — the other direction of
// the old bug: a card built from the catalog with no Scryfall record
// at all. Its whole mana ability is a ProducedFunc, and it used to
// contribute nothing.
func TestProducibleReadsATokenLandsProducedFunc(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Demo Chosen Land",
		TypeLine:   "Land",
		Owner:      me.ID,
		Controller: me.ID,
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost:      true,
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
	})
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(me.ID, id, "Demo Chosen Land", nil) })
	answerColor(t, g, me.ID, "R")

	got := producible(g, id)
	if len(got) != 1 || got[0] != "R" {
		t.Errorf("a token land with a ProducedFunc could produce %v, want {R}", got)
	}
}

// TestTwoExoticOrchardsStillSeeNothing — CR 106.6b. The guard is the
// reason DerivesFromOtherSources exists; evaluating every OTHER
// ProducedFunc must not have opened the recursion back up.
func TestTwoExoticOrchardsStillSeeNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	mine := seedPermanentWithOracle(g, me.ID, "Exotic Orchard", "Land", exoticOrchardOracle)
	seedPermanentWithOracle(g, them.ID, "Exotic Orchard", "Land", exoticOrchardOracle)
	// A Reflecting Pool on the far side is the other half of the
	// circle, and an imported one carries Scryfall's five colours —
	// which must not leak through the guard either.
	pool := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: pool, Name: "Reflecting Pool", TypeLine: "Land",
		OracleID: reflectingPoolOracle, ProducedMana: allFiveColors,
		Owner: them.ID, Controller: them.ID,
	})

	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if pick := manaPickFor(g, me.ID); pick != nil {
		t.Errorf("a colour pick was queued with options %v; two Orchards and a Pool see nothing (CR 106.6b)", pick.ColorOptions)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want empty", me.ManaPool)
	}
}

// TestProducibleIsAPureRead — CR 106.7 asks a question; it must not
// tap, spend or queue anything, which is also what makes it undo-safe.
func TestProducibleIsAPureRead(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	riderGiveImportedCommander(g, me, "Two-Colour Commander", "{U}{B}", []string{"U", "B"})
	seedChosenColorLand(t, g, them.ID, "Thriving Isle", thrivingIsleOracle, "R")
	seedPermanentWithOracle(g, me.ID, "Command Tower", "Land", commandTowerOracle)
	before := g.Clone()

	for _, c := range g.Battlefield.Cards {
		if got := producible(g, c.InstanceID); got == nil && c.Name == "Command Tower" {
			t.Errorf("Command Tower could produce nothing")
		}
	}

	after := g.Clone()
	if !reflect.DeepEqual(before.Battlefield.Cards, after.Battlefield.Cards) {
		t.Errorf("reading CR 106.7 changed the battlefield")
	}
	if len(g.PendingChoices) != len(before.PendingChoices) {
		t.Errorf("reading CR 106.7 queued a choice")
	}
}

// TestDerivedManaAbilitiesDeclareTheGuard — the catalog lint. A
// ProducedFunc built from ProducedFromOpponentLands or
// ProducedFromOwnLands reads what OTHER permanents could produce and
// MUST be marked DerivesFromOtherSources, or CR 106.7's reader calls
// back into it and two of them recurse until the stack runs out. The
// reverse holds too: nothing else may claim the guard, because a
// guarded ability contributes nothing to a derivation.
func TestDerivedManaAbilitiesDeclareTheGuard(t *testing.T) {
	derivedConstructors := map[string]bool{
		"ProducedFromOpponentLands": true,
		"ProducedFromOwnLands":      true,
	}
	for _, s := range All() {
		for i, a := range s.ManaAbilities {
			from := constructorOf(a.ProducedFunc)
			derives := derivedConstructors[from]
			switch {
			case derives && !a.DerivesFromOtherSources:
				t.Errorf("%s mana ability %d builds its ProducedFunc from %s, which reads other permanents' producible mana, but does not set DerivesFromOtherSources — CR 106.7's reader will recurse into it",
					s.Name, i, from)
			case !derives && a.DerivesFromOtherSources:
				t.Errorf("%s mana ability %d sets DerivesFromOtherSources (ProducedFunc built by %q) without deriving from other sources; the guard would drop it from every derivation for nothing",
					s.Name, i, from)
			}
		}
	}
}

// constructorOf names the function a closure was built by —
// "ProducedFromOwnLands" for the value ProducedFromOwnLands()
// returns. Every closure a constructor makes shares its code pointer,
// so this identifies the SHAPE rather than the instance. The runtime
// name is dotted and carries an `init.N` segment for a closure built
// inside a package init, so the LAST named segment before the
// `funcN` tail is the one that matters:
//
//	…/effects.init.430.ProducedFromOpponentLands.func1
func constructorOf(fn func(*game.Game, uuid.UUID, uuid.UUID) string) string {
	if fn == nil {
		return ""
	}
	f := runtime.FuncForPC(reflect.ValueOf(fn).Pointer())
	if f == nil {
		return ""
	}
	name := f.Name()
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	parts := strings.Split(name, ".")
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if p == "" || strings.HasPrefix(p, "func") || strings.HasPrefix(p, "init") {
			continue
		}
		if i == 0 {
			// Only the package name is left: a plain named function,
			// not a closure.
			break
		}
		return p
	}
	return name
}
