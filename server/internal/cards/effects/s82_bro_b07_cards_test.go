package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b07_cards_test.go — The Brothers' War, S82 batch 7 (#5482). Each
// test plays the real catalog entry and asserts what its text says.

const (
	s82BroLatNamOracle      = "0d5a1870-6d26-47e7-a86c-f3ff15cb0707"
	s82BroMechanicOracle    = "a7e7be75-646d-4c99-8388-15bb422c8e25"
	s82BroAnointerOracle    = "22fa5324-ed3b-4b97-89b4-31c1f86d6ead"
	s82BroEvangelOracle     = "dc0bc051-2d5c-4b7f-9902-4f9d5ef8f19d"
	s82BroStalkerOracle     = "5a0c215a-2164-4cb8-89ff-2558b6b42dca"
	s82BroPatrolOracle      = "ef625caf-c5d3-47f4-b1fb-bd040c5d864c"
	s82BroPhalanxOracle     = "265f461e-d076-4119-8ea4-96360305f3a0"
	s82BroArchitectOracle   = "3194fe3d-57d3-4dab-ab39-3778b75150fa"
	s82BroDissidentOracle   = "6ad61622-a70f-423c-99e3-b1ef6ead0b31"
	s82BroSteelseekerOracle = "dacf7f0d-9ce2-4e28-98a5-e7ec270a4c83"
	s82BroWurmletOracle     = "a44ceede-9934-4d4e-8b8c-e087b3ef74d4"
	s82BroMatrixOracle      = "2f6d129c-e6ad-4f35-b8b1-56a101fdd5db"
)

func TestS82BroB07EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		s82BroLatNamOracle:      "Lat-Nam Adept",
		s82BroMechanicOracle:    "Thopter Mechanic",
		s82BroAnointerOracle:    "Gurgling Anointer",
		s82BroEvangelOracle:     "Evangel of Synthesis",
		s82BroStalkerOracle:     "Trench Stalker",
		s82BroPatrolOracle:      "Perimeter Patrol",
		s82BroPhalanxOracle:     "Phalanx Vanguard",
		s82BroArchitectOracle:   "Thopter Architect",
		s82BroDissidentOracle:   "Yotian Dissident",
		s82BroSteelseekerOracle: "Sarinth Steelseeker",
		s82BroWurmletOracle:     "Teething Wurmlet",
		s82BroMatrixOracle:      "Symmetry Matrix",
	}
	for oracle, name := range want {
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

func s82BroDraw(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	hobA5Draw(t, g, p, nil)
}

func s82BroRobot(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	tmtB3MakeToken(g, owner, "1/1 colorless Robot artifact")
	passPriorityAroundTable(t, g)
}

// --- the "second card each turn" creatures ---

func TestS82BroB07SecondDrawGrowsOnTheSecondCardOnly(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Lat-Nam Adept", "Creature — Human Wizard", s82BroLatNamOracle},
		{"Thopter Mechanic", "Creature — Human Artificer", s82BroMechanicOracle},
		{"Gurgling Anointer", "Creature — Phyrexian Horror", s82BroAnointerOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[1]
			id := b27Push(g, me.ID, c.name, c.typeLine, c.oracle, "", 2, 2)
			for i, want := range []int{0, 1, 1} {
				s82BroDraw(t, g, me)
				if got := plusOnes(g, id); got != want {
					t.Fatalf("%d counters after draw %d, want %d", got, i+1, want)
				}
			}
		})
	}
}

func TestS82BroB07SecondDrawIgnoresAnOpponentsDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	id := b27Push(g, me.ID, "Lat-Nam Adept", "Creature — Human Wizard", s82BroLatNamOracle, "", 3, 3)
	s82BroDraw(t, g, opp)
	s82BroDraw(t, g, opp)
	if got := plusOnes(g, id); got != 0 {
		t.Errorf("an opponent's second draw gave %d counters", got)
	}
}

func TestS82BroB07ThopterMechanicLeavesAFlyingThopterWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b27Push(g, me.ID, "Thopter Mechanic", "Creature — Human Artificer", s82BroMechanicOracle, "{1}{U}", 2, 1, "U")
	b27Kill(g, id)
	passPriorityAroundTable(t, g)
	tid := battlefieldIDNamed(g, me.ID, "Thopter")
	tok, ok := g.LookupCardForEffect(tid)
	if !ok || !tok.IsArtifact() || !tok.IsCreature() || tok.Power != 1 || tok.Toughness != 1 || len(tok.Colors) != 0 {
		t.Fatalf("Thopter token = %+v, want a 1/1 colorless artifact creature", tok)
	}
	assertKeywords(t, g, tid, "flying")
}

