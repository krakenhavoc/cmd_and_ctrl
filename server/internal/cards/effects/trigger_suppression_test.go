package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_suppression_test.go — #1735, the proof cards for the trigger
// suppressor: Torpor Orb, Hushwing Gryff, Tocatli Honor Guard,
// Hushbringer, and the third paragraph of Elesh Norn, Mother of
// Machines. The engine rules (the order against the doubler and the
// once-per-batch guard, which board is asked, a delayed trigger) are
// pinned in game/trigger_suppression_test.go. These tests use the real
// cards against real catalog triggers.

const (
	torporOrbOracle         = "97326cad-b13c-4e52-82ce-850a39e5ff08"
	hushwingGryffOracle     = "a0d615b3-7ee3-43d7-ad37-5f7f1544a8db"
	tocatliHonorGuardOracle = "ac946047-4933-4b55-a4eb-85abde1839bb"
	hushbringerOracle       = "3d21f710-6bbc-41ae-a8e5-02debe3e02bd"
	doorkeeperThrullOracle  = "0eaea4fd-a377-4541-9bdd-14921034a975"
	suppSoulWardenOracle    = "f3fad295-1af2-4ecc-8546-b121ad6be27b"
)

// suppSeats returns the active seat and the next one, an opponent.
func suppSeats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func suppPermanent(g *game.Game, controller uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: controller, Controller: controller,
	})
}

// castMulldrifterExpectingNoDraw casts Mulldrifter for the active seat
// and checks that the hand is the same size after it resolves as it
// was with the spell on the stack, so its "draw two" never triggered.
func castMulldrifterExpectingNoDraw(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me, _ := suppSeats(g)
	id := castWithCost(t, g, "Mulldrifter", "Creature — Elemental", "{4}{U}", mulldrifterOracle)
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	// passPriorityAroundTable stops at a prompt, so an order or target
	// prompt here would leave the hand untouched for the wrong reason.
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts are waiting, want none: something triggered", len(g.PendingChoices))
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size = %d, want %d: Mulldrifter's enters trigger should not have triggered", got, before)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("Mulldrifter is not on the battlefield")
	}
	return id
}

// --- Torpor Orb ---------------------------------------------------------

func TestTorporOrbStopsACreaturesOwnEntersTrigger(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := suppSeats(g)
	suppPermanent(g, opp.ID, "Torpor Orb", "Artifact", torporOrbOracle)
	castMulldrifterExpectingNoDraw(t, g)
}

// Evoke's sacrifice is the creature's own enters trigger (CR 702.74a),
// so under Torpor Orb an evoked Mulldrifter neither draws nor is
// sacrificed. This is the interaction the card is known for.
func TestTorporOrbKeepsAnEvokedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := suppSeats(g)
	suppPermanent(g, me.ID, "Torpor Orb", "Artifact", torporOrbOracle)
	id := castWithAltCost(t, g, "Mulldrifter", "Creature — Elemental", mulldrifterOracle, "evoke")
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("an evoked creature under Torpor Orb left %d prompts, want none", len(g.PendingChoices))
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the evoked Mulldrifter was sacrificed under Torpor Orb")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size = %d, want %d: the evoked Mulldrifter drew", got, before)
	}
}

// A Treasure is not a creature: Torpor Orb lets Reckless Fireweaver ping
// for it. Doorkeeper Thrull (below) is the one that stops it.
func TestTorporOrbLetsAnArtifactEnteringTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	suppPermanent(g, me.ID, "Torpor Orb", "Artifact", torporOrbOracle)
	pushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", recklessFireweaverOracle, false)
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1: a Treasure entering is not a creature entering", i+1, b, got)
		}
	}
}

// --- Doorkeeper Thrull ----------------------------------------------------

// "Artifacts and creatures entering": a Treasure being created no
// longer triggers Reckless Fireweaver, and a creature's own enters
// trigger is stopped too.
func TestDoorkeeperThrullStopsArtifactsAndCreaturesEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := suppSeats(g)
	thrull := suppPermanent(g, opp.ID, "Doorkeeper Thrull", "Creature — Thrull", doorkeeperThrullOracle)
	pushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", recklessFireweaverOracle, false)
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts are waiting, want none", len(g.PendingChoices))
	}
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d: %d -> %d: Reckless Fireweaver triggered under Doorkeeper Thrull", i+1, b, got)
		}
	}
	castMulldrifterExpectingNoDraw(t, g)
	for _, kw := range []string{"flash", "flying"} {
		if !hasEffectiveKeyword(t, g, thrull, kw) {
			t.Errorf("Doorkeeper Thrull lacks %s", kw)
		}
	}
}

// --- Tocatli Honor Guard ---------------------------------------------

// Another permanent's "whenever another creature enters" is stopped:
// Soul Warden gains nothing.
func TestTocatliHonorGuardStopsSoulWarden(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := suppSeats(g)
	suppPermanent(g, opp.ID, "Tocatli Honor Guard", "Creature — Human Soldier", tocatliHonorGuardOracle)
	suppPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", suppSoulWardenOracle)
	life := me.Life
	castWithCost(t, g, "Bear", "Creature — Bear", "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life = %d, want %d: Soul Warden triggered under Tocatli Honor Guard", me.Life, life)
	}
}

// --- Hushwing Gryff ---------------------------------------------------

