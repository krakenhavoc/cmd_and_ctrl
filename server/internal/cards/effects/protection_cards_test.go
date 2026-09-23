package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// protection_cards_test.go — #662, the catalog half. One assertion
// per card that lost its protection caveat, and each one is about the
// thing that caveat said was missing rather than about the Spec
// struct: a test that only read PrintedKeywords back would have
// passed before the engine change too.

const (
	baneslayerOracle    = "0e11792b-7fe5-4208-aa0b-e5d09b2b65fe"
	swordFireIceOracle  = "2ccdc60a-49a9-44b9-a7af-0ebf18b26785"
	motherOfRunesOracle = "60433b48-d27f-413c-905c-43839b1943f1"
	swordBodyMindOracle = "fac42229-4f5f-4d04-85dd-5031d4e435aa"
)

// protectionsOn reads the engine's one quality reader off a
// battlefield permanent, as printed strings.
func protectionsOn(t *testing.T, g *game.Game, id uuid.UUID) []string {
	t.Helper()
	var out []string
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != id {
				continue
			}
			for _, q := range game.ProtectionQualities(c) {
				out = append(out, q.Printed)
			}
		}
	})
	return out
}

func hasString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// TestBaneslayerAngelIsProtectedFromDemonsAndDragons, and the
// changeling case with it: her quality is a SUBTYPE, so CR 702.73a
// has to reach it (#939's AllCreatureTypes). A test that only used a
// printed Demon would pass for a matcher that never asks the flag.
func TestBaneslayerAngelIsProtectedFromDemonsAndDragons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	angel := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Baneslayer Angel",
		TypeLine: "Creature — Angel", OracleID: baneslayerOracle,
		Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	// Every blocker here FLIES. Baneslayer Angel does too, and the
	// flying check runs before the protection one — a grounded blocker
	// would be refused for the wrong reason and the test would pass
	// without protection existing at all.
	demon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Shivan Demon",
		TypeLine: "Creature — Demon", Power: 4, Toughness: 4,
		Keywords: []string{"flying"},
		Owner:    opp.ID, Controller: opp.ID,
	})
	changeling := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mistform Ultimus",
		TypeLine: "Creature — Shapeshifter", Power: 2, Toughness: 2,
		Keywords: []string{game.KeywordChangeling, "flying"},
		Owner:    opp.ID, Controller: opp.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wind Drake",
		TypeLine: "Creature — Drake", Power: 2, Toughness: 2,
		Keywords: []string{"flying"},
		Owner:    opp.ID, Controller: opp.ID,
	})

	got := protectionsOn(t, g, angel)
	if !hasString(got, "Demons") || !hasString(got, "Dragons") {
		t.Fatalf("Baneslayer Angel's protections = %q", got)
	}

	// CR 702.16f, through the one block check.
	card := func(id uuid.UUID) *game.Card {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				return &g.Battlefield.Cards[i]
			}
		}
		t.Fatalf("card %v is not on the battlefield", id)
		return nil
	}
	g.ReadSnapshot(func() {
		if r := g.BlockPairRefusalLocked(card(angel), card(demon)); r.Reason != game.BlockReasonProtection {
			t.Errorf("a flying Demon blocking Baneslayer Angel: reason = %q, want protection", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(card(angel), card(changeling)); r.Reason != game.BlockReasonProtection {
			t.Errorf("a changeling is every creature type (CR 702.73a), so it is a Demon: reason = %q", r.Reason)
		}
		if r := g.BlockPairRefusalLocked(card(angel), card(bear)); !r.Legal() {
			t.Errorf("a Drake is neither a Demon nor a Dragon and flies: reason = %q", r.Reason)
		}
	})
}

// The two Swords grant their protections to the EQUIPPED CREATURE
// (layer 6), and the Sword itself is a colourless artifact — which is
// the asymmetry CR 702.16b turns on and the one a controller-based
// implementation would get wrong.
func TestTheSwordsGrantTheirProtectionsToTheEquippedCreature(t *testing.T) {
	for _, tc := range []struct {
		name    string
		oracle  string
		printed [2]string
	}{
		{"Sword of Fire and Ice", swordFireIceOracle, [2]string{"red", "blue"}},
		{"Sword of Feast and Famine", swordFeastOracle, [2]string{"black", "green"}},
		{"Sword of Body and Mind", swordBodyMindOracle, [2]string{"green", "blue"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]

			bear := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Grizzly Bears",
				TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
				Owner: me.ID, Controller: me.ID,
			})
			sword := pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: tc.name,
				TypeLine: "Artifact — Equipment", OracleID: tc.oracle,
				Owner: me.ID, Controller: me.ID,
			})

			if got := protectionsOn(t, g, bear); len(got) != 0 {
				t.Fatalf("the bear has %q before the Equipment is attached", got)
			}
			g.WithWriteLock(func() {
				if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
					t.Fatalf("attach: %v", err)
				}
			})
			got := protectionsOn(t, g, bear)
			for _, want := range tc.printed {
				if !hasString(got, want) {
					t.Errorf("equipped creature's protections = %q, want %q among them", got, want)
				}
			}
			// The Sword itself stays a colourless artifact — its own
			// damage trigger is a source with no colour.
			if p := protectionsOn(t, g, sword); len(p) != 0 {
				t.Errorf("the Equipment granted itself %q", p)
			}
		})
	}
}

