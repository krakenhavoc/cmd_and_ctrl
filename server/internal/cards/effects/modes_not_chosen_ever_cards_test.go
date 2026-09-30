package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modes_not_chosen_ever_cards_test.go — ADR 0097 PR 2 (#1749): the
// cards whose modal clause says "choose one that hasn't been chosen",
// with no duration. The memory lasts as long as the object, so these
// tests cross turns, and the ones whose last bullet moves the card
// check that the new object starts with no memory (CR 400.7).

// --- Demonic Pact ----------------------------------------------------

func TestDemonicPactRunsOutAndMustChooseToLose(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	b12Permanent(g, me.ID, "Demonic Pact", "Enchantment")
	pact := findBattlefieldByName(g, "Demonic Pact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pact {
				g.Battlefield.Cards[i].OracleID = "19a2f0a0-9e68-4982-a5f5-b77d805befd7"
			}
		}
	})
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Upkeep 1: draw two.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	noModePick(t, g, me.ID, "only YOUR upkeep")
	advanceToUpkeepOf(t, g, seat)
	hand := handSize(me)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+2 {
		t.Errorf("draw two: hand %d → %d", hand, handSize(me))
	}

	// Upkeep 2: the draw is remembered across turns.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	advanceToUpkeepOf(t, g, seat)
	oppHand := handSize(opp)
	chooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, opp.ID)
	if handSize(opp) != oppHand-2 {
		t.Errorf("target opponent discards two: %d → %d", oppHand, handSize(opp))
	}

	// Upkeep 3: 4 damage to any target and you gain 4.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	advanceToUpkeepOf(t, g, seat)
	life := me.Life
	chooseModeNow(t, g, me.ID, []int{0, 3}, 0)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if zoneOf(g, victim) == game.ZoneBattlefield {
		t.Error("4 damage kills the 2/2")
	}
	if me.Life != life+4 {
		t.Errorf("and you gain 4 life: %d → %d", life, me.Life)
	}

	// Upkeep 4: only "You lose the game" is left, and it must be taken.
	advanceToUpkeepOf(t, g, (seat+1)%4)
	advanceToUpkeepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{3}, 3)
	passPriorityAroundTable(t, g)
	if !me.Eliminated {
		t.Error("the fourth mode: you lose the game")
	}
}

// --- Gandalf the Grey ------------------------------------------------

func TestGandalfTheGreyFourBulletsThenTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	gandalf := b12Push(g, me.ID, "Gandalf the Grey", "Legendary Creature — Avatar Wizard",
		"dfd12e3c-2b2a-4461-a08d-09d4e6c4626e", 3, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	noModePick(t, g, me.ID, "a creature spell is not an instant or sorcery")
	passPriorityAroundTable(t, g)

	// Copy the spell that triggered it.
	spell := castCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	pickCard(t, g, me.ID, spell)
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
	for i := 0; i < 8 && triggerOnStack(g, gandalf) != nil; i++ {
		passPriorityOnce(t, g)
	}
	if spells() != before+1 {
		t.Errorf("the copy is put on the stack: %d spells → %d", before, spells())
	}
	passPriorityAroundTable(t, g)

	lives := make([]int, 4)
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	castCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	chooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.ID != me.ID && p.Life != lives[i]-3 {
			t.Errorf("3 damage to each opponent: %s %d → %d", p.Name, lives[i], p.Life)
		}
	}

	castCatalogSpell(t, g, "Test Instant 2", "Instant", "", nil)
	chooseModeNow(t, g, me.ID, []int{0, 3}, 0)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("tap target permanent")
	}

	castCatalogSpell(t, g, "Test Instant 3", "Instant", "", nil)
	chooseModeNow(t, g, me.ID, []int{3}, 3)
	passPriorityAroundTable(t, g)
	if z := zoneOf(g, gandalf); z != game.ZoneLibrary {
		t.Fatalf("Gandalf is put on top of his owner's library, zone %q", z)
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != gandalf {
		t.Errorf("on TOP of the library: top is %q", top.Name)
	}
}

// passPriorityOnce passes priority a single time.
func passPriorityOnce(t *testing.T, g *game.Game) {
	t.Helper()
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
}

// --- Henrika Domnathi ------------------------------------------------

func henrikaCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   henrikaDomnathiOracle,
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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	henrika := pushBattlefieldCardWithTimestamp(g, henrikaCard(me.ID))
	fodder := b12Creature(g, me.ID, "My Fodder", "Creature — Human", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	flyer := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bat", TypeLine: "Creature — Bat", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID,
	})
	ground := b12Creature(g, me.ID, "Ground Pounder", "Creature — Ogre", 3, 3)

	advanceToStepOf(t, g, seat, game.StepBeginCombat)
	hand, life := handSize(me), me.Life
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 1)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 || me.Life != life-1 {
		t.Errorf("you draw a card and lose 1 life: hand %d → %d, life %d → %d", hand, handSize(me), life, me.Life)
	}

	advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	advanceToStepOf(t, g, seat, game.StepBeginCombat)
	chooseModeNow(t, g, me.ID, []int{0, 2}, 0)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, fodder)
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if zoneOf(g, fodder) == game.ZoneBattlefield || zoneOf(g, theirs) == game.ZoneBattlefield {
		t.Error("each player sacrifices a creature of their choice")
	}

	advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	advanceToStepOf(t, g, seat, game.StepBeginCombat)
	chooseModeNow(t, g, me.ID, []int{2}, 2)
	passPriorityAroundTable(t, g)
	c := battlefieldCardFor(g, henrika)
	if c == nil || c.ActiveFace != 1 || c.Name != "Henrika, Infernal Seer" {
		t.Fatalf("Transform Henrika: %+v", c)
	}

	// The back face's pump reaches the flyer and Henrika, not the Ogre.
	b16Activate(t, g, me.ID, henrika, 0, game.ActivateAbilityParams{})
	if p := effectiveOf(t, g, flyer).Power; p != 2 {
		t.Errorf("a creature with flying gets +1/+0: %d", p)
	}
	if p := effectiveOf(t, g, ground).Power; p != 3 {
		t.Errorf("a creature with none of the three does not: %d", p)
	}
	if p := effectiveOf(t, g, henrika).Power; p != 4 {
		t.Errorf("Henrika herself has all three: %d", p)
	}

	advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	advanceToStepOf(t, g, seat, game.StepBeginCombat)
	noModePick(t, g, me.ID, "the back face has no such trigger")
}

// --- Survivor's Med Kit ----------------------------------------------

func TestSurvivorsMedKitEachBulletOnceForThatObject(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	kit := pushCatalogPermanent(g, me.ID, "Survivor's Med Kit", "Artifact", "fa08400f-b7ca-4c5c-b7ca-a6583b783878", false)
	untap := func() {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(kit) })
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterRad, 3) })

	hand := handSize(me)
	b16Activate(t, g, me.ID, kit, 0, game.ActivateAbilityParams{Modes: []int{0}})
	if handSize(me) != hand+1 {
		t.Errorf("Stimpak draws: %d → %d", hand, handSize(me))
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
	if b16Tapped(t, g, kit) {
		t.Error("a refused activation paid nothing — the Kit is still untapped")
	}

	b16Activate(t, g, me.ID, kit, 0, game.ActivateAbilityParams{Modes: []int{1}})
	if n := b16CountNamed(g, "Food"); n != 1 {
		t.Errorf("Fancy Lads Snack Cakes makes a Food: %d", n)
	}
	untap()
	b16Activate(t, g, me.ID, kit, 0, game.ActivateAbilityParams{Modes: []int{2}, Targets: b16TargetPlayer(opp.ID)})
	if n := opp.Counters[game.CounterRad]; n != 0 {
		t.Errorf("RadAway removes every rad counter: %d left", n)
	}
	if zoneOf(g, kit) != game.ZoneGraveyard {
		t.Error("and sacrifices the Kit")
	}
}

// --- Kimoyo Beads ----------------------------------------------------

func TestKimoyoBeadsPrimeBeadResetsTheMemory(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushCatalogPermanent(g, me.ID, "Kimoyo Beads", "Artifact", "c7bdbf7a-8054-4e91-bb32-84f47e98553c", false)
	next := func() {
		t.Helper()
		advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
		advanceToEndStepOf(t, g, seat)
	}

	advanceToEndStepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 1)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Soldier"); n != 2 {
		t.Errorf("two Soldiers: %d", n)
	}
	advanceToEndStepOf(t, g, (seat+1)%4)
	noModePick(t, g, me.ID, "only YOUR end step")

	advanceToEndStepOf(t, g, seat)
	life := me.Life
	chooseModeNow(t, g, me.ID, []int{0, 2}, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("Prime Bead gains 3: %d → %d", life, me.Life)
	}

	// Exiled and returned, the Beads are a new object and remember
	// nothing: all three are on offer again.
	next()
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
}

// --- Sol'Kanar the Tainted -------------------------------------------

func TestSolKanarTheTaintedPassesToAnOpponentAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	solkanar := b12Push(g, me.ID, "Sol'Kanar the Tainted", "Legendary Creature — Elemental Demon",
		"4efcdefc-e49d-4bce-8581-69037bb48c0a", 5, 5)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)

	advanceToEndStepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the damage bullet asks for its target")
	}
	if slices.Contains(pick.PickTargetCards, solkanar) {
		t.Error("\"other\": Sol'Kanar is not its own target")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if zoneOf(g, theirs) == game.ZoneBattlefield {
		t.Error("3 damage kills the 3/3")
	}

	advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	advanceToEndStepOf(t, g, seat)
	lives := make([]int, 4)
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	chooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i] - 2
		if p.ID == me.ID {
			want = lives[i] + 2
		}
		if p.Life != want {
			t.Errorf("each opponent loses 2 and you gain 2: %s %d → %d", p.Name, lives[i], p.Life)
		}
	}

	advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
	advanceToEndStepOf(t, g, seat)
	chooseModeNow(t, g, me.ID, []int{0, 3}, 3)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	passPriorityAroundTable(t, g)
	now := findBattlefieldByName(g, "Sol'Kanar the Tainted")
	c := battlefieldCardFor(g, now)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("Sol'Kanar returns under the chosen opponent's control: %+v", c)
	}
	if len(c.ModesChosen) != 0 {
		t.Errorf("a new object remembers no modes: %v", c.ModesChosen)
	}
}

