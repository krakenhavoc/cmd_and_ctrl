package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b6_test.go — Planechase plane cards, batch 6 (ADR 0216 slice 4,
// #4870): each plane is played face up on a Planechase table
// (planes_harness_test.go) and its static, triggered and chaos abilities are
// driven by the engine.

var (
	planeGreatAerie = planeFixture{"00f00001-bd84-4dbf-a707-f4b1549100d4", "The Great Aerie", "Plane — Tarkir"}
)

func TestGreatAerieBolstersOnEntryAndEachUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGreatAerie)
	g, _ := newPlaneGame(t, planeGreatAerie)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	small := pushVanillaCreature(g, me.ID, "Small", 1, 1)
	big := pushVanillaCreature(g, me.ID, "Big", 4, 4)
	planeswalkToPlane(t, g, planeGreatAerie.Oracle)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, small, game.CounterPlusOne); got != 3 {
		t.Errorf("least-toughness creature has %d counters, want 3", got)
	}
	if got := countersOn(g, big, game.CounterPlusOne); got != 0 {
		t.Errorf("the big creature got %d counters", got)
	}
}

func TestGreatAerieChaosDealsToughnessBothWays(t *testing.T) {
	g, _ := newPlaneGame(t, planeGreatAerie)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := pushVanillaCreature(g, me.ID, "Wall", 0, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Ogre", 5, 2)
	rollBlank(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, mine)
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	// Wall deals 3 to Ogre (toughness 2: dies); Ogre deals 2 to Wall (survives).
	g.WithWriteLock(func() {
		if _, ok := g.LookupCardForEffect(theirs); ok && g.FindCardZoneForEffect(theirs).Kind == game.ZoneBattlefield {
			t.Error("the opposing creature survived 3 damage")
		}
		if z := g.FindCardZoneForEffect(mine); z == nil || z.Kind != game.ZoneBattlefield {
			t.Error("the wall died to 2 damage")
		}
	})
}

var planeGreatForest = planeFixture{"8c56bb41-a8d8-4b0f-bbb7-be8466dc9ed1", "The Great Forest", "Plane — Lorwyn"}

func TestGreatForestCreaturesAssignDamageByToughness(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGreatForest)
	g, _ := newPlaneGame(t, planeGreatForest)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	wall := pushVanillaCreature(g, me.ID, "Wall", 0, 4)
	glass := pushVanillaCreature(g, me.ID, "Glass Cannon", 5, 1)
	life := opp.Life
	pr7bAttack(t, g, opp.ID, wall, glass)
	advanceTo(t, g, game.StepEndCombat)
	if got := life - opp.Life; got != 5 {
		t.Errorf("the defender lost %d life, want 4 (toughness of the wall) + 1 (toughness of the cannon)", got)
	}
}

func TestGreatForestChaosGivesPlusTwoToughnessAndTrample(t *testing.T) {
	g, _ := newPlaneGame(t, planeGreatForest)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	rollBlank(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample on my creature")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("trample on an opposing creature")
	}
	var tough int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c, _ := g.LookupCardForEffect(bear)
		tough = c.CurrentToughness()
	})
	if tough != 4 {
		t.Errorf("toughness %d, want 4", tough)
	}
}

var planeThePit = planeFixture{"b60e8617-fc35-4c8a-beb0-4688b88484c4", "The Pit", "Plane — The Abyss"}

func TestThePitEachPlayerChoosesATokenAndTheDemonBites(t *testing.T) {
	requirePlaneRegisteredFull(t, planeThePit)
	g, _ := newPlaneGame(t, planeThePit)
	advanceToMain(t, g)
	planeswalkToPlane(t, g, planeThePit.Oracle)
	passPriorityAroundTable(t, g)
	// Four prompts, one per player in turn order; the active player takes
	// the Demon, the rest Angels.
	for i := 0; i < 4; i++ {
		ch := latestChoiceOfKind(g, game.PendingChoiceOptionPick)
		if ch == nil {
			t.Fatalf("prompt %d missing", i)
		}
		idx := 0
		if i == 0 {
			idx = 1
		}
		if err := g.ResolveOptionPick(ch.ID, ch.Chooser, idx); err != nil {
			t.Fatalf("answering prompt %d: %v", i, err)
		}
	}
	passPriorityAroundTable(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var demons, angels int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.Name {
			case "Demon":
				demons++
			case "Angel":
				angels++
			}
		}
	})
	if demons != 1 || angels != 3 {
		t.Fatalf("demons %d angels %d, want 1 and 3", demons, angels)
	}
	// Next turn cycle: the Demon's controller reaches an upkeep with no other
	// creature, so it deals 6 damage to them.
	life := me.Life
	for i := 0; i < 4; i++ {
		toNextUpkeep(t, g)
		passPriorityAroundTable(t, g)
	}
	if g.Seats[g.Turn.ActiveSeat] != me {
		t.Fatalf("not back at the Demon's controller")
	}
	if got := life - me.Life; got != 6 {
		t.Errorf("the Demon's controller lost %d life, want 6", got)
	}
}