// Animar's protections are printed data, and the caveat they replaced
// said he "can be targeted and damaged by white and black sources".
// Both halves are asserted, because they are different checks.
func TestAnimarIsProtectedFromWhiteAndBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	animar := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Animar, Soul of Elements",
		TypeLine: "Legendary Creature — Elemental", OracleID: animarOracle,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	got := protectionsOn(t, g, animar)
	if !hasString(got, "white") || !hasString(got, "black") {
		t.Fatalf("Animar's protections = %q", got)
	}

	// Targeting, CR 702.16b — an opponent's white spell.
	white := game.NewCard("Swords to Plowshares", opp.ID)
	white.TypeLine = "Instant"
	white.Colors = []string{"W"}
	red := game.NewCard("Chaos Warp", opp.ID)
	red.TypeLine = "Instant"
	red.Colors = []string{"R"}
	g.WithWriteLock(func() {
		g.Stack.PushTop(white)
		g.Stack.PushTop(red)
	})
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != animar {
				continue
			}
			w := findStackCard(g, white.InstanceID)
			r := findStackCard(g, red.InstanceID)
			if game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceObject(opp.ID, w)) {
				t.Error("a white spell may not target Animar")
			}
			if !game.CanBeTargetedBy(c, game.ZoneBattlefield, game.SourceObject(opp.ID, r)) {
				t.Error("a RED spell may: Animar has no protection from red")
			}
		}
	})

	// Damage, CR 702.16e.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(white.InstanceID, animar, 2); err != nil {
			t.Fatalf("white damage: %v", err)
		}
		if err := g.DealDamageToCreatureForEffect(red.InstanceID, animar, 1); err != nil {
			t.Fatalf("red damage: %v", err)
		}
	})
	var marked int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == animar {
				marked = c.DamageMarked
			}
		}
	})
	if marked != 1 {
		t.Errorf("Animar has %d damage marked, want 1 (the white 2 prevented, the red 1 dealt)", marked)
	}
}

func findStackCard(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Stack.Cards {
		if g.Stack.Cards[i].InstanceID == id {
			return &g.Stack.Cards[i]
		}
	}
	return nil
}

// Mother of Runes is the choose-a-color half (#742's existing kind).
// The Spec assertion is about the shape — a tapped activated ability
// with a "creature you control" clause — and the behavioural one is
// that the colour the prompt returns becomes a real, enforced token.
func TestMotherOfRunesGrantsTheChosenColour(t *testing.T) {
	spec, ok := Lookup(motherOfRunesOracle)
	if !ok {
		t.Fatal("Mother of Runes is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v, caveats = %q", spec.Completeness, spec.Caveats)
	}
	if len(spec.Activated) != 1 || !spec.Activated[0].Cost.Tap || spec.Activated[0].Targets == nil {
		t.Fatalf("activated = %+v, want one {T} ability with a target clause", spec.Activated)
	}

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	// The grant the prompt's continuation makes, with the token the
	// engine's own reader mints for the chosen colour.
	g.WithWriteLock(func() {
		token := game.ProtectionFromColor("R")
		if token != "protection from red" {
			t.Fatalf("ProtectionFromColor(R) = %q", token)
		}
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: bear})
		if err := (GrantKeywordUntilEOT{
			Target: bear, Keywords: []string{token}, Label: "test",
		}).Apply(ctx); err != nil {
			t.Fatalf("grant: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if got := protectionsOn(t, g, bear); !hasString(got, "red") {
		t.Errorf("the granted protection is not readable: %q", got)
	}
}

// TestYawgmothIsProtectedFromHumans (issue #1117) is Baneslayer
// Angel's subtype case again, checked against the SOURCE of an
// ability rather than a spell: CR 702.16b tests the object doing the
// targeting, not its controller, so a Human creature's own ability
// cannot target Yawgmoth even though a non-Human creature's can.
func TestYawgmothIsProtectedFromHumans(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	yawgmoth := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Yawgmoth, Thran Physician",
		TypeLine: "Legendary Creature — Human Cleric", OracleID: yawgmothOracle,
		Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	got := protectionsOn(t, g, yawgmoth)
	if !hasString(got, "Humans") {
		t.Fatalf("Yawgmoth's protections = %q", got)
	}

	human := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Human Source",
		TypeLine: "Creature — Human Soldier", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	nonHuman := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin Source",
		TypeLine: "Creature — Goblin", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})

	card := func(id uuid.UUID) *game.Card {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				return &g.Battlefield.Cards[i]
			}
		}
		t.Fatalf("card %v is not on the battlefield", id)
		return nil
	}
	g.ReadSnapshot(func() {
		if game.CanBeTargetedBy(card(yawgmoth), game.ZoneBattlefield, game.SourceObject(opp.ID, card(human))) {
			t.Error("an ability from a Human source may not target Yawgmoth")
		}
		if !game.CanBeTargetedBy(card(yawgmoth), game.ZoneBattlefield, game.SourceObject(opp.ID, card(nonHuman))) {
			t.Error("an ability from a non-Human source may target Yawgmoth")
		}
	})
}
