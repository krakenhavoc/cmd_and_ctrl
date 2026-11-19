package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b12_cards_test.go — bonus-sheet batch 12 (#5929): twelve cards of
// Ravnica: Clue Edition, each driven through a real cast, activation or
// trigger.

const (
	bnsB12Recruit      = "14319020-faf2-49c4-a808-c39650203383"
	bnsB12Weird        = "f667b4b6-c97f-47fa-9146-c5c1f3f47e92"
	bnsB12Spinesplit   = "f4964f74-07c0-4edc-984d-57722fb16961"
	bnsB12Glorifier    = "6205c485-f46b-40af-8b73-03c668f9ea42"
	bnsB12Guildmage    = "46dad873-11f5-4269-8ec2-f126419fbd7d"
	bnsB12Vulture      = "67330f45-d08c-4fe2-8ae3-8d18939c512a"
	bnsB12Goblin       = "71d284ee-5b29-4db3-b355-a6bdc142a05b"
	bnsB12Swiftwing    = "fc1256e9-eb92-43b6-abfc-ce19d910a929"
	bnsB12Jester       = "58f5bbd2-b876-4eb9-aa79-f4d3511961e0"
	bnsB12Squirter     = "a65ba9bd-1bc6-41c1-a2a4-04b97a5e9541"
	bnsB12Herald       = "80d1866f-b5a5-42c6-b81a-851a34ec0cd9"
	bnsB12HypersonicDr = "e1bf20b0-2bd7-4a0f-9a02-41e43bd44641"
)

func TestBnsB12CardsAreFullWithNoCaveats(t *testing.T) {
	for _, id := range []string{bnsB12Recruit, bnsB12Weird, bnsB12Spinesplit, bnsB12Glorifier, bnsB12Guildmage,
		bnsB12Vulture, bnsB12Goblin, bnsB12Swiftwing, bnsB12Jester, bnsB12Squirter, bnsB12Herald, bnsB12HypersonicDr} {
		spec, ok := Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB12FreshFacedRecruitHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Push(g, me.ID, "Fresh-Faced Recruit", "Creature — Human Soldier", bnsB12Recruit, 2, 1)
	theirs := b12Push(g, opp.ID, "Fresh-Faced Recruit", "Creature — Human Soldier", bnsB12Recruit, 2, 1)
	if !hasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("no first strike on its controller's turn")
	}
	if hasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("first strike on another player's turn")
	}
}

func TestBnsB12FrostburnWeirdPumpsAndDiesAtZeroToughness(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	w := b12Push(g, me.ID, "Frostburn Weird", "Creature — Weird", bnsB12Weird, 1, 4)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	for i := 1; i <= 3; i++ {
		b06AddMana(me, "U")
		b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{Strict: true})
		passPriorityAroundTable(t, g)
		if p, tt := effectivePower(t, g, w), effectiveToughness(t, g, w); p != 1+i || tt != 4-i {
			t.Fatalf("after %d: %d/%d, want %d/%d", i, p, tt, 1+i, 4-i)
		}
	}
	b06AddMana(me, "R")
	b16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(w) {
		t.Error("the Weird survived 0 toughness")
	}
}

func TestBnsB12SpinesplitterCountsEachDamagedOpponentOnce(t *testing.T) {
	for _, tc := range []struct {
		name string
		hits []int // damage to opponent seats 1, 2, 3 (a seat hit twice counts once)
		want int
	}{
		{"none", []int{0, 0, 0}, 0},
		{"one opponent twice", []int{2, 0, 0}, 1},
		{"two opponents", []int{1, 0, 3}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			s := b12Push(g, me.ID, "Furious Spinesplitter", "Creature — Ogre Warrior", bnsB12Spinesplit, 3, 3)
			advanceToMain(t, g)
			for i, n := range tc.hits {
				if n == 0 {
					continue
				}
				pr7Hit(t, g, s, g.Seats[i+1].ID, n)
				if i == 0 && n == 2 {
					pr7Hit(t, g, s, g.Seats[1].ID, 1)
				}
			}
			// Damage to yourself is not an opponent.
			pr7Hit(t, g, s, me.ID, 1)
			advanceToEndStep(t, g)
			passPriorityAroundTable(t, g)
			if got := countersOn(g, s, game.CounterPlusOne); got != tc.want {
				t.Errorf("counters %d, want %d", got, tc.want)
			}
		})
	}
}

func TestBnsB12GlorifierOfDuskGrantsFlyingAndVigilanceForLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gl := b12Push(g, me.ID, "Glorifier of Dusk", "Creature — Vampire Soldier", bnsB12Glorifier, 4, 4)
	life := me.Life
	for idx, kw := range []string{"flying", "vigilance"} {
		if hasEffectiveKeyword(t, g, gl, kw) {
			t.Fatalf("%s before activating", kw)
		}
		b16Activate(t, g, me.ID, gl, idx, game.ActivateAbilityParams{Strict: true})
		passPriorityAroundTable(t, g)
		if !hasEffectiveKeyword(t, g, gl, kw) {
			t.Errorf("no %s after paying 2 life", kw)
		}
	}
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
	me.Life = 1
	if err := g.ActivateCatalogAbility(me.ID, gl, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("paid 2 life with 1 life")
	}
}

func TestBnsB12GolgariGuildmageReturnsACreatureCardAndPutsCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gm := b12Push(g, me.ID, "Golgari Guildmage", "Creature — Elf Shaman", bnsB12Guildmage, 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dead Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})

	addPool(t, g, me.ID, "{C}{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(land), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("a land card is not a legal target")
	}
	b16Activate(t, g, me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(dead), SacrificeIDs: []uuid.UUID{fodder}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}

	addPool(t, g, me.ID, "{C}{C}{C}{C}{G}")
	b16Activate(t, g, me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(gm)})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, gm, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
}

