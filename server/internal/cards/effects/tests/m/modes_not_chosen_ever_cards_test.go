package testsm

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	cardsh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/h"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_not_chosen_ever_cards_test.go — ADR 0097 PR 2 (#1749): the
// cards whose modal clause says "choose one that hasn't been chosen",
// with no duration. The memory lasts as long as the object, so these
// tests cross turns, and the ones whose last bullet moves the card
// check that the new object starts with no memory (CR 400.7).

// --- Demonic Pact ----------------------------------------------------

func TestDemonicPactRunsOutAndMustChooseToLose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	effectstest.B12Permanent(g, me.ID, "Demonic Pact", "Enchantment")
	pact := effectstest.FindBattlefieldByName(g, "Demonic Pact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pact {
				g.Battlefield.Cards[i].OracleID = "19a2f0a0-9e68-4982-a5f5-b77d805befd7"
			}
		}
	})
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Upkeep 1: draw two.
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%4)
	effectstest.NoModePick(t, g, me.ID, "only YOUR upkeep")
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	hand := effectstest.HandSize(me)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+2 {
		t.Errorf("draw two: hand %d → %d", hand, effectstest.HandSize(me))
	}

	// Upkeep 2: the draw is remembered across turns.
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%4)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	oppHand := effectstest.HandSize(opp)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, opp.ID)
	if effectstest.HandSize(opp) != oppHand-2 {
		t.Errorf("target opponent discards two: %d → %d", oppHand, effectstest.HandSize(opp))
	}

	// Upkeep 3: 4 damage to any target and you gain 4.
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%4)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	life := me.Life
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 3}, 0)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, victim) == game.ZoneBattlefield {
		t.Error("4 damage kills the 2/2")
	}
	if me.Life != life+4 {
		t.Errorf("and you gain 4 life: %d → %d", life, me.Life)
	}

	// Upkeep 4: only "You lose the game" is left, and it must be taken.
	effectstest.AdvanceToUpkeepOf(t, g, (seat+1)%4)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	effectstest.ChooseModeNow(t, g, me.ID, []int{3}, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Eliminated {
		t.Error("the fourth mode: you lose the game")
	}
}

// --- Gandalf the Grey ------------------------------------------------

func TestGandalfTheGreyFourBulletsThenTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	gandalf := effectstest.B12Push(g, me.ID, "Gandalf the Grey", "Legendary Creature — Avatar Wizard",
		"dfd12e3c-2b2a-4461-a08d-09d4e6c4626e", 3, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.NoModePick(t, g, me.ID, "a creature spell is not an instant or sorcery")
	effectstest.PassPriorityAroundTable(t, g)

	// Copy the spell that triggered it.
	spell := effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	effectstest.PickCard(t, g, me.ID, spell)
	spells := func() int {
		n := 0
		for _, it := range g.StackMeta {
			if it != nil && it.Kind != game.StackItemTriggered {
				n++
			}
		}
		return n
	}
	before := spells()
	for i := 0; i < 8 && effectstest.TriggerOnStack(g, gandalf) != nil; i++ {
		effectstest.PassPriorityOnce(t, g)
	}
	if spells() != before+1 {
		t.Errorf("the copy is put on the stack: %d spells → %d", before, spells())
	}
	effectstest.PassPriorityAroundTable(t, g)

	lives := make([]int, 4)
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.ID != me.ID && p.Life != lives[i]-3 {
			t.Errorf("3 damage to each opponent: %s %d → %d", p.Name, lives[i], p.Life)
		}
	}

	effectstest.CastCatalogSpell(t, g, "Test Instant 2", "Instant", "", nil)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 3}, 0)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("tap target permanent")
	}

	effectstest.CastCatalogSpell(t, g, "Test Instant 3", "Instant", "", nil)
	effectstest.ChooseModeNow(t, g, me.ID, []int{3}, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOf(g, gandalf); z != game.ZoneLibrary {
		t.Fatalf("Gandalf is put on top of his owner's library, zone %q", z)
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != gandalf {
		t.Errorf("on TOP of the library: top is %q", top.Name)
	}
}

// --- Henrika Domnathi ------------------------------------------------

func henrikaCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   cardsh.HenrikaDomnathiOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Henrika Domnathi", TypeLine: "Legendary Creature — Vampire", ManaCost: "{2}{B}{B}",
				Colors: []string{"B"}, Power: 1, Toughness: 3},
			{Name: "Henrika, Infernal Seer", TypeLine: "Legendary Creature — Vampire",
				Colors: []string{"B"}, Power: 3, Toughness: 4},
		},
	}
	c.SetFace(0)
	return c
}

