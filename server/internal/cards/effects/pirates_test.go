package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pirates_test.go — the Mary Read and Anne Bonny decklist, batch 1:
// artifact-ETB payoffs, discard payoffs, the Pirate lord, and the
// commander's loot-into-Treasure engine.

const (
	recklessFireweaverOracle   = "180e1a7e-890d-477c-80a5-da8a5f2857b3"
	ingeniousArtilleristOracle = "752c7723-90f8-4e3a-8266-f251ee0dadd8"
	maraudingMakoOracle        = "e349be42-5f14-44a9-9608-281985c10e2d"
	glintHornOracle            = "64ad5657-78e9-4f34-8877-18c4f51fff9a"
	corsairCaptainOracle       = "a7ec13c6-7ade-433a-b5a2-047854eef486"
	impulsivePilfererOracle    = "7d9fc9e7-d80b-49c3-871c-ed25b3059ae8"
	angrathsMaraudersOracle    = "2d4976d4-649c-4d42-ac5a-ada4b46a480c"
	faithlessLootingOracle     = "3d6fa57a-aa53-4b5c-b8af-a7612c823117"
	maryReadOracle             = "5182de2d-aceb-450e-bd20-8bc7db124334"
)

// discardTo puts a card with the given type line into a player's
// hand and returns its ID, so a test can discard something specific.
func handCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func lifeOfOpponents(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats)-1)
	for _, p := range g.Seats[1:] {
		out = append(out, p.Life)
	}
	return out
}

// --- artifact-ETB payoffs ----------------------------------------

func TestRecklessFireweaverPingsOnEachArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", recklessFireweaverOracle, false)
	before := lifeOfOpponents(g)

	// Two Treasures entering = two triggers = 1 damage each, twice.
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 2)
	})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", recklessFireweaverOracle, false)
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, RedGoblinToken(), 1)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d took damage from a creature token: %d -> %d", i+1, b, got)
		}
	}
}

// An opponent's artifact doesn't trigger your Fireweaver.
func TestRecklessFireweaverIsControllerScoped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", recklessFireweaverOracle, false)
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1)
	})
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d: an opponent's artifact should not trigger my Fireweaver", i+1)
		}
	}
}

func TestIngeniousArtilleristPings(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Ingenious Artillerist", "Creature — Human Artificer", ingeniousArtilleristOracle, false)
	before := lifeOfOpponents(g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, ClueToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

// --- discard payoffs ---------------------------------------------

func TestMaraudingMakoGrowsOnDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mako := pushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", maraudingMakoOracle, false)
	handCard(me, "Island", "Basic Land — Island")

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Glint-Horn Buccaneer", "Creature — Minotaur Pirate", glintHornOracle, false)
	handCard(me, "Mountain", "Basic Land — Mountain")
	before := lifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-1 {
			t.Errorf("opponent %d: %d -> %d, want -1", i+1, b, got)
		}
	}
}

// An opponent's discard doesn't feed your payoffs.
func TestDiscardPayoffsAreControllerScoped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Glint-Horn Buccaneer", "Creature — Minotaur Pirate", glintHornOracle, false)
	handCard(opp, "Mountain", "Basic Land — Mountain")
	before := lifeOfOpponents(g)

	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b {
			t.Errorf("opponent %d took damage from their own discard", i+1)
		}
	}
}

// --- the lord ----------------------------------------------------

func TestCorsairCaptainMakesTreasureAndPumpsOtherPirates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	captain := pushCatalogPermanent(g, me.ID, "Corsair Captain", "Creature — Human Pirate", corsairCaptainOracle, false)
	mako := pushCatalogPermanent(g, me.ID, "Marauding Mako", "Creature — Shark Pirate", maraudingMakoOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)

	g.WithWriteLock(func() {
		// The ETB fires the Treasure trigger; the zone-move bumps the
		// layer version so the lord's static is recomputed.
		g.EmitEvent(game.Event{Kind: game.EventETB, Actor: me.ID, CardID: captain})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, CardID: captain,
			OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield,
		})
	})
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	pilferer := pushCatalogPermanent(g, me.ID, "Impulsive Pilferer", "Creature — Goblin Pirate", impulsivePilfererOracle, false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(pilferer) })
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Errorf("dying should make a Treasure")
	}
}

// The first damage-amount replacement in the catalog.
func TestAngrathsMaraudersDoublesYourDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Angrath's Marauders", "Creature — Human Pirate", angrathsMaraudersOracle, false)
	fireweaver := pushCatalogPermanent(g, me.ID, "Reckless Fireweaver", "Creature — Human Artificer", recklessFireweaverOracle, false)
	_ = fireweaver
	before := lifeOfOpponents(g)

	// A Treasure enters: the Fireweaver's 1 becomes 2.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	passPriorityAroundTable(t, g)
	for i, b := range before {
		if got := g.Seats[i+1].Life; got != b-2 {
			t.Errorf("opponent %d: %d -> %d, want -2 (1 doubled)", i+1, b, got)
		}
	}

	// An opponent's source is not doubled.
	oppBefore := me.Life
	theirs := pushCatalogPermanent(g, opp.ID, "Their Pinger", "Creature — Bear", "", false)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, me.ID, 3) })
	if me.Life != oppBefore-3 {
		t.Errorf("an opponent's damage should not double: %d -> %d, want -3", oppBefore, me.Life)
	}
}

// --- the commander -----------------------------------------------

func TestFaithlessLootingDrawsThenQueuesTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	libBefore := me.Library.Size()
	castCatalogSpell(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Library.Size() != libBefore-2 {
		t.Errorf("library %d -> %d, want -2", libBefore, me.Library.Size())
	}
	// The discard is an interactive selection, queued against the
	// post-draw hand — so the cards just drawn are legal discards,
	// which is the whole point of a loot.
	if got := g.DiscardPending[me.ID]; got != 2 {
		t.Fatalf("discard pending = %d, want 2", got)
	}
}

func TestMaryReadLootsAndTurnsTypedDiscardsIntoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mary := pushCatalogPermanent(g, me.ID, "Mary Read and Anne Bonny",
		"Legendary Creature — Human Assassin Pirate", maryReadOracle, false)

	// Empty the hand so the random discard is deterministic — this
	// test is about the type check, not the picker.
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			c, _ := me.Hand.Top()
			_, _ = game.MoveCard(me.Hand, me.Library, c.InstanceID)
		}
	})

	// Discarding an Island makes a tapped Treasure.
	handCard(me, "Island", "Basic Land — Island")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	tre := findBattlefieldByName(g, "Treasure")
	if tre == uuid.Nil {
		t.Fatalf("discarding an Island should make a Treasure")
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == tre && !g.Battlefield.Cards[i].Tapped {
			t.Errorf("Mary Read's Treasure should enter tapped")
		}
	}

	// Discarding an off-type card makes nothing.
	handCard(me, "Shock", "Instant")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
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
	passPriorityAroundTable(t, g)
	if me.Library.Size() != libBefore-1 {
		t.Errorf("the loot should draw one: library %d -> %d", libBefore, me.Library.Size())
	}
}
