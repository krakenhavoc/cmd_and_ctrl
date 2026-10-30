package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m9_cards_test.go — Marvel Super Heroes Commander, M–Z batch 9 (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mscM9ShuriOracle     = "94b2a7a6-9e9a-437f-9cde-111b1b02a68e"
	mscM9MirageOracle    = "60733a3a-dfc9-429a-af2d-f0ce9a58d533"
	mscM9TransmitOracle  = "23714aed-0b26-4e21-b290-36c21815471c"
	mscM9FalconOracle    = "359332d6-c42a-4dfa-b148-3296d7e74661"
	mscM9WaspJanetOracle = "8d35c7c8-1342-40b7-82ec-2f7973ee8af8"
	mscM9TVAOracle       = "e9638ad1-4e1d-4721-91f5-8cf13d0d0d0d"
	mscM9StormOracle     = "5d936655-8e46-4e30-8f9c-4fa1a4d34a17"
	mscM9HometownOracle  = "ef40224a-5b49-48b0-ad29-c6ce391548f5"
	mscM9TitaniumOracle  = "84544d35-7acf-4643-965d-66af90fae51b"
	mscM9PortalOracle    = "3ceac6ce-acde-42ab-b79d-ff153074d2d8"
	mscM9WarMachineOracl = "0a512e55-793e-4936-bd27-80d749f4657e"
	mscM9ValeriaOracle   = "89931cf0-f20b-49b5-88c6-2a92aee8c50c"
	mscM9WaspWinsomeOrcl = "4b706959-ea38-4808-870a-7c35d83f7ab2"
)

func TestMscM9EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM9ShuriOracle:     "Shuri, Vibranium Technologist",
		mscM9MirageOracle:    "Mysterio's Mirage",
		mscM9TransmitOracle:  "Team Transmitter",
		mscM9FalconOracle:    "The Falcon, Sam Wilson",
		mscM9WaspJanetOracle: "The Wasp, Janet Van Dyne",
		mscM9TVAOracle:       "TVA Bureaucrat",
		mscM9StormOracle:     "Storm, Shaker of Skies",
		mscM9HometownOracle:  "Spider-Man, Hometown Hero",
		mscM9TitaniumOracle:  "Titanium Man",
		mscM9PortalOracle:    "The Spot's Portal",
		mscM9WarMachineOracl: "War Machine, James Rhodes",
		mscM9ValeriaOracle:   "Valeria Richards, Precocious",
		mscM9WaspWinsomeOrcl: "The Wasp, Winsome Avenger",
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

// mscM9Hand puts a card with the given types in a hand.
func mscM9Hand(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// mscM9Settle passes priority until the table is quiet.
func mscM9Settle(t *testing.T, g *game.Game) {
	t.Helper()
	passPriorityAroundTable(t, g)
}

func mscM9Target(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// --- Shuri, Vibranium Technologist ---------------------------------------------------

func mscM9CastShuri(t *testing.T, g *game.Game, mode int) {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	castAndResolveCreature(t, g, "Shuri, Vibranium Technologist", "Legendary Creature — Human Artificer Hero", mscM9ShuriOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick prompt from the ETB: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
	mscM9Settle(t, g)
}

func TestMscM9ShuriRobotMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	mscM9CastShuri(t, g, 0)

	if got := mshB1Tokens(g, me.ID, "Robot"); got != 1 {
		t.Fatalf("Robot tokens = %d, want 1", got)
	}
	tok := mshB1Token(t, g, me.ID, "Robot")
	if !tok.HasSubtype("Hero") || !tok.IsArtifact() || !tok.IsCreature() || len(tok.Colors) != 0 {
		t.Errorf("token = %s %v, want a colorless Robot Hero artifact creature", tok.TypeLine, tok.Colors)
	}
	if tok.Power != 1 || tok.Toughness != 1 {
		t.Errorf("token is %d/%d, want 1/1", tok.Power, tok.Toughness)
	}
	if !hasEffectiveKeyword(t, g, tok.InstanceID, "flying") {
		t.Error("the token lacks flying")
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d: the Robot bullet draws nothing", me.Hand.Size(), hand)
	}
}

func TestMscM9ShuriDrawMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	mscM9CastShuri(t, g, 1)

	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d after the draw bullet", me.Hand.Size(), hand+1)
	}
	if got := mshB1Tokens(g, me.ID, "Robot"); got != 0 {
		t.Errorf("Robot tokens = %d, want 0 on the draw bullet", got)
	}
}

// --- Mysterio's Mirage ---------------------------------------------------------------

