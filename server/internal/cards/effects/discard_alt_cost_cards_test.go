package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// discard_alt_cost_cards_test.go — ADR 0135 §2 (#2412): alternative costs
// that discard cards, from hand, against the real catalog. Snag, Abolish,
// Flameshot and Outbreak discard one card of a land type; Foil discards
// an Island card and another card under a set rule; The Infamous
// Cruelclaw grants "cast it by discarding a card".

const (
	snagOracle       = "feb5e9a9-500a-4f20-881d-c97d052953ce"
	abolishOracle    = "d6d7faa8-ecc9-408b-9c46-a0f62c74f567"
	flameshotOracle  = "26640648-6400-4b80-a78b-f519ab9cd09e"
	outbreakOracle   = "05079cee-b190-42d7-b59f-130cb545cab7"
	foilOracle       = "7dde4eb6-e9d7-4259-abc2-3af738e0f00f"
	cruelclawOracle  = "e229d55c-9513-437d-bd67-015ac46aaa18"
	discardAltCostKy = "discard"
)

func TestDiscardAltCostCardsAreFull(t *testing.T) {
	for _, oracle := range []string{snagOracle, abolishOracle, flameshotOracle, outbreakOracle, foilOracle, cruelclawOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
	// The four one-card discards name their land type, and charge no mana.
	for _, c := range []struct{ oracle, sub string }{
		{snagOracle, "Forest"}, {abolishOracle, "Plains"}, {flameshotOracle, "Mountain"}, {outbreakOracle, "Swamp"},
	} {
		offers := game.AlternativeCostsFor(c.oracle)
		if len(offers) != 1 || offers[0].Key != discardAltCostKy || offers[0].ManaCost != "" ||
			offers[0].DiscardFromHand == nil || offers[0].CardPaymentCount() != 1 {
			t.Errorf("%s offers %+v, want one free discard of a %s card", c.oracle, offers, c.sub)
		}
	}
}

// discardTable is a main phase on seat 0's turn with an empty hand.
func discardTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	return g, me, g.Seats[1]
}

func handCardOf(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	})
	return id
}

func castDiscarding(g *game.Game, p *game.Player, card uuid.UUID, targets []game.TargetRef, discard ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		AlternativeCost: discardAltCostKy, AltCostIDs: discard, Strict: true, Targets: targets,
	})
}

