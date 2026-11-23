package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// --- Gandalf, Spark Starter --------------------------------------------------

func TestHobA11GandalfDividesThreeDamageOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	gandalf := effectstest.CastCatalogSpell(t, g, "Gandalf, Spark Starter", "Legendary Creature — Avatar Wizard", hobA11GandalfOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, gandalf, "reach") {
		t.Error("Gandalf has reach")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{a: 2, b: 1}, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 2 || d2 != 1 {
		t.Errorf("damage %d/%d, want 2/1", d1, d2)
	}
}

// A single target takes all three, and a player is a legal target.
func TestHobA11GandalfCanPutAllThreeOnOnePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Gandalf, Spark Starter", "Legendary Creature — Avatar Wizard", hobA11GandalfOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asks for targets")
	}
	refs := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, refs, map[uuid.UUID]int{opp.ID: 3}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
}

// A split that does not add up to three is refused.
func TestHobA11GandalfRefusesASplitThatIsNotThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	effectstest.CastCatalogSpell(t, g, "Gandalf, Spark Starter", "Legendary Creature — Avatar Wizard", hobA11GandalfOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
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
	bolg := effectstest.CastCatalogSpell(t, g, "Bolg of the North", "Legendary Creature — Goblin Soldier", hobA11BolgOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return bolg
}

// Sacrificing a 4-power creature to hit a 1/2: 4 damage, 2 of it excess, so
// the Goblin amass puts two counters on a new Army.
func TestHobA11BolgSacrificeDealsItsPowerAndAmassesTheExcess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 4, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 2)
	castHobA11Bolg(t, g)

	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the sacrifice happens before the reflexive trigger asks for a target")
	}
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(victim) {
		t.Error("4 damage kills a 1/2")
	}
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2 (4 damage less the 2 it needed)", got)
	}
	if !army.HasSubtype("Goblin") {
		t.Error("the amassed Army is a Goblin")
	}
}

// The damage is the creature's power as it last stood: an anthem counts.
func TestHobA11BolgCountsAnAnthemInThePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 1)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: effectstest.GloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	castHobA11Bolg(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("Army counters = %d, want 1 (a 2-power sacrifice into a 1/1)", got)
	}
}

// Exactly lethal is no excess: the creature dies and nothing is amassed.
func TestHobA11BolgWithNoExcessAmassesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 3)
	castHobA11Bolg(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 4, 4)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 2)
	castHobA11Bolg(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(victim) {
		t.Error("declining changes nothing")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no sacrifice, no reflexive trigger")
	}
}

// "Another creature": Bolg himself is not offered to be sacrificed, and he is
// not "another target creature" for his own damage either.
func TestHobA11BolgOffersOnlyOtherCreaturesAndCannotHitHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Ogre", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 2)
	bolg := castHobA11Bolg(t, g)

	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	for _, id := range c.ChooseCards {
		if id == bolg {
			t.Error("Bolg cannot sacrifice himself")
		}
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
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
		effectstest.B12Permanent(g, owner, "Mountain", "Basic Land — Mountain")
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Desert Were-Worm", TypeLine: "Creature — Dragon Wurm",
		OracleID: hobA11WormOracle, Power: 0, Toughness: 5, Owner: owner, Controller: owner,
	})
}

