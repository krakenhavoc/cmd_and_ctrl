package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// huatli_poet_of_unity_test.go pins Huatli, Poet of Unity // Roar of the
// Fifth People (#5402, ADR 0093 amendment 2026-11-10): the front face's
// land tutor and transform, the four chapters, the chapter II grant of a
// granting static (to creatures that enter later too, ended by the Saga
// losing its abilities or leaving), a restore, and the bot path.

func huatliRow() cards.Card {
	row := transformRow(huatliPoetOfUnityOracle, "Huatli, Poet of Unity", "Legendary Creature — Human Warrior Bard", "{2}{G}",
		"Roar of the Fifth People", "Enchantment — Saga", "", "", []string{"G"})
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = "2", "3"
	return row
}

func TestHuatliPoetOfUnityBothFacesAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		huatliPoetOfUnityOracle:        "Huatli, Poet of Unity",
		huatliPoetOfUnityOracle + "#1": "Roar of the Fifth People",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// huatliCast casts Huatli in seat 0's main phase and takes the Forest
// with her enter trigger.
func huatliCast(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pushLibraryCardForTest(me, game.Card{Name: "Nonbasic Vista", TypeLine: "Land"})
	huatli := importAndCast(t, g, huatliRow(), me)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Huatli's enter trigger asked no search")
	}
	if searchOptionNamed(g, c, "Nonbasic Vista") != uuid.Nil {
		t.Error("a nonbasic land is findable")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	passPriorityAroundTable(t, g)
	return huatli
}

// huatliTransform activates the front face's ability and returns the
// Saga, with chapter I's trigger resolved.
func huatliTransform(t *testing.T, g *game.Game, me *game.Player, huatli uuid.UUID) uuid.UUID {
	t.Helper()
	floatForTest(g, me, "CCCRW")
	if err := g.ActivateCatalogAbility(me.ID, huatli, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("transform ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	var saga uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == huatliPoetOfUnityOracle && TransformedPermanent(c) {
				saga = c.InstanceID
			}
		}
	})
	if saga == uuid.Nil {
		t.Fatal("Huatli did not return transformed")
	}
	if c := e2Card(t, g, saga); c.Name != "Roar of the Fifth People" || c.Counters[game.CounterLore] != 1 {
		t.Fatalf("back face %q with %d lore counters, want Roar of the Fifth People on chapter I", c.Name, c.Counters[game.CounterLore])
	}
	passPriorityAroundTable(t, g)
	return saga
}

// roarTable is a game where seat 0's Roar of the Fifth People has just
// resolved chapter II, with chapter I's two Dinosaurs and an opponent's
// creature on the battlefield.
type roarTable struct {
	g          *game.Game
	me, opp    *game.Player
	seat       int
	saga       uuid.UUID
	dinos      []uuid.UUID
	enemy      uuid.UUID
	handBefore int
}

func newRoarTable(t *testing.T) roarTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	seat := g.Turn.ActiveSeat
	huatli := huatliCast(t, g, me)
	saga := huatliTransform(t, g, me, huatli)
	var dinos []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Dinosaur" && c.Controller == me.ID {
				dinos = append(dinos, c.InstanceID)
			}
		}
	})
	if len(dinos) != 2 {
		t.Fatalf("chapter I made %d Dinosaurs, want 2", len(dinos))
	}
	if d := e2Card(t, g, dinos[0]); d.Power != 3 || d.Toughness != 3 || !d.IsCreature() {
		t.Errorf("chapter I token is %d/%d", d.Power, d.Toughness)
	}
	if grantedManaRowsOf(g, dinos[0]) != 0 {
		t.Fatal("a Dinosaur taps for mana before chapter II")
	}
	enemy := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if n := loreCountersOn(g, saga); n != 2 {
		t.Fatalf("lore counters %d, want 2", n)
	}
	return roarTable{g: g, me: me, opp: opp, seat: seat, saga: saga, dinos: dinos, enemy: enemy, handBefore: me.Hand.Size()}
}

func TestHuatliFindsABasicLandAndTransforms(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	huatli := huatliCast(t, g, me)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.Name == "Forest" {
			inHand = true
		}
	}
	if !inHand {
		t.Fatal("the basic land is not in hand")
	}
	if spec, _ := Lookup(huatliPoetOfUnityOracle); len(spec.Activated) != 1 || !spec.Activated[0].SorcerySpeed {
		t.Error("the transform is not a sorcery-speed activation")
	}
	huatliTransform(t, g, me, huatli)
}

func TestRoarOfTheFifthPeopleChapterIIGivesYourCreaturesTheManaAbility(t *testing.T) {
	r := newRoarTable(t)
	g, me := r.g, r.me
	for _, d := range r.dinos {
		if grantedManaRowsOf(g, d) != 1 {
			t.Errorf("Dinosaur %s has no granted mana ability", d)
		}
	}
	if grantedManaRowsOf(g, r.enemy) != 0 {
		t.Error("an opponent's creature has the mana ability")
	}
	if grantedManaRowsOf(g, r.saga) != 0 {
		t.Error("the Saga (not a creature) has the mana ability")
	}
	late := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Late Elk",
		TypeLine: "Creature — Elk", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if grantedManaRowsOf(g, late) != 1 {
		t.Error("a creature that entered after chapter II has no mana ability")
	}
	for i, color := range []string{"R", "G"} {
		if err := tapGrantedMana(t, g, me.ID, r.dinos[i], color); err != nil {
			t.Fatalf("tapping a Dinosaur for {%s}: %v", color, err)
		}
	}
	if err := tapGrantedMana(t, g, me.ID, late, "W"); err != nil {
		t.Fatalf("tapping the Elk for {W}: %v", err)
	}
	for _, color := range []string{"R", "G", "W"} {
		if got := me.ManaPool.CountColor(color); got != 1 {
			t.Errorf("{%s} in pool = %d, want 1", color, got)
		}
	}
	if !b16Tapped(t, g, r.dinos[0]) {
		t.Error("the granted {T} did not tap the Dinosaur")
	}
}

