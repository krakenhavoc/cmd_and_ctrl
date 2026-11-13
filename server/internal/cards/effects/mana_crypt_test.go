package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mana Crypt (Eternal Masters, #5542).

const manaCryptOracle = "2c63e4e1-89d2-4bc6-a232-94e75c4b1c8a"

func TestManaCryptIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(manaCryptOracle)
	if !ok {
		t.Fatal("Mana Crypt is not registered")
	}
	if spec.Name != "Mana Crypt" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// manaCryptUpkeepFlip walks a fresh table to the upkeep of seat 1, who
// controls a Mana Crypt, runs `before` while the trigger waits on the
// stack, resolves it, calls heads and reports whether the flip was won,
// with seat 1's and seat 2's life before the trigger resolved.
func manaCryptUpkeepFlip(t *testing.T, before func(g *game.Game, crypt uuid.UUID)) (g *game.Game, won bool, life, otherLife int) {
	t.Helper()
	g = newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	crypt := pushCatalogPermanent(g, me.ID, "Mana Crypt", "Artifact", manaCryptOracle, false)
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, crypt) == nil {
		t.Fatal("no upkeep trigger on its controller's upkeep")
	}
	if before != nil {
		before(g, crypt)
	}
	life, otherLife = me.Life, other.Life
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if ask == nil {
		t.Fatal("no coin call was asked")
	}
	if ask.Chooser != me.ID {
		t.Fatalf("the coin call was asked of %s, want the Crypt's controller", ask.Chooser)
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
	return g, evs[len(evs)-1].Won, life, otherLife
}

// TestManaCryptDealsThreeToYouOnlyWhenYouLoseTheFlip flips on fresh
// tables until it has seen both faces: a won flip changes nothing, a lost
// one costs its controller 3 life and nobody else any.
func TestManaCryptDealsThreeToYouOnlyWhenYouLoseTheFlip(t *testing.T) {
	lost, won := 0, 0
	for table := 0; table < 40 && (lost == 0 || won == 0); table++ {
		g, w, life, otherLife := manaCryptUpkeepFlip(t, nil)
		me, other := g.Seats[1], g.Seats[2]
		want := life - 3
		if w {
			won++
			want = life
		} else {
			lost++
		}
		if me.Life != want {
			t.Errorf("won %v: controller life %d -> %d, want %d", w, life, me.Life, want)
		}
		if other.Life != otherLife {
			t.Errorf("won %v: another player's life %d -> %d", w, otherLife, other.Life)
		}
	}
	if lost == 0 || won == 0 {
		t.Fatalf("saw %d lost and %d won flips; want both", lost, won)
	}
}

// TestManaCryptStillDealsDamageAfterItLeaves: the Crypt bounced with
// its trigger on the stack still flips, and a lost flip still deals the
// 3 damage, from the Crypt as it last existed.
func TestManaCryptStillDealsDamageAfterItLeaves(t *testing.T) {
	for table := 0; table < 40; table++ {
		g, w, life, _ := manaCryptUpkeepFlip(t, func(g *game.Game, crypt uuid.UUID) {
			g.WithWriteLock(func() { _ = g.BounceToHandForEffect(crypt) })
			if g.Battlefield.Contains(crypt) {
				t.Fatal("the Crypt is still on the battlefield")
			}
		})
		if w {
			continue
		}
		if me := g.Seats[1]; me.Life != life-3 {
			t.Errorf("a lost flip after the Crypt left: life %d -> %d, want 3 damage", life, me.Life)
		}
		return
	}
	t.Fatal("no flip was lost")
}

func TestManaCryptDoesNotTriggerOnAnotherPlayersUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[2]
	crypt := pushCatalogPermanent(g, owner.ID, "Mana Crypt", "Artifact", manaCryptOracle, false)
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, crypt) != nil {
		t.Fatal("the Crypt triggered on another player's upkeep")
	}
}

func TestManaCryptTapsForTwoColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	crypt := pushCatalogPermanent(g, me.ID, "Mana Crypt", "Artifact", manaCryptOracle, false)
	if err := g.ActivateManaAbility(me.ID, crypt, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("pool %v, want {C}{C}", got)
	}
	if !mustBattlefieldCard(t, g, crypt).Tapped {
		t.Error("the Crypt is untapped after tapping for mana")
	}
}
