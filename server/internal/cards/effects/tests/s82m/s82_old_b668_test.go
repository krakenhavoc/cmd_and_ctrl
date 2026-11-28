package testss82m

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 668: Prophecy, activated abilities.

const (
	oldB668MercInformer = "245e26ec-39eb-4af8-99af-8cc22edbc261"
	oldB668Greel        = "2b84a10e-4c5c-407f-8b5b-a23d402f0529"
	oldB668Alexi        = "3f60de36-ed63-4d08-a012-fc16e91da46d"
	oldB668Healer       = "4377cc84-19d3-48e0-9588-d18b4fbf4e6d"
	oldB668Battlewagon  = "4a2df44b-5d36-4c4e-b982-836d9d7e0505"
	oldB668Eagle        = "640b67d8-75a4-44d4-a2f5-7323bad4e1a8"
	oldB668Idol         = "78b364b7-d180-484c-a1d9-9d7a4e8abb57"
	oldB668Jolrael      = "7972d118-c105-4b33-8c3b-ec85da538471"
	oldB668RebelInform  = "85ef10fd-4a48-4a37-baad-f663fb5439e0"
	oldB668Jeweled      = "cddf2e46-79dc-4d6f-bb67-db55d966e8e2"
)

func oldB668Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

func TestOldB668Registered(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB668MercInformer: "Mercenary Informer", oldB668Greel: "Greel, Mind Raker", oldB668Alexi: "Alexi, Zephyr Mage",
		oldB668Healer: "Troubled Healer", oldB668Battlewagon: "Keldon Battlewagon", oldB668Eagle: "Stormwatch Eagle",
		oldB668Idol: "Chimeric Idol", oldB668Jolrael: "Jolrael, Empress of Beasts", oldB668RebelInform: "Rebel Informer",
		oldB668Jeweled: "Jeweled Spirit",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// A colored spell that targets a creature, registered for the test.
func oldB668TargetingSpell(t *testing.T, g *game.Game, p *game.Player, oracle, color string) uuid.UUID {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{OracleID: oracle, Name: oracle, Targets: effects.TargetCreature("target creature"),
		OnResolve: func(*game.StackItem, *effects.Context) error { return nil }})
	id := effectstest.HandCardFull(p, oracle, "Instant", "{"+color+"}", oracle, []string{color})
	g.WithWriteLock(func() { effectstest.FillPoolColored(p, color, 1) })
	return id
}