func TestHobA11WormGetsTwoPowerPerMountainYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	worm := hobA11Worm(g, me.ID, 0)
	if p := effectstest.E2Card(t, g, worm).CurrentPower(); p != 0 {
		t.Errorf("no Mountains: power %d, want 0", p)
	}
	for i := 0; i < 3; i++ {
		effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	effectstest.B12Permanent(g, opp.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	c := effectstest.E2Card(t, g, worm)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 5 {
		t.Errorf("three Mountains: %d/%d, want 6/5 (an opponent's Mountain and a Forest do not count)", c.CurrentPower(), c.CurrentToughness())
	}
}

// Twelve power attacking: the attackers untap, one more combat follows, and
// attacking with twelve again in it does not trigger a third.
func TestHobA11WormUntapsAttackersAndAddsOneCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	worm := hobA11Worm(g, me.ID, 5)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	start := len(g.Events)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{worm, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountResolved(g, start, hobA11WormLabel) != 1 {
		t.Fatal("the Worm's trigger did not resolve once")
	}
	for _, id := range []uuid.UUID{worm, bear} {
		if c, _ := effectstest.BattlefieldCard(g, id); c.Tapped {
			t.Errorf("%s is still tapped: all attacking creatures untap", c.Name)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after the first combat: %+v, want a second combat", next)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{worm, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker in the second combat: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, hobA11WormLabel); n != 1 {
		t.Errorf("the trigger resolved %d times: it is once per turn, so the second combat adds nothing", n)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after the second combat: %+v, want the postcombat main phase", next)
	}
}

// Total power counts the creatures that ATTACK: a Worm of ten attacking alone
// is short of twelve even with a 2/2 at home.
func TestHobA11WormCountsOnlyTheAttackersPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	worm := hobA11Worm(g, me.ID, 5)
	effectstest.PushVanillaCreature(g, me.ID, "Bear at home", 2, 2)
	start := len(g.Events)

	effectstest.DeclareAttack(t, g, opp.ID, worm)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, hobA11WormLabel); n != 0 {
		t.Errorf("the trigger resolved %d times with only 10 power attacking", n)
	}
	if c, _ := effectstest.BattlefieldCard(g, worm); !c.Tapped {
		t.Error("the Worm stays tapped: nothing untapped it")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if next := effectstest.StepOnce(t, g); next.Step != game.StepPostcombatMain {
		t.Errorf("after combat: %+v, want the postcombat main phase (no additional combat)", next)
	}
}

// Twelve from several creatures is one trigger, not one per attacker.
func TestHobA11WormTriggersOncePerAttackNotPerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	worm := hobA11Worm(g, me.ID, 3)
	a := effectstest.PushVanillaCreature(g, me.ID, "Ogre A", 3, 3)
	b := effectstest.PushVanillaCreature(g, me.ID, "Ogre B", 3, 3)
	start := len(g.Events)

	effectstest.DeclareAttack(t, g, opp.ID, worm, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountResolved(g, start, hobA11WormLabel); n != 1 {
		t.Errorf("the trigger resolved %d times, want 1 (6 + 3 + 3 = 12)", n)
	}
}

// --- Bilbo, Luckwearer // Burglar's Plot ----------------------------------------

var (
	hobA11BilboCreature = effectstest.HobA6Face("Bilbo, Luckwearer", "Legendary Creature — Halfling Rogue", "{1}{U}", 1, 1, "U")
	hobA11BurglarsPlot  = effectstest.HobA6Face("Burglar's Plot", "Sorcery — Adventure", "{4}{U}", 0, 0, "U")
)

func TestHobA11BilboCantBeBlockedAndLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bilbo := effectstest.B12Push(g, me.ID, "Bilbo, Luckwearer", "Legendary Creature — Halfling Rogue", hobA11BilboOracle, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.PushLibraryTopForTest(me, "Drawn Card")
	handBefore := len(me.Hand.Cards)
	life := opp.Life

	effectstest.DeclareAttack(t, g, opp.ID, bilbo)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bilbo); err == nil {
		t.Fatal("a creature blocked an unblockable Bilbo")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want -1", life, opp.Life)
	}
	if len(me.Hand.Cards) != handBefore+1 {
		t.Fatalf("hand = %d, want %d: the draw comes before the discard", len(me.Hand.Cards), handBefore+1)
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != handBefore {
		t.Errorf("hand = %d, want %d after draw then discard", got, handBefore)
	}
}

func hobA11Control(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Controller
}

// An artifact and an artifact creature share a card type: control swaps, and
// the Adventure goes to exile to be cast later.
func TestHobA11BurglarsPlotExchangesTwoPermanentsThatShareAType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := effectstest.HobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	mine := effectstest.B12Permanent(g, me.ID, "Mighty Hammer", "Artifact — Equipment")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(mine), effectstest.HobA6Card(theirs)})

	if hobA11Control(t, g, mine) != opp.ID || hobA11Control(t, g, theirs) != me.ID {
		t.Error("the Hammer and the Golem should have changed hands")
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobA11BurglarsPlotWorksOnTwoCreaturesAndTheOrderOfPicksDoesNotMatter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := effectstest.HobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(theirs), effectstest.HobA6Card(mine)})

	if hobA11Control(t, g, mine) != opp.ID || hobA11Control(t, g, theirs) != me.ID {
		t.Error("the creatures should have changed hands")
	}
}

