package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b100_cards_test.go — Commander Masters creatures, GAP batch 100
// (#5948). Each test plays the real catalog entry through the engine.

const (
	gapCryptSliverOracle      = "b0bc2d60-20d0-4ca2-be86-65b297226ac7"
	gapDemonlordBelzenlokOrcl = "667ece0e-a474-4e11-8ac4-6c8fc354457c"
	gapLazotepSliverOracle    = "2017d355-403a-4408-9779-94d1125e373f"
	gapLegionVanguardOracle   = "9f2c8398-1606-400b-a593-2635656e8c4b"
	gapLotlethGiantOracle     = "5e0bd136-e7e9-49f0-910b-a9fb0b9c481a"
	gapMindwrackHarpyOracle   = "f20f27d1-2c50-41d5-b1af-44a0e7ca2e00"
	gapPriestBloodRiteOracle  = "441cdd7d-6c4c-4a91-a118-7eaa35276660"
	gapSerratedScorpionOracle = "abbdc656-02e9-43a0-8e2c-263d6c37c273"
)

func TestGapB100EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		gapCryptSliverOracle:      "Crypt Sliver",
		gapDemonlordBelzenlokOrcl: "Demonlord Belzenlok",
		gapLazotepSliverOracle:    "Lazotep Sliver",
		gapLegionVanguardOracle:   "Legion Vanguard",
		gapLotlethGiantOracle:     "Lotleth Giant",
		gapMindwrackHarpyOracle:   "Mindwrack Harpy",
		gapPriestBloodRiteOracle:  "Priest of the Blood Rite",
		gapSerratedScorpionOracle: "Serrated Scorpion",
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

// --- Crypt Sliver ---

func TestGapB100CryptSliverLetsASliverTapToRegenerateAnySliver(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Crypt Sliver", "Creature — Sliver", gapCryptSliverOracle, false)
	mine := tspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	theirs := tspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
	bear := seedBear(g, opp.ID)
	idx, ref := grantedActivatedIndex(t, g, mine)
	if idx < 0 {
		t.Fatal("a Sliver has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, idx, game.ActivateAbilityParams{Ref: ref, Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Error("the ability targeted a non-Sliver")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, idx, game.ActivateAbilityParams{Ref: ref, Targets: cardRefs(theirs), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, mine); c == nil || !c.Tapped {
		t.Error("the activating Sliver was not tapped")
	}
	_ = g.DestroyPermanentForEffect(theirs)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the targeted Sliver was destroyed despite its regeneration shield")
	}
}

// --- Lazotep Sliver ---

func TestGapB100LazotepSliverGivesYourSliversAfflictTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lazotep := b12Push(g, me.ID, "Lazotep Sliver", "Creature — Zombie Sliver", gapLazotepSliverOracle, 4, 4)
	sliver := tspB1Sliver(g, me.ID, "Other Sliver", 2, 2)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	theirs := tspB1Sliver(g, opp.ID, "Their Sliver", 2, 2)
	blockerA := b12Creature(g, opp.ID, "Blocker A", "Creature — Wall", 0, 9)
	blockerB := b12Creature(g, opp.ID, "Blocker B", "Creature — Wall", 0, 9)
	life := opp.Life
	brAttack(t, g, lazotep, sliver, human)
	if err := g.DeclareBlocker(blockerA, sliver); err != nil {
		t.Fatalf("block: %v", err)
	}
	if err := g.DeclareBlocker(blockerB, human); err != nil {
		t.Fatalf("block: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d -> %d, want 2 lost to the blocked Sliver only", life, opp.Life)
	}
	_ = theirs
}

