package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_sz_b3_cards_test.go — The Lost Caverns of Ixalan, S–Z batch 3
// (#4470). Each test plays the real catalog entry and asserts what its
// text says.

const (
	lciSz3SongOracle        = "87300967-1a59-4dce-8876-f008ff3bfb0b"
	lciSz3SovereignOracle   = "087b2490-e415-44e2-a674-3b5619a0fad5"
	lciSz3TorchOracle       = "d2da26d6-b0f7-4af0-9ce8-fb8d82fe2d85"
	lciSz3BelligerentOracle = "9a0c9023-9431-4f98-b167-ab4c7215319c"
	lciSz3CrackshotOracle   = "b1ce2971-fffa-40b3-9bfd-29afdb929e74"
	lciSz3TreasureMapOracle = "0b55eac6-a745-4bf4-8926-5ce83bc38d7d"
	lciSz3ChompOracle       = "d417f574-38a4-405c-83bb-49c5cf542648"
)

func TestLciSz3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciSz3SongOracle:               "Song of Stupefaction",
		lciSz3SovereignOracle:          "Sovereign Okinec Ahau",
		lciSz3TorchOracle:              "Sunfire Torch",
		lciSz3BelligerentOracle:        "The Belligerent",
		lciSz3CrackshotOracle:          "Thousand Moons Crackshot",
		lciSz3TreasureMapOracle:        "Treasure Map",
		lciSz3TreasureMapOracle + "#1": "Treasure Cove",
		lciSz3ChompOracle:              "Triumphant Chomp",
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

// --- Song of Stupefaction ----------------------------------------------------

func TestLciSz3SongOfStupefactionMaysMillAndShrinksByTheDescendCount(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lciSz2GraveyardPermanents(me, 2)
		spmB17Top(me, game.Card{Name: "Milled Bolt", TypeLine: "Instant"})
		spmB17Top(me, game.Card{Name: "Milled Rock", TypeLine: "Artifact"})
		victim := pushVanillaCreature(g, opp.ID, "Victim", 5, 5)

		castCatalogSpell(t, g, "Song of Stupefaction", "Enchantment — Aura", lciSz3SongOracle, cardRefs(victim))
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, accept)
		passPriorityAroundTable(t, g)

		// Two permanent cards to start; the mill adds the Rock, not the Bolt.
		want := 3
		if !accept {
			want = 2
		}
		c := apaLive(g, victim)
		if got := c.CurrentPower(); got != 5-want {
			t.Errorf("accept=%v: power %d, want %d", accept, got, 5-want)
		}
		if got := c.CurrentToughness(); got != 5 {
			t.Errorf("accept=%v: toughness %d, want 5 (-X/-0)", accept, got)
		}
	}
}

func TestLciSz3SongOfStupefactionGrowsWithTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	castCatalogSpell(t, g, "Song of Stupefaction", "Enchantment — Aura", lciSz3SongOracle, cardRefs(victim))
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := apaLive(g, victim).CurrentPower(); got != 5 {
		t.Fatalf("power %d with an empty graveyard, want 5", got)
	}
	// The count is the Aura controller's graveyard, read live: an
	// opponent's permanent cards don't count, and yours do as they arrive.
	for _, p := range []*game.Player{opp, me, me} {
		spmB17Top(p, game.Card{Name: "Milled Rock", TypeLine: "Artifact"})
		spmB17Top(p, game.Card{Name: "Milled Rock", TypeLine: "Artifact"})
		g.WithWriteLock(func() {
			if _, err := g.MillToZoneForEffect(p.ID, 2, game.ZoneGraveyard); err != nil {
				t.Fatalf("mill: %v", err)
			}
		})
	}
	if got := apaLive(g, victim).CurrentPower(); got != 1 {
		t.Errorf("power %d after four permanent cards reach your graveyard, want 1", got)
	}
}

// --- Sovereign Okinec Ahau ---------------------------------------------------

func TestLciSz3SovereignOkinecAhauAddsTheDifferenceOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	sovereign := b12Push(g, me.ID, "Sovereign Okinec Ahau", "Legendary Creature — Cat Noble", lciSz3SovereignOracle, 3, 4)
	grown := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grown Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterPlusOne: 2},
	})
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterPlusOne: 1},
	})

	declareAttack(t, g, opp.ID, sovereign)
	passPriorityAroundTable(t, g)

	if got := tmtB1Card(t, g, grown).Counters[game.CounterPlusOne]; got != 4 {
		t.Errorf("Grown Bear has %d +1/+1 counters, want 4 (2 plus the difference of 2)", got)
	}
	if got := tmtB1Card(t, g, plain).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("Plain Bear has %d counters; its power is its base power", got)
	}
	if got := tmtB1Card(t, g, sovereign).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("the Sovereign has %d counters; its power is its base power", got)
	}
	if got := tmtB1Card(t, g, theirs).Counters[game.CounterPlusOne]; got != 1 {
		t.Errorf("an opponent's creature has %d counters, want 1 (only creatures you control)", got)
	}
	if spec, _ := Lookup(lciSz3SovereignOracle); len(spec.Triggered) < 1 || spec.Triggered[0].Key != "Sovereign Okinec Ahau — ward {2}" {
		t.Error("the Sovereign's first trigger is not its ward {2}")
	}
}

