package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b783_cards_test.go — Portal, batch 783 (S82, #5630). Each test plays
// the real catalog entry through the engine.

const (
	oldB783SightOracle   = "20370c3b-231f-4d9d-8b6e-f1eb25fa4b5d"
	oldB783StareOracle   = "4135131d-4653-4767-ad1d-68c9bf393c3a"
	oldB783GazeOracle    = "9225016d-adfd-43c6-99cd-d41a7e0d35d6"
	oldB783OmenOracle    = "9b8c44a0-82ef-4beb-a762-c2a2a21380f3"
	oldB783BargainOracle = "dcf105c1-d37b-4e92-a420-9872a2187764"
	oldB783BeastOracle   = "5ae03181-a436-4bad-be3a-6c9f6c0ed4d6"
	oldB783HordeOracle   = "ad07cb55-cc4f-40be-b16a-bd3d3ca94249"
	oldB783WurmOracle    = "cdab50a8-1e02-4ba5-8c09-e2837e7652f7"
	oldB783PlantOracle   = "e822bf3d-3a29-4a02-9ae8-e2830ce70f15"
	oldB783KnightOracle  = "ed5429bb-233a-4528-bf7d-df5f6b192b1c"
)

func TestOldB783EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB783SightOracle: "Sorcerous Sight", oldB783StareOracle: "Baleful Stare", oldB783GazeOracle: "Withering Gaze",
		oldB783OmenOracle: "Omen", oldB783BargainOracle: "Cruel Bargain", oldB783BeastOracle: "Undying Beast",
		oldB783HordeOracle: "Pillaging Horde", oldB783WurmOracle: "Thundering Wurm",
		oldB783PlantOracle: "Plant Elemental", oldB783KnightOracle: "Mercenary Knight",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB783Opp(g *game.Game) (*game.Player, *game.Player) {
	me := g.Seats[g.Turn.ActiveSeat]
	return me, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB783PlayerTarget(p *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: p.ID}}
}

func oldB783InGraveyard(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Graveyard.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB783SorcerousSightLooksAtTheHandAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB783Opp(g)
	card := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Secret", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Sorcerous Sight", "Sorcery", oldB783SightOracle, oldB783PlayerTarget(me)); err == nil {
		t.Fatal("Sorcerous Sight targeted its own caster")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Sorcerous Sight", "Sorcery", oldB783SightOracle, oldB783PlayerTarget(opp))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, me.Hand.Size())
	}
}

func TestOldB783BalefulStareAndWitheringGazeDrawOnePerMatchingCard(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, land, color string
	}{
		{"Baleful Stare", oldB783StareOracle, "Basic Land — Mountain", "R"},
		{"Withering Gaze", oldB783GazeOracle, "Basic Land — Forest", "G"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := oldB783Opp(g)
			add := func(c game.Card) {
				c.InstanceID, c.Owner, c.Controller = uuid.New(), opp.ID, opp.ID
				opp.Hand.PushTop(c)
			}
			add(game.Card{Name: "Land", TypeLine: tc.land})
			add(game.Card{Name: "Colored", TypeLine: "Creature — Bear", Colors: []string{tc.color}, Power: 1, Toughness: 1})
			add(game.Card{Name: "Both", TypeLine: tc.land, Colors: []string{tc.color}})
			add(game.Card{Name: "Other", TypeLine: "Sorcery", Colors: []string{"U"}})
			add(game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
			effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
			hand := me.Hand.Size()
			effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, oldB783PlayerTarget(opp))
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != hand+3 {
				t.Errorf("hand %d -> %d, want three drawn (the land, the colored card, the card that is both once)", hand, me.Hand.Size())
			}
		})
	}
}

func TestOldB783OmenReordersMayShuffleAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB783Opp(g)
	a := effectstest.PushLibraryCard(me, game.Card{Name: "Bottom Of Three", TypeLine: "Sorcery"})
	b := effectstest.PushLibraryCard(me, game.Card{Name: "Middle Of Three", TypeLine: "Sorcery"})
	c := effectstest.PushLibraryCard(me, game.Card{Name: "Top Of Three", TypeLine: "Sorcery"})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Omen", "Sorcery", oldB783OmenOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	look := effectstest.LookAtTopChoiceFor(g, me.ID)
	if look == nil {
		t.Fatal("no look-at-the-top prompt")
	}
	if me.Hand.Size() != hand {
		t.Fatal("a card was drawn before the order and shuffle were settled")
	}
	if err := g.ResolveLookAtTop(look.ID, me.ID, []uuid.UUID{a, b, c}); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	effectstest.AnswerConfirmFor(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(a) {
		t.Errorf("hand %d -> %d, want the old bottom card drawn", hand, me.Hand.Size())
	}
}

func TestOldB783CruelBargainDrawsFourThenLosesHalfYourLifeRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB783Opp(g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	me.Life = 15
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Cruel Bargain", "Sorcery", oldB783BargainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d -> %d, want four drawn", hand, me.Hand.Size())
	}
	if me.Life != 7 {
		t.Errorf("life %d, want 7 (15 - 8)", me.Life)
	}
}