func oldB668Informer(t *testing.T, name, oracle, typeLine, barredColor, mana, tribe string) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	informer := effectstest.B12Push(g, me.ID, name, typeLine, oracle, 2, 1)
	barred := oldB668TargetingSpell(t, g, me, "test-b668-"+strings.ToLower(barredColor), barredColor)
	if err := g.CastSpell(me.ID, barred, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: informer}}}); err == nil {
		t.Errorf("a %s spell targeted %s", barredColor, name)
	}
	other := "R"
	if barredColor == "R" {
		other = "G"
	}
	fine := oldB668TargetingSpell(t, g, me, "test-b668-"+strings.ToLower(other), other)
	if err := g.CastSpell(me.ID, fine, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: informer}}}); err != nil {
		t.Errorf("a %s spell may target it: %v", other, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The tuck: a nontoken member of the tribe goes to the bottom of its owner's library; a token can't be targeted.
	theirs := effectstest.B12Creature(g, opp.ID, "Their "+tribe, "Creature — "+tribe, 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: tribe + " Token", TypeLine: "Token Creature — " + tribe,
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatMana(t, g, me, mana)
	if err := effectstest.ActivateOn(t, g, me.ID, informer, 0, token); err == nil {
		t.Errorf("a token %s was targeted", tribe)
	}
	if err := effectstest.ActivateOn(t, g, me.ID, informer, 0, theirs); err != nil {
		t.Fatalf("the tuck: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Library.Contains(theirs) {
		t.Fatalf("the %s should be in its owner's library", tribe)
	}
	if bottom, err := opp.Library.Bottom(); err != nil || bottom.InstanceID != theirs {
		t.Errorf("the card should be on the bottom of the library")
	}
}

func TestOldB668MercenaryInformer(t *testing.T) {
	oldB668Informer(t, "Mercenary Informer", oldB668MercInformer, "Creature — Human Rebel Mercenary", "B", "{W}{W}{W}", "Mercenary")
}

func TestOldB668RebelInformer(t *testing.T) {
	oldB668Informer(t, "Rebel Informer", oldB668RebelInform, "Creature — Human Mercenary Rebel", "W", "{B}{B}{B}", "Rebel")
}

// Greel: pays {X}{B}, taps, discards two; the target discards X at random.
func TestOldB668Greel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	greel := effectstest.B12Push(g, me.ID, "Greel, Mind Raker", "Legendary Creature — Horror Spellshaper", oldB668Greel, 3, 3)
	d1 := effectstest.SpmB12Hand(me, "One", "Instant")
	d2 := effectstest.SpmB12Hand(me, "Two", "Instant")
	for i := 0; i < 4; i++ {
		effectstest.SpmB12Hand(opp, "Theirs", "Instant")
	}
	before := opp.Hand.Size()
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, greel, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(d1) || !me.Graveyard.Contains(d2) {
		t.Error("the two cards were not discarded as a cost")
	}
	if got := before - opp.Hand.Size(); got != 2 {
		t.Errorf("the target discarded %d cards, want X = 2", got)
	}
	if !oldB668Tapped(t, g, greel) {
		t.Error("Greel should be tapped")
	}
}

// Alexi: X target creatures return to their owners' hands.
func TestOldB668Alexi(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	alexi := effectstest.B12Push(g, me.ID, "Alexi, Zephyr Mage", "Legendary Creature — Human Spellshaper", oldB668Alexi, 3, 3)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 2)
	d1 := effectstest.SpmB12Hand(me, "One", "Instant")
	d2 := effectstest.SpmB12Hand(me, "Two", "Instant")
	effectstest.FloatMana(t, g, me, "{U}{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, alexi, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) || !opp.Hand.Contains(a) || !opp.Hand.Contains(b) {
		t.Error("the two targets should be back in their owner's hand")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("an untargeted creature stays")
	}
}

// Troubled Healer: sacrificing a land prevents the next 2 damage to the target.
func TestOldB668TroubledHealer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	healer := effectstest.B12Push(g, me.ID, "Troubled Healer", "Creature — Human Cleric", oldB668Healer, 1, 2)
	land := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("activated without naming a land to sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, healer, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land should have been sacrificed")
	}
	life := me.Life
	effectstest.BoltFrom(t, g, opp, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 1 {
		t.Errorf("a Bolt dealt %d, want 1 (2 of 3 prevented)", got)
	}
}

// Keldon Battlewagon: can't block; pumps by the tapped creature's power; if it
// attacks it is sacrificed at end of combat.
func TestOldB668KeldonBattlewagon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wagon := effectstest.B12Push(g, me.ID, "Keldon Battlewagon", "Artifact Creature — Juggernaut", oldB668Battlewagon, 0, 3)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 4, 4)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.ActivateCatalogAbility(me.ID, wagon, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{ally}}); err != nil {
		t.Fatalf("the pump: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, wagon); p != 4 {
		t.Fatalf("power %d, want +4 for the 4-power creature tapped", p)
	}
	if err := g.DeclareAttacker(wagon, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wagon) {
		t.Error("the Battlewagon should be sacrificed at end of combat")
	}
}

func TestOldB668KeldonBattlewagonCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wagon := effectstest.B12Push(g, me.ID, "Keldon Battlewagon", "Artifact Creature — Juggernaut", oldB668Battlewagon, 0, 3)
	atk := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(atk, me.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wagon, atk); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("the Battlewagon blocked: err = %v, want ErrIllegalBlock", err)
	}
}

// Stormwatch Eagle: sacrificing a land returns it to hand.
func TestOldB668StormwatchEagle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	eagle := effectstest.B12Push(g, me.ID, "Stormwatch Eagle", "Creature — Bird", oldB668Eagle, 2, 1)
	land := effectstest.FreeSpellPermanent(g, me.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, eagle, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(eagle) || !me.Hand.Contains(eagle) {
		t.Error("the Eagle should be back in its owner's hand")
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land should have been sacrificed")
	}
}

