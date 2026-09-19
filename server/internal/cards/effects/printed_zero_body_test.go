package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// printed_zero_body_test.go — the catalogue's proof for #690 and #691.
//
// game.Card.ToughnessIsKnown is the rule and game/toughness_known_test.go
// is its unit test; this file is the cards that were wrong before it.
//
//   - #690, through the layer pass: Consuming Aberration and Lord of
//     Extinction print `*` and this engine SIZES them, so an empty
//     graveyard makes them real 0/0s and CR 704.5f applies.
//   - #691, through the printing: Hangarback Walker, Wildwood Scourge,
//     Mikaeus and Multani print a real 0/0 and die the moment they
//     enter with nothing holding them up.

// pushPrinted seats a battlefield permanent with a PRINTING behind it
// — the ScryfallID every imported card carries and no bare fixture
// does. It is what tells the engine the 0 in Toughness is printed
// rather than the importer's stand-in.
func pushPrinted(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		ScryfallID: uuid.NewString(),
		Power:      power, Toughness: toughness,
		Owner: owner, Controller: owner,
	})
}

// castPrinted is b12PlayFromHand with a printing behind the card, and
// it resolves the spell before returning.
func castPrinted(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		ScryfallID: uuid.NewString(),
		Owner:      active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// --- #690: coded characteristic-defining abilities -------------------

// Lord of Extinction with every graveyard empty is a 0/0 and dies, as
// its own card comment has claimed since it was written. Before #690
// the toughness check skipped it as a `*` stand-in and it sat on the
// battlefield as an unkillable 0/0.
func TestLordOfExtinctionWithEmptyGraveyardsDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	lord := b12Push(g, me.ID, "Lord of Extinction", "Creature — Elemental", b35LordOfExtinctionOracle, 0, 0)

	runStateChecksViaDraw(t, g)
	if g.Battlefield.Contains(lord) {
		t.Error("Lord of Extinction stayed on the battlefield with every graveyard empty — CR 704.5f (#690)")
	}
	if !me.Graveyard.Contains(lord) {
		t.Error("it did not reach its owner's graveyard")
	}
}

// One card in one graveyard and the Lord is a 1/1 that lives. The
// fix is about the number, not about the card.
func TestLordOfExtinctionWithOneCardInAGraveyardLives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	pushGraveyardCardForTest(me, "Dead One")
	lord := b12Push(g, me.ID, "Lord of Extinction", "Creature — Elemental", b35LordOfExtinctionOracle, 0, 0)

	runStateChecksViaDraw(t, g)
	if !g.Battlefield.Contains(lord) {
		t.Fatal("a 1/1 Lord of Extinction died")
	}
	if got := effectiveToughness(t, g, lord); got != 1 {
		t.Errorf("toughness %d, want 1", got)
	}
}

// Consuming Aberration against opponents who have discarded nothing.
// Same rule, the other CDA the issue names.
func TestConsumingAberrationWithEmptyOpponentGraveyardsDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	// The Aberration reads OPPONENTS' graveyards, so its controller's
	// own dead cards must not save it.
	pushGraveyardCardForTest(me, "My Dead")
	aberration := b12Push(g, me.ID, "Consuming Aberration", "Creature — Horror", b14ConsumingAberrationOracle, 0, 0)

	runStateChecksViaDraw(t, g)
	if g.Battlefield.Contains(aberration) {
		t.Error("Consuming Aberration stayed on the battlefield with every opponent graveyard empty — CR 704.5f (#690)")
	}
}

// The skip covered the whole creature, so a coded `*` creature could
// not be killed by damage either. A 2/2 Lord of Extinction takes two
// and dies (CR 704.5g).
func TestCodedStarCreatureDiesToLethalDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for _, p := range g.Seats {
		p.Graveyard.Cards = nil
	}
	pushGraveyardCardForTest(opp, "Dead A")
	pushGraveyardCardForTest(opp, "Dead B")
	lord := b12Push(g, me.ID, "Lord of Extinction", "Creature — Elemental", b35LordOfExtinctionOracle, 0, 0)
	if got := effectiveToughness(t, g, lord); got != 2 {
		t.Fatalf("toughness %d, want 2", got)
	}

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, lord, 2) })
	runStateChecksViaDraw(t, g)
	if g.Battlefield.Contains(lord) {
		t.Error("a 2/2 Lord of Extinction survived two damage — the stand-in skip was skipping its damage check too")
	}
}

// --- #691: a printed 0/0 with a printing behind it -------------------

// Hangarback Walker cast for X=0 (CR 601.2b: a legal announcement)
// enters with no counters, is the 0/0 it prints, and is put into its
// owner's graveyard at once — making no Thopters, because it had no
// counters when it died.
func TestHangarbackWalkerCastForXZeroDiesAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := castPrinted(t, g, "Hangarback Walker", "Artifact Creature — Construct",
		b14HangarbackWalkerOracle, game.CastSpellParams{XValue: 0})

	if g.Battlefield.Contains(walker) {
		t.Error("a Hangarback Walker cast for X=0 stayed on the battlefield — CR 704.5f (#691)")
	}
	if !me.Graveyard.Contains(walker) {
		t.Error("it did not reach its owner's graveyard")
	}
	if n := countBattlefieldNamed(g, me.ID, "Thopter"); n != 0 {
		t.Errorf("no +1/+1 counters means no Thopters, got %d", n)
	}
}