// Snag from hand for a Forest card and no mana, under the strict gate.
// A non-Forest is refused with nothing paid; a nonbasic Forest pays it
// (the ruling); the Forest is discarded with the spell on the stack.
func TestSnagIsCastByDiscardingAForestCard(t *testing.T) {
	g, me, _ := discardTable(t)
	snag := handCardOf(g, me, "Snag", "Instant", "{3}{G}", snagOracle)
	mountain := handCardOf(g, me, "Mountain", "Basic Land — Mountain", "", "")
	trop := handCardOf(g, me, "Tropical Island", "Land — Forest Island", "", "")

	if err := castDiscarding(g, me, snag, nil, mountain); err == nil {
		t.Fatal("a Mountain paid \"discard a Forest card\"")
	}
	if !me.Hand.Contains(mountain) || !me.Hand.Contains(snag) {
		t.Fatal("the refused cast paid something")
	}
	if err := castDiscarding(g, me, snag, nil); err == nil {
		t.Fatal("Snag was cast for its alternative cost with nothing discarded")
	}
	if err := castDiscarding(g, me, snag, nil, trop); err != nil {
		t.Fatalf("Snag for Tropical Island: %v", err)
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack = %d, want Snag", g.Stack.Size())
	}
	if me.Hand.Contains(trop) || !me.Graveyard.Contains(trop) {
		t.Fatal("the Forest card was not discarded while Snag was on the stack")
	}
	if !me.Hand.Contains(mountain) {
		t.Error("the Mountain was discarded too")
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(snag) {
		t.Error("Snag did not resolve to the graveyard")
	}
}

// Abolish for a Plains card destroys its target; Flameshot divides 3
// among creatures for a Mountain card; Outbreak shrinks a chosen type
// for a Swamp card.
func TestAbolishFlameshotAndOutbreakForTheirLandCards(t *testing.T) {
	t.Run("Abolish", func(t *testing.T) {
		g, me, opp := discardTable(t)
		rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
		abolish := handCardOf(g, me, "Abolish", "Instant", "{1}{W}{W}", abolishOracle)
		plains := handCardOf(g, me, "Plains", "Basic Land — Plains", "", "")
		if err := castDiscarding(g, me, abolish, []game.TargetRef{{Kind: game.TargetCard, ID: rock}}, plains); err != nil {
			t.Fatalf("Abolish for a Plains card: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("Abolish did not destroy the artifact")
		}
		if !me.Graveyard.Contains(plains) {
			t.Error("the Plains card was not discarded")
		}
	})
	t.Run("Flameshot", func(t *testing.T) {
		g, me, opp := discardTable(t)
		a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
		b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
		shot := handCardOf(g, me, "Flameshot", "Sorcery", "{3}{R}", flameshotOracle)
		mountain := handCardOf(g, me, "Mountain", "Basic Land — Mountain", "", "")
		err := g.CastSpell(me.ID, shot, game.CastSpellParams{
			AlternativeCost: discardAltCostKy, AltCostIDs: []uuid.UUID{mountain}, Strict: true,
			Targets:      cardRefs(a, b),
			Distribution: map[uuid.UUID]int{a: 2, b: 1},
		})
		if err != nil {
			t.Fatalf("Flameshot for a Mountain card: %v", err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(a) {
			t.Error("the creature dealt 2 survived")
		}
		if !g.Battlefield.Contains(b) {
			t.Error("the creature dealt 1 died")
		}
	})
	t.Run("Outbreak", func(t *testing.T) {
		g, me, opp := discardTable(t)
		elf := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
		outbreak := handCardOf(g, me, "Outbreak", "Sorcery", "{3}{B}", outbreakOracle)
		swamp := handCardOf(g, me, "Swamp", "Basic Land — Swamp", "", "")
		if err := castDiscarding(g, me, outbreak, nil, swamp); err != nil {
			t.Fatalf("Outbreak for a Swamp card: %v", err)
		}
		passPriorityAroundTable(t, g)
		answerCreatureType(t, g, me.ID, "Elf")
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(elf) {
			t.Error("the 1/1 Elf survived -1/-1")
		}
		if !g.Battlefield.Contains(bear) {
			t.Error("the Bear was shrunk too")
		}
	})
}

// Foil: "an Island card and another card". Two non-Islands are refused,
// one Island alone is refused, an Island and anything else pays it, and
// two Islands pay it.
func TestFoilDiscardsAnIslandCardAndAnotherCard(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *game.Player, uuid.UUID, []game.TargetRef) {
		g, me, _ := discardTable(t)
		bolt := handCardOf(g, me, "Bolt Stand-in", "Instant", "{R}", "")
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err != nil {
			t.Fatalf("cast the spell to counter: %v", err)
		}
		foil := handCardOf(g, me, "Foil", "Instant", "{2}{U}{U}", foilOracle)
		return g, me, foil, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}
	}

	t.Run("two non-Islands", func(t *testing.T) {
		g, me, foil, at := setup(t)
		a := handCardOf(g, me, "Mountain", "Basic Land — Mountain", "", "")
		b := handCardOf(g, me, "Spare", "Sorcery", "{2}", "")
		if offerWithKey(grantedCardOffers(g, me.ID, foil, game.ZoneHand), discardAltCostKy) != nil {
			t.Error("Foil's offer is listed with no Island card in hand")
		}
		if err := castDiscarding(g, me, foil, at, a, b); err == nil {
			t.Fatal("two non-Island cards paid Foil")
		}
		if !me.Hand.Contains(a) || !me.Hand.Contains(b) || !me.Hand.Contains(foil) {
			t.Fatal("the refused cast paid something")
		}
	})
	t.Run("an Island alone", func(t *testing.T) {
		g, me, foil, at := setup(t)
		isl := handCardOf(g, me, "Island", "Basic Land — Island", "", "")
		if offerWithKey(grantedCardOffers(g, me.ID, foil, game.ZoneHand), discardAltCostKy) != nil {
			t.Error("Foil's offer is listed with one card to discard")
		}
		if err := castDiscarding(g, me, foil, at, isl); err == nil {
			t.Fatal("one Island paid Foil")
		}
		if err := castDiscarding(g, me, foil, at, isl, isl); err == nil {
			t.Fatal("one Island named twice paid Foil")
		}
	})
	t.Run("an Island and another card", func(t *testing.T) {
		g, me, foil, at := setup(t)
		isl := handCardOf(g, me, "Island", "Basic Land — Island", "", "")
		spare := handCardOf(g, me, "Spare", "Sorcery", "{2}", "")
		if offerWithKey(grantedCardOffers(g, me.ID, foil, game.ZoneHand), discardAltCostKy) == nil {
			t.Fatal("Foil's offer is not listed with an Island card and another card in hand")
		}
		// The other card named first: the matching, not the order, decides.
		if err := castDiscarding(g, me, foil, at, spare, isl); err != nil {
			t.Fatalf("Foil for an Island card and another card: %v", err)
		}
		if !me.Graveyard.Contains(isl) || !me.Graveyard.Contains(spare) {
			t.Fatal("both cards were not discarded")
		}
		passPriorityAroundTable(t, g)
		if g.Stack.Size() != 0 {
			t.Errorf("stack = %d after Foil resolved", g.Stack.Size())
		}
	})
	t.Run("two Islands", func(t *testing.T) {
		g, me, foil, at := setup(t)
		a := handCardOf(g, me, "Island A", "Basic Land — Island", "", "")
		b := handCardOf(g, me, "Island B", "Basic Land — Island", "", "")
		if err := castDiscarding(g, me, foil, at, a, b); err != nil {
			t.Fatalf("Foil for two Islands: %v", err)
		}
	})
}

// The wire and the bot: Foil's pay_options carry each_of with the two
// groups, `discards` is set, and every enumerated Foil payment is one the
// engine accepts (an Island and the cheapest other card).
func TestFoilOfferShipsItsGroupsAndTheEnumeratorPaysASet(t *testing.T) {
	g, me, _ := discardTable(t)
	bolt := handCardOf(g, me, "Bolt Stand-in", "Instant", "{R}", "")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err != nil {
		t.Fatalf("cast the spell to counter: %v", err)
	}
	foil := handCardOf(g, me, "Foil", "Instant", "{2}{U}{U}", foilOracle)
	handCardOf(g, me, "Mountain A", "Basic Land — Mountain", "", "")
	handCardOf(g, me, "Mountain B", "Basic Land — Mountain", "", "")
	isl := handCardOf(g, me, "Island", "Basic Land — Island", "", "")

	var offer *protocol.AlternativeCostView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID == foil.String() {
				for j := range s.Hand.Cards[i].AlternativeCosts {
					if s.Hand.Cards[i].AlternativeCosts[j].Key == discardAltCostKy {
						offer = &s.Hand.Cards[i].AlternativeCosts[j]
					}
				}
			}
		}
	}
	if offer == nil || offer.PayOptions == nil {
		t.Fatalf("Foil's discard offer is not on the wire: %+v", offer)
	}
	if !offer.Discards {
		t.Error("the offer does not say it discards")
	}
	po := offer.PayOptions
	if po.Min != 2 || po.Max != 2 || len(po.Cards) != 3 {
		t.Errorf("pay_options = min %d max %d cards %v, want 2/2 over the three cards", po.Min, po.Max, po.Cards)
	}
	if len(po.EachOf) != 2 || po.EachOf[0].Label != "an Island card" || len(po.EachOf[0].Cards) != 1 ||
		po.EachOf[0].Cards[0] != isl.String() || len(po.EachOf[1].Cards) != 3 {
		t.Errorf("each_of = %+v, want the Island alone, then every card", po.EachOf)
	}

	type move struct {
		AlternativeCost string   `json:"alternative_cost"`
		AltCostIDs      []string `json:"alt_cost_ids"`
	}
	found := 0
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != foil {
			continue
		}
		var p move
		if err := json.Unmarshal(m.Params, &p); err != nil || p.AlternativeCost != discardAltCostKy {
			continue
		}
		found++
		if len(p.AltCostIDs) != 2 {
			t.Fatalf("enumerated Foil payment %v is not two cards", p.AltCostIDs)
		}
		hasIsland := p.AltCostIDs[0] == isl.String() || p.AltCostIDs[1] == isl.String()
		if !hasIsland {
			t.Errorf("enumerated Foil payment %v has no Island card", p.AltCostIDs)
		}
	}
	if found == 0 {
		t.Fatal("the enumerator offers no Foil cast for its discard")
	}
}

