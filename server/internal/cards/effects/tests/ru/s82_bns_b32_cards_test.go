package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b32_cards_test.go — Foundations Jumpstart, BNS batch 32 (#5929).

const (
	bnsB32MosquitoOracle  = "f8fca88f-5eab-4854-913c-44261720a4d2"
	bnsB32WhelpOracle     = "a9b1dd4a-2eaf-4448-a767-da705d3ec32b"
	bnsB32HeraldOracle    = "8331fcd2-b91d-45eb-8361-74aa73e4ff69"
	bnsB32HeroOracle      = "fc594e38-4297-44d4-b11b-3ab804cfcbdd"
	bnsB32BlightfangOracl = "342eb666-838c-43d0-bf52-2f704f2a80e3"
	bnsB32BansheeOracle   = "6156bd59-d623-4612-8e86-b5228d5de750"
	bnsB32YouthsOracle    = "9c239027-c5cb-43c6-b4f3-93c269a1d853"
	bnsB32NeonateOracle   = "a9a1d92d-150b-48e1-b41b-dd746b7c1626"
	bnsB32DivinerOracle   = "56c930c4-77e8-4547-8d55-663dc68ccf13"
	bnsB32RangerOracle    = "0b5d174c-7d4d-4c31-a582-8169d3bc932a"
	bnsB32MascotOracle    = "d160887f-f78e-4a66-9766-0b7dd852a784"
)

func TestBnsB32EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB32MosquitoOracle:  "Heartstabber Mosquito",
		bnsB32WhelpOracle:     "Hellkite Whelp",
		bnsB32HeraldOracle:    "Herald of the Sun",
		bnsB32HeroOracle:      "Hero of the Games",
		bnsB32BlightfangOracl: "Hooded Blightfang",
		bnsB32BansheeOracle:   "Howling Banshee",
		bnsB32YouthsOracle:    "Incorrigible Youths",
		bnsB32NeonateOracle:   "Insolent Neonate",
		bnsB32DivinerOracle:   "Ixalli's Diviner",
		bnsB32RangerOracle:    "Jadelight Ranger",
		bnsB32MascotOracle:    "Jubilant Mascot",
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

func TestBnsB32HeartstabberMosquitoDestroysOnlyWhenKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Heartstabber Mosquito", "Creature — Insect", bnsB32MosquitoOracle, nil, nil, nil); err != nil {
		t.Fatalf("unkicked CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Fatal("an unkicked Mosquito triggered")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("an unkicked Mosquito destroyed a creature")
	}

	if _, err := effectstest.CastWithOptionalCosts(t, g, "Heartstabber Mosquito", "Creature — Insect", bnsB32MosquitoOracle, nil, []int{0}, nil); err != nil {
		t.Fatalf("kicked CastSpell: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the kicked Mosquito did not destroy the creature")
	}
}

func TestBnsB32HellkiteWhelpPingsOnlyTheDefendersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := effectstest.BlbB1Seats(g)
	whelp := effectstest.BlbB1Push(g, me.ID, "Hellkite Whelp", "Creature — Dragon", bnsB32WhelpOracle, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, other.ID, "Other Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, whelp)
	if c := effectstest.LatestPickTarget(g, me.ID); c != nil {
		effectstest.PickCard(t, g, me.ID, victim)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 1 {
		t.Errorf("the defender's creature has %d damage, want 1", got)
	}
	if got := effectstest.DamageMarkedOn(g, bystander); got != 0 {
		t.Errorf("a bystander's creature has %d damage, want 0", got)
	}
}

func TestBnsB32HeraldOfTheSunCountersAnotherFlyer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	herald := effectstest.B12Push(g, me.ID, "Herald of the Sun", "Creature — Angel", bnsB32HeraldOracle, 4, 4)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.FloatForTest(g, me, "WWWW")
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(herald)}); err == nil {
		t.Fatal("the Herald targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a creature without flying was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, herald, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(flyer)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, flyer, game.CounterPlusOne); got != 1 {
		t.Errorf("the flyer has %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, herald, game.CounterPlusOne); got != 0 {
		t.Errorf("the Herald has %d counters, want 0", got)
	}
}

func TestBnsB32HeroOfTheGamesPumpsTheTeamWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.AdvanceToMain(t, g)
	hero := effectstest.B12Push(g, me.ID, "Hero of the Games", "Creature — Human Soldier", bnsB32HeroOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	decoy := effectstest.B12Creature(g, me.ID, "Decoy", "Creature — Bear", 2, 2)

	// A spell aimed at another creature does not trigger it.
	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(decoy))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, hero); p != 3 {
		t.Fatalf("Hero is %d power after a spell targeted another creature, want 3", p)
	}
	g.WithWriteLock(func() {})

	effectstest.CastCatalogSpell(t, g, "Hearts on Fire", "Instant", effectstest.J25b1HeartsOnFireOracle, effectstest.CardRefs(hero))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 3 {
		t.Errorf("Bear power %d, want 3 (the Hero's +1)", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature is %d power, want 2", p)
	}
}

func TestBnsB32HowlingBansheeDrainsEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats {
		before[s.ID] = s.Life
	}
	effectstest.CastHoldCreature(t, g, "Howling Banshee", "Creature — Spirit", bnsB32BansheeOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats {
		if got := effectstest.LifeOf(g, s.ID); got != before[s.ID]-3 {
			t.Errorf("seat %s at %d, want %d", s.ID, got, before[s.ID]-3)
		}
	}
}

func TestBnsB32IncorrigibleYouthsHasMadnessAndHaste(t *testing.T) {
	spec, ok := effects.Lookup(bnsB32YouthsOracle)
	if !ok || spec.Madness != "{2}{R}" {
		t.Fatalf("madness cost %q, want {2}{R}", spec.Madness)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	y := effectstest.B12Push(g, me.ID, "Incorrigible Youths", "Creature — Vampire", bnsB32YouthsOracle, 4, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, y, "haste") {
		t.Error("the Youths have no haste")
	}
}

func TestBnsB32InsolentNeonateSacrificesForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	neonate := effectstest.B12Push(g, me.ID, "Insolent Neonate", "Creature — Vampire", bnsB32NeonateOracle, 1, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, neonate, "menace") {
		t.Error("the Neonate has no menace")
	}
	for me.Hand.Size() > 0 {
		me.Hand.Cards = me.Hand.Cards[:0]
	}
	if err := g.ActivateCatalogAbility(me.ID, neonate, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no card to discard")
	}
	if !g.Battlefield.Contains(neonate) {
		t.Fatal("the Neonate was sacrificed by a refused activation")
	}
	filler := game.NewCard("Filler", me.ID)
	filler.TypeLine = "Sorcery"
	me.Hand.PushTop(filler)
	effectstest.PutOnTopForTest(me, "Drawn Card", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, neonate, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{filler.InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(neonate) {
		t.Error("the Neonate was not sacrificed")
	}
	if !me.Graveyard.Contains(filler.InstanceID) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d, want 1 (the drawn card)", me.Hand.Size())
	}
}

func TestBnsB32IxallisDivinerExplores(t *testing.T) {
	t.Run("a land goes to the hand", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		top := effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
		d := effectstest.Msc17Play(t, g, me, game.Card{Name: "Ixalli's Diviner", TypeLine: "Creature — Human Druid", OracleID: bnsB32DivinerOracle, Power: 0, Toughness: 3})
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(top) {
			t.Error("the revealed land did not go to the hand")
		}
		if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 0 {
			t.Errorf("%d counters after a land, want 0", got)
		}
	})
	t.Run("a nonland gives a counter", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
		d := effectstest.Msc17Play(t, g, me, game.Card{Name: "Ixalli's Diviner", TypeLine: "Creature — Human Druid", OracleID: bnsB32DivinerOracle, Power: 0, Toughness: 3})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 1 {
			t.Errorf("%d counters after a nonland, want 1", got)
		}
		effectstest.AnswerExploreQuestion(t, g, me.ID, false)
		effectstest.PassPriorityAroundTable(t, g)
	})
}