func TestMscM9MirageMakesAnIllusionVillainAfterYouDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Mysterio's Mirage", "Enchantment", mscM9MirageOracle, 0, 0)
	mscM9Hand(me, "Spare", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })

	advanceToEndStepOf(t, g, 0)
	mscM9Settle(t, g)
	if got := mshB1Tokens(g, me.ID, "Illusion"); got != 1 {
		t.Fatalf("Illusion tokens = %d, want 1", got)
	}
	tok := mshB1Token(t, g, me.ID, "Illusion")
	if !tok.HasSubtype("Villain") || tok.Power != 3 || tok.Toughness != 3 || len(tok.Colors) != 1 || tok.Colors[0] != "U" {
		t.Errorf("token = %s %d/%d %v, want a 3/3 blue Illusion Villain", tok.TypeLine, tok.Power, tok.Toughness, tok.Colors)
	}
}

func TestMscM9MirageTriggersOncePerTurnNoMatterHowManyDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Mysterio's Mirage", "Enchantment", mscM9MirageOracle, 0, 0)
	mscM9Hand(me, "Spare 1", "Creature — Bear")
	mscM9Hand(me, "Spare 2", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 2) })

	advanceToEndStepOf(t, g, 0)
	mscM9Settle(t, g)
	if got := mshB1Tokens(g, me.ID, "Illusion"); got != 1 {
		t.Errorf("Illusion tokens = %d after two discards, want 1", got)
	}
}

func TestMscM9MirageDoesNothingWithoutADiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Mysterio's Mirage", "Enchantment", mscM9MirageOracle, 0, 0)

	advanceToEndStepOf(t, g, 0)
	mscM9Settle(t, g)
	if got := mshB1Tokens(g, me.ID, "Illusion"); got != 0 {
		t.Errorf("Illusion tokens = %d with no discard, want 0", got)
	}
}

func TestMscM9MirageIgnoresAnOpponentsDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Mysterio's Mirage", "Enchantment", mscM9MirageOracle, 0, 0)
	mscM9Hand(opp, "Spare", "Creature — Bear")
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 1) })

	advanceToEndStepOf(t, g, 0)
	mscM9Settle(t, g)
	if got := mshB1Tokens(g, me.ID, "Illusion"); got != 0 {
		t.Errorf("Illusion tokens = %d after only an opponent discarded, want 0", got)
	}
}

// --- Team Transmitter ----------------------------------------------------------------

func TestMscM9TransmitterGainsLifeForYourHeroesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Team Transmitter", "Artifact", mscM9TransmitOracle, 0, 0)

	life := me.Life
	castFree(t, g, me, "A Hero", "Creature — Human Hero")
	if me.Life != life+1 {
		t.Errorf("life = %d after a Hero entered, want %d", me.Life, life+1)
	}
	castFree(t, g, me, "A Bear", "Creature — Bear")
	if me.Life != life+1 {
		t.Errorf("life = %d after a non-Hero entered, want %d", me.Life, life+1)
	}
}

func TestMscM9TransmitterTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rock := b12Push(g, me.ID, "Team Transmitter", "Artifact", mscM9TransmitOracle, 0, 0)

	b28TapForMana(t, g, me.ID, rock, "U")
	if me.ManaPool.Total() != 1 {
		t.Errorf("mana pool = %d, want 1", me.ManaPool.Total())
	}
	if !battlefieldCardPtr(t, g, rock).Tapped {
		t.Error("the Transmitter did not tap")
	}
}

// --- The Falcon, Sam Wilson ----------------------------------------------------------

func TestMscM9FalconPutsACounterOnEachOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	a := b12Creature(g, me.ID, "Pal A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Pal B", "Creature — Bear", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	falcon := castAndResolveCreature(t, g, "The Falcon, Sam Wilson", "Legendary Creature — Human Hero", mscM9FalconOracle)
	mscM9Settle(t, g)

	for _, id := range []uuid.UUID{a, b} {
		if got := battlefieldCardPtr(t, g, id).Counters[game.CounterPlusOne]; got != 1 {
			t.Errorf("my creature has %d +1/+1 counters, want 1", got)
		}
	}
	if got := battlefieldCardPtr(t, g, falcon).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("The Falcon has %d counters, want 0 (each other creature)", got)
	}
	if got := battlefieldCardPtr(t, g, theirs).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", got)
	}
	if !hasEffectiveKeyword(t, g, falcon, "flying") {
		t.Error("The Falcon lacks flying")
	}
}