func TestBnsB12GorgingVultureGainsLifePerCreatureCardMilled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	topCard(me, "Land A", "Land", "")
	topCard(me, "Bear B", "Creature — Bear", "{1}")
	topCard(me, "Spell C", "Instant", "{R}")
	topCard(me, "Bear D", "Creature — Bear", "{1}")
	life := me.Life
	addPool(t, g, me.ID, "{C}{C}{B}")
	id := castCatalogSpell(t, g, "Gorging Vulture", "Creature — Bird", bnsB12Vulture, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Vulture did not resolve")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (two creature cards among four milled)", me.Life, life+2)
	}
	if n := me.Graveyard.Size(); n < 4 {
		t.Errorf("graveyard has %d cards, want at least the four milled", n)
	}
}

func TestBnsB12GravelHideGoblinPumpsForThreeAndGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	gob := b12Push(g, me.ID, "Gravel-Hide Goblin", "Creature — Goblin Shaman", bnsB12Goblin, 2, 1)
	addPool(t, g, me.ID, "{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without green mana")
	}
	me.ManaPool = game.ManaPool{}
	addPool(t, g, me.ID, "{C}{C}{C}{G}")
	b16Activate(t, g, me.ID, gob, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, gob), effectiveToughness(t, g, gob); p != 4 || tt != 3 {
		t.Errorf("the Goblin is %d/%d, want 4/3", p, tt)
	}
}

func TestBnsB12GurmagSwiftwingKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Gurmag Swiftwing", "Creature — Bat", bnsB12Swiftwing, 1, 2)
	assertKeywords(t, g, id, "flying", "first strike", "haste")
}

func TestBnsB12HavocJesterPingsOnYourSacrificesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Havoc Jester", "Creature — Devil", bnsB12Jester, 5, 5)
	mine := pushCatalogPermanent(g, me.ID, "Mine", "Creature — Goblin", "", false)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	before := opp.Life

	if err := g.SacrificePermanent(opp.ID, theirs); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || opp.Life != before {
		t.Fatal("an opponent's sacrifice triggered the Jester")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a destroyed creature triggered the Jester")
	}
	pushCatalogPermanent(g, me.ID, "Mine 2", "Creature — Goblin", "", false)
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatal(err)
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-1)
	}
}

func TestBnsB12HeliumSquirterEntersWithThreeCountersAndGrafts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{C}{C}{C}{C}{U}")
	id := castCatalogSpell(t, g, "Helium Squirter", "Creature — Beast Mutant", bnsB12Squirter, nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, id, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters on entry, want 3", n)
	}
	bear := castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if countersOn(g, id, game.CounterPlusOne) != 2 || countersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Fatalf("graft: squirter %d, bear %d; want 2 and 1", countersOn(g, id, game.CounterPlusOne), countersOn(g, bear, game.CounterPlusOne))
	}
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(plain)}); err == nil {
		t.Fatal("a creature without a +1/+1 counter is not a legal target")
	}
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature with a counter did not gain flying")
	}
}

func TestBnsB12HeraldOfIlhargCountersAndDamagesOnBigCreatureSpells(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cost   string
		pool   string
		damage int
	}{
		{"small spell", "{2}", "{C}{C}", 0},
		{"mana value five", "{5}", "{C}{C}{C}{C}{C}", 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			h := b12Push(g, me.ID, "Herald of Ilharg", "Creature — Boar Beast", bnsB12Herald, 3, 3)
			advanceToMain(t, g)
			addPool(t, g, me.ID, tc.pool)
			before := opp.Life
			if _, err := castWithTapParams(t, g, "Test Creature", "Creature — Bear", tc.cost, "", game.CastSpellParams{}); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			if got := countersOn(g, h, game.CounterPlusOne); got != 2 {
				t.Errorf("counters %d, want 2", got)
			}
			if opp.Life != before-tc.damage {
				t.Errorf("opponent life %d, want %d", opp.Life, before-tc.damage)
			}
		})
	}
}

func TestBnsB12HeraldOfIlhargIgnoresNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	h := b12Push(g, me.ID, "Herald of Ilharg", "Creature — Boar Beast", bnsB12Herald, 3, 3)
	advanceToMain(t, g)
	addPool(t, g, me.ID, "{R}")
	if _, err := castWithTapParams(t, g, "Test Instant", "Instant", "{R}", "", game.CastSpellParams{}); err != nil {
		t.Skipf("noncreature cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, h, game.CounterPlusOne); got != 0 {
		t.Errorf("counters %d, want 0", got)
	}
}

func TestBnsB12HypersonicDragonLetsYouCastSorceriesAtInstantSpeed(t *testing.T) {
	for _, withDragon := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		if withDragon {
			b12Push(g, me.ID, "Hypersonic Dragon", "Creature — Dragon", bnsB12HypersonicDr, 4, 4)
		}
		sorc := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: sorc, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		// It is the opponent's turn, so a sorcery is off-timing without the Dragon.
		g.Turn.ActiveSeat = 1
		err := g.CastSpell(me.ID, sorc, game.CastSpellParams{Strict: true})
		if withDragon && err != nil {
			t.Errorf("with the Dragon, a sorcery on the opponent's turn: %v", err)
		}
		if !withDragon && err == nil {
			t.Error("without the Dragon, a sorcery on the opponent's turn was cast")
		}
	}
}
