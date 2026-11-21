package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b66_test.go — Fallout creatures and enchantments (GAP batch 66,
// #5948), played through the real catalog entries.

const (
	gapB66GaryClone   = "d9af0da0-5e10-40c5-afb2-3f44c016b927"
	gapB66Sierra      = "e4fb15b6-02bb-4923-953a-69784ca780b8"
	gapB66YesMan      = "16b8eab0-ccbd-4dbf-9998-82a4d85d0f77"
	gapB66AnimalFrnd  = "ae4249de-1b8c-4b15-8960-180557a15065"
	gapB66StrongBack  = "097992a9-7b88-4c3f-8191-1ff10a4a735d"
	gapB66WellRested  = "f8cebb33-9235-4f63-8171-15af8b8e1436"
	gapB66Vault87     = "1eb22223-edd4-42bf-a127-e16191cbd1c1"
	gapB66Sprint      = "99a6d7fe-40a1-4f2f-a3f6-6777884d5d1a"
	gapB66Vault21     = "8d83d816-518c-4557-a3c4-0255f772c57a"
	gapB66GoblinPaint = "67ca79d7-9064-4605-8625-b1cfe5cb1b45"
)

func TestGapB66CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB66GaryClone: "Gary Clone", gapB66Sierra: "Sierra, Nuka's Biggest Fan", gapB66YesMan: "Yes Man, Personal Securitron",
		gapB66AnimalFrnd: "Animal Friend", gapB66StrongBack: "Strong Back", gapB66WellRested: "Well Rested",
		gapB66Vault87: "Vault 87: Forced Evolution", gapB66Sprint: "Grim Reaper's Sprint", gapB66Vault21: "Vault 21: House Gambit",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// castAura casts the catalog Aura at host, free (no mana cost on the test card).
func gapB66Aura(t *testing.T, g *game.Game, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := castCatalogSpell(t, g, name, "Enchantment — Aura", oracle, []game.TargetRef{{Kind: game.TargetCard, ID: host}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatalf("%s did not enter", name)
	}
	return aura
}

// --- Gary Clone -----------------------------------------------------------

func TestGapB66GaryCloneSquadCopiesAndEveryCloneGetsPlusOneWhenOneAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	id := tmtB11Hand(g, me, "Gary Clone", "Creature — Human Citizen", gapB66GaryClone, "{1}{W}", 1, 3)
	floatMana(t, g, me, "{W}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with squad paid once: %v", err)
	}
	settleOrderingTriggers(t, g)
	if got := len(tmtB1Named(g, "Gary Clone")); got != 2 {
		t.Fatalf("%d Gary Clones after squad, want 2", got)
	}
	// An older Clone and an unrelated creature on the board; the older
	// Clone is the attacker.
	attacker := b12Push(g, me.ID, "Gary Clone", "Creature — Human Citizen", gapB66GaryClone, 1, 3)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, c := range tmtB1Named(g, "Gary Clone") {
		if got := effectivePower(t, g, c); got != 2 {
			t.Errorf("a Gary Clone has power %d, want 2", got)
		}
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("an unnamed creature has power %d, want its own 2", got)
	}
}

// --- Sierra, Nuka's Biggest Fan -------------------------------------------

func TestGapB66SierraGrowsAndMakesFoodOncePerDamageAndPumpsWhenAFoodIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sierra := b27Push(g, me.ID, "Sierra, Nuka's Biggest Fan", "Legendary Creature — Human Citizen", gapB66Sierra, "{3}{W}", 3, 4, "W")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dealCombatDamageToPlayer(g, sierra, opp.ID, 3)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, sierra, "quest"); got != 1 {
		t.Fatalf("%d quest counters after combat damage, want 1", got)
	}
	foods := tmtB1Named(g, "Food")
	if len(foods) != 1 {
		t.Fatalf("%d Foods, want 1", len(foods))
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, foods[0], 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("crack the Food: %v", err)
	}
	b27Settle(t, g, me.ID, bear)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the target is %d/%d, want 3/3 (+1/+1 for one quest counter)", p, tt)
	}
}

// --- Yes Man, Personal Securitron -----------------------------------------