// --- Sunfire Torch -----------------------------------------------------------

func lciSz3TorchAttack(t *testing.T) (g *game.Game, me, opp *game.Player, bear, torch uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear = b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	torch = seedEquipment(g, me.ID, "Sunfire Torch", lciSz3TorchOracle)
	equipTo(t, g, me.ID, torch, bear)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Fatalf("equipped power %d, want 3", got)
	}
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	return g, me, opp, bear, torch
}

func TestLciSz3SunfireTorchSacrificesItselfToDealTwo(t *testing.T) {
	g, me, opp, bear, torch := lciSz3TorchAttack(t)
	life := opp.Life
	answerMayChoice(t, g, me.ID, true)
	if g.Battlefield.Contains(torch) {
		t.Fatal("the Torch was not sacrificed")
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d life, want 2", got)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature left")
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d once the Torch is gone, want 2", got)
	}
}

func TestLciSz3SunfireTorchDecliningKeepsIt(t *testing.T) {
	g, me, opp, _, torch := lciSz3TorchAttack(t)
	life := opp.Life
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(torch) {
		t.Error("the Torch left after declining")
	}
	if opp.Life != life {
		t.Errorf("life %d -> %d after declining", life, opp.Life)
	}
}

// --- The Belligerent ---------------------------------------------------------

func TestLciSz3TheBelligerentMakesATreasureAndOpensTheTopOfTheLibrary(t *testing.T) {
	s := newCrewScene(t)
	ship := s.pushVehicle("The Belligerent", lciSz3BelligerentOracle, 5, 5)
	small := s.pushCreature("Small", "Creature — Human", 2, 2)
	crewer := s.pushCreature("Crewer", "Creature — Human", 3, 3)
	toMain(t, s.g)
	if err := s.g.ActivateCatalogAbility(s.me.ID, ship, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Fatal("crewed with total power 2")
	}
	s.crewAndResolve(t, ship, 0, crewer)
	top := spmB17Top(s.me, game.Card{Name: "Top Instant", TypeLine: "Instant", ManaCost: "{1}{U}"})
	if perm := grantedPermissionOn(s.g, s.me.ID, top, game.ZoneLibrary); perm.Granted() {
		t.Fatal("the top is open before the attack")
	}
	treasures := countTreasuresControlledBy(s.g, s.me.ID)

	declareAttack(t, s.g, s.opp.ID, ship)
	passPriorityAroundTable(t, s.g)
	if got := countTreasuresControlledBy(s.g, s.me.ID); got != treasures+1 {
		t.Errorf("Treasures %d -> %d, want one more", treasures, got)
	}
	if perm := grantedPermissionOn(s.g, s.me.ID, top, game.ZoneLibrary); !perm.Granted() {
		t.Fatal("the attack opened nothing")
	}
	visible := false
	s.g.ReadSnapshot(func() { visible = s.g.LibraryTopVisibleToLocked(s.me.ID, s.me.ID) })
	if !visible {
		t.Error("the top card is not visible to its owner")
	}
	floatMana(t, s.g, s.me, "{C}{U}")
	if err := s.g.CastSpell(s.me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast from the top for its mana cost: %v", err)
	}
	passPriorityAroundTable(t, s.g)

	nextTurnOfMine(t, s.g)
	next := spmB17Top(s.me, game.Card{Name: "Later Instant", TypeLine: "Instant", ManaCost: "{U}"})
	if perm := grantedPermissionOn(s.g, s.me.ID, next, game.ZoneLibrary); perm.Granted() {
		t.Error("the permission outlived the turn")
	}
}

// --- Thousand Moons Crackshot ------------------------------------------------