func TestBnsB32JadelightRangerExploresTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// Two lands: both go to the hand, no counter.
	a := effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
	b := effectstest.PutOnTopForTest(me, "Island", "Basic Land — Island")
	r := effectstest.Msc17Play(t, g, me, game.Card{Name: "Jadelight Ranger", TypeLine: "Creature — Merfolk Scout Ranger", OracleID: bnsB32RangerOracle, Power: 2, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Errorf("both lands should be in hand: %v %v", me.Hand.Contains(a), me.Hand.Contains(b))
	}
	if got := effectstest.CountersOn(g, r, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters after two lands, want 0", got)
	}
}

func TestBnsB32JubilantMascotSupportsOnlyWhenPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		mascot := effectstest.PushCatalogPermanent(g, me.ID, "Jubilant Mascot", "Creature — Homunculus", bnsB32MascotOracle, false)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.B06AddMana(me, "W", "W", "W", "W")
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.LatestPickTarget(g, me.ID)
		if c == nil {
			t.Fatal("no target prompt for support")
		}
		if err := g.ResolvePickTargets(c.ID, me.ID, effectstest.CardRefs(bear)); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 1
		}
		if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != want {
			t.Errorf("pay=%v: Bear has %d counters, want %d", pay, got, want)
		}
		if got := effectstest.CountersOn(g, mascot, game.CounterPlusOne); got != 0 {
			t.Errorf("pay=%v: the Mascot put a counter on itself (%d)", pay, got)
		}
	}
}

func TestBnsB32HoodedBlightfangDrainsForEachDeathtouchAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Hooded Blightfang", "Creature — Snake", bnsB32BlightfangOracl, 1, 4)
	touch := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Toucher", TypeLine: "Creature — Snake",
		Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}, Owner: me.ID, Controller: me.ID})
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.DeclareAttack(t, g, opp.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, me.ID) != myLife || effectstest.LifeOf(g, opp.ID) != oppLife {
		t.Fatalf("a creature without deathtouch drained: %d/%d", effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID))
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, _ := effectstest.BlbB1Seats(g2)
	effectstest.BlbB1Push(g2, me2.ID, "Hooded Blightfang", "Creature — Snake", bnsB32BlightfangOracl, 1, 4)
	touch2 := effectstest.PushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Toucher", TypeLine: "Creature — Snake",
		Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}, Owner: me2.ID, Controller: me2.ID})
	_ = touch
	effectstest.DeclareAttack(t, g2, opp2.ID, touch2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.LifeOf(g2, opp2.ID); got != oppLife-1 {
		t.Errorf("the defender is at %d, want %d (one drain)", got, oppLife-1)
	}
	if got := effectstest.LifeOf(g2, me2.ID); got != myLife+1 {
		t.Errorf("you are at %d, want %d", got, myLife+1)
	}
}

func TestBnsB32HoodedBlightfangDestroysAPlaneswalkerItsDeathtouchCreatureHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	effectstest.BlbB1Push(g, me.ID, "Hooded Blightfang", "Creature — Snake", bnsB32BlightfangOracl, 1, 4)
	touch := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Toucher", TypeLine: "Creature — Snake",
		Power: 1, Toughness: 1, Keywords: []string{"deathtouch"}, Owner: me.ID, Controller: me.ID})
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walkerA := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker A", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 9}, Owner: opp.ID, Controller: opp.ID})
	walkerB := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker B", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 9}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(touch, walkerA); err != nil {
		t.Fatalf("attack walker A: %v", err)
	}
	if err := g.DeclareAttacker(plain, walkerB); err != nil {
		t.Fatalf("attack walker B: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walkerA) {
		t.Error("the planeswalker hit by a deathtouch creature survived")
	}
	if !g.Battlefield.Contains(walkerB) {
		t.Error("the planeswalker hit by a creature without deathtouch was destroyed")
	}
}