func TestThePitChaosEachPlayerSacrificesANonartifactCreature(t *testing.T) {
	g, _ := newPlaneGame(t, planeThePit)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Ogre", 3, 3)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, id := range []uuid.UUID{mine, theirs} {
			if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneBattlefield {
				t.Errorf("%v was not sacrificed", id)
			}
		}
	})
}

var planeWesternCloud = planeFixture{"4abae901-eb47-4318-bf64-322a893c2833", "The Western Cloud", "Plane — Gobakhan"}

func TestWesternCloudPreventsDamageToThePlanarControllersCreatures(t *testing.T) {
	requirePlaneRegisteredFull(t, planeWesternCloud)
	g, _ := newPlaneGame(t, planeWesternCloud)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 4, 4)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	treasures := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treasure" && c.Controller == me.ID && c.Tapped {
				treasures++
			}
		}
		if z := g.FindCardZoneForEffect(mine); z == nil || z.Kind != game.ZoneBattlefield {
			t.Error("the planar controller's creature died despite the prevention")
		}
		if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneBattlefield {
			t.Error("a 4/4 died to three 1-damage pings")
		}
	})
	if treasures != 3 {
		t.Errorf("%d tapped Treasures, want 3", treasures)
	}
	if got := damageOn(t, g, mine); got != 0 {
		t.Errorf("%d damage marked on my creature, want 0", got)
	}
	if got := damageOn(t, g, theirs); got != 3 {
		t.Errorf("%d damage marked on the opposing creature, want 3", got)
	}
}

var planeTheWilds = planeFixture{"cb7ae16a-9982-4782-9e20-07bb771d0601", "The Wilds", "Plane — Eldraine"}

func countFood(g *game.Game, owner uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Food" && c.Controller == owner {
				n++
			}
		}
	})
	return n
}

func TestTheWildsMakesFoodOnEntryAndAtUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTheWilds)
	g, _ := newPlaneGame(t, planeTheWilds)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	planeswalkToPlane(t, g, planeTheWilds.Oracle)
	passPriorityAroundTable(t, g)
	if got := countFood(g, me.ID); got != 1 {
		t.Fatalf("%d Food after planeswalking in, want 1", got)
	}
	next := toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if got := countFood(g, next.ID); got != 1 {
		t.Errorf("%d Food at the next planar controller's upkeep, want 1", got)
	}
}

func TestTheWildsChaosEdictPaysOneOrTwoFood(t *testing.T) {
	for _, tc := range []struct {
		name      string
		toughness int
		want      int
	}{{"small", 3, 1}, {"big", 4, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g, _ := newPlaneGame(t, planeTheWilds)
			advanceToMain(t, g)
			me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
			victim := pushVanillaCreature(g, opp.ID, "Victim", 1, tc.toughness)
			rollChaos(t, g, me.ID)
			passPriorityAroundTable(t, g)
			answerPickTargetPlayer(t, g, opp.ID)
			passPriorityAroundTable(t, g)
			b6AnswerSacrifices(t, g)
			g.ReadSnapshot(func() {
				if z := g.FindCardZoneForEffect(victim); z != nil && z.Kind == game.ZoneBattlefield {
					t.Error("the creature was not sacrificed")
				}
			})
			if got := countFood(g, me.ID); got != tc.want {
				t.Errorf("%d Food, want %d", got, tc.want)
			}
		})
	}
}

// b6AnswerSacrifices answers every open sacrifice prompt with its first
// candidate, then lets the table settle.
func b6AnswerSacrifices(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if isSacrificePrompt(c) {
				open = c
				break
			}
		}
		if open == nil {
			break
		}
		opts := sacrificeOptions(open)
		if len(opts) == 0 {
			t.Fatalf("a sacrifice prompt with no candidates")
		}
		var err error
		if open.Kind == game.PendingChoiceSacrifice {
			err = g.ResolveSacrificeChoice(open.ID, open.Chooser, opts[0])
		} else {
			err = g.ResolveChooseCards(open.ID, open.Chooser, opts[:1])
		}
		if err != nil {
			t.Fatalf("answering the sacrifice prompt: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
}

var planeTowashi = planeFixture{"0c1cb8e2-8ed4-4a0e-926b-4df5466c607c", "Towashi", "Plane — Kamigawa"}

func TestTowashiModifiedCreaturesTrampleAndDraw(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTowashi)
	g, _ := newPlaneGame(t, planeTowashi)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	modded := pushVanillaCreature(g, me.ID, "Modded", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, modded, game.CounterPlusOne, 1) })
	g.BumpLayerVersionForTest()
	if hasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("an unmodified creature has trample")
	}
	if !hasEffectiveKeyword(t, g, modded, "trample") {
		t.Fatal("a modified creature has no trample")
	}
	before := me.Hand.Size()
	life := opp.Life
	pr7bAttack(t, g, opp.ID, plain, modded)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1 (only the modified attacker)", got)
	}
	if opp.Life >= life {
		t.Error("no combat damage dealt")
	}
}

