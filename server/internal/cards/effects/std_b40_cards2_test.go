package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b40_cards2_test.go — batch 40, the cards that needed an engine seam:
// Case of the Burning Masks (distinct damage sources), Sudden Setback (a
// spell put in a library without being countered), Lost in the Maze (an X
// target count on a trigger), Cryptic Coat and Expose the Culprit (cloak),
// Concealed Weapon (an Equipment with disguise) and Forensic Gadgeteer (an
// artifact-scoped activation reduction).

const stdB40DisguiseCreatureOracle = "26a7f92d-6914-466d-b7c7-9ac121dd5118"

// --- Case of the Burning Masks -------------------------------------------------

func stdB40Bolt(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(target))
	passPriorityAroundTable(t, g)
}

func stdB40Sources(g *game.Game, p uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() { n = g.DamageSourcesDealtThisTurnForEffect(p) })
	return n
}

func TestStdB40CaseOfTheBurningMasksDealsThreeToAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Case of the Burning Masks", "Enchantment — Case", stdB40MasksOracle, 0, 0)
	p := latestPickTarget(g, me.ID)
	if p == nil || !hasID(p.PickTargetCards, victim) || hasID(p.PickTargetCards, mine) {
		t.Fatalf("a creature an opponent controls: %+v", p)
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage did not kill the 3/3")
	}
}

func TestStdB40CaseOfTheBurningMasksCountsDistinctSourcesOfItsController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 9, 9)
	id := castHoldCreature(t, g, "Case of the Burning Masks", "Enchantment — Case", stdB40MasksOracle, 0, 0)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if got := stdB40Sources(g, me.ID); got != 1 {
		t.Fatalf("%d sources after the Case's own damage, want 1", got)
	}
	// The same source dealing damage again is still one source.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureKeyedForEffect(id, victim, 1, game.KeyedTail{}); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if got := stdB40Sources(g, me.ID); got != 1 {
		t.Errorf("the Case dealing damage twice made %d sources, want 1", got)
	}
	stdB40Bolt(t, g, victim)
	if got := stdB40Sources(g, me.ID); got != 2 {
		t.Errorf("the Case and a Bolt are %d sources, want 2", got)
	}
	if got := stdB40Sources(g, opp.ID); got != 0 {
		t.Errorf("the opponent controlled %d sources of damage", got)
	}
	if mkmAC2EndStepSolves(t, g, id) {
		t.Fatal("two sources solved the Case")
	}
}

func TestStdB40CaseOfTheBurningMasksSolvesAtTheEndStepAndExilesThreeToPlayOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 9, 9)
	var tops []uuid.UUID
	g.WithWriteLock(func() { me.Library.Cards = nil })
	for i := 0; i < 5; i++ {
		tops = append(tops, pushLibraryCardForTest(me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: lightningBoltOracle}))
	}
	id := castHoldCreature(t, g, "Case of the Burning Masks", "Enchantment — Case", stdB40MasksOracle, 0, 0)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	stdB40Bolt(t, g, victim)
	stdB40Bolt(t, g, victim)
	if got := stdB40Sources(g, me.ID); got != 3 {
		t.Fatalf("%d sources, want 3", got)
	}
	if !mkmAC2EndStepSolves(t, g, id) {
		t.Fatal("three sources did not solve the Case")
	}
	// Solved — Sacrifice this Case: exile the top three, choose one, play it.
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(id) {
		t.Error("the Case was not sacrificed")
	}
	c := pendingChooseCards(t, g, me.ID)
	if len(c.ChooseCards) != 3 {
		t.Fatalf("%d cards offered, want the 3 exiled", len(c.ChooseCards))
	}
	_ = tops
	chosen, other := c.ChooseCards[0], c.ChooseCards[1]
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{chosen}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if err := g.CastSpell(me.ID, other, game.CastSpellParams{Strict: true, FromZone: "exile", Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Error("a card that was not chosen could be cast from exile")
	}
	if err := g.CastSpell(me.ID, chosen, game.CastSpellParams{Strict: true, FromZone: "exile", Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Errorf("the chosen card could not be cast this turn: %v", err)
	}
}

// --- Sudden Setback ------------------------------------------------------------

func TestStdB40SuddenSetbackLetsThePermanentsOwnerChooseTopOrBottom(t *testing.T) {
	for _, bottom := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		rock := pushArtifactForTest(g, opp.ID, "Their Mind Stone")
		castCatalogSpell(t, g, "Sudden Setback", "Instant", stdB40SetbackOracle, cardRefs(rock))
		passPriorityAroundTable(t, g)
		if putInLibraryChoiceFor(g, me.ID) != nil {
			t.Fatal("the caster was asked; the owner chooses")
		}
		c := putInLibraryChoiceFor(g, opp.ID)
		if c == nil {
			t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
		}
		var top, bot []uuid.UUID
		if bottom {
			bot = []uuid.UUID{rock}
		} else {
			top = []uuid.UUID{rock}
		}
		if err := g.ResolvePutInLibrary(c.ID, opp.ID, bot, top); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
		if g.Battlefield.Contains(rock) {
			t.Fatal("the artifact is still on the battlefield")
		}
		cards := opp.Library.Cards
		if bottom && cards[0].InstanceID != rock {
			t.Errorf("the bottom card is %q, want the artifact", cards[0].Name)
		}
		if !bottom && cards[len(cards)-1].InstanceID != rock {
			t.Errorf("the top card is %q, want the artifact", cards[len(cards)-1].Name)
		}
	}
}

