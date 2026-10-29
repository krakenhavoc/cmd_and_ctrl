package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a11_cards_test.go — The Hobbit / Tales of Middle-earth, half A, eleventh
// batch (#4454): a Goblin with a reflexive fling, a divided-damage Wizard, an
// extra-combat Worm, two Adventures, an Elf mana creature, a bounce with a
// gift, Saruman and a graveyard thief. Each test plays the real catalog entry
// and asserts what the card's text says.

const (
	hobA11BolgOracle        = "88522a0f-5377-4522-97f4-4148bef954af"
	hobA11GandalfOracle     = "cd1a9d87-4c91-4456-8d88-19e0d6e0b200"
	hobA11WormOracle        = "81c2913f-42b7-4cfa-adc0-b3416fbe2ec6"
	hobA11BilboOracle       = "a73d9109-36e0-40ea-ab06-5c5f7eb4ac54"
	hobA11GlamdringOracle   = "2802069f-201e-43a7-b5d3-43a95951a2ec"
	hobA11GambitOracle      = "ef71ef07-8a35-41ec-a851-d95e9e3a221b"
	hobA11WeavemasterOracle = "28ffddca-a3c4-467d-a2dc-2c5e62393e11"
	hobA11SarumanOracle     = "c2d04138-cc17-4dd0-9a4c-9c87df72ed01"
	hobA11ShadowOracle      = "4f0c72ff-cd93-46a6-8649-2cd9ce8b8625"
)

func TestHobA11EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hobA11BolgOracle:             "Bolg of the North",
		hobA11GandalfOracle:          "Gandalf, Spark Starter",
		hobA11WormOracle:             "Desert Were-Worm",
		hobA11BilboOracle:            "Bilbo, Luckwearer",
		hobA11BilboOracle + "#1":     "Burglar's Plot",
		hobA11GlamdringOracle:        "Glamdring, Foe-hammer",
		hobA11GlamdringOracle + "#1": "Gleam of Death",
		hobA11GambitOracle:           "Bilbo's Gambit",
		hobA11WeavemasterOracle:      "Woodland Weavemaster",
		hobA11SarumanOracle:          "Saruman of Many Colors",
		hobA11ShadowOracle:           "Shadow of the Enemy",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// --- Gandalf, Spark Starter --------------------------------------------------

func TestHobA11GandalfDividesThreeDamageOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := b12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	gandalf := castCatalogSpell(t, g, "Gandalf, Spark Starter", "Legendary Creature — Avatar Wizard", hobA11GandalfOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, gandalf, "reach") {
		t.Error("Gandalf has reach")
	}
	b17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 1}, a, b)
	passPriorityAroundTable(t, g)
	if d1, d2 := e2Card(t, g, a).DamageMarked, e2Card(t, g, b).DamageMarked; d1 != 2 || d2 != 1 {
		t.Errorf("damage %d/%d, want 2/1", d1, d2)
	}
}

// A single target takes all three, and a player is a legal target.
func TestHobA11GandalfCanPutAllThreeOnOnePlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	castCatalogSpell(t, g, "Gandalf, Spark Starter", "Legendary Creature — Avatar Wizard", hobA11GandalfOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for targets")
	}
	refs := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{opp.ID: 3}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
}

// A split that does not add up to three is refused.
func TestHobA11GandalfRefusesASplitThatIsNotThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	castCatalogSpell(t, g, "Gandalf, Spark Starter", "Legendary Creature — Avatar Wizard", hobA11GandalfOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for targets")
	}
	refs := []game.TargetRef{{Kind: game.TargetCard, ID: a}}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{a: 2}); err == nil {
		t.Error("a split of 2 was accepted for 3 damage")
	}
}

// --- Bolg of the North ---------------------------------------------------------

// castHobA11Bolg casts Bolg and resolves the spell, leaving the enters
// trigger's "you may sacrifice" prompt open.
func castHobA11Bolg(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	bolg := castCatalogSpell(t, g, "Bolg of the North", "Legendary Creature — Goblin Soldier", hobA11BolgOracle, nil)
	passPriorityAroundTable(t, g)
	return bolg
}

// Sacrificing a 4-power creature to hit a 1/2: 4 damage, 2 of it excess, so
// the Goblin amass puts two counters on a new Army.
func TestHobA11BolgSacrificeDealsItsPowerAndAmassesTheExcess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Ogre", 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 2)
	castHobA11Bolg(t, g)

	answerOwnPermanents(t, g, me.ID, fodder)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the sacrifice happens before the reflexive trigger asks for a target")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("4 damage kills a 1/2")
	}
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2 (4 damage less the 2 it needed)", got)
	}
	if !army.HasSubtype("Goblin") {
		t.Error("the amassed Army is a Goblin")
	}
}

