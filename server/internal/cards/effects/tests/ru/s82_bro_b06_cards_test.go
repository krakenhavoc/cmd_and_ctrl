package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b06_cards_test.go — S82 BRO batch 6 (#5482): the mill-three-take-one
// cycle and creature-cards-in-graveyard counts. Every test plays the real
// catalog card through the engine.

const (
	broB06Chaplain   = "719aeb50-c8cf-4cd8-b74c-d2cef408b52e"
	broB06Prowler    = "8d1a367a-542e-4b94-b466-c922c5fb717b"
	broB06Archaeo    = "bf037457-82e3-45fd-bc17-5bc25e66c15e"
	broB06Gigamole   = "68f4e3e5-2f0d-4d2d-84f6-d1e824a0dcf0"
	broB06Scrapsmith = "d1559911-6621-4fd6-8b3e-16e82baf849e"
	broB06Harvest    = "d44627c0-1617-4706-bbeb-0169745364c7"
	broB06Razormaw   = "8551e1df-47eb-419e-8dd2-e96b9a450912"
	broB06Courser    = "5d13aa6a-5a04-45e0-bb33-daa707ffaf9f"
	broB06Skullflay  = "0b2a59a3-cf87-48fe-99e2-4c2063119b55"
	broB06Pallbearer = "18da85d4-2a51-45c0-a18c-bce9bc39fa64"
	broB06Butcher    = "2978ba77-c0ce-4411-99c7-ef05dd602c8d"
	broB06Ritualist  = "b52148c3-fe97-4cc0-a814-a5a0f5d426f6"
	broB06Remorse    = "bb4507c1-1e5d-48b5-8fc0-1d26bd1c9e72"
)

func TestBroB06EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		broB06Chaplain: "Airlift Chaplain", broB06Prowler: "Blanchwood Prowler", broB06Archaeo: "Fallaji Archaeologist",
		broB06Gigamole: "Ravenous Gigamole", broB06Scrapsmith: "Tomakul Scrapsmith", broB06Harvest: "Wasteful Harvest",
		broB06Razormaw: "Burrowing Razormaw", broB06Courser: "Gaea's Courser", broB06Skullflay: "Gixian Skullflayer",
		broB06Pallbearer: "Gnarlroot Pallbearer", broB06Butcher: "Battlefield Butcher",
		broB06Ritualist: "Deathbloom Ritualist", broB06Remorse: "Overwhelming Remorse",
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

// broB06Cycle is one row of the mill-three-take-one cycle: the card, a
// milled card it may take, and a milled card it may not.
type broB06Row struct {
	name, typeLine, oracle string
	takeable               game.Card
	notTakeable            game.Card
}

func broB06CycleRows() []broB06Row {
	return []broB06Row{
		{"Airlift Chaplain", "Creature — Human Cleric", broB06Chaplain,
			game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
			game.Card{Name: "Big Ox", TypeLine: "Creature — Ox", ManaCost: "{3}{G}"}},
		{"Blanchwood Prowler", "Creature — Elemental", broB06Prowler,
			game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"},
			game.Card{Name: "Milled Bolt", TypeLine: "Instant", ManaCost: "{R}"}},
		{"Fallaji Archaeologist", "Creature — Human Scout", broB06Archaeo,
			game.Card{Name: "Milled Bolt", TypeLine: "Instant", ManaCost: "{R}"},
			game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"}},
		{"Ravenous Gigamole", "Creature — Mole Horror", broB06Gigamole,
			game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
			game.Card{Name: "Milled Bolt", TypeLine: "Instant", ManaCost: "{R}"}},
		{"Tomakul Scrapsmith", "Creature — Human Artificer", broB06Scrapsmith,
			game.Card{Name: "Milled Rock", TypeLine: "Artifact", ManaCost: "{2}"},
			game.Card{Name: "Milled Bolt", TypeLine: "Instant", ManaCost: "{R}"}},
	}
}

func TestBroB06CycleTakesAMilledMatchAndSkipsTheCounter(t *testing.T) {
	for _, row := range broB06CycleRows() {
		t.Run(row.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			ids := effectstest.SeedSearchLibrary(me, row.takeable, row.notTakeable,
				broB06Filler(row, "Filler"),
				game.Card{Name: "Stays", TypeLine: "Basic Land — Forest"})
			src := effectstest.CastCatalogSpell(t, g, row.name, row.typeLine, row.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if pick == nil {
				t.Fatal("no prompt for a card from among the milled cards")
			}
			if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[0] {
				t.Fatalf("offered %v, want only the matching milled card", pick.ChooseCards)
			}
			effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
			if !me.Hand.Contains(ids[0]) {
				t.Error("the card did not go to hand")
			}
			if !me.Graveyard.Contains(ids[1]) || !me.Graveyard.Contains(ids[2]) {
				t.Error("the other milled cards are not in the graveyard")
			}
			if me.Graveyard.Contains(ids[3]) {
				t.Error("a fourth card was milled")
			}
			if got := effectstest.CountersOn(g, src, game.CounterPlusOne); got != 0 {
				t.Errorf("a card was taken and the creature has %d counters", got)
			}
		})
	}
}

func TestBroB06CycleGrowsWhenNothingIsTakenOrNothingMatches(t *testing.T) {
	for _, row := range broB06CycleRows() {
		for _, decline := range []bool{false, true} {
			name := row.name + " nothing matches"
			if decline {
				name = row.name + " declined"
			}
			t.Run(name, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me := g.Seats[0]
				first := row.notTakeable
				if decline {
					first = row.takeable
				}
				effectstest.SeedSearchLibrary(me, first, broB06Filler(row, "Miss 1"), broB06Filler(row, "Miss 2"))
				src := effectstest.CastCatalogSpell(t, g, row.name, row.typeLine, row.oracle, nil)
				effectstest.PassPriorityAroundTable(t, g)
				if decline {
					if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
						t.Fatal("no prompt with a matching card milled")
					}
					effectstest.AnswerChooseCards(t, g, me.ID)
				} else if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
					t.Fatal("a prompt with no matching card among the milled cards")
				}
				if got := effectstest.CountersOn(g, src, game.CounterPlusOne); got != 1 {
					t.Errorf("%d +1/+1 counters, want 1", got)
				}
			})
		}
	}
}

