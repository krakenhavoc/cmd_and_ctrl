package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not registered", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestBnsB12FreshFacedRecruitHasFirstStrikeOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Push(g, me.ID, "Fresh-Faced Recruit", "Creature — Human Soldier", bnsB12Recruit, 2, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Fresh-Faced Recruit", "Creature — Human Soldier", bnsB12Recruit, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "first strike") {
		t.Error("no first strike on its controller's turn")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "first strike") {
		t.Error("first strike on another player's turn")
	}
}

func TestBnsB12FrostburnWeirdPumpsAndDiesAtZeroToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	w := effectstest.B12Push(g, me.ID, "Frostburn Weird", "Creature — Weird", bnsB12Weird, 1, 4)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no mana")
	}
	for i := 1; i <= 3; i++ {
		effectstest.B06AddMana(me, "U")
		effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tt := effectstest.EffectivePower(t, g, w), effectstest.EffectiveToughness(t, g, w); p != 1+i || tt != 4-i {
			t.Fatalf("after %d: %d/%d, want %d/%d", i, p, tt, 1+i, 4-i)
		}
	}
	effectstest.B06AddMana(me, "R")
	effectstest.B16Activate(t, g, me.ID, w, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			s := effectstest.B12Push(g, me.ID, "Furious Spinesplitter", "Creature — Ogre Warrior", bnsB12Spinesplit, 3, 3)
			effectstest.AdvanceToMain(t, g)
			for i, n := range tc.hits {
				if n == 0 {
					continue
				}
				effectstest.Pr7Hit(t, g, s, g.Seats[i+1].ID, n)
				if i == 0 && n == 2 {
					effectstest.Pr7Hit(t, g, s, g.Seats[1].ID, 1)
				}
			}
			// Damage to yourself is not an opponent.
			effectstest.Pr7Hit(t, g, s, me.ID, 1)
			effectstest.AdvanceToEndStep(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, s, game.CounterPlusOne); got != tc.want {
				t.Errorf("counters %d, want %d", got, tc.want)
			}
		})
	}
}

func TestBnsB12GlorifierOfDuskGrantsFlyingAndVigilanceForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gl := effectstest.B12Push(g, me.ID, "Glorifier of Dusk", "Creature — Vampire Soldier", bnsB12Glorifier, 4, 4)
	life := me.Life
	for idx, kw := range []string{"flying", "vigilance"} {
		if effectstest.HasEffectiveKeyword(t, g, gl, kw) {
			t.Fatalf("%s before activating", kw)
		}
		effectstest.B16Activate(t, g, me.ID, gl, idx, game.ActivateAbilityParams{Strict: true})
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasEffectiveKeyword(t, g, gl, kw) {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gm := effectstest.B12Push(g, me.ID, "Golgari Guildmage", "Creature — Elf Shaman", bnsB12Guildmage, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dead Land", TypeLine: "Land", Owner: me.ID, Controller: me.ID})

	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(land), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("a land card is not a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, gm, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(dead), SacrificeIDs: []uuid.UUID{fodder}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}

	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{G}")
	effectstest.B16Activate(t, g, me.ID, gm, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(gm)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, gm, game.CounterPlusOne); got != 1 {
		t.Errorf("counters %d, want 1", got)
	}
}

func TestBnsB12GorgingVultureGainsLifePerCreatureCardMilled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.TopCard(me, "Land A", "Land", "")
	effectstest.TopCard(me, "Bear B", "Creature — Bear", "{1}")
	effectstest.TopCard(me, "Spell C", "Instant", "{R}")
	effectstest.TopCard(me, "Bear D", "Creature — Bear", "{1}")
	life := me.Life
	effectstest.AddPool(t, g, me.ID, "{C}{C}{B}")
	id := effectstest.CastCatalogSpell(t, g, "Gorging Vulture", "Creature — Bird", bnsB12Vulture, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	gob := effectstest.B12Push(g, me.ID, "Gravel-Hide Goblin", "Creature — Goblin Shaman", bnsB12Goblin, 2, 1)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, gob, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without green mana")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{G}")
	effectstest.B16Activate(t, g, me.ID, gob, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.EffectivePower(t, g, gob), effectstest.EffectiveToughness(t, g, gob); p != 4 || tt != 3 {
		t.Errorf("the Goblin is %d/%d, want 4/3", p, tt)
	}
}

func TestBnsB12GurmagSwiftwingKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Gurmag Swiftwing", "Creature — Bat", bnsB12Swiftwing, 1, 2)
	effectstest.AssertKeywords(t, g, id, "flying", "first strike", "haste")
}

func TestBnsB12HavocJesterPingsOnYourSacrificesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Havoc Jester", "Creature — Devil", bnsB12Jester, 5, 5)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Creature — Goblin", "", false)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	before := opp.Life

	if err := g.SacrificePermanent(opp.ID, theirs); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != before {
		t.Fatal("an opponent's sacrifice triggered the Jester")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a destroyed creature triggered the Jester")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Mine 2", "Creature — Goblin", "", false)
	if err := g.SacrificePermanent(me.ID, rock); err != nil {
		t.Fatal(err)
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-1)
	}
}

func TestBnsB12HeliumSquirterEntersWithThreeCountersAndGrafts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}{C}{U}")
	id := effectstest.CastCatalogSpell(t, g, "Helium Squirter", "Creature — Beast Mutant", bnsB12Squirter, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, id, game.CounterPlusOne); n != 3 {
		t.Fatalf("%d counters on entry, want 3", n)
	}
	bear := effectstest.CastCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, id, game.CounterPlusOne) != 2 || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Fatalf("graft: squirter %d, bear %d; want 2 and 1", effectstest.CountersOn(g, id, game.CounterPlusOne), effectstest.CountersOn(g, bear, game.CounterPlusOne))
	}
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(plain)}); err == nil {
		t.Fatal("a creature without a +1/+1 counter is not a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			h := effectstest.B12Push(g, me.ID, "Herald of Ilharg", "Creature — Boar Beast", bnsB12Herald, 3, 3)
			effectstest.AdvanceToMain(t, g)
			effectstest.AddPool(t, g, me.ID, tc.pool)
			before := opp.Life
			if _, err := effectstest.CastWithTapParams(t, g, "Test Creature", "Creature — Bear", tc.cost, "", game.CastSpellParams{}); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, h, game.CounterPlusOne); got != 2 {
				t.Errorf("counters %d, want 2", got)
			}
			if opp.Life != before-tc.damage {
				t.Errorf("opponent life %d, want %d", opp.Life, before-tc.damage)
			}
		})
	}
}

func TestBnsB12HeraldOfIlhargIgnoresNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := effectstest.B12Push(g, me.ID, "Herald of Ilharg", "Creature — Boar Beast", bnsB12Herald, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{R}")
	if _, err := effectstest.CastWithTapParams(t, g, "Test Instant", "Instant", "{R}", "", game.CastSpellParams{}); err != nil {
		t.Skipf("noncreature cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, h, game.CounterPlusOne); got != 0 {
		t.Errorf("counters %d, want 0", got)
	}
}

func TestBnsB12HypersonicDragonLetsYouCastSorceriesAtInstantSpeed(t *testing.T) {
	for _, withDragon := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		if withDragon {
			effectstest.B12Push(g, me.ID, "Hypersonic Dragon", "Creature — Dragon", bnsB12HypersonicDr, 4, 4)
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