// The damage is the creature's power as it last stood: an anthem counts.
func TestHobA11BolgCountsAnAnthemInThePower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Ogre", 1, 1)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 1)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: gloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	castHobA11Bolg(t, g)
	answerOwnPermanents(t, g, me.ID, fodder)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if got := countersOn(g, armyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("Army counters = %d, want 1 (a 2-power sacrifice into a 1/1)", got)
	}
}

// Exactly lethal is no excess: the creature dies and nothing is amassed.
func TestHobA11BolgWithNoExcessAmassesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Ogre", 3, 3)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 3)
	castHobA11Bolg(t, g)
	answerOwnPermanents(t, g, me.ID, fodder)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("3 damage kills a 1/3")
	}
	for _, c := range g.Battlefield.Cards {
		if game.IsArmy(c) {
			t.Error("no excess damage: no Army is made")
		}
	}
}

// "You may": declining sacrifices nothing and deals no damage.
func TestHobA11BolgMayDeclineTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Ogre", 4, 4)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 2)
	castHobA11Bolg(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(victim) {
		t.Error("declining changes nothing")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no sacrifice, no reflexive trigger")
	}
}

// "Another creature": Bolg himself is not offered to be sacrificed, and he is
// not "another target creature" for his own damage either.
func TestHobA11BolgOffersOnlyOtherCreaturesAndCannotHitHimself(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Ogre", 2, 2)
	b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 2)
	bolg := castHobA11Bolg(t, g)

	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	for _, id := range c.ChooseCards {
		if id == bolg {
			t.Error("Bolg cannot sacrifice himself")
		}
	}
	answerOwnPermanents(t, g, me.ID, fodder)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, id := range p.PickTargetCards {
		if id == bolg {
			t.Error("Bolg is not \"another target creature\"")
		}
	}
}

// --- Desert Were-Worm -------------------------------------------------------------

const hobA11WormLabel = "Desert Were-Worm — untap all attacking creatures; after this phase, an additional combat phase"

// hobA11Worm puts a Desert Were-Worm (a printed 0/5) and `mountains` Mountains
// under `owner`.
func hobA11Worm(g *game.Game, owner uuid.UUID, mountains int) uuid.UUID {
	for i := 0; i < mountains; i++ {
		b12Permanent(g, owner, "Mountain", "Basic Land — Mountain")
	}
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Desert Were-Worm", TypeLine: "Creature — Dragon Wurm",
		OracleID: hobA11WormOracle, Power: 0, Toughness: 5, Owner: owner, Controller: owner,
	})
}

func TestHobA11WormGetsTwoPowerPerMountainYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worm := hobA11Worm(g, me.ID, 0)
	if p := e2Card(t, g, worm).CurrentPower(); p != 0 {
		t.Errorf("no Mountains: power %d, want 0", p)
	}
	for i := 0; i < 3; i++ {
		b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	b12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	c := e2Card(t, g, worm)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 5 {
		t.Errorf("three Mountains: %d/%d, want 6/5 (an opponent's Mountain and a Forest do not count)", c.CurrentPower(), c.CurrentToughness())
	}
}

// Twelve power attacking: the attackers untap, one more combat follows, and
// attacking with twelve again in it does not trigger a third.
func TestHobA11WormUntapsAttackersAndAddsOneCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	worm := hobA11Worm(g, me.ID, 5)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{worm, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if countResolved(g, start, hobA11WormLabel) != 1 {
		t.Fatal("the Worm's trigger did not resolve once")
	}
	for _, id := range []uuid.UUID{worm, bear} {
		if c, _ := battlefieldCard(g, id); c.Tapped {
			t.Errorf("%s is still tapped: all attacking creatures untap", c.Name)
		}
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after the first combat: %+v, want a second combat", next)
	}
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{worm, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker in the second combat: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, hobA11WormLabel); n != 1 {
		t.Errorf("the trigger resolved %d times: it is once per turn, so the second combat adds nothing", n)
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the second combat: %+v, want the postcombat main phase", next)
	}
}

