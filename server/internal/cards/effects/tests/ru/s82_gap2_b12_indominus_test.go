package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b12_indominus_test.go — S82 GAP2 batch 12: Indominus Rex, Alpha,
// through the real cast, the as-enters discard and the enters trigger.

const gap2B12Indominus = "723a66c3-2c2c-411c-bba5-922e8601981b"

func TestGap2B12IndominusIsFull(t *testing.T) {
	spec, ok := effects.Lookup(gap2B12Indominus)
	if !ok {
		t.Fatal("Indominus Rex, Alpha is not registered")
	}
	if spec.Name != "Indominus Rex, Alpha" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func gap2B12EntryDiscard(g *game.Game, me uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryDiscardFromHand && c.Chooser == me {
			return c
		}
	}
	return nil
}

// gap2B12CastRex casts a 6/6 Indominus Rex from the active player's hand.
func gap2B12CastRex(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Indominus Rex, Alpha", TypeLine: "Legendary Creature — Dinosaur Mutant",
		OracleID: gap2B12Indominus, Owner: me.ID, Controller: me.ID, Power: 6, Toughness: 6})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("casting the Rex: %v", err)
	}
	return id
}

func gap2B12InHand(p *game.Player, name, typeLine string, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID, Keywords: keywords})
	return id
}

func TestGap2B12IndominusDiscardsForKeywordCountersThenDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dragon := gap2B12InHand(me, "Test Dragon", "Creature — Dragon", "flying", "deathtouch")
	bird := gap2B12InHand(me, "Test Bird", "Creature — Bird", "flying")
	bear := gap2B12InHand(me, "Test Bear", "Creature — Bear")
	spell := gap2B12InHand(me, "Test Spell", "Instant")

	rex := gap2B12CastRex(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := gap2B12EntryDiscard(g, me.ID)
	if c == nil {
		t.Fatalf("no as-enters discard prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, []uuid.UUID{spell}); err == nil {
		t.Error("a noncreature card was accepted")
	}
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, []uuid.UUID{dragon, bird, bear}); err != nil {
		t.Fatalf("discarding three creature cards: %v", err)
	}
	for _, id := range []uuid.UUID{dragon, bird, bear} {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
			t.Errorf("a discarded card is in %+v", z)
		}
	}
	var card game.Card
	for _, b := range g.Battlefield.Cards {
		if b.Name == "Indominus Rex, Alpha" {
			card = b
		}
	}
	if card.InstanceID == uuid.Nil {
		t.Fatalf("the Rex did not enter (cast id %s, zone %+v, pending %+v, stack %d)", rex, g.FindCardZoneForEffect(rex), g.PendingChoices, len(g.Stack.Cards))
	}
	if card.Counters[game.CounterFlying] != 1 || card.Counters[game.CounterDeathtouch] != 1 || len(card.Counters) != 2 {
		t.Errorf("counters %v, want one flying and one deathtouch", card.Counters)
	}
	if !game.HasKeyword(&card, "flying") || !game.HasKeyword(&card, "deathtouch") {
		t.Error("the keyword counters did not give the keywords")
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d after the enters trigger, want %d (one card per counter)", got, hand+2)
	}
}

func TestGap2B12IndominusDiscardingNothingDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gap2B12InHand(me, "Test Bird", "Creature — Bird", "flying")
	gap2B12CastRex(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	c := gap2B12EntryDiscard(g, me.ID)
	if c == nil {
		t.Fatalf("no as-enters discard prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, nil); err != nil {
		t.Fatalf("discarding nothing: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d: no counters, no draw", got, hand)
	}
}