func TestHenrikaDomnathiEachBulletOnceThenTransforms(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	henrika := effectstest.PushBattlefieldCardWithTimestamp(g, henrikaCard(me.ID))
	fodder := effectstest.B12Creature(g, me.ID, "My Fodder", "Creature — Human", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bat", TypeLine: "Creature — Bat", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	ground := effectstest.B12Creature(g, me.ID, "Ground Pounder", "Creature — Ogre", 3, 3)

	effectstest.AdvanceToStepOf(t, g, seat, game.StepBeginCombat)
	hand, life := effectstest.HandSize(me), me.Life
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+1 || me.Life != life-1 {
		t.Errorf("you draw a card and lose 1 life: hand %d → %d, life %d → %d", hand, effectstest.HandSize(me), life, me.Life)
	}

	effectstest.AdvanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepBeginCombat)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 2}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, fodder) == game.ZoneBattlefield || effectstest.ZoneOf(g, theirs) == game.ZoneBattlefield {
		t.Error("each player sacrifices a creature of their choice")
	}

	effectstest.AdvanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepBeginCombat)
	effectstest.ChooseModeNow(t, g, me.ID, []int{2}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardFor(g, henrika)
	if c == nil || c.ActiveFace != 1 || c.Name != "Henrika, Infernal Seer" {
		t.Fatalf("Transform Henrika: %+v", c)
	}

	// The back face's pump reaches the flyer and Henrika, not the Ogre.
	effectstest.B16Activate(t, g, me.ID, henrika, 0, game.ActivateAbilityParams{})
	if p := effectstest.EffectiveOf(t, g, flyer).Power; p != 2 {
		t.Errorf("a creature with flying gets +1/+0: %d", p)
	}
	if p := effectstest.EffectiveOf(t, g, ground).Power; p != 3 {
		t.Errorf("a creature with none of the three does not: %d", p)
	}
	if p := effectstest.EffectiveOf(t, g, henrika).Power; p != 4 {
		t.Errorf("Henrika herself has all three: %d", p)
	}

	effectstest.AdvanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	effectstest.AdvanceToStepOf(t, g, seat, game.StepBeginCombat)
	effectstest.NoModePick(t, g, me.ID, "the back face has no such trigger")
}

// --- Survivor's Med Kit ----------------------------------------------

func TestSurvivorsMedKitEachBulletOnceForThatObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	kit := effectstest.PushCatalogPermanent(g, me.ID, "Survivor's Med Kit", "Artifact", "fa08400f-b7ca-4c5c-b7ca-a6583b783878", false)
	untap := func() {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(kit) })
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterRad, 3) })

	hand := effectstest.HandSize(me)
	effectstest.B16Activate(t, g, me.ID, kit, 0, game.ActivateAbilityParams{Modes: []int{0}})
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("Stimpak draws: %d → %d", hand, effectstest.HandSize(me))
	}
	untap()
	for _, modes := range activationModesOffered(g, me.ID, kit) {
		if slices.Contains(modes, 0) {
			t.Errorf("the used bullet is never offered: %v", modes)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, kit, 0, game.ActivateAbilityParams{Modes: []int{0}}); err == nil {
		t.Fatal("a used bullet is refused")
	}
	if effectstest.B16Tapped(t, g, kit) {
		t.Error("a refused activation paid nothing — the Kit is still untapped")
	}

	effectstest.B16Activate(t, g, me.ID, kit, 0, game.ActivateAbilityParams{Modes: []int{1}})
	if n := effectstest.B16CountNamed(g, "Food"); n != 1 {
		t.Errorf("Fancy Lads Snack Cakes makes a Food: %d", n)
	}
	untap()
	effectstest.B16Activate(t, g, me.ID, kit, 0, game.ActivateAbilityParams{Modes: []int{2}, Targets: effectstest.B16TargetPlayer(opp.ID)})
	if n := opp.Counters[game.CounterRad]; n != 0 {
		t.Errorf("RadAway removes every rad counter: %d left", n)
	}
	if effectstest.ZoneOf(g, kit) != game.ZoneGraveyard {
		t.Error("and sacrifices the Kit")
	}
}

// --- Kimoyo Beads ----------------------------------------------------

func TestKimoyoBeadsPrimeBeadResetsTheMemory(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushCatalogPermanent(g, me.ID, "Kimoyo Beads", "Artifact", "c7bdbf7a-8054-4e91-bb32-84f47e98553c", false)
	next := func() {
		t.Helper()
		effectstest.AdvanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
		effectstest.AdvanceToEndStepOf(t, g, seat)
	}

	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Soldier"); n != 2 {
		t.Errorf("two Soldiers: %d", n)
	}
	effectstest.AdvanceToEndStepOf(t, g, (seat+1)%4)
	effectstest.NoModePick(t, g, me.ID, "only YOUR end step")

	effectstest.AdvanceToEndStepOf(t, g, seat)
	life := me.Life
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 2}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("Prime Bead gains 3: %d → %d", life, me.Life)
	}

	// Exiled and returned, the Beads are a new object and remember
	// nothing: all three are on offer again.
	next()
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
}

// --- Sol'Kanar the Tainted -------------------------------------------