// Total power counts the creatures that ATTACK: a Worm of ten attacking alone
// is short of twelve even with a 2/2 at home.
func TestHobA11WormCountsOnlyTheAttackersPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	worm := hobA11Worm(g, me.ID, 5)
	pushVanillaCreature(g, me.ID, "Bear at home", 2, 2)
	start := len(g.Events)

	declareAttack(t, g, opp.ID, worm)
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, hobA11WormLabel); n != 0 {
		t.Errorf("the trigger resolved %d times with only 10 power attacking", n)
	}
	if c, _ := battlefieldCard(g, worm); !c.Tapped {
		t.Error("the Worm stays tapped: nothing untapped it")
	}
	advanceTo(t, g, game.StepEndCombat)
	if next := stepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after combat: %+v, want the postcombat main phase (no additional combat)", next)
	}
}

// Twelve from several creatures is one trigger, not one per attacker.
func TestHobA11WormTriggersOncePerAttackNotPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	worm := hobA11Worm(g, me.ID, 3)
	a := pushVanillaCreature(g, me.ID, "Ogre A", 3, 3)
	b := pushVanillaCreature(g, me.ID, "Ogre B", 3, 3)
	start := len(g.Events)

	declareAttack(t, g, opp.ID, worm, a, b)
	passPriorityAroundTable(t, g)
	if n := countResolved(g, start, hobA11WormLabel); n != 1 {
		t.Errorf("the trigger resolved %d times, want 1 (6 + 3 + 3 = 12)", n)
	}
}

// --- Bilbo, Luckwearer // Burglar's Plot ----------------------------------------

var (
	hobA11BilboCreature = hobA6Face("Bilbo, Luckwearer", "Legendary Creature — Halfling Rogue", "{1}{U}", 1, 1, "U")
	hobA11BurglarsPlot  = hobA6Face("Burglar's Plot", "Sorcery — Adventure", "{4}{U}", 0, 0, "U")
)

func TestHobA11BilboCantBeBlockedAndLoots(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bilbo := b12Push(g, me.ID, "Bilbo, Luckwearer", "Legendary Creature — Halfling Rogue", hobA11BilboOracle, 1, 1)
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	pushLibraryTopForTest(me, "Drawn Card")
	handBefore := len(me.Hand.Cards)
	life := opp.Life

	declareAttack(t, g, opp.ID, bilbo)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bilbo); err == nil {
		t.Fatal("a creature blocked an unblockable Bilbo")
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, opp.Life)
	}
	if len(me.Hand.Cards) != handBefore+1 {
		t.Fatalf("hand = %d, want %d: the draw comes before the discard", len(me.Hand.Cards), handBefore+1)
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("hand = %d, want %d after draw then discard", got, handBefore)
	}
}

func hobA11Control(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Controller
}

// An artifact and an artifact creature share a card type: control swaps, and
// the Adventure goes to exile to be cast later.
func TestHobA11BurglarsPlotExchangesTwoPermanentsThatShareAType(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := hobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	mine := b12Permanent(g, me.ID, "Mighty Hammer", "Artifact — Equipment")
	theirs := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)

	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{hobA6Card(mine), hobA6Card(theirs)})

	if hobA11Control(t, g, mine) != opp.ID || hobA11Control(t, g, theirs) != me.ID {
		t.Error("the Hammer and the Golem should have changed hands")
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestHobA11BurglarsPlotWorksOnTwoCreaturesAndTheOrderOfPicksDoesNotMatter(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := hobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)

	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{hobA6Card(theirs), hobA6Card(mine)})

	if hobA11Control(t, g, mine) != opp.ID || hobA11Control(t, g, theirs) != me.ID {
		t.Error("the creatures should have changed hands")
	}
}

// A land is never a target, and the two must share a type: an enchantment and
// a creature do not, an artifact creature and an artifact land (a land) are
// refused too.
func TestHobA11BurglarsPlotRefusesTargetsThatShareNoTypeOrAreLands(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := hobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	aura := b12Permanent(g, me.ID, "My Wall of Words", "Enchantment")
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	vault := b12Permanent(g, me.ID, "Darksteel Citadel", "Artifact Land")
	golem := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)

	if err := hobA6Cast(g, me, id, 1, []game.TargetRef{hobA6Card(aura), hobA6Card(ogre)}); err == nil {
		t.Error("an enchantment and a creature share no card type")
	}
	if err := hobA6Cast(g, me, id, 1, []game.TargetRef{hobA6Card(vault), hobA6Card(golem)}); err == nil {
		t.Error("a land is not a legal target, even one that shares a type")
	}
	if err := hobA6Cast(g, me, id, 1, []game.TargetRef{hobA6Card(ogre), hobA6Card(ogre)}); err == nil {
		t.Error("the two targets are different permanents")
	}
}