// --- The Wasp, Janet Van Dyne --------------------------------------------------------

func mscM9CastWaspJanet(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	wasp := castAndResolveCreature(t, g, "The Wasp, Janet Van Dyne", "Legendary Creature — Human Hero", mscM9WaspJanetOracle)
	pickCard(t, g, me.ID, target)
	return wasp
}

func TestMscM9WaspJanetDealsFourToATappedCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	big := b12Creature(g, opp.ID, "Tapped Giant", "Creature — Giant", 6, 6)
	battlefieldCardPtr(t, g, big).Tapped = true

	mscM9CastWaspJanet(t, g, big)
	mscM9Settle(t, g)
	if got := battlefieldCardPtr(t, g, big).DamageMarked; got != 4 {
		t.Errorf("damage on the tapped creature = %d, want 4", got)
	}
}

func TestMscM9WaspJanetCannotTargetAnUntappedOrYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	untapped := b12Creature(g, opp.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Tapped Bear", "Creature — Bear", 2, 2)
	battlefieldCardPtr(t, g, mine).Tapped = true
	tapped := b12Creature(g, opp.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	battlefieldCardPtr(t, g, tapped).Tapped = true

	castAndResolveCreature(t, g, "The Wasp, Janet Van Dyne", "Legendary Creature — Human Hero", mscM9WaspJanetOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range []uuid.UUID{untapped, mine} {
		if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(id)); err == nil {
			t.Errorf("an untapped or own creature %s was accepted as a target", id)
		}
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(tapped)); err != nil {
		t.Fatalf("the tapped opposing creature was refused: %v", err)
	}
}

// The damage is dealt by The Wasp's last-known information (CR 608.2h) when
// she has left the battlefield with the ability on the stack.
func TestMscM9WaspJanetStillDamagesAfterLeaving(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	big := b12Creature(g, opp.ID, "Tapped Giant", "Creature — Giant", 6, 6)
	battlefieldCardPtr(t, g, big).Tapped = true

	wasp := mscM9CastWaspJanet(t, g, big)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wasp) })
	mscM9Settle(t, g)
	if g.Battlefield.Contains(wasp) {
		t.Fatal("The Wasp is still on the battlefield")
	}
	if got := battlefieldCardPtr(t, g, big).DamageMarked; got != 4 {
		t.Errorf("damage after The Wasp left = %d, want 4", got)
	}
}

// --- TVA Bureaucrat ------------------------------------------------------------------

func TestMscM9TVABureaucratGrowsAndEvadesOnYourNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tva := b12Push(g, me.ID, "TVA Bureaucrat", "Creature — Human Advisor", mscM9TVAOracle, 1, 3)

	castFree(t, g, me, "A Creature", "Creature — Bear")
	if got := effectivePower(t, g, tva); got != 1 {
		t.Errorf("power after a creature spell = %d, want 1", got)
	}
	assertRestrictions(t, g, tva, 0)

	castFree(t, g, me, "A Sorcery", "Sorcery")
	if got := effectivePower(t, g, tva); got != 2 {
		t.Errorf("power after a noncreature spell = %d, want 2", got)
	}
	assertRestrictions(t, g, tva, game.CantBeBlocked)
}

func TestMscM9TVABureaucratIgnoresAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	tva := b12Push(g, me.ID, "TVA Bureaucrat", "Creature — Human Advisor", mscM9TVAOracle, 1, 3)

	castFree(t, g, opp, "Their Instant", "Instant")
	if got := effectivePower(t, g, tva); got != 1 {
		t.Errorf("power after an opponent's spell = %d, want 1", got)
	}
	assertRestrictions(t, g, tva, 0)
}

// --- Storm, Shaker of Skies ----------------------------------------------------------

func TestMscM9StormDestroysAFlyingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bird := pushPermanent(g, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 2, Toughness: 2, Keywords: []string{"flying"}})
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	castAndResolveCreature(t, g, "Storm, Shaker of Skies", "Legendary Creature — Mutant Hero", mscM9StormOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(bear)); err == nil {
		t.Error("a creature without flying was accepted as a target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(bird)); err != nil {
		t.Fatalf("flying creature refused: %v", err)
	}
	mscM9Settle(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("the flying creature survived")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the ground creature was destroyed")
	}
}

