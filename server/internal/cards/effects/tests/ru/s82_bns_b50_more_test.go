package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func bnsB50Counters(g *game.Game, id uuid.UUID) int {
	if c, ok := effectstest.BattlefieldCard(g, id); ok {
		return c.Counters[game.CounterPlusOne]
	}
	return 0
}

func TestBnsB50EverethSacrificesForCountersAndLifelinkOnATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	evereth := effectstest.B12Push(g, me.ID, "Evereth, Viceroy of Plunder", "Legendary Creature — Vampire Soldier", bnsB50EverethOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	treasure := effectstest.B12Permanent(g, me.ID, "Treasure", "Artifact — Treasure")
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B16Activate(t, g, me.ID, evereth, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if got := bnsB50Counters(g, evereth); got != 1 {
		t.Fatalf("%d counters, want 1", got)
	}
	if effectstest.HasKeywordOf(g, evereth, "lifelink") {
		t.Error("a creature is not a Treasure: no lifelink")
	}
	effectstest.B16Activate(t, g, me.ID, evereth, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}})
	if got := bnsB50Counters(g, evereth); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
	if !effectstest.HasKeywordOf(g, evereth, "lifelink") {
		t.Error("sacrificing a Treasure grants lifelink until end of turn")
	}
	// Another creature or artifact only: it can't sacrifice itself, and a
	// sorcery-speed ability is refused in combat.
	if err := g.ActivateCatalogAbility(me.ID, evereth, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{evereth}}); err == nil {
		t.Error("Evereth sacrificed itself")
	}
	more := effectstest.B12Creature(g, me.ID, "More Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, evereth, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{more}}); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestBnsB50EverethDiesAndYouMayPayForItsPowerInDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	evereth := effectstest.B12Push(g, me.ID, "Evereth, Viceroy of Plunder", "Legendary Creature — Vampire Soldier", bnsB50EverethOracle, 2, 2)
	effectstest.AddCountersForTest(g, evereth, game.CounterPlusOne, 2)
	before := map[uuid.UUID]int{}
	for _, s := range g.Seats[1:] {
		before[s.ID] = s.Life
	}
	myLife := me.Life
	effectstest.B15Destroy(t, g, evereth)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "C", "R")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, s := range g.Seats[1:] {
		if s.Life != before[s.ID]-4 {
			t.Errorf("%s at %d, want %d (4 damage: its last-known power)", s.Name, s.Life, before[s.ID]-4)
		}
	}
	if me.Life != myLife {
		t.Error("the controller isn't an opponent")
	}
}

func TestBnsB50EverethDeclinedOrUnpaidDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	evereth := effectstest.B12Push(g, me.ID, "Evereth, Viceroy of Plunder", "Legendary Creature — Vampire Soldier", bnsB50EverethOracle, 2, 2)
	life := g.Seats[1].Life
	effectstest.B15Destroy(t, g, evereth)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Seats[1].Life != life {
		t.Errorf("life %d → %d: declining the payment deals nothing", life, g.Seats[1].Life)
	}
}

func TestBnsB50ApheliaAttackMayPayForASnake(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		aphelia := effectstest.B12Push(g, me.ID, "Aphelia, Viper Whisperer", "Legendary Creature — Gorgon Assassin", bnsB50ApheliaOracle, 1, 3)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(aphelia, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.B06AddMana(me, "C", "B")
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		snake := effectstest.FindBattlefieldByName(g, "Snake")
		if pay && snake == uuid.Nil {
			t.Error("paying {1}{B/G} creates a Snake")
		}
		if !pay && snake != uuid.Nil {
			t.Error("declining creates no Snake")
		}
		if pay && !effectstest.HasKeywordOf(g, snake, "deathtouch") {
			t.Error("the Snake has deathtouch")
		}
	}
}

func TestBnsB50ApheliaActivationHalvesTheLifeOfAPlayerAGorgonOrSnakeHits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aphelia := effectstest.B12Push(g, me.ID, "Aphelia, Viper Whisperer", "Legendary Creature — Gorgon Assassin", bnsB50ApheliaOracle, 1, 3)
	snake := effectstest.B12Creature(g, me.ID, "Snake", "Token Creature — Snake", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	opp.Life = 40
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "B")
	effectstest.B16Activate(t, g, me.ID, aphelia, 0, game.ActivateAbilityParams{})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{aphelia, snake, bear} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	// 1 + 1 + 2 damage leaves 36; the Gorgon and the Snake together are one
	// "one or more" trigger: half of 36 is 18, so 18 remain.
	if opp.Life != 18 {
		t.Errorf("opponent at %d, want 18 (36 less half, once)", opp.Life)
	}
}

func TestBnsB50ApheliaNonGorgonDamageAloneDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aphelia := effectstest.B12Push(g, me.ID, "Aphelia, Viper Whisperer", "Legendary Creature — Gorgon Assassin", bnsB50ApheliaOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	opp.Life = 40
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "B")
	effectstest.B16Activate(t, g, me.ID, aphelia, 0, game.ActivateAbilityParams{})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 38 {
		t.Errorf("opponent at %d, want 38: a Bear is neither a Gorgon nor a Snake", opp.Life)
	}
}