// All or nothing (CR 701.12b): when one target is gone as the spell resolves
// nothing changes hands.
func TestHobA11BurglarsPlotDoesNothingWhenATargetIsGone(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := hobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)

	if err := hobA6Cast(g, me, id, 1, []game.TargetRef{hobA6Card(mine), hobA6Card(theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	moveToGraveyard(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)

	if hobA11Control(t, g, mine) != me.ID {
		t.Error("my Bear changed hands although its partner was gone")
	}
}

// --- Glamdring, Foe-hammer // Gleam of Death -----------------------------------

var (
	hobA11GlamdringFace = hobA6Face("Glamdring, Foe-hammer", "Legendary Artifact — Equipment", "{2}", 0, 0)
	hobA11GleamOfDeath  = hobA6Face("Gleam of Death", "Sorcery — Adventure", "{3}{U}", 0, 0, "U")
)

func hobA11GlamdringOnCreature(t *testing.T, g *game.Game, me *game.Player, power int) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	glamdring := pushCatalogPermanent(g, me.ID, "Glamdring, Foe-hammer", "Legendary Artifact — Equipment", hobA11GlamdringOracle, false)
	host := b12Creature(g, me.ID, "Wielder", "Creature — Human", power, 3)
	equipTo(t, g, me.ID, glamdring, host)
	return glamdring
}

func TestHobA11GlamdringDiscountsYourInstantsAndSorceriesByTheWielderPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobA11GlamdringOnCreature(t, g, me, 3)

	if got := priceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 2 {
		t.Errorf("sorcery of mana value 5: %d, want 2 (3 less)", got)
	}
	if got := priceInHand(t, g, me, "Big Instant", "Instant", "{4}{U}"); got != 2 {
		t.Errorf("instant of mana value 5: %d, want 2 (3 less)", got)
	}
	if got := priceInHand(t, g, me, "Cheap Instant", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant of mana value 2: %d, want 1 (the generic is all it can spend against)", got)
	}
	if got := priceInHand(t, g, me, "Big Ogre", "Creature — Ogre", "{4}{R}"); got != 5 {
		t.Errorf("creature spell: %d, want 5 (not an instant or sorcery)", got)
	}
	if got := priceInHand(t, g, opp, "Their Sorcery", "Sorcery", "{4}{R}"); got != 5 {
		t.Errorf("an opponent's sorcery: %d, want 5 (spells YOU cast)", got)
	}
}

// X moves with the wielder's power, and an unattached Glamdring gives nothing.
func TestHobA11GlamdringFollowsThePowerAndNeedsAWielder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	glamdring := pushCatalogPermanent(g, me.ID, "Glamdring, Foe-hammer", "Legendary Artifact — Equipment", hobA11GlamdringOracle, false)
	if got := priceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 5 {
		t.Errorf("unattached: %d, want 5", got)
	}
	host := b12Creature(g, me.ID, "Wielder", "Creature — Human", 1, 3)
	equipTo(t, g, me.ID, glamdring, host)
	if got := priceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 4 {
		t.Errorf("1-power wielder: %d, want 4", got)
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: gloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := priceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 3 {
		t.Errorf("an anthem makes the wielder a 2-power creature: %d, want 3", got)
	}
}

func TestHobA11GlamdringEquipsForTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	glamdring := pushCatalogPermanent(g, me.ID, "Glamdring, Foe-hammer", "Legendary Artifact — Equipment", hobA11GlamdringOracle, false)
	host := b12Creature(g, me.ID, "Wielder", "Creature — Human", 1, 3)
	if got := equipPrice(t, g, me.ID, glamdring, host); got != "{2}" {
		t.Errorf("equip costs %s, want {2}", got)
	}
}