func TestMscM9StormDestroysAnArtifactOrAnEnchantment(t *testing.T) {
	for _, tc := range []struct{ name, typeLine string }{
		{"Rock", "Artifact"},
		{"Aura Thing", "Enchantment"},
	} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		target := b12Push(g, opp.ID, tc.name, tc.typeLine, "", 0, 0)

		castAndResolveCreature(t, g, "Storm, Shaker of Skies", "Legendary Creature — Mutant Hero", mscM9StormOracle)
		pickCard(t, g, me.ID, target)
		mscM9Settle(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("%s survived", tc.name)
		}
	}
}

func TestMscM9StormMayChooseNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rock := b12Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)

	castAndResolveCreature(t, g, "Storm, Shaker of Skies", "Legendary Creature — Mutant Hero", mscM9StormOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	mscM9Settle(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed with no target chosen")
	}
}

// --- Spider-Man, Hometown Hero -------------------------------------------------------

func TestMscM9HometownHeroMakesALowPowerCreatureUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	small := b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)

	spidey := castAndResolveCreature(t, g, "Spider-Man, Hometown Hero", "Legendary Creature — Spider Human Hero", mscM9HometownOracle)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(big)); err == nil {
		t.Error("a creature with power 3 was accepted as a target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(small)); err != nil {
		t.Fatalf("power-2 creature refused: %v", err)
	}
	mscM9Settle(t, g)
	assertRestrictions(t, g, small, game.CantBeBlocked)
	assertRestrictions(t, g, big, 0)
	if !hasEffectiveKeyword(t, g, spidey, "reach") {
		t.Error("Spider-Man lacks reach")
	}
}

// --- Titanium Man --------------------------------------------------------------------

func mscM9TitaniumAttacks(t *testing.T, g *game.Game, mode int) (titanium uuid.UUID, mp *game.PendingChoice) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	titanium = b27Push(g, me.ID, "Titanium Man", "Legendary Artifact Creature — Human Villain", mscM9TitaniumOracle, "{2}{B}{B}", 3, 4, "B")
	battlefieldCardPtr(t, g, titanium).SummonedThisTurn = false
	declareAttack(t, g, opp.ID, titanium)
	mp = latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mp == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mp.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
	return titanium, mp
}

func TestMscM9TitaniumManFlyingMode(t *testing.T) {
	g := newCatalogGame(t)
	titanium, _ := mscM9TitaniumAttacks(t, g, 0)
	mscM9Settle(t, g)
	if !hasEffectiveKeyword(t, g, titanium, "flying") {
		t.Error("Titanium Man did not gain flying")
	}
}

func TestMscM9TitaniumManDamageMode(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	mscM9TitaniumAttacks(t, g, 1)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the damage bullet")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	mscM9Settle(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-1)
	}
}

// --- The Spot's Portal ---------------------------------------------------------------

func TestMscM9PortalTucksTheCreatureAndCostsTwoLifeWithoutAVillain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life

	castCatalogSpell(t, g, "The Spot's Portal", "Instant", mscM9PortalOracle, []game.TargetRef{mscM9Target(bear)})
	mscM9Settle(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature stayed on the battlefield")
	}
	if bottom, err := opp.Library.Bottom(); err != nil || bottom.InstanceID != bear {
		t.Errorf("bottom of the owner's library = %v (%v), want the creature", bottom.Name, err)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
}

func TestMscM9PortalCostsNothingWithAVillain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life

	castCatalogSpell(t, g, "The Spot's Portal", "Instant", mscM9PortalOracle, []game.TargetRef{mscM9Target(bear)})
	mscM9Settle(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature stayed on the battlefield")
	}
	if me.Life != life {
		t.Errorf("life = %d, want %d with a Villain on my side", me.Life, life)
	}
}

// The Villain test is made after the creature has gone: tucking your own
// only Villain leaves you without one.
func TestMscM9PortalOnYourOnlyVillainStillCostsLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	villain := b12Creature(g, me.ID, "Henchman", "Creature — Human Villain", 1, 1)
	life := me.Life

	castCatalogSpell(t, g, "The Spot's Portal", "Instant", mscM9PortalOracle, []game.TargetRef{mscM9Target(villain)})
	mscM9Settle(t, g)
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d: the tucked Villain no longer counts", me.Life, life-2)
	}
}

// --- War Machine, James Rhodes -------------------------------------------------------

func TestMscM9WarMachineTapsATargetOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wm := b27Push(g, me.ID, "War Machine, James Rhodes", "Legendary Artifact Creature — Human Hero", mscM9WarMachineOracl, "{3}{W}", 3, 2, "W", "U")
	battlefieldCardPtr(t, g, wm).SummonedThisTurn = false
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, wm)
	pickCard(t, g, me.ID, blocker)
	mscM9Settle(t, g)
	if !battlefieldCardPtr(t, g, blocker).Tapped {
		t.Error("the targeted creature did not tap")
	}
	if !hasEffectiveKeyword(t, g, wm, "flying") {
		t.Error("War Machine lacks flying")
	}
}