func TestGapB66YesManGivesAwayItselfDrawsTwoAndPaysSoldiersToItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	seedLibrary(me, "A", "B", "C", "D")
	yes := pushCatalogPermanent(g, me.ID, "Yes Man, Personal Securitron", "Legendary Artifact Creature — Robot", gapB66YesMan, false)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, yes, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, yes); c.Controller != opp.ID {
		t.Fatalf("Yes Man is controlled by %v, want the opponent", c.Controller)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
	if got := countersOn(g, yes, "quest"); got != 1 {
		t.Errorf("%d quest counters, want 1", got)
	}
	b27Kill(g, yes)
	passPriorityAroundTable(t, g)
	soldiers := tmtB1Named(g, "Soldier")
	if len(soldiers) != 1 {
		t.Fatalf("%d Soldiers, want 1", len(soldiers))
	}
	if c := layeredCard(t, g, soldiers[0]); c.Controller != me.ID || !c.Tapped {
		t.Errorf("the Soldier is controlled by %v, tapped %v; want my tapped token (the owner's)", c.Controller, c.Tapped)
	}
}

// --- Animal Friend --------------------------------------------------------

func TestGapB66AnimalFriendMakesASquirrelWithACounterForEachOtherAttachment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	gapB66Aura(t, g, "Animal Friend", gapB66AnimalFrnd, bear)
	gapB66Aura(t, g, "Goblin War Paint", gapB66GoblinPaint, bear)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	squirrels := tmtB1Named(g, "Squirrel")
	if len(squirrels) != 1 {
		t.Fatalf("%d Squirrels, want 1", len(squirrels))
	}
	if got := plusOnes(g, squirrels[0]); got != 1 {
		t.Errorf("the Squirrel has %d +1/+1 counters, want 1 (War Paint; Animal Friend itself is not counted)", got)
	}
}

// --- Well Rested ----------------------------------------------------------

func TestGapB66WellRestedPaysOffOnlyTheFirstUntapEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	seedLibrary(me, "A", "B", "C", "D")
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	gapB66Aura(t, g, "Well Rested", gapB66WellRested, bear)
	life, hand := me.Life, me.Hand.Size()
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bear) })
		passPriorityAroundTable(t, g)
	}
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("%d +1/+1 counters after two untaps in one turn, want 2 (once each turn)", got)
	}
	if me.Life != life+2 || me.Hand.Size() != hand+1 {
		t.Errorf("life %+d, cards %+d; want +2 and +1", me.Life-life, me.Hand.Size()-hand)
	}
}

// --- Strong Back ----------------------------------------------------------

func TestGapB66StrongBackPumpsPerAttachmentAndMakesAurasOnItsHostCostThreeLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	host := b12Push(g, me.ID, "Host", "Creature — Bear", "", 2, 2)
	other := b12Push(g, me.ID, "Other", "Creature — Bear", "", 2, 2)
	gapB66Aura(t, g, "Strong Back", gapB66StrongBack, host)
	if p, _ := ptOf(t, g, host); p != 4 {
		t.Fatalf("the host has power %d with one Aura, want 4", p)
	}
	// Goblin War Paint is {1}{R}: three less on the host leaves {R}.
	paint := handCardFull(me, "Goblin War Paint", "Enchantment — Aura", "{1}{R}", gapB66GoblinPaint, []string{"R"})
	floatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, paint, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}}); err == nil {
		t.Fatal("an Aura targeting another creature was cast for {R}")
	}
	if err := g.CastSpell(me.ID, paint, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}}}); err != nil {
		t.Fatalf("an Aura targeting the enchanted creature was refused for {R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, host); p != 2+2*2+2 {
		t.Errorf("the host has power %d with two Auras, want %d", p, 2+2*2+2)
	}
}

func TestGapB66StrongBackDeclaresBothCostReductionsReadingTargets(t *testing.T) {
	spec, _ := Lookup(gapB66StrongBack)
	if len(spec.CostModifiers) != 2 {
		t.Fatalf("%d cost modifiers, want 2 (Equip abilities, Aura spells)", len(spec.CostModifiers))
	}
	for i, m := range spec.CostModifiers {
		if !m.ReadsTargets {
			t.Errorf("modifier %d does not read targets", i)
		}
	}
	if !spec.CostModifiers[0].Activations || spec.CostModifiers[1].Activations {
		t.Error("the first modifier prices activations and the second prices casts")
	}
}

// --- Grim Reaper's Sprint -------------------------------------------------

