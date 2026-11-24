package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 old sets, batch 008 group A (#5462): Bow of Nylea, Braid of Fire,
// Olivia, Mobilized for War and Sarulf, Realm Eater.

const (
	oldB008Bow    = "361b965f-2ce7-49f8-84e7-7325ea0c948d"
	oldB008Braid  = "796009b3-73b5-42c0-a490-66f12be5a42e"
	oldB008Olivia = "1f8d387e-e4bb-4e76-a173-f46f7b107b95"
	oldB008Sarulf = "789964f5-79c5-4329-b68f-b15b0d54b0b2"
)

func TestS82OldB008ACardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB008Bow: "Bow of Nylea", oldB008Braid: "Braid of Fire",
		oldB008Olivia: "Olivia, Mobilized for War", oldB008Sarulf: "Sarulf, Realm Eater",
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

// --- Braid of Fire ---------------------------------------------------------------

func TestS82OldB008BraidAddsRedManaPerAgeCounterEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	braid := effectstest.PushPermanentForTest(g, owner.ID, "Braid of Fire", oldB008Braid, "Enchantment")
	for turn := 1; turn <= 2; turn++ {
		if turn > 1 {
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.AdvanceToUpkeepOf(t, g, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.AgeCountersOn(t, g, braid); got != turn {
			t.Fatalf("turn %d: %d age counters, want %d", turn, got, turn)
		}
		before := owner.ManaPool.Total()
		effectstest.AnswerPayUnless(t, g, owner.ID, true)
		if got := owner.ManaPool.Total() - before; got != turn {
			t.Errorf("turn %d: %d mana added, want %d", turn, got, turn)
		}
		if _, ok := effectstest.BattlefieldCard(g, braid); !ok {
			t.Fatalf("turn %d: paying sacrificed the Braid", turn)
		}
	}
}

func TestS82OldB008BraidIsSacrificedWhenTheUpkeepIsDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	braid := effectstest.PushPermanentForTest(g, owner.ID, "Braid of Fire", oldB008Braid, "Enchantment")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	before := owner.ManaPool.Total()
	effectstest.AnswerPayUnless(t, g, owner.ID, false)
	if _, ok := effectstest.BattlefieldCard(g, braid); ok {
		t.Error("declining the cumulative upkeep left the Braid on the battlefield")
	}
	if owner.ManaPool.Total() != before {
		t.Error("declining added mana")
	}
}

// --- Bow of Nylea ----------------------------------------------------------------

func bowOfNylea(t *testing.T, g *game.Game, owner *game.Player) uuid.UUID {
	t.Helper()
	id := effectstest.PushPermanentForTest(g, owner.ID, "Bow of Nylea", oldB008Bow, "Legendary Enchantment Artifact")
	owner.ManaPool.AddMana(game.ManaToken{Color: "G"})
	owner.ManaPool.AddMana(game.ManaToken{Color: "C"})
	return id
}

func TestS82OldB008BowGivesAttackingCreaturesDeathtouchOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bowOfNylea(t, g, me)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	home := effectstest.PushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, attacker), "deathtouch") {
		t.Error("an attacking creature I control lacks deathtouch")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, home), "deathtouch") {
		t.Error("a creature that did not attack has deathtouch")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "deathtouch") {
		t.Error("an opponent's creature has deathtouch")
	}
}

func TestS82OldB008BowCounterMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bow := bowOfNylea(t, g, me)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{0}, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	if c, _ := effectstest.BattlefieldCard(g, bow); !c.Tapped {
		t.Error("the Bow was not tapped")
	}
}

func TestS82OldB008BowDamageModeNeedsAFlyingTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bow := bowOfNylea(t, g, me)
	ground := effectstest.PushVanillaCreature(g, opp.ID, "Ground", 2, 5)
	flier := effectstest.B12Push(g, opp.ID, "Flier", "Creature — Bird", "", 1, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}, Targets: effectstest.CardTarget(ground)}); err == nil {
		t.Fatal("the damage mode accepted a creature without flying")
	}
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{1}, Targets: effectstest.CardTarget(flier)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, flier)
	if c.DamageMarked != 2 {
		t.Errorf("flier has %d damage, want 2", c.DamageMarked)
	}
}

func TestS82OldB008BowLifeMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bow := bowOfNylea(t, g, me)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{2}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestS82OldB008BowTucksUpToFourGraveyardCardsOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bow := bowOfNylea(t, g, me)
	var ids []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four"} {
		ids = append(ids, effectstest.PushGraveyardCardForTest(me, n))
	}
	left := effectstest.PushGraveyardCardForTest(me, "Left Behind")
	effectstest.AdvanceToMain(t, g)
	size := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{3}, Targets: effectstest.CardTargets(ids...)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil {
			t.Logf("pending choice %v (%s)", c.Kind, c.Chooser)
		}
	}
	if got := me.Library.Size() - size; got != 4 {
		t.Errorf("library grew by %d, want 4", got)
	}
	for _, id := range ids {
		if me.Graveyard.Contains(id) {
			t.Errorf("a targeted card is still in the graveyard")
		}
	}
	if !me.Graveyard.Contains(left) {
		t.Error("an untargeted card left the graveyard")
	}
}