func TestTowashiChaosDistributesThreeCounters(t *testing.T) {
	g, _ := newPlaneGame(t, planeTowashi)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTargetsDivided(ch.ID, ch.Chooser, []game.TargetRef{cardRef(a), cardRef(b)}, map[uuid.UUID]int{a: 1, b: 2}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if x, y := countersOn(g, a, game.CounterPlusOne), countersOn(g, b, game.CounterPlusOne); x != 1 || y != 2 {
		t.Errorf("counters %d and %d, want 1 and 2", x, y)
	}
}

var (
	planeTrugaJungle = planeFixture{"d858e6db-9842-41d1-8006-574c8dcd98ed", "Truga Jungle", "Plane — Ergamon"}
	planeTurriIsland = planeFixture{"1300b5e7-d741-4723-bb15-dabd04cb0c11", "Turri Island", "Plane — Ir"}
)

func TestTrugaJungleLandsTapForAnyColorAndChaosTakesLands(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTrugaJungle)
	g, _ := newPlaneGame(t, planeTrugaJungle)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := b12Permanent(g, me.ID, "Odd Land", "Land")
	theirs := b12Permanent(g, opp.ID, "Their Land", "Land")
	g.BumpLayerVersionForTest()
	for _, id := range []uuid.UUID{mine, theirs} {
		var n int
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			c, _ := g.LookupCardForEffect(id)
			n = len(game.ManaAbilitiesForCard(c))
		})
		if n != 1 {
			t.Errorf("a land has %d mana abilities, want the granted one", n)
		}
	}
	// Chaos: top three are land, creature, land.
	libraryTop(me,
		cardNamed("Bottom Land", "Basic Land — Forest"),
		cardNamed("Bear", "Creature — Bear"),
		cardNamed("Top Land", "Basic Land — Plains"))
	hand := me.Hand.Size()
	lib := me.Library.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	// The rest (one creature) goes to the bottom: a single card needs no order.
	for i := 0; i < 4; i++ {
		if ch := latestChoiceOfKind(g, game.PendingChoicePutInLibrary); ch != nil {
			t.Fatalf("an order was asked for a single card")
		}
		passPriorityAroundTable(t, g)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("%d cards entered the hand, want the 2 lands", got)
	}
	if got := lib - me.Library.Size(); got != 2 {
		t.Errorf("library fell by %d, want 2", got)
	}
}

func TestTurriIslandCreatureSpellsCostTwoLessAndChaosTakesCreatures(t *testing.T) {
	requirePlaneRegisteredFull(t, planeTurriIsland)
	g, _ := newPlaneGame(t, planeTurriIsland)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	creature := func(cost string) int {
		return b6Price(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: cost, Power: 2, Toughness: 2})
	}
	if got := creature("{3}{G}"); got != 2 {
		t.Errorf("a {3}{G} creature spell costs %d, want 2", got)
	}
	if got := creature("{1}{G}"); got != 1 {
		t.Errorf("a {1}{G} creature spell costs %d, want 1 (only generic mana is reduced)", got)
	}
	if got := b6Price(t, g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{3}"}); got != 3 {
		t.Errorf("a noncreature spell costs %d, want 3", got)
	}

	libraryTop(me,
		cardNamed("Spell", "Sorcery"),
		cardNamed("Bear", "Creature — Bear"),
		cardNamed("Elf", "Creature — Elf"))
	hand, yard := me.Hand.Size(), me.Graveyard.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("%d cards entered the hand, want the 2 creatures", got)
	}
	if got := me.Graveyard.Size() - yard; got != 1 {
		t.Errorf("%d cards entered the graveyard, want the sorcery", got)
	}
}

var planeUndercityReaches = planeFixture{"f3f5a033-d063-4250-9cfe-4c3c1a2725f1", "Undercity Reaches", "Plane — Ravnica"}

