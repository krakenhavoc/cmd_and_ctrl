package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// until_end_of_turn_test.go is the card-level half of S32: Giant
// Growth and Overrun driven through a real cast, with the layer
// engine reading the result. The registry's own mechanics (undo,
// layer ordering, the end-step expiry rule) are pinned in
// server/internal/game/turn_scoped_statics_test.go.

const (
	giantGrowthOracle = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
	overrunOracle     = "204f9afe-c20b-4933-b5cd-aa572784762a"
)

// advanceToNextSeatsTurn walks the cursor past this seat's cleanup
// step. Turn.Round counts rounds, so the active seat is the
// reliable "the turn ended" marker.
func advanceToNextSeatsTurn(t *testing.T, g *game.Game) {
	t.Helper()
	start := g.Turn.ActiveSeat
	for i := 0; i < 30; i++ {
		if g.Turn.ActiveSeat != start {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("cursor never left seat %d's turn", start)
}

// TestGiantGrowthPumpsUntilCleanup is the S32 acceptance card: the
// +3/+3 is on the wire the moment the spell resolves and gone after
// the cleanup step.
func TestGiantGrowthPumpsUntilCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      me.ID,
		Controller: me.ID,
	})

	castCatalogSpell(t, g, "Giant Growth", "Instant", giantGrowthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 5 {
		t.Fatalf("post-resolution P/T = %d/%d, want 5/5", p, tough)
	}
	if n := len(g.ScopedEffects); n != 1 {
		t.Fatalf("ScopedEffects = %d, want 1", n)
	}

	advanceToNextSeatsTurn(t, g)

	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("ScopedEffects = %d after cleanup, want 0", n)
	}
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("post-cleanup P/T = %d/%d, want 2/2", p, tough)
	}
}

// TestGiantGrowthComposesWithAnthem — CR 613.7 at the card level:
// the pump is a layer 7c modification, so it adds to Glorious
// Anthem's +1/+1 instead of replacing it.
func TestGiantGrowthComposesWithAnthem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   gloriousAnthemOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      me.ID,
		Controller: me.ID,
	})
	if p := effectivePower(t, g, bear); p != 3 {
		t.Fatalf("setup: anthem power = %d, want 3", p)
	}

	castCatalogSpell(t, g, "Giant Growth", "Instant", giantGrowthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("anthem + Giant Growth P/T = %d/%d, want 6/6", p, tough)
	}
}

// TestOverrunPumpsAndGrantsTrample is the mass-grant acceptance:
// both halves land on every creature you control, and neither lands
// on an opponent's.
func TestOverrunPumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Enemy Bears",
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: opp.ID, Controller: opp.ID,
	})
	myLand := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Overrun", "Sorcery", overrunOracle, nil)
	passPriorityAroundTable(t, g)

	if p, tough := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("my creature = %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("my creature did not gain trample")
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("opponent's creature = %d power, want 2", p)
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("opponent's creature gained trample")
	}
	if p := effectivePower(t, g, myLand); p != 0 {
		t.Errorf("my land = %d power, want 0 (Overrun hits creatures only)", p)
	}

	advanceToNextSeatsTurn(t, g)
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("post-cleanup power = %d, want 2", p)
	}
	if hasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("trample survived the cleanup step")
	}
}

// TestOverrunMissesCreaturesThatArriveAfterIt is CR 611.2c: a
// one-shot continuous effect affects only the permanents that were
// on the battlefield when it resolved. The pump must NOT follow a
// creature cast later in the same turn.
func TestOverrunMissesCreaturesThatArriveAfterIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]

	early := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Early Bears",
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Overrun", "Sorcery", overrunOracle, nil)
	passPriorityAroundTable(t, g)

	late := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Late Bears",
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	if p := effectivePower(t, g, early); p != 5 {
		t.Errorf("creature present at resolution = %d power, want 5", p)
	}
	if p := effectivePower(t, g, late); p != 2 {
		t.Errorf("creature that arrived after resolution = %d power, want 2 (CR 611.2c)", p)
	}
	if hasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a creature that arrived after Overrun resolved gained trample (CR 611.2c)")
	}
}

// TestBoostUntilEOTDropsAFlickeredCreature is CR 400.7: a permanent
// that leaves and returns is a new object, so the grant stops
// applying even though the instance ID is reused by the engine.
func TestBoostUntilEOTDropsAFlickeredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Giant Growth", "Instant", giantGrowthOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 5 {
		t.Fatalf("setup: power = %d, want 5", p)
	}

	// Blink it out and back — same instance ID, new entry stamp.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneExile},
		bear,
	); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneExile},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		bear,
	); err != nil {
		t.Fatalf("return: %v", err)
	}

	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("power after blink = %d, want 2 (CR 400.7 — a new object)", p)
	}
}

// TestAangGainsLifelinkOnLessonCast restores the clause Aang, the
// Last Airbender shipped without in S23: "Whenever you cast a
// Lesson spell, Aang gains lifelink until end of turn." Lifelink is
// one of the twelve keywords the combat code honours, so this grant
// is real, not cosmetic.
func TestAangGainsLifelinkOnLessonCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Aang, the Last Airbender",
		TypeLine:   "Legendary Creature — Human Avatar Ally",
		OracleID:   aangTheLastAirbenderOracle,
		Power:      3, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})
	if hasEffectiveKeyword(t, g, aang, "lifelink") {
		t.Fatal("setup: Aang already has lifelink")
	}

	// A Lesson spell. Non-catalog on purpose — the trigger reads the
	// cast card's type line off the stack, so any Lesson works.
	castCatalogSpell(t, g, "Introduction to Prophecy", "Sorcery — Lesson", "", nil)
	passPriorityAroundTable(t, g)

	if !hasEffectiveKeyword(t, g, aang, "lifelink") {
		t.Error("Aang did not gain lifelink from a Lesson cast")
	}
	if !hasEffectiveKeyword(t, g, aang, "flying") {
		t.Error("the grant clobbered Aang's printed flying")
	}

	advanceToNextSeatsTurn(t, g)
	if hasEffectiveKeyword(t, g, aang, "lifelink") {
		t.Error("lifelink survived the cleanup step")
	}
}

// TestAangIgnoresNonLessonCasts — the trigger is type-gated, not a
// blanket "whenever you cast anything".
func TestAangIgnoresNonLessonCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aang := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Aang, the Last Airbender",
		TypeLine:   "Legendary Creature — Human Avatar Ally",
		OracleID:   aangTheLastAirbenderOracle,
		Power:      3, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Divination Filler", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)

	if hasEffectiveKeyword(t, g, aang, "lifelink") {
		t.Error("a non-Lesson sorcery granted lifelink")
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("ScopedEffects = %d, want 0", n)
	}
}

// hasEffectiveKeyword forces a recompute and reports whether the
// keyword is in the card's post-layer abilities — the same read
// game.HasKeyword does from the combat code.
func hasEffectiveKeyword(t *testing.T, g *game.Game, cardID uuid.UUID, kw string) bool {
	t.Helper()
	found := false
	present := false
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.InstanceID != cardID {
				continue
			}
			found = true
			present = game.HasKeyword(c, kw)
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return present
}
