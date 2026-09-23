package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// earthbend_cards_test.go — #1178, the catalog half of the AVATAR
// keyword action. The engine half (the layers, the duration, the
// delayed return, the CR 614 window on the count) is in
// game/earthbend_test.go; everything here is a real card, cast and
// resolved through the stack.
//
// Three proofs, chosen for what each one adds:
//
//	Earthbending Lesson   the keyword and nothing else — a fixed count
//	Rockalanche           a count read at RESOLUTION, plus flashback
//	The Legend of Kyoshi  the keyword beside a second clause about the
//	                      same land, and the caveat #1179 shipped

const (
	earthbendingLessonOracle = "5e113f0f-4469-4276-82e1-0a804f3de444"
	rockalancheOracle        = "d1ecfad3-e79a-447b-a932-68ef728eaf1f"
)

// pushEarthbendLand parks an untapped land of the given subtype on the
// battlefield under `owner`, old enough that summoning sickness is
// not why anything below can or cannot attack.
func pushEarthbendLand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		Owner:      owner,
		Controller: owner,
	})
}

// earthbentBody reads the land's post-layer power/toughness and
// whether it is a Land Creature with haste.
func earthbentBody(t *testing.T, g *game.Game, land uuid.UUID) (power, toughness int, creatureLand, hasty bool) {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != land {
			continue
		}
		return c.CurrentPower(), c.CurrentToughness(),
			c.IsLand() && c.IsCreature(), game.HasKeyword(c, "haste")
	}
	t.Fatalf("the land is not on the battlefield")
	return 0, 0, false, false
}

// --- Earthbending Lesson --------------------------------------------

// TestEarthbendingLessonAnimatesTheTargetLand is the whole card: one
// keyword action, four counters, and a land that is now a 4/4 hasty
// creature that still taps for mana.
func TestEarthbendingLessonAnimatesTheTargetLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Earthbending Lesson", "Sorcery — Lesson", earthbendingLessonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	power, toughness, creatureLand, hasty := earthbentBody(t, g, land)
	if !creatureLand {
		t.Error("the land is not a Land Creature — earthbend adds Creature and keeps Land")
	}
	if power != 4 || toughness != 4 {
		t.Errorf("power/toughness = %d/%d, want 4/4 (base 0/0 plus four +1/+1 counters)", power, toughness)
	}
	if !hasty {
		t.Error("the animated land has no haste")
	}
	if got := countersOn(g, land, game.CounterPlusOne); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4", got)
	}
}

// TestEarthbendingLessonsLandComesBackTappedWhenItDies is the fourth
// sentence, on a real card: kill the animated land and it returns
// tapped, as a plain land again.
func TestEarthbendingLessonsLandComesBackTappedWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Earthbending Lesson", "Sorcery — Lesson", earthbendingLessonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(land); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	passPriorityAroundTable(t, g)

	if me.Graveyard.Contains(land) {
		t.Fatal("the dead land was never returned")
	}
	back := findBattlefieldCardByID(g, findBattlefieldByName(g, "Forest"))
	if back == nil {
		t.Fatal("no Forest on the battlefield after the return")
	}
	if !back.Tapped {
		t.Error("it returned untapped; earthbend says tapped")
	}
	if back.IsCreature() {
		t.Error("the returned land is still animated — CR 400.7 makes it a new object")
	}
}

// --- Rockalanche ----------------------------------------------------

// TestRockalancheCountsForestsAtResolution: X is not announced, it is
// read when the spell resolves, and every Forest counts — basic or
// not.
func TestRockalancheCountsForestsAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")
	pushEarthbendLand(g, me.ID, "Stomping Ground", "Land — Mountain Forest")
	target := pushEarthbendLand(g, me.ID, "Wastes", "Basic Land — Wastes")
	// An opponent's Forest is not "you control".
	pushEarthbendLand(g, g.Seats[1].ID, "Their Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Rockalanche", "Sorcery — Lesson", rockalancheOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)

	if got := countersOn(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 — a basic Forest and a Stomping Ground, "+
			"not the Wastes it targeted and not the opponent's", got)
	}
	power, _, creatureLand, hasty := earthbentBody(t, g, target)
	if !creatureLand || !hasty || power != 2 {
		t.Errorf("the target is not a 2/2 hasty Land Creature: power=%d land+creature=%v haste=%v",
			power, creatureLand, hasty)
	}
}

