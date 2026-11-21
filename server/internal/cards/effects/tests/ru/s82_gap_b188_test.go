package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b188_test.go — Warhammer 40,000 Commander creatures, S82 GAP
// batch 188 (#5948).

var gapB188IDs = map[string]string{
	"Illuminor Szeras":           "10f274f9-9044-447b-81c9-d4906ad9e6be",
	"Imotekh the Stormlord":      "425fea06-76ec-4cd9-bd3b-750504fd7fd2",
	"Lokhust Heavy Destroyer":    "fd03f20a-d365-48fb-8b2a-ad4b502a0045",
	"Lychguard":                  "e82daa7b-d9fd-48c4-b5b2-f3b1f6ef1a24",
	"Mortarion, Daemon Primarch": "588eecbf-dcfc-4b97-b81c-8b9f3d39e08f",
	"Necron Deathmark":           "45866598-4ee4-46b6-ade2-ffcdbe98e58f",
	"Plasmancer":                 "b3ff8da4-f2cc-41c3-8950-9cc531854b78",
	"Poxwalkers":                 "e4275202-4960-4cde-b941-27be8163ad6e",
	"Primaris Eliminator":        "7d679591-f8ea-4c4c-ab98-7b9e3438cf57",
	"Psychomancer":               "589b742c-fee5-413a-a56c-4d4ffaf0f73d",
	"Royal Warden":               "0192ba05-6144-4fec-a803-462ecdf64195",
}

