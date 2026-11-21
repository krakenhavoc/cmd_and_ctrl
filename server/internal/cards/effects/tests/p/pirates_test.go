package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- artifact-ETB payoffs ----------------------------------------

func TestRecklessFireweaverPingsOnEachArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", effectstest.RecklessFireweaverOracle, false)
	before := effectstest.LifeOfOpponents(g)

	// Two Treasures entering = two triggers = 1 damage each, twice.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 2)
	})
	effectstest.PassPriorityAroundTable(t, g)

	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2", i+1, b, got)
		}
	}
	if me.Life != 40 && me.Life != g.Seats[0].Life {
		t.Errorf("the Fireweaver's controller should be untouched")
	}
}

// A creature entering is not an artifact entering.
func TestRecklessFireweaverIgnoresNonArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", effectstest.RecklessFireweaverOracle, false)
	before := effectstest.LifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.RedGoblinToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d took damage from a creature token: %d -> %d", i+1, b, got)
		}
	}
}

// An opponent's artifact doesn't trigger your Fireweaver.
func TestRecklessFireweaverIsControllerScoped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", effectstest.RecklessFireweaverOracle, false)
	before := effectstest.LifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d: an opponent's artifact should not trigger my Fireweaver", i+1)
		}
	}
}

func TestIngeniousArtilleristPings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Ingenious Artillerist", "Creature — Human Artificer", effectstest.IngeniousArtilleristOracle, false)
	before := effectstest.LifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.ClueToken(), 1) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

// --- discard payoffs ---------------------------------------------

func TestMaraudingMakoGrowsOnDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mako := effectstest.PushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", effectstest.MaraudingMakoOracle, false)
	effectstest.HandCard(me, "Island", "Basic Land — Island")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)

	var counters int
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == mako {
			counters = g.Battlefield.Cards[i].Counters["+1/+1"]
		}
	}
	if counters != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", counters)
	}
}

func TestGlintHornPingsOnDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Glint-Horn Buccaneer", "Creature — Minotaur Pirate", effectstest.GlintHornOracle, false)
	effectstest.HandCard(me, "Mountain", "Basic Land — Mountain")
	before := effectstest.LifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

// An opponent's discard doesn't feed your payoffs.
func TestDiscardPayoffsAreControllerScoped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Glint-Horn Buccaneer", "Creature — Minotaur Pirate", effectstest.GlintHornOracle, false)
	effectstest.HandCard(opp, "Mountain", "Basic Land — Mountain")
	before := effectstest.LifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d took damage from their own discard", i+1)
		}
	}
}

// --- the lord ----------------------------------------------------

func TestCorsairCaptainMakesTreasureAndPumpsOtherPirates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	captain := effectstest.PushCatalogPermanent(g, me.ID, "Corsair Captain", "Creature — Human Pirate", effectstest.CorsairCaptainOracle, false)
	mako := effectstest.PushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", effectstest.MaraudingMakoOracle, false)
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() {
		// The ETB fires the Treasure trigger; the zone-move bumps the
		// layer version so the lord's static is recomputed.
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: captain})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, CardID: captain,
			OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield,
		})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Errorf("the Captain's ETB should make a Treasure")
	}

	// ReadSnapshot forces the layer recompute, the way the anthem
	// tests do.
	eff := func(id uuid.UUID) (int, int) {
		p, tou := -1, -1
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == id {
					e := c.Effective()
					p, tou = e.Power, e.Toughness
					return
				}
			}
		})
		return p, tou
	}
	// The Mako is a Pirate: pumped. The Bear isn't. The Captain is a
	// Pirate but excluded by "other".
	if p, tou := eff(mako); p != 2 || tou != 2 {
		t.Errorf("Mako under the lord = %d/%d, want 2/2", p, tou)
	}
	if p, tou := eff(bear); p != 1 || tou != 1 {
		t.Errorf("non-Pirate should be untouched, got %d/%d", p, tou)
	}
	if p, tou := eff(captain); p != 1 || tou != 1 {
		t.Errorf("the Captain must not pump itself, got %d/%d", p, tou)
	}
}

// --- ramp + doubler ----------------------------------------------

func TestImpulsivePilfererMakesTreasureOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	pilferer := effectstest.PushCatalogPermanent(g, me.ID, "Impulsive Pilferer", "Creature — Goblin Pirate", effectstest.ImpulsivePilfererOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pilferer) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Errorf("dying should make a Treasure")
	}
}

// The first damage-amount replacement in the catalog.
func TestAngrathsMaraudersDoublesYourDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Angrath's Marauders", "Creature — Human Pirate", effectstest.AngrathsMaraudersOracle, false)
	fireweaver := effectstest.PushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", effectstest.RecklessFireweaverOracle, false)
	_ = fireweaver
	before := effectstest.LifeOfOpponents(g)

	// A Treasure enters: the Fireweaver's 1 becomes 2.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2 (1 doubled)", i+1, b, got)
		}
	}

	// An opponent's source is not doubled.
	oppBefore := me.Life
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Pinger", "Creature — Bear", "", false)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 3) })
	if me.Life != oppBefore-3 {
		t.Errorf("an opponent's damage should not double: %d -> %d, want -3", oppBefore, me.Life)
	}
}

// --- the commander -----------------------------------------------

func TestFaithlessLootingDrawsThenQueuesTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	libBefore := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Faithless Looting", "Sorcery", effectstest.FaithlessLootingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if me.Library.Size() != libBefore-2 {
		t.Errorf("library %d -> %d, want -2", libBefore, me.Library.Size())
	}
	// The discard is an interactive selection, queued against the
	// post-draw hand — so the cards just drawn are legal discards,
	// which is the whole point of a loot.
	if got := effectstest.DiscardOwed(g, me.ID); got != 2 {
		t.Fatalf("discard pending = %d, want 2", got)
	}
}

func TestMaryReadLootsAndTurnsTypedDiscardsIntoTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mary := effectstest.PushCatalogPermanent(g, me.ID, "Mary Read and Anne Bonny",
		"Legendary Creature — Human Assassin Pirate", effectstest.MaryReadOracle, false)

	// Empty the hand so the random discard is deterministic — this
	// test is about the type check, not the picker.
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			c, _ := me.Hand.Top()
			_, _ = game.MoveCard(me.Hand, me.Library, c.InstanceID)
		}
	})

	// Discarding an Island makes a tapped Treasure.
	effectstest.HandCard(me, "Island", "Basic Land — Island")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	tre := effectstest.FindBattlefieldByName(g, "Treasure")
	if tre == uuid.Nil {
		t.Fatalf("discarding an Island should make a Treasure")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == tre && !g.Battlefield.Cards[i].Tapped {
			t.Errorf("Mary Read's Treasure should enter tapped")
		}
	}

	// Discarding an off-type card makes nothing.
	effectstest.HandCard(me, "Shock", "Instant")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	count := 0
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Treasure" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("an off-type discard made a Treasure: %d on the battlefield", count)
	}

	// The tap ability loots.
	libBefore := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, mary, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Mary Read's loot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore-1 {
		t.Errorf("the loot should draw one: library %d -> %d", libBefore, me.Library.Size())
	}
}
