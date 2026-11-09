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