func TestOldB783UndyingBeastGoesOnTopOfItsOwnersLibraryWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB783Opp(g)
	id := effectstest.B12Push(g, me.ID, "Undying Beast", "Creature — Beast", oldB783BeastOracle, 3, 2)
	effectstest.SeedLibrary(me, "A", "B")
	oppLib := opp.Library.Size()
	effectstest.B25Destroy(g, id)
	effectstest.PassPriorityAroundTable(t, g)
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != id {
		t.Errorf("the top of my library is %v, want the Beast", top.Name)
	}
	if opp.Library.Size() != oppLib {
		t.Error("an opponent's library changed")
	}
}

func oldB783Enter(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string, p, tt int) uuid.UUID {
	t.Helper()
	id := effectstest.EnterCard(t, g, me.ID, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle, Power: p, Toughness: tt})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB783PillagingHordeDiscardsAtRandomToStay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB783Opp(g)
	id := oldB783Enter(t, g, me, "Pillaging Horde", "Creature — Human Barbarian", oldB783HordeOracle, 5, 5)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the Horde was sacrificed although a card was discarded")
	}
}

func TestOldB783PillagingHordeDiscardCountsOneCardFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB783Opp(g)
	before := len(me.Graveyard.Cards)
	hand := me.Hand.Size()
	oldB783Enter(t, g, me, "Pillaging Horde", "Creature — Human Barbarian", oldB783HordeOracle, 5, 5)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 || len(me.Graveyard.Cards) != before+1 {
		t.Errorf("hand %d -> %d, graveyard %d -> %d, want one random card discarded", hand, me.Hand.Size(), before, len(me.Graveyard.Cards))
	}
}

func TestOldB783PillagingHordeIsSacrificedIfYouDeclineOrHaveNoCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB783Opp(g)
	id := oldB783Enter(t, g, me, "Pillaging Horde", "Creature — Human Barbarian", oldB783HordeOracle, 5, 5)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) {
		t.Error("the Horde stayed although the controller chose to sacrifice it")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, _ := oldB783Opp(g2)
	for me2.Hand.Size() > 0 {
		_, _ = me2.Hand.Remove(me2.Hand.Cards[0].InstanceID)
	}
	id2 := oldB783Enter(t, g2, me2, "Pillaging Horde", "Creature — Human Barbarian", oldB783HordeOracle, 5, 5)
	if g2.Battlefield.Contains(id2) {
		t.Error("the Horde stayed with no card to discard")
	}
}

func TestOldB783ThunderingWurmAndMercenaryKnightNeedTheRightCardToStay(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tt                  int
		right, wrong           string
	}{
		{"Thundering Wurm", "Creature — Wurm", oldB783WurmOracle, 4, 4, "Basic Land — Forest", "Creature — Bear"},
		{"Mercenary Knight", "Creature — Human Mercenary Knight", oldB783KnightOracle, 4, 4, "Creature — Bear", "Basic Land — Forest"},
	} {
		t.Run(tc.name+" stays for the right card", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := oldB783Opp(g)
			right := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: right, Name: "Right", TypeLine: tc.right, Owner: me.ID, Controller: me.ID})
			id := oldB783Enter(t, g, me, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tt)
			effectstest.AnswerDiscard(t, g, me.ID, right)
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(id) || !oldB783InGraveyard(me, right) {
				t.Error("the creature was sacrificed although the right card was discarded")
			}
		})
		t.Run(tc.name+" is sacrificed with only the wrong card", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := oldB783Opp(g)
			wrong := uuid.New()
			for me.Hand.Size() > 0 {
				_, _ = me.Hand.Remove(me.Hand.Cards[0].InstanceID)
			}
			me.Hand.PushTop(game.Card{InstanceID: wrong, Name: "Wrong", TypeLine: tc.wrong, Owner: me.ID, Controller: me.ID})
			id := oldB783Enter(t, g, me, tc.name, tc.typeLine, tc.oracle, tc.p, tc.tt)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(id) {
				t.Error("the creature stayed with no right card to discard")
			}
			if oldB783InGraveyard(me, wrong) {
				t.Error("the wrong card was discarded")
			}
		})
	}
}

func TestOldB783PlantElementalSacrificesAForestOrItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB783Opp(g)
	forest := effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	island := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	id := oldB783Enter(t, g, me, "Plant Elemental", "Creature — Plant Elemental", oldB783PlantOracle, 3, 4)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || g.Battlefield.Contains(forest) || !g.Battlefield.Contains(island) {
		t.Error("the Forest was not sacrificed in place of the Elemental, or the Island was")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, _ := oldB783Opp(g2)
	effectstest.PushLandFor(g2, me2.ID, "Island", "Basic Land — Island")
	id2 := oldB783Enter(t, g2, me2, "Plant Elemental", "Creature — Plant Elemental", oldB783PlantOracle, 3, 4)
	if g2.Battlefield.Contains(id2) {
		t.Error("the Elemental stayed with no Forest to sacrifice")
	}
}
