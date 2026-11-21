package testscr

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestDiscardAltCostCardsAreFull(t *testing.T) {
	for _, oracle := range []string{effectstest.SnagOracle, effectstest.AbolishOracle, effectstest.FlameshotOracle, effectstest.OutbreakOracle, effectstest.FoilOracle, effectstest.CruelclawOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
	}
	// The four one-card discards name their land type, and charge no mana.
	for _, c := range []struct{ oracle, sub string }{
		{effectstest.SnagOracle, "Forest"}, {effectstest.AbolishOracle, "Plains"}, {effectstest.FlameshotOracle, "Mountain"}, {effectstest.OutbreakOracle, "Swamp"},
	} {
		offers := game.AlternativeCostsFor(c.oracle)
		if len(offers) != 1 || offers[0].Key != effectstest.DiscardAltCostKy || offers[0].ManaCost != "" ||
			offers[0].DiscardFromHand == nil || offers[0].CardPaymentCount() != 1 {
			t.Errorf("%s offers %+v, want one free discard of a %s card", c.oracle, offers, c.sub)
		}
	}
}

func castDiscarding(g *game.Game, p *game.Player, card uuid.UUID, targets []game.TargetRef, discard ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		AlternativeCost: effectstest.DiscardAltCostKy, AltCostIDs: discard, Strict: true, Targets: targets,
	})
}

// Snag from hand for a Forest card and no mana, under the strict gate.
// A non-Forest is refused with nothing paid; a nonbasic Forest pays it
// (the ruling); the Forest is discarded with the spell on the stack.
func TestSnagIsCastByDiscardingAForestCard(t *testing.T) {
	g, me, _ := effectstest.DiscardTable(t)
	snag := effectstest.HandCardOf(g, me, "Snag", "Instant", "{3}{G}", effectstest.SnagOracle)
	mountain := effectstest.HandCardOf(g, me, "Mountain", "Basic Land — Mountain", "", "")
	trop := effectstest.HandCardOf(g, me, "Tropical Island", "Land — Forest Island", "", "")

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
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(snag) {
		t.Error("Snag did not resolve to the graveyard")
	}
}

