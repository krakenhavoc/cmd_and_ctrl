package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// malcolm_alluring_scoundrel_test.go — the counter, the loot and the
// threshold. The free cast is asserted all the way through to a
// resolved spell, because a permission that renders but cannot be
// claimed would look identical from the graveyard.

const malcolmOracle = "3bba3b36-f8d7-4fd6-892a-797226210b6e"

// pushMalcolm seeds Malcolm ready to act, with `chorus` counters
// already on him.
func pushMalcolm(t *testing.T, g *game.Game, controller uuid.UUID, chorus int) uuid.UUID {
	t.Helper()
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Malcolm, Alluring Scoundrel",
		TypeLine: "Legendary Creature — Siren Pirate", OracleID: malcolmOracle,
		ManaCost: "{1}{U}", Power: 2, Toughness: 1,
		Owner: controller, Controller: controller,
	})
	if chorus > 0 {
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(id, malcolmChorusCounter, chorus); err != nil {
				t.Fatalf("seed chorus counters: %v", err)
			}
		})
	}
	return id
}

func TestMalcolmHasFlashAndFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushMalcolm(t, g, me.ID, 0)
	got := effectiveAbilities(t, g, id)
	for _, want := range []string{"flash", "flying"} {
		found := false
		for _, a := range got {
			if a == want {
				found = true
			}
		}
		if !found {
			t.Errorf("abilities %v missing %q", got, want)
		}
	}
}

// TestMalcolmCountsThenLootsAndOffersNothingBelowFour is the ordinary
// hit: one counter, draw one, discard one, and no free cast.
func TestMalcolmCountsThenLootsAndOffersNothingBelowFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushMalcolm(t, g, me.ID, 0)
	advanceToMain(t, g)

	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	dealCombatDamageToPlayer(g, id, opp.ID, 2)
	passPriorityAroundTable(t, g)

	if c := findTestCard(g, id); c == nil || c.Counters[malcolmChorusCounter] != 1 {
		t.Fatalf("chorus counters after one hit = %v, want 1", c)
	}
	// The draw is above the discard, so the prompt is over a hand that
	// already holds the new card.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand while the discard prompt is open = %d, want %d — draw, THEN discard", got, hand+1)
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if got := me.Graveyard.Size(); got != graves+1 {
		t.Errorf("graveyard %d → %d, want one card discarded", graves, got)
	}
	if offer := latestChoiceOfKind(g, game.PendingChoiceMayCast); offer != nil {
		t.Error("a free cast was offered at one chorus counter; the card prints four or more")
	}
}

// TestMalcolmAtFourChorusCountersCastsTheDiscardedCardForFree is the
// last sentence, end to end: the fourth counter arrives, the offer is
// made, and taking it lets the discarded card be cast out of the
// graveyard for nothing.
func TestMalcolmAtFourChorusCountersCastsTheDiscardedCardForFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := pushMalcolm(t, g, me.ID, 3)
	advanceToMain(t, g)

	// A real spell to pitch, so the free cast has something to prove.
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: pitch, Name: "Pitched Spell", TypeLine: "Instant",
		ManaCost: "{4}{U}{U}", Owner: me.ID, Controller: me.ID,
	})

	dealCombatDamageToPlayer(g, id, opp.ID, 2)
	passPriorityAroundTable(t, g)

	if c := findTestCard(g, id); c == nil || c.Counters[malcolmChorusCounter] != 4 {
		t.Fatalf("chorus counters = %v, want 4", c)
	}
	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(pitch) {
		t.Fatal("the pitched card is not in the graveyard")
	}
	offer := latestChoiceOfKind(g, game.PendingChoiceMayCast)
	if offer == nil {
		t.Fatalf("no free cast was offered at four chorus counters: %+v", g.PendingChoices)
	}
	if offer.MayCastCard != pitch {
		t.Errorf("the offer names %v, want the discarded card %v", offer.MayCastCard, pitch)
	}
	if err := g.ResolveMayCast(offer.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}

	var perm *game.CastPermission
	g.ReadSnapshot(func() { perm = g.CastPermissionOnCardByIDForEffect(pitch) })
	if perm == nil {
		t.Fatal("taking the offer granted no cast permission")
	}
	if perm.Cost != "{0}" {
		t.Errorf("granted cost = %q, want \"{0}\" — the cast is free", perm.Cost)
	}

	// And it is really castable: no mana in pool, out of the graveyard.
	if err := g.CastSpell(me.ID, pitch, game.CastSpellParams{
		Strict: true, FromZone: "graveyard",
	}); err != nil {
		t.Fatalf("the granted free cast: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool after a free cast: %d tokens, want 0", len(me.ManaPool))
	}
}