func TestUndercityReachesLetsTheCreaturesControllerDrawAndChaosLiftsHandSize(t *testing.T) {
	requirePlaneRegisteredFull(t, planeUndercityReaches)
	g, _ := newPlaneGame(t, planeUndercityReaches)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := me.Hand.Size()
	pr7bAttack(t, g, opp.ID, bear)
	var ch *game.PendingChoice
	for i := 0; i < 8 && ch == nil; i++ {
		if ch = latestChoiceOfKind(g, game.PendingChoiceConfirm); ch == nil {
			if _, err := g.AdvanceStep(); err != nil && latestChoiceOfKind(g, game.PendingChoiceConfirm) == nil {
				t.Fatalf("AdvanceStep: %v", err)
			}
			ch = latestChoiceOfKind(g, game.PendingChoiceConfirm)
		}
	}
	if ch == nil {
		t.Fatal("the creature's controller was not asked")
	}
	if ch.Chooser != me.ID {
		t.Errorf("asked %v, want the creature's controller", ch.Chooser)
	}
	if err := g.ResolveConfirm(ch.ID, ch.Chooser, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.MaxHandSize; got != game.NoMaxHandSize {
		t.Errorf("max hand size %d, want none", got)
	}
}

var planeUnyaro = planeFixture{"503fcc43-dafd-484b-a97f-bbd11eed66de", "Unyaro", "Plane — Zhalfir"}

func b6PhasedOut(g *game.Game, id uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() { out = g.IsPhasedOutForEffect(id) })
	return out
}

func TestUnyaroPhasesCreaturesOutUntilAPlayerPlaneswalks(t *testing.T) {
	requirePlaneRegisteredFull(t, planeUnyaro)
	g, _ := newPlaneGame(t, planeUnyaro)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	// Not planeswalked to this turn (it was the starting plane): nothing.
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if b6PhasedOut(g, mine) || b6PhasedOut(g, theirs) {
		t.Fatal("creatures phased out though the plane was not planeswalked to this turn")
	}
	endTurn(t, g)

	// A later turn, planeswalking to Unyaro.
	g2, _ := newPlaneGame(t, planeUnyaro)
	advanceToMain(t, g2)
	me2, opp2 := g2.Seats[g2.Turn.ActiveSeat], opponentOf(g2)
	mine2 := pushVanillaCreature(g2, me2.ID, "Mine", 2, 2)
	theirs2 := pushVanillaCreature(g2, opp2.ID, "Theirs", 2, 2)
	planeswalkToPlane(t, g2, planeUnyaro.Oracle)
	passPriorityAroundTable(t, g2)
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if !b6PhasedOut(g2, mine2) || !b6PhasedOut(g2, theirs2) {
		t.Fatalf("every creature should be phased out (mine %v theirs %v)", b6PhasedOut(g2, mine2), b6PhasedOut(g2, theirs2))
	}
	// A phased-out creature skips the untap-step phase-in.
	next := toNextUpkeep(t, g2)
	_ = next
	if !b6PhasedOut(g2, mine2) || !b6PhasedOut(g2, theirs2) {
		t.Fatal("a creature phased in at an untap step")
	}
	g2.WithWriteLock(func() { _ = g2.PlaneswalkForEffect(g2.Seats[g2.Turn.ActiveSeat].ID) })
	if b6PhasedOut(g2, mine2) || b6PhasedOut(g2, theirs2) {
		t.Error("creatures stayed phased out after a planeswalk")
	}
}

func TestUnyaroChaosMakesTwoKnights(t *testing.T) {
	g, _ := newPlaneGame(t, planeUnyaro)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Knight" && c.Controller == me.ID {
				n++
			}
		}
	})
	if n != 2 {
		t.Errorf("%d Knights, want 2", n)
	}
}

var planeValorsReach = planeFixture{"61b5984a-de6a-4d5a-8b20-e2fb3b553a34", "Valor's Reach", "Plane — Kylem"}

func TestValorsReachDoubleStrikesExactlyTwoAttackersOnce(t *testing.T) {
	requirePlaneRegisteredFull(t, planeValorsReach)
	for _, n := range []int{1, 2, 3} {
		g, plane := newPlaneGame(t, planeValorsReach)
		advanceToMain(t, g)
		me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
		var ids []uuid.UUID
		for i := 0; i < n; i++ {
			ids = append(ids, pushVanillaCreature(g, me.ID, "Bear", 2, 2))
		}
		pr7bAttack(t, g, opp.ID, ids...)
		passPriorityAroundTable(t, g)
		for _, id := range ids {
			if got, want := hasEffectiveKeyword(t, g, id, "double strike"), n == 2; got != want {
				t.Errorf("%d attackers: double strike %v, want %v", n, got, want)
			}
		}
		_ = plane
	}
}

func TestValorsReachChaosUntapsAndAddsCombatInAMainPhase(t *testing.T) {
	g, _ := newPlaneGame(t, planeValorsReach)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, a)
	passPriorityAroundTable(t, g)
	if tapped := tmtB1Card(t, g, a).Tapped; tapped {
		t.Error("the creature was not untapped")
	}
	// Combat, then the added combat and the added main phase.
	combats := 0
	for i := 0; i < 40 && g.Turn.Step != game.StepEnd; i++ {
		if g.Turn.Step == game.StepBeginCombat {
			combats++
		}
		finishBlockDeclarations(t, g)
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	if combats != 2 {
		t.Errorf("%d combat phases this turn, want 2", combats)
	}
}

