package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func ringTarget(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// "When this creature enters, the Ring tempts you": the creature is on
// the battlefield, and the only creature, so it becomes the
// Ring-bearer.
func TestRingPoolAEntersTemptsYou(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Enraged Huorn", "Creature — Treefolk", effectstest.EnragedHuornOracle},
		{"Relentless Rohirrim", "Creature — Human Knight", effectstest.RelentlessRohirrimOracle},
		{"Uruk-hai Berserker", "Creature — Orc Berserker", effectstest.UrukHaiBerserkerOracle},
		{"Stalwarts of Osgiliath", "Creature — Human Soldier", effectstest.StalwartsOfOsgiliathOracle},
		{"Bilbo, Retired Burglar", "Legendary Creature — Halfling Rogue", effectstest.BilboRetiredBurglarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.CastRingCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.RingSettle(t, g)
			if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != id {
				t.Fatalf("after entering: tempted %d times, Ring-bearer %v; want 1 and the creature", effectstest.RingCount(g, me.ID), effectstest.RingBearerOf(g, me.ID))
			}
		})
	}
}

// A noncreature permanent that tempts on entering: no creature, so no
// Ring-bearer, and the tempt still counts (CR 701.54d).
func TestRingPoolANoncreatureEntersTemptsYou(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Inherited Envelope", "Artifact", effectstest.InheritedEnvelopeOracle},
		{"Fiery Inscription", "Enchantment", effectstest.FieryInscriptionOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.CastRingCreature(t, g, tc.name, tc.typeLine, tc.oracle, 0, 0)
			effectstest.RingSettle(t, g)
			if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != uuid.Nil {
				t.Fatalf("tempted %d times with Ring-bearer %v; want 1 and none", effectstest.RingCount(g, me.ID), effectstest.RingBearerOf(g, me.ID))
			}
		})
	}
}

// "When this creature dies, the Ring tempts you": it is in the
// graveyard, so the other creature is chosen. Bilbo's "leaves the
// battlefield" is the same.
func TestRingPoolADiesTemptsYou(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Mirrormere Guardian", "Creature — Dwarf Soldier", effectstest.MirrormereGuardianOracle},
		{"Rohirrim Lancer", "Creature — Human Knight", effectstest.RohirrimLancerOracle},
		{"Took Reaper", "Creature — Halfling Peasant", effectstest.TookReaperOracle},
		{"Bilbo, Retired Burglar", "Legendary Creature — Halfling Rogue", effectstest.BilboRetiredBurglarOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			id := effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
			effectstest.RingSettle(t, g)
			if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != bear {
				t.Fatalf("after it died: tempted %d times, Ring-bearer %v; want 1 and the other creature", effectstest.RingCount(g, me.ID), effectstest.RingBearerOf(g, me.ID))
			}
		})
	}
}

// Bilbo: "Whenever Bilbo deals combat damage to a player, create a
// Treasure token."
func TestBilboRetiredBurglarMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bilbo := effectstest.PushCatalogPermanent(g, me.ID, "Bilbo, Retired Burglar", "Legendary Creature — Halfling Rogue", effectstest.BilboRetiredBurglarOracle, false)
	effectstest.DealCombatDamageToPlayer(g, bilbo, opp.ID, 1)
	effectstest.RingSettle(t, g)
	if n := effectstest.OnBattlefieldNamed(g, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after Bilbo connected, want 1", n)
	}
}

// Birthday Escape: "Draw a card. The Ring tempts you."
func TestBirthdayEscapeDrawsAndTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := me.Hand.Size()
	effectstest.RingCastSettled(t, g, "Birthday Escape", "Sorcery", effectstest.BirthdayEscapeOracle)
	if me.Hand.Size() != hand+1 || effectstest.RingCount(g, me.ID) != 1 {
		t.Fatalf("hand %d → %d, tempted %d times; want +1 (the spell left, a card came) and 1", hand, me.Hand.Size(), effectstest.RingCount(g, me.ID))
	}
}

// Bombadil's Song: +1/+1 and hexproof until end of turn, then the
// tempt, which may choose the same creature.
func TestBombadilsSongPumpsAndTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.RingCastSettled(t, g, "Bombadil's Song", "Instant", effectstest.BombadilsSongOracle, ringTarget(bear))
	if p, tt := effectstest.PtOf(t, g, bear); p != 3 || tt != 3 {
		t.Fatalf("the bear is %d/%d, want 3/3", p, tt)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "hexproof") {
		t.Fatal("the bear has no hexproof")
	}
	if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != bear {
		t.Fatal("the Ring did not tempt, or did not choose the only creature")
	}
}

// Claim the Precious: the target is destroyed before the tempt, so the
// only creature left to choose is the caster's own.
func TestClaimThePreciousDestroysThenTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.RingCastSettled(t, g, "Claim the Precious", "Sorcery", effectstest.ClaimThePreciousOracle, ringTarget(theirs))
	if ringOnBattlefield(g, theirs) {
		t.Fatal("the target survived")
	}
	if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != mine {
		t.Fatal("the Ring did not tempt the caster onto their creature")
	}

	// Destroying your own only creature leaves nothing to choose: the
	// tempt waited for the destruction.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	only := effectstest.B12Creature(g2, me2.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.RingCastSettled(t, g2, "Claim the Precious", "Sorcery", effectstest.ClaimThePreciousOracle, ringTarget(only))
	if effectstest.RingCount(g2, me2.ID) != 1 || effectstest.RingBearerOf(g2, me2.ID) != uuid.Nil {
		t.Fatal("the destroyed creature was chosen as the Ring-bearer")
	}
}

