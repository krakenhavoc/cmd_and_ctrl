package testsm

import (
	"testing"

	"github.com/google/uuid"

	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMalcolmHasFlashAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushMalcolm(t, g, me.ID, 0)
	got := effectstest.EffectiveAbilities(t, g, id)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushMalcolm(t, g, me.ID, 0)
	effectstest.AdvanceToMain(t, g)

	hand, graves := me.Hand.Size(), me.Graveyard.Size()
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.FindTestCard(g, id); c == nil || c.Counters[cardslu.MalcolmChorusCounter] != 1 {
		t.Fatalf("chorus counters after one hit = %v, want 1", c)
	}
	// The draw is above the discard, so the prompt is over a hand that
	// already holds the new card.
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand while the discard prompt is open = %d, want %d — draw, THEN discard", got, hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Graveyard.Size(); got != graves+1 {
		t.Errorf("graveyard %d → %d, want one card discarded", graves, got)
	}
	if offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMayCast); offer != nil {
		t.Error("a free cast was offered at one chorus counter; the card prints four or more")
	}
}

// TestMalcolmAtFourChorusCountersCastsTheDiscardedCardForFree is the
// last sentence, end to end: the fourth counter arrives, the offer is
// made, and taking it lets the discarded card be cast out of the
// graveyard for nothing.
func TestMalcolmAtFourChorusCountersCastsTheDiscardedCardForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushMalcolm(t, g, me.ID, 3)
	effectstest.AdvanceToMain(t, g)

	// A real spell to pitch, so the free cast has something to prove.
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: pitch, Name: "Pitched Spell", TypeLine: "Instant",
		ManaCost: "{4}{U}{U}", Owner: me.ID, Controller: me.ID,
	})

	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.FindTestCard(g, id); c == nil || c.Counters[cardslu.MalcolmChorusCounter] != 4 {
		t.Fatalf("chorus counters = %v, want 4", c)
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(pitch) {
		t.Fatal("the pitched card is not in the graveyard")
	}
	offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMayCast)
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