// Chimeric Idol: taps all your lands and becomes a 3/3 Turtle artifact creature.
func TestOldB668ChimericIdol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	idol := effectstest.PushCatalogPermanent(g, me.ID, "Chimeric Idol", "Artifact", oldB668Idol, false)
	mine := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	theirs := effectstest.FreeSpellPermanent(g, opp.ID, "Island", "Basic Land — Island")
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB668Tapped(t, g, mine) || oldB668Tapped(t, g, theirs) {
		t.Error("only my lands should be tapped")
	}
	p, th := effectstest.EffectivePTOf(t, g, idol)
	var typeLine string
	g.ReadSnapshot(func() { typeLine = effectstest.FindBattlefieldCardForTest(g, idol).TypeLine })
	if p != 3 || th != 3 || !strings.Contains(typeLine, "Artifact") {
		t.Errorf("idol is %d/%d %q, want a 3/3 artifact creature", p, th, typeLine)
	}
	var turtle bool
	g.ReadSnapshot(func() { turtle = effectstest.FindBattlefieldCardForTest(g, idol).HasSubtypeWord("Turtle") })
	if !turtle {
		t.Error("the idol should be a Turtle")
	}
}

// Jolrael: the target player's lands become 3/3 creatures; lands that arrive later don't.
func TestOldB668Jolrael(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jolrael := effectstest.B12Push(g, me.ID, "Jolrael, Empress of Beasts", "Legendary Creature — Human Spellshaper", oldB668Jolrael, 3, 3)
	theirs := effectstest.FreeSpellPermanent(g, opp.ID, "Island", "Basic Land — Island")
	mine := effectstest.FreeSpellPermanent(g, me.ID, "Forest", "Basic Land — Forest")
	d1 := effectstest.SpmB12Hand(me, "One", "Instant")
	d2 := effectstest.SpmB12Hand(me, "Two", "Instant")
	effectstest.FloatMana(t, g, me, "{G}{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, jolrael, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{d1, d2},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var creature, land bool
	var p, th int
	g.ReadSnapshot(func() {
		c := effectstest.FindBattlefieldCardForTest(g, theirs)
		creature, land, p, th = c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness()
	})
	if !creature || !land || p != 3 || th != 3 {
		t.Errorf("their Island: creature=%v land=%v %d/%d, want a 3/3 land creature", creature, land, p, th)
	}
	var mineCreature bool
	g.ReadSnapshot(func() { mineCreature = effectstest.FindBattlefieldCardForTest(g, mine).IsCreature() })
	if mineCreature {
		t.Error("my own land is not animated")
	}
	later := effectstest.FreeSpellPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	var laterCreature bool
	g.ReadSnapshot(func() { laterCreature = effectstest.FindBattlefieldCardForTest(g, later).IsCreature() })
	if laterCreature {
		t.Error("a land that enters later is not animated")
	}
}

// Jeweled Spirit: two lands sacrificed; protection from artifacts, or from a colour chosen as it resolves.
func TestOldB668JeweledSpiritArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spirit := effectstest.B12Push(g, me.ID, "Jeweled Spirit", "Creature — Spirit", oldB668Jeweled, 3, 3)
	l1 := effectstest.FreeSpellPermanent(g, me.ID, "Plains", "Basic Land — Plains")
	l2 := effectstest.FreeSpellPermanent(g, me.ID, "Plains", "Basic Land — Plains")
	if err := g.ActivateCatalogAbility(me.ID, spirit, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{l1}}); err == nil {
		t.Fatal("activated with only one land sacrificed")
	}
	if err := g.ActivateCatalogAbility(me.ID, spirit, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{l1, l2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, spirit, "protection from artifacts")
}

func TestOldB668JeweledSpiritColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spirit := effectstest.B12Push(g, me.ID, "Jeweled Spirit", "Creature — Spirit", oldB668Jeweled, 3, 3)
	l1 := effectstest.FreeSpellPermanent(g, me.ID, "Plains", "Basic Land — Plains")
	l2 := effectstest.FreeSpellPermanent(g, me.ID, "Plains", "Basic Land — Plains")
	if err := g.ActivateCatalogAbility(me.ID, spirit, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{l1, l2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, spirit, game.ProtectionFromColor("R"))
}
