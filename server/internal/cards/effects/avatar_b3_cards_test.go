package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b3_cards_test.go — Avatar: The Last Airbender, batch 3 (#4459):
// every card cast, triggered or activated through the engine.

const (
	avB3SandbendersStormOracle  = "dfc8b8de-be1d-4f7f-bb21-4761a0e7db4e"
	avB3SouthernAirTempleOracle = "987c1144-69f6-4441-b0a4-d23f931df81f"
	avB3SukiOracle              = "70ea1a9d-ec52-4c11-8c4d-0d957e5491ad"
	avB3TeamAvatarOracle        = "f95f7b0c-2dd6-46e1-ae02-f51d275e7120"
	avB3UnitedFrontOracle       = "c23f260b-b7d3-4c25-a08a-dabdb99e1791"
	avB3VengefulVillagersOracle = "ad6442c2-7aa2-4c24-8133-f679a3499bde"
	avB3WaterTribeCaptainOracle = "0c6c4921-c6cb-4b4b-ac9d-10ccc89a5d1d"
	avB3WaterTribeRallierOracle = "64acb083-a0a3-46ec-9c93-b2ac36878703"
	avB3YipYipOracle            = "5453550f-9136-4568-baf8-c3d20e33fbb6"
	avB3AccumulateWisdomOracle  = "a63eb744-71c4-447d-9372-c3a734f022c5"
	avB3SanctumOracle           = "132859dd-de66-45c6-8af4-ab5e202a17b0"
)

func TestAvatarB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB3SandbendersStormOracle:  "Sandbenders' Storm",
		avB3SouthernAirTempleOracle: "Southern Air Temple",
		avB3SukiOracle:              "Suki, Courageous Rescuer",
		avB3TeamAvatarOracle:        "Team Avatar",
		avB3UnitedFrontOracle:       "United Front",
		avB3VengefulVillagersOracle: "Vengeful Villagers",
		avB3WaterTribeCaptainOracle: "Water Tribe Captain",
		avB3WaterTribeRallierOracle: "Water Tribe Rallier",
		avB3YipYipOracle:            "Yip Yip!",
		avB3AccumulateWisdomOracle:  "Accumulate Wisdom",
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

// avB3HandSpell puts a catalog card in the active seat's hand.
func avB3HandSpell(g *game.Game, name, typeLine, oracle, manaCost string) uuid.UUID {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Owner: me.ID, Controller: me.ID,
	})
	return id
}

func avB3AddMana(t *testing.T, g *game.Game, who uuid.UUID, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	})
}

func avB3Count(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// --- Sandbenders' Storm ----------------------------------------------------

func TestAvatarB3SandbendersStormDestroysOnlyABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	small := b12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Giant", 4, 4)

	storm := avB3HandSpell(g, "Sandbenders' Storm", "Instant", avB3SandbendersStormOracle, "{3}{W}")
	avB3AddMana(t, g, me.ID, "{C}{C}{C}{W}")
	err := g.CastSpell(me.ID, storm, game.CastSpellParams{Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small, Mode: 0}}})
	if err == nil {
		t.Fatal("Sandbenders' Storm targeted a creature with power 3")
	}
	if err := g.CastSpell(me.ID, storm, game.CastSpellParams{Modes: []int{0}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big, Mode: 0}}}); err != nil {
		t.Fatalf("cast at the 4/4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 4/4 survived")
	}
	if !g.Battlefield.Contains(small) {
		t.Error("the 3/3 was destroyed")
	}
}

func TestAvatarB3SandbendersStormEarthbendsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushEarthbendLand(g, me.ID, "Plains", "Basic Land — Plains")
	b02cCastModalSpell(t, g, "Sandbenders' Storm", "Instant", avB3SandbendersStormOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: land, Mode: 0}})
	passPriorityAroundTable(t, g)
	power, toughness, creatureLand, hasty := earthbentBody(t, g, land)
	if !creatureLand || !hasty || power != 3 || toughness != 3 {
		t.Errorf("the land is %d/%d, creature land %v, haste %v; want a hasty 3/3 land creature", power, toughness, creatureLand, hasty)
	}
}

