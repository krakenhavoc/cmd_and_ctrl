package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b14_cards_test.go — Kamigawa: Neon Dynasty, S82 batch 14 (#5493).
// Each test plays the real catalog entry and asserts what its text says.

const (
	neoWretchedSorrowOracle = "6e638332-fba1-4e91-9237-fdff23826640"
	neoBurgeoningOracle     = "a81fd70c-c2bc-4863-97a0-facde86f024c"
	neoOtherworldlyOracle   = "ed06e421-480c-40c8-8c3e-1e466e5327b2"
	neoRecklessJoyOracle    = "3b3ecaf2-7e95-4b62-9bc0-a61985508808"
	neoInvokeCalamityOracle = "5b3f041e-ad4a-47ea-bdc4-1be2353f2e18"
	neoInvokeJusticeOracle  = "3e5cd60f-8fef-4152-a957-0a803997b3da"
	neoInvokeAncientsOracle = "868b3c65-d753-4f32-b6ca-31e58782ea92"
	neoScrapWelderOracle    = "4ebce651-ef95-4a96-b720-a23757d641c8"
	neoTameshiOracle        = "b47b3cfe-7005-4b33-9d41-83b60808e12b"
)

func TestS82NeoB14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		neoWretchedSorrowOracle: "March of Wretched Sorrow",
		neoBurgeoningOracle:     "March of Burgeoning Life",
		neoOtherworldlyOracle:   "March of Otherworldly Light",
		neoRecklessJoyOracle:    "March of Reckless Joy",
		neoInvokeCalamityOracle: "Invoke Calamity",
		neoInvokeJusticeOracle:  "Invoke Justice",
		neoInvokeAncientsOracle: "Invoke the Ancients",
		neoScrapWelderOracle:    "Scrap Welder",
		neoTameshiOracle:        "Tameshi, Reality Architect",
	}
	for oracle, name := range want {
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

func neoMana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func neoCast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, name, typeLine, cost, oracle, params)
}

func neoCardTarget(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- March of Wretched Sorrow ---

func TestS82NeoB14WretchedSorrowDamagesAndGainsXForTheDiscountedPrice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := apcB1Push(g, opp.ID, "Victim", "Creature — Ogre", "", "{3}", 5, 5)
	black := pushHandCardWithManaCost(me, "Black Card", "Creature — Zombie", "{1}{B}")
	advanceToMain(t, g)
	neoMana(me, "B", "B")
	life := me.Life
	// X = 3, one black card exiled: {3}{B} - {2} = {1}{B}.
	spell, err := neoCast(t, g, "March of Wretched Sorrow", "Instant", "{X}{B}", neoWretchedSorrowOracle, game.CastSpellParams{
		XValue: 3, Targets: []game.TargetRef{neoCardTarget(victim)}, ExileFromHandIDs: []uuid.UUID{black}, Strict: true,
	})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(black) {
		t.Error("the black card was not exiled as the cost")
	}
	if got := g.StackMeta[spell].XValue; got != 3 {
		t.Errorf("X = %d, want 3", got)
	}
	passPriorityAroundTable(t, g)
	c := findCardAnywhere(t, g, victim)
	if c.DamageMarked != 3 {
		t.Errorf("the creature has %d damage, want 3", c.DamageMarked)
	}
	if me.Life-life != 3 {
		t.Errorf("gained %d life, want 3", me.Life-life)
	}
}

func TestS82NeoB14WretchedSorrowHitsAPlaneswalkerAndRefusesALandOrPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 4}, Owner: opp.ID, Controller: opp.ID})
	land := apcB1Push(g, opp.ID, "Plains", "Basic Land — Plains", "", "", 0, 0)
	advanceToMain(t, g)
	neoMana(me, "B", "B", "B", "B", "B", "B")
	for _, bad := range []game.TargetRef{neoCardTarget(land), playerTarget(opp.ID)} {
		if _, err := neoCast(t, g, "March of Wretched Sorrow", "Instant", "{X}{B}", neoWretchedSorrowOracle, game.CastSpellParams{
			XValue: 2, Targets: []game.TargetRef{bad}, Strict: true,
		}); err == nil {
			t.Fatalf("an illegal target %v was accepted", bad)
		}
	}
	if _, err := neoCast(t, g, "March of Wretched Sorrow", "Instant", "{X}{B}", neoWretchedSorrowOracle, game.CastSpellParams{
		XValue: 2, Targets: []game.TargetRef{neoCardTarget(walker)}, Strict: true,
	}); err != nil {
		t.Fatalf("a planeswalker was refused: %v", err)
	}
}

