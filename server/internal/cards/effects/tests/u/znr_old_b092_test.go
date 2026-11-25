package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// znr_old_b092_test.go — Zendikar Rising counter cards, S82 OLD batch 92
// (#5508). Every test plays the real catalog entry through the engine.

const (
	oldB092ShadowcatOracle   = "043ded41-2506-4dab-91c4-142d9287e7d7"
	oldB092HungerOracle      = "517be4da-9aa0-4a83-a559-962df0450f2c"
	oldB092SpringmantleOracl = "83865eee-709a-4a34-94f1-9b818f31323f"
	oldB092TacticianOracle   = "99262e50-94c1-4796-a7bf-11d7b6164839"
	oldB092GeckoOracle       = "a9596b05-e0bb-4e70-82fe-5db60f87bf3d"
	oldB092AmuletOracle      = "b3998d9c-56b1-486c-818c-a6882e999d54"
	oldB092ClericOracle      = "d7721c5e-b3c5-45c8-b757-935603dfe44b"
	oldB092VerazolOracle     = "d85e5398-fa97-4bcd-8b93-008f3f435e22"
	oldB092SolidarityOracle  = "f903b76f-d405-427a-b115-5066eb311e10"
	oldB092GloomhunterOracle = "70ab65ff-724f-4126-b472-97fa03ca3a34"
	oldB092BearType          = "Creature — Bear"
	oldB092ClericType        = "Creature — Human Cleric"
	oldB092RogueType         = "Creature — Human Rogue"
	oldB092WarriorType       = "Creature — Human Warrior"
	oldB092WizardType        = "Creature — Human Wizard"
)

func TestOldB092EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB092ShadowcatOracle: "Skyclave Shadowcat", oldB092HungerOracle: "Oblivion's Hunger",
		oldB092SpringmantleOracl: "Springmantle Cleric", oldB092TacticianOracle: "Paired Tactician",
		oldB092GeckoOracle: "Vine Gecko", oldB092AmuletOracle: "Relic Amulet",
		oldB092ClericOracle: "Cleric of Life's Bond", oldB092VerazolOracle: "Verazol, the Split Current",
		oldB092SolidarityOracle: "Strength of Solidarity",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB092Cast puts a card with the given real cost into the active player's
// hand, loads exactly the mana the cast needs and casts it.
func oldB092Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, colors []string, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.HandCardFull(me, name, typeLine, cost, oracle, colors)
	if mana != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("mana: %v", err)
		}
	}
	params.Strict = true
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestOldB092SkyclaveShadowcatSacrificesForCountersAndDrawsWhenACounteredCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc18Library(me, 6)
	cat := effectstest.B12Push(g, me.ID, "Skyclave Shadowcat", "Creature — Cat Horror", oldB092ShadowcatOracle, 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, cat, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no other creature to sacrifice")
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB092BearType, 1, 1)
	effectstest.ApaMana(me, "B", "B")
	effectstest.ApaActivate(t, g, me, cat, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if got := effectstest.CountersOn(g, cat, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still there")
	}
	hand := me.Hand.Size()
	plain := effectstest.B12Creature(g, me.ID, "Plain", oldB092BearType, 1, 1)
	effectstest.SbondDie(t, g, me.ID, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("drew for a creature with no counter: hand %d -> %d", hand, me.Hand.Size())
	}
	grown := effectstest.B12Creature(g, me.ID, "Grown", oldB092BearType, 1, 1)
	effectstest.PutCounters(t, g, grown, game.CounterPlusOne, 1)
	effectstest.SbondDie(t, g, me.ID, grown)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1 for a creature with a counter", hand, me.Hand.Size())
	}
}

