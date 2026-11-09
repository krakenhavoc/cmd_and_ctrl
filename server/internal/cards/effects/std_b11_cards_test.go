package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b11_cards_test.go — Standard set leftovers, batches 11 and 12
// (#4452): each test plays the real catalog entry through the engine.

const (
	sb11JavelineerOracle  = "bd31487e-e95a-4d90-8c85-742aa83c3e90"
	sb11RevenantOracle    = "bbea79bc-71d4-4100-92f4-783c2c4b46da"
	sb11ConquerorOracle   = "f1266a1b-bae6-4ded-b4d5-7586a5b6c5d0"
	sb11GuardianOracle    = "41129dc1-be88-4198-94f9-032817a79bc7"
	sb11MaelstromOracle   = "49e9fba7-8465-4bbb-95db-73a7e149f494"
	sb11DancersOracle     = "83c8f1ca-194e-4cb6-9d5f-029133c1edbc"
	sb11BattlefrontOracle = "892a83a0-a3cb-4cfb-9106-4008e4e75c0f"
	sb11PriestOracle      = "3a38bf69-511f-4851-870f-8a65641d832e"
	sb11RoadwatcherOracle = "0c3d86b6-5dcd-4016-9b4a-d4f66b1e963f"
	sb11HellkiteOracle    = "2a3b29c3-f69d-4671-8907-600d21528153"
	sb11CallOracle        = "3ceb23f5-abb1-4569-a1e4-1eed9a9babcf"
	sb11SarkhanOracle     = "71cb48e2-2328-471c-a04a-7b6461f378f4"
	sb11GlidecarOracle    = "8f40d73b-9106-448e-8649-5d62939f0d38"
	sb11StrategistOracle  = "1b788680-28e3-4378-9607-bd42acb4fbad"
	sb11HarrierOracle     = "23102d8a-3a93-4e83-af5b-401c85ba9dbc"
	sb11DeathBegetsOracle = "53e86135-3b24-4618-bcc0-af4d81e672dd"
)

// --- Grim Javelineer ------------------------------------------------------

func TestSb11GrimJavelineerPumpsAnAttackerAndSurveilsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Grim Javelineer", "Creature — Human Warrior", sb11JavelineerOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if hasID(pick.PickTargetCards, idle) {
		t.Error("a creature that isn't attacking is not a legal target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, bear); p != 3 {
		t.Errorf("the target has power %d, want 3", p)
	}
	killCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("the target died this turn: a surveil 1 prompt")
	}
}

func TestSb11GrimJavelineerSurveilsNothingForAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Grim Javelineer", "Creature — Human Warrior", sb11JavelineerOracle, 3, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear, other)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	killCreature(t, g, me.ID, other)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) != nil {
		t.Error("another creature dying surveils nothing")
	}
}

// --- Qarsi Revenant -------------------------------------------------------

func TestSb11QarsiRevenantRenewPutsThreeKeywordCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushBattlefieldCreature(g, me.ID)
	id, _ := activateFromGraveyard(t, g, "Qarsi Revenant", "Creature — Vampire",
		sb11RevenantOracle, 3, 3, "{B}{B}{B}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile Qarsi Revenant")
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "deathtouch", "lifelink"} {
		if countersOn(g, target, kw) != 1 {
			t.Errorf("no %s counter on the target", kw)
		}
		if !hasEffectiveKeyword(t, g, target, kw) {
			t.Errorf("the target has no %s from its counter", kw)
		}
	}
}

// --- Death Begets Life ----------------------------------------------------

func TestSb11DeathBegetsLifeDrawsForEachPermanentDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ench := b12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	art := b12Permanent(g, opp.ID, "Rock", "Artifact")
	for i := 0; i < 6; i++ {
		pushLibraryTopForTest(me, "Filler")
	}
	before := handSize(me)
	castCatalogSpell(t, g, "Death Begets Life", "Sorcery", sb11DeathBegetsOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, ench} {
		if g.Battlefield.Contains(id) {
			t.Error("a creature or enchantment survived")
		}
	}
	if !g.Battlefield.Contains(art) {
		t.Error("an artifact was destroyed")
	}
	if got := handSize(me) - before; got != 3 {
		t.Errorf("drew %d cards, want 3 (one per permanent destroyed)", got)
	}
}