// TestRockalancheWithNoForestsStillAnimatesAndReturnsTheLand is
// "earthbend 0" on a real card: the land becomes a 0/0, the toughness
// state-based action kills it, and the delayed return hands it back
// tapped. Not a fizzle — that distinction is what
// KeywordAction.actsAtZeroCount exists for.
func TestRockalancheWithNoForestsStillAnimatesAndReturnsTheLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushEarthbendLand(g, me.ID, "Wastes", "Basic Land — Wastes")

	castCatalogSpell(t, g, "Rockalanche", "Sorcery — Lesson", rockalancheOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)

	back := findBattlefieldCardByID(g, findBattlefieldByName(g, "Wastes"))
	if back == nil {
		t.Fatal("the 0/0 land died and never came back")
	}
	if !back.Tapped {
		t.Error("it returned untapped")
	}
	if back.IsCreature() {
		t.Error("the returned land is still a creature")
	}
}

// TestRockalancheIsCastableFromTheGraveyardForFlashback pins the half
// of the card that is not earthbend: the zone and the price are
// declared separately, so the graveyard cast is offered.
func TestRockalancheIsCastableFromTheGraveyardForFlashback(t *testing.T) {
	spec, ok := Lookup(rockalancheOracle)
	if !ok {
		t.Fatal("Rockalanche is not registered")
	}
	if len(spec.CastableZones) != 1 || spec.CastableZones[0] != game.ZoneGraveyard {
		t.Errorf("CastableZones = %v, want the graveyard", spec.CastableZones)
	}
	if len(spec.AlternativeCosts) != 1 || spec.AlternativeCosts[0].Key != "flashback" {
		t.Errorf("AlternativeCosts = %v, want one flashback", spec.AlternativeCosts)
	}
	if got := spec.AlternativeCosts[0].ManaCost; got != "{5}{G}" {
		t.Errorf("flashback cost = %q, want {5}{G}", got)
	}
}

// --- The Legend of Kyoshi, chapter II --------------------------------

// TestKyoshiChapterTwoEarthbendsAndMakesItAnIsland is the caveat
// #1179 shipped, now cleared: chapter II earthbends X (the cards in
// hand) AND makes the same land an Island. Both clauses, one target.
func TestKyoshiChapterTwoEarthbendsAndMakesItAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")

	saga := importAndCast(t, g, kyoshiRow(), me)
	passPriorityAroundTable(t, g)
	if got := loreCountersOn(g, saga); got != 1 {
		t.Fatalf("lore counters after entry = %d, want 1", got)
	}

	advanceToPrecombatMainOf(t, g, seat)
	hand := me.Hand.Size()
	pickCard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)

	if got := countersOn(g, land, game.CounterPlusOne); got != hand {
		t.Errorf("+1/+1 counters = %d, want %d (the cards in hand when the chapter resolved)", got, hand)
	}
	power, _, creatureLand, hasty := earthbentBody(t, g, land)
	if !creatureLand || !hasty {
		t.Errorf("the chosen land is not a hasty Land Creature: land+creature=%v haste=%v", creatureLand, hasty)
	}
	if power != hand {
		t.Errorf("power = %d, want %d", power, hand)
	}
	c := findBattlefieldCardByID(g, findBattlefieldByName(g, "Forest"))
	if c == nil || !c.HasSubtype("Island") {
		t.Error(`the chapter's second clause did not apply: "that land becomes an Island in addition to its other types"`)
	}
}

// TestTheLegendOfKyoshiShipsWithoutTheEarthbendCaveat is the catalog
// claim #1178 came to fix. #1179 registered the front face at
// `caveats` with one note about earthbend; the keyword exists now, so
// the face is Full and the note is gone.
func TestTheLegendOfKyoshiShipsWithoutTheEarthbendCaveat(t *testing.T) {
	spec, ok := Lookup(theLegendOfKyoshiOracleID)
	if !ok {
		t.Fatal("The Legend of Kyoshi is not registered")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("Completeness = %q, want %q — earthbend ships now",
			spec.Completeness, CompletenessFull)
	}
	if len(spec.Caveats) != 0 {
		t.Errorf("Caveats = %v, want none", spec.Caveats)
	}
}

// TestHardenedScalesSeesAnEarthbend: the animation registers its layer
// effects and the counters go on straight after, so the placement has
// to see the land as the creature it just became — Hardened Scales
// asks "a creature you control". Before #1282's recompute it read the
// stale cache, saw a plain land, and added nothing.
func TestHardenedScalesSeesAnEarthbend(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", me.ID)
	land := pushEarthbendLand(g, me.ID, "Forest", "Basic Land — Forest")

	castCatalogSpell(t, g, "Earthbending Lesson", "Sorcery — Lesson", earthbendingLessonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}})
	passPriorityAroundTable(t, g)

	if got := countersOn(g, land, game.CounterPlusOne); got != 5 {
		t.Errorf("+1/+1 counters = %d, want 5 — Hardened Scales adds one to an earthbend's counters", got)
	}
}
