package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b412_test.go — Gatecrash, batch 412 (#5565): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB412Drake    = "1987bb61-a31b-4d5a-910f-88e9cbc6c8b3"
	oldB412Burst    = "1ab592fa-adfb-4879-af6c-0204eab323fb"
	oldB412Raiders  = "7904dd02-f231-4538-b1bb-81bce69309b1"
	oldB412Zameck   = "98250fad-7fd0-431e-a87e-551185d26ec1"
	oldB412Fire     = "b1e5c507-d110-4f9e-a58a-939f9961e464"
	oldB412Swimmer  = "ce02ef24-d53e-4025-bc45-ca27aba89fe9"
	oldB412PlusOne  = game.CounterPlusOne
	oldB412Blaze    = "blaze"
	oldB412PTFormat = "%s is %d/%d, want %d/%d"
)

func TestOldB412CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB412Drake:   "Sapphire Drake",
		oldB412Burst:   "Burst of Strength",
		oldB412Raiders: "Rubblebelt Raiders",
		oldB412Zameck:  "Zameck Guildmage",
		oldB412Fire:    "Five-Alarm Fire",
		oldB412Swimmer: "Nimbus Swimmer",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func oldB412Counter(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, oldB412PlusOne, n) })
}

// --- Sapphire Drake ---

func TestOldB412SapphireDrakeGivesFlyingToYourCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Sapphire Drake", OracleID: oldB412Drake, TypeLine: "Creature — Drake", Power: 4, Toughness: 4, Keywords: []string{"flying"}})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	oldB412Counter(g, bear, 1)
	oldB412Counter(g, theirs, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("my creature with a +1/+1 counter has no flying")
	}
	for name, id := range map[string]uuid.UUID{"a creature without a counter": plain, "an opponent's creature with a counter": theirs} {
		if effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
			t.Errorf("%s has flying", name)
		}
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, oldB412PlusOne, -1) })
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("flying stayed after the counter left")
	}
}

// --- Burst of Strength ---

func TestOldB412BurstOfStrengthPutsACounterAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardByID(g, bear).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Burst of Strength", "Instant", oldB412Burst, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, oldB412PlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
	if effectstest.B16Tapped(t, g, bear) {
		t.Error("the creature is still tapped")
	}
	if p, tg := effectstest.E2Card(t, g, bear).CurrentPower(), effectstest.E2Card(t, g, bear).CurrentToughness(); p != 3 || tg != 3 {
		t.Errorf(oldB412PTFormat, "Bear", p, tg, 3, 3)
	}
}

// --- Rubblebelt Raiders ---

func TestOldB412RubblebeltRaidersCountsEveryAttackerYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raiders := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Rubblebelt Raiders", OracleID: oldB412Raiders, TypeLine: "Creature — Human Warrior", Power: 3, Toughness: 3})
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 1, 1)
	effectstest.AttackWith(t, g, opp.ID, raiders, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, raiders, oldB412PlusOne); n != 3 {
		t.Errorf("%d counters, want 3 (three attackers)", n)
	}
}

func TestOldB412RubblebeltRaidersAloneGetsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raiders := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Rubblebelt Raiders", OracleID: oldB412Raiders, TypeLine: "Creature — Human Warrior", Power: 3, Toughness: 3})
	effectstest.AttackWith(t, g, opp.ID, raiders)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, raiders, oldB412PlusOne); n != 1 {
		t.Errorf("%d counters, want 1", n)
	}
}

// --- Zameck Guildmage ---

func TestOldB412ZameckGuildmageGivesCreaturesACounterThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	zg := effectstest.PushCatalogPermanent(g, me.ID, "Zameck Guildmage", "Creature — Elf Wizard", oldB412Zameck, false)
	for i := 0; i < 2; i++ {
		effectstest.Acrb1AddMana(t, g, me.ID, "{G}{U}")
		if err := g.ActivateCatalogAbility(me.ID, zg, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.CastAndResolveCreature(t, g, "Hill Giant", "Creature — Giant", "")
	effectstest.PassPriorityAroundTable(t, g)
	giant := effectstest.FindBattlefieldByName(g, "Hill Giant")
	if n := effectstest.CountersOn(g, giant, oldB412PlusOne); n != 2 {
		t.Errorf("a creature cast after two activations entered with %d counters, want 2", n)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.CastAndResolveCreature(t, g, "Next Giant", "Creature — Giant", "")
	if n := effectstest.CountersOn(g, effectstest.FindBattlefieldByName(g, "Next Giant"), oldB412PlusOne); n != 0 {
		t.Errorf("a creature entered with %d counters after the turn ended, want 0", n)
	}
}

func TestOldB412ZameckGuildmageDrawsByRemovingACounterFromAnyOfYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	zg := effectstest.PushCatalogPermanent(g, me.ID, "Zameck Guildmage", "Creature — Elf Wizard", oldB412Zameck, false)
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{U}")
	if err := g.ActivateCatalogAbility(me.ID, zg, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{zg}, CounterCounts: []int{1}}); err == nil {
		t.Fatal("activated with no +1/+1 counter on a creature to remove")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB412Counter(g, bear, 1)
	effectstest.SeedLibraryTop(me, "Top", "Sorcery", "")
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, zg, 1, game.ActivateAbilityParams{Strict: true, CounterSourceIDs: []uuid.UUID{bear}, CounterCounts: []int{1}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, oldB412PlusOne); n != 0 {
		t.Errorf("the Bear still has %d counters", n)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand is %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- Five-Alarm Fire ---

func TestOldB412FiveAlarmFireCountsEachCreatureThatDealsCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fire := effectstest.B12Push(g, me.ID, "Five-Alarm Fire", "Enchantment", oldB412Fire, 0, 0)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, a, opp.ID, 2)
	effectstest.DealCombatDamageToPlayer(g, a, opp.ID, 1) // a second recipient of the same creature's damage
	effectstest.DealCombatDamageToPlayer(g, b, opp.ID, 2)
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, fire, oldB412Blaze); n != 2 {
		t.Errorf("%d blaze counters, want 2 (one per creature of mine that dealt combat damage)", n)
	}
}

func TestOldB412FiveAlarmFireNeedsFiveCountersToBurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fire := effectstest.B12Push(g, me.ID, "Five-Alarm Fire", "Enchantment", oldB412Fire, 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(fire, oldB412Blaze, 4) })
	if err := g.ActivateCatalogAbility(me.ID, fire, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err == nil {
		t.Fatal("activated with only four blaze counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(fire, oldB412Blaze, 2) })
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, fire, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("life %d → %d, want 5 damage", life, opp.Life)
	}
	if n := effectstest.CountersOn(g, fire, oldB412Blaze); n != 1 {
		t.Errorf("%d blaze counters left, want 1 (six minus five)", n)
	}
}

// --- Nimbus Swimmer ---

func TestOldB412NimbusSwimmerEntersWithXCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Nimbus Swimmer", OracleID: oldB412Swimmer, TypeLine: "Creature — Leviathan",
		Power: 0, Toughness: 0, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, oldB412PlusOne); n != 3 {
		t.Errorf("%d counters, want 3", n)
	}
	if p, tg := effectstest.E2Card(t, g, id).CurrentPower(), effectstest.E2Card(t, g, id).CurrentToughness(); p != 3 || tg != 3 {
		t.Errorf(oldB412PTFormat, "Swimmer", p, tg, 3, 3)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("no flying")
	}
}