func TestSb11DeathBegetsLifeDoesNotCountAnIndestructibleCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Golem", Power: 2, Toughness: 2,
		Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 4; i++ {
		pushLibraryTopForTest(me, "Filler")
	}
	before := handSize(me)
	castCatalogSpell(t, g, "Death Begets Life", "Sorcery", sb11DeathBegetsOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible creature was destroyed")
	}
	if got := handSize(me) - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Clarion Conqueror ----------------------------------------------------

func TestSb11ClarionConquerorShutsOffArtifactsCreaturesAndPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	birds := b21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", actBirdsOracle, 0, 1, "G")
	theirBirds := b21Push(g, them.ID, "Birds of Paradise", "Creature — Bird", actBirdsOracle, 0, 1, "G")
	solRing := b21Push(g, me.ID, "Sol Ring", "Artifact", actSolRingOracle, 0, 0)
	forest := b21Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	pw := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("setup: Sol Ring before the Conqueror: %v", err)
	}
	b21Push(g, me.ID, "Clarion Conqueror", "Creature — Dragon", sb11ConquerorOracle, 3, 3, "W")
	for _, c := range []struct {
		who  *game.Player
		id   uuid.UUID
		name string
	}{{me, birds, "your Birds"}, {them, theirBirds, "their Birds"}} {
		if err := g.ActivateManaAbility(c.who.ID, c.id, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
			t.Errorf("%s: err = %v, want ErrCantActivate (mana abilities included)", c.name, err)
		}
	}
	if err := g.ActivateLoyalty(me.ID, pw, "+1", 1); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("a loyalty ability: err = %v, want ErrCantActivate", err)
	}
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); errors.Is(err, game.ErrCantActivate) {
		t.Errorf("a plain land is none of the three types: err = %v", err)
	}
}

// --- Karakyk Guardian -----------------------------------------------------

func sb11Guardian(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := pushDiesCreatureForTest(g, owner, "Karakyk Guardian", sb11GuardianOracle, "Creature — Dragon", 6, 5)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

func TestSb11KarakykGuardianHasHexproofUntilItDealsAnyDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := sb11Guardian(g, me.ID)
	if !effectiveAbilitiesContain(t, g, guardian, "hexproof") || creatureTargetableBy(g, opp.ID, guardian) {
		t.Fatal("before it has dealt damage an opponent can't target it")
	}
	if !creatureTargetableBy(g, me.ID, guardian) {
		t.Fatal("its controller still can")
	}
	// Damage that is fully prevented is not dealt; a zero never marks it.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(guardian, opp.ID, 0); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if !effectiveAbilitiesContain(t, g, guardian, "hexproof") {
		t.Fatal("0 damage is no damage")
	}
	// Noncombat damage counts, unlike Ruric Thar's combat-only gate.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(guardian, opp.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if effectiveAbilitiesContain(t, g, guardian, "hexproof") || !creatureTargetableBy(g, opp.ID, guardian) {
		t.Error("after dealing noncombat damage it has lost hexproof")
	}
}

func TestSb11KarakykGuardianLosesHexproofInCombatAndGetsItBackAsANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := sb11Guardian(g, me.ID)
	attackWith(t, g, opp.ID, guardian)
	if opp.Life != 40-6 {
		t.Fatalf("opponent life %d, want 34", opp.Life)
	}
	if effectiveAbilitiesContain(t, g, guardian, "hexproof") {
		t.Fatal("after combat damage it has lost hexproof")
	}
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(guardian); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	for _, c := range me.Hand.Cards {
		if c.InstanceID == guardian && c.DealtDamage {
			t.Error("the card in hand still remembers dealing damage")
		}
	}
}

// --- Maelstrom of the Spirit Dragon -----------------------------------------