func TestSolKanarTheTaintedPassesToAnOpponentAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	solkanar := effectstest.B12Push(g, me.ID, "Sol'Kanar the Tainted", "Legendary Creature — Elemental Demon",
		"4efcdefc-e49d-4bce-8581-69037bb48c0a", 5, 5)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)

	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the damage bullet asks for its target")
	}
	if slices.Contains(pick.PickTargetCards, solkanar) {
		t.Error("\"other\": Sol'Kanar is not its own target")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, theirs) == game.ZoneBattlefield {
		t.Error("3 damage kills the 3/3")
	}

	effectstest.AdvanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	effectstest.AdvanceToEndStepOf(t, g, seat)
	lives := make([]int, 4)
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 2
		if p.ID == me.ID {
			want = lives[i] + 2
		}
		if p.Life != want {
			t.Errorf("each opponent loses 2 and you gain 2: %s %d → %d", p.Name, lives[i], p.Life)
		}
	}

	effectstest.AdvanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 3}, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	now := effectstest.FindBattlefieldByName(g, "Sol'Kanar the Tainted")
	c := effectstest.BattlefieldCardFor(g, now)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("Sol'Kanar returns under the chosen opponent's control: %+v", c)
	}
	if len(c.ModesChosen) != 0 {
		t.Errorf("a new object remembers no modes: %v", c.ModesChosen)
	}
}

// --- Zuko, Conflicted ------------------------------------------------

func TestZukoConflictedLosesTwoWithEveryBullet(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	zuko := effectstest.B12Push(g, me.ID, "Zuko, Conflicted", "Legendary Creature — Human Rogue",
		"b5fd82b9-77de-4358-9ce7-915cc809a889", 2, 3)
	nextMain := func() {
		t.Helper()
		effectstest.AdvanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
		effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("you lose 2 life: %d → %d", life, me.Life)
	}
	if n := effectstest.PoolCount(me, "R"); n != 1 {
		t.Errorf("Add {R}: %d red in the pool", n)
	}

	nextMain()
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, zuko, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on Zuko: %d", n)
	}

	nextMain()
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 3}, 3)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardFor(g, effectstest.FindBattlefieldByName(g, "Zuko, Conflicted"))
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("Zuko returns under the opponent's control: %+v", c)
	}
}

// --- Gollum, Riddle Master -------------------------------------------

func TestGollumRiddleMasterTriggersOnTheChosenQuality(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % 4
	opp := g.Seats[oppSeat]
	gollum := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gollum, Riddle Master", TypeLine: "Legendary Creature — Halfling Horror",
		OracleID: "514b451b-814d-45fb-a2ba-8fe6f0bdad60", Power: 3, Toughness: 1,
		Owner: me.ID, Controller: me.ID, ChosenOption: "odd",
	})
	effectstest.AdvanceToMainOf(t, g, oppSeat)
	cast := func(name, cost string) {
		t.Helper()
		id := effectstest.HandCardFull(opp, name, "Sorcery", cost, "", nil)
		if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell %s: %v", name, err)
		}
	}

	cast("Two Drop", "{1}{B}")
	effectstest.NoModePick(t, g, me.ID, "mana value 2 is even; Gollum chose odd")
	effectstest.PassPriorityAroundTable(t, g)
	cast("One Drop", "{B}")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, gollum, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on Gollum: %d", n)
	}
	hand := effectstest.HandSize(me)
	cast("Three Drop", "{2}{B}")
	effectstest.ChooseModeNow(t, g, me.ID, []int{1, 2}, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HandSize(me) != hand+1 {
		t.Errorf("draw a card: %d → %d", hand, effectstest.HandSize(me))
	}
}

// --- Three Bowls of Porridge -----------------------------------------

func TestThreeBowlsOfPorridgeEachBulletOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bowls := effectstest.PushCatalogPermanent(g, me.ID, "Three Bowls of Porridge", "Artifact — Food",
		"5369e40a-5fd6-4a9d-bdc9-4af510449649", false)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	untap := func() { g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bowls) }) }

	effectstest.B16Activate(t, g, me.ID, bowls, 0, game.ActivateAbilityParams{Modes: []int{0}, Targets: effectstest.B16TargetCard(small)})
	if effectstest.ZoneOf(g, small) == game.ZoneBattlefield {
		t.Error("2 damage kills the 2/2")
	}
	untap()
	for _, modes := range activationModesOffered(g, me.ID, bowls) {
		if slices.Contains(modes, 0) {
			t.Errorf("the used bullet is never offered: %v", modes)
		}
	}
	effectstest.B16Activate(t, g, me.ID, bowls, 0, game.ActivateAbilityParams{Modes: []int{1}, Targets: effectstest.B16TargetCard(big)})
	if !effectstest.B16Tapped(t, g, big) {
		t.Error("tap target creature")
	}
	untap()
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, bowls, 0, game.ActivateAbilityParams{Modes: []int{2}})
	if me.Life != life+3 || effectstest.ZoneOf(g, bowls) != game.ZoneGraveyard {
		t.Errorf("sacrifice this artifact, gain 3: life %d → %d, zone %q", life, me.Life, effectstest.ZoneOf(g, bowls))
	}
}