// --- Southern Air Temple ---------------------------------------------------

func TestAvatarB3SouthernAirTempleCountsShrinesAndGrowsOnAnotherShrine(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Old Shrine", "Enchantment — Shrine")
	b12Permanent(g, opp.ID, "Their Shrine", "Enchantment — Shrine")

	castCatalogSpell(t, g, "Southern Air Temple", "Legendary Enchantment — Shrine", avB3SouthernAirTempleOracle, nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 2 {
		t.Fatalf("the bear has %d counters, want 2 (two Shrines you control)", got)
	}
	if got := plusOnes(g, theirs); got != 0 {
		t.Errorf("the opponent's bear has %d counters, want 0", got)
	}

	castCatalogSpell(t, g, "Sanctum of Fruitful Harvest", "Legendary Enchantment — Shrine", avB3SanctumOracle, nil)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 3 {
		t.Errorf("after another Shrine entered the bear has %d counters, want 3", got)
	}
}

// --- Suki, Courageous Rescuer ----------------------------------------------

func TestAvatarB3SukiAnthemAndOneAllyPerTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	suki := pushCatalogPermanent(g, me.ID, "Suki, Courageous Rescuer", "Legendary Creature — Human Warrior Ally", avB3SukiOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	if p := hobA4Power(t, g, bear); p != 3 {
		t.Errorf("the bear has power %d, want 3", p)
	}
	if p := hobA4Power(t, g, suki); p != 1 {
		t.Errorf("Suki has power %d, want her own 1 (other creatures only)", p)
	}

	for _, id := range []uuid.UUID{bear, wolf} {
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(id); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	if got := avB3Count(g, me.ID, "Ally"); got != 1 {
		t.Errorf("%d Ally tokens after two permanents left, want 1 (only once each turn)", got)
	}
}

func TestAvatarB3SukiDoesNothingOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Suki, Courageous Rescuer", "Legendary Creature — Human Warrior Ally", avB3SukiOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	aangAdvanceToMain(t, g, 1)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if got := avB3Count(g, me.ID, "Ally"); got != 0 {
		t.Errorf("%d Ally tokens on an opponent's turn, want 0", got)
	}
}

// --- Team Avatar -----------------------------------------------------------

func TestAvatarB3TeamAvatarPumpsTheLoneAttackerByCreatureCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Team Avatar", "Enchantment", avB3TeamAvatarOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, me.ID, "My Ox", "Creature — Ox", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 5 {
		t.Errorf("the lone attacker has power %d, want 5 (+3 for three creatures)", got)
	}
}

func TestAvatarB3TeamAvatarDoesNothingWhenTwoAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Team Avatar", "Enchantment", avB3TeamAvatarOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	declareAttack(t, g, opp.ID, bear, wolf)
	passPriorityAroundTable(t, g)
	if a, b := hobA4Power(t, g, bear), hobA4Power(t, g, wolf); a != 2 || b != 2 {
		t.Errorf("powers %d and %d, want 2 and 2", a, b)
	}
}

