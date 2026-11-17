package effects

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 — The Brothers' War, batch 4: artifact creatures with unearth.

type broB04Card struct {
	name, typeLine, oracle string
	power, toughness       int
	unearth                string
}

var broB04Cards = []broB04Card{
	{"Ashnod's Harvester", "Artifact Creature — Construct", "9ea87566-43da-4ebd-a904-ae8b8830c3a2", 3, 1, "{1}{B}"},
	{"Combat Courier", "Artifact Creature — Construct", "54d32245-c796-4e8e-b61e-c218e3902565", 1, 1, "{U}"},
	{"Mishra's Juggernaut", "Artifact Creature — Juggernaut", "fa6fa2f5-eb9b-402e-82f4-e0a8d30df245", 5, 3, "{5}{R}"},
	{"Reconstructed Thopter", "Artifact Creature — Thopter", "c87ff8e8-a3da-4848-9881-fbc896acf7c5", 2, 1, "{2}"},
	{"Scrapwork Cohort", "Artifact Creature — Soldier", "149694e6-5cf0-4983-9546-a4bd579b2ade", 3, 1, "{2}{W}"},
	{"Scrapwork Mutt", "Artifact Creature — Dog", "5906b378-c7a7-44fd-a63b-55a347d97854", 2, 1, "{1}{R}"},
	{"Scrapwork Rager", "Artifact Creature — Phyrexian Horror", "552867ec-5741-4e34-813f-4aaf185661f7", 2, 2, "{3}{B}"},
	{"Simian Simulacrum", "Artifact Creature — Ape", "77eb44ca-53e9-4924-b19f-31732f14ef14", 2, 1, "{2}{G}{G}"},
	{"Tocasia's Onulet", "Artifact Creature — Construct", "150dd67b-4de2-4fcd-b753-e4096de911a2", 4, 4, "{3}{W}"},
	{"Yotian Frontliner", "Artifact Creature — Soldier", "cd39e485-71ba-40cb-8608-f5a0dfb7ef5f", 1, 1, "{W}"},
	{"Heavyweight Demolisher", "Artifact Creature — Construct", "fe71ada0-ac86-4035-9b6b-4c24ae65034c", 8, 6, "{6}{R}{R}"},
	{"Artificer's Dragon", "Artifact Creature — Dragon", "4284ff9f-4085-4874-abc7-95467cb759a5", 4, 4, "{3}{R}{R}"},
}

// Each card carries exactly its graveyard unearth ability, at the printed
// cost, sorcery speed, and not offered anywhere else.
func TestBroB04UnearthAbilities(t *testing.T) {
	for _, tc := range broB04Cards {
		var found *game.ActivatedAbilityShape
		for _, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle}) {
			if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
				ab := ab
				found = &ab
			}
		}
		if found == nil {
			t.Errorf("%s: no graveyard ability", tc.name)
			continue
		}
		if found.Cost.Mana != tc.unearth || !found.SorcerySpeed {
			t.Errorf("%s: unearth cost %q sorcery %v, want %q sorcery", tc.name, found.Cost.Mana, found.SorcerySpeed, tc.unearth)
		}
		if game.AbilityFunctionsFromZone(*found, game.ZoneBattlefield) {
			t.Errorf("%s: unearth is offered on the battlefield", tc.name)
		}
	}
}

func broB04Unearth(t *testing.T, g *game.Game, tc broB04Card) (uuid.UUID, *game.Player) {
	t.Helper()
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.oracle})
	idx := -1
	for i, ab := range abilities {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("%s: no unearth", tc.name)
	}
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	id := pushCatalogGraveyardCard(me, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
	rfCrAAddMana(t, g, me, broB04Pool(tc.unearth))
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: unearth: %v", tc.name, err)
	}
	passPriorityAroundTable(t, g)
	return id, me
}

// Every card unearths: it returns with haste.
func TestBroB04EveryCardUnearthsWithHaste(t *testing.T) {
	for _, tc := range broB04Cards {
		g := newCatalogGame(t)
		id, _ := broB04Unearth(t, g, tc)
		if p := latestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID); p != nil {
			_ = g.ResolvePickTarget(p.ID, g.Seats[g.Turn.ActiveSeat].ID, game.TargetRef{Kind: game.TargetCard, ID: id})
			passPriorityAroundTable(t, g)
		}
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s: not on the battlefield after unearth", tc.name)
			continue
		}
		c, _ := g.LookupCardForEffect(id)
		if !game.HasKeyword(&c, "haste") {
			t.Errorf("%s: unearthed without haste", tc.name)
		}
	}
}

func TestBroB04CombatCourierSacrificesToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	c := rfCrAPush(g, me.ID, "Combat Courier", "Artifact Creature — Construct", broB04Cards[1].oracle, 1, 1)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	rfCrAAddMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, c, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c) {
		t.Error("Courier was not sacrificed")
	}
	if !me.Graveyard.Contains(c) {
		t.Error("Courier is not in the graveyard (it should be unearthable)")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want +1", hand, len(me.Hand.Cards))
	}
}