var planeAcademyTolariaWest = planeFixture{"15b979de-c8ee-4664-9ca7-6c4eb3346967", "Academy at Tolaria West", "Plane — Dominaria"}

func TestAcademyAtTolariaWestDrawsSevenOnlyWithAnEmptyHand(t *testing.T) {
	requirePlaneRegisteredFull(t, planeAcademyTolariaWest)
	g, _ := newPlaneGame(t, planeAcademyTolariaWest)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("hand changed from %d to %d with cards in hand", hand, me.Hand.Size())
	}

	g2, _ := newPlaneGame(t, planeAcademyTolariaWest)
	advanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	rollChaos(t, g2, me2.ID)
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != 0 {
		t.Fatalf("the chaos ability left %d cards in hand", me2.Hand.Size())
	}
	lib := me2.Library.Size()
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	if me2.Hand.Size() != 7 || lib-me2.Library.Size() != 7 {
		t.Errorf("hand %d after the end step, want 7", me2.Hand.Size())
	}
}

var planeAgyrem = planeFixture{"38f84e55-049c-441e-b4e2-1e207ab5dbe5", "Agyrem", "Plane — Ravnica"}

func b6Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.Nil, id, 99) })
	g.RunStateChecksForTest()
}

func TestAgyremReturnsWhiteToTheBattlefieldAndNonwhiteToHand(t *testing.T) {
	requirePlaneRegisteredFull(t, planeAgyrem)
	g, _ := newPlaneGame(t, planeAgyrem)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	white := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	b6Kill(g, white)
	b6Kill(g, green)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	// Two delayed triggers fire at once: their controller orders them.
	if c := latestChoiceOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
		var ids []uuid.UUID
		for _, it := range g.PendingTriggers {
			if it != nil {
				ids = append(ids, it.ID)
			}
		}
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, ids); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
	}
	for i := 0; i < 4; i++ {
		passPriorityAroundTable(t, g)
	}
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(white); z == nil || z.Kind != game.ZoneBattlefield {
			t.Error("the white creature did not return to the battlefield")
		}
		if z := g.FindCardZoneForEffect(green); z == nil || z.Kind != game.ZoneHand {
			t.Errorf("the nonwhite creature is not in its owner's hand")
		}
	})
}

func TestAgyremChaosStopsAttacksUntilAPlaneswalk(t *testing.T) {
	g, _ := newPlaneGame(t, planeAgyrem)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	next := toNextUpkeep(t, g)
	bear := pushVanillaCreature(g, next.ID, "Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("a creature attacked the planar controller who rolled chaos")
	}
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	if err := g.DeclareAttacker(bear, other.ID); err != nil {
		t.Errorf("attacking someone else was refused: %v", err)
	}
	// A planeswalk ends it: the next turn's attacker may attack the roller.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(next.ID) })
	for _, seat := range g.Seats {
		for _, pl := range seat.Statics {
			if pl.CantAttack.Protected != uuid.Nil {
				t.Error("the can't-attack restriction outlived the planeswalk")
			}
		}
	}
}

var planeAkoum = planeFixture{"25650a32-4014-4065-ad01-7357c3ad3995", "Akoum", "Plane — Zendikar"}

func TestAkoumLetsEveryoneCastEnchantmentsAtInstantSpeed(t *testing.T) {
	requirePlaneRegisteredFull(t, planeAkoum)
	g, _ := newPlaneGame(t, planeAkoum)
	advanceToMain(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	me := g.Seats[g.Turn.ActiveSeat]
	ench := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: ench, Name: "Shrine", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err == nil {
		t.Fatal("a creature was cast at the beginning of combat")
	}
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{}); err != nil {
		t.Errorf("an enchantment could not be cast with flash: %v", err)
	}
}

func TestAkoumChaosDestroysAnUnenchantedCreature(t *testing.T) {
	g, _ := newPlaneGame(t, planeAkoum)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	plain := pushVanillaCreature(g, opp.ID, "Plain", 2, 2)
	aura := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Aura", TypeLine: "Enchantment — Aura", Owner: opp.ID, Controller: opp.ID})
	enchanted := pushVanillaCreature(g, opp.ID, "Enchanted", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aura {
				g.Battlefield.Cards[i].AttachedTo = cardRef(enchanted)
			}
		}
	})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(enchanted)); err == nil {
		t.Fatal("an enchanted creature was accepted as the target")
	}
	answerPickTarget(t, g, plain)
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(plain); z != nil && z.Kind == game.ZoneBattlefield {
			t.Error("the unenchanted creature survived")
		}
	})
}

var planeAretopolis = planeFixture{"d6ab4159-e04e-4991-8a9b-9da302c98e9d", "Aretopolis", "Plane — Kephalai"}

