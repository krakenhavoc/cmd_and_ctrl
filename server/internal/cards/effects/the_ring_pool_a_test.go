package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ring_pool_a_test.go — ADR 0114 PR 5, pool A: the Ring cards
// whose other text is an ordinary spell, an enters or dies trigger, or
// a sibling's ability.

const (
	bilboRetiredBurglarOracle  = "2e7e3be5-04dd-4ff4-b121-3e65b3e00699"
	birthdayEscapeOracle       = "4f958678-2a32-4a30-8fe7-47df9ada6b2f"
	bombadilsSongOracle        = "60092fb4-98db-4f2d-a5d7-24b5f11c8acc"
	claimThePreciousOracle     = "8b0a0991-2327-45ea-b47f-c31db367fa1d"
	dreadfulAsTheStormOracle   = "bd9f4f20-cdef-4a20-b4cb-4f25c137a786"
	enragedHuornOracle         = "8a7dbf4e-f121-40c5-a688-3398e04c1311"
	fieryInscriptionOracle     = "111889d4-bcca-4b1f-ad48-3077e2f5136f"
	gollumsBiteOracle          = "cf0492de-cec5-4455-839f-212246b7e9ea"
	horsesOfTheBruinenOracle   = "100d1c9e-15aa-4bab-9fb3-76c2cd6abc2f"
	inheritedEnvelopeOracle    = "2f1ae834-f6b8-458c-ac0d-3510ddb85193"
	mirrormereGuardianOracle   = "97c188c4-3f27-428e-a4f6-4213f44898fc"
	rangersFirebrandOracle     = "796d3836-d820-440c-b0ad-5907b3cfcb10"
	relentlessRohirrimOracle   = "1212cc74-ff66-479e-8e1b-504446e1c6d9"
	rohirrimLancerOracle       = "504fac88-1674-4157-becb-0204ab8844bf"
	samsDesperateRescueOracle  = "5a3336f1-614a-4b3f-8752-522a34d0417a"
	soothingOfSmeagolOracle    = "35a17b42-9d3f-49ef-8218-2780c5ef5f08"
	stalwartsOfOsgiliathOracle = "d41b2a4f-e137-4e61-99eb-6261ab3edd03"
	tookReaperOracle           = "bfd8daa9-029f-41aa-a947-cdfde3f4d6b6"
	urukHaiBerserkerOracle     = "0e353a41-b523-4c22-9bda-d10460ba99c7"
	theBlackBreathOracle       = "c0afbb32-dc50-436a-bd24-2990c103a4f8"
)

// ringCastSettled casts a catalog spell with targets and settles the
// stack, failing the test on a ring_bearer prompt (arrange at most one
// creature).
func ringCastSettled(t *testing.T, g *game.Game, name, typeLine, oracle string, targets ...game.TargetRef) {
	t.Helper()
	castCatalogSpell(t, g, name, typeLine, oracle, targets)
	ringSettle(t, g)
	if ringPrompt(g, g.Seats[g.Turn.ActiveSeat].ID) != nil {
		t.Fatal("setup: the tempt asked for a Ring-bearer; arrange one creature")
	}
}

func ringTarget(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// "When this creature enters, the Ring tempts you": the creature is on
// the battlefield, and the only creature, so it becomes the
// Ring-bearer.
func TestRingPoolAEntersTemptsYou(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Enraged Huorn", "Creature — Treefolk", enragedHuornOracle},
		{"Relentless Rohirrim", "Creature — Human Knight", relentlessRohirrimOracle},
		{"Uruk-hai Berserker", "Creature — Orc Berserker", urukHaiBerserkerOracle},
		{"Stalwarts of Osgiliath", "Creature — Human Soldier", stalwartsOfOsgiliathOracle},
		{"Bilbo, Retired Burglar", "Legendary Creature — Halfling Rogue", bilboRetiredBurglarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := castRingCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			ringSettle(t, g)
			if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != id {
				t.Fatalf("after entering: tempted %d times, Ring-bearer %v; want 1 and the creature", ringCount(g, me.ID), ringBearerOf(g, me.ID))
			}
		})
	}
}

// A noncreature permanent that tempts on entering: no creature, so no
// Ring-bearer, and the tempt still counts (CR 701.54d).
func TestRingPoolANoncreatureEntersTemptsYou(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Inherited Envelope", "Artifact", inheritedEnvelopeOracle},
		{"Fiery Inscription", "Enchantment", fieryInscriptionOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			castRingCreature(t, g, tc.name, tc.typeLine, tc.oracle, 0, 0)
			ringSettle(t, g)
			if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != uuid.Nil {
				t.Fatalf("tempted %d times with Ring-bearer %v; want 1 and none", ringCount(g, me.ID), ringBearerOf(g, me.ID))
			}
		})
	}
}

