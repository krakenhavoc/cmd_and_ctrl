package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestHeartOfYavimayaSacrificesAForestAndEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.PushSacrificeCandidate(g, me.ID, "Forest", "Basic Land — Forest", true)
	effectstest.PushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", false)

	heart := effectstest.PlayLandFromHand(t, g, "Heart of Yavimaya", effectstest.SacrificeInsteadLandOracleIDs["Heart of Yavimaya"])

	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("Heart of Yavimaya entered without asking for a Forest")
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 1..1 — the sacrifice is not a \"may\"", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != forest {
		t.Errorf("candidates %v, want just the Forest (tapped is fine)", c.ChooseCards)
	}
	if _, ok := effectstest.BattlefieldCard(g, heart); ok {
		t.Fatal("the land entered before the sacrifice was chosen")
	}
	// Not a "may": the empty answer is refused.
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, nil); err == nil {
		t.Error("declining a mandatory sacrifice was accepted")
	}

	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, forest)
	if _, ok := effectstest.BattlefieldCard(g, heart); !ok {
		t.Fatal("Heart of Yavimaya did not enter after the sacrifice")
	}
	if !me.Graveyard.Contains(forest) {
		t.Error("the sacrificed Forest is not in the graveyard")
	}
	if n := effectstest.SacrificeEventsFor(g, forest); n != 1 {
		t.Errorf("%d EventSacrifice for the Forest, want 1", n)
	}
}

func TestHeartOfYavimayaWithNoForestGoesToTheGraveyardAndSpendsTheDrop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", false)

	heart := effectstest.PlayLandFromHand(t, g, "Heart of Yavimaya", effectstest.SacrificeInsteadLandOracleIDs["Heart of Yavimaya"])

	if c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID); c != nil {
		t.Error("asked for a Forest with none on the battlefield")
	}
	if _, ok := effectstest.BattlefieldCard(g, heart); ok {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.PushSacrificeCandidate(g, me.ID, "Plains", "Basic Land — Plains", false)
	b := effectstest.PushSacrificeCandidate(g, me.ID, "Island", "Basic Land — Island", false)
	tapped := effectstest.PushSacrificeCandidate(g, me.ID, "Swamp", "Basic Land — Swamp", true)

	vale := effectstest.PlayLandFromHand(t, g, "Lotus Vale", effectstest.SacrificeInsteadLandOracleIDs["Lotus Vale"])

	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
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
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, a, b)
	if _, ok := effectstest.BattlefieldCard(g, vale); !ok {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	only := effectstest.PushSacrificeCandidate(g, me.ID, "Plains", "Basic Land — Plains", false)

	ruins := effectstest.PlayLandFromHand(t, g, "Scorched Ruins", effectstest.SacrificeInsteadLandOracleIDs["Scorched Ruins"])

	if c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID); c != nil {
		t.Error("asked with only one untapped land")
	}
	if !me.Graveyard.Contains(ruins) {
		t.Fatal("the land is not in the graveyard")
	}
	if _, ok := effectstest.BattlefieldCard(g, only); !ok {
		t.Error("the one untapped land was sacrificed anyway")
	}
}

func TestBalduvianTradingPostNeedsAnUntappedMountain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", true)
	untapped := effectstest.PushSacrificeCandidate(g, me.ID, "Mountain", "Basic Land — Mountain", false)

	effectstest.PlayLandFromHand(t, g, "Balduvian Trading Post", effectstest.SacrificeInsteadLandOracleIDs["Balduvian Trading Post"])

	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("no prompt")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != untapped {
		t.Errorf("candidates %v, want only the untapped Mountain", c.ChooseCards)
	}
}

func TestKjeldoranOutpostMakesASoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := effectstest.PushSacrificeCandidate(g, me.ID, "Plains", "Basic Land — Plains", false)
	outpost := effectstest.PlayLandFromHand(t, g, "Kjeldoran Outpost", effectstest.SacrificeInsteadLandOracleIDs["Kjeldoran Outpost"])
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, plains)

	before := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Soldier" {
			before++
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, outpost, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	for name, oracleID := range effectstest.SacrificeInsteadLandOracleIDs {
		spec, ok := effects.Lookup(oracleID)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s: registered as %q", name, spec.Name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %q with %d caveats", name, spec.Completeness, len(spec.Caveats))
		}
		if len(spec.Replacements) != 1 || spec.Replacements[0].EntryCardChoice == nil ||
			spec.Replacements[0].EntryCardChoice.Action != game.EntryCardSacrifice {
			t.Errorf("%s: no sacrifice entry choice", name)
		}
	}
}