// Gleam of Death mills six and takes every instant and sorcery among them,
// leaving the rest in the graveyard and the seventh card in the library.
func TestHobA11GleamOfDeathMillsSixAndTakesTheInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	activeSeat := g.Seats[g.Turn.ActiveSeat]
	activeSeat.Library.Cards = nil
	deep := uuid.New()
	var ids []uuid.UUID
	for _, c := range []struct{ name, typeLine string }{
		{"Deep Sorcery", "Sorcery"}, // the seventh card: stays in the library
		{"A Land", "Basic Land — Island"},
		{"Instant One", "Instant"},
		{"An Ogre", "Creature — Ogre"},
		{"Sorcery One", "Sorcery"},
		{"An Aura", "Enchantment — Aura"},
		{"Instant Two", "Instant"},
	} {
		id := uuid.New()
		if c.name == "Deep Sorcery" {
			id = deep
		}
		ids = append(ids, id)
		activeSeat.Library.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: c.typeLine, Owner: activeSeat.ID, Controller: activeSeat.ID})
	}
	// Top first: Instant Two, An Aura, Sorcery One, An Ogre, Instant One, A Land | Deep Sorcery.
	me, id := hobA6HandAdventure(t, g, hobA11GlamdringOracle, hobA11GlamdringFace, hobA11GleamOfDeath)
	hobA6MustCast(t, g, me, id, 1, nil)

	for _, i := range []int{2, 4, 6} {
		if !me.Hand.Contains(ids[i]) {
			t.Errorf("card %d (an instant or sorcery) should be in hand", i)
		}
	}
	for _, i := range []int{1, 3, 5} {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("card %d (not an instant or sorcery) should stay in the graveyard", i)
		}
	}
	if !me.Library.Contains(deep) {
		t.Error("the seventh card was not milled and must stay in the library")
	}
	hobA6WentOnAnAdventure(t, g, id)
}

// --- Bilbo's Gambit -------------------------------------------------------------

// hobA11ShockOnTheStack has the opponent cast a Shock at the caster and
// returns it.
func hobA11ShockOnTheStack(t *testing.T, g *game.Game, caster, opponent *game.Player) uuid.UUID {
	t.Helper()
	advanceToMain(t, g) // walking the turn clears the stack, so get there first
	shock := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant",
		OracleID: b5ShockOracle, Owner: opponent.ID, Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, shock, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent CastSpell Shock: %v", err)
	}
	return shock
}

func hobA11CastAnything(g *game.Game, who *game.Player) error {
	id := uuid.New()
	who.Hand.PushTop(game.Card{InstanceID: id, Name: "Shock", TypeLine: "Instant", OracleID: b5ShockOracle, Owner: who.ID, Controller: who.ID})
	return g.CastSpell(who.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: who.ID}}})
}

// Promised: the opponent gets a Treasure, the spell goes back to its owner's
// hand, and nobody, the caster included, can cast a spell for the rest of the
// turn.
func TestHobA11GambitWithGiftBouncesTheSpellAndLocksTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	shock := hobA11ShockOnTheStack(t, g, caster, opp)

	mustCastWithGift(t, g, "Bilbo's Gambit", "Instant", hobA11GambitOracle, cardTarget(shock), opp.ID)
	passPriorityAroundTable(t, g)

	if g.Stack.Contains(shock) || !opp.Hand.Contains(shock) {
		t.Error("the Shock should be back in its owner's hand")
	}
	if opp.Graveyard.Contains(shock) {
		t.Error("it is returned, not countered")
	}
	if n, _ := countTokensNamed(g, opp.ID, "Treasure"); n != 1 {
		t.Errorf("the opponent has %d Treasures, want the 1 gift", n)
	}
	for _, p := range []*game.Player{caster, opp, third} {
		if err := hobA11CastAnything(g, p); err == nil {
			t.Errorf("%s cast a spell after the gift was promised", p.ID)
		}
	}
}

// Not promised: only the bounce, no Treasure and no lock.
func TestHobA11GambitWithoutGiftOnlyBounces(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	shock := hobA11ShockOnTheStack(t, g, caster, opp)

	mustCastWithGift(t, g, "Bilbo's Gambit", "Instant", hobA11GambitOracle, cardTarget(shock), uuid.Nil)
	passPriorityAroundTable(t, g)

	if !opp.Hand.Contains(shock) {
		t.Error("the Shock should be back in its owner's hand")
	}
	if n, _ := countTokensNamed(g, opp.ID, "Treasure"); n != 0 {
		t.Errorf("the opponent has %d Treasures, want none", n)
	}
	if err := hobA11CastAnything(g, caster); err != nil {
		t.Errorf("without the gift anyone may still cast: %v", err)
	}
}

