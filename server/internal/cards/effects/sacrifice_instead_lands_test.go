package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_instead_lands_test.go — ADR 0098 Decision 11: "If this land
// would enter, sacrifice <n> <kind> instead. If you do, put this land
// onto the battlefield. If you don't, put it into its owner's
// graveyard."

var sacrificeInsteadLandOracleIDs = map[string]string{
	"Heart of Yavimaya":      "6c9a854c-0509-4ed4-9d94-c45b823b65e5",
	"Kjeldoran Outpost":      "8b370db5-dfb9-4ea0-9017-bae3e767b041",
	"Lake of the Dead":       "bdf476e5-1d57-4b17-b45b-d52fd75aadeb",
	"Balduvian Trading Post": "7647940e-c99c-401c-ad1d-9ec730f66b6f",
	"Soldevi Excavations":    "5baa7abe-5bdf-40ce-9a83-a93b7cae71a3",
	"Lotus Vale":             "01fc5bb3-ebd7-4ab4-8aef-2ece1e1d9b7c",
	"Scorched Ruins":         "6ee68855-c8c5-422b-88da-163c09a96416",
}

// pushSacrificeCandidate puts a basic-typed land onto the active seat's battlefield.
func pushSacrificeCandidate(g *game.Game, owner uuid.UUID, name, typeLine string, tapped bool) uuid.UUID {
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
	if tapped {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		})
	}
	return id
}

func sacrificeEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventSacrifice && ev.CardID == id {
			n++
		}
	}
	return n
}

func TestHeartOfYavimayaSacrificesAForestAndEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := pushSacrificeCandidate(g, me.ID, "Forest", "Basic Land — Forest", true)
	pushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", false)

	heart := playLandFromHand(t, g, "Heart of Yavimaya", sacrificeInsteadLandOracleIDs["Heart of Yavimaya"])

	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("Heart of Yavimaya entered without asking for a Forest")
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 1..1 — the sacrifice is not a \"may\"", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != forest {
		t.Errorf("candidates %v, want just the Forest (tapped is fine)", c.ChooseCards)
	}
	if _, ok := battlefieldCard(g, heart); ok {
		t.Fatal("the land entered before the sacrifice was chosen")
	}
	// Not a "may": the empty answer is refused.
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, nil); err == nil {
		t.Error("declining a mandatory sacrifice was accepted")
	}

	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, forest)
	if _, ok := battlefieldCard(g, heart); !ok {
		t.Fatal("Heart of Yavimaya did not enter after the sacrifice")
	}
	if !me.Graveyard.Contains(forest) {
		t.Error("the sacrificed Forest is not in the graveyard")
	}
	if n := sacrificeEventsFor(g, forest); n != 1 {
		t.Errorf("%d EventSacrifice for the Forest, want 1", n)
	}
}

func TestHeartOfYavimayaWithNoForestGoesToTheGraveyardAndSpendsTheDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", false)

	heart := playLandFromHand(t, g, "Heart of Yavimaya", sacrificeInsteadLandOracleIDs["Heart of Yavimaya"])

	if c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID); c != nil {
		t.Error("asked for a Forest with none on the battlefield")
	}
	if _, ok := battlefieldCard(g, heart); ok {
		t.Fatal("Heart of Yavimaya entered for free")
	}
	if !me.Graveyard.Contains(heart) {
		t.Fatal("Heart of Yavimaya is not in its owner's graveyard")
	}
	// Owner decision 7: the play happened.
	if g.LandsPlayedThisTurn[me.ID] != 1 {
		t.Errorf("lands played this turn = %d, want 1", g.LandsPlayedThisTurn[me.ID])
	}
}

func TestLotusValeNeedsTwoUntappedLands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushSacrificeCandidate(g, me.ID, "Plains", "Basic Land — Plains", false)
	b := pushSacrificeCandidate(g, me.ID, "Island", "Basic Land — Island", false)
	tapped := pushSacrificeCandidate(g, me.ID, "Swamp", "Basic Land — Swamp", true)

	vale := playLandFromHand(t, g, "Lotus Vale", sacrificeInsteadLandOracleIDs["Lotus Vale"])

	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("Lotus Vale did not ask")
	}
	if c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 2..2", c.ChooseMin, c.ChooseMax)
	}
	for _, id := range c.ChooseCards {
		if id == tapped {
			t.Error("a tapped land is a candidate for \"two untapped lands\"")
		}
	}
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, []uuid.UUID{a}); err == nil {
		t.Error("sacrificing one land for \"two untapped lands\" was accepted")
	}
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	if _, ok := battlefieldCard(g, vale); !ok {
		t.Fatal("Lotus Vale did not enter")
	}
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the two lands were not sacrificed")
	}
}

// TestScorchedRuinsWithOneUntappedLandSacrificesNothing is the ruling: "If
// you don't sacrifice the lands, Lotus Vale never enters — it goes
// directly to your graveyard." With one untapped land it cannot, and
// the one it has is not spent.
func TestScorchedRuinsWithOneUntappedLandSacrificesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	only := pushSacrificeCandidate(g, me.ID, "Plains", "Basic Land — Plains", false)

	ruins := playLandFromHand(t, g, "Scorched Ruins", sacrificeInsteadLandOracleIDs["Scorched Ruins"])

	if c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID); c != nil {
		t.Error("asked with only one untapped land")
	}
	if !me.Graveyard.Contains(ruins) {
		t.Fatal("the land is not in the graveyard")
	}
	if _, ok := battlefieldCard(g, only); !ok {
		t.Error("the one untapped land was sacrificed anyway")
	}
}

func TestBalduvianTradingPostNeedsAnUntappedMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", true)
	untapped := pushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", false)

	playLandFromHand(t, g, "Balduvian Trading Post", sacrificeInsteadLandOracleIDs["Balduvian Trading Post"])

	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != untapped {
		t.Errorf("candidates %v, want only the untapped Mountain", c.ChooseCards)
	}
}

func TestKjeldoranOutpostMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := pushSacrificeCandidate(g, me.ID, "Plains", "Basic Land — Plains", false)
	outpost := playLandFromHand(t, g, "Kjeldoran Outpost", sacrificeInsteadLandOracleIDs["Kjeldoran Outpost"])
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, plains)

	before := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			before++
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, outpost, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	after := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			after++
		}
	}
	if after != before+1 {
		t.Errorf("Soldier tokens %d → %d, want one more", before, after)
	}
}

func TestEverySacrificeInsteadLandIsRegisteredComplete(t *testing.T) {
	for name, oracleID := range sacrificeInsteadLandOracleIDs {
		spec, ok := Lookup(oracleID)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s: registered as %q", name, spec.Name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %q with %d caveats", name, spec.Completeness, len(spec.Caveats))
		}
		if len(spec.Replacements) != 1 || spec.Replacements[0].EntryCardChoice == nil ||
			spec.Replacements[0].EntryCardChoice.Action != game.EntryCardSacrifice {
			t.Errorf("%s: no sacrifice entry choice", name)
		}
	}
}
