package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b131_cards_test.go — S82 OLD batch 131 (#5514): Theros Beyond
// Death graveyard recursion. Every test plays the real catalog card through
// the engine.

const (
	b131Storyteller = "0d208c1e-c273-437f-9aea-136400e450e1"
	b131Herald      = "1692602a-0c44-485c-aadc-c9c70ed70ea8"
	b131Evangel     = "4d6e38f6-480b-4f81-a5d4-342fc45e39b0"
	b131Glory       = "a41297c4-5c2d-475d-b1f5-439df30fbad7"
	b131Acolyte     = "d57c246f-18e5-4523-898c-1509db738363"
	b131Klothys     = "e8473969-195f-41ee-a84d-463cc8c0ef02"
	b131Sentinel    = "eef75645-1aa5-4085-8cbd-8b6fec2cbdc2"
	// Arcane Flight (Enchant creature, +1/+1 and flying), a catalog Aura.
	b131ArcaneFlight = "40eee955-bec6-455a-addf-68cdb84c76b9"
	b131Bear         = "Creature — Bear"
)

func TestB131EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b131Storyteller: "Lagonna-Band Storyteller", b131Herald: "Storm Herald", b131Evangel: "Dawn Evangel",
		b131Glory: "Rise to Glory", b131Acolyte: "Acolyte of Affliction", b131Klothys: "Klothys, God of Destiny",
		b131Sentinel: "Underworld Sentinel",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func b131Grave(p *game.Player, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// b131Settle answers whatever a trigger asks and lets the stack empty.
func b131Settle(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	for i := 0; i < 8; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, who) != nil || effectstest.LatestTriggerPrompt(g, who) != nil {
			return
		}
	}
}

func TestB131LagonnaBandStorytellerPutsAnEnchantmentOnTopAndGainsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ench := b131Grave(me, "Some Aura", "Enchantment — Aura", "{2}{W}", "", 0, 0)
	creature := b131Grave(me, "Some Bear", b131Bear, "{1}{G}", "", 2, 2)
	life := me.Life
	effectstest.B12PlayFromHand(t, g, "Lagonna-Band Storyteller", "Creature — Centaur Advisor", b131Storyteller, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: creature}); err == nil {
			t.Error("a creature card was a legal target")
		}
		effectstest.PickCard(t, g, me.ID, ench)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(ench) {
		t.Error("the enchantment is still in the graveyard")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ench {
		t.Error("the enchantment is not on top of the library")
	}
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want +3", life, me.Life)
	}
}

func TestB131LagonnaBandStorytellerDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ench := b131Grave(me, "Some Aura", "Enchantment — Aura", "{2}{W}", "", 0, 0)
	life := me.Life
	effectstest.B12PlayFromHand(t, g, "Lagonna-Band Storyteller", "Creature — Centaur Advisor", b131Storyteller, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, ench)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(ench) || me.Life != life {
		t.Errorf("declined: enchantment in graveyard %v, life %d → %d", me.Graveyard.Contains(ench), life, me.Life)
	}
}

func TestB131AcolyteOfAfflictionMillsTwoThenReturnsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b131Grave(me, "Old Forest", "Basic Land — Forest", "", "", 0, 0)
	spell := b131Grave(me, "Old Bolt", "Instant", "{R}", "", 0, 0)
	effectstest.SeedLibrary(me, "A", "B", "C")
	libBefore := me.Library.Size()
	effectstest.B12PlayFromHand(t, g, "Acolyte of Affliction", "Creature — Human Cleric", b131Acolyte, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != libBefore-2 {
		t.Errorf("library %d → %d, want two milled", libBefore, got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt for the return")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("the land card did not return to hand")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("a non-permanent card left the graveyard")
	}
}

func TestB131DawnEvangelReturnsASmallCreatureWhenAnEnchantedCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Dawn Evangel", "Enchantment Creature — Human Cleric", b131Evangel, false)
	small := b131Grave(me, "Small Bear", b131Bear, "{1}{G}", "", 2, 2)
	big := b131Grave(me, "Big Bear", b131Bear, "{3}{G}", "", 4, 4)
	victim := effectstest.PushVanillaCreature(g, me.ID, "Victim", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Arcane Flight", "Enchantment — Aura", b131ArcaneFlight, effectstest.B16TargetCard(victim))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, victim)
	g.RunStateChecksForTest()
	b131Settle(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
			t.Error("a mana value 4 card was a legal target")
		}
		effectstest.PickCard(t, g, me.ID, small)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(small) {
		t.Error("the small creature card did not return to hand")
	}
	if !me.Graveyard.Contains(big) {
		t.Error("the big creature card left the graveyard")
	}
}

func TestB131DawnEvangelIgnoresACreatureWithNoAuraOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Dawn Evangel", "Enchantment Creature — Human Cleric", b131Evangel, false)
	small := b131Grave(me, "Small Bear", b131Bear, "{1}{G}", "", 2, 2)
	victim := effectstest.PushVanillaCreature(g, me.ID, "Victim", 2, 2)
	effectstest.P1Destroy(t, g, victim)
	g.RunStateChecksForTest()
	b131Settle(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Error("the trigger fired for a creature with no Aura attached")
	}
	if me.Hand.Contains(small) {
		t.Error("a creature card returned with no Aura attached")
	}
}