func TestGapB100LazotepSliverAmassesWhenANontokenSliverOfYoursDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Lazotep Sliver", "Creature — Zombie Sliver", gapLazotepSliverOracle, false)
	mine := tspB1Sliver(g, me.ID, "My Sliver", 1, 1)
	theirs := tspB1Sliver(g, opp.ID, "Their Sliver", 1, 1)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	armies := func() (n, counters int, isSliver bool) {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Army") {
				n++
				counters += c.Counters[game.CounterPlusOne]
				isSliver = c.HasSubtype("Sliver")
			}
		}
		return
	}
	for _, other := range []uuid.UUID{theirs, human} {
		_ = g.DestroyPermanentForEffect(other)
		passPriorityAroundTable(t, g)
		if n, _, _ := armies(); n != 0 {
			t.Fatal("an Army was made for a creature that is not a Sliver of yours")
		}
	}
	_ = g.DestroyPermanentForEffect(mine)
	passPriorityAroundTable(t, g)
	n, counters, isSliver := armies()
	if n != 1 || counters != 2 || !isSliver {
		t.Fatalf("Armies %d with %d counters (Sliver %v), want one Sliver Army with two", n, counters, isSliver)
	}
	// A token Sliver dying does not amass again.
	tok := uuid.Nil
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Army") {
			tok = c.InstanceID
		}
	}
	_ = g.DestroyPermanentForEffect(tok)
	passPriorityAroundTable(t, g)
	if n, _, _ := armies(); n != 0 {
		t.Errorf("a token Sliver's death made %d new Army", n)
	}
}

// --- Legion Vanguard ---

func TestGapB100LegionVanguardExploresWhenYouSacrificeAnotherCreature(t *testing.T) {
	for _, land := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		vanguard := pushCatalogPermanent(g, me.ID, "Legion Vanguard", "Creature — Vampire Soldier", gapLegionVanguardOracle, false)
		fodder := b12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
		top := game.Card{InstanceID: uuid.New(), Name: "Top Card", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID}
		if land {
			top.TypeLine = "Basic Land — Forest"
			top.ManaCost = ""
		}
		g.WithWriteLock(func() { me.Library.PushTop(top) })
		b06AddMana(me, "G")
		if err := g.ActivateCatalogAbility(me.ID, vanguard, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if g.Battlefield.Contains(fodder) {
			t.Error("the creature was not sacrificed")
		}
		passPriorityAroundTable(t, g)
		if land {
			if !me.Hand.Contains(top.InstanceID) || counterOn(g, vanguard, game.CounterPlusOne) != 0 {
				t.Error("a land on top should go to the hand with no counter")
			}
		} else if counterOn(g, vanguard, game.CounterPlusOne) != 1 {
			t.Error("a nonland card should put a +1/+1 counter on the Vanguard")
		}
	}
}

// --- Lotleth Giant ---

func TestGapB100LotlethGiantDealsOneDamagePerCreatureCardInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		pushGraveyardPermanent(me, "Dead Creature", "Creature — Bear", "{1}{G}")
	}
	pushGraveyardPermanent(me, "Dead Rock", "Artifact", "{1}")
	pushGraveyardPermanent(opp, "Their Dead", "Creature — Bear", "{1}{G}")
	life, mine := opp.Life, me.Life
	enterCard(t, g, me.ID, game.Card{Name: "Lotleth Giant", TypeLine: "Creature — Zombie Giant", OracleID: gapLotlethGiantOracle, Power: 6, Toughness: 5})
	if p := latestPickTarget(g, me.ID); p != nil {
		pickPlayerTarget(t, g, me.ID, opp.ID)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 damage", life, opp.Life)
	}
	if me.Life != mine {
		t.Error("the controller took damage")
	}
}

// --- Mindwrack Harpy ---

func TestGapB100MindwrackHarpyMillsEachPlayerThreeAtYourBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		seedCheapLibrary(g, p, 10)
	}
	pushCatalogPermanent(g, me.ID, "Mindwrack Harpy", "Enchantment Creature — Harpy", gapMindwrackHarpyOracle, false)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Graveyard.Size()
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := p.Graveyard.Size() - before[p.ID]; got != 3 {
			t.Errorf("a player milled %d cards, want 3", got)
		}
	}
}