func TestAretopolisScrollsGainLifeDrawCardsAndPlaneswalkAtTen(t *testing.T) {
	requirePlaneRegisteredFull(t, planeAretopolis)
	g, _ := newPlaneGame(t, planeAretopolis)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	planeswalkToPlane(t, g, planeAretopolis.Oracle)
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 1 {
		t.Fatalf("gained %d life on the first scroll counter, want 1", got)
	}
	hand := me.Hand.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d with two scroll counters, want 2", got)
	}
	// Eight more counters bring it to ten: it planeswalks away.
	for i := 0; i < 8; i++ {
		rollChaos(t, g, me.ID)
		passPriorityAroundTable(t, g)
	}
	if _, ok := faceUpPlane(g, planeAretopolis.Oracle); ok {
		t.Error("Aretopolis stayed face up with ten scroll counters")
	}
}

var planeAstralArena = planeFixture{"805f9dc2-e172-481d-bae7-9a136a3d1e49", "Astral Arena", "Plane — Kolbahan"}

func TestAstralArenaAllowsOneAttackerAndOneBlocker(t *testing.T) {
	requirePlaneRegisteredFull(t, planeAstralArena)
	g, _ := newPlaneGame(t, planeAstralArena)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 2, 2)
	b1 := pushVanillaCreature(g, opp.ID, "Block 1", 2, 2)
	b2 := pushVanillaCreature(g, opp.ID, "Block 2", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("the first attacker was refused: %v", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err == nil {
		t.Fatal("a second attacker was accepted")
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(b1, a); err != nil {
		t.Fatalf("the first blocker was refused: %v", err)
	}
	if err := g.DeclareBlocker(b2, a); err == nil {
		t.Fatal("a second blocker was accepted")
	}
}

func TestAstralArenaChaosDamagesEveryCreature(t *testing.T) {
	g, _ := newPlaneGame(t, planeAstralArena)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	small := pushVanillaCreature(g, me.ID, "Small", 2, 2)
	big := pushVanillaCreature(g, opp.ID, "Big", 3, 3)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(small); z != nil && z.Kind == game.ZoneBattlefield {
			t.Error("a 2/2 survived 2 damage")
		}
	})
	if got := damageOn(t, g, big); got != 2 {
		t.Errorf("%d damage on the 3/3, want 2", got)
	}
}

var planeBant = planeFixture{"82e67c57-6d7a-44cc-ba63-77ce887ab075", "Bant", "Plane — Alara"}

func TestBantGivesEveryCreatureExaltedAndChaosGrantsIndestructible(t *testing.T) {
	requirePlaneRegisteredFull(t, planeBant)
	g, plane := newPlaneGame(t, planeBant)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	red := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	g.BumpLayerVersionForTest()
	if !hasEffectiveKeyword(t, g, red, "exalted") {
		t.Fatal("a creature does not have exalted")
	}
	// Two creatures with exalted, one attacks alone: +2/+2.
	pr7bAttack(t, g, opp.ID, green)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, green); p != 4 {
		t.Errorf("power %d after attacking alone with two exalted instances, want 4", p)
	}
	_ = plane
}

func TestBantChaosPutsADivinityCounterOnlyOnGreenWhiteOrBlue(t *testing.T) {
	g, _ := newPlaneGame(t, planeBant)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	red := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(red)); err == nil {
		t.Fatal("a red creature was accepted")
	}
	answerPickTarget(t, g, green)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, green, "indestructible") {
		t.Fatal("the creature is not indestructible")
	}
	// Walk away from the plane: it keeps indestructible while it has the counter.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if !hasEffectiveKeyword(t, g, green, "indestructible") {
		t.Error("indestructible ended with the plane")
	}
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, green, "divinity", -1) })
	g.BumpLayerVersionForTest()
	if hasEffectiveKeyword(t, g, green, "indestructible") {
		t.Error("indestructible outlived the divinity counter")
	}
}

var planeCelestineReef = planeFixture{"2979597d-c3cc-4585-a9b7-7fa15641b030", "Celestine Reef", "Plane — Luvion"}

func TestCelestineReefOnlyFliersAndIslandwalkersAttackAndChaosProtectsUntilAPlaneswalk(t *testing.T) {
	requirePlaneRegisteredFull(t, planeCelestineReef)
	g, _ := newPlaneGame(t, planeCelestineReef)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	flier := pushVanillaCreature(g, me.ID, "Flier", 2, 2)
	walker := pushVanillaCreature(g, me.ID, "Walker", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case flier:
				c.Keywords = append(c.Keywords, "flying")
			case walker:
				c.Keywords = append(c.Keywords, "islandwalk")
			}
		}
	})
	g.BumpLayerVersionForTest()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(plain, opp.ID); err == nil {
		t.Error("a creature without flying or islandwalk attacked")
	}
	if err := g.DeclareAttacker(flier, opp.ID); err != nil {
		t.Errorf("a flier could not attack: %v", err)
	}
	if err := g.DeclareAttacker(walker, opp.ID); err != nil {
		t.Errorf("an islandwalker could not attack: %v", err)
	}
}