// CR 613.6, 613.8a: the Saga losing all abilities after chapter II takes
// its granted static, and the creatures lose the mana ability with it.
func TestRoarOfTheFifthPeopleLosingItsAbilitiesEndsTheManaAbility(t *testing.T) {
	r := newRoarTable(t)
	g := r.g
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(r.saga),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
	})
	for _, d := range r.dinos {
		if grantedManaRowsOf(g, d) != 0 {
			t.Error("a Dinosaur kept the mana ability after the Saga lost all abilities")
		}
	}
}

// CR 613.7: a creature that loses all abilities after the grant loses the
// mana ability; its neighbour keeps it.
func TestRoarOfTheFifthPeopleRecipientThatLosesItsAbilitiesLosesTheManaAbility(t *testing.T) {
	r := newRoarTable(t)
	g := r.g
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(r.dinos[0]),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
	})
	if grantedManaRowsOf(g, r.dinos[0]) != 0 {
		t.Error("the Dinosaur that lost all abilities kept the mana ability")
	}
	if grantedManaRowsOf(g, r.dinos[1]) != 1 {
		t.Error("the other Dinosaur lost the mana ability")
	}
}

func TestRoarOfTheFifthPeopleChapterIIIFindsADinosaurAndIVThenTheSagaGoes(t *testing.T) {
	r := newRoarTable(t)
	g, me := r.g, r.me
	late := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Late Elk",
		TypeLine: "Creature — Elk", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})

	// III — a Dinosaur card, revealed, to hand; a non-Dinosaur is not findable.
	pushLibraryCardForTest(me, game.Card{Name: "Library Raptor", TypeLine: "Creature — Dinosaur", ManaCost: "{1}{G}"})
	pushLibraryCardForTest(me, game.Card{Name: "Library Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	advanceToPrecombatMainOf(t, g, r.seat)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("chapter III: no search prompt")
	}
	if searchOptionNamed(g, c, "Library Bear") != uuid.Nil {
		t.Error("a non-Dinosaur card is findable")
	}
	answerSearchNamed(t, g, me.ID, "Library Raptor")
	passPriorityAroundTable(t, g)
	found := false
	for _, h := range me.Hand.Cards {
		if h.Name == "Library Raptor" {
			found = true
		}
	}
	if !found {
		t.Fatal("chapter III did not put the Dinosaur card into hand")
	}

	// IV — Dinosaurs you control gain double strike and trample; the Elk
	// does not. The Saga is sacrificed, and the mana ability ends with it.
	advanceToPrecombatMainOf(t, g, r.seat)
	passPriorityAroundTable(t, g)
	for _, d := range r.dinos {
		if !hasKeywordOnBattlefield(t, g, d, "double strike") || !hasKeywordOnBattlefield(t, g, d, "trample") {
			t.Error("a Dinosaur lacks double strike or trample after chapter IV")
		}
	}
	if hasKeywordOnBattlefield(t, g, late, "double strike") || hasKeywordOnBattlefield(t, g, late, "trample") {
		t.Error("a non-Dinosaur gained chapter IV's keywords")
	}
	if findBattlefieldCardByID(g, r.saga) != nil {
		t.Fatal("the Saga is not sacrificed after chapter IV")
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, d := range append([]uuid.UUID{late}, r.dinos...) {
		if grantedManaRowsOf(g, d) != 0 {
			t.Error("a creature kept the mana ability after the Saga left")
		}
	}
}

// The Saga holding its chapter II grant is a restore point, and a
// Dinosaur still taps for mana after a restore.
func TestRoarOfTheFifthPeopleGrantSurvivesARestore(t *testing.T) {
	r := newRoarTable(t)
	restored := restoreRoundTrip(t, r.g, true)
	for _, d := range r.dinos {
		if grantedManaRowsOf(restored, d) != 1 {
			t.Error("restored: a Dinosaur lost the mana ability")
		}
	}
	if err := tapGrantedMana(t, restored, restored.Seats[0].ID, r.dinos[0], "G"); err != nil {
		t.Errorf("restored: tapping a Dinosaur for {G}: %v", err)
	}
}

// The bot path: the enumerator offers the transform with the mana for it,
// and each creature's granted mana ability after chapter II.
func TestHuatliAndRoarOfTheFifthPeopleOfferTheBotTheirMoves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	huatli := huatliCast(t, g, me)
	floatForTest(g, me, "CCCRW")
	offersTransform := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeActivateAbility && m.Source == huatli {
			offersTransform = true
		}
	}
	if !offersTransform {
		t.Error("the enumerator offers no transform for Huatli")
	}

	r := newRoarTable(t)
	offered := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(r.g, r.me.ID) {
		if m.Type == legal.TypeActivateManaAbility {
			offered[m.Source] = true
		}
	}
	for _, d := range r.dinos {
		if !offered[d] {
			t.Errorf("the enumerator offers no mana ability for Dinosaur %s", d)
		}
	}
	if offered[r.saga] {
		t.Error("the enumerator offers a mana ability on the Saga")
	}
}