// Flash is how the Gryff is played: cast in response to a creature
// spell, it resolves first, and the creature then enters with nothing
// triggering.
func TestHushwingGryffFlashedInStopsTheCreatureBelowIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := suppSeats(g)
	castWithCost(t, g, "Mulldrifter", "Creature — Elemental", "{4}{U}", mulldrifterOracle)
	before := me.Hand.Size()
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass to the opponent: %v", err)
	}
	if g.Turn.PriorityHolder != (g.Turn.ActiveSeat+1)%len(g.Seats) {
		t.Fatalf("setup: priority is with seat %d, want the next seat", g.Turn.PriorityHolder)
	}
	gryff := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: gryff, Name: "Hushwing Gryff", TypeLine: "Creature — Hippogriff",
		ManaCost: "{2}{W}", OracleID: hushwingGryffOracle, Power: 2, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, gryff, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash in Hushwing Gryff: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gryff) {
		t.Fatal("Hushwing Gryff did not resolve")
	}
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size = %d, want %d: Mulldrifter drew under Hushwing Gryff", got, before)
	}
	for _, kw := range []string{"flash", "flying"} {
		if !hasEffectiveKeyword(t, g, gryff, kw) {
			t.Errorf("Hushwing Gryff lacks %s", kw)
		}
	}
}

// --- Hushbringer --------------------------------------------------------

// Both halves: a creature dying does not trigger Blood Artist (no target
// prompt, no drain), and a creature entering does not trigger
// Mulldrifter.
func TestHushbringerStopsDiesAndEntersTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := suppSeats(g)
	hush := suppPermanent(g, opp.ID, "Hushbringer", "Creature — Faerie", hushbringerOracle)
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	victim := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	meLife, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 || me.Life != meLife || opp.Life != oppLife {
		t.Errorf("Blood Artist triggered under Hushbringer: choices=%d me %d→%d opp %d→%d",
			len(g.PendingChoices), meLife, me.Life, oppLife, opp.Life)
	}

	castMulldrifterExpectingNoDraw(t, g)
	for _, kw := range []string{"flying", "lifelink"} {
		if !hasEffectiveKeyword(t, g, hush, kw) {
			t.Errorf("Hushbringer lacks %s", kw)
		}
	}
}

// Without a suppressor the same death does trigger Blood Artist. This
// is the control for the test above: it is Hushbringer that stops the
// trigger, not the setup.
func TestBloodArtistStillTriggersWithoutHushbringer(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := suppSeats(g)
	pushCatalogPermanent(g, me.ID, "Blood Artist", "Creature — Vampire", bloodArtistOracle, false)
	victim := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	if passUntilPickTarget(t, g, me.ID) == nil {
		t.Fatal("Blood Artist did not trigger on a creature's death")
	}
}

// --- Elesh Norn, Mother of Machines ------------------------------------

// The third paragraph: an opponent's permanent's own enters trigger is
// stopped. The opponent here is the active seat casting Mulldrifter.
func TestEleshNornStopsAnOpponentsEntersTrigger(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := suppSeats(g)
	pushEleshNorn(g, opp.ID)
	castMulldrifterExpectingNoDraw(t, g)
}

// An opponent's evoked creature stays: its evoke sacrifice is its own
// enters trigger, and it is an ability of a permanent that Elesh Norn's
// controller's opponent controls.
func TestEleshNornKeepsAnOpponentsEvokedCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := suppSeats(g)
	pushEleshNorn(g, opp.ID)
	id := castWithAltCost(t, g, "Mulldrifter", "Creature — Elemental", mulldrifterOracle, "evoke")
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts are waiting, want none: the evoked Mulldrifter's triggers triggered", len(g.PendingChoices))
	}
	if !g.Battlefield.Contains(id) {
		t.Error("an opponent's evoked Mulldrifter was sacrificed under Elesh Norn")
	}
}

// Both paragraphs across two players: my Soul Warden triggers twice and
// the opponent's does not trigger at all.
func TestEleshNornDoublesMineAndStopsTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := suppSeats(g)
	pushEleshNorn(g, me.ID)
	suppPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", suppSoulWardenOracle)
	suppPermanent(g, opp.ID, "Soul Warden", "Creature — Human Cleric", suppSoulWardenOracle)
	meLife, oppLife := me.Life, opp.Life
	castWithCost(t, g, "Bear", "Creature — Bear", "{1}{G}", "")
	passPriorityAroundTable(t, g)
	if me.Life != meLife+2 {
		t.Errorf("my life = %d, want %d: my Soul Warden should trigger twice", me.Life, meLife+2)
	}
	if opp.Life != oppLife {
		t.Errorf("opponent's life = %d, want %d: their Soul Warden should not trigger", opp.Life, oppLife)
	}
}

// Two Elesh Norns on opposite sides: each stops the other player's
// triggers, so nothing is left for either to double.
func TestTwoEleshNornsOnOppositeSidesStopEverything(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := suppSeats(g)
	pushEleshNorn(g, me.ID)
	pushEleshNorn(g, opp.ID)
	suppPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", suppSoulWardenOracle)
	meLife := me.Life
	castMulldrifterExpectingNoDraw(t, g)
	if me.Life != meLife {
		t.Errorf("my life = %d, want %d: my Soul Warden triggered under the opponent's Elesh Norn", me.Life, meLife)
	}
}

func pushEleshNorn(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elesh Norn, Mother of Machines", TypeLine: "Legendary Creature — Phyrexian Praetor",
		OracleID: eleshNornOracle, Power: 4, Toughness: 7, Owner: controller, Controller: controller,
	})
}