func TestCelestineReefChaosGatesLastUntilAPlaneswalk(t *testing.T) {
	g, _ := newPlaneGame(t, planeCelestineReef)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	gates := func() int {
		n := 0
		for _, s := range me.Statics {
			if s.GameEnd.CantLose {
				n++
			}
		}
		return n
	}
	if gates() == 0 {
		t.Fatal("no can't-lose gate was granted")
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if gates() != 0 {
		t.Error("the can't-lose gate outlived the planeswalk")
	}
}

var planeEdgeOfMalacol = planeFixture{"cb467901-a68d-4922-8ceb-867b582c8fe7", "Edge of Malacol", "Plane — Belenon"}

func TestEdgeOfMalacolPutsCountersInsteadOfUntapping(t *testing.T) {
	requirePlaneRegisteredFull(t, planeEdgeOfMalacol)
	g, _ := newPlaneGame(t, planeEdgeOfMalacol)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	setTapped := func(id uuid.UUID, v bool) {
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = v
				}
			}
		})
	}
	tapped := pushVanillaCreature(g, me.ID, "Tapped", 2, 2)
	fresh := pushVanillaCreature(g, me.ID, "Untapped", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	setTapped(tapped, true)
	setTapped(theirs, true)
	// Pass the turn around the table until it is my untap step again.
	for i := 0; i < len(g.Seats); i++ {
		endTurn(t, g)
		advanceToMainOf(t, g, g.Turn.ActiveSeat)
	}
	if g.Seats[g.Turn.ActiveSeat] != me {
		t.Fatalf("not back at my turn")
	}
	if got := countersOn(g, tapped, game.CounterPlusOne); got != 2 {
		t.Errorf("the tapped creature has %d counters, want 2", got)
	}
	if !tmtB1Card(t, g, tapped).Tapped {
		t.Error("the replaced creature untapped as well")
	}
	if got := countersOn(g, fresh, game.CounterPlusOne); got != 0 {
		t.Errorf("an already untapped creature got %d counters", got)
	}
}

func TestEdgeOfMalacolChaosUntapsYourCreatures(t *testing.T) {
	g, _ := newPlaneGame(t, planeEdgeOfMalacol)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == a {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if tmtB1Card(t, g, a).Tapped {
		t.Error("the creature stayed tapped")
	}
	if got := countersOn(g, a, game.CounterPlusOne); got != 0 {
		t.Errorf("the chaos untap was replaced: %d counters", got)
	}
}

var planeElorenWilds = planeFixture{"23fe1c21-c396-4ebf-a8b0-22d697952b34", "Eloren Wilds", "Plane — Shandalar"}

func TestElorenWildsDoublesAnyPermanentsManaAndChaosBansCasting(t *testing.T) {
	requirePlaneRegisteredFull(t, planeElorenWilds)
	g, _ := newPlaneGame(t, planeElorenWilds)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	forest := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	tapForMana(t, g, me.ID, forest)
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("a tapped Forest made %d mana, want 2", got)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	spell := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Ritual", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, spell, game.CastSpellParams{}); err == nil {
		t.Error("the banned player cast a spell")
	}
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	for _, st := range opp.Statics {
		if st.CastBan.Kind != game.CastBanNone {
			t.Error("the cast ban outlived the planeswalk")
		}
	}
}

var planeFeedingGrounds = planeFixture{"b1eff630-b56c-48a5-ac21-836a60157b35", "Feeding Grounds", "Plane — Muraganda"}

func TestFeedingGroundsReducesRedAndGreenSpellsAndChaosCountsManaValue(t *testing.T) {
	requirePlaneRegisteredFull(t, planeFeedingGrounds)
	g, _ := newPlaneGame(t, planeFeedingGrounds)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	cost := func(colors ...string) int {
		return b6Price(t, g, me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{2}", Colors: colors})
	}
	if got := cost("R"); got != 1 {
		t.Errorf("a red {2} spell costs %d, want 1", got)
	}
	if got := cost("G"); got != 1 {
		t.Errorf("a green {2} spell costs %d, want 1", got)
	}
	if got := cost("R", "G"); got != 0 {
		t.Errorf("a red-green {2} spell costs %d, want 0 (both reductions)", got)
	}
	if got := cost("U"); got != 2 {
		t.Errorf("a blue {2} spell costs %d, want 2", got)
	}
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("%d counters, want 3 (the creature's mana value)", got)
	}
}

// b6Price is what the board charges `seat` to cast `card` from hand, as a mana
// value (the real cast is free in these tables: nothing demands the payment).
func b6Price(t *testing.T, g *game.Game, seat *game.Player, card game.Card) int {
	t.Helper()
	base, err := game.ParseCost(card.ManaCost)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", card.ManaCost, err)
	}
	card.InstanceID, card.Owner, card.Controller = uuid.New(), seat.ID, seat.ID
	out, err := g.ApplyCostModifiers(base, game.CostQuery{Card: card, Controller: seat.ID, FromZone: game.ZoneHand})
	if err != nil {
		t.Fatalf("ApplyCostModifiers: %v", err)
	}
	return out.ManaValue()
}