func TestS82BroB07GurglingAnointerReturnsACreatureUpToItsLastKnownPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b27Push(g, me.ID, "Gurgling Anointer", "Creature — Phyrexian Horror", s82BroAnointerOracle, "{2}{B}", 1, 3, "B")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterPlusOne, 2) })
	small := socB6Graveyard(me, "Small", "Creature — Ogre", "{3}")
	big := socB6Graveyard(me, "Big", "Creature — Ogre", "{4}")
	land := socB6Graveyard(me, "Rock", "Artifact", "{1}")
	theirs := socB6Graveyard(opp, "Theirs", "Creature — Ogre", "{1}")
	b27Kill(g, id)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the dies trigger asked for no target")
	}
	for _, bad := range []uuid.UUID{big, land, theirs, id} {
		if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bad}}); err == nil {
			t.Errorf("illegal target %v was accepted (power 3 as it died)", bad)
		}
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: small}}); err != nil {
		t.Fatalf("pick the mana value 3 creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(small) {
		t.Error("the creature card did not return")
	}
}

// --- draw-count statics ---

func TestS82BroB07TrenchStalkerHasDeathtouchAndLifelinkAfterTwoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	id := b27Push(g, me.ID, "Trench Stalker", "Creature — Beast Horror", s82BroStalkerOracle, "{4}{B}", 4, 5, "B")
	s82BroDraw(t, g, me)
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("has %s after one draw", kw)
		}
	}
	s82BroDraw(t, g, me)
	assertKeywords(t, g, id, "deathtouch", "lifelink")
}

func TestS82BroB07EvangelLootsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := g.TurnTallyFor(me.ID).CardsDrawn
	castCatalogSpell(t, g, "Evangel of Synthesis", "Creature — Phyrexian Human Cleric", s82BroEvangelOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := g.TurnTallyFor(me.ID).CardsDrawn - before; got != 1 {
		t.Fatalf("drew %d cards from the loot, want 1", got)
	}
	discard := me.Hand.Cards[0].InstanceID
	answerDiscard(t, g, me.ID, discard)
	if !me.Graveyard.Contains(discard) {
		t.Fatal("the loot did not discard")
	}
}

func TestS82BroB07EvangelGetsMenaceAndPowerAfterTwoDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	id := b27Push(g, me.ID, "Evangel of Synthesis", "Creature — Phyrexian Human Cleric", s82BroEvangelOracle, "{U}{B}", 2, 3, "U", "B")
	s82BroDraw(t, g, me)
	if effectivePower(t, g, id) != 2 || hasEffectiveKeyword(t, g, id, "menace") {
		t.Fatalf("after one draw: power %d, menace %v", effectivePower(t, g, id), hasEffectiveKeyword(t, g, id, "menace"))
	}
	s82BroDraw(t, g, me)
	if effectivePower(t, g, id) != 3 {
		t.Errorf("power %d after two draws, want 3", effectivePower(t, g, id))
	}
	assertKeywords(t, g, id, "menace")
}

// --- artifact-enters creatures ---

func TestS82BroB07PumpOnArtifactsLastsTheTurnAndCountsOnlyYours(t *testing.T) {
	for _, c := range []struct{ name, oracle, typeLine string }{
		{"Phalanx Vanguard", s82BroPhalanxOracle, "Creature — Human Soldier"},
		{"Perimeter Patrol", s82BroPatrolOracle, "Creature — Human Soldier"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b27Push(g, me.ID, c.name, c.typeLine, c.oracle, "", 2, 2)
			s82BroRobot(t, g, opp.ID)
			if effectivePower(t, g, id) != 2 {
				t.Fatalf("an opponent's artifact pumped it to %d", effectivePower(t, g, id))
			}
			tmtB3MakeToken(g, me.ID, "2/2 red Mutant")
			passPriorityAroundTable(t, g)
			if effectivePower(t, g, id) != 2 {
				t.Fatalf("a nonartifact pumped it to %d", effectivePower(t, g, id))
			}
			s82BroRobot(t, g, me.ID)
			s82BroRobot(t, g, me.ID)
			if effectivePower(t, g, id) != 4 {
				t.Fatalf("power %d after two artifacts, want 4", effectivePower(t, g, id))
			}
			advanceToUpkeepOf(t, g, 1)
			if effectivePower(t, g, id) != 2 {
				t.Errorf("power %d next turn, want the pump gone", effectivePower(t, g, id))
			}
		})
	}
}

func TestS82BroB07PhalanxVanguardHasVigilance(t *testing.T) {
	g := newCatalogGame(t)
	id := b27Push(g, g.Seats[0].ID, "Phalanx Vanguard", "Creature — Human Soldier", s82BroPhalanxOracle, "{1}{W}", 2, 2, "W")
	assertKeywords(t, g, id, "vigilance")
}