// The Infamous Cruelclaw: the trigger exiles down to a nonland card,
// which may be cast by discarding a card and no mana, and not for free.
func TestInfamousCruelclawCastsTheExiledCardByDiscarding(t *testing.T) {
	g, me, _ := discardTable(t)
	hit := uuid.New()
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	})
	claw := pushCatalogPermanent(g, me.ID, "The Infamous Cruelclaw", "Legendary Creature — Weasel Mercenary", cruelclawOracle, false)
	var err error
	g.WithWriteLock(func() {
		err = cruelclawHits(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: claw})
	})
	if err != nil {
		t.Fatalf("Cruelclaw's trigger: %v", err)
	}
	if !g.Exile.Contains(hit) || !g.Exile.Contains(land) {
		t.Fatal("the land and the two-drop are not both in exile")
	}
	if o := offerWithKey(grantedCardOffers(g, me.ID, hit, game.ZoneExile), cruelclawAltCostKey); o != nil {
		t.Fatalf("offered with an empty hand: %+v", o)
	}
	spare := handCardOf(g, me, "Spare", "Sorcery", "{2}", "")
	o := offerWithKey(grantedCardOffers(g, me.ID, hit, game.ZoneExile), cruelclawAltCostKey)
	if o == nil || o.ManaCost != "" || o.DiscardFromHand == nil {
		t.Fatalf("offer on the exiled two-drop = %+v", o)
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: cruelclawAltCostKey}); err == nil {
		t.Fatal("the exiled card was cast without discarding")
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true,
		AlternativeCost: cruelclawAltCostKey, AltCostIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("cast by discarding a card: %v", err)
	}
	if !me.Graveyard.Contains(spare) {
		t.Error("the card was not discarded")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hit) {
		t.Error("the two-drop did not resolve onto the battlefield")
	}
}