func TestBroB04ScrapworkCohortMakesASoldier(t *testing.T) {
	g := newCatalogGame(t)
	before := len(g.Battlefield.Cards)
	broB04Unearth(t, g, broB04Cards[4])
	if len(g.Battlefield.Cards)-before != 2 {
		t.Fatalf("battlefield grew by %d, want Cohort + Soldier token", len(g.Battlefield.Cards)-before)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Soldier") && c.IsArtifact() && c.IsCreature() {
			found = true
		}
	}
	if !found {
		t.Error("no Soldier artifact creature token")
	}
}

func TestBroB04ScrapworkRagerDrawsAndLosesLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	hand, life := len(me.Hand.Cards), me.Life
	broB04Unearth(t, g, broB04Cards[6])
	if len(me.Hand.Cards) != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d; want +1 card, -1 life", hand, len(me.Hand.Cards), life, me.Life)
	}
}

func TestBroB04TocasiasOnuletGainsTwoWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	id, me := broB04Unearth(t, g, broB04Cards[8])
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if !g.Exile.Contains(id) {
		t.Error("the unearthed Onulet should be exiled")
	}
}

func TestBroB04ArtificersDragonPumpsArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	d := rfCrAPush(g, me.ID, "Artificer's Dragon", "Artifact Creature — Dragon", broB04Cards[11].oracle, 4, 4)
	art := rfCrAPush(g, me.ID, "Test Golem", "Artifact Creature — Golem", "test-golem", 2, 2)
	plain := rfCrAPush(g, me.ID, "Test Bear", "Creature — Bear", "test-bear", 2, 2)
	rfCrAAddMana(t, g, me, "{R}")
	b16Activate(t, g, me.ID, d, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if effectivePower(t, g, d) != 5 || effectivePower(t, g, art) != 3 {
		t.Errorf("artifact creatures: dragon %d golem %d, want 5 and 3", effectivePower(t, g, d), effectivePower(t, g, art))
	}
	if effectivePower(t, g, plain) != 2 {
		t.Errorf("a non-artifact creature got the bonus: %d", effectivePower(t, g, plain))
	}
}

func TestBroB04SimianSimulacrumPutsCountersOnATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	bear := rfCrAPush(g, me.ID, "Test Bear", "Creature — Bear", "test-bear", 2, 2)
	broB04Unearth(t, g, broB04Cards[7])
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
	}
	if got := effectivePowerWithCounters(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4", got)
	}
}

func TestBroB04AshnodsHarvesterExilesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	h := rfCrAPush(g, me.ID, "Ashnod's Harvester", "Artifact Creature — Construct", broB04Cards[0].oracle, 3, 1)
	victim := pushGraveyardCardForTest(opp, "Doomed")
	declareAttack(t, g, opp.ID, h)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
	}
	if !g.Exile.Contains(victim) {
		t.Fatal("the graveyard card was not exiled")
	}
}

func TestBroB04YotianFrontlinerPumpsAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	f := rfCrAPush(g, me.ID, "Yotian Frontliner", "Artifact Creature — Soldier", broB04Cards[9].oracle, 1, 1)
	bear := rfCrAPush(g, me.ID, "Test Bear", "Creature — Bear", "test-bear", 2, 2)
	declareAttack(t, g, opp.ID, f)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
	}
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power %d, want 3", got)
	}
	if got := effectivePower(t, g, f); got != 1 {
		t.Errorf("Frontliner power %d, want 1 (\"another\")", got)
	}
}

func TestBroB04MishrasJuggernautMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	j := b7Put(g, me.ID, "Mishra's Juggernaut", "Artifact Creature — Juggernaut", broB04Cards[2].oracle, 5, 3)
	if !hasEffectiveKeyword(t, g, j, "trample") {
		t.Error("no trample")
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(j, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	if opp.Life != 40-5 {
		t.Errorf("opponent life %d, want 35", opp.Life)
	}
}

func TestBroB04HeavyweightDemolisherTapsUnlessPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		d := rfCrAPush(g, me.ID, "Heavyweight Demolisher", "Artifact Creature — Construct", broB04Cards[10].oracle, 8, 6)
		if !hasEffectiveKeyword(t, g, d, "menace") {
			t.Error("no menace")
		}
		walkToUpkeepOfSeat0(t, g)
		if pay {
			rfCrAAddMana(t, g, me, "{C}{C}{C}")
		}
		passPriorityAroundTable(t, g)
		upkeepPromptFor(t, g, me.ID)
		answerPayUnless(t, g, me.ID, pay)
		if got := isTapped(g, d); got == pay {
			t.Errorf("pay=%v: tapped=%v, want %v", pay, got, !pay)
		}
	}
}

// broB04Pool turns a printed cost into mana to float: generic {N} as N {C}.
func broB04Pool(cost string) string {
	return regexp.MustCompile(`\{(\d+)\}`).ReplaceAllStringFunc(cost, func(m string) string {
		n, _ := strconv.Atoi(m[1 : len(m)-1])
		return strings.Repeat("{C}", n)
	})
}