// The gift is made before the other effects, and it is the CHOSEN opponent's.
func TestHobA11GambitGivesTheTreasureToTheChosenOpponent(t *testing.T) {
	g := newCatalogGame(t)
	caster, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	shock := hobA11ShockOnTheStack(t, g, caster, opp)

	mustCastWithGift(t, g, "Bilbo's Gambit", "Instant", hobA11GambitOracle, cardTarget(shock), other.ID)
	passPriorityAroundTable(t, g)

	if n, _ := countTokensNamed(g, other.ID, "Treasure"); n != 1 {
		t.Errorf("the chosen opponent has %d Treasures, want 1", n)
	}
	if n, _ := countTokensNamed(g, opp.ID, "Treasure"); n != 0 {
		t.Errorf("the spell's owner has %d Treasures, want none", n)
	}
}

// --- Woodland Weavemaster ---------------------------------------------------------

func TestHobA11WeavemasterTapsForManaEqualToItsPowerRestrictedToElves(t *testing.T) {
	g, me, _ := spendTable(t)
	weaver := b12Push(g, me.ID, "Woodland Weavemaster", "Creature — Elf Druid", hobA11WeavemasterOracle, 1, 2)
	if !hasEffectiveKeyword(t, g, weaver, "vigilance") {
		t.Error("Weavemaster has vigilance")
	}
	if err := g.ActivateManaAbility(me.ID, weaver, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap Weavemaster: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "G" {
		t.Fatalf("pool = %v, want a single {G}", got)
	}
	cost, _ := game.ParseCost("{G}")
	elf := game.Card{Name: "Elf", TypeLine: "Creature — Elf Warrior", ManaCost: "{G}"}
	bear := game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{G}"}
	elfRock := game.Card{Name: "Elven Rock", TypeLine: "Artifact — Elf", ManaCost: "{2}"}
	rock := game.Card{Name: "Rock", TypeLine: "Artifact", ManaCost: "{2}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(elf)) {
		t.Error("refused an Elf spell")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(elf)) {
		t.Error("refused an ability of an Elf source")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(elfRock)) {
		t.Error("refused an ability of an Elf that is not a creature")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("paid for a spell that is not an Elf")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(rock)) {
		t.Error("paid for an ability of a source that is not an Elf")
	}
}

// X is the Weavemaster's power as the ability is activated, so the +1/+1 from
// another Elf entering counts, and the mana is all one colour.
func TestHobA11WeavemasterGrowsWhenAnotherElfEntersAndMakesThatMuchMana(t *testing.T) {
	g, me, opp := spendTable(t)
	weaver := b12Push(g, me.ID, "Woodland Weavemaster", "Creature — Elf Druid", hobA11WeavemasterOracle, 1, 2)

	b12Creature(g, opp.ID, "Their Elf", "Creature — Elf Scout", 1, 1)
	passPriorityAroundTable(t, g)
	if p := e2Card(t, g, weaver).CurrentPower(); p != 1 {
		t.Fatalf("an opponent's Elf entering changed the power to %d", p)
	}
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if p := e2Card(t, g, weaver).CurrentPower(); p != 1 {
		t.Fatalf("a Bear entering changed the power to %d: only an Elf counts", p)
	}
	elf := handSpell(me, "Elf", "Creature — Elf Warrior", "")
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast an Elf: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := e2Card(t, g, weaver)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Fatalf("after another Elf: %d/%d, want 2/3", c.CurrentPower(), c.CurrentToughness())
	}
	if err := g.ActivateManaAbility(me.ID, weaver, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap Weavemaster: %v", err)
	}
	if got := poolColors(me); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("pool = %v, want {U}{U}: X mana of one colour", got)
	}
}

// --- Saruman of Many Colors -------------------------------------------------------

// Paying the ward means discarding an enchantment, instant or sorcery card:
// a creature card is refused, and the spell then resolves.
func TestHobA11SarumanWardTakesAnEnchantmentInstantOrSorceryCard(t *testing.T) {
	g, _, opp, warded := wardTable(t, "Saruman of Many Colors", "Legendary Creature — Avatar Wizard", hobA11SarumanOracle)
	creature := handCardForTest(opp, "An Ogre", "Creature — Ogre", "")
	sorcery := handCardForTest(opp, "A Sorcery", "Sorcery", "")
	aura := handCardForTest(opp, "An Aura", "Enchantment — Aura", "")
	boltWardedCreature(t, g, opp, warded)

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	c := discardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the payer is not asked to discard")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{creature}); err == nil {
		t.Fatal("a creature card paid the ward")
	}
	answerDiscard(t, g, opp.ID, aura)
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(aura) || !opp.Hand.Contains(sorcery) || !opp.Hand.Contains(creature) {
		t.Error("only the chosen Aura is discarded")
	}
	if !boltResolved(g, warded) {
		t.Error("the ward was paid, so the Lightning Bolt should have resolved")
	}
}