func TestSb11MaelstromManaCastsOnlyDragonAndOmenSpells(t *testing.T) {
	t.Run("a Dragon spell is paid", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		land := pushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", sb11MaelstromOracle, false)
		castFromHandForTest(t, g, me, "Spark Dragon", "Creature — Dragon", "{R}", "", autoPay)
		if !permanentTapped(g, land) {
			t.Error("the Maelstrom was not tapped for the Dragon")
		}
	})
	t.Run("an Elf is refused", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		pushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", sb11MaelstromOracle, false)
		_, err := castFromHandErr(g, me, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", autoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Errorf("an Elf off the restricted mana: %v, want insufficient mana", err)
		}
	})
	t.Run("an Omen spell is paid", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		land := pushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", sb11MaelstromOracle, false)
		params := autoPay
		params.Face = 1
		castFromHand(t, g, tdmOmenCards["Sagu Wildling"], params)
		if !permanentTapped(g, land) {
			t.Error("the Maelstrom was not tapped for the Omen")
		}
	})
	t.Run("an ordinary instant is refused", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		pushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", sb11MaelstromOracle, false)
		_, err := castFromHandErr(g, me, "Shock", "Instant", "{R}", "", autoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Errorf("an ordinary instant off the restricted mana: %v, want insufficient mana", err)
		}
	})
}

func TestSb11MaelstromSacrificesToFetchADragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	land := pushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", sb11MaelstromOracle, false)
	dragon := pushLibraryCardForTest(me, game.Card{Name: "Spark Dragon", TypeLine: "Creature — Dragon"})
	pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Dragon was offered")
	}
	answerSearchByID(t, g, me.ID, dragon)
	if !me.Hand.Contains(dragon) {
		t.Error("the Dragon is not in hand")
	}
}

// --- Lotuslight Dancers -----------------------------------------------------

func sb11DancersSetup(t *testing.T) (g *game.Game, me *game.Player, ids map[string]uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	ids = map[string]uuid.UUID{}
	for _, c := range []game.Card{
		{Name: "Black A", TypeLine: "Creature — Zombie", Colors: []string{"B"}},
		{Name: "Black B", TypeLine: "Creature — Zombie", Colors: []string{"B"}},
		{Name: "Green", TypeLine: "Creature — Elf", Colors: []string{"G"}},
		{Name: "Blue", TypeLine: "Creature — Merfolk", Colors: []string{"U"}},
		{Name: "Three", TypeLine: "Creature — Dragon", Colors: []string{"B", "G", "U"}},
		{Name: "White", TypeLine: "Creature — Human", Colors: []string{"W"}},
	} {
		ids[c.Name] = pushLibraryCardForTest(me, c)
	}
	castAndResolveCreature(t, g, "Lotuslight Dancers", "Creature — Zombie Bard", sb11DancersOracle)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "White") != uuid.Nil {
		t.Error("a white card was offered")
	}
	return g, me, ids
}

func TestSb11LotuslightDancersPutsOneCardPerColorIntoTheGraveyard(t *testing.T) {
	g, me, ids := sb11DancersSetup(t)
	answerSearchByID(t, g, me.ID, ids["Black A"], ids["Green"], ids["Blue"])
	for _, n := range []string{"Black A", "Green", "Blue"} {
		if !me.Graveyard.Contains(ids[n]) {
			t.Errorf("%s is not in the graveyard", n)
		}
	}
	if me.Graveyard.Contains(ids["Black B"]) || me.Graveyard.Contains(ids["White"]) {
		t.Error("an unchosen card is in the graveyard")
	}
}

func TestSb11LotuslightDancersLetsAMulticoloredCardStandForOneColor(t *testing.T) {
	g, me, ids := sb11DancersSetup(t)
	// The three-colour card is the green one; Black A and Blue fill the rest.
	answerSearchByID(t, g, me.ID, ids["Three"], ids["Black A"], ids["Blue"])
	if !me.Graveyard.Contains(ids["Three"]) || !me.Graveyard.Contains(ids["Black A"]) || !me.Graveyard.Contains(ids["Blue"]) {
		t.Error("the three cards were not all found")
	}
}

