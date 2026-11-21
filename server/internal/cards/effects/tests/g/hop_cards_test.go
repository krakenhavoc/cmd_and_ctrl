package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsw "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/w"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Beast Hunt -------------------------------------------------------------

func TestHopBeastHuntTakesTheCreaturesAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Bear(), effectstest.C2b15Bear())
	effectstest.DftABCast(t, g, me, game.Card{Name: "Beast Hunt", TypeLine: "Sorcery", OracleID: hopBeastHuntOracle})
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Hideous End", "Instant", hopHideousEndOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the target was not destroyed")
	}
	if opp.Life != before-2 {
		t.Errorf("its controller's life %d -> %d, want a loss of 2", before, opp.Life)
	}
}

func TestHopHideousEndRefusesABlackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	black := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie", Colors: []string{"B"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Hideous End", "Instant", hopHideousEndOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: black}}); err == nil {
		t.Error("a black creature was accepted as the target")
	}
}

// --- Kor Sanctifiers --------------------------------------------------------

func TestHopKorSanctifiersDestroysOnlyWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	relic := effectstest.B12Permanent(g, opp.ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Kor Sanctifiers", hopKorTypeLine, hopKorSanctifiersOracle, nil, nil, nil); err != nil {
		t.Fatalf("unkicked CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Fatal("an unkicked Kor Sanctifiers triggered")
	}
	if !g.Battlefield.Contains(relic) {
		t.Fatal("an unkicked Kor Sanctifiers destroyed the artifact")
	}

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Kor Sanctifiers", hopKorTypeLine, hopKorSanctifiersOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("kicked CastSpell: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a creature was accepted as the target")
	}
	effectstest.PickCard(t, g, me.ID, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the kicked Kor Sanctifiers did not destroy the artifact")
	}
}

// --- Whiplash Trap ----------------------------------------------------------

func TestHopWhiplashTrapReturnsTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Whiplash Trap", hopTrapTypeLine, hopWhiplashTrapOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHand(me, mine) || !effectstest.InHand(opp, theirs) {
		t.Errorf("returned to their owners' hands: mine %v, theirs %v", effectstest.InHand(me, mine), effectstest.InHand(opp, theirs))
	}
}

func TestHopWhiplashTrapNeedsTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Whiplash Trap", hopTrapTypeLine, hopWhiplashTrapOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Error("cast with one target")
	}
}

func TestHopWhiplashTrapCostNeedsOneOpponentWithTwoCreaturesEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp2 := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	check := func() bool {
		var ok bool
		g.ReadSnapshot(func() { ok = cardsw.AnOpponentHadTwoOrMoreCreaturesEnterThisTurn(g, me.ID) })
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
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
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
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHand(opp, a) || !effectstest.InHand(opp, b) {
		t.Error("the trap-cast Whiplash Trap did not return both creatures")
	}
}
