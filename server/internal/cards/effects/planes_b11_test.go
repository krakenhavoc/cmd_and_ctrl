package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_b11_test.go — Planechase plane cards, batches 11 and 12 (the gold-
// bordered punk and pssc planes, ADR 0216 slice 4, #4870). Every test plays
// the real catalog entry face up on a Planechase table.

var (
	planeSquirrels = planeFixture{"7aac9628-f112-4840-91b4-7bb5cd896685", "We Hope You Like Squirrels", "Plane — Secret Lair"}
)

func TestWeHopeYouLikeSquirrelsGivesTheActingPlayerASquirrel(t *testing.T) {
	requirePlaneRegisteredFull(t, planeSquirrels)
	g, _ := newPlaneGame(t, planeSquirrels)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Squirrels for the caster after a spell, want 1", n)
	}
	next := toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	m := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" && c.Controller == next.ID {
			m++
		}
	}
	if m != 1 {
		t.Errorf("the upkeep player has %d Squirrels, want 1", m)
	}
}

func TestWeHopeYouLikeSquirrelsChaosGrowsSquirrels(t *testing.T) {
	g, _ := newPlaneGame(t, planeSquirrels)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sq := b12Creature(g, me.ID, "Squirrel", "Creature — Squirrel", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rollBlank(t, g, me.ID)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p := tmtB1Card(t, g, sq).CurrentPower(); p != 2 {
		t.Errorf("the Squirrel is %d power, want 2", p)
	}
	if p := tmtB1Card(t, g, bear).CurrentPower(); p != 2 {
		t.Errorf("the Bear is %d power, want 2", p)
	}
}

var planeFblthp = planeFixture{"ec2f6e72-8a2b-44f0-8241-7fe29c601d96", "Fblthp: Completely, Utterly, Totally Lost", "Plane — Secret Lair"}

func TestFblthpMakesAMapOnPlaneswalkAndUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeFblthp)
	g, _ := newPlaneGame(t, planeFblthp)
	advanceToMain(t, g)
	planeswalkToPlane(t, g, planeFblthp.Oracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Map"); n != 1 {
		t.Fatalf("%d Maps after planeswalking to the plane, want 1", n)
	}
	toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Map"); n != 2 {
		t.Errorf("%d Maps after the next upkeep, want 2", n)
	}
}

func TestFblthpChaosLootsTwo(t *testing.T) {
	g, plane := newPlaneGame(t, planeFblthp)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 1 {
		t.Fatalf("chaos put %d items on the stack, want 1", n)
	}
}

var (
	planeBlackLotusLounge = planeFixture{"ddea0dcf-e8d1-44ea-9ce4-009fa8bbfb95", "Black Lotus Lounge", "Plane — MagicCon"}
	planeLilGiri          = planeFixture{"5b35fbd0-65a0-4973-b2c9-841b11c0d426", "Li'l Giri Saves the Day", "Plane — Secret Lair"}
)

func TestBlackLotusLoungeMakesALotusOnPlaneswalkAndUpkeep(t *testing.T) {
	requirePlaneRegisteredFull(t, planeBlackLotusLounge)
	g, _ := newPlaneGame(t, planeBlackLotusLounge)
	advanceToMain(t, g)
	planeswalkToPlane(t, g, planeBlackLotusLounge.Oracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Black Lotus"); n != 1 {
		t.Fatalf("%d Black Lotus tokens after planeswalking here, want 1", n)
	}
	toNextUpkeep(t, g)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Black Lotus"); n != 2 {
		t.Errorf("%d Black Lotus tokens after the next upkeep, want 2", n)
	}
}

func TestBlackLotusLoungeChaosDrawsPerArtifactToken(t *testing.T) {
	g, _ := newPlaneGame(t, planeBlackLotusLounge)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	other := opponentOf(g)
	for range 2 {
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(other.ID, TreasureToken(), 1) })
	b12Permanent(g, me.ID, "Nontoken Rock", "Artifact")
	myHand, theirHand := me.Hand.Size(), other.Hand.Size()
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != myHand+2 {
		t.Errorf("the planar controller drew %d, want 2 (two artifact tokens)", got-myHand)
	}
	if got := other.Hand.Size(); got != theirHand+1 {
		t.Errorf("the other player drew %d, want 1", got-theirHand)
	}
}