func TestMscM9WarMachineMayChooseNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	wm := b27Push(g, me.ID, "War Machine, James Rhodes", "Legendary Artifact Creature — Human Hero", mscM9WarMachineOracl, "{3}{W}", 3, 2, "W", "U")
	battlefieldCardPtr(t, g, wm).SummonedThisTurn = false
	blocker := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, wm)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing no target: %v", err)
	}
	mscM9Settle(t, g)
	if battlefieldCardPtr(t, g, blocker).Tapped {
		t.Error("a creature tapped with no target chosen")
	}
}

// --- Valeria Richards, Precocious ----------------------------------------------------

func TestMscM9ValeriaTrimsOnlyNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Valeria Richards, Precocious", "Legendary Creature — Human Scientist Hero", mscM9ValeriaOracle, 3, 3)

	spell := handCardFull(me, "A Sorcery", "Sorcery", "{2}{U}", "", []string{"U"})
	creature := handCardFull(me, "A Bear", "Creature — Bear", "{2}{U}", "", []string{"U"})
	price, err := g.PriceCast(me.ID, mscM9HandCard(me, spell), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast sorcery: %v", err)
	}
	if price.Total.Generic != 1 {
		t.Errorf("noncreature spell generic = %d, want 1", price.Total.Generic)
	}
	price, err = g.PriceCast(me.ID, mscM9HandCard(me, creature), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast creature: %v", err)
	}
	if price.Total.Generic != 2 {
		t.Errorf("creature spell generic = %d, want 2", price.Total.Generic)
	}
}

func mscM9HandCard(p *game.Player, id uuid.UUID) game.Card {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	return game.Card{}
}

func TestMscM9ValeriaDrawsOnlyOnTheFirstNoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	seedLibrary(me, "A", "B", "C")
	b12Push(g, me.ID, "Valeria Richards, Precocious", "Legendary Creature — Human Scientist Hero", mscM9ValeriaOracle, 3, 3)

	castFree(t, g, me, "A Creature", "Creature — Bear")
	hand := me.Hand.Size()
	castFree(t, g, me, "First", "Sorcery")
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand = %d after the first noncreature spell, want %d", me.Hand.Size(), hand+1)
	}
	hand = me.Hand.Size()
	castFree(t, g, me, "Second", "Sorcery")
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d after the second noncreature spell, want %d", me.Hand.Size(), hand)
	}
}

// --- The Wasp, Winsome Avenger -------------------------------------------------------

func TestMscM9WaspWinsomeGivesAHeroHexproofAndTapsOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	hero := b12Creature(g, me.ID, "Pal Hero", "Creature — Human Hero", 2, 2)
	bear := b12Creature(g, me.ID, "Pal Bear", "Creature — Bear", 2, 2)

	wasp := castAndResolveCreature(t, g, "The Wasp, Winsome Avenger", "Legendary Creature — Human Hero", mscM9WaspWinsomeOrcl)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(bear)); err == nil {
		t.Error("a non-Hero was accepted as the hexproof target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, mscM9Target(hero)); err != nil {
		t.Fatalf("Hero refused: %v", err)
	}
	mscM9Settle(t, g)
	if !hasEffectiveKeyword(t, g, hero, "hexproof") {
		t.Error("the Hero did not gain hexproof")
	}
	for _, kw := range []string{"flash", "flying"} {
		if !hasEffectiveKeyword(t, g, wasp, kw) {
			t.Errorf("The Wasp lacks %s", kw)
		}
	}

	// A later attack taps a creature the defender controls.
	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	advanceToMain(t, g2)
	w := b27Push(g2, me2.ID, "The Wasp, Winsome Avenger", "Legendary Creature — Human Hero", mscM9WaspWinsomeOrcl, "{1}{U}", 2, 1, "U")
	battlefieldCardPtr(t, g2, w).SummonedThisTurn = false
	victim := b12Creature(g2, opp2.ID, "Victim", "Creature — Bear", 2, 2)
	declareAttack(t, g2, opp2.ID, w)
	pickCard(t, g2, me2.ID, victim)
	mscM9Settle(t, g2)
	if !battlefieldCardPtr(t, g2, victim).Tapped {
		t.Error("the defending player's creature did not tap")
	}
	_ = opp
}