// "When this creature dies, the Ring tempts you": it is in the
// graveyard, so the other creature is chosen. Bilbo's "leaves the
// battlefield" is the same.
func TestRingPoolADiesTemptsYou(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Mirrormere Guardian", "Creature — Dwarf Soldier", mirrormereGuardianOracle},
		{"Rohirrim Lancer", "Creature — Human Knight", rohirrimLancerOracle},
		{"Took Reaper", "Creature — Halfling Peasant", tookReaperOracle},
		{"Bilbo, Retired Burglar", "Legendary Creature — Halfling Rogue", bilboRetiredBurglarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			ringSettle(t, g)
			if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != bear {
				t.Fatalf("after it died: tempted %d times, Ring-bearer %v; want 1 and the other creature", ringCount(g, me.ID), ringBearerOf(g, me.ID))
			}
		})
	}
}

// Bilbo: "Whenever Bilbo deals combat damage to a player, create a
// Treasure token."
func TestBilboRetiredBurglarMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bilbo := pushCatalogPermanent(g, me.ID, "Bilbo, Retired Burglar", "Legendary Creature — Halfling Rogue", bilboRetiredBurglarOracle, false)
	dealCombatDamageToPlayer(g, bilbo, opp.ID, 1)
	ringSettle(t, g)
	if n := onBattlefieldNamed(g, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after Bilbo connected, want 1", n)
	}
}

// Birthday Escape: "Draw a card. The Ring tempts you."
func TestBirthdayEscapeDrawsAndTempts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	ringCastSettled(t, g, "Birthday Escape", "Sorcery", birthdayEscapeOracle)
	if me.Hand.Size() != hand+1 || ringCount(g, me.ID) != 1 {
		t.Fatalf("hand %d → %d, tempted %d times; want +1 (the spell left, a card came) and 1", hand, me.Hand.Size(), ringCount(g, me.ID))
	}
}

// Bombadil's Song: +1/+1 and hexproof until end of turn, then the
// tempt, which may choose the same creature.
func TestBombadilsSongPumpsAndTempts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringCastSettled(t, g, "Bombadil's Song", "Instant", bombadilsSongOracle, ringTarget(bear))
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Fatalf("the bear is %d/%d, want 3/3", p, tt)
	}
	if !containsString(effectiveAbilities(t, g, bear), "hexproof") {
		t.Fatal("the bear has no hexproof")
	}
	if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != bear {
		t.Fatal("the Ring did not tempt, or did not choose the only creature")
	}
}

// Claim the Precious: the target is destroyed before the tempt, so the
// only creature left to choose is the caster's own.
func TestClaimThePreciousDestroysThenTempts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringCastSettled(t, g, "Claim the Precious", "Sorcery", claimThePreciousOracle, ringTarget(theirs))
	if ringOnBattlefield(g, theirs) {
		t.Fatal("the target survived")
	}
	if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != mine {
		t.Fatal("the Ring did not tempt the caster onto their creature")
	}

	// Destroying your own only creature leaves nothing to choose: the
	// tempt waited for the destruction.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	only := b12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringCastSettled(t, g2, "Claim the Precious", "Sorcery", claimThePreciousOracle, ringTarget(only))
	if ringCount(g2, me2.ID) != 1 || ringBearerOf(g2, me2.ID) != uuid.Nil {
		t.Fatal("the destroyed creature was chosen as the Ring-bearer")
	}
}

// Dreadful as the Storm: base 5/5 until end of turn, counters on top.
func TestDreadfulAsTheStormSetsBaseFiveFive(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 1) })
	ringCastSettled(t, g, "Dreadful as the Storm", "Instant", dreadfulAsTheStormOracle, ringTarget(bear))
	if p, tt := ptOf(t, g, bear); p != 6 || tt != 6 {
		t.Fatalf("the bear is %d/%d, want 6/6 (base 5/5 and a counter)", p, tt)
	}
	if ringCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt")
	}
}

// Fiery Inscription: "Whenever you cast an instant or sorcery spell,
// this enchantment deals 2 damage to each opponent."
func TestFieryInscriptionBurnsOnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Fiery Inscription", "Enchantment", fieryInscriptionOracle, false)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	ringCastSettled(t, g, "Birthday Escape", "Sorcery", birthdayEscapeOracle)
	for _, p := range g.Seats[1:] {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("opponent at %d, want %d", p.Life, lives[p.ID]-2)
		}
	}
	if me.Life != lives[me.ID] {
		t.Errorf("the caster's life changed to %d", me.Life)
	}
}