func TestSb11LotuslightDancersRefusesTwoCardsForTheSameColorAndAllowsFewer(t *testing.T) {
	g, me, ids := sb11DancersSetup(t)
	c := searchChoiceFor(g, me.ID)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids["Black A"], ids["Black B"]}); err == nil {
		t.Fatal("two black cards were accepted for one black quality")
	}
	answerSearchByID(t, g, me.ID, ids["Green"])
	if !me.Graveyard.Contains(ids["Green"]) || me.Graveyard.Size() != 1 {
		t.Errorf("a single card should be found: graveyard %d", me.Graveyard.Size())
	}
}

// --- United Battlefront -----------------------------------------------------

func TestSb11UnitedBattlefrontPutsUpToTwoCheapNoncreatureNonlandPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	deep := plTop(me, "Deep Card", "Artifact", "{1}")
	art4 := plTop(me, "Four Rock", "Artifact", "{4}")
	inst := plTop(me, "Cheap Instant", "Instant", "{1}")
	bear := plTop(me, "Bear", "Creature — Bear", "{2}")
	land := plTop(me, "Forest", "Basic Land — Forest", "")
	art1 := plTop(me, "Rock", "Artifact", "{1}")
	ench := plTop(me, "Rite", "Enchantment", "{2}")
	art2 := plTop(me, "Other Rock", "Artifact", "{3}")
	castCatalogSpell(t, g, "United Battlefront", "Sorcery", sb11BattlefrontOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose_cards prompt")
	}
	for _, bad := range []uuid.UUID{inst, bear, land, art4, deep} {
		if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bad}); err == nil {
			t.Error("an illegal card was accepted")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{art1, ench, art2}); err == nil {
		t.Error("three cards were accepted for up to two")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{art1, ench}); err != nil {
		t.Fatalf("a legal pick: %v", err)
	}
	if !g.Battlefield.Contains(art1) || !g.Battlefield.Contains(ench) {
		t.Error("the two chosen permanents did not enter")
	}
	for _, id := range []uuid.UUID{art2, inst, bear, land, art4} {
		if g.Battlefield.Contains(id) || !me.Library.Contains(id) {
			t.Error("an unchosen card left the library")
		}
	}
	// The five looked-at cards not put are the bottom five, in some order,
	// and the eighth card, never looked at, is not among them.
	bottom := plBottomIDs(me, 5)
	for _, id := range []uuid.UUID{art2, inst, bear, land, art4} {
		if !bottom[id] {
			t.Error("a looked-at card that was not put is not on the bottom")
		}
	}
	if bottom[deep] {
		t.Error("the card below the top seven was moved to the bottom")
	}
}

// --- Severance Priest -----------------------------------------------------

// sb11PriestTable casts the Priest against seat 1, whose hand is a Forest,
// a Lightning Bolt and a Big Bear ({3}{R}{R}); when `take` is true the Bear
// is chosen, otherwise the "you may" is declined.
func sb11PriestTable(t *testing.T, take bool) (g *game.Game, me, opp *game.Player, priest, bear uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := revealHand(opp, rhForest(), rhBolt(),
		game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{R}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3})
	bear = ids[2]
	priest = castCatalogSpell(t, g, "Severance Priest", "Creature — Djinn Cleric", sb11PriestOracle, nil)
	passPriorityAroundTable(t, g)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	pick := openVariant(t, g)
	if !sameIDs(pick.DiscardOptions, ids[1:]) {
		t.Fatalf("options %v: want the two nonland cards", pick.DiscardOptions)
	}
	var picked []uuid.UUID
	if take {
		picked = []uuid.UUID{bear}
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, picked); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	return g, me, opp, priest, bear
}