// A hand with no such card cannot pay at all: countered with no prompt.
func TestHobA11SarumanWardCountersTheSpellWhenYouHoldNoSuchCard(t *testing.T) {
	g, _, opp, warded := wardTable(t, "Saruman of Many Colors", "Legendary Creature — Avatar Wizard", hobA11SarumanOracle)
	handCardForTest(opp, "An Ogre", "Creature — Ogre", "")
	handCardForTest(opp, "A Land", "Basic Land — Forest", "")
	bolt := boltWardedCreature(t, g, opp, warded)

	if prompt := passUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatal("a payer who cannot pay is not asked")
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || boltResolved(g, warded) {
		t.Error("the Bolt should have been countered")
	}
}

func TestHobA11SarumanWardDeclinedCountersTheSpell(t *testing.T) {
	g, _, opp, warded := wardTable(t, "Saruman of Many Colors", "Legendary Creature — Avatar Wizard", hobA11SarumanOracle)
	sorcery := handCardForTest(opp, "A Sorcery", "Sorcery", "")
	bolt := boltWardedCreature(t, g, opp, warded)

	prompt := passUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || !opp.Hand.Contains(sorcery) || boltResolved(g, warded) {
		t.Error("declining counters the Bolt and discards nothing")
	}
}

// hobA11SarumanTable puts Saruman under the active player and casts their
// first instant (mana value 1); the next cast is the second spell.
func hobA11SarumanTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	saruman := pushDiesCreatureForTest(g, me.ID, "Saruman of Many Colors", hobA11SarumanOracle, "Legendary Creature — Avatar Wizard", 5, 4)
	toMainForCost(t, g)
	first := handSpell(me, "First Spell", "Instant", "{1}")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{}); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if saruman == uuid.Nil || len(g.PendingTriggers) != 0 {
		t.Fatal("the first spell of the turn does not trigger Saruman")
	}
	return g, me, opp, saruman
}

// Each opponent mills two; the reflexive trigger exiles an instant, sorcery or
// enchantment card from an opponent's graveyard with mana value at most the
// spell's, and the copy is cast for free.
func TestHobA11SarumanMillsExilesCopiesAndCastsTheCopyFree(t *testing.T) {
	g, me, opp, _ := hobA11SarumanTable(t)
	bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	tooBig := gyCard(opp, "Big Instant", "Instant", "{3}{R}{R}", "")
	ogre := gyCard(opp, "Dead Ogre", "Creature — Ogre", "{1}", "")
	mine := gyCard(me, "My Own Sorcery", "Sorcery", "{1}", "")
	oppLibrary := opp.Library.Size()

	second := handSpell(me, "Second Spell", "Instant", "{1}{R}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p != me && p.Graveyard.Size() < 2 && p != opp {
			t.Errorf("an opponent did not mill two (graveyard %d)", p.Graveyard.Size())
		}
	}
	if opp.Library.Size() != oppLibrary-2 {
		t.Fatalf("opponent library %d → %d, want -2", oppLibrary, opp.Library.Size())
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger asks for a target once cards were milled")
	}
	if !hasID(p.PickTargetCards, bolt) {
		t.Error("the Bolt (mana value 1) is a legal target")
	}
	for name, id := range map[string]uuid.UUID{"a mana value 5 instant": tooBig, "a creature": ogre, "my own graveyard's sorcery": mine} {
		if hasID(p.PickTargetCards, id) {
			t.Errorf("%s is not a legal target", name)
		}
	}
	pickCard(t, g, me.ID, bolt)
	passPriorityAroundTable(t, g)

	if !inExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	life := opp.Life
	// Strict mana with an empty pool: only a free cast gets through.
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy without paying {R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3: the copy resolves as Lightning Bolt", life, opp.Life)
	}
	if !inExile(g, bolt) {
		t.Error("the original stays exiled")
	}
}

// The ceiling is the second spell's mana value: a mana value 2 spell reaches a
// mana value 2 card but not a 3.
func TestHobA11SarumanCeilingIsTheSecondSpellsManaValue(t *testing.T) {
	g, me, opp, _ := hobA11SarumanTable(t)
	two := gyCard(opp, "Two Sorcery", "Sorcery", "{1}{U}", "")
	three := gyCard(opp, "Three Sorcery", "Sorcery", "{2}{U}", "")
	aura := gyCard(opp, "Two Aura", "Enchantment — Aura", "{1}{W}", "")

	second := handSpell(me, "Second Spell", "Instant", "{1}{R}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if !hasID(p.PickTargetCards, two) || !hasID(p.PickTargetCards, aura) || hasID(p.PickTargetCards, three) {
		t.Errorf("legal targets %v: mana value 2 or less, enchantments included", p.PickTargetCards)
	}
}

