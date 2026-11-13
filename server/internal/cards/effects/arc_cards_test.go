package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Archenemy (ARC) card coverage, #5580.

const (
	arcChandrasOutrageOracle    = "47437865-0032-4f47-b0ab-034cc841bb84"
	arcPlummetOracle            = "85bde6ac-3dd4-4946-8b57-24f57e3eae2b"
	arcSorcerersStrongboxOracle = "c4f58768-35b2-4e17-9129-c9e39b8ed339"
)

func TestArcEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		arcChandrasOutrageOracle:    "Chandra's Outrage",
		arcPlummetOracle:            "Plummet",
		arcSorcerersStrongboxOracle: "Sorcerer's Strongbox",
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

// --- Chandra's Outrage ------------------------------------------------------

func TestArcChandrasOutrageHitsTheCreatureAndItsController(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	before := opp.Life
	castCatalogSpell(t, g, "Chandra's Outrage", "Instant", arcChandrasOutrageOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(big)
	if !ok || !g.Battlefield.Contains(big) || c.DamageMarked != 4 {
		t.Errorf("the 5/5 has %d damage marked (on battlefield %v), want 4", c.DamageMarked, g.Battlefield.Contains(big))
	}
	if opp.Life != before-2 {
		t.Errorf("its controller's life %d -> %d, want 2 damage", before, opp.Life)
	}
}

func TestArcChandrasOutrageHitsTheControllerOfACreatureItKills(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	castCatalogSpell(t, g, "Chandra's Outrage", "Instant", arcChandrasOutrageOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("4 damage did not kill the 2/2")
	}
	if opp.Life != before-2 {
		t.Errorf("its controller's life %d -> %d, want 2 damage", before, opp.Life)
	}
}

func TestArcChandrasOutrageDoesNothingWhenItsTargetLeft(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	castCatalogSpell(t, g, "Chandra's Outrage", "Instant", arcChandrasOutrageOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Errorf("a spell whose target left dealt %d damage to the player", before-opp.Life)
	}
}

// --- Plummet ----------------------------------------------------------------

func TestArcPlummetDestroysOnlyAFlyer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bird := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"},
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
	})
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Plummet", "Instant", arcPlummetOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("a creature without flying was accepted as the target")
	}
	castCatalogSpell(t, g, "Plummet", "Instant", arcPlummetOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bird}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flyer was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature without flying left the battlefield")
	}
}

// --- Sorcerer's Strongbox ---------------------------------------------------

// arcStrongboxFlip activates the Strongbox as me, resolves the ability,
// runs `before` while the coin call is open, calls heads and reports
// whether the flip was won.
func arcStrongboxFlip(t *testing.T, g *game.Game, me *game.Player, box uuid.UUID, before func()) bool {
	t.Helper()
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(box) })
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, box, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if ask == nil {
		t.Fatal("no coin call was asked")
	}
	if before != nil {
		before()
	}
	flips := len(randomEvents(g, game.EventFlipCoin))
	if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	passPriorityAroundTable(t, g)
	evs := randomEvents(g, game.EventFlipCoin)
	if len(evs) != flips+1 {
		t.Fatalf("%d coins were flipped, want 1", len(evs)-flips)
	}
	return evs[len(evs)-1].Won
}

// TestArcSorcerersStrongboxDrawsThreeAndGoesOnAWin flips until it has seen
// both faces: a lost flip leaves the Strongbox and the hand alone, a won
// one sacrifices it and draws three. A won flip ends a table, so a run
// that has not lost yet starts a fresh one; every table winning its first
// flip, 40 times running, is the only way to miss the lost branch.
func TestArcSorcerersStrongboxDrawsThreeAndGoesOnAWin(t *testing.T) {
	lost, won := 0, 0
	for table := 0; table < 40 && (lost == 0 || won == 0); table++ {
		g := newCatalogGame(t)
		me := g.Seats[0]
		box := pushCatalogPermanent(g, me.ID, "Sorcerer's Strongbox", "Artifact", arcSorcerersStrongboxOracle, false)
		for i := 0; i < 40; i++ {
			hand := me.Hand.Size()
			if !arcStrongboxFlip(t, g, me, box, nil) {
				lost++
				if !g.Battlefield.Contains(box) || me.Hand.Size() != hand {
					t.Fatalf("a lost flip: on battlefield %v, hand %d -> %d", g.Battlefield.Contains(box), hand, me.Hand.Size())
				}
				continue
			}
			won++
			if g.Battlefield.Contains(box) || !me.Graveyard.Contains(box) {
				t.Error("a won flip did not sacrifice the Strongbox")
			}
			if me.Hand.Size() != hand+3 {
				t.Errorf("a won flip: hand %d -> %d, want three cards drawn", hand, me.Hand.Size())
			}
			break
		}
	}
	if lost == 0 || won == 0 {
		t.Fatalf("saw %d lost and %d won flips; want both", lost, won)
	}
}

// TestArcSorcerersStrongboxStillDrawsWhenItIsNoLongerYours is the ruling:
// a Strongbox someone else controls by the time the flip is won is not
// sacrificed, and its activator still draws three cards.
func TestArcSorcerersStrongboxStillDrawsWhenItIsNoLongerYours(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	box := pushCatalogPermanent(g, me.ID, "Sorcerer's Strongbox", "Artifact", arcSorcerersStrongboxOracle, false)
	steal := func() {
		g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, box, opp.ID, game.IndefiniteDuration(), "theft") })
	}
	for i := 0; i < 40; i++ {
		hand, oppHand := me.Hand.Size(), opp.Hand.Size()
		if !arcStrongboxFlip(t, g, me, box, steal) {
			g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, box, me.ID, game.IndefiniteDuration(), "return") })
			continue
		}
		c, ok := g.LookupCardForEffect(box)
		if !g.Battlefield.Contains(box) || !ok || c.Controller != opp.ID {
			t.Error("a Strongbox its activator no longer controls was sacrificed")
		}
		if me.Hand.Size() != hand+3 || opp.Hand.Size() != oppHand {
			t.Errorf("hands: mine %d -> %d (want +3), theirs %d -> %d (want unchanged)", hand, me.Hand.Size(), oppHand, opp.Hand.Size())
		}
		return
	}
	t.Fatal("no flip was won")
}