func TestStdB40SuddenSetbackPutsASpellInItsOwnersLibraryWithoutCounteringIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "Their Top")
	needle := castNeedleAt(t, g, opp, me.ID)
	castCounterAt(t, g, me, "Sudden Setback", stdB40SetbackOracle, needle)
	passPriorityAroundTable(t, g)
	if putInLibraryChoiceFor(g, me.ID) != nil {
		t.Fatal("the caster was asked; the spell's owner chooses")
	}
	c := putInLibraryChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatalf("the spell's owner was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{needle}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(needle) {
		t.Fatal("the spell is still on the stack")
	}
	cards := opp.Library.Cards
	if cards[len(cards)-1].InstanceID != needle {
		t.Errorf("the top card is %q, want the spell", cards[len(cards)-1].Name)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterSpell && ev.CardID == needle {
			t.Error("the spell was reported as countered")
		}
	}
}

func TestStdB40SuddenSetbackCantTargetALand(t *testing.T) {
	g := newCatalogGame(t)
	land := b12Permanent(g, g.Seats[1].ID, "Their Island", "Basic Land — Island")
	if err := castCatalogSpellErr(t, g, "Sudden Setback", "Instant", stdB40SetbackOracle, cardRefs(land)); err == nil {
		t.Error("a land was accepted as the target")
	}
}

// --- Lost in the Maze ----------------------------------------------------------

func stdB40CastMaze(t *testing.T, g *game.Game, me *game.Player, x int) uuid.UUID {
	t.Helper()
	id := avB1Hand(me, game.Card{Name: "Lost in the Maze", TypeLine: "Enchantment", OracleID: stdB40MazeOracle})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast with X = %d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestStdB40LostInTheMazeTapsXTargetsAndStunsTheOnesYouDontControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	spare := b12Creature(g, opp.ID, "Their Spare", "Creature — Ogre", 3, 3)
	stdB40CastMaze(t, g, me, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err == nil {
		t.Fatal("one target was accepted for X = 2")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs},
	}); err != nil {
		t.Fatalf("two targets for X = 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs} {
		if !b16Tapped(t, g, id) {
			t.Errorf("%s was not tapped", id)
		}
	}
	if b16Tapped(t, g, spare) {
		t.Error("an untargeted creature was tapped")
	}
	if got := battlefieldCardCopy(t, g, mine).Counters[game.CounterStun]; got != 0 {
		t.Errorf("my own creature got %d stun counters", got)
	}
	if got := battlefieldCardCopy(t, g, theirs).Counters[game.CounterStun]; got != 1 {
		t.Errorf("their creature has %d stun counters, want 1", got)
	}
	// Tapped creatures you control have hexproof.
	if !hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("my tapped creature has no hexproof")
	}
	if hasEffectiveKeyword(t, g, theirs, "hexproof") {
		t.Error("their tapped creature has hexproof")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(mine) })
	if hasEffectiveKeyword(t, g, mine, "hexproof") {
		t.Error("my untapped creature has hexproof")
	}
}

func TestStdB40LostInTheMazeWithXZeroTargetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 3)
	stdB40CastMaze(t, g, me, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("X = 0 asked for targets")
	}
	if b16Tapped(t, g, theirs) {
		t.Error("a creature was tapped with X = 0")
	}
}

// --- Forensic Gadgeteer --------------------------------------------------------

func TestStdB40ForensicGadgeteerInvestigatesOnAnArtifactSpellOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Forensic Gadgeteer", "Creature — Vedalken Artificer Detective", stdB40GadgeteerOracle, false)
	castCatalogSpell(t, g, "Shock", "Instant", b5ShockOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 0 {
		t.Fatalf("an instant made %d Clues", got)
	}
	castCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("an artifact spell made %d Clues, want 1", got)
	}
}