// --- Priest of the Blood Rite ---

func TestGapB100PriestOfTheBloodRiteMakesADemonAndCostsTwoLifeEachUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	enterCard(t, g, me.ID, game.Card{Name: "Priest of the Blood Rite", TypeLine: "Creature — Human Cleric", OracleID: gapPriestBloodRiteOracle, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	var demon *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == me.ID && c.Name == "Demon" {
			demon = c
		}
	}
	if demon == nil || !demon.IsToken() || demon.Power != 5 || demon.Toughness != 5 || len(demon.Colors) != 1 || demon.Colors[0] != "B" {
		t.Fatalf("the token is %+v, want a 5/5 black Demon token", demon)
	}
	assertKeywords(t, g, demon.InstanceID, "flying")
	life := me.Life
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("life %d changed on an opponent's upkeep", me.Life)
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d after your upkeep", me.Life, life-2)
	}
}

// --- Serrated Scorpion ---

func TestGapB100SerratedScorpionDrainsEachOpponentWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	scorpion := pushCatalogPermanent(g, me.ID, "Serrated Scorpion", "Creature — Scorpion", gapSerratedScorpionOracle, false)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	_ = g.DestroyPermanentForEffect(scorpion)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 2
		if p.ID == me.ID {
			want = lives[p.ID] + 2
		}
		if p.Life != want {
			t.Errorf("a player's life is %d, want %d", p.Life, want)
		}
	}
}

// --- Demonlord Belzenlok ---

func TestGapB100DemonlordBelzenlokDigsUntilACheapCardAndDamagesYouPerCard(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cards []game.Card // top first
		hand  int
		dmg   int
	}{
		{"a cheap first card stops at once", []game.Card{
			{Name: "Forest A", TypeLine: "Basic Land — Forest"},
			{Name: "Cheap Spell", TypeLine: "Sorcery", ManaCost: "{2}"},
			{Name: "Never Reached", TypeLine: "Sorcery", ManaCost: "{1}"},
		}, 1, 1},
		{"a mana value 4 or more card repeats the process", []game.Card{
			{Name: "Forest A", TypeLine: "Basic Land — Forest"},
			{Name: "Big Spell", TypeLine: "Sorcery", ManaCost: "{3}{G}"},
			{Name: "Forest B", TypeLine: "Basic Land — Forest"},
			{Name: "Huge Spell", TypeLine: "Sorcery", ManaCost: "{6}"},
			{Name: "Cheap Spell", TypeLine: "Sorcery", ManaCost: "{1}"},
			{Name: "Never Reached", TypeLine: "Sorcery", ManaCost: "{1}"},
		}, 3, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			g.WithWriteLock(func() {
				me.Library.Cards = nil
				for i := len(tc.cards) - 1; i >= 0; i-- {
					c := tc.cards[i]
					c.InstanceID, c.Owner, c.Controller = uuid.New(), me.ID, me.ID
					me.Library.PushTop(c)
				}
			})
			hand, life := me.Hand.Size(), me.Life
			enterCard(t, g, me.ID, game.Card{Name: "Demonlord Belzenlok", TypeLine: "Legendary Creature — Elder Demon",
				OracleID: gapDemonlordBelzenlokOrcl, Power: 6, Toughness: 6})
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.hand {
				t.Errorf("%d cards put into the hand, want %d", got, tc.hand)
			}
			if me.Life != life-tc.dmg {
				t.Errorf("life %d, want %d", me.Life, life-tc.dmg)
			}
			if tc.hand == 3 {
				lands := 0
				for _, c := range g.Exile.Cards {
					if c.Name == "Forest A" || c.Name == "Forest B" {
						lands++
					}
				}
				if lands != 2 {
					t.Errorf("%d lands left in exile, want 2", lands)
				}
			}
		})
	}
}