// --- Zuko, Conflicted ------------------------------------------------

func TestZukoConflictedLosesTwoWithEveryBullet(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%4]
	zuko := b12Push(g, me.ID, "Zuko, Conflicted", "Legendary Creature — Human Rogue",
		"b5fd82b9-77de-4358-9ce7-915cc809a889", 2, 3)
	nextMain := func() {
		t.Helper()
		advanceToStepOf(t, g, (seat+1)%4, game.StepUpkeep)
		advanceToPrecombatMainOf(t, g, seat)
	}

	advanceToPrecombatMainOf(t, g, seat)
	life := me.Life
	chooseModeNow(t, g, me.ID, []int{0, 1, 2, 3}, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("you lose 2 life: %d → %d", life, me.Life)
	}
	if n := poolCount(me, "R"); n != 1 {
		t.Errorf("Add {R}: %d red in the pool", n)
	}

	nextMain()
	chooseModeNow(t, g, me.ID, []int{0, 1, 3}, 1)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, zuko, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on Zuko: %d", n)
	}

	nextMain()
	chooseModeNow(t, g, me.ID, []int{0, 3}, 3)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	passPriorityAroundTable(t, g)
	c := battlefieldCardFor(g, findBattlefieldByName(g, "Zuko, Conflicted"))
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("Zuko returns under the opponent's control: %+v", c)
	}
}

// --- Gollum, Riddle Master -------------------------------------------

func TestGollumRiddleMasterTriggersOnTheChosenQuality(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	oppSeat := (seat + 1) % 4
	opp := g.Seats[oppSeat]
	gollum := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Gollum, Riddle Master", TypeLine: "Legendary Creature — Halfling Horror",
		OracleID: "514b451b-814d-45fb-a2ba-8fe6f0bdad60", Power: 3, Toughness: 1,
		Owner: me.ID, Controller: me.ID, ChosenOption: "odd",
	})
	advanceToMainOf(t, g, oppSeat)
	cast := func(name, cost string) {
		t.Helper()
		id := handCardFull(opp, name, "Sorcery", cost, "", nil)
		if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("CastSpell %s: %v", name, err)
		}
	}

	cast("Two Drop", "{1}{B}")
	noModePick(t, g, me.ID, "mana value 2 is even; Gollum chose odd")
	passPriorityAroundTable(t, g)
	cast("One Drop", "{B}")
	chooseModeNow(t, g, me.ID, []int{0, 1, 2}, 0)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, gollum, game.CounterPlusOne); n != 1 {
		t.Errorf("a +1/+1 counter on Gollum: %d", n)
	}
	hand := handSize(me)
	cast("Three Drop", "{2}{B}")
	chooseModeNow(t, g, me.ID, []int{1, 2}, 2)
	passPriorityAroundTable(t, g)
	if handSize(me) != hand+1 {
		t.Errorf("draw a card: %d → %d", hand, handSize(me))
	}
}

// --- Three Bowls of Porridge -----------------------------------------

func TestThreeBowlsOfPorridgeEachBulletOnce(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bowls := pushCatalogPermanent(g, me.ID, "Three Bowls of Porridge", "Artifact — Food",
		"5369e40a-5fd6-4a9d-bdc9-4af510449649", false)
	small := b12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 4, 4)
	untap := func() { g.WithWriteLock(func() { _ = g.UntapTargetForEffect(bowls) }) }

	b16Activate(t, g, me.ID, bowls, 0, game.ActivateAbilityParams{Modes: []int{0}, Targets: b16TargetCard(small)})
	if zoneOf(g, small) == game.ZoneBattlefield {
		t.Error("2 damage kills the 2/2")
	}
	untap()
	for _, modes := range activationModesOffered(g, me.ID, bowls) {
		if slices.Contains(modes, 0) {
			t.Errorf("the used bullet is never offered: %v", modes)
		}
	}
	b16Activate(t, g, me.ID, bowls, 0, game.ActivateAbilityParams{Modes: []int{1}, Targets: b16TargetCard(big)})
	if !b16Tapped(t, g, big) {
		t.Error("tap target creature")
	}
	untap()
	life := me.Life
	b16Activate(t, g, me.ID, bowls, 0, game.ActivateAbilityParams{Modes: []int{2}})
	if me.Life != life+3 || zoneOf(g, bowls) != game.ZoneGraveyard {
		t.Errorf("sacrifice this artifact, gain 3: life %d → %d, zone %q", life, me.Life, zoneOf(g, bowls))
	}
}