// Abolish for a Plains card destroys its target; Flameshot divides 3
// among creatures for a Mountain card; Outbreak shrinks a chosen type
// for a Swamp card.
func TestAbolishFlameshotAndOutbreakForTheirLandCards(t *testing.T) {
	t.Run("Abolish", func(t *testing.T) {
		g, me, opp := effectstest.DiscardTable(t)
		rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
		abolish := effectstest.HandCardOf(g, me, "Abolish", "Instant", "{1}{W}{W}", effectstest.AbolishOracle)
		plains := effectstest.HandCardOf(g, me, "Plains", "Basic Land — Plains", "", "")
		if err := castDiscarding(g, me, abolish, []game.TargetRef{{Kind: game.TargetCard, ID: rock}}, plains); err != nil {
			t.Fatalf("Abolish for a Plains card: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(rock) {
			t.Error("Abolish did not destroy the artifact")
		}
		if !me.Graveyard.Contains(plains) {
			t.Error("the Plains card was not discarded")
		}
	})
	t.Run("Flameshot", func(t *testing.T) {
		g, me, opp := effectstest.DiscardTable(t)
		a := effectstest.B12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
		b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
		shot := effectstest.HandCardOf(g, me, "Flameshot", "Sorcery", "{3}{R}", effectstest.FlameshotOracle)
		mountain := effectstest.HandCardOf(g, me, "Mountain", "Basic Land — Mountain", "", "")
		err := g.CastSpell(me.ID, shot, game.CastSpellParams{
			AlternativeCost: effectstest.DiscardAltCostKy, AltCostIDs: []uuid.UUID{mountain}, Strict: true,
			Targets:      effectstest.CardRefs(a, b),
			Distribution: map[uuid.UUID]int{a: 2, b: 1},
		})
		if err != nil {
			t.Fatalf("Flameshot for a Mountain card: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(a) {
			t.Error("the creature dealt 2 survived")
		}
		if !g.Battlefield.Contains(b) {
			t.Error("the creature dealt 1 died")
		}
	})
	t.Run("Outbreak", func(t *testing.T) {
		g, me, opp := effectstest.DiscardTable(t)
		elf := effectstest.B12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
		outbreak := effectstest.HandCardOf(g, me, "Outbreak", "Sorcery", "{3}{B}", effectstest.OutbreakOracle)
		swamp := effectstest.HandCardOf(g, me, "Swamp", "Basic Land — Swamp", "", "")
		if err := castDiscarding(g, me, outbreak, nil, swamp); err != nil {
			t.Fatalf("Outbreak for a Swamp card: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerCreatureType(t, g, me.ID, "Elf")
		effectstest.PassPriorityAroundTable(t, g)
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
		g, me, _ := effectstest.DiscardTable(t)
		bolt := effectstest.HandCardOf(g, me, "Bolt Stand-in", "Instant", "{R}", "")
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err != nil {
			t.Fatalf("cast the spell to counter: %v", err)
		}
		foil := effectstest.HandCardOf(g, me, "Foil", "Instant", "{2}{U}{U}", effectstest.FoilOracle)
		return g, me, foil, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}
	}

	t.Run("two non-Islands", func(t *testing.T) {
		g, me, foil, at := setup(t)
		a := effectstest.HandCardOf(g, me, "Mountain", "Basic Land — Mountain", "", "")
		b := effectstest.HandCardOf(g, me, "Spare", "Sorcery", "{2}", "")
		if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, foil, game.ZoneHand), effectstest.DiscardAltCostKy) != nil {
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
		isl := effectstest.HandCardOf(g, me, "Island", "Basic Land — Island", "", "")
		if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, foil, game.ZoneHand), effectstest.DiscardAltCostKy) != nil {
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
		isl := effectstest.HandCardOf(g, me, "Island", "Basic Land — Island", "", "")
		spare := effectstest.HandCardOf(g, me, "Spare", "Sorcery", "{2}", "")
		if effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, foil, game.ZoneHand), effectstest.DiscardAltCostKy) == nil {
			t.Fatal("Foil's offer is not listed with an Island card and another card in hand")
		}
		// The other card named first: the matching, not the order, decides.
		if err := castDiscarding(g, me, foil, at, spare, isl); err != nil {
			t.Fatalf("Foil for an Island card and another card: %v", err)
		}
		if !me.Graveyard.Contains(isl) || !me.Graveyard.Contains(spare) {
			t.Fatal("both cards were not discarded")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Stack.Size() != 0 {
			t.Errorf("stack = %d after Foil resolved", g.Stack.Size())
		}
	})
	t.Run("two Islands", func(t *testing.T) {
		g, me, foil, at := setup(t)
		a := effectstest.HandCardOf(g, me, "Island A", "Basic Land — Island", "", "")
		b := effectstest.HandCardOf(g, me, "Island B", "Basic Land — Island", "", "")
		if err := castDiscarding(g, me, foil, at, a, b); err != nil {
			t.Fatalf("Foil for two Islands: %v", err)
		}
	})
}

// The wire and the bot: Foil's pay_options carry each_of with the two
// groups, `discards` is set, and every enumerated Foil payment is one the
// engine accepts (an Island and the cheapest other card).
func TestFoilOfferShipsItsGroupsAndTheEnumeratorPaysASet(t *testing.T) {
	g, me, _ := effectstest.DiscardTable(t)
	bolt := effectstest.HandCardOf(g, me, "Bolt Stand-in", "Instant", "{R}", "")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err != nil {
		t.Fatalf("cast the spell to counter: %v", err)
	}
	foil := effectstest.HandCardOf(g, me, "Foil", "Instant", "{2}{U}{U}", effectstest.FoilOracle)
	effectstest.HandCardOf(g, me, "Mountain A", "Basic Land — Mountain", "", "")
	effectstest.HandCardOf(g, me, "Mountain B", "Basic Land — Mountain", "", "")
	isl := effectstest.HandCardOf(g, me, "Island", "Basic Land — Island", "", "")

	var offer *protocol.AlternativeCostView
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if s.Hand.Cards[i].InstanceID == foil.String() {
				for j := range s.Hand.Cards[i].AlternativeCosts {
					if s.Hand.Cards[i].AlternativeCosts[j].Key == effectstest.DiscardAltCostKy {
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
		if err := json.Unmarshal(m.Params, &p); err != nil || p.AlternativeCost != effectstest.DiscardAltCostKy {
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
	g, me, _ := effectstest.DiscardTable(t)
	hit := uuid.New()
	land := uuid.New()
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	})
	claw := effectstest.PushCatalogPermanent(g, me.ID, "The Infamous Cruelclaw", "Legendary Creature — Weasel Mercenary", effectstest.CruelclawOracle, false)
	var err error
	g.WithWriteLock(func() {
		err = cardsth.CruelclawHits(g, &game.StackItem{ID: uuid.New(), Controller: me.ID, SourceCardID: claw})
	})
	if err != nil {
		t.Fatalf("Cruelclaw's trigger: %v", err)
	}
	if !g.Exile.Contains(hit) || !g.Exile.Contains(land) {
		t.Fatal("the land and the two-drop are not both in exile")
	}
	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, hit, game.ZoneExile), cardsth.CruelclawAltCostKey); o != nil {
		t.Fatalf("offered with an empty hand: %+v", o)
	}
	spare := effectstest.HandCardOf(g, me, "Spare", "Sorcery", "{2}", "")
	o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, hit, game.ZoneExile), cardsth.CruelclawAltCostKey)
	if o == nil || o.ManaCost != "" || o.DiscardFromHand == nil {
		t.Fatalf("offer on the exiled two-drop = %+v", o)
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true, AlternativeCost: cardsth.CruelclawAltCostKey}); err == nil {
		t.Fatal("the exiled card was cast without discarding")
	}
	if err := g.CastSpell(me.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true,
		AlternativeCost: cardsth.CruelclawAltCostKey, AltCostIDs: []uuid.UUID{spare}}); err != nil {
		t.Fatalf("cast by discarding a card: %v", err)
	}
	if !me.Graveyard.Contains(spare) {
		t.Error("the card was not discarded")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hit) {
		t.Error("the two-drop did not resolve onto the battlefield")
	}
}