// Gollum's Bite: -2/-2, and from the graveyard "{3}{B}, Exile this
// card from your graveyard: The Ring tempts you."
func TestGollumsBiteShrinksAndTemptsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringCastSettled(t, g, "Gollum's Bite", "Instant", gollumsBiteOracle, ringTarget(giant))
	if p, tt := ptOf(t, g, giant); p != 1 || tt != 1 {
		t.Fatalf("the giant is %d/%d, want 1/1", p, tt)
	}
	if ringCount(g, me.ID) != 0 {
		t.Fatal("casting Gollum's Bite tempted")
	}

	id, _ := activateFromGraveyard(t, g, "Gollum's Bite", "Instant", gollumsBiteOracle, 0, 0, "{C}{C}{C}{B}", game.ActivateAbilityParams{})
	ringSettle(t, g)
	if ringCount(g, me.ID) != 1 {
		t.Fatal("the graveyard ability did not tempt")
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the card was not exiled as the cost")
	}
}

// Horses of the Bruinen: both creatures return, then the scry, then the
// tempt.
func TestHorsesOfTheBruinenBouncesScriesAndTempts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	b := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Horses of the Bruinen", "Sorcery", horsesOfTheBruinenOracle, []game.TargetRef{ringTarget(a), ringTarget(b)})
	ringSettle(t, g)
	if ringOnBattlefield(g, a) || ringOnBattlefield(g, b) {
		t.Fatal("a target is still on the battlefield")
	}
	if ringCount(g, me.ID) != 0 {
		t.Fatal("the Ring tempted before the scry was answered")
	}
	answerScryKeepAll(t, g, me.ID)
	ringSettle(t, g)
	if ringCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt after the scry")
	}
}

// Ranger's Firebrand: 2 damage to any target, then the tempt.
func TestRangersFirebrandBurnsAndTempts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	ringCastSettled(t, g, "Ranger's Firebrand", "Sorcery", rangersFirebrandOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-2 || ringCount(g, me.ID) != 1 {
		t.Fatalf("opponent %d → %d, tempted %d; want -2 and 1", life, opp.Life, ringCount(g, me.ID))
	}
}

// Sam's Desperate Rescue: a creature card back to hand, then the tempt.
func TestSamsDesperateRescueReturnsAndTempts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardCardTyped(me, "Grizzly Bears", "Creature — Bear")
	ringCastSettled(t, g, "Sam's Desperate Rescue", "Sorcery", samsDesperateRescueOracle, ringTarget(dead))
	if _, ok := cardInZone(me.Hand, dead); !ok {
		t.Fatal("the creature card is not in hand")
	}
	if ringCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt")
	}
}

// Soothing of Sméagol: a nontoken creature returns, then the tempt, so
// the returned creature is not chosen.
func TestSoothingOfSmeagolBouncesThenTempts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	only := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	ringCastSettled(t, g, "Soothing of Sméagol", "Instant", soothingOfSmeagolOracle, ringTarget(only))
	if ringOnBattlefield(g, only) {
		t.Fatal("the target is still on the battlefield")
	}
	if ringCount(g, me.ID) != 1 || ringBearerOf(g, me.ID) != uuid.Nil {
		t.Fatal("the returned creature was chosen, or the Ring did not tempt")
	}
}

// Stalwarts of Osgiliath: a +1/+1 counter on the second card drawn in
// a turn, not the first.
func TestStalwartsOfOsgiliathGrowsOnTheSecondDraw(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	id := pushCatalogPermanent(g, opp.ID, "Stalwarts of Osgiliath", "Creature — Human Soldier", stalwartsOfOsgiliathOracle, false)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	ringSettle(t, g)
	if n := ringBFCard(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Fatalf("%d counters after the first draw, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	ringSettle(t, g)
	if n := ringBFCard(t, g, id).Counters["+1/+1"]; n != 1 {
		t.Fatalf("%d counters after the second draw, want 1", n)
	}
}

// The Black Breath: opponents' creatures shrink, the caster's do not.
func TestTheBlackBreathShrinksOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ringCastSettled(t, g, "The Black Breath", "Sorcery", theBlackBreathOracle)
	if p, tt := ptOf(t, g, theirs); p != 2 || tt != 2 {
		t.Fatalf("their giant is %d/%d, want 2/2", p, tt)
	}
	if p, tt := ptOf(t, g, mine); p != 2 || tt != 2 {
		t.Fatalf("my bear is %d/%d, want 2/2", p, tt)
	}
	if ringCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt")
	}
}