func TestOldB092OblivionsHungerGrantsIndestructibleAndDrawsOnlyWithACounter(t *testing.T) {
	for _, counter := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		effectstest.Msc18Library(me, 6)
		bear := effectstest.B12Creature(g, me.ID, "Bear", oldB092BearType, 2, 2)
		if counter {
			effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
		}
		hand := me.Hand.Size()
		oldB092Cast(t, g, "Oblivion's Hunger", "Instant", "{1}{B}", oldB092HungerOracle, []string{"B"}, "{B}{B}",
			game.CastSpellParams{Targets: effectstest.CardTarget(bear)})
		if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
			t.Error("no indestructible")
		}
		want := hand // the spell was added and left
		if counter {
			want++
		}
		if me.Hand.Size() != want {
			t.Errorf("counter=%v: hand %d, want %d", counter, me.Hand.Size(), want)
		}
	}
}

func TestOldB092SpringmantleClericEntersWithACounterPerColorSpent(t *testing.T) {
	for _, tc := range []struct {
		mana string
		want int
	}{{"{G}{G}{G}{G}{G}", 1}, {"{G}{G}{R}{R}{U}", 3}} {
		g := effectstest.NewCatalogGame(t)
		id := oldB092Cast(t, g, "Springmantle Cleric", "Creature — Elf Cleric", "{4}{G}", oldB092SpringmantleOracl, []string{"G"}, tc.mana, game.CastSpellParams{})
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.mana, got, tc.want)
		}
	}
}

func TestOldB092PairedTacticianGrowsOnlyWhenAnotherWarriorAttacksToo(t *testing.T) {
	for _, partner := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		tact := effectstest.B12Push(g, me.ID, "Paired Tactician", "Creature — Human Warrior", oldB092TacticianOracle, 3, 2)
		attackers := []uuid.UUID{tact}
		if partner {
			attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Fighter", oldB092WarriorType, 2, 2))
		} else {
			effectstest.B12Creature(g, me.ID, "Idle Fighter", oldB092WarriorType, 2, 2)
			attackers = append(attackers, effectstest.B12Creature(g, me.ID, "Bear", oldB092BearType, 2, 2))
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		for _, id := range attackers {
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountersOn(g, tact, game.CounterPlusOne); (got == 1) != partner {
			t.Errorf("partner=%v: %d counters", partner, got)
		}
	}
}

func TestOldB092VineGeckoDiscountsTheFirstKickedSpellOnlyAndGrowsOnEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	gecko := effectstest.B12Push(g, me.ID, "Vine Gecko", "Creature — Elemental Lizard", oldB092GeckoOracle, 2, 2)
	kicked := func(mana string) error {
		id := effectstest.HandCardFull(me, "Ghastly Gloomhunter", "Creature — Zombie Bat", "{1}{B}", oldB092GloomhunterOracle, []string{"B"})
		me.ManaPool.EmptyPool()
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			return err
		}
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}})
		if err != nil {
			_, _ = me.Hand.Remove(id)
			me.ManaPool.EmptyPool()
		}
		return err
	}
	// {1}{B} plus kicker {3}{B} is six mana: five with the first-kicked discount.
	if err := kicked("{B}{B}{B}{B}"); err == nil {
		t.Fatal("cast a kicked spell for four mana")
	}
	if err := kicked("{B}{B}{B}{B}{B}"); err != nil {
		t.Fatalf("the first kicked spell for five mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := kicked("{B}{B}{B}{B}{B}"); err == nil {
		t.Fatal("the second kicked spell was discounted too")
	}
	if err := kicked("{B}{B}{B}{B}{B}{B}"); err != nil {
		t.Fatalf("the second kicked spell for six mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, gecko, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 (one per kicked spell)", got)
	}
}

func TestOldB092RelicAmuletChargesOnYourSpellsAndDealsThatMuchDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	amulet := effectstest.B12Push(g, me.ID, "Relic Amulet", "Artifact", oldB092AmuletOracle, 0, 0)
	for _, c := range []struct {
		name, typeLine string
	}{{"Quick Instant", "Instant"}, {"Slow Sorcery", "Sorcery"}, {"Other Wizard", oldB092WizardType}, {"Plain Bear", oldB092BearType}} {
		id := effectstest.MayhemHandCard(g, me, c.name, c.typeLine, "{0}", "")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", c.name, err)
		}
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountersOn(g, amulet, game.CounterCharge); got != 3 {
		t.Fatalf("%d charge counters, want 3 (the Bear does not count)", got)
	}
	ogre := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Ogre", "Creature — Ogre", 3, 3)
	effectstest.ApaMana(me, "C", "C")
	effectstest.ApaActivate(t, g, me, amulet, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(ogre)})
	if g.Battlefield.Contains(ogre) {
		t.Error("3 damage did not kill the 3/3")
	}
	if got := effectstest.CountersOn(g, amulet, game.CounterCharge); got != 0 {
		t.Errorf("%d charge counters left, want 0", got)
	}
}