func TestS82BroB07ThopterArchitectGivesAnyCreatureFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Thopter Architect", "Creature — Human Artificer", s82BroArchitectOracle, "{3}{W}", 2, 3, "W")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tmtB3MakeToken(g, me.ID, "1/1 colorless Robot artifact")
	answerTriggerTargets(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, theirs, "flying")
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, theirs, "flying") {
		t.Error("flying outlived the turn")
	}
}

func TestS82BroB07YotianDissidentPutsACounterOnTargetCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Yotian Dissident", "Creature — Human Artificer", s82BroDissidentOracle, "{G}{W}", 1, 1, "G", "W")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tmtB3MakeToken(g, me.ID, "1/1 colorless Robot artifact")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 1 || plusOnes(g, theirs) != 0 {
		t.Errorf("counters: bear %d, theirs %d; want 1, 0", plusOnes(g, bear), plusOnes(g, theirs))
	}
}

func TestS82BroB07TeethingWurmletGrowsOnceAPerTurnGainsLifeEachTimeAndGetsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b27Push(g, me.ID, "Teething Wurmlet", "Creature — Wurm", s82BroWurmletOracle, "{G}", 1, 1, "G")
	life := me.Life
	s82BroRobot(t, g, me.ID)
	if me.Life != life+1 || plusOnes(g, id) != 1 {
		t.Fatalf("first artifact: life %+d, counters %d; want +1, 1", me.Life-life, plusOnes(g, id))
	}
	if hasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("deathtouch with one artifact")
	}
	s82BroRobot(t, g, me.ID)
	if me.Life != life+2 || plusOnes(g, id) != 1 {
		t.Fatalf("second artifact: life %+d, counters %d; want +2, 1", me.Life-life, plusOnes(g, id))
	}
	if hasEffectiveKeyword(t, g, id, "deathtouch") {
		t.Error("deathtouch with two artifacts")
	}
	s82BroRobot(t, g, me.ID)
	assertKeywords(t, g, id, "deathtouch")
	if me.Life != life+3 || plusOnes(g, id) != 1 {
		t.Errorf("third artifact: life %+d, counters %d; want +3, 1", me.Life-life, plusOnes(g, id))
	}
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	s82BroRobot(t, g, me.ID)
	if plusOnes(g, id) != 2 {
		t.Errorf("counters %d on the next turn's first artifact, want 2", plusOnes(g, id))
	}
}

// --- Symmetry Matrix ---

func TestS82BroB07SymmetryMatrixPaysOneToDrawForAnEqualPowerAndToughnessCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Symmetry Matrix", "Artifact", s82BroMatrixOracle, "{4}", 0, 0)
	// 3/1 enters: nothing.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("3/1 red Elemental"), 1) })
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a 3/1 prompted a payment")
	}
	// 2/2 enters, pay {1}: draw.
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 red Soldier"), 1) })
	passPriorityAroundTable(t, g)
	floatMana(t, g, me, "{C}")
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d → %d after paying, want +1", hand, me.Hand.Size())
	}
	// Decline: no draw.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("2/2 red Soldier"), 1) })
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("a declined payment drew a card")
	}
}

// --- Sarinth Steelseeker ---

func s82BroSteelseekerScene(t *testing.T, top [3]string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Sarinth Steelseeker", "Creature — Human Artificer Scout", s82BroSteelseekerOracle, "{1}{G}", 1, 2, "G")
	ids := s82Library(me, top)
	s82BroRobot(t, g, me.ID)
	return g, me, ids[0]
}

func TestS82BroB07SteelseekerTakesALandIntoYourHand(t *testing.T) {
	g, me, land := s82BroSteelseekerScene(t, [3]string{"Forest", "Basic Land — Forest", ""})
	answerConfirmFor(t, g, me, true)
	if !me.Hand.Contains(land) {
		t.Fatal("the land is not in hand")
	}
}

func TestS82BroB07SteelseekerOffersTheGraveyardForANonlandOrADeclinedLand(t *testing.T) {
	g, me, bear := s82BroSteelseekerScene(t, [3]string{"Bear", "Creature — Bear", "{1}{G}"})
	answerConfirmFor(t, g, me, true)
	if !me.Graveyard.Contains(bear) {
		t.Fatal("the nonland card did not go to the graveyard")
	}

	g, me, land := s82BroSteelseekerScene(t, [3]string{"Forest", "Basic Land — Forest", ""})
	answerConfirmFor(t, g, me, false) // not into hand
	answerConfirmFor(t, g, me, true)  // into the graveyard
	if !me.Graveyard.Contains(land) || me.Hand.Contains(land) {
		t.Fatal("a declined land should be offered to, and go to, the graveyard")
	}

	g, me, top := s82BroSteelseekerScene(t, [3]string{"Bear", "Creature — Bear", "{1}{G}"})
	answerConfirmFor(t, g, me, false)
	if !me.Library.Contains(top) {
		t.Error("declining both should leave the card on top")
	}
}