func TestGapB66SprintUntapsAndAddsACombatWhenCastInYourMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	toMain(t, g)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	gapB66Aura(t, g, "Grim Reaper's Sprint", gapB66Sprint, bear)
	if isTapped(g, bear) {
		t.Error("the creature is still tapped")
	}
	if p, _ := ptOf(t, g, bear); p != 4 {
		t.Errorf("the enchanted creature has power %d, want 4", p)
	}
	combats := 0
	for i := 0; i < 40 && g.Turn.Step != game.StepPostcombatMain; i++ {
		if stepOnce(t, g).Step == game.StepBeginCombat {
			combats++
		}
	}
	if combats != 2 {
		t.Errorf("%d combat phases before the second main phase, want 2", combats)
	}
}

func TestGapB66SprintCostsThreeLessAfterACreatureDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := b12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	victim := b12Creature(g, g.Seats[1].ID, "Victim", "Creature — Bear", 1, 1)
	sprint := handCardFull(me, "Grim Reaper's Sprint", "Enchantment — Aura", "{4}{R}", gapB66Sprint, []string{"R"})
	floatMana(t, g, me, "{R}{C}")
	if err := g.CastSpell(me.ID, sprint, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("the Aura was cast for two mana before any creature died")
	}
	b27Kill(g, victim)
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, sprint, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("morbid cast for {1}{R}: %v", err)
	}
}

// --- Vault 87: Forced Evolution -------------------------------------------

func gapB66Chapter(t *testing.T, g *game.Game, saga uuid.UUID, pick uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	b27Settle(t, g, g.Seats[0].ID, pick)
}

func TestGapB66Vault87StealsAMutantlessCreatureMakesAMutantAndDrawsItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C", "D", "E", "F")
	mutant := b12Creature(g, opp.ID, "Their Mutant", "Creature — Mutant", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	saga := castCatalogSpell(t, g, "Vault 87: Forced Evolution", "Enchantment — Saga", gapB66Vault87, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for chapter I")
	}
	if hasID(pick.PickTargetCards, mutant) {
		t.Error("a Mutant is offered as a target of chapter I")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, theirs); c.Controller != me.ID {
		t.Fatalf("their Bear is controlled by %v, want me", c.Controller)
	}
	gapB66Chapter(t, g, saga, mine)
	if got := plusOnes(g, mine); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !layeredCard(t, g, mine).HasSubtype("Mutant") {
		t.Error("the creature did not become a Mutant")
	}
	hand := me.Hand.Size()
	gapB66Chapter(t, g, saga, uuid.Nil)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3 (my Mutant is a 3/3)", got)
	}
}

func TestGapB66Vault87TheftEndsWhenTheSagaLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	saga := castCatalogSpell(t, g, "Vault 87: Forced Evolution", "Enchantment — Saga", gapB66Vault87, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	b27Kill(g, saga)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, theirs); c.Controller != opp.ID {
		t.Errorf("their Bear is controlled by %v after the Saga left, want its owner", c.Controller)
	}
}

// --- Vault 21: House Gambit -----------------------------------------------

func TestGapB66Vault21RummagesThenMakesATreasureForEachCardSharingAManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	seedLibrary(me, "A", "B", "C", "D")
	saga := castCatalogSpell(t, g, "Vault 21: House Gambit", "Enchantment — Saga", gapB66Vault21, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed %d at chapter I, want 1", discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after discard and draw, want %d", got, hand)
	}
	// Chapter II rummages the same way.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed %d at chapter II, want 1", discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	// Chapter III: three 2s and a 3 and a land: the 2s share a value.
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
	})
	var picks []uuid.UUID
	for i, cost := range []string{"{1}{W}", "{2}", "{R}{R}", "{3}"} {
		picks = append(picks, tmtB11Hand(g, me, "Spell "+string(rune('A'+i)), "Sorcery", "", cost, 0, 0))
	}
	tmtB11Hand(g, me, "Forest", "Basic Land — Forest", "", "", 0, 0)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(saga, game.CounterLore, 1) })
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatalf("no card choice for chapter III: %+v", g.PendingChoices)
	}
	if len(c.ChooseCards) != 4 {
		t.Errorf("%d cards offered, want the 4 nonland cards", len(c.ChooseCards))
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("%d Treasures, want 3 (three cards of mana value 2)", got)
	}
}
