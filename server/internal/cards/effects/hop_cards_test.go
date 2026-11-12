package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planechase (HOP) card coverage, #5584.

const (
	hopBeastHuntOracle      = "069c8fa1-fa04-4620-9356-a618158dad3e"
	hopHideousEndOracle     = "f1eafb69-e7c0-4a22-9e73-6bc32168a7c5"
	hopKorSanctifiersOracle = "55d8ac5b-622c-42a9-83dc-f2c1c6c2506f"
	hopWhiplashTrapOracle   = "1bd26692-c030-4493-883e-2d26b3a1d025"
	hopKorTypeLine          = "Creature — Kor Cleric"
	hopTrapTypeLine         = "Instant — Trap"
)

func TestHopEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		hopBeastHuntOracle:      "Beast Hunt",
		hopHideousEndOracle:     "Hideous End",
		hopKorSanctifiersOracle: "Kor Sanctifiers",
		hopWhiplashTrapOracle:   "Whiplash Trap",
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

// --- Beast Hunt -------------------------------------------------------------

func TestHopBeastHuntTakesTheCreaturesAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	lib := dftMPLibraryTop(me, c2b15Bear(), c2b15Forest(), c2b15Bear(), c2b15Bear())
	dftABCast(t, g, me, game.Card{Name: "Beast Hunt", TypeLine: "Sorcery", OracleID: hopBeastHuntOracle})
	for _, i := range []int{0, 2} {
		if !me.Hand.Contains(lib[i]) {
			t.Errorf("creature %d is not in hand", i)
		}
	}
	if !me.Graveyard.Contains(lib[1]) {
		t.Error("the land revealed was not put into the graveyard")
	}
	if !me.Library.Contains(lib[3]) {
		t.Error("the fourth card left the library")
	}
}

// --- Hideous End ------------------------------------------------------------

func TestHopHideousEndDestroysAndItsControllerLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	castCatalogSpell(t, g, "Hideous End", "Instant", hopHideousEndOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the target was not destroyed")
	}
	if opp.Life != before-2 {
		t.Errorf("its controller's life %d -> %d, want a loss of 2", before, opp.Life)
	}
}

func TestHopHideousEndRefusesABlackCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	black := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if err := castCatalogSpellErr(t, g, "Hideous End", "Instant", hopHideousEndOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: black}}); err == nil {
		t.Error("a black creature was accepted as the target")
	}
}

// --- Kor Sanctifiers --------------------------------------------------------

func TestHopKorSanctifiersDestroysOnlyWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := b12Permanent(g, opp.ID, "Relic", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	if _, err := castWithOptionalCosts(t, g, "Kor Sanctifiers", hopKorTypeLine, hopKorSanctifiersOracle, nil, nil, nil); err != nil {
		t.Fatalf("unkicked CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Fatal("an unkicked Kor Sanctifiers triggered")
	}
	if !g.Battlefield.Contains(relic) {
		t.Fatal("an unkicked Kor Sanctifiers destroyed the artifact")
	}

	if _, err := castWithOptionalCosts(t, g, "Kor Sanctifiers", hopKorTypeLine, hopKorSanctifiersOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("kicked CastSpell: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the target")
	}
	pickCard(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the kicked Kor Sanctifiers did not destroy the artifact")
	}
}

// --- Whiplash Trap ----------------------------------------------------------

func TestHopWhiplashTrapReturnsTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Whiplash Trap", hopTrapTypeLine, hopWhiplashTrapOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if !inHand(me, mine) || !inHand(opp, theirs) {
		t.Errorf("returned to their owners' hands: mine %v, theirs %v", inHand(me, mine), inHand(opp, theirs))
	}
}

func TestHopWhiplashTrapNeedsTwoTargets(t *testing.T) {
	g := newCatalogGame(t)
	bear := b12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Whiplash Trap", hopTrapTypeLine, hopWhiplashTrapOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("cast with one target")
	}
}

func TestHopWhiplashTrapCostNeedsOneOpponentWithTwoCreaturesEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	advanceToMain(t, g)
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = anOpponentHadTwoOrMoreCreaturesEnterThisTurn(g, me.ID) })
		return ok
	}
	enter := func(who uuid.UUID, typeLine string, n int) {
		g.WithWriteLock(func() {
			if _, err := g.CreateTokensForEffect(who, game.Card{Name: "Token", TypeLine: typeLine, Power: 1, Toughness: 1}, n, game.TokenEntryOptions{}); err != nil {
				t.Errorf("CreateTokensForEffect: %v", err)
			}
		})
	}
	if check() {
		t.Fatal("offered before any creature entered")
	}
	enter(me.ID, "Token Creature — Bear", 3)
	enter(opp.ID, "Token Creature — Bear", 1)
	enter(opp2.ID, "Token Creature — Bear", 1)
	enter(opp.ID, "Token Artifact — Treasure", 2)
	if check() {
		t.Error("offered with no single opponent at two creatures")
	}

	// The alternative cost is refused while the condition is false.
	trap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: trap, Name: "Whiplash Trap", TypeLine: hopTrapTypeLine, ManaCost: "{3}{U}{U}",
		OracleID: hopWhiplashTrapOracle, Owner: me.ID, Controller: me.ID})
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err == nil {
		t.Fatal("the trap cost was accepted with no opponent at two creatures")
	}

	enter(opp2.ID, "Token Creature — Bear", 1)
	if !check() {
		t.Fatal("not offered once an opponent had two creatures enter")
	}
	if err := g.CastSpell(me.ID, trap, game.CastSpellParams{AlternativeCost: "trap", Targets: targets}); err != nil {
		t.Fatalf("the trap cost was refused once an opponent had two creatures enter: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inHand(opp, a) || !inHand(opp, b) {
		t.Error("the trap-cast Whiplash Trap did not return both creatures")
	}
}