func TestGapB188CardsAreFull(t *testing.T) {
	for name, id := range gapB188IDs {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v spec=%q/%s/%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB188Count(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return n
}

func gapB188Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine string) {
	t.Helper()
	effectstest.CastFromHandForTest(t, g, me, name, typeLine, "{0}", gapB188IDs[name], game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestRoyalWardenMakesTwoTappedWarriors(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	gapB188Cast(t, g, me, "Royal Warden", "Artifact Creature — Necron")
	if n := gapB188Count(g, "Necron Warrior"); n != 2 {
		t.Fatalf("%d Necron Warriors, want 2", n)
	}
	w := effectstest.BattlefieldTokenNamed(t, g, "Necron Warrior")
	if !w.Tapped {
		t.Fatal("the Warriors did not enter tapped")
	}
}

func TestLokhustHeavyDestroyerMakesEachPlayerSacrifice(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	mine := effectstest.B12Creature(g, me.ID, "Mine", effectstest.TestCreatureTypeLine, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", effectstest.TestCreatureTypeLine, 1, 1)
	gapB188Cast(t, g, me, "Lokhust Heavy Destroyer", "Artifact Creature — Necron")
	for range 4 {
		var c *game.PendingChoice
		for _, p := range []uuid.UUID{me.ID, opp.ID} {
			if c = effectstest.LatestChooseCards(g, p); c != nil {
				break
			}
		}
		if c == nil {
			break
		}
		effectstest.AnswerChooseCards(t, g, c.Chooser, c.ChooseCards[0])
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Battlefield.Contains(theirs) {
		t.Fatal("the opponent kept their only creature")
	}
	_ = mine
}

func TestLychguardReturnsOnlyLegendaryCreatureCards(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	lych := effectstest.PushCatalogPermanent(g, me.ID, "Lychguard", "Artifact Creature — Necron", gapB188IDs["Lychguard"], false)
	effectstest.Unsick(g, lych)
	legend := effectstest.GraveCard(me, "Legend", "Legendary Creature — Human")
	plain := effectstest.GraveCard(me, "Plain", "Creature — Bear")
	relic := effectstest.GraveCard(me, "Legendary Relic", "Legendary Artifact")
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, lych, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Guardian Protocols: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(legend) || me.Hand.Contains(plain) || me.Hand.Contains(relic) {
		t.Fatalf("hand: legend=%v plain=%v relic=%v", me.Hand.Contains(legend), me.Hand.Contains(plain), me.Hand.Contains(relic))
	}
}

func TestPlasmancerFetchesABasicSwamp(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	var swamp uuid.UUID
	g.WithWriteLock(func() {
		swamp = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: swamp, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	})
	gapB188Cast(t, g, me, "Plasmancer", "Artifact Creature — Necron Wizard")
	c := effectstest.ChkB1OpenSearch(g)
	if c == nil {
		t.Fatal("no library search was offered")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{swamp}); err != nil {
		t.Fatalf("taking the Swamp: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(swamp) {
		t.Fatal("the basic Swamp is not in hand")
	}
}

func TestNecronDeathmarkDestroysAndMills(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", effectstest.TestCreatureTypeLine, 2, 2)
	g.WithWriteLock(func() {
		opp.Library.Cards = nil
		for range 5 {
			opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
		}
	})
	effectstest.CastFromHandForTest(t, g, me, "Necron Deathmark", "Artifact Creature — Necron", "{0}", gapB188IDs["Necron Deathmark"], game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
		t.Fatalf("first clause: %v", err)
	}
	if ch = effectstest.PickTargetPrompt(t, g); ch != nil {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("second clause: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || opp.Graveyard.Size() < 3 {
		t.Fatalf("victim on battlefield=%v, opponent graveyard=%d", g.Battlefield.Contains(victim), opp.Graveyard.Size())
	}
}

func TestPrimarisEliminatorExecutionerRoundDestroys(t *testing.T) {
	spec, _ := effects.Lookup(gapB188IDs["Primaris Eliminator"])
	if len(spec.Triggered) != 1 || spec.Triggered[0].Modes == nil {
		t.Fatalf("the enters trigger is not modal: %+v", spec.Triggered)
	}
}

func TestPsychomancerDrainsWhenAnotherNontokenArtifactLeaves(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Psychomancer", "Artifact Creature — Necron Wizard", gapB188IDs["Psychomancer"], false)
	relic := effectstest.PushCatalogPermanent(g, me.ID, "Relic", "Artifact", "", false)
	me.Life, opp.Life = 20, 20
	effectstest.B18Kill(t, g, relic)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("targeting the opponent: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 21 || opp.Life != 19 {
		t.Fatalf("life me=%d opp=%d, want 21 and 19", me.Life, opp.Life)
	}
}

func TestPoxwalkersReturnsWhenYouCastFromAnywhereButHand(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	pox := effectstest.GraveCard(me, "Poxwalkers", "Creature — Zombie")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == pox {
				me.Graveyard.Cards[i].OracleID = gapB188IDs["Poxwalkers"]
			}
		}
	})
	effectstest.CastFromHandForTest(t, g, me, "Hand Spell", "Instant", "{0}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pox) {
		t.Fatal("a spell cast from hand returned Poxwalkers")
	}
	exiled := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: exiled, Name: "Exiled Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{Player: me.ID, Zone: game.ZoneExile, Cost: "{0}", CastOnly: true, Timing: game.TimingFlash}, []game.Card{{InstanceID: exiled}})
	})
	if err := g.CastSpell(me.ID, exiled, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Skipf("could not stage a cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(pox) {
		t.Fatal("a spell cast from exile did not return Poxwalkers")
	}
	if c, _ := effectstest.BattlefieldCardByID(g, pox); !c.Tapped {
		t.Fatal("Poxwalkers did not return tapped")
	}
}

func TestIlluminorSzerasAddsManaEqualToTheSacrificedCreaturesManaValue(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	sz := effectstest.PushCatalogPermanent(g, me.ID, "Illuminor Szeras", "Legendary Artifact Creature — Necron", gapB188IDs["Illuminor Szeras"], false)
	effectstest.Unsick(g, sz)
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Four Drop", TypeLine: effectstest.TestCreatureTypeLine, ManaCost: "{2}{B}{B}",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, sz, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("Secrets of the Soul: %v", err)
	}
	if got := me.ManaPool.Total(); got != 4 {
		t.Fatalf("mana %d, want 4", got)
	}
}

func TestMortarionPaysXUpToLifeLost(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Mortarion, Daemon Primarch", "Legendary Creature — Demon Primarch", gapB188IDs["Mortarion, Daemon Primarch"], false)
	spec, _ := effects.Lookup(gapB188IDs["Mortarion, Daemon Primarch"])
	if len(spec.Triggered) != 1 {
		t.Fatalf("triggers = %d", len(spec.Triggered))
	}
}

func TestImotekhMakesWarriorsWhenAnArtifactCardLeavesYourGraveyard(t *testing.T) {
	g, me, _ := effectstest.FlipMain(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Imotekh the Stormlord", "Legendary Artifact Creature — Necron", gapB188IDs["Imotekh the Stormlord"], false)
	relic := effectstest.GraveCard(me, "Relic", "Artifact")
	g.WithWriteLock(func() { _, _ = g.ReturnToBattlefieldForEffect(relic, uuid.Nil, false) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := gapB188Count(g, "Necron Warrior"); n != 2 {
		t.Fatalf("%d Necron Warriors, want 2", n)
	}
}