func TestBnsB50GodsendPumpsAndExilesABlockerAndBansItsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	godsend := effectstest.B12Push(g, me.ID, "Godsend", "Legendary Artifact — Equipment", bnsB50GodsendOracle, 0, 0)
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Bear", 2, 2)
	effectstest.AttachForTest(t, g, godsend, runner)
	if p, tt := effectstest.EffectivePTOf(t, g, runner); p != 5 || tt != 5 {
		t.Fatalf("equipped creature %d/%d, want 5/5", p, tt)
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Names", "Creature — Wall", 0, 4)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, runner, blocker)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	rp := effectstest.LatestRevealPickFor(g, me.ID)
	if rp == nil {
		t.Fatalf("no prompt to exile one of those creatures: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{blocker}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if !g.Exile.Contains(blocker) {
		t.Fatal("the blocker is exiled")
	}
	// The opponent can't cast a spell with the same name as the exiled card.
	same := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: same, Name: "Wall of Names", TypeLine: "Instant", ManaCost: "{0}", Owner: opp.ID, Controller: opp.ID})
	other := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: other, Name: "Something Else", TypeLine: "Instant", ManaCost: "{0}", Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, same, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast a spell named like the exiled card")
	}
	if err := g.CastSpell(opp.ID, other, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("control: a differently named spell is castable: %v", err)
	}
	// You aren't bound by it.
	mine := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mine, Name: "Wall of Names", TypeLine: "Instant", ManaCost: "{0}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the ban binds opponents only: %v", err)
	}
}

func TestBnsB50GodsendMayDeclineAndExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	godsend := effectstest.B12Push(g, me.ID, "Godsend", "Legendary Artifact — Equipment", bnsB50GodsendOracle, 0, 0)
	runner := effectstest.B12Creature(g, me.ID, "Runner", "Creature — Bear", 2, 2)
	effectstest.AttachForTest(t, g, godsend, runner)
	blocker := effectstest.B12Creature(g, opp.ID, "Wall of Names", "Creature — Wall", 0, 4)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, runner, blocker)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	rp := effectstest.LatestRevealPickFor(g, me.ID)
	if rp == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if g.Exile.Contains(blocker) || !g.Battlefield.Contains(blocker) {
		t.Error("declining exiles nothing")
	}
}

func TestBnsB50GodsendTriggersWhenTheEquippedCreatureBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	godsend := effectstest.B12Push(g, opp.ID, "Godsend", "Legendary Artifact — Equipment", bnsB50GodsendOracle, 0, 0)
	guard := effectstest.B12Creature(g, opp.ID, "Guard", "Creature — Bear", 2, 2)
	effectstest.AttachForTest(t, g, godsend, guard)
	attacker := effectstest.B12Creature(g, me.ID, "Raider", "Creature — Bear", 2, 2)
	effectstest.Pr7aBlockedCombat(t, g, opp.ID, attacker, guard)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	rp := effectstest.LatestRevealPickFor(g, opp.ID)
	if rp == nil {
		t.Fatalf("the equipped blocker's trigger offers the attacker: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{attacker}); err != nil {
		t.Fatal(err)
	}
	if !g.Exile.Contains(attacker) {
		t.Error("the attacker is exiled")
	}
}

func TestBnsB50AvernaPutsALandFromACascadeOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Averna, the Chaos Bloom", "Legendary Creature — Elemental Shaman", bnsB50AvernaOracle, 4, 2)
	effectstest.SeedCheapLibrary(g, me, 1)
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	effectstest.CastWithCost(t, g, "Shardless Agent", "Artifact Creature — Human Rogue", "{1}{G}{U}", effectstest.ShardlessAgentOracle)
	effectstest.PassPriorityAroundTable(t, g)
	rp := effectstest.LatestRevealPickFor(g, me.ID)
	if rp == nil {
		t.Fatalf("no land prompt: %+v", g.PendingChoices)
	}
	if effectstest.CountMayCastPrompts(g) != 0 {
		t.Fatal("the land is put before the cast is decided")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	// An exiled card is a new object when it enters.
	entered := effectstest.FindBattlefieldByName(g, "Mountain")
	if entered == uuid.Nil || !effectstest.B27Tapped(t, g, entered) {
		t.Error("the land enters the battlefield tapped")
	}
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Fatalf("%d cast offers after the land, want 1 for the nonland hit", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, true)
	found := false
	for _, c := range g.Stack.Cards {
		if c.Name == "Cheap Spell" {
			found = true
		}
	}
	if !found {
		t.Error("the cascaded spell is cast")
	}
}

func TestBnsB50AvernaMayDeclineTheLandAndItGoesToTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Averna, the Chaos Bloom", "Legendary Creature — Elemental Shaman", bnsB50AvernaOracle, 4, 2)
	effectstest.SeedCheapLibrary(g, me, 1)
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	effectstest.CastWithCost(t, g, "Shardless Agent", "Artifact Creature — Human Rogue", "{1}{G}{U}", effectstest.ShardlessAgentOracle)
	effectstest.PassPriorityAroundTable(t, g)
	rp := effectstest.LatestRevealPickFor(g, me.ID)
	if rp == nil {
		t.Fatal("no land prompt")
	}
	if err := g.ResolveRevealPick(rp.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if effectstest.FindBattlefieldByName(g, "Mountain") != uuid.Nil {
		t.Error("declined: the land stays off the battlefield")
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	inLibrary := false
	for _, c := range me.Library.Cards {
		if c.Name == "Mountain" {
			inLibrary = true
		}
	}
	if !inLibrary {
		t.Error("the uncast and unchosen cards go to the bottom of the library")
	}
}

func TestBnsB50AvernaIsInertWithoutAnyLandInTheCascade(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Averna, the Chaos Bloom", "Legendary Creature — Elemental Shaman", bnsB50AvernaOracle, 4, 2)
	effectstest.SeedCheapLibrary(g, me, 1)
	effectstest.CastWithCost(t, g, "Shardless Agent", "Artifact Creature — Human Rogue", "{1}{G}{U}", effectstest.ShardlessAgentOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestRevealPickFor(g, me.ID) != nil {
		t.Error("no exiled land: no prompt")
	}
	if effectstest.CountMayCastPrompts(g) != 1 {
		t.Error("the normal cast offer is made")
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}
