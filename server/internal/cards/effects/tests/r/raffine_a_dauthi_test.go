package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDauthiVoidwalkerIsRegisteredAndWhole(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.DauthiOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// A card going to an opponent's graveyard is exiled with a void counter;
// one going to the Voidwalker's controller's own graveyard is not.
func TestDauthiVoidwalkerExilesOpponentsCardsWithAVoidCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dauthi Voidwalker", "Creature — Dauthi Rogue", effectstest.DauthiOracle, false)

	theirs := millOne(t, g, opp)
	if opp.Graveyard.Contains(theirs) {
		t.Fatal("the milled card reached the opponent's graveyard")
	}
	if got := effectstest.VoidCounters(g, theirs); got != 1 {
		t.Errorf("void counters on the exiled card = %d, want 1", got)
	}

	mine := millOne(t, g, me)
	if !me.Graveyard.Contains(mine) {
		t.Error("my own milled card was exiled; only an opponent's graveyard is replaced")
	}
}

// A dying opponent's creature never died: it is exiled with the counter.
func TestDauthiVoidwalkerExilesADyingOpponentCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Dauthi Voidwalker", "Creature — Dauthi Rogue", effectstest.DauthiOracle, false)
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	if opp.Graveyard.Contains(bear) {
		t.Fatal("the creature reached the graveyard")
	}
	if got := effectstest.VoidCounters(g, bear); got != 1 {
		t.Errorf("void counters = %d, want 1", got)
	}
}

// The ability: tap and sacrifice, choose a void-countered card an opponent
// owns, and play it free this turn. A card without a counter is not offered.
func TestDauthiVoidwalkerSacrificesToPlayAVoidCardForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	walker := effectstest.PushCatalogPermanent(g, me.ID, "Dauthi Voidwalker", "Creature — Dauthi Rogue", effectstest.DauthiOracle, false)
	effectstest.AdvanceToMainOf(t, g, seat)

	// Two exiled opponent cards: one void-countered creature, one plain.
	voided := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: voided, Name: "Voided Giant", TypeLine: "Creature — Giant",
		ManaCost: "{4}{R}{R}", Power: 6, Toughness: 6, Owner: opp.ID, Controller: opp.ID,
		Counters: map[string]int{"void": 1}})
	plain := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: plain, Name: "Plain Exiled", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	mineVoid := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: mineVoid, Name: "My Void Card", TypeLine: "Creature — Bear",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"void": 1}})

	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(walker) {
		t.Fatal("the Voidwalker was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)

	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no choice prompt")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != voided {
		t.Fatalf("candidates = %v, want only the opponent's void-countered card", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{voided}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if err := g.CastSpell(me.ID, voided, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast for free from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(voided) {
		t.Error("the free-cast creature did not enter the battlefield")
	}
	// The plain card was never castable.
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("a card without a void counter was castable")
	}
}