func TestS82OldB008BowRefusesFiveGraveyardTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bow := bowOfNylea(t, g, me)
	var ids []uuid.UUID
	for _, n := range []string{"One", "Two", "Three", "Four", "Five"} {
		ids = append(ids, effectstest.PushGraveyardCardForTest(me, n))
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{3}, Targets: effectstest.CardTargets(ids...)}); err == nil {
		t.Error("five targets were accepted for up to four")
	}
}

// --- Olivia, Mobilized for War ----------------------------------------------------

func enterCreatureFromHand(t *testing.T, g *game.Game, p *game.Player, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestS82OldB008OliviaDiscardBoostsTheEnteringCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Olivia, Mobilized for War", "Legendary Creature — Vampire Knight", oldB008Olivia, 3, 3)
	fodder := effectstest.PushGraveyardCardForTest(me, "unused")
	_ = fodder
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Discard Me", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	bear := enterCreatureFromHand(t, g, me, "Bear")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := hand - me.Hand.Size(); got != 1 {
		t.Errorf("hand fell by %d, want 1 (the discard)", got)
	}
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters on the Bear, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear has no haste")
	}
	c, _ := effectstest.BattlefieldCard(g, bear)
	if !c.HasSubtype("Vampire") {
		t.Errorf("the Bear is %q, want a Vampire as well", c.TypeLine)
	}
}

func TestS82OldB008OliviaDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Olivia, Mobilized for War", "Legendary Creature — Vampire Knight", oldB008Olivia, 3, 3)
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Keep Me", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	bear := enterCreatureFromHand(t, g, me, "Bear")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if me.Hand.Size() != hand {
		t.Error("declining still discarded")
	}
	if n := effectstest.CounterCount(g, bear, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters after declining", n)
	}
}

func TestS82OldB008OliviaIgnoresItselfAndOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.B12Push(g, me.ID, "Olivia, Mobilized for War", "Legendary Creature — Vampire Knight", oldB008Olivia, 3, 3)
	enterCreatureFromHand(t, g, opp, "Their Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("an opponent's creature entering asked me to discard")
	}
}

// --- Sarulf, Realm Eater ---------------------------------------------------------

func TestS82OldB008SarulfGrowsWhenAnOpponentsPermanentGoesToAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sarulf := effectstest.B12Push(g, me.ID, "Sarulf, Realm Eater", "Legendary Creature — Wolf", oldB008Sarulf, 3, 3)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sarulf, game.CounterPlusOne); n != 0 {
		t.Fatalf("my own creature dying gave %d counters", n)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sarulf, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters after an opponent's permanent died, want 1", n)
	}
}

func sarulfWithCounters(g *game.Game, owner uuid.UUID, n int) uuid.UUID {
	id := effectstest.B12Push(g, owner, "Sarulf, Realm Eater", "Legendary Creature — Wolf", oldB008Sarulf, 3, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, n) })
	return id
}

func TestS82OldB008SarulfUpkeepExilesEachOtherNonlandPermanentUpToTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	sarulf := sarulfWithCounters(g, me.ID, 2)
	cheap := effectstest.B12Push(g, opp.ID, "Cheap", "Artifact", "", 0, 0)
	pricey := effectstest.B12Push(g, opp.ID, "Pricey", "Artifact", "", 0, 0)
	land := effectstest.PushPermanentForTest(g, opp.ID, "Forest", "", "Basic Land — Forest")
	setManaValue(g, cheap, "{1}{1}")
	setManaValue(g, pricey, "{3}")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptional(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, sarulf); !ok {
		t.Fatal("Sarulf exiled itself")
	}
	if n := effectstest.CounterCount(g, sarulf, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters left, want all removed", n)
	}
	if _, ok := effectstest.BattlefieldCard(g, cheap); ok {
		t.Error("the mana value 2 permanent survived")
	}
	if _, ok := effectstest.BattlefieldCard(g, pricey); !ok {
		t.Error("the mana value 3 permanent was exiled with only 2 counters removed")
	}
	if _, ok := effectstest.BattlefieldCard(g, land); !ok {
		t.Error("a land was exiled")
	}
}

func TestS82OldB008SarulfUpkeepDeclinedKeepsTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	sarulf := sarulfWithCounters(g, me.ID, 2)
	other := effectstest.B12Push(g, opp.ID, "Cheap", "Artifact", "", 0, 0)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptional(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, sarulf, game.CounterPlusOne); n != 2 {
		t.Errorf("%d counters, want 2 after declining", n)
	}
	if _, ok := effectstest.BattlefieldCard(g, other); !ok {
		t.Error("declining still exiled")
	}
}

func TestS82OldB008SarulfWithoutCountersDoesNotTriggerAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	sarulf := effectstest.B12Push(g, me.ID, "Sarulf, Realm Eater", "Legendary Creature — Wolf", oldB008Sarulf, 3, 3)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.TriggerOnStack(g, sarulf) != nil {
		t.Error("the upkeep trigger went on the stack with no +1/+1 counters")
	}
}

// setManaValue gives a battlefield card a printed mana cost.
func setManaValue(g *game.Game, id uuid.UUID, cost string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ManaCost = cost
			}
		}
	})
}