// The same Walker cast for X=1 enters as a 1/1 and lives. X=0 is the
// only announcement that kills it.
func TestHangarbackWalkerCastForXOneLives(t *testing.T) {
	g := newCatalogGame(t)
	walker := castPrinted(t, g, "Hangarback Walker", "Artifact Creature — Construct",
		b14HangarbackWalkerOracle, game.CastSpellParams{XValue: 1})

	if !g.Battlefield.Contains(walker) {
		t.Fatal("a Hangarback Walker cast for X=1 died")
	}
	if got := b12Counter(t, g, walker, "+1/+1"); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

// Wildwood Scourge cast for X=0 dies to the toughness check, which is
// what its card comment has always said and what the engine did not
// do until #691.
func TestWildwoodScourgeCastForXZeroDiesAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scourge := castPrinted(t, g, "Wildwood Scourge", "Creature — Hydra",
		b35WildwoodScourgeOracle, game.CastSpellParams{XValue: 0})

	if g.Battlefield.Contains(scourge) {
		t.Error("a Wildwood Scourge cast for X=0 stayed on the battlefield — CR 704.5f (#691)")
	}
	if !me.Graveyard.Contains(scourge) {
		t.Error("it did not reach its owner's graveyard")
	}
}

// Mikaeus, the Lunarch cast for X=0 dies rather than sticking around
// to be grown with his tap ability. He carried a player-facing caveat
// saying so; #691 deletes it.
func TestMikaeusCastForXZeroDiesAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	mikaeus := castPrinted(t, g, "Mikaeus, the Lunarch", "Legendary Creature — Human Cleric",
		"82f3faa8-39fa-450b-843f-d60a4c36d8f7", game.CastSpellParams{XValue: 0})

	if g.Battlefield.Contains(mikaeus) {
		t.Error("Mikaeus cast for X=0 stayed on the battlefield — CR 704.5f (#691)")
	}
	for _, spec := range []string{"82f3faa8-39fa-450b-843f-d60a4c36d8f7"} {
		s, ok := Lookup(spec)
		if !ok {
			t.Fatalf("no spec for %s", spec)
		}
		for _, c := range s.Caveats {
			if containsFold(c, "X=0") {
				t.Errorf("Mikaeus still declares an X=0 caveat: %q", c)
			}
		}
	}
}

// Multani, Yavimaya's Avatar is the same printed 0/0 reached by a
// layer 7c modify rather than an entry counter: with no land anywhere
// he is a 0/0 and dies.
func TestMultaniWithNoLandsDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	multani := pushPrinted(g, me.ID, "Multani, Yavimaya's Avatar",
		"Legendary Creature — Elemental Avatar", "4b8bf64b-4800-45ff-81c6-2857f34999b5", 0, 0)

	runStateChecksViaDraw(t, g)
	if g.Battlefield.Contains(multani) {
		t.Error("Multani with no lands anywhere stayed on the battlefield — CR 704.5f (#691)")
	}
}

// One land and Multani is a 1/1 that lives.
func TestMultaniWithOneLandLives(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	multani := pushPrinted(g, me.ID, "Multani, Yavimaya's Avatar",
		"Legendary Creature — Elemental Avatar", "4b8bf64b-4800-45ff-81c6-2857f34999b5", 0, 0)

	runStateChecksViaDraw(t, g)
	if !g.Battlefield.Contains(multani) {
		t.Fatal("a 1/1 Multani died")
	}
	if got := effectiveToughness(t, g, multani); got != 1 {
		t.Errorf("toughness %d, want 1", got)
	}
}

// A declined Clone enters as the printed 0/0 Clone actually is, and
// it dies, which is what the printed card does. It used to survive —
// a deviation clone.go declared to players — for the same reason
// every printed 0/0 did.
func TestDeclinedCloneDiesAsItsPrintedZeroZero(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Clone", TypeLine: "Creature — Shapeshifter",
		OracleID: oracleClone, ScryfallID: uuid.NewString(),
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Clone: %v", err)
	}
	resolveWithCopyChoice(t, g, uuid.Nil)
	runStateChecksViaDraw(t, g)

	if g.Battlefield.Contains(id) {
		t.Error("a declined Clone stayed on the battlefield as a 0/0 — CR 704.5f")
	}
}

// A 0/0 fixture with NO printing behind it is still skipped: the
// engine has no toughness for it, and every catalogue test that seats
// a body it does not care about depends on that staying true.
func TestFixtureWithNoPrintingIsStillSkipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bodyless Bear", "Creature — Bear", 0, 0)

	runStateChecksViaDraw(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a 0/0 fixture with no printing was killed — the stand-in skip regressed")
	}
}