// --- March of Otherworldly Light ---

func TestS82NeoB14OtherworldlyLightExilesWithinXAndRefusesAboveIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := apcB1Push(g, opp.ID, "Small", "Creature — Ogre", "", "{2}", 2, 2)
	rock := apcB1Push(g, opp.ID, "Rock", "Artifact", "", "{3}", 0, 0)
	big := apcB1Push(g, opp.ID, "Big", "Creature — Ogre", "", "{4}", 4, 4)
	land := apcB1Push(g, opp.ID, "Plains", "Basic Land — Plains", "", "", 0, 0)
	advanceToMain(t, g)
	neoMana(me, "W", "W", "W", "W", "W", "W", "W", "W")
	cast := func(target uuid.UUID, x int) error {
		_, err := neoCast(t, g, "March of Otherworldly Light", "Instant", "{X}{W}", neoOtherworldlyOracle, game.CastSpellParams{
			XValue: x, Targets: []game.TargetRef{neoCardTarget(target)}, Strict: true,
		})
		return err
	}
	if err := cast(big, 3); err == nil {
		t.Error("a mana value 4 creature was accepted at X = 3")
	}
	if err := cast(land, 5); err == nil {
		t.Error("a land was accepted")
	}
	if err := cast(small, 2); err != nil {
		t.Fatalf("mana value 2 at X = 2: %v", err)
	}
	if err := cast(rock, 3); err != nil {
		t.Fatalf("a mana value 3 artifact at X = 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{small, rock} {
		if !g.Exile.Contains(id) || g.Battlefield.Contains(id) {
			t.Errorf("%v was not exiled", id)
		}
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the mana value 4 creature should still be on the battlefield")
	}
}

// --- March of Burgeoning Life ---

func TestS82NeoB14BurgeoningLifeFetchesTheSameNameTappedAndWantsLessThanX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elf := apcB1Push(g, opp.ID, "Llanowar Elves", "Creature — Elf Druid", "", "{G}", 1, 1)
	ogre := apcB1Push(g, opp.ID, "Hill Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3)
	twin := pushLibraryCardForTest(me, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Power: 1, Toughness: 1})
	other := pushLibraryCardForTest(me, game.Card{Name: "Fyndhorn Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Power: 1, Toughness: 1})
	advanceToMain(t, g)
	neoMana(me, "G", "G", "G", "G", "G", "G")
	cast := func(target uuid.UUID, x int) error {
		_, err := neoCast(t, g, "March of Burgeoning Life", "Instant", "{X}{G}", neoBurgeoningOracle, game.CastSpellParams{
			XValue: x, Targets: []game.TargetRef{neoCardTarget(target)}, Strict: true,
		})
		return err
	}
	if err := cast(ogre, 3); err == nil {
		t.Error("a mana value 3 creature was accepted at X = 3: it must be LESS than X")
	}
	if err := cast(elf, 1); err == nil {
		t.Error("a mana value 1 creature was accepted at X = 1")
	}
	if err := cast(elf, 2); err != nil {
		t.Fatalf("mana value 1 at X = 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the search asked for nothing")
	}
	answerSearchByID(t, g, me.ID, twin)
	if !g.Battlefield.Contains(twin) {
		t.Fatal("the creature with the same name did not enter")
	}
	if !findCardAnywhere(t, g, twin).Tapped {
		t.Error("it did not enter tapped")
	}
	if !me.Library.Contains(other) {
		t.Error("a creature with another name left the library")
	}
}

// --- March of Reckless Joy ---

func TestS82NeoB14RecklessJoyExilesXAndLetsYouPlayTwoOfThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var top []uuid.UUID
	for _, n := range []string{"Card A", "Card B", "Card C"} {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		top = append(top, id)
	}
	advanceToMain(t, g)
	neoMana(me, "R", "R", "R", "R")
	if _, err := neoCast(t, g, "March of Reckless Joy", "Instant", "{X}{R}", neoRecklessJoyOracle, game.CastSpellParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range top {
		if !g.Exile.Contains(id) {
			t.Fatalf("%v was not exiled", id)
		}
	}
	played := 0
	for _, id := range top {
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
			played++
			passPriorityAroundTable(t, g)
		}
	}
	if played != 2 {
		t.Errorf("played %d of the exiled cards, want exactly 2", played)
	}
}