func TestLciSz3ThousandMoonsCrackshotPaysToTapACreature(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me, opp := g.Seats[0], g.Seats[1]
		crackshot := b12Push(g, me.ID, "Thousand Moons Crackshot", "Creature — Human Soldier", lciSz3CrackshotOracle, 2, 2)
		victim := pushVanillaCreature(g, opp.ID, "Victim", 3, 3)

		declareAttack(t, g, opp.ID, crackshot)
		// The pool empties at every step change, so the mana is floated now.
		b06AddMana(me, "W", "C", "C")
		passPriorityAroundTable(t, g)
		answerPayUnless(t, g, me.ID, pay)
		if pay {
			answerPickTarget(t, g, victim)
		}
		passPriorityAroundTable(t, g)
		if got := tappedNow(g, victim); got != pay {
			t.Errorf("pay=%v: victim tapped %v", pay, got)
		}
		wantPool := 3
		if pay {
			wantPool = 0
		}
		if got := len(poolColors(me)); got != wantPool {
			t.Errorf("pay=%v: pool %v, want %d mana left", pay, poolColors(me), wantPool)
		}
	}
}

// --- Treasure Map // Treasure Cove ------------------------------------------

func lciSz3TreasureMapCard(owner uuid.UUID, landmarks int) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   lciSz3TreasureMapOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Treasure Map", TypeLine: "Artifact", ManaCost: "{2}"},
			{Name: "Treasure Cove", TypeLine: "Land"},
		},
	}
	if landmarks > 0 {
		c.Counters = map[string]int{"landmark": landmarks}
	}
	c.SetFace(0)
	return c
}

func lciSz3ActivateMap(t *testing.T, g *game.Game, me *game.Player, mapID uuid.UUID) {
	t.Helper()
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, mapID, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
}

func TestLciSz3TreasureMapScriesAndCountsLandmarks(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	mapID := pushBattlefieldCardWithTimestamp(g, lciSz3TreasureMapCard(me.ID, 0))
	lciSz3ActivateMap(t, g, me, mapID)

	c := tmtB1Card(t, g, mapID)
	if c.Name != "Treasure Map" {
		t.Errorf("transformed on the first counter: %q", c.Name)
	}
	if got := c.Counters["landmark"]; got != 1 {
		t.Errorf("landmark counters %d, want 1", got)
	}
	if got := countTreasuresControlledBy(g, me.ID); got != 0 {
		t.Errorf("%d Treasures before the third counter", got)
	}
}

func TestLciSz3TreasureMapTransformsOnTheThirdLandmark(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	mapID := pushBattlefieldCardWithTimestamp(g, lciSz3TreasureMapCard(me.ID, 2))
	lciSz3ActivateMap(t, g, me, mapID)

	c := tmtB1Card(t, g, mapID)
	if c.Name != "Treasure Cove" || !c.IsLand() {
		t.Fatalf("it is %q %q, want the Treasure Cove land", c.Name, c.TypeLine)
	}
	if got := c.Counters["landmark"]; got != 0 {
		t.Errorf("landmark counters %d, want them removed", got)
	}
	if got := countTreasuresControlledBy(g, me.ID); got != 3 {
		t.Errorf("Treasures %d, want 3", got)
	}

	// Treasure Cove: {T}, Sacrifice a Treasure: Draw a card.
	nextTurnOfMine(t, g)
	treasure := tmtB1Named(g, "Treasure")[0]
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, mapID, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}}); err != nil {
		t.Fatalf("activate the Cove: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
	if g.Battlefield.Contains(treasure) {
		t.Error("the Treasure was not sacrificed")
	}
	if got := countTreasuresControlledBy(g, me.ID); got != 2 {
		t.Errorf("Treasures %d, want 2", got)
	}
}

// --- Triumphant Chomp --------------------------------------------------------

func TestLciSz3TriumphantChompDealsTheGreaterOfTwoAndTheBiggestDinosaur(t *testing.T) {
	for _, tc := range []struct {
		name string
		dino int
		want int
	}{
		{"no Dinosaur", 0, 2},
		{"a 1-power Dinosaur", 1, 2},
		{"a 5-power Dinosaur", 5, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			if tc.dino > 0 {
				b12Creature(g, me.ID, "Dino", "Creature — Dinosaur", tc.dino, tc.dino)
			}
			// A bigger non-Dinosaur and an opponent's bigger Dinosaur don't count.
			b12Creature(g, me.ID, "Big Bear", "Creature — Bear", 9, 9)
			b12Creature(g, opp.ID, "Their Dino", "Creature — Dinosaur", 8, 8)
			victim := pushVanillaCreature(g, opp.ID, "Victim", 10, 10)

			castCatalogSpell(t, g, "Triumphant Chomp", "Sorcery", lciSz3ChompOracle, cardRefs(victim))
			passPriorityAroundTable(t, g)
			if got := damageMarkedOn(g, victim); got != tc.want {
				t.Errorf("damage %d, want %d", got, tc.want)
			}
		})
	}
}