func TestB131RiseToGloryReturnsACreatureAndAnAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b131Grave(me, "Grave Bear", b131Bear, "{1}{G}", "", 2, 2)
	aura := b131Grave(me, "Arcane Flight", "Enchantment — Aura", "{1}{U}", b131ArcaneFlight, 0, 0)
	spec, _ := effects.Lookup(b131Glory)
	if spec.Modes == nil || len(spec.Modes.Options) != 2 || spec.Modes.Min != 1 || spec.Modes.Max != 2 {
		t.Fatalf("modes %+v, want choose one or both of two", spec.Modes)
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rise to Glory", TypeLine: "Sorcery", OracleID: b131Glory, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0, 1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: aura}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature card did not return to the battlefield")
	}
	if !g.Battlefield.Contains(aura) {
		t.Error("the Aura card did not return to the battlefield")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("the returned Aura should enchant the returned Bear: power %d, want 3", p)
	}
}

// b131ToFirstMain walks to the beginning of the precombat main phase and
// answers Klothys's target with the given card.
func b131ToFirstMain(t *testing.T, g *game.Game, me, target uuid.UUID) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if effectstest.LatestPickTarget(g, me) != nil {
			effectstest.PickCard(t, g, me, target)
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("Klothys's trigger never asked for a target")
}

func TestB131KlothysExilesALandForManaAndAnythingElseForLifeAndDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, me.ID, "Klothys, God of Destiny", "Legendary Enchantment Creature — God", b131Klothys, false)
	if c := effectstest.GaLayered(t, g, g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID); c.IsCreature() {
		t.Error("Klothys is a creature below seven devotion")
	}
	land := b131Grave(opp, "Their Forest", "Basic Land — Forest", "", "", 0, 0)
	spell := b131Grave(opp, "Their Bolt", "Instant", "{R}", "", 0, 0)
	b131ToFirstMain(t, g, me.ID, spell)
	life, oppLife := me.Life, opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(spell) {
		t.Error("the targeted card is still in the graveyard")
	}
	if me.Life != life+2 || opp.Life != oppLife-2 {
		t.Errorf("non-land: my life %d → %d, their life %d → %d, want +2 and -2", life, me.Life, oppLife, opp.Life)
	}
	if !opp.Graveyard.Contains(land) {
		t.Error("the other card left the graveyard")
	}
}

func TestB131KlothysLandAddsRedOrGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, me.ID, "Klothys, God of Destiny", "Legendary Enchantment Creature — God", b131Klothys, false)
	land := b131Grave(opp, "Their Forest", "Basic Land — Forest", "", "", 0, 0)
	b131ToFirstMain(t, g, me.ID, land)
	life, oppLife := me.Life, opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(land) {
		t.Error("the land card is still in the graveyard")
	}
	if me.Life != life || opp.Life != oppLife {
		t.Errorf("a land also drained: my life %d → %d, their life %d → %d", life, me.Life, oppLife, opp.Life)
	}
	if c := effectstest.ManaPickFor(g, me.ID); c != nil {
		if len(c.ColorOptions) != 2 {
			t.Errorf("colour options %v, want red and green", c.ColorOptions)
		}
		if err := g.ResolveManaChoice(c.ID, me.ID, c.ColorOptions[0]); err != nil {
			t.Fatalf("ResolveManaChoice: %v", err)
		}
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana in the pool %d, want 1", me.ManaPool.Total())
	}
}