func TestSb11SeverancePriestExilesACardAndItsOwnerGetsASpiritWhenItLeaves(t *testing.T) {
	g, me, opp, priest, bear := sb11PriestTable(t, true)
	if !g.Exile.Contains(bear) {
		t.Fatal("the chosen card was not exiled")
	}
	killCreature(t, g, me.ID, priest)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := len(battlefieldIDsNamed(g, "Spirit")); got != 1 {
		t.Fatalf("%d Spirits, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Spirit" {
			continue
		}
		if c.Controller != opp.ID {
			t.Error("the exiled card's owner creates the token, not the Priest's controller")
		}
		if p, tt := ptOf(t, g, c.InstanceID); p != 5 || tt != 5 {
			t.Errorf("the Spirit is %d/%d, want 5/5 (the Bear's mana value)", p, tt)
		}
		if !c.HasColor("W") {
			t.Error("the Spirit is not white")
		}
	}
}

func TestSb11SeverancePriestMakesNothingWhenYouDeclineTheCard(t *testing.T) {
	g, me, opp, priest, _ := sb11PriestTable(t, false)
	if g.Exile.Size() != 0 {
		t.Fatal("a card was exiled though the choice was declined")
	}
	if opp.Hand.Size() != 3 {
		t.Errorf("the hand has %d cards, want 3", opp.Hand.Size())
	}
	killCreature(t, g, me.ID, priest)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := len(battlefieldIDsNamed(g, "Spirit")); got != 0 {
		t.Errorf("%d Spirits though nothing was exiled", got)
	}
}

// --- Yathan Roadwatcher -----------------------------------------------------

func TestSb11YathanRoadwatcherMillsFourThenReturnsACheapCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := pushGraveyardCardTyped(me, "Cheap Bear", "Creature — Bear")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == cheap {
			me.Graveyard.Cards[i].ManaCost = "{2}"
		}
	}
	dear := pushGraveyardCardTyped(me, "Dear Bear", "Creature — Bear")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == dear {
			me.Graveyard.Cards[i].ManaCost = "{4}"
		}
	}
	lib := me.Library.Size()
	rfdCast(t, g, "Yathan Roadwatcher", "Creature — Human Scout", sb11RoadwatcherOracle, 3, 3)
	if got := me.Library.Size(); got != lib-4 {
		t.Fatalf("library = %d, want %d (milled four)", got, lib-4)
	}
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if hasID(pick.PickTargetCards, dear) {
		t.Error("a creature card with mana value 4 is not a legal target")
	}
	pickTriggerTarget(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the cheap creature card did not return to the battlefield")
	}
	if g.Battlefield.Contains(dear) {
		t.Error("the mana value 4 creature returned")
	}
}

func TestSb11YathanRoadwatcherDoesNothingWhenItWasNotCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lib := me.Library.Size()
	pushCatalogPermanent(g, me.ID, "Yathan Roadwatcher", "Creature — Human Scout", sb11RoadwatcherOracle, false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	passPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib {
		t.Errorf("library = %d, want %d: an entry that was not a cast mills nothing", got, lib)
	}
}

// --- Magmatic Hellkite ------------------------------------------------------

func TestSb11MagmaticHellkiteDestroysANonbasicAndItsControllerFetchesAStunnedBasic(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	coffers := seedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	basic := seedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	swamp := b07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	b07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	castCatalogSpell(t, g, "Magmatic Hellkite", "Creature — Dragon", sb11HellkiteOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if hasID(pick.PickTargetCards, basic) {
		t.Error("a basic land is not a legal target")
	}
	pickCard(t, g, me.ID, coffers)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(coffers) || !opp.Graveyard.Contains(coffers) {
		t.Error("the targeted land should be destroyed")
	}
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("the Hellkite's controller does not search")
	}
	if searchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the land's controller gets no search prompt")
	}
	answerSearchByID(t, g, opp.ID, swamp)
	card, ok := battlefieldCard(g, swamp)
	if !ok {
		t.Fatal("the basic land did not reach the battlefield")
	}
	if !card.Tapped {
		t.Error("the basic land enters tapped")
	}
	if card.Counters[game.CounterStun] != 1 {
		t.Errorf("the basic land has %d stun counters, want 1", card.Counters[game.CounterStun])
	}
}