func TestStdB40ForensicGadgeteerTakesOneOffArtifactAbilitiesButLeavesOneMana(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Forensic Gadgeteer", "Creature — Vedalken Artificer Detective", stdB40GadgeteerOracle, false)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	for _, tc := range []struct{ cost, want string }{
		{"{2}", "{1}"}, {"{3}", "{2}"}, {"{1}", "{1}"}, {"{2}{G}", "{1}{G}"}, {"{R}{R}", "{R}{R}"}, {"{1}{G}", "{G}"},
	} {
		if got := lcPrice(t, g, me.ID, rock, tc.cost, 0); got != tc.want {
			t.Errorf("an artifact's %s costs %s, want %s", tc.cost, got, tc.want)
		}
	}
	if got := lcPrice(t, g, me.ID, bear, "{3}", 0); got != "{3}" {
		t.Errorf("a creature's {3} costs %s, want {3}", got)
	}
	if got := lcPrice(t, g, opp.ID, theirs, "{3}", 0); got != "{3}" {
		t.Errorf("an opponent's artifact's {3} costs %s, want {3}", got)
	}
}

// --- Concealed Weapon ----------------------------------------------------------

func TestStdB40ConcealedWeaponIsAnEquipmentWithDisguise(t *testing.T) {
	ac := game.AlternativeCostByKey(stdB40WeaponOracle, "disguise")
	if ac == nil || ac.FaceDown == nil || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != "{2}{R}" || ac.ManaCost != MorphCastCost {
		t.Fatalf("disguise declaration = %+v, want disguise {2}{R}", ac)
	}
}

func TestStdB40ConcealedWeaponAttachesWhenTurnedFaceUpForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	id := mkmMPCastFaceDown(t, g, "Concealed Weapon", "Artifact — Equipment", stdB40WeaponOracle, 0, 0)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if p, tough := mkmMPPT(t, g, id); p != 2 || tough != 2 {
		t.Fatalf("face down it is %d/%d, want a 2/2", p, tough)
	}
	mkmMPTurnFaceUp(t, g, me, id, "R", "C", "C")
	p := latestPickTarget(g, me.ID)
	if p == nil || !hasID(p.PickTargetCards, bear) {
		t.Fatalf("the face-up trigger offered %+v", p)
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := battlefieldCardCopy(t, g, id).AttachedTo.ID; got != bear {
		t.Fatalf("attached to %s, want the Bear", got)
	}
	if p, _ := mkmMPPT(t, g, bear); p != 5 {
		t.Errorf("the Bear's power is %d, want 5", p)
	}
}

// --- Expose the Culprit --------------------------------------------------------

func TestStdB40ExposeTheCulpritTurnsAFaceDownCreatureFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_ = opp
	id := mkmMPCastFaceDown(t, g, "Essence of Antiquity", "Artifact Creature — Golem", stdB40DisguiseCreatureOracle, 5, 5)
	if !findCardAnywhere(t, g, id).FaceDownIsPermanent() {
		t.Fatal("not face down")
	}
	castModal(t, g, "Expose the Culprit", "Instant", stdB40ExposeOracle, []int{0},
		[]game.TargetRef{modeRef(game.TargetCard, id, 0, 0)})
	passPriorityAroundTable(t, g)
	if findCardAnywhere(t, g, id).FaceDownIsPermanent() {
		t.Error("the creature is still face down")
	}
	_ = me
}

func TestStdB40ExposeTheCulpritCloaksYourFaceUpDisguiseCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Essence of Antiquity has disguise; a Bear does not.
	creeper := b12Push(g, me.ID, "Essence of Antiquity", "Artifact Creature — Golem", stdB40DisguiseCreatureOracle, 5, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castModal(t, g, "Expose the Culprit", "Instant", stdB40ExposeOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	c := pendingChooseCards(t, g, me.ID)
	if !hasID(c.ChooseCards, creeper) || hasID(c.ChooseCards, bear) {
		t.Fatalf("offered %v, want only the disguise creature", c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{creeper}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	passPriorityAroundTable(t, g)
	// It comes back as a new object (CR 400.7).
	creeper = findBattlefieldByName(g, "Essence of Antiquity")
	if creeper == uuid.Nil {
		t.Fatal("the creature did not come back")
	}
	got := findCardAnywhere(t, g, creeper)
	if !g.Battlefield.Contains(creeper) || !got.FaceDownIsPermanent() || got.FaceDownKind != game.FaceDownCloaked {
		t.Fatalf("the creature is not back as a cloaked permanent: %+v", got)
	}
	if p, tough := mkmMPPT(t, g, creeper); p != 2 || tough != 2 {
		t.Errorf("cloaked it is %d/%d, want 2/2", p, tough)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the Bear was touched")
	}
}