// A land is never a target, and the two must share a type: an enchantment and
// a creature do not, an artifact creature and an artifact land (a land) are
// refused too.
func TestHobA11BurglarsPlotRefusesTargetsThatShareNoTypeOrAreLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := effectstest.HobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	aura := effectstest.B12Permanent(g, me.ID, "My Wall of Words", "Enchantment")
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	vault := effectstest.B12Permanent(g, me.ID, "Darksteel Citadel", "Artifact Land")
	golem := effectstest.B12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)

	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(aura), effectstest.HobA6Card(ogre)}); err == nil {
		t.Error("an enchantment and a creature share no card type")
	}
	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(vault), effectstest.HobA6Card(golem)}); err == nil {
		t.Error("a land is not a legal target, even one that shares a type")
	}
	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(ogre), effectstest.HobA6Card(ogre)}); err == nil {
		t.Error("the two targets are different permanents")
	}
}

// All or nothing (CR 701.12a): when one target is gone as the spell resolves
// nothing changes hands.
func TestHobA11BurglarsPlotDoesNothingWhenATargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me, id := effectstest.HobA6HandAdventure(t, g, hobA11BilboOracle, hobA11BilboCreature, hobA11BurglarsPlot)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)

	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(mine), effectstest.HobA6Card(theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.MoveToGraveyard(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)

	if hobA11Control(t, g, mine) != me.ID {
		t.Error("my Bear changed hands although its partner was gone")
	}
}

// --- Glamdring, Foe-hammer // Gleam of Death -----------------------------------

var (
	hobA11GlamdringFace = effectstest.HobA6Face("Glamdring, Foe-hammer", "Legendary Artifact — Equipment", "{2}", 0, 0)
	hobA11GleamOfDeath  = effectstest.HobA6Face("Gleam of Death", "Sorcery — Adventure", "{3}{U}", 0, 0, "U")
)

func hobA11GlamdringOnCreature(t *testing.T, g *game.Game, me *game.Player, power int) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	glamdring := effectstest.PushCatalogPermanent(g, me.ID, "Glamdring, Foe-hammer", "Legendary Artifact — Equipment", hobA11GlamdringOracle, false)
	host := effectstest.B12Creature(g, me.ID, "Wielder", "Creature — Human", power, 3)
	effectstest.EquipTo(t, g, me.ID, glamdring, host)
	return glamdring
}

func TestHobA11GlamdringDiscountsYourInstantsAndSorceriesByTheWielderPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hobA11GlamdringOnCreature(t, g, me, 3)

	if got := effectstest.PriceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 2 {
		t.Errorf("sorcery of mana value 5: %d, want 2 (3 less)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Big Instant", "Instant", "{4}{U}"); got != 2 {
		t.Errorf("instant of mana value 5: %d, want 2 (3 less)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Cheap Instant", "Instant", "{1}{R}"); got != 1 {
		t.Errorf("instant of mana value 2: %d, want 1 (the generic is all it can spend against)", got)
	}
	if got := effectstest.PriceInHand(t, g, me, "Big Ogre", "Creature — Ogre", "{4}{R}"); got != 5 {
		t.Errorf("creature spell: %d, want 5 (not an instant or sorcery)", got)
	}
	if got := effectstest.PriceInHand(t, g, opp, "Their Sorcery", "Sorcery", "{4}{R}"); got != 5 {
		t.Errorf("an opponent's sorcery: %d, want 5 (spells YOU cast)", got)
	}
}

// X moves with the wielder's power, and an unattached Glamdring gives nothing.
func TestHobA11GlamdringFollowsThePowerAndNeedsAWielder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	glamdring := effectstest.PushCatalogPermanent(g, me.ID, "Glamdring, Foe-hammer", "Legendary Artifact — Equipment", hobA11GlamdringOracle, false)
	if got := effectstest.PriceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 5 {
		t.Errorf("unattached: %d, want 5", got)
	}
	host := effectstest.B12Creature(g, me.ID, "Wielder", "Creature — Human", 1, 3)
	effectstest.EquipTo(t, g, me.ID, glamdring, host)
	if got := effectstest.PriceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 4 {
		t.Errorf("1-power wielder: %d, want 4", got)
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
		OracleID: effectstest.GloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if got := effectstest.PriceInHand(t, g, me, "Big Sorcery", "Sorcery", "{4}{R}"); got != 3 {
		t.Errorf("an anthem makes the wielder a 2-power creature: %d, want 3", got)
	}
}

func TestHobA11GlamdringEquipsForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	glamdring := effectstest.PushCatalogPermanent(g, me.ID, "Glamdring, Foe-hammer", "Legendary Artifact — Equipment", hobA11GlamdringOracle, false)
	host := effectstest.B12Creature(g, me.ID, "Wielder", "Creature — Human", 1, 3)
	if got := effectstest.EquipPrice(t, g, me.ID, glamdring, host); got != "{2}" {
		t.Errorf("equip costs %s, want {2}", got)
	}
}