var planeFieldsOfSummer = planeFixture{"81a0d96a-7ac8-4947-8779-2102c9b33d38", "Fields of Summer", "Plane — Moag"}

func TestFieldsOfSummerLetsTheCasterGainTwoAndChaosLetsYouGainTen(t *testing.T) {
	requirePlaneRegisteredFull(t, planeFieldsOfSummer)
	g, _ := newPlaneGame(t, planeFieldsOfSummer)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	castCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	ch := latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt)
	passPriorityAroundTable(t, g)
	if ch == nil {
		ch = latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt)
	}
	if ch == nil {
		for _, c := range g.PendingChoices {
			t.Logf("choice kind %s chooser %v", c.Kind, c.Chooser)
		}
		t.Logf("stack %d pending %d life %d", len(g.StackMeta), len(g.PendingTriggers), me.Life)
		t.Fatal("the caster was not asked")
	}
	if ch.Chooser != me.ID {
		t.Errorf("asked %v, want the caster", ch.Chooser)
	}
	if err := g.ResolveTriggerPrompt(ch.ID, ch.Chooser, true); err != nil {
		t.Fatalf("answering: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2", got)
	}
	life = me.Life
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch = latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt)
	if ch == nil {
		t.Fatal("the planar controller was not asked about the chaos life")
	}
	if err := g.ResolveTriggerPrompt(ch.ID, ch.Chooser, true); err != nil {
		t.Fatalf("answering: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 10 {
		t.Errorf("gained %d life, want 10", got)
	}
}

var planeFurnaceLayer = planeFixture{"1799d361-85e7-4c04-aa04-6f83e6499573", "Furnace Layer", "Plane — New Phyrexia"}

func TestFurnaceLayerDiscardsAtRandomAndChaosDestroys(t *testing.T) {
	requirePlaneRegisteredFull(t, planeFurnaceLayer)
	g, _ := newPlaneGame(t, planeFurnaceLayer)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// Make every hand a single land so whoever is picked pays 3 life.
	var lives []int
	for _, p := range g.Seats {
		for p.Hand.Size() > 0 {
			_, _ = p.Hand.PopTop()
		}
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: p.ID, Controller: p.ID})
		lives = append(lives, p.Life)
	}
	planeswalkToPlane(t, g, planeFurnaceLayer.Oracle)
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && isSacrificePrompt(c) {
			t.Fatal("unexpected sacrifice prompt")
		}
	}
	// A discard prompt is a choose_cards of the discard kind; answer it.
	for i := 0; i < 4; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards && len(c.ChooseCards) > 0 {
				_ = g.ResolveChooseCards(c.ID, c.Chooser, c.ChooseCards[:1])
			}
		}
		passPriorityAroundTable(t, g)
	}
	lost, emptied := 0, 0
	for i, p := range g.Seats {
		if lives[i]-p.Life == 3 {
			lost++
		}
		if p.Hand.Size() == 0 {
			emptied++
		}
	}
	if lost != 1 || emptied != 1 {
		t.Errorf("%d players lost 3 life and %d discarded; want exactly one player to do both", lost, emptied)
	}

	opp := opponentOf(g)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	land := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	ch := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, cardRef(land)); err == nil {
		t.Fatal("a land was accepted as the target")
	}
	answerPickTarget(t, g, rock)
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKind(g, game.PendingChoiceTriggerPrompt); c != nil {
		if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, true); err != nil {
			t.Fatalf("answering: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	g.ReadSnapshot(func() {
		if z := g.FindCardZoneForEffect(rock); z != nil && z.Kind == game.ZoneBattlefield {
			t.Error("the nonland permanent was not destroyed")
		}
	})
}

var planeGoldmeadow = planeFixture{"f12400e5-8288-48e1-b3b7-0cc4f8d9ca67", "Goldmeadow", "Plane — Lorwyn"}

func countGoats(g *game.Game, owner uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Goat" && c.Controller == owner {
				n++
			}
		}
	})
	return n
}

func TestGoldmeadowGivesTheLandsControllerThreeGoats(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGoldmeadow)
	g, _ := newPlaneGame(t, planeGoldmeadow)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	playLandFromHand(t, g, "Plains", "")
	passPriorityAroundTable(t, g)
	if got := countGoats(g, me.ID); got != 3 {
		t.Errorf("%d Goats for the land's controller, want 3", got)
	}
	if got := countGoats(g, opp.ID); got != 0 {
		t.Errorf("%d Goats for an opponent", got)
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := countGoats(g, me.ID); got != 4 {
		t.Errorf("%d Goats after chaos, want 4", got)
	}
}