func TestSb11MagmaticHellkiteSearchHappensEvenIfTheLandIsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tough Land", TypeLine: "Land", Keywords: []string{"indestructible"},
		Owner: opp.ID, Controller: opp.ID,
	})
	swamp := b07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	castCatalogSpell(t, g, "Magmatic Hellkite", "Creature — Dragon", sb11HellkiteOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, tough)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible land was destroyed")
	}
	if searchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the land's controller still searches")
	}
	answerSearchByID(t, g, opp.ID, swamp)
	if !g.Battlefield.Contains(swamp) {
		t.Error("the basic land did not enter")
	}
}

// --- Call the Spirit Dragons -------------------------------------------------

func sb11CallTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Call the Spirit Dragons", "Enchantment", sb11CallOracle, false)
	return g, me
}

func sb11Dragon(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return b21Push(g, owner, name, "Creature — Dragon", "", 3, 3, colors...)
}

func TestSb11CallTheSpiritDragonsGivesDragonsIndestructible(t *testing.T) {
	g, me := sb11CallTable(t)
	opp := g.Seats[0]
	mine := sb11Dragon(g, me.ID, "Mine", "R")
	theirs := sb11Dragon(g, opp.ID, "Theirs", "R")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !hasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("a Dragon you control is indestructible")
	}
	if hasEffectiveKeyword(t, g, theirs, "indestructible") || hasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("an opponent's Dragon and a non-Dragon get nothing")
	}
}

func TestSb11CallTheSpiritDragonsWinsWithCountersOnFiveDifferentDragons(t *testing.T) {
	g, me := sb11CallTable(t)
	var dragons []uuid.UUID
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		dragons = append(dragons, sb11Dragon(g, me.ID, "Dragon "+c, c))
	}
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	for _, d := range dragons {
		if countersOn(g, d, game.CounterPlusOne) != 1 {
			t.Error("a Dragon did not get its counter")
		}
	}
	if g.State != game.StateEnded {
		t.Error("counters on five Dragons did not win the game")
	}
}

func TestSb11CallTheSpiritDragonsOneFiveColorDragonGetsFiveCountersAndNoWin(t *testing.T) {
	g, me := sb11CallTable(t)
	d := sb11Dragon(g, me.ID, "Spirit", "W", "U", "B", "R", "G")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, d, game.CounterPlusOne); got != 5 {
		t.Errorf("%d counters, want 5: one for each colour", got)
	}
	if g.State == game.StateEnded {
		t.Error("five counters on one Dragon won the game")
	}
}

func TestSb11CallTheSpiritDragonsAsksWhichOfTwoSameColorDragonsGetsTheCounter(t *testing.T) {
	g, me := sb11CallTable(t)
	a := sb11Dragon(g, me.ID, "Red A", "R")
	b := sb11Dragon(g, me.ID, "Red B", "R")
	green := sb11Dragon(g, me.ID, "Green", "G")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the two red Dragons")
	}
	if countersOn(g, green, game.CounterPlusOne) != 0 {
		t.Error("green comes after red in the walk, its counter waits for the red answer")
	}
	answerChooseCards(t, g, me.ID, b)
	passPriorityAroundTable(t, g)
	if countersOn(g, b, game.CounterPlusOne) != 1 || countersOn(g, a, game.CounterPlusOne) != 0 {
		t.Error("the chosen red Dragon, and only it, got the counter")
	}
	if countersOn(g, green, game.CounterPlusOne) != 1 {
		t.Error("the green Dragon got no counter")
	}
}

// --- Sarkhan, Dragon Ascendant ------------------------------------------------

func sb11CastSarkhan(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Sarkhan, Dragon Ascendant", "Legendary Creature — Human Druid", sb11SarkhanOracle, nil)
	passPriorityAroundTable(t, g)
	return id
}

func TestSb11SarkhanBeholdsADragonYouControlForATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := sb11Dragon(g, me.ID, "Dragon", "R")
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sb11CastSarkhan(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no behold prompt")
	}
	answerChooseCards(t, g, me.ID, dragon)
	if got := len(battlefieldIDsNamed(g, "Treasure")); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

