package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

var broB04Cards = []effectstest.BroB04Card{
	{Name: "Ashnod's Harvester", TypeLine: "Artifact Creature — Construct", Oracle: "9ea87566-43da-4ebd-a904-ae8b8830c3a2", Power: 3, Toughness: 1, Unearth: "{1}{B}"},
	{Name: "Combat Courier", TypeLine: "Artifact Creature — Construct", Oracle: "54d32245-c796-4e8e-b61e-c218e3902565", Power: 1, Toughness: 1, Unearth: "{U}"},
	{Name: "Mishra's Juggernaut", TypeLine: "Artifact Creature — Juggernaut", Oracle: "fa6fa2f5-eb9b-402e-82f4-e0a8d30df245", Power: 5, Toughness: 3, Unearth: "{5}{R}"},
	{Name: "Reconstructed Thopter", TypeLine: "Artifact Creature — Thopter", Oracle: "c87ff8e8-a3da-4848-9881-fbc896acf7c5", Power: 2, Toughness: 1, Unearth: "{2}"},
	{Name: "Scrapwork Cohort", TypeLine: "Artifact Creature — Soldier", Oracle: "149694e6-5cf0-4983-9546-a4bd579b2ade", Power: 3, Toughness: 1, Unearth: "{2}{W}"},
	{Name: "Scrapwork Mutt", TypeLine: "Artifact Creature — Dog", Oracle: "5906b378-c7a7-44fd-a63b-55a347d97854", Power: 2, Toughness: 1, Unearth: "{1}{R}"},
	{Name: "Scrapwork Rager", TypeLine: "Artifact Creature — Phyrexian Horror", Oracle: "552867ec-5741-4e34-813f-4aaf185661f7", Power: 2, Toughness: 2, Unearth: "{3}{B}"},
	{Name: "Simian Simulacrum", TypeLine: "Artifact Creature — Ape", Oracle: "77eb44ca-53e9-4924-b19f-31732f14ef14", Power: 2, Toughness: 1, Unearth: "{2}{G}{G}"},
	{Name: "Tocasia's Onulet", TypeLine: "Artifact Creature — Construct", Oracle: "150dd67b-4de2-4fcd-b753-e4096de911a2", Power: 4, Toughness: 4, Unearth: "{3}{W}"},
	{Name: "Yotian Frontliner", TypeLine: "Artifact Creature — Soldier", Oracle: "cd39e485-71ba-40cb-8608-f5a0dfb7ef5f", Power: 1, Toughness: 1, Unearth: "{W}"},
	{Name: "Heavyweight Demolisher", TypeLine: "Artifact Creature — Construct", Oracle: "fe71ada0-ac86-4035-9b6b-4c24ae65034c", Power: 8, Toughness: 6, Unearth: "{6}{R}{R}"},
	{Name: "Artificer's Dragon", TypeLine: "Artifact Creature — Dragon", Oracle: "4284ff9f-4085-4874-abc7-95467cb759a5", Power: 4, Toughness: 4, Unearth: "{3}{R}{R}"},
}

// Each card carries exactly its graveyard unearth ability, at the printed
// cost, sorcery speed, and not offered anywhere else.
func TestBroB04UnearthAbilities(t *testing.T) {
	for _, tc := range broB04Cards {
		var found *game.ActivatedAbilityShape
		for _, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: tc.Oracle}) {
			if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
				ab := ab
				found = &ab
			}
		}
		if found == nil {
			t.Errorf("%s: no graveyard ability", tc.Name)
			continue
		}
		if found.Cost.Mana != tc.Unearth || !found.SorcerySpeed {
			t.Errorf("%s: unearth cost %q sorcery %v, want %q sorcery", tc.Name, found.Cost.Mana, found.SorcerySpeed, tc.Unearth)
		}
		if game.AbilityFunctionsFromZone(*found, game.ZoneBattlefield) {
			t.Errorf("%s: unearth is offered on the battlefield", tc.Name)
		}
	}
}

// Every card unearths: it returns with haste.
func TestBroB04EveryCardUnearthsWithHaste(t *testing.T) {
	for _, tc := range broB04Cards {
		g := effectstest.NewCatalogGame(t)
		id, _ := effectstest.BroB04Unearth(t, g, tc)
		if p := effectstest.LatestPickTarget(g, g.Seats[g.Turn.ActiveSeat].ID); p != nil {
			_ = g.ResolvePickTarget(p.ID, g.Seats[g.Turn.ActiveSeat].ID, game.TargetRef{Kind: game.TargetCard, ID: id})
			effectstest.PassPriorityAroundTable(t, g)
		}
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s: not on the battlefield after unearth", tc.Name)
			continue
		}
		c, _ := g.LookupCardForEffect(id)
		if !game.HasKeyword(&c, "haste") {
			t.Errorf("%s: unearthed without haste", tc.Name)
		}
	}
}

func TestBroB04CombatCourierSacrificesToDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	c := effectstest.RfCrAPush(g, me.ID, "Combat Courier", "Artifact Creature — Construct", broB04Cards[1].Oracle, 1, 1)
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	hand := len(me.Hand.Cards)
	effectstest.RfCrAAddMana(t, g, me, "{C}{C}")
	effectstest.B16Activate(t, g, me.ID, c, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	before := len(g.Battlefield.Cards)
	effectstest.BroB04Unearth(t, g, broB04Cards[4])
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	hand, life := len(me.Hand.Cards), me.Life
	effectstest.BroB04Unearth(t, g, broB04Cards[6])
	if len(me.Hand.Cards) != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d -> %d, life %d -> %d; want +1 card, -1 life", hand, len(me.Hand.Cards), life, me.Life)
	}
}

func TestBroB04TocasiasOnuletGainsTwoWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.BroB04Unearth(t, g, broB04Cards[8])
	life := me.Life
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
	if !g.Exile.Contains(id) {
		t.Error("the unearthed Onulet should be exiled")
	}
}

func TestBroB04ArtificersDragonPumpsArtifactCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	d := effectstest.RfCrAPush(g, me.ID, "Artificer's Dragon", "Artifact Creature — Dragon", broB04Cards[11].Oracle, 4, 4)
	art := effectstest.RfCrAPush(g, me.ID, "Test Golem", "Artifact Creature — Golem", "test-golem", 2, 2)
	plain := effectstest.RfCrAPush(g, me.ID, "Test Bear", "Creature — Bear", "test-bear", 2, 2)
	effectstest.RfCrAAddMana(t, g, me, "{R}")
	effectstest.B16Activate(t, g, me.ID, d, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, d) != 5 || effectstest.EffectivePower(t, g, art) != 3 {
		t.Errorf("artifact creatures: dragon %d golem %d, want 5 and 3", effectstest.EffectivePower(t, g, d), effectstest.EffectivePower(t, g, art))
	}
	if effectstest.EffectivePower(t, g, plain) != 2 {
		t.Errorf("a non-artifact creature got the bonus: %d", effectstest.EffectivePower(t, g, plain))
	}
}

func TestBroB04SimianSimulacrumPutsCountersOnATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	bear := effectstest.RfCrAPush(g, me.ID, "Test Bear", "Creature — Bear", "test-bear", 2, 2)
	effectstest.BroB04Unearth(t, g, broB04Cards[7])
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, bear); got != 4 {
		t.Errorf("Bear power %d, want 4", got)
	}
}

func TestBroB04AshnodsHarvesterExilesOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	h := effectstest.RfCrAPush(g, me.ID, "Ashnod's Harvester", "Artifact Creature — Construct", broB04Cards[0].Oracle, 3, 1)
	victim := effectstest.PushGraveyardCardForTest(opp, "Doomed")
	effectstest.DeclareAttack(t, g, opp.ID, h)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Exile.Contains(victim) {
		t.Fatal("the graveyard card was not exiled")
	}
}

func TestBroB04YotianFrontlinerPumpsAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	f := effectstest.RfCrAPush(g, me.ID, "Yotian Frontliner", "Artifact Creature — Soldier", broB04Cards[9].Oracle, 1, 1)
	bear := effectstest.RfCrAPush(g, me.ID, "Test Bear", "Creature — Bear", "test-bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, f)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("Bear power %d, want 3", got)
	}
	if got := effectstest.EffectivePower(t, g, f); got != 1 {
		t.Errorf("Frontliner power %d, want 1 (\"another\")", got)
	}
}

func TestBroB04MishrasJuggernautMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	j := effectstest.B7Put(g, me.ID, "Mishra's Juggernaut", "Artifact Creature — Juggernaut", broB04Cards[2].Oracle, 5, 3)
	if !effectstest.HasEffectiveKeyword(t, g, j, "trample") {
		t.Error("no trample")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(j, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if opp.Life != 40-5 {
		t.Errorf("opponent life %d, want 35", opp.Life)
	}
}

func TestBroB04HeavyweightDemolisherTapsUnlessPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		d := effectstest.RfCrAPush(g, me.ID, "Heavyweight Demolisher", "Artifact Creature — Construct", broB04Cards[10].Oracle, 8, 6)
		if !effectstest.HasEffectiveKeyword(t, g, d, "menace") {
			t.Error("no menace")
		}
		effectstest.WalkToUpkeepOfSeat0(t, g)
		if pay {
			effectstest.RfCrAAddMana(t, g, me, "{C}{C}{C}")
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.UpkeepPromptFor(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		if got := effectstest.IsTapped(g, d); got == pay {
			t.Errorf("pay=%v: tapped=%v, want %v", pay, got, !pay)
		}
	}
}