func TestLilGiriMakesLilGiriAndProtectsLegendaryFood(t *testing.T) {
	requirePlaneRegisteredFull(t, planeLilGiri)
	g, _ := newPlaneGame(t, planeLilGiri)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	planeswalkToPlane(t, g, planeLilGiri.Oracle)
	passPriorityAroundTable(t, g)
	ids := tmtB1Named(g, "Li'l Giri")
	if len(ids) != 1 {
		t.Fatalf("%d Li'l Giri after planeswalking here, want 1", len(ids))
	}
	c := tmtB1Card(t, g, ids[0])
	if !c.IsLegendary() || !c.IsCreature() || !c.IsArtifact() || !c.HasSubtypeWord("Food") || !c.HasSubtypeWord("Warrior") || c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("Li'l Giri is %q %d/%d", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "hexproof") || !game.HasKeyword(&c, "indestructible") {
		t.Errorf("Li'l Giri lacks hexproof or indestructible: %v", c.Effective().Abilities)
	}
	plain := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if game.HasKeyword(planesB11Ptr(tmtB1Card(t, g, plain)), "hexproof") {
		t.Error("a nonlegendary nonFood creature has hexproof")
	}
	// Chaos: X = planar die rolls this turn.
	g.WithWriteLock(func() { g.Planechase.DieTurnSeq, g.Planechase.DieRolls = g.Turn.Seq, 3 })
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, ids[0])
	passPriorityAroundTable(t, g)
	if p := tmtB1Card(t, g, ids[0]).CurrentPower(); p != 4 {
		t.Errorf("Li'l Giri is %d power after 3 rolls' chaos, want 4", p)
	}
	_ = game.CounterPlusOne
}

func planesB11Ptr[T any](v T) *T { return &v }

var (
	planeKitchen       = planeFixture{"1d2bad25-dbe3-4b51-a78b-4a95dd5669fe", "Asmoranomardicadaistinaculdacar's Kitchen", "Plane — Hell"}
	planeJalirasShow   = planeFixture{"eea700a1-a5f5-4b9e-aa03-8e092deb7396", "Jalira's Show", "Plane — Las Vegas"}
	planesB11ElfOracle = "00000000-0000-4000-8000-0000000b1101"
)

func TestKitchenMakesFoodOnPlaneswalkAndWhenACreatureEnters(t *testing.T) {
	requirePlaneRegisteredFull(t, planeKitchen)
	g, _ := newPlaneGame(t, planeKitchen)
	advanceToMain(t, g)
	me, other := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	planeswalkToPlane(t, g, planeKitchen.Oracle)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Food"); n != 1 {
		t.Fatalf("%d Food after planeswalking here, want 1", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 green Squirrel"), 1) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Food"); n != 2 {
		t.Errorf("%d Food after a creature entered under the planar controller, want 2", n)
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(other.ID, TokenCard("1/1 green Squirrel"), 1) })
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Food"); n != 2 {
		t.Errorf("%d Food after a creature entered under another player, want 2", n)
	}
}

func TestKitchenGainingLifeDrainsAnOpponent(t *testing.T) {
	g, _ := newPlaneGame(t, planeKitchen)
	advanceToMain(t, g)
	me, other := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	life := other.Life
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, other.ID)
	passPriorityAroundTable(t, g)
	if got := life - other.Life; got != 3 {
		t.Errorf("the opponent lost %d, want 3", got)
	}
}

func TestKitchenChaosSacrificesForCardsAndLife(t *testing.T) {
	g, plane := newPlaneGame(t, planeKitchen)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Rock", "Artifact")
	life, hand := me.Life, me.Hand.Size()
	rollChaos(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 1 {
		t.Fatalf("chaos put %d items on the stack, want 1", n)
	}
	passPriorityAroundTable(t, g)
	answerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {a}})
	passPriorityAroundTable(t, g)
	if onBattlefieldForTest(g, a) {
		t.Error("the chosen creature was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if got := life - me.Life; got != 1 {
		t.Errorf("lost %d life, want 1", got)
	}
}

func onBattlefieldForTest(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestJalirasShowMakesAPerformerWhenYouAttack(t *testing.T) {
	requirePlaneRegisteredFull(t, planeJalirasShow)
	g, _ := newPlaneGame(t, planeJalirasShow)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)
	if n := countBattlefieldByName(g, "Performer"); n != 1 {
		t.Errorf("%d Performers after one attack with two creatures, want 1", n)
	}
}