func TestB131StormHeraldReturnsAurasOntoYourCreaturesAndExilesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := b131Grave(me, "Arcane Flight", "Enchantment — Aura", "{1}{U}", b131ArcaneFlight, 0, 0)
	filler := b131Grave(me, "Old Bolt", "Instant", "{R}", "", 0, 0)
	effectstest.B12PlayFromHand(t, g, "Storm Herald", "Creature — Human Shaman", b131Herald, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt for the Auras")
	}
	for _, c := range pick.ChooseCards {
		if c == filler {
			t.Error("a non-Aura card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, aura)
	effectstest.AnswerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		for _, c := range g.PendingChoices {
			t.Logf("pending: %v", c.Kind)
		}
		t.Fatal("the Aura did not return to the battlefield")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("enchanted Bear power %d, want 3", p)
	}
	// If the Aura would leave the battlefield, it is exiled instead.
	effectstest.P1Destroy(t, g, aura)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || me.Graveyard.Contains(aura) {
		t.Error("the Aura did not leave to exile")
	}
}

func TestB131StormHeraldExilesTheAurasAtYourNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := b131Grave(me, "Arcane Flight", "Enchantment — Aura", "{1}{U}", b131ArcaneFlight, 0, 0)
	effectstest.B12PlayFromHand(t, g, "Storm Herald", "Creature — Human Shaman", b131Herald, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, aura)
	effectstest.AnswerAuraHost(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not return")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || me.Graveyard.Contains(aura) {
		t.Error("the Aura was not exiled at the beginning of the end step")
	}
}

func TestB131StormHeraldChoosingNoneLeavesTheGraveyardAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := b131Grave(me, "Arcane Flight", "Enchantment — Aura", "{1}{U}", b131ArcaneFlight, 0, 0)
	effectstest.B12PlayFromHand(t, g, "Storm Herald", "Creature — Human Shaman", b131Herald, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(aura) {
		t.Error("an unchosen Aura left the graveyard")
	}
}

func TestB131UnderworldSentinelExilesOnAttackAndReturnsEverythingWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	sentinel := effectstest.PushCatalogPermanent(g, me.ID, "Underworld Sentinel", "Creature — Skeleton Soldier", b131Sentinel, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == sentinel {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 5
			}
		}
	})
	one := b131Grave(me, "First Bear", b131Bear, "{1}{G}", "", 2, 2)
	two := b131Grave(me, "Second Bear", b131Bear, "{2}{G}", "", 3, 3)
	spell := b131Grave(me, "Old Bolt", "Instant", "{R}", "", 0, 0)
	effectstest.AttackWithTrigger(t, g, sentinel, opp.ID, one)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(one) {
		t.Fatal("the targeted creature card is still in the graveyard")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("an instant was exiled")
	}
	_ = two
	effectstest.P1Destroy(t, g, sentinel)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	// A card returned from exile is a new object with a new ID.
	if n := effectstest.NamedOnBattlefield(g, "First Bear", me.ID); n != 1 {
		t.Error("the exiled creature card did not return when the Sentinel died")
	}
	if !me.Graveyard.Contains(two) {
		t.Error("a card the Sentinel never exiled left the graveyard")
	}
}

func TestB131StormHeraldNeverEnchantsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", b131Bear, 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	aura := b131Grave(me, "Arcane Flight", "Enchantment — Aura", "{1}{U}", b131ArcaneFlight, 0, 0)
	effectstest.B12PlayFromHand(t, g, "Storm Herald", "Creature — Human Shaman", b131Herald, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, aura)
	c := effectstest.AuraHostChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no aura_host prompt with two creatures of mine to enchant")
	}
	var sawMine bool
	for _, id := range c.PickTargetCards {
		if id == theirs {
			t.Error("an opponent's creature was offered as a host")
		}
		if id == mine {
			sawMine = true
		}
	}
	if !sawMine {
		t.Error("my own creature was not offered as a host")
	}
	if len(c.PickTargetPlayers) != 0 {
		t.Errorf("players offered as hosts: %v", c.PickTargetPlayers)
	}
}