func TestSb11SarkhanBeholdsADragonCardInHandAndRevealsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hand := handSpell(me, "Hand Dragon", "Creature — Dragon", "{4}{R}")
	handSpell(me, "Hand Bear", "Creature — Bear", "{1}{G}")
	sb11CastSarkhan(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no behold prompt")
	}
	answerChooseCards(t, g, me.ID, hand)
	if got := len(battlefieldIDsNamed(g, "Treasure")); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
	revealed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == hand {
			revealed = true
		}
	}
	if !revealed {
		t.Error("the Dragon card from hand was not revealed")
	}
	if !me.Hand.Contains(hand) {
		t.Error("beholding moved the card out of the hand")
	}
}

func TestSb11SarkhanDecliningOrHavingNoDragonMakesNoTreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dragon := sb11Dragon(g, me.ID, "Dragon", "R")
	_ = dragon
	sb11CastSarkhan(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no behold prompt")
	}
	answerChooseCards(t, g, me.ID)
	if got := len(battlefieldIDsNamed(g, "Treasure")); got != 0 {
		t.Errorf("%d Treasures after declining", got)
	}
	g2 := newCatalogGame(t)
	sb11CastSarkhan(t, g2)
	if chooseCardsChoiceFor(g2, g2.Seats[0].ID) != nil {
		t.Error("nothing to behold: no prompt expected")
	}
}

func TestSb11SarkhanGrowsAndBecomesAFlyingDragonWhenADragonEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sarkhan := b12Push(g, me.ID, "Sarkhan, Dragon Ascendant", "Legendary Creature — Human Druid", sb11SarkhanOracle, 2, 2)
	if hasEffectiveKeyword(t, g, sarkhan, "flying") {
		t.Fatal("setup: Sarkhan flies before a Dragon entered")
	}
	castCatalogSpell(t, g, "Spark Dragon", "Creature — Dragon", "", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, sarkhan, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, sarkhan, "flying") {
		t.Error("Sarkhan has no flying")
	}
	if !layeredCard(t, g, sarkhan).HasSubtype("Dragon") {
		t.Error("Sarkhan is not a Dragon")
	}
	endTurn(t, g)
	if hasEffectiveKeyword(t, g, sarkhan, "flying") || layeredCard(t, g, sarkhan).HasSubtype("Dragon") {
		t.Error("the type and flying should end with the turn")
	}
	if got := countersOn(g, sarkhan, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter is permanent: %d", got)
	}
}

// --- Voyager Glidecar ---------------------------------------------------------

func TestSb11VoyagerGlidecarTapsThreeOthersToBecomeAFlyingCreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	car := pushVehicleForTest(g, me.ID, "Voyager Glidecar", sb11GlidecarOracle, 2, 3)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, car, "flying") || layeredCard(t, g, car).IsCreature() {
		t.Fatal("setup: the Vehicle is not a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}}); err == nil {
		t.Fatal("two creatures paid a cost that needs three")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b, c}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !b12Card(t, g, id).Tapped {
			t.Error("a creature was not tapped for the cost")
		}
	}
	if !layeredCard(t, g, car).IsCreature() || !hasEffectiveKeyword(t, g, car, "flying") {
		t.Error("the Vehicle is not a flying creature")
	}
	if p, tt := ptOf(t, g, car); p != 3 || tt != 4 {
		t.Errorf("the Vehicle is %d/%d, want 3/4 with its counter", p, tt)
	}
	endTurn(t, g)
	if layeredCard(t, g, car).IsCreature() || hasEffectiveKeyword(t, g, car, "flying") {
		t.Error("the animation should end with the turn")
	}
	if countersOn(g, car, game.CounterPlusOne) != 1 {
		t.Error("the counter stays")
	}
}

func TestSb11VoyagerGlidecarScriesWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Voyager Glidecar", "Artifact — Vehicle", sb11GlidecarOracle, nil)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 1 prompt")
	}
}

// --- Sabotage Strategist ------------------------------------------------------