func TestAvatarB3TeamAvatarDiscardDealsDamageEqualToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	b12Creature(g, me.ID, "My Ox", "Creature — Ox", 2, 2)
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 4)
	team := pushCatalogHandCard(me, "Team Avatar", "Enchantment", avB3TeamAvatarOracle)

	if err := hobB4Activate(t, g, me.ID, team, 0, "{C}{W}", cardRefs(victim)); err == nil {
		t.Fatal("activated with only {1}{W} in the pool")
	}
	if err := hobB4Activate(t, g, me.ID, team, 0, "{W}", cardRefs(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := zoneKindOf(g, team); got != game.ZoneGraveyard {
		t.Fatalf("Team Avatar is in %v, want the graveyard", got)
	}
	passPriorityAroundTable(t, g)
	c := findBattlefieldCardByID(g, victim)
	if c == nil {
		t.Fatal("the 3/4 died to damage equal to three creatures")
	}
	if c.DamageMarked != 3 {
		t.Errorf("the target has %d damage, want exactly 3 (three creatures you control)", c.DamageMarked)
	}
}

// --- United Front ----------------------------------------------------------

func TestAvatarB3UnitedFrontMakesXAlliesThenCountersEveryone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castXSpell(t, g, "United Front", "Sorcery", avB3UnitedFrontOracle, "{X}{W}{W}", 2, nil)
	passPriorityAroundTable(t, g)
	if got := avB3Count(g, me.ID, "Ally"); got != 2 {
		t.Fatalf("%d Ally tokens, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Ally" && plusOnes(g, c.InstanceID) != 1 {
			t.Errorf("an Ally token has %d counters, want 1", plusOnes(g, c.InstanceID))
		}
	}
	if plusOnes(g, bear) != 1 || plusOnes(g, theirs) != 0 {
		t.Errorf("counters: mine %d, theirs %d; want 1, 0", plusOnes(g, bear), plusOnes(g, theirs))
	}
}

// --- Vengeful Villagers ----------------------------------------------------

func avB3VillagersAttack(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	villagers := pushCatalogPermanent(g, me.ID, "Vengeful Villagers", "Creature — Human Citizen", avB3VengefulVillagersOracle, false)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, villagers)
	answerTriggerTargets(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if !isTapped(g, victim) {
		t.Fatal("the target was not tapped")
	}
	return g, me, villagers, fodder, victim
}

func TestAvatarB3VengefulVillagersSacrificeStuns(t *testing.T) {
	g, me, _, fodder, victim := avB3VillagersAttack(t)
	answerChooseCards(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if got := countersOn(g, victim, game.CounterStun); got != 1 {
		t.Errorf("stun counters %d, want 1", got)
	}
}

func TestAvatarB3VengefulVillagersDeclineDoesNotStun(t *testing.T) {
	g, me, _, fodder, victim := avB3VillagersAttack(t)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) {
		t.Error("a creature was sacrificed after declining")
	}
	if got := countersOn(g, victim, game.CounterStun); got != 0 {
		t.Errorf("stun counters %d, want 0", got)
	}
}

// --- Water Tribe Captain ---------------------------------------------------

func TestAvatarB3WaterTribeCaptainPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	captain := pushCatalogPermanent(g, me.ID, "Water Tribe Captain", "Creature — Human Soldier Ally", avB3WaterTribeCaptainOracle, true)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := hobB4Activate(t, g, me.ID, captain, 0, "{C}{C}{C}{C}", nil); err == nil {
		t.Fatal("activated with four mana")
	}
	if err := hobB4Activate(t, g, me.ID, captain, 0, "{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the bear is %d/%d, want 3/3", p, tt)
	}
	if p := effectivePower(t, g, captain); p != 2 {
		t.Errorf("the Captain has power %d, want 2 (the test permanent's 1, +1)", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("the opponent's bear has power %d, want 2", p)
	}
}

// --- Water Tribe Rallier ---------------------------------------------------