func TestOldB092ClericOfLifesBondGainsForOtherClericsAndGrowsOnTheFirstGainOfTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	cleric := effectstest.B12Push(g, me.ID, "Cleric of Life's Bond", "Creature — Vampire Cleric", oldB092ClericOracle, 2, 2)
	life := me.Life
	oldB092Cast(t, g, "Plain Bear", oldB092BearType, "{0}", "", nil, "", game.CastSpellParams{})
	if me.Life != life {
		t.Error("gained life for a Bear")
	}
	oldB092Cast(t, g, "Other Cleric", oldB092ClericType, "{0}", "", nil, "", game.CastSpellParams{})
	if me.Life != life+1 {
		t.Errorf("life %d, want %d for another Cleric", me.Life, life+1)
	}
	if got := effectstest.CountersOn(g, cleric, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want 1 from the first gain of the turn", got)
	}
	oldB092Cast(t, g, "Second Cleric", oldB092ClericType, "{0}", "", nil, "", game.CastSpellParams{})
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if got := effectstest.CountersOn(g, cleric, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters, want still 1: the second gain is not the first", got)
	}
}

func TestOldB092VerazolEntersWithACounterPerManaSpentAndCopiesAKickedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	// X = 2: {2}{G}{U} is four mana, so four counters.
	verazol := oldB092Cast(t, g, "Verazol, the Split Current", "Legendary Creature — Serpent", "{X}{G}{U}", oldB092VerazolOracle, []string{"G", "U"},
		"{G}{U}{G}{G}", game.CastSpellParams{XValue: 2})
	if got := effectstest.CountersOn(g, verazol, game.CounterPlusOne); got != 4 {
		t.Fatalf("%d counters, want 4 (one per mana spent)", got)
	}
	before := effectstest.CountBattlefieldNamed(g, me.ID, "Ghastly Gloomhunter")
	id := effectstest.HandCardFull(me, "Ghastly Gloomhunter", "Creature — Zombie Bat", "{1}{B}", oldB092GloomhunterOracle, []string{"B"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}{B}{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast kicked: %v", err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, verazol, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 after paying two", got)
	}
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Ghastly Gloomhunter"); got != before+2 {
		t.Errorf("%d Gloomhunters, want %d (the spell and its token copy)", got, before+2)
	}
}

func TestOldB092VerazolDoesNotAskWithoutTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	verazol := effectstest.B12Push(g, me.ID, "Verazol, the Split Current", "Legendary Creature — Serpent", oldB092VerazolOracle, 0, 0)
	effectstest.PutCounters(t, g, verazol, game.CounterPlusOne, 1)
	id := effectstest.HandCardFull(me, "Ghastly Gloomhunter", "Creature — Zombie Bat", "{1}{B}", oldB092GloomhunterOracle, []string{"B"})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}{B}{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatal(err)
	}
	g.RunStateChecksForTest()
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil {
		t.Error("asked to remove counters with only one")
	}
}

func TestOldB092StrengthOfSolidarityPutsACounterPerPartyMember(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	target := effectstest.B12Creature(g, me.ID, "Healer", oldB092ClericType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Thief", oldB092RogueType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Bear", oldB092BearType, 1, 1)
	oldB092Cast(t, g, "Strength of Solidarity", "Sorcery", "{G}", oldB092SolidarityOracle, []string{"G"}, "{G}",
		game.CastSpellParams{Targets: effectstest.CardTarget(target)})
	if got := effectstest.CountersOn(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2 (a party of two)", got)
	}
}
