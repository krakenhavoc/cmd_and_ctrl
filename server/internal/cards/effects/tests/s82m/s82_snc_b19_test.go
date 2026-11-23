package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b19_test.go — Streets of New Capenna, S82 batch 19: the
// exile-from-hand creature cycle and the three hideaway cards.

type sncB19Cycle struct {
	name, oracle, typeLine, cost string
	p, t                         int
	pool                         string // enough mana for the printed cost minus nothing: the whole cost
	colors                       []string
}

var sncB19CycleCards = []sncB19Cycle{
	{"Masked Bandits", "84ac3096-098a-477a-8907-fc30ac15f00f", "Creature — Raccoon Rogue", "{3}{B}{R}{G}", 5, 5, "BRGCCC", []string{"B", "R", "G"}},
	{"Rakish Revelers", "7db1402e-a434-45d2-9d21-9e0faeeaff10", "Creature — Elf Druid Rogue", "{2}{R}{G}{W}", 5, 3, "RGWCC", []string{"R", "G", "W"}},
	{"Spara's Adjudicators", "02718076-4c71-4bf5-988f-e6f94fbf0aef", "Creature — Cat Citizen", "{2}{G}{W}{U}", 4, 4, "GWUCC", []string{"G", "W", "U"}},
	{"Glamorous Outlaw", "cd4f7912-81c0-44dd-87c2-a98e955fe5a8", "Creature — Vampire Rogue", "{3}{U}{B}{R}", 4, 5, "UBRCCC", []string{"U", "B", "R"}},
	{"Shattered Seraph", "92eae6af-b915-4eb8-93c4-527f732bb712", "Creature — Angel Rogue", "{4}{W}{U}{B}", 4, 4, "WUBCCCC", []string{"W", "U", "B"}},
}

func TestSncB19AllCardsAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{
		"84ac3096-098a-477a-8907-fc30ac15f00f", "7db1402e-a434-45d2-9d21-9e0faeeaff10",
		"02718076-4c71-4bf5-988f-e6f94fbf0aef", "cd4f7912-81c0-44dd-87c2-a98e955fe5a8",
		"92eae6af-b915-4eb8-93c4-527f732bb712", "befc5986-f06c-40a4-98ee-ffcace4ceaa8",
		"2c2e4b19-ac88-426a-9f1b-ac305cd512e8", "e19e1e83-75e5-4cfa-8d50-8ed0b027a202",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v ok=%v", id, spec.Name, ok)
		}
	}
}

func sncB19Table(t *testing.T, c sncB19Cycle) (g *game.Game, me *game.Player, card, land uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	card = uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: card, Name: c.name, TypeLine: c.typeLine, ManaCost: c.cost, OracleID: c.oracle,
		Power: c.p, Toughness: c.t, Owner: me.ID, Controller: me.ID,
	})
	land = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Wastes", TypeLine: "Basic Land", Owner: me.ID, Controller: me.ID,
	})
	return
}

func sncB19Activate(t *testing.T, g *game.Game, me *game.Player, card, target uuid.UUID) error {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	return g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}, Strict: true,
	})
}

// The hand ability exiles as a cost, the land taps for one of the three
// colours, the cast from exile is open, and casting ends the grant.
func TestSncB19CycleHandAbilityGrantsTheLandAndOpensTheCast(t *testing.T) {
	for _, c := range sncB19CycleCards {
		t.Run(c.name, func(t *testing.T) {
			g, me, card, land := sncB19Table(t, c)
			if err := sncB19Activate(t, g, me, card, land); err != nil {
				t.Fatalf("activate from hand: %v", err)
			}
			if !effectstest.InExile(g, card) || me.Hand.Contains(card) {
				t.Fatal("the exile is a cost: the card is in exile as the ability goes on the stack")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.GrantedManaRowsOf(g, land); n != 1 {
				t.Fatalf("the land has %d granted mana abilities, want 1", n)
			}
			// Each printed colour is a legal pick, and the land taps once.
			if err := effectstest.TapGrantedMana(t, g, me.ID, land, c.colors[1]); err != nil {
				t.Fatalf("tap the land: %v", err)
			}
			if got := effectstest.PoolSize(g, me.ID); got != 1 {
				t.Errorf("pool = %d, want one mana", got)
			}
			var perm *game.CastPermission
			g.ReadSnapshot(func() {
				cc, _ := g.LookupCardForEffect(card)
				perm = g.CastPermissionOnCardForEffect(cc, game.ZoneExile)
			})
			if perm == nil || perm.Player != me.ID || perm.Cost != "" || perm.Duration.Kind != game.WhileInZone {
				t.Fatalf("cast permission = %+v, want the holder's at the printed cost while exiled", perm)
			}
			effectstest.FloatForTest(g, me, c.pool)
			if err := g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
				t.Fatalf("cast from exile: %v", err)
			}
			if n := effectstest.GrantedManaRowsOf(g, land); n != 0 {
				t.Errorf("the land keeps %d granted mana abilities after the cast", n)
			}
		})
	}
}

// Refusals: not from the battlefield or without a land target or the {2}.
func TestSncB19CycleHandAbilityRefusals(t *testing.T) {
	c := sncB19CycleCards[0]
	g, me, card, land := sncB19Table(t, c)
	nonland := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	if err := sncB19Activate(t, g, me, card, nonland); err == nil {
		t.Error("a nonland target was accepted")
	}
	if me.Hand.Contains(card) == false {
		t.Fatal("a refused activation must leave the card in hand")
	}
	// Short of {2}.
	g2, me2, card2, land2 := sncB19Table(t, c)
	if err := g2.AddManaForEffect(me2.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g2.ActivateCatalogAbility(me2.ID, card2, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land2}}, Strict: true,
	}); err == nil {
		t.Error("activated with one mana")
	}
	_ = land
}

func sncB19Cast(t *testing.T, g *game.Game, me *game.Player, c sncB19Cycle) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: c.name, TypeLine: c.typeLine, ManaCost: c.cost, OracleID: c.oracle,
		Power: c.p, Toughness: c.t, Owner: me.ID, Controller: me.ID,
	})
	effectstest.FloatForTest(g, me, c.pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", c.name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSncB19RakishRevelersMakesACitizen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	sncB19Cast(t, g, me, sncB19CycleCards[1])
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Citizen" && c.Controller == me.ID && c.Power == 1 && c.Toughness == 1 {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("Citizens = %d, want 1", n)
	}
}

func TestSncB19ShatteredSeraphGainsThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	sncB19Cast(t, g, me, sncB19CycleCards[4])
	if me.Life != before+3 {
		t.Fatalf("life = %d, want %d", me.Life, before+3)
	}
}

func TestSncB19GlamorousOutlawDamagesEachOpponentAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	sncB19Cast(t, g, me, sncB19CycleCards[3])
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		want := before[s.ID]
		if s.ID != me.ID {
			want -= 2
		}
		if s.Life != want {
			t.Errorf("%s life = %d, want %d", s.ID, s.Life, want)
		}
	}
}

func TestSncB19SparasAdjudicatorsStopsAnOpposingCreatureUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	c := sncB19CycleCards[2]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: c.typeLine, ManaCost: c.cost, OracleID: c.oracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, c.pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, theirs, game.CantAttackOrBlock)
	if card := effectstest.LayeredCard(t, g, mine); game.Restricted(&card, game.CantAttack) || game.Restricted(&card, game.CantBlock) {
		t.Fatal("my own creature was restricted")
	}
}