func TestAvatarB3WaterTribeRallierWaterbendsAndTakesASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	rallier := pushCatalogPermanent(g, me.ID, "Water Tribe Rallier", "Creature — Human Soldier Ally", avB3WaterTribeRallierOracle, true)
	var tappers []uuid.UUID
	for i := 0; i < 3; i++ {
		tappers = append(tappers, b12Creature(g, me.ID, "Helper", "Creature — Bear", 1, 1))
	}
	tappers = append(tappers, rallier)
	// Top first: Giant, Scout, Spell, Land.
	big, small := uuid.New(), uuid.New()
	for _, c := range []game.Card{
		{InstanceID: uuid.New(), Name: "Land", TypeLine: "Land"},
		{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery"},
		{InstanceID: small, Name: "Scout", TypeLine: "Creature — Human", Power: 3, Toughness: 1},
		{InstanceID: big, Name: "Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5},
	} {
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
	}
	library := me.Library.Size()

	// Four taps and no mana pay four of the five: refused.
	if err := g.ActivateCatalogAbility(me.ID, rallier, 0, game.ActivateAbilityParams{WaterbendIDs: tappers, Strict: true}); err == nil {
		t.Fatal("waterbend {5} paid with four taps and no mana")
	}
	avB3AddMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, rallier, 0, game.ActivateAbilityParams{WaterbendIDs: tappers, Strict: true}); err != nil {
		t.Fatalf("waterbend {5}: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	for _, id := range pick.ChooseCards {
		if id == big {
			t.Error("the power-5 creature was offered")
		}
	}
	answerChooseCards(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(small) {
		t.Error("the power-3 creature is not in hand")
	}
	if me.Library.Size() != library-1 {
		t.Errorf("library %d -> %d, want one card fewer", library, me.Library.Size())
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top == big {
		t.Error("the Giant is still on top; the rest go to the bottom")
	}
}

// --- Yip Yip! --------------------------------------------------------------

func TestAvatarB3YipYipGivesAnAllyFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ally := b12Creature(g, me.ID, "Sokka", "Creature — Human Warrior Ally", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "Yip Yip!", "Instant — Lesson", avB3YipYipOracle, cardRefs(ally))
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Yip Yip!", "Instant — Lesson", avB3YipYipOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)

	for _, tc := range []struct {
		id     uuid.UUID
		flying bool
	}{{ally, true}, {bear, false}} {
		if p, tt := effectivePower(t, g, tc.id), effectiveToughness(t, g, tc.id); p != 4 || tt != 4 {
			t.Errorf("%v is %d/%d, want 4/4", tc.id, p, tt)
		}
		c := findBattlefieldCardByID(g, tc.id)
		if got := game.HasKeyword(c, "flying"); got != tc.flying {
			t.Errorf("%s flying = %v, want %v", c.Name, got, tc.flying)
		}
	}
}

func TestAvatarB3YipYipRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := avB3HandSpell(g, "Yip Yip!", "Instant — Lesson", avB3YipYipOracle, "{W}")
	avB3AddMana(t, g, me.ID, "{W}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(theirs), Strict: true}); err == nil {
		t.Fatal("Yip Yip! targeted a creature an opponent controls")
	}
}

// --- Accumulate Wisdom -----------------------------------------------------

func avB3Lessons(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old Lesson", TypeLine: "Sorcery — Lesson", Owner: p.ID, Controller: p.ID})
	}
}

func TestAvatarB3AccumulateWisdomTakesOneBelowThreeLessons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	avB3Lessons(me, 2)
	ids := seedLibrary(me, "A", "B", "C", "D")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Accumulate Wisdom", "Instant — Lesson", avB3AccumulateWisdomOracle, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 3 {
		t.Fatalf("want a pick of exactly one of three, got %+v", pick)
	}
	answerChooseCards(t, g, me.ID, ids[1])
	if order := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[0], ids[2]}, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 || !me.Hand.Contains(ids[1]) {
		t.Errorf("hand grew by %d, want B alone", got)
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != ids[3] {
		t.Error("D is no longer on top; A and C belong on the bottom")
	}
}

func TestAvatarB3AccumulateWisdomTakesAllThreeWithThreeLessons(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	avB3Lessons(me, 3)
	ids := seedLibrary(me, "A", "B", "C", "D")
	library := me.Library.Size()
	castCatalogSpell(t, g, "Accumulate Wisdom", "Instant — Lesson", avB3AccumulateWisdomOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !me.Hand.Contains(id) {
			t.Errorf("%v is not in hand", id)
		}
	}
	if me.Hand.Contains(ids[3]) || me.Library.Size() != library-3 {
		t.Errorf("library %d -> %d, want three fewer and D still there", library, me.Library.Size())
	}
}

// A wipe on your turn that takes Suki with another permanent still
// triggers: a leaves-the-battlefield ability looks back (CR 603.10a).
func TestAvatarB3SukiSeesAPermanentLeaveBesideHer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	suki := pushCatalogPermanent(g, me.ID, "Suki, Courageous Rescuer", "Legendary Creature — Human Warrior Ally", avB3SukiOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{suki, bear}) })
	passPriorityAroundTable(t, g)
	if got := avB3Count(g, me.ID, "Ally"); got != 1 {
		t.Errorf("%d Ally tokens after Suki and the bear left together, want 1", got)
	}
}
