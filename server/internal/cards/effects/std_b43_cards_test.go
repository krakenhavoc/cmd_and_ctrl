package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b43_cards_test.go — Standard set leftovers, batches 43-46 (#4452).
// Each test plays the real catalog entry and asserts what its text says.

const (
	stdSentryOracle  = "6b869c33-a77c-4eb7-9df9-9815d70f7444"
	stdGossipOracle  = "19dc29a7-8ccc-4e54-829b-86cc102b082d"
	stdHellionOracle = "e121f9cb-ed25-4be1-9822-b68cce2f2623"
)

func TestStdB43HellionDisguiseAndReach(t *testing.T) {
	spec, ok := Lookup(stdHellionOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Riftburst Hellion: %+v", spec)
	}
	ac := game.AlternativeCostByKey(stdHellionOracle, "disguise")
	if ac == nil || ac.FaceDown == nil || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != "{4}{R/G}{R/G}" {
		t.Fatalf("disguise offer: %+v", ac)
	}
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "reach" {
		t.Errorf("keywords %v, want reach", spec.PrintedKeywords)
	}
}

func TestStdB43SumalaSentryCountsBothWhenAPermanentTurnsUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sentry := b12Push(g, me.ID, "Sumala Sentry", "Creature — Elf Archer", stdSentryOracle, 3, 3)
	perf := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	// Entering face down is not turning face up.
	if c := tmtB1Card(t, g, sentry); c.Counters["+1/+1"] != 0 {
		t.Fatalf("a permanent entering face down gave the Sentry a counter")
	}
	mkmMPTurnFaceUp(t, g, me, perf, "R")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	if c := tmtB1Card(t, g, sentry); c.Counters["+1/+1"] != 1 {
		t.Errorf("Sentry counters %d, want 1", c.Counters["+1/+1"])
	}
	if c := tmtB1Card(t, g, perf); c.Counters["+1/+1"] != 1 {
		t.Errorf("turned-up permanent counters %d, want 1", c.Counters["+1/+1"])
	}
}

func stdGossipMana(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	gossip := pushCatalogPermanent(g, me.ID, "Tin Street Gossip", "Creature — Lizard Advisor", stdGossipOracle, false)
	b28TapForMana(t, g, me.ID, gossip, "")
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("Gossip made %d mana, want {R}{G}", got)
	}
	return g, me
}

func TestStdB43TinStreetGossipManaPaysAFaceDownCast(t *testing.T) {
	g, me := stdGossipMana(t)
	b06AddMana(me, "C")
	id := handCardFull(me, "Pyrotechnic Performer", mkmMPPerformerTypeLine, "{1}{R}", mkmMPPerformerOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "disguise", Strict: true}); err != nil {
		t.Fatalf("a face-down cast paid with Gossip mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want 0", got)
	}
}

func TestStdB43TinStreetGossipManaCannotPayAFaceUpCast(t *testing.T) {
	g, me := stdGossipMana(t)
	id := handCardFull(me, "Pyrotechnic Performer", mkmMPPerformerTypeLine, "{1}{R}", mkmMPPerformerOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a face-up cast was paid with Gossip's restricted mana")
	}
}

func TestStdB43TinStreetGossipManaTurnsACreatureFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gossip := pushCatalogPermanent(g, me.ID, "Tin Street Gossip", "Creature — Lizard Advisor", stdGossipOracle, false)
	perf := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	b28TapForMana(t, g, me.ID, gossip, "")
	if err := g.PerformSpecialAction(me.ID, perf, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up with Gossip mana: %v", err)
	}
	if findCardAnywhere(t, g, perf).FaceDown {
		t.Fatal("still face down")
	}
}

const stdTolsimirOracle = "0e9c7380-7584-43e6-8c31-5d1bb1685fe0"

func TestStdB43TolsimirCreatesVojaWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Tolsimir, Midnight's Light", "Legendary Creature — Elf Scout", stdTolsimirOracle, 3, 3)
	passPriorityAroundTable(t, g)
	var voja game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Voja Fenstalker" {
				voja = c
			}
		}
	})
	if voja.InstanceID == [16]byte{} || voja.Power != 5 || voja.Toughness != 5 || !voja.HasSubtype("Wolf") ||
		!hasKeywordOnBattlefield(t, g, voja.InstanceID, "trample") || len(voja.Colors) != 2 {
		t.Fatalf("Voja Fenstalker is %+v, want a legendary 5/5 green and white Wolf with trample", voja)
	}
}

func stdTolsimirBoard(g *game.Game) (me, opp *game.Player, tolsimir, wolf, victim uuid.UUID) {
	me, opp = seats1684(g)
	tolsimir = typedCreature(g, me.ID, "Tolsimir, Midnight's Light", "Legendary Creature — Elf Scout", stdTolsimirOracle, 3, 3, "haste")
	wolf = typedCreature(g, me.ID, "Test Wolf", "Creature — Wolf", "", 2, 2, "haste")
	victim = reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	return
}

func TestStdB43TolsimirForcesABlockOnAWolfWhenHeAttacksToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, tolsimir, wolf, victim := stdTolsimirBoard(g)
	attackAndTrigger(t, g, tolsimir, wolf)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	recs := blocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != wolf {
		t.Fatalf("records = %+v, want one naming the Wolf", recs)
	}
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, tolsimir, victim))
	if err := reqBlock(t, g, wolf, victim); err != nil {
		t.Fatalf("the forced block: %v", err)
	}
	_ = opp
}

func TestStdB43TolsimirDoesNothingWhenHeStaysHome(t *testing.T) {
	g := newCatalogGame(t)
	_, _, _, wolf, _ := stdTolsimirBoard(g)
	attackAndTrigger(t, g, wolf)
	passPriorityAroundTable(t, g)
	if recs := blocksAttackerRecords(g); len(recs) != 0 {
		t.Fatalf("records = %+v, want none: Tolsimir did not attack", recs)
	}
}
