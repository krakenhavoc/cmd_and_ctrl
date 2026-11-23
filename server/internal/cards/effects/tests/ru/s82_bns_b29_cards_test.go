package testsru

import (
	"github.com/google/uuid"

	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b29_cards_test.go — Foundations Jumpstart, BNS batch 29 (#5929, S82).
// Each test plays the real catalog entry.

const (
	bnsB29Druid      = "8d7933c4-4c74-492f-a9cd-50b339a4b29e"
	bnsB29Coilbug    = "a421582a-f673-4f43-beb5-28ee785da11e"
	bnsB29Disciple   = "8db63f1a-0d63-4b11-966d-818d83a082f9"
	bnsB29Skirmisher = "66578806-9148-479d-80e1-6c3e3b67e452"
	bnsB29Wurm       = "69191459-36ca-4e71-b4bb-0270c4cdaee9"
	bnsB29Epicure    = "bdb319a6-a1cd-44f5-999d-af1bb55b2fc4"
	bnsB29Champion   = "d432857f-cff6-421e-949b-ce2f65854e1d"
	bnsB29Diviner    = "bd6bc87d-fc75-43bc-a7d7-49d89dabcd3a"
	bnsB29Raptor     = "0a44da1d-7914-4592-a00f-7e40e0d07f91"
	bnsB29Skulker    = "1b6e0d65-202e-4fd3-861e-8d2eaffe3269"
	bnsB29Aviator    = "47f221a4-b51a-46b9-a4fc-46bc8ec7ddd3"
)

func TestBnsB29CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB29Druid: "Druid of the Anima", bnsB29Coilbug: "Durable Coilbug",
		bnsB29Disciple: "Elderfang Disciple", bnsB29Skirmisher: "Elite Skirmisher",
		bnsB29Wurm: "Enlisted Wurm", bnsB29Epicure: "Epicure of Blood",
		bnsB29Champion: "Expedition Champion", bnsB29Diviner: "Expedition Diviner",
		bnsB29Raptor: "Expedition Raptor", bnsB29Skulker: "Expedition Skulker",
		bnsB29Aviator: "Experimental Aviator",
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

func TestBnsB29DruidTapsForRedGreenOrWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Druid of the Anima", "Creature — Elf Druid", bnsB29Druid, true)
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("a summoning-sick Druid tapped for mana")
	}
	id := effectstest.PushCatalogPermanent(g, me.ID, "Druid of the Anima", "Creature — Elf Druid", bnsB29Druid, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the Druid made black mana")
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("tap for white: %v", err)
	}
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("the Druid did not tap")
	}
}

func TestBnsB29CoilbugReturnsFromTheGraveyardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gy := effectstest.PushCatalogGraveyardCard(me, "Durable Coilbug", "Creature — Insect", bnsB29Coilbug, 2, 2)
	if err := effectstest.BroB01Activate(t, g, me, gy, 0, "{C}{C}{C}{C}{C}"); err == nil {
		t.Fatal("returned with no black mana")
	}
	if err := effectstest.BroB01Activate(t, g, me, gy, 0, "{B}{C}{C}{C}{C}"); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gy) {
		t.Error("the Coilbug did not return to hand")
	}
	bf := effectstest.PushCatalogPermanent(g, me.ID, "Durable Coilbug", "Creature — Insect", bnsB29Coilbug, false)
	if err := effectstest.BroB01Activate(t, g, me, bf, 0, "{B}{C}{C}{C}{C}"); err == nil {
		t.Error("activated from the battlefield")
	}
}

func TestBnsB29DiscipleMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastAndResolveCreature(t, g, "Elderfang Disciple", "Creature — Elf Cleric", bnsB29Disciple)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		owed := effectstest.DiscardOwed(g, p.ID)
		if i == 0 && owed != 0 {
			t.Errorf("you owe nothing, got %d", owed)
		}
		if i != 0 && owed != 1 {
			t.Errorf("seat %d owes one discard, got %d", i, owed)
		}
	}
}

func TestBnsB29SkirmisherHeroicMayTapATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.AdvanceToMain(t, g)
	sk := effectstest.B12Push(g, me.ID, "Elite Skirmisher", "Creature — Human Soldier", bnsB29Skirmisher, 3, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(sk))
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TmtB1Card(t, g, bear).Tapped {
		t.Error("the heroic trigger did not tap the creature")
	}
}

func TestBnsB29SkirmisherHeroicDeclinedTapsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	effectstest.AdvanceToMain(t, g)
	sk := effectstest.B12Push(g, me.ID, "Elite Skirmisher", "Creature — Human Soldier", bnsB29Skirmisher, 3, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(sk))
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Card(t, g, bear).Tapped {
		t.Error("the creature was tapped though the controller declined")
	}
}

func TestBnsB29WurmCascadesWhenCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"})
	effectstest.B38CastCreature(t, g, "Enlisted Wurm", "Creature — Wurm", bnsB29Wurm, 5, 5)
	if g.Exile.Size() == 0 && me.Library.Size() == 20 {
		t.Error("cascade exiled nothing off the top of the library")
	}
}

func TestBnsB29EpicureDrainsWhenYouGainLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Epicure of Blood", "Creature — Vampire", bnsB29Epicure, false)
	before := []int{g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 4) })
	effectstest.PassPriorityAroundTable(t, g)
	for i, o := range g.Seats[1:] {
		if o.Life != before[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want a loss of 1", i, before[i], o.Life)
		}
	}
	mine := me.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[1].ID, 3) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine {
		t.Error("an opponent's life gain drained you")
	}
}

func TestBnsB29ChampionGetsPlusTwoWithAnotherWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Expedition Champion", "Creature — Human Warrior", bnsB29Champion, 2, 3)
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 2 {
		t.Errorf("power alone = %d, want 2 (it is not another Warrior)", p)
	}
	other := effectstest.B12Creature(g, g.Seats[1].ID, "Their Warrior", "Creature — Human Warrior", 1, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 2 {
		t.Errorf("power with an opponent's Warrior = %d, want 2", p)
	}
	mine := effectstest.B12Creature(g, me.ID, "My Warrior", "Creature — Human Warrior", 1, 1)
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 4 || tough != 3 {
		t.Errorf("with another Warrior: %d/%d, want 4/3", p, tough)
	}
	_ = other
	effectstest.KillCreature(t, g, me.ID, mine)
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 2 {
		t.Errorf("after the Warrior left: power %d, want 2", p)
	}
}

func TestBnsB29SkulkerHasDeathtouchWithAnotherRogue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Expedition Skulker", "Creature — Vampire Rogue", bnsB29Skulker, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("deathtouch with no other Rogue")
	}
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Rogue", "Creature — Human Rogue", 1, 1)
	if effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("deathtouch off an opponent's Rogue")
	}
	effectstest.B12Creature(g, me.ID, "My Rogue", "Creature — Human Rogue", 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("no deathtouch with another Rogue")
	}
}

func TestBnsB29DivinerDrawsOnDeathOnlyWithAnotherWizard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	alone := effectstest.B12Push(g, me.ID, "Expedition Diviner", "Creature — Merfolk Wizard", bnsB29Diviner, 3, 2)
	hand := me.Hand.Size()
	effectstest.SbondDie(t, g, me.ID, alone)
	if me.Hand.Size() != hand {
		t.Errorf("drew with no other Wizard: hand %d -> %d", hand, me.Hand.Size())
	}
	d := effectstest.B12Push(g, me.ID, "Expedition Diviner", "Creature — Merfolk Wizard", bnsB29Diviner, 3, 2)
	effectstest.B12Creature(g, me.ID, "Other Wizard", "Creature — Human Wizard", 1, 1)
	hand = me.Hand.Size()
	// A read through the layers, as any state change does in play, so the grant is on the Diviner when it dies.
	_ = effectstest.GaLayered(t, g, d)
	effectstest.SbondDie(t, g, me.ID, d)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 with another Wizard", hand, me.Hand.Size())
	}
}

func TestBnsB29RaptorSupportsTwoOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Ons2Opponent(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	raptor := effectstest.CastHoldCreature(t, g, "Expedition Raptor", "Creature — Bird", bnsB29Raptor, 2, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no support target prompt")
	}
	if effectstest.HasID(p.PickTargetCards, raptor) {
		t.Error("the Raptor was offered as its own support target")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(mine, theirs)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, mine, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, theirs, game.CounterPlusOne) != 1 {
		t.Error("support did not put a +1/+1 counter on each target")
	}
	if effectstest.CountersOn(g, raptor, game.CounterPlusOne) != 0 {
		t.Error("the Raptor put a counter on itself")
	}
}

func TestBnsB29AviatorMakesTwoFlyingThopters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Experimental Aviator", "Creature — Human Artificer", bnsB29Aviator)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Thopter"); n != 2 {
		t.Fatalf("%d Thopters, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && (!effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "flying") || len(c.Colors) != 0) {
			t.Error("a Thopter is not a colourless flier")
		}
	}
}