func TestBroB06AirliftChaplainOffersAPlainsAndASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Milled Plains", TypeLine: "Basic Land — Plains"},
		game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
		game.Card{Name: "Stays", TypeLine: "Instant"})
	effectstest.CastCatalogSpell(t, g, "Airlift Chaplain", "Creature — Human Cleric", broB06Chaplain, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if len(pick.ChooseCards) != 2 {
		t.Fatalf("offered %v, want the Plains and the cheap Bear but not the Forest", pick.ChooseCards)
	}
	for _, id := range pick.ChooseCards {
		if id == ids[1] {
			t.Error("a Forest was offered")
		}
	}
}

func TestBroB06WastefulHarvestMillsFiveAndTakesAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Milled Bolt", TypeLine: "Instant"},
		game.Card{Name: "Milled Ox", TypeLine: "Creature — Ox"},
		game.Card{Name: "Milled Wrath", TypeLine: "Sorcery"},
		game.Card{Name: "Milled Land", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Milled Rock", TypeLine: "Artifact"},
		game.Card{Name: "Stays", TypeLine: "Creature — Bear"})
	effectstest.CastCatalogSpell(t, g, "Wasteful Harvest", "Instant", broB06Harvest, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for a permanent card")
	}
	if len(pick.ChooseCards) != 3 {
		t.Fatalf("offered %v, want the creature, land and artifact", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1])
	if !me.Hand.Contains(ids[1]) {
		t.Error("the creature did not go to hand")
	}
	for _, i := range []int{0, 2, 3, 4} {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("milled card %d is not in the graveyard", i)
		}
	}
	if me.Graveyard.Contains(ids[5]) {
		t.Error("a sixth card was milled")
	}
}

// broB06Filler is a card the row's filter does not take.
func broB06Filler(row broB06Row, name string) game.Card {
	c := row.notTakeable
	c.Name = name
	return c
}