// With nothing to mill there is no "one or more cards milled", so no trigger.
func TestHobA11SarumanWithNoCardsMilledDoesNothing(t *testing.T) {
	g, me, opp, _ := hobA11SarumanTable(t)
	gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	for _, p := range g.Seats {
		if p != me {
			p.Library.Cards = nil
		}
	}
	second := handSpell(me, "Second Spell", "Instant", "{1}{R}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no cards were milled, so there is no reflexive trigger")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies made", n)
	}
}

// Only the second spell each turn triggers it, and not a third.
func TestHobA11SarumanTriggersOnTheSecondSpellOnly(t *testing.T) {
	g, me, opp, _ := hobA11SarumanTable(t)
	second := handSpell(me, "Second Spell", "Instant", "{1}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	milled := opp.Graveyard.Size()
	if milled < 2 {
		t.Fatalf("the second spell should mill two, graveyard %d", milled)
	}
	third := handSpell(me, "Third Spell", "Instant", "{1}")
	if err := g.CastSpell(me.ID, third, game.CastSpellParams{}); err != nil {
		t.Fatalf("third spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != milled {
		t.Errorf("the third spell milled again (graveyard %d → %d)", milled, opp.Graveyard.Size())
	}
}

// --- Shadow of the Enemy ---------------------------------------------------------

func TestHobA11ShadowExilesTheCreatureCardsAndLetsYouCastThemWithAnyMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ogre := gyCard(opp, "Dead Ogre", "Creature — Ogre", "{G}{G}", "")
	elf := gyCard(opp, "Dead Elf", "Creature — Elf", "{1}{G}", "")
	sorcery := gyCard(opp, "A Sorcery", "Sorcery", "{1}", "")
	mine := gyCard(me, "My Bear", "Creature — Bear", "{1}{G}", "")

	castCatalogSpell(t, g, "Shadow of the Enemy", "Sorcery", hobA11ShadowOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)

	if !inExile(g, ogre) || !inExile(g, elf) {
		t.Fatal("every creature card in the graveyard is exiled")
	}
	if !opp.Graveyard.Contains(sorcery) || !me.Graveyard.Contains(mine) {
		t.Error("only the creature cards of the target player's graveyard are exiled")
	}
	g.ReadSnapshot(func() {
		perm := g.CastPermissionOnCardByIDForEffect(ogre)
		if perm.Player != me.ID || !perm.CastOnly || !perm.AnyType {
			t.Errorf("permission = %+v, want the caster's, cast-only, any type of mana", perm)
		}
	})
	// Black and blue mana pay a green cost; strict mana is on.
	floatForTest(g, me, "UB")
	if err := g.CastSpell(me.ID, ogre, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the Ogre with {U}{B} for {G}{G}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) {
		t.Error("the Ogre should have been cast and resolved under the caster's control")
	}
	if c, _ := battlefieldCard(g, ogre); c.Controller != me.ID {
		t.Error("the caster controls the creature they cast")
	}
	if !inExile(g, elf) {
		t.Error("the other card stays in exile, still castable")
	}
}

// You may target yourself, and a graveyard with no creature cards is fine.
func TestHobA11ShadowCanTargetYourOwnGraveyardAndAnEmptyOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := gyCard(me, "My Bear", "Creature — Bear", "{1}{G}", "")
	castCatalogSpell(t, g, "Shadow of the Enemy", "Sorcery", hobA11ShadowOracle, []game.TargetRef{playerTarget(me.ID)})
	passPriorityAroundTable(t, g)
	if !inExile(g, mine) {
		t.Error("your own creature card is exiled when you target yourself")
	}

	g2 := newCatalogGame(t)
	opp2 := g2.Seats[(g2.Turn.ActiveSeat+1)%4]
	castCatalogSpell(t, g2, "Shadow of the Enemy", "Sorcery", hobA11ShadowOracle, []game.TargetRef{playerTarget(opp2.ID)})
	passPriorityAroundTable(t, g2)
	if len(g2.Exile.Cards) != 0 {
		t.Errorf("%d cards in exile from an empty graveyard", len(g2.Exile.Cards))
	}
}