// Gleam of Death mills six and takes every instant and sorcery among them,
// leaving the rest in the graveyard and the seventh card in the library.
func TestHobA11GleamOfDeathMillsSixAndTakesTheInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	activeSeat := g.Seats[g.Turn.ActiveSeat]
	activeSeat.Library.Cards = nil
	deep := uuid.New()
	var ids []uuid.UUID
	for _, c := range []struct{ Name, TypeLine string }{
		{"Deep Sorcery", "Sorcery"}, // the seventh card: stays in the library
		{"A Land", "Basic Land — Island"},
		{"Instant One", "Instant"},
		{"An Ogre", "Creature — Ogre"},
		{"Sorcery One", "Sorcery"},
		{"An Aura", "Enchantment — Aura"},
		{"Instant Two", "Instant"},
	} {
		id := uuid.New()
		if c.Name == "Deep Sorcery" {
			id = deep
		}
		ids = append(ids, id)
		activeSeat.Library.PushTop(game.Card{InstanceID: id, Name: c.Name, TypeLine: c.TypeLine, Owner: activeSeat.ID, Controller: activeSeat.ID})
	}
	// Top first: Instant Two, An Aura, Sorcery One, An Ogre, Instant One, A Land | Deep Sorcery.
	me, id := effectstest.HobA6HandAdventure(t, g, hobA11GlamdringOracle, hobA11GlamdringFace, hobA11GleamOfDeath)
	effectstest.HobA6MustCast(t, g, me, id, 1, nil)

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
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

// --- Bilbo's Gambit -------------------------------------------------------------

// hobA11ShockOnTheStack has the opponent cast a Shock at the caster and
// returns it.
func hobA11ShockOnTheStack(t *testing.T, g *game.Game, caster, opponent *game.Player) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g) // walking the turn clears the stack, so get there first
	shock := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant",
		OracleID: effectstest.B5ShockOracle, Owner: opponent.ID, Controller: opponent.ID,
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
	who.Hand.PushTop(game.Card{InstanceID: id, Name: "Shock", TypeLine: "Instant", OracleID: effectstest.B5ShockOracle, Owner: who.ID, Controller: who.ID})
	return g.CastSpell(who.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: who.ID}}})
}