func broB06FillGraveyard(g *game.Game, p *game.Player, creatures int) {
	for i := 0; i < creatures; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Fallen", TypeLine: "Creature — Bear",
			Owner: p.ID, Controller: p.ID})
	}
	p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
}

func TestBroB06BurrowingRazormawMillsFourWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	maw := effectstest.B12Push(g, me.ID, "Burrowing Razormaw", "Creature — Beast", broB06Razormaw, 4, 2)
	before := me.Graveyard.Size()
	effectstest.KillCreature(t, g, me.ID, maw)
	effectstest.PassPriorityAroundTable(t, g)
	// The Razormaw itself is one of the cards in the graveyard.
	if got := me.Graveyard.Size() - before; got != 5 {
		t.Errorf("graveyard grew by %d, want the Razormaw and four milled cards", got)
	}
}

func TestBroB06GaeasCourserDrawsOnlyWithThreeCreatureCards(t *testing.T) {
	for _, n := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		courser := effectstest.B12Push(g, me.ID, "Gaea's Courser", "Creature — Centaur Soldier", broB06Courser, 4, 5)
		broB06FillGraveyard(g, me, n)
		before := effectstest.HandSize(me)
		effectstest.AttackWith(t, g, opp.ID, courser)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if n >= 3 {
			want = 1
		}
		if got := effectstest.HandSize(me) - before; got != want {
			t.Errorf("with %d creature cards drew %d, want %d", n, got, want)
		}
	}
}

func TestBroB06GixianSkullflayerGrowsOnlyWithThreeCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	flayer := effectstest.B12Push(g, me.ID, "Gixian Skullflayer", "Creature — Phyrexian Human Assassin", broB06Skullflay, 2, 3)
	broB06FillGraveyard(g, me, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, flayer, game.CounterPlusOne); got != 0 {
		t.Fatalf("%d counters with two creature cards", got)
	}
	broB06FillGraveyard(g, me, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, flayer, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters with three creature cards, want 1", got)
	}
}

func TestBroB06GnarlrootPallbearerPumpsByTheCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	broB06FillGraveyard(g, me, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Gnarlroot Pallbearer", "Creature — Treefolk Druid", broB06Pallbearer, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("Bear is %d/%d, want 5/5", p, tg)
	}
}

func TestBroB06BattlefieldButcherDiscountsItsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butcher := effectstest.PushCatalogPermanent(g, me.ID, "Battlefield Butcher", "Creature — Human Soldier", broB06Butcher, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, butcher, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {5} with two mana and no creature cards")
	}
	broB06FillGraveyard(g, me, 3)
	effectstest.B06AddMana(me, "C", "C")
	_ = opp
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, butcher, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("opponent life %d → %d, want -2", before, opp.Life)
	}
}

func TestBroB06DeathbloomRitualistAddsOneColorPerCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rit := effectstest.B12Push(g, me.ID, "Deathbloom Ritualist", "Creature — Elf Warlock", broB06Ritualist, 3, 5)
	broB06FillGraveyard(g, me, 3)
	if err := g.ActivateManaAbility(me.ID, rit, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 3 || got[0] != "G" || got[1] != "G" || got[2] != "G" {
		t.Errorf("pool = %v, want three green", got)
	}
}

func TestBroB06OverwhelmingRemorseCostsLessAndExilesCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Their Forest", "Basic Land — Forest", "", false)
	effectstest.AdvanceToMain(t, g)
	broB06FillGraveyard(g, me, 3)
	effectstest.B06AddMana(me, "B", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Overwhelming Remorse", "Instant", "{4}{B}", broB06Remorse, game.CastSpellParams{
		Strict: true, Targets: effectstest.B16TargetCard(land)}); err == nil {
		t.Fatal("targeted a land")
	}
	effectstest.B06AddMana(me, "B", "C")
	if _, err := effectstest.CastFromHandErr(g, me, "Overwhelming Remorse", "Instant", "{4}{B}", broB06Remorse, game.CastSpellParams{
		Strict: true, Targets: effectstest.B16TargetCard(theirs)}); err != nil {
		t.Fatalf("cast with three creature cards for {1}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature was not exiled")
	}
}