func TestSb11SabotageStrategistShrinksOnlyTheCreaturesThatAttackedYou(t *testing.T) {
	g := newCatalogGame(t)
	me, third := g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Sabotage Strategist", "Creature — Vedalken Ranger", sb11StrategistOracle, 2, 2)
	a := b12Creature(g, g.Seats[0].ID, "Attacker A", "Creature — Bear", 3, 3)
	b := b12Creature(g, g.Seats[0].ID, "Attacker B", "Creature — Bear", 3, 3)
	c := b12Creature(g, g.Seats[0].ID, "Attacker C", "Creature — Bear", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, at := range []struct {
		id  uuid.UUID
		def uuid.UUID
	}{{a, me.ID}, {b, me.ID}, {c, third.ID}} {
		if err := g.DeclareAttacker(at.id, at.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, _ := ptOf(t, g, id); p != 2 {
			t.Errorf("a creature that attacked you has power %d, want 2", p)
		}
	}
	if p, _ := ptOf(t, g, c); p != 3 {
		t.Errorf("a creature that attacked another player has power %d, want 3", p)
	}
	endTurn(t, g)
	if p, _ := ptOf(t, g, a); p != 3 {
		t.Errorf("-1/-0 should end with the turn: power %d", p)
	}
}

func TestSb11SabotageStrategistIgnoresAttacksOnOtherPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, third := g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Sabotage Strategist", "Creature — Vedalken Ranger", sb11StrategistOracle, 2, 2)
	a := b12Creature(g, g.Seats[0].ID, "Attacker A", "Creature — Bear", 3, 3)
	declareAttack(t, g, third.ID, a)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, a); p != 3 {
		t.Errorf("an attack on another player shrank the attacker: power %d", p)
	}
}

func TestSb11SabotageStrategistExhaustPutsThreeCountersOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	sb := b12Push(g, me.ID, "Sabotage Strategist", "Creature — Vedalken Ranger", sb11StrategistOracle, 2, 2)
	b06AddMana(me, "C", "C", "C", "C", "C", "U", "U", "C", "C", "C", "C", "C", "U", "U")
	b16Activate(t, g, me.ID, sb, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, sb, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, sb, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("exhaust was activated a second time")
	}
}

// --- Spikeshell Harrier -------------------------------------------------------

// sb11Harrier casts the Harrier and aims it at the opponent's creature, then
// lets it resolve.
func sb11Harrier(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Spikeshell Harrier", "Artifact Creature — Robot Turtle", sb11HarrierOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
}

func TestSb11SpikeshellHarrierBouncesAndReducesTheHighestSpeedByOne(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.SetSpeedForTest(opp.ID, 3)
	g.SetSpeedForTest(g.Seats[2].ID, 2)
	sb11Harrier(t, g, bear)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Error("the creature was not returned to its owner's hand")
	}
	if got := g.SpeedOf(opp.ID); got != 2 {
		t.Errorf("the opponent's speed is %d, want 2", got)
	}
}

func TestSb11SpikeshellHarrierLeavesSpeedAloneWhenAnotherPlayerIsAsFast(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.SetSpeedForTest(opp.ID, 3)
	g.SetSpeedForTest(g.Seats[2].ID, 3)
	sb11Harrier(t, g, bear)
	if got := g.SpeedOf(opp.ID); got != 3 {
		t.Errorf("speed is %d, want 3: it is not greater than each other player's", got)
	}
}

func TestSb11SpikeshellHarrierNeverReducesSpeedBelowOneAndCanTakeAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	car := pushVehicleForTest(g, opp.ID, "Their Vehicle", "", 3, 3)
	g.SetSpeedForTest(opp.ID, 1)
	sb11Harrier(t, g, car)
	if !opp.Hand.Contains(car) {
		t.Error("the Vehicle was not returned")
	}
	if got := g.SpeedOf(opp.ID); got != 1 {
		t.Errorf("speed is %d, want 1: it can't go below 1", got)
	}
}

func TestSb11SpikeshellHarrierFromMaxSpeedEndsMaxSpeed(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.SetSpeedForTest(opp.ID, game.MaxSpeed)
	sb11Harrier(t, g, bear)
	if g.SpeedOf(opp.ID) != game.MaxSpeed-1 {
		t.Errorf("speed is %d, want %d", g.SpeedOf(opp.ID), game.MaxSpeed-1)
	}
}

var (
	_ = uuid.Nil
	_ = game.ZoneHand
)