func TestJalirasShowChaosDestroysAndRevealsForTheVictim(t *testing.T) {
	g, _ := newPlaneGame(t, planeJalirasShow)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	victim := b12Creature(g, opp.ID, "Doomed Bear", "Creature — Bear", 2, 2)
	putOnTopOfLibrary(g, opp, "Elf", "Creature — Elf", planesB11ElfOracle)
	putOnTopOfLibrary(g, opp, "Spell", "Sorcery", "")
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if onBattlefieldForTest(g, victim) {
		t.Error("the target was not destroyed")
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elf" && c.Controller == opp.ID {
			found = true
		}
	}
	if !found {
		t.Error("the victim's controller did not put the revealed creature card onto the battlefield")
	}
}

func TestJalirasShowPlaneswalkingHereAlsoTriggers(t *testing.T) {
	g, _ := newPlaneGame(t, planeJalirasShow)
	advanceToMain(t, g)
	opp := opponentOf(g)
	victim := b12Creature(g, opp.ID, "Doomed Bear", "Creature — Bear", 2, 2)
	planeswalkToPlane(t, g, planeJalirasShow.Oracle)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, victim)
	passPriorityAroundTable(t, g)
	if onBattlefieldForTest(g, victim) {
		t.Error("planeswalking here did not destroy the target")
	}
}

var (
	planeCityHall = planeFixture{"36aac8aa-d6f0-4e53-a280-73b01e040f35", "City Hall", "Plane — Chicago"}
	planeGhirapur = planeFixture{"6515e8d6-26ff-4716-a62d-a5847d0933e7", "Ghirapur Grand Prix", "Plane — Avishkar"}
)

func TestCityHallMakesThePlanarControllerTheMonarch(t *testing.T) {
	requirePlaneRegisteredFull(t, planeCityHall)
	g, _ := newPlaneGame(t, planeCityHall)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	planeswalkToPlane(t, g, planeCityHall.Oracle)
	passPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Errorf("the monarch is %v, want the planar controller", g.Monarch)
	}
}

func TestCityHallChaosTreasuresForThoseWhoAcceptAndTheyCantAttackYou(t *testing.T) {
	g, _ := newPlaneGame(t, planeCityHall)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	// Turn order from the active player: me says no, the next says yes, the rest no.
	for _, p := range seatsFromActiveForTest(g) {
		yes := p == opp.ID
		answerOptionFor(t, g, p, map[bool]int{true: 0, false: 1}[yes])
	}
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == opp.ID && c.Tapped {
			tapped++
		}
		if c.Name == "Treasure" && c.Controller != opp.ID {
			t.Errorf("a player who declined got a Treasure")
		}
	}
	if tapped != 2 {
		t.Errorf("%d tapped Treasures for the player who accepted, want 2", tapped)
	}
}

func seatsFromActiveForTest(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for i := range g.Seats {
		out = append(out, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)].ID)
	}
	return out
}

func answerOptionFor(t *testing.T, g *game.Game, who uuid.UUID, idx int) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick && c.Chooser == who {
			if err := g.ResolveOptionPick(c.ID, who, idx); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			return
		}
	}
	t.Fatalf("no option prompt open for %v", who)
}

func TestGhirapurGrandPrixSpeedCountersAndTreasures(t *testing.T) {
	requirePlaneRegisteredFull(t, planeGhirapur)
	g, _ := newPlaneGame(t, planeGhirapur)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// The harness walks away from the face-up plane first (its own
	// planeswalk-away ability), then back to it: two triggers to order.
	planeswalkToPlane(t, g, planeGhirapur.Oracle)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.Speed != 1 {
			t.Fatalf("%v has speed %d after the planeswalk, want 1 (all: %d %d %d %d; pending %d stack %d)", p.ID, p.Speed,
				g.Seats[0].Speed, g.Seats[1].Speed, g.Seats[2].Speed, g.Seats[3].Speed, len(g.PendingTriggers), len(g.StackMeta))
		}
	}
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.SetSpeedForTest(me.ID, 3)
	advanceTo(t, g, game.StepEnd)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if p := tmtB1Card(t, g, bear).CurrentPower(); p != 5 {
		t.Errorf("the bear is %d power, want 5 (three counters)", p)
	}
}

func TestGhirapurGrandPrixPlaneswalkingAwayPaysTheFastest(t *testing.T) {
	g, _ := newPlaneGame(t, planeGhirapur)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	g.SetSpeedForTest(me.ID, 3)
	g.SetSpeedForTest(opp.ID, 2)
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	passPriorityAroundTable(t, g)
	mine, theirs := 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			mine++
		}
		if c.Name == "Treasure" && c.Controller == opp.ID {
			theirs++
		}
	}
	if mine != 3 || theirs != 0 {
		t.Errorf("Treasures gained: planar controller %d, opponent %d; want 3 and 0", mine, theirs)
	}
}