// Promised: the opponent gets a Treasure, the spell goes back to its owner's
// hand, and nobody, the caster included, can cast a spell for the rest of the
// turn.
func TestHobA11GambitWithGiftBouncesTheSpellAndLocksTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	third := g.Seats[2]
	shock := hobA11ShockOnTheStack(t, g, caster, opp)

	effectstest.MustCastWithGift(t, g, "Bilbo's Gambit", "Instant", hobA11GambitOracle, effectstest.CardTarget(shock), opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Stack.Contains(shock) || !opp.Hand.Contains(shock) {
		t.Error("the Shock should be back in its owner's hand")
	}
	if opp.Graveyard.Contains(shock) {
		t.Error("it is returned, not countered")
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Treasure"); n != 1 {
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
	g := effectstest.NewCatalogGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	shock := hobA11ShockOnTheStack(t, g, caster, opp)

	effectstest.MustCastWithGift(t, g, "Bilbo's Gambit", "Instant", hobA11GambitOracle, effectstest.CardTarget(shock), uuid.Nil)
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Hand.Contains(shock) {
		t.Error("the Shock should be back in its owner's hand")
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Treasure"); n != 0 {
		t.Errorf("the opponent has %d Treasures, want none", n)
	}
	if err := hobA11CastAnything(g, caster); err != nil {
		t.Errorf("without the gift anyone may still cast: %v", err)
	}
}

// The gift is made before the other effects, and it is the CHOSEN opponent's.
func TestHobA11GambitGivesTheTreasureToTheChosenOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	shock := hobA11ShockOnTheStack(t, g, caster, opp)

	effectstest.MustCastWithGift(t, g, "Bilbo's Gambit", "Instant", hobA11GambitOracle, effectstest.CardTarget(shock), other.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if n, _ := effectstest.CountTokensNamed(g, other.ID, "Treasure"); n != 1 {
		t.Errorf("the chosen opponent has %d Treasures, want 1", n)
	}
	if n, _ := effectstest.CountTokensNamed(g, opp.ID, "Treasure"); n != 0 {
		t.Errorf("the spell's owner has %d Treasures, want none", n)
	}
}

// --- Woodland Weavemaster ---------------------------------------------------------

func TestHobA11WeavemasterTapsForManaEqualToItsPowerRestrictedToElves(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	weaver := effectstest.B12Push(g, me.ID, "Woodland Weavemaster", "Creature — Elf Druid", hobA11WeavemasterOracle, 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, weaver, "vigilance") {
		t.Error("Weavemaster has vigilance")
	}
	if err := g.ActivateManaAbility(me.ID, weaver, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("tap Weavemaster: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
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
	g, me, opp := effectstest.SpendTable(t)
	weaver := effectstest.B12Push(g, me.ID, "Woodland Weavemaster", "Creature — Elf Druid", hobA11WeavemasterOracle, 1, 2)

	effectstest.B12Creature(g, opp.ID, "Their Elf", "Creature — Elf Scout", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.E2Card(t, g, weaver).CurrentPower(); p != 1 {
		t.Fatalf("an opponent's Elf entering changed the power to %d", p)
	}
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.E2Card(t, g, weaver).CurrentPower(); p != 1 {
		t.Fatalf("a Bear entering changed the power to %d: only an Elf counts", p)
	}
	elf := effectstest.HandSpell(me, "Elf", "Creature — Elf Warrior", "")
	if err := g.CastSpell(me.ID, elf, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast an Elf: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.E2Card(t, g, weaver)
	if c.CurrentPower() != 2 || c.CurrentToughness() != 3 {
		t.Fatalf("after another Elf: %d/%d, want 2/3", c.CurrentPower(), c.CurrentToughness())
	}
	if err := g.ActivateManaAbility(me.ID, weaver, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("tap Weavemaster: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("pool = %v, want {U}{U}: X mana of one colour", got)
	}
}

// --- Saruman of Many Colors -------------------------------------------------------

// Paying the ward means discarding an enchantment, instant or sorcery card:
// a creature card is refused, and the spell then resolves.
func TestHobA11SarumanWardTakesAnEnchantmentInstantOrSorceryCard(t *testing.T) {
	g, _, opp, warded := effectstest.WardTable(t, "Saruman of Many Colors", "Legendary Creature — Avatar Wizard", hobA11SarumanOracle)
	creature := effectstest.HandCardForTest(opp, "An Ogre", "Creature — Ogre", "")
	sorcery := effectstest.HandCardForTest(opp, "A Sorcery", "Sorcery", "")
	aura := effectstest.HandCardForTest(opp, "An Aura", "Enchantment — Aura", "")
	effectstest.BoltWardedCreature(t, g, opp, warded)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the payer is not asked to discard")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{creature}); err == nil {
		t.Fatal("a creature card paid the ward")
	}
	effectstest.AnswerDiscard(t, g, opp.ID, aura)
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(aura) || !opp.Hand.Contains(sorcery) || !opp.Hand.Contains(creature) {
		t.Error("only the chosen Aura is discarded")
	}
	if !effectstest.BoltResolved(g, warded) {
		t.Error("the ward was paid, so the Lightning Bolt should have resolved")
	}
}

// A hand with no such card cannot pay at all: countered with no prompt.
func TestHobA11SarumanWardCountersTheSpellWhenYouHoldNoSuchCard(t *testing.T) {
	g, _, opp, warded := effectstest.WardTable(t, "Saruman of Many Colors", "Legendary Creature — Avatar Wizard", hobA11SarumanOracle)
	effectstest.HandCardForTest(opp, "An Ogre", "Creature — Ogre", "")
	effectstest.HandCardForTest(opp, "A Land", "Basic Land — Forest", "")
	bolt := effectstest.BoltWardedCreature(t, g, opp, warded)

	if prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID); prompt != nil {
		t.Fatal("a payer who cannot pay is not asked")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || effectstest.BoltResolved(g, warded) {
		t.Error("the Bolt should have been countered")
	}
}

func TestHobA11SarumanWardDeclinedCountersTheSpell(t *testing.T) {
	g, _, opp, warded := effectstest.WardTable(t, "Saruman of Many Colors", "Legendary Creature — Avatar Wizard", hobA11SarumanOracle)
	sorcery := effectstest.HandCardForTest(opp, "A Sorcery", "Sorcery", "")
	bolt := effectstest.BoltWardedCreature(t, g, opp, warded)

	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) || !opp.Hand.Contains(sorcery) || effectstest.BoltResolved(g, warded) {
		t.Error("declining counters the Bolt and discards nothing")
	}
}

// hobA11SarumanTable puts Saruman under the active player and casts their
// first instant (mana value 1); the next cast is the second spell.
func hobA11SarumanTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	saruman := effectstest.PushDiesCreatureForTest(g, me.ID, "Saruman of Many Colors", hobA11SarumanOracle, "Legendary Creature — Avatar Wizard", 5, 4)
	effectstest.ToMainForCost(t, g)
	first := effectstest.HandSpell(me, "First Spell", "Instant", "{1}")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{}); err != nil {
		t.Fatalf("first spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	bolt := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	tooBig := effectstest.GyCard(opp, "Big Instant", "Instant", "{3}{R}{R}", "")
	ogre := effectstest.GyCard(opp, "Dead Ogre", "Creature — Ogre", "{1}", "")
	mine := effectstest.GyCard(me, "My Own Sorcery", "Sorcery", "{1}", "")
	oppLibrary := opp.Library.Size()

	second := effectstest.HandSpell(me, "Second Spell", "Instant", "{1}{R}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p != me && p.Graveyard.Size() < 2 && p != opp {
			t.Errorf("an opponent did not mill two (graveyard %d)", p.Graveyard.Size())
		}
	}
	if opp.Library.Size() != oppLibrary-2 {
		t.Fatalf("opponent library %d → %d, want -2", oppLibrary, opp.Library.Size())
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger asks for a target once cards were milled")
	}
	if !effectstest.HasID(p.PickTargetCards, bolt) {
		t.Error("the Bolt (mana value 1) is a legal target")
	}
	for name, id := range map[string]uuid.UUID{"a mana value 5 instant": tooBig, "a creature": ogre, "my own graveyard's sorcery": mine} {
		if effectstest.HasID(p.PickTargetCards, id) {
			t.Errorf("%s is not a legal target", name)
		}
	}
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	life := opp.Life
	// Strict mana with an empty pool: only a free cast gets through.
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy without paying {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3: the copy resolves as Lightning Bolt", life, opp.Life)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the original stays exiled")
	}
}

// The ceiling is the second spell's mana value: a mana value 2 spell reaches a
// mana value 2 card but not a 3.
func TestHobA11SarumanCeilingIsTheSecondSpellsManaValue(t *testing.T) {
	g, me, opp, _ := hobA11SarumanTable(t)
	two := effectstest.GyCard(opp, "Two Sorcery", "Sorcery", "{1}{U}", "")
	three := effectstest.GyCard(opp, "Three Sorcery", "Sorcery", "{2}{U}", "")
	aura := effectstest.GyCard(opp, "Two Aura", "Enchantment — Aura", "{1}{W}", "")

	second := effectstest.HandSpell(me, "Second Spell", "Instant", "{1}{R}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if !effectstest.HasID(p.PickTargetCards, two) || !effectstest.HasID(p.PickTargetCards, aura) || effectstest.HasID(p.PickTargetCards, three) {
		t.Errorf("legal targets %v: mana value 2 or less, enchantments included", p.PickTargetCards)
	}
}

// With nothing to mill there is no "one or more cards milled", so no trigger.
func TestHobA11SarumanWithNoCardsMilledDoesNothing(t *testing.T) {
	g, me, opp, _ := hobA11SarumanTable(t)
	effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	for _, p := range g.Seats {
		if p != me {
			p.Library.Cards = nil
		}
	}
	second := effectstest.HandSpell(me, "Second Spell", "Instant", "{1}{R}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no cards were milled, so there is no reflexive trigger")
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies made", n)
	}
}

// Only the second spell each turn triggers it, and not a third.
func TestHobA11SarumanTriggersOnTheSecondSpellOnly(t *testing.T) {
	g, me, opp, _ := hobA11SarumanTable(t)
	second := effectstest.HandSpell(me, "Second Spell", "Instant", "{1}")
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{}); err != nil {
		t.Fatalf("second spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	milled := opp.Graveyard.Size()
	if milled < 2 {
		t.Fatalf("the second spell should mill two, graveyard %d", milled)
	}
	third := effectstest.HandSpell(me, "Third Spell", "Instant", "{1}")
	if err := g.CastSpell(me.ID, third, game.CastSpellParams{}); err != nil {
		t.Fatalf("third spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != milled {
		t.Errorf("the third spell milled again (graveyard %d → %d)", milled, opp.Graveyard.Size())
	}
}

// --- Shadow of the Enemy ---------------------------------------------------------

func TestHobA11ShadowExilesTheCreatureCardsAndLetsYouCastThemWithAnyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ogre := effectstest.GyCard(opp, "Dead Ogre", "Creature — Ogre", "{G}{G}", "")
	elf := effectstest.GyCard(opp, "Dead Elf", "Creature — Elf", "{1}{G}", "")
	sorcery := effectstest.GyCard(opp, "A Sorcery", "Sorcery", "{1}", "")
	mine := effectstest.GyCard(me, "My Bear", "Creature — Bear", "{1}{G}", "")

	effectstest.CastCatalogSpell(t, g, "Shadow of the Enemy", "Sorcery", hobA11ShadowOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, ogre) || !effectstest.InExile(g, elf) {
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
	effectstest.FloatForTest(g, me, "UB")
	if err := g.CastSpell(me.ID, ogre, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the Ogre with {U}{B} for {G}{G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) {
		t.Error("the Ogre should have been cast and resolved under the caster's control")
	}
	if c, _ := effectstest.BattlefieldCard(g, ogre); c.Controller != me.ID {
		t.Error("the caster controls the creature they cast")
	}
	if !effectstest.InExile(g, elf) {
		t.Error("the other card stays in exile, still castable")
	}
}

// You may target yourself, and a graveyard with no creature cards is fine.
func TestHobA11ShadowCanTargetYourOwnGraveyardAndAnEmptyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.GyCard(me, "My Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Shadow of the Enemy", "Sorcery", hobA11ShadowOracle, []game.TargetRef{effectstest.PlayerTarget(me.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, mine) {
		t.Error("your own creature card is exiled when you target yourself")
	}

	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[(g2.Turn.ActiveSeat+1)%4]
	effectstest.CastCatalogSpell(t, g2, "Shadow of the Enemy", "Sorcery", hobA11ShadowOracle, []game.TargetRef{effectstest.PlayerTarget(opp2.ID)})
	effectstest.PassPriorityAroundTable(t, g2)
	if len(g2.Exile.Cards) != 0 {
		t.Errorf("%d cards in exile from an empty graveyard", len(g2.Exile.Cards))
	}
}