// Dreadful as the Storm: base 5/5 until end of turn, counters on top.
func TestDreadfulAsTheStormSetsBaseFiveFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, "+1/+1", 1) })
	effectstest.RingCastSettled(t, g, "Dreadful as the Storm", "Instant", effectstest.DreadfulAsTheStormOracle, ringTarget(bear))
	if p, tt := effectstest.PtOf(t, g, bear); p != 6 || tt != 6 {
		t.Fatalf("the bear is %d/%d, want 6/6 (base 5/5 and a counter)", p, tt)
	}
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt")
	}
}

// Fiery Inscription: "Whenever you cast an instant or sorcery spell,
// this enchantment deals 2 damage to each opponent."
func TestFieryInscriptionBurnsOnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Fiery Inscription", "Enchantment", effectstest.FieryInscriptionOracle, false)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.RingCastSettled(t, g, "Birthday Escape", "Sorcery", effectstest.BirthdayEscapeOracle)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.RingCastSettled(t, g, "Gollum's Bite", "Instant", effectstest.GollumsBiteOracle, ringTarget(giant))
	if p, tt := effectstest.PtOf(t, g, giant); p != 1 || tt != 1 {
		t.Fatalf("the giant is %d/%d, want 1/1", p, tt)
	}
	if effectstest.RingCount(g, me.ID) != 0 {
		t.Fatal("casting Gollum's Bite tempted")
	}

	id, _ := effectstest.ActivateFromGraveyard(t, g, "Gollum's Bite", "Instant", effectstest.GollumsBiteOracle, 0, 0, "{C}{C}{C}{B}", game.ActivateAbilityParams{})
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("the graveyard ability did not tempt")
	}
	if !g.Exile.Contains(id) {
		t.Fatal("the card was not exiled as the cost")
	}
}

// Horses of the Bruinen: both creatures return, then the scry, then the
// tempt.
func TestHorsesOfTheBruinenBouncesScriesAndTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	b := effectstest.B12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Horses of the Bruinen", "Sorcery", effectstest.HorsesOfTheBruinenOracle, []game.TargetRef{ringTarget(a), ringTarget(b)})
	effectstest.RingSettle(t, g)
	if ringOnBattlefield(g, a) || ringOnBattlefield(g, b) {
		t.Fatal("a target is still on the battlefield")
	}
	if effectstest.RingCount(g, me.ID) != 0 {
		t.Fatal("the Ring tempted before the scry was answered")
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.RingSettle(t, g)
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt after the scry")
	}
}

// Ranger's Firebrand: 2 damage to any target, then the tempt.
func TestRangersFirebrandBurnsAndTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	effectstest.RingCastSettled(t, g, "Ranger's Firebrand", "Sorcery", effectstest.RangersFirebrandOracle,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-2 || effectstest.RingCount(g, me.ID) != 1 {
		t.Fatalf("opponent %d → %d, tempted %d; want -2 and 1", life, opp.Life, effectstest.RingCount(g, me.ID))
	}
}

// Sam's Desperate Rescue: a creature card back to hand, then the tempt.
func TestSamsDesperateRescueReturnsAndTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardCardTyped(me, "Grizzly Bears", "Creature — Bear")
	effectstest.RingCastSettled(t, g, "Sam's Desperate Rescue", "Sorcery", effectstest.SamsDesperateRescueOracle, ringTarget(dead))
	if _, ok := effectstest.CardInZone(me.Hand, dead); !ok {
		t.Fatal("the creature card is not in hand")
	}
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt")
	}
}

// Soothing of Sméagol: a nontoken creature returns, then the tempt, so
// the returned creature is not chosen.
func TestSoothingOfSmeagolBouncesThenTempts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	only := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.RingCastSettled(t, g, "Soothing of Sméagol", "Instant", effectstest.SoothingOfSmeagolOracle, ringTarget(only))
	if ringOnBattlefield(g, only) {
		t.Fatal("the target is still on the battlefield")
	}
	if effectstest.RingCount(g, me.ID) != 1 || effectstest.RingBearerOf(g, me.ID) != uuid.Nil {
		t.Fatal("the returned creature was chosen, or the Ring did not tempt")
	}
}

// Stalwarts of Osgiliath: a +1/+1 counter on the second card drawn in
// a turn, not the first.
func TestStalwartsOfOsgiliathGrowsOnTheSecondDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	id := effectstest.PushCatalogPermanent(g, opp.ID, "Stalwarts of Osgiliath", "Creature — Human Soldier", effectstest.StalwartsOfOsgiliathOracle, false)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	effectstest.RingSettle(t, g)
	if n := effectstest.RingBFCard(t, g, id).Counters["+1/+1"]; n != 0 {
		t.Fatalf("%d counters after the first draw, want 0", n)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	effectstest.RingSettle(t, g)
	if n := effectstest.RingBFCard(t, g, id).Counters["+1/+1"]; n != 1 {
		t.Fatalf("%d counters after the second draw, want 1", n)
	}
}

// The Black Breath: opponents' creatures shrink, the caster's do not.
func TestTheBlackBreathShrinksOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.RingCastSettled(t, g, "The Black Breath", "Sorcery", effectstest.TheBlackBreathOracle)
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Fatalf("their giant is %d/%d, want 2/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, mine); p != 2 || tt != 2 {
		t.Fatalf("my bear is %d/%d, want 2/2", p, tt)
	}
	if effectstest.RingCount(g, me.ID) != 1 {
		t.Fatal("the Ring did not tempt")
	}
}
