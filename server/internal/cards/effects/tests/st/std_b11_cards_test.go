package testsst

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Grim Javelineer ------------------------------------------------------

func TestSb11GrimJavelineerPumpsAnAttackerAndSurveilsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Grim Javelineer", "Creature — Human Warrior", effectstest.Sb11JavelineerOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt: whenever you attack")
	}
	if effectstest.HasID(pick.PickTargetCards, idle) {
		t.Error("a creature that isn't attacking is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 3 {
		t.Errorf("the target has power %d, want 3", p)
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Error("the target died this turn: a surveil 1 prompt")
	}
}

func TestSb11GrimJavelineerSurveilsNothingForAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Grim Javelineer", "Creature — Human Warrior", effectstest.Sb11JavelineerOracle, 3, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear, other)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, me.ID, other)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) != nil {
		t.Error("another creature dying surveils nothing")
	}
}

// --- Qarsi Revenant -------------------------------------------------------

func TestSb11QarsiRevenantRenewPutsThreeKeywordCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushBattlefieldCreature(g, me.ID)
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Qarsi Revenant", "Creature — Vampire",
		effectstest.Sb11RevenantOracle, 3, 3, "{B}{B}{B}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile Qarsi Revenant")
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, kw := range []string{"flying", "deathtouch", "lifelink"} {
		if effectstest.CountersOn(g, target, kw) != 1 {
			t.Errorf("no %s counter on the target", kw)
		}
		if !effectstest.HasEffectiveKeyword(t, g, target, kw) {
			t.Errorf("the target has no %s from its counter", kw)
		}
	}
}

// --- Death Begets Life ----------------------------------------------------

func TestSb11DeathBegetsLifeDrawsForEachPermanentDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, opp.ID, "Pacifism", "Enchantment — Aura")
	art := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryTopForTest(me, "Filler")
	}
	before := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Death Begets Life", "Sorcery", effectstest.Sb11DeathBegetsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, ench} {
		if g.Battlefield.Contains(id) {
			t.Error("a creature or enchantment survived")
		}
	}
	if !g.Battlefield.Contains(art) {
		t.Error("an artifact was destroyed")
	}
	if got := effectstest.HandSize(me) - before; got != 3 {
		t.Errorf("drew %d cards, want 3 (one per permanent destroyed)", got)
	}
}

func TestSb11DeathBegetsLifeDoesNotCountAnIndestructibleCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Golem", Power: 2, Toughness: 2,
		Keywords: []string{"indestructible"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryTopForTest(me, "Filler")
	}
	before := effectstest.HandSize(me)
	effectstest.CastCatalogSpell(t, g, "Death Begets Life", "Sorcery", effectstest.Sb11DeathBegetsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible creature was destroyed")
	}
	if got := effectstest.HandSize(me) - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- Clarion Conqueror ----------------------------------------------------

func TestSb11ClarionConquerorShutsOffArtifactsCreaturesAndPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]
	birds := effectstest.B21Push(g, me.ID, "Birds of Paradise", "Creature — Bird", effectstest.ActBirdsOracle, 0, 1, "G")
	theirBirds := effectstest.B21Push(g, them.ID, "Birds of Paradise", "Creature — Bird", effectstest.ActBirdsOracle, 0, 1, "G")
	solRing := effectstest.B21Push(g, me.ID, "Sol Ring", "Artifact", effectstest.ActSolRingOracle, 0, 0)
	forest := effectstest.B21Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	pw := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := g.ActivateManaAbility(me.ID, solRing, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("setup: Sol Ring before the Conqueror: %v", err)
	}
	effectstest.B21Push(g, me.ID, "Clarion Conqueror", "Creature — Dragon", effectstest.Sb11ConquerorOracle, 3, 3, "W")
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
	id := effectstest.PushDiesCreatureForTest(g, owner, "Karakyk Guardian", effectstest.Sb11GuardianOracle, "Creature — Dragon", 6, 5)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := sb11Guardian(g, me.ID)
	if !effectstest.EffectiveAbilitiesContain(t, g, guardian, "hexproof") || effectstest.CreatureTargetableBy(g, opp.ID, guardian) {
		t.Fatal("before it has dealt damage an opponent can't target it")
	}
	if !effectstest.CreatureTargetableBy(g, me.ID, guardian) {
		t.Fatal("its controller still can")
	}
	// Damage that is fully prevented is not dealt; a zero never marks it.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(guardian, opp.ID, 0); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if !effectstest.EffectiveAbilitiesContain(t, g, guardian, "hexproof") {
		t.Fatal("0 damage is no damage")
	}
	// Noncombat damage counts, unlike Ruric Thar's combat-only gate.
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(guardian, opp.ID, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if effectstest.EffectiveAbilitiesContain(t, g, guardian, "hexproof") || !effectstest.CreatureTargetableBy(g, opp.ID, guardian) {
		t.Error("after dealing noncombat damage it has lost hexproof")
	}
}

func TestSb11KarakykGuardianLosesHexproofInCombatAndGetsItBackAsANewObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	guardian := sb11Guardian(g, me.ID)
	effectstest.AttackWith(t, g, opp.ID, guardian)
	if opp.Life != 40-6 {
		t.Fatalf("opponent life %d, want 34", opp.Life)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, guardian, "hexproof") {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", effectstest.Sb11MaelstromOracle, false)
		effectstest.CastFromHandForTest(t, g, me, "Spark Dragon", "Creature — Dragon", "{R}", "", effectstest.AutoPay)
		if !effectstest.PermanentTapped(g, land) {
			t.Error("the Maelstrom was not tapped for the Dragon")
		}
	})
	t.Run("an Elf is refused", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", effectstest.Sb11MaelstromOracle, false)
		_, err := effectstest.CastFromHandErr(g, me, "Llanowar Elves", "Creature — Elf Druid", "{G}", "", effectstest.AutoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Errorf("an Elf off the restricted mana: %v, want insufficient mana", err)
		}
	})
	t.Run("an Omen spell is paid", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		land := effectstest.PushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", effectstest.Sb11MaelstromOracle, false)
		params := effectstest.AutoPay
		params.Face = 1
		effectstest.CastFromHand(t, g, effectstest.TdmOmenCards["Sagu Wildling"], params)
		if !effectstest.PermanentTapped(g, land) {
			t.Error("the Maelstrom was not tapped for the Omen")
		}
	})
	t.Run("an ordinary instant is refused", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		effectstest.PushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", effectstest.Sb11MaelstromOracle, false)
		_, err := effectstest.CastFromHandErr(g, me, "Shock", "Instant", "{R}", "", effectstest.AutoPay)
		var short *game.InsufficientManaError
		if !errors.As(err, &short) {
			t.Errorf("an ordinary instant off the restricted mana: %v, want insufficient mana", err)
		}
	})
}

func TestSb11MaelstromSacrificesToFetchADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Maelstrom of the Spirit Dragon", "Land", effectstest.Sb11MaelstromOracle, false)
	dragon := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Spark Dragon", TypeLine: "Creature — Dragon"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Dragon was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, dragon)
	if !me.Hand.Contains(dragon) {
		t.Error("the Dragon is not in hand")
	}
}

// --- Lotuslight Dancers -----------------------------------------------------

func sb11DancersSetup(t *testing.T) (g *game.Game, me *game.Player, ids map[string]uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
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
		ids[c.Name] = effectstest.PushLibraryCardForTest(me, c)
	}
	effectstest.CastAndResolveCreature(t, g, "Lotuslight Dancers", "Creature — Zombie Bard", effectstest.Sb11DancersOracle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if effectstest.SearchOptionNamed(g, c, "White") != uuid.Nil {
		t.Error("a white card was offered")
	}
	return g, me, ids
}

func TestSb11LotuslightDancersPutsOneCardPerColorIntoTheGraveyard(t *testing.T) {
	g, me, ids := sb11DancersSetup(t)
	effectstest.AnswerSearchByID(t, g, me.ID, ids["Black A"], ids["Green"], ids["Blue"])
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
	effectstest.AnswerSearchByID(t, g, me.ID, ids["Three"], ids["Black A"], ids["Blue"])
	if !me.Graveyard.Contains(ids["Three"]) || !me.Graveyard.Contains(ids["Black A"]) || !me.Graveyard.Contains(ids["Blue"]) {
		t.Error("the three cards were not all found")
	}
}

func TestSb11LotuslightDancersRefusesTwoCardsForTheSameColorAndAllowsFewer(t *testing.T) {
	g, me, ids := sb11DancersSetup(t)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids["Black A"], ids["Black B"]}); err == nil {
		t.Fatal("two black cards were accepted for one black quality")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids["Green"])
	if !me.Graveyard.Contains(ids["Green"]) || me.Graveyard.Size() != 1 {
		t.Errorf("a single card should be found: graveyard %d", me.Graveyard.Size())
	}
}

// --- United Battlefront -----------------------------------------------------

func TestSb11UnitedBattlefrontPutsUpToTwoCheapNoncreatureNonlandPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	deep := effectstest.PlTop(me, "Deep Card", "Artifact", "{1}")
	art4 := effectstest.PlTop(me, "Four Rock", "Artifact", "{4}")
	inst := effectstest.PlTop(me, "Cheap Instant", "Instant", "{1}")
	bear := effectstest.PlTop(me, "Bear", "Creature — Bear", "{2}")
	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	art1 := effectstest.PlTop(me, "Rock", "Artifact", "{1}")
	ench := effectstest.PlTop(me, "Rite", "Enchantment", "{2}")
	art2 := effectstest.PlTop(me, "Other Rock", "Artifact", "{3}")
	effectstest.CastCatalogSpell(t, g, "United Battlefront", "Sorcery", effectstest.Sb11BattlefrontOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
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
	bottom := effectstest.PlBottomIDs(me, 5)
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
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt(),
		game.Card{Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{3}{R}{R}", Colors: []string{"R"}, Power: 3, Toughness: 3})
	bear = ids[2]
	priest = effectstest.CastCatalogSpell(t, g, "Severance Priest", "Creature — Djinn Cleric", effectstest.Sb11PriestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, ids[1:]) {
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
	effectstest.KillCreature(t, g, me.ID, priest)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.BattlefieldIDsNamed(g, "Spirit")); got != 1 {
		t.Fatalf("%d Spirits, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Spirit" {
			continue
		}
		if c.Controller != opp.ID {
			t.Error("the exiled card's owner creates the token, not the Priest's controller")
		}
		if p, tt := effectstest.PtOf(t, g, c.InstanceID); p != 5 || tt != 5 {
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
	effectstest.KillCreature(t, g, me.ID, priest)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(effectstest.BattlefieldIDsNamed(g, "Spirit")); got != 0 {
		t.Errorf("%d Spirits though nothing was exiled", got)
	}
}

// --- Yathan Roadwatcher -----------------------------------------------------

func TestSb11YathanRoadwatcherMillsFourThenReturnsACheapCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cheap := effectstest.PushGraveyardCardTyped(me, "Cheap Bear", "Creature — Bear")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == cheap {
			me.Graveyard.Cards[i].ManaCost = "{2}"
		}
	}
	dear := effectstest.PushGraveyardCardTyped(me, "Dear Bear", "Creature — Bear")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == dear {
			me.Graveyard.Cards[i].ManaCost = "{4}"
		}
	}
	lib := me.Library.Size()
	effectstest.RfdCast(t, g, "Yathan Roadwatcher", "Creature — Human Scout", effectstest.Sb11RoadwatcherOracle, 3, 3)
	if got := me.Library.Size(); got != lib-4 {
		t.Fatalf("library = %d, want %d (milled four)", got, lib-4)
	}
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the reflexive trigger")
	}
	if effectstest.HasID(pick.PickTargetCards, dear) {
		t.Error("a creature card with mana value 4 is not a legal target")
	}
	effectstest.PickTriggerTarget(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Error("the cheap creature card did not return to the battlefield")
	}
	if g.Battlefield.Contains(dear) {
		t.Error("the mana value 4 creature returned")
	}
}

func TestSb11YathanRoadwatcherDoesNothingWhenItWasNotCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lib := me.Library.Size()
	effectstest.PushCatalogPermanent(g, me.ID, "Yathan Roadwatcher", "Creature — Human Scout", effectstest.Sb11RoadwatcherOracle, false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib {
		t.Errorf("library = %d, want %d: an entry that was not a cast mills nothing", got, lib)
	}
}

// --- Magmatic Hellkite ------------------------------------------------------

func TestSb11MagmaticHellkiteDestroysANonbasicAndItsControllerFetchesAStunnedBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	coffers := effectstest.SeedLandOnBattlefield(g, opp.ID, "Cabal Coffers", "Land")
	basic := effectstest.SeedLandOnBattlefield(g, opp.ID, "Swamp", "Basic Land — Swamp")
	swamp := effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Magmatic Hellkite", "Creature — Dragon", effectstest.Sb11HellkiteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if effectstest.HasID(pick.PickTargetCards, basic) {
		t.Error("a basic land is not a legal target")
	}
	effectstest.PickCard(t, g, me.ID, coffers)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(coffers) || !opp.Graveyard.Contains(coffers) {
		t.Error("the targeted land should be destroyed")
	}
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("the Hellkite's controller does not search")
	}
	if effectstest.SearchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the land's controller gets no search prompt")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, swamp)
	card, ok := effectstest.BattlefieldCard(g, swamp)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tough := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tough Land", TypeLine: "Land", Keywords: []string{"indestructible"},
		Owner: opp.ID, Controller: opp.ID,
	})
	swamp := effectstest.B07LibraryLand(opp, "Swamp", "Basic Land — Swamp")
	effectstest.CastCatalogSpell(t, g, "Magmatic Hellkite", "Creature — Dragon", effectstest.Sb11HellkiteOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, tough)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tough) {
		t.Error("an indestructible land was destroyed")
	}
	if effectstest.SearchChoiceFor(g, opp.ID) == nil {
		t.Fatal("the land's controller still searches")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, swamp)
	if !g.Battlefield.Contains(swamp) {
		t.Error("the basic land did not enter")
	}
}

// --- Call the Spirit Dragons -------------------------------------------------

func sb11CallTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Call the Spirit Dragons", "Enchantment", effectstest.Sb11CallOracle, false)
	return g, me
}

func sb11Dragon(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	return effectstest.B21Push(g, owner, name, "Creature — Dragon", "", 3, 3, colors...)
}

func TestSb11CallTheSpiritDragonsGivesDragonsIndestructible(t *testing.T) {
	g, me := sb11CallTable(t)
	opp := g.Seats[0]
	mine := sb11Dragon(g, me.ID, "Mine", "R")
	theirs := sb11Dragon(g, opp.ID, "Theirs", "R")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, mine, "indestructible") {
		t.Error("a Dragon you control is indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "indestructible") || effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("an opponent's Dragon and a non-Dragon get nothing")
	}
}

func TestSb11CallTheSpiritDragonsWinsWithCountersOnFiveDifferentDragons(t *testing.T) {
	g, me := sb11CallTable(t)
	var dragons []uuid.UUID
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		dragons = append(dragons, sb11Dragon(g, me.ID, "Dragon "+c, c))
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, d := range dragons {
		if effectstest.CountersOn(g, d, game.CounterPlusOne) != 1 {
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
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, d, game.CounterPlusOne); got != 5 {
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
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the two red Dragons")
	}
	if effectstest.CountersOn(g, green, game.CounterPlusOne) != 0 {
		t.Error("green comes after red in the walk, its counter waits for the red answer")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountersOn(g, b, game.CounterPlusOne) != 1 || effectstest.CountersOn(g, a, game.CounterPlusOne) != 0 {
		t.Error("the chosen red Dragon, and only it, got the counter")
	}
	if effectstest.CountersOn(g, green, game.CounterPlusOne) != 1 {
		t.Error("the green Dragon got no counter")
	}
}

// --- Sarkhan, Dragon Ascendant ------------------------------------------------

func sb11CastSarkhan(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Sarkhan, Dragon Ascendant", "Legendary Creature — Human Druid", effectstest.Sb11SarkhanOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSb11SarkhanBeholdsADragonYouControlForATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := sb11Dragon(g, me.ID, "Dragon", "R")
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sb11CastSarkhan(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no behold prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dragon)
	if got := len(effectstest.BattlefieldIDsNamed(g, "Treasure")); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
}

func TestSb11SarkhanBeholdsADragonCardInHandAndRevealsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hand := effectstest.HandSpell(me, "Hand Dragon", "Creature — Dragon", "{4}{R}")
	effectstest.HandSpell(me, "Hand Bear", "Creature — Bear", "{1}{G}")
	sb11CastSarkhan(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no behold prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, hand)
	if got := len(effectstest.BattlefieldIDsNamed(g, "Treasure")); got != 1 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dragon := sb11Dragon(g, me.ID, "Dragon", "R")
	_ = dragon
	sb11CastSarkhan(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no behold prompt")
	}
	effectstest.AnswerChooseCards(t, g, me.ID)
	if got := len(effectstest.BattlefieldIDsNamed(g, "Treasure")); got != 0 {
		t.Errorf("%d Treasures after declining", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	sb11CastSarkhan(t, g2)
	if effectstest.ChooseCardsChoiceFor(g2, g2.Seats[0].ID) != nil {
		t.Error("nothing to behold: no prompt expected")
	}
}

func TestSb11SarkhanGrowsAndBecomesAFlyingDragonWhenADragonEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sarkhan := effectstest.B12Push(g, me.ID, "Sarkhan, Dragon Ascendant", "Legendary Creature — Human Druid", effectstest.Sb11SarkhanOracle, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, sarkhan, "flying") {
		t.Fatal("setup: Sarkhan flies before a Dragon entered")
	}
	effectstest.CastCatalogSpell(t, g, "Spark Dragon", "Creature — Dragon", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, sarkhan, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, sarkhan, "flying") {
		t.Error("Sarkhan has no flying")
	}
	if !effectstest.LayeredCard(t, g, sarkhan).HasSubtype("Dragon") {
		t.Error("Sarkhan is not a Dragon")
	}
	effectstest.EndTurn(t, g)
	if effectstest.HasEffectiveKeyword(t, g, sarkhan, "flying") || effectstest.LayeredCard(t, g, sarkhan).HasSubtype("Dragon") {
		t.Error("the type and flying should end with the turn")
	}
	if got := effectstest.CountersOn(g, sarkhan, game.CounterPlusOne); got != 1 {
		t.Errorf("the counter is permanent: %d", got)
	}
}

// --- Voyager Glidecar ---------------------------------------------------------

func TestSb11VoyagerGlidecarTapsThreeOthersToBecomeAFlyingCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	car := effectstest.PushVehicleForTest(g, me.ID, "Voyager Glidecar", effectstest.Sb11GlidecarOracle, 2, 3)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, car, "flying") || effectstest.LayeredCard(t, g, car).IsCreature() {
		t.Fatal("setup: the Vehicle is not a creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b}}); err == nil {
		t.Fatal("two creatures paid a cost that needs three")
	}
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{a, b, c}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectstest.B12Card(t, g, id).Tapped {
			t.Error("a creature was not tapped for the cost")
		}
	}
	if !effectstest.LayeredCard(t, g, car).IsCreature() || !effectstest.HasEffectiveKeyword(t, g, car, "flying") {
		t.Error("the Vehicle is not a flying creature")
	}
	if p, tt := effectstest.PtOf(t, g, car); p != 3 || tt != 4 {
		t.Errorf("the Vehicle is %d/%d, want 3/4 with its counter", p, tt)
	}
	effectstest.EndTurn(t, g)
	if effectstest.LayeredCard(t, g, car).IsCreature() || effectstest.HasEffectiveKeyword(t, g, car, "flying") {
		t.Error("the animation should end with the turn")
	}
	if effectstest.CountersOn(g, car, game.CounterPlusOne) != 1 {
		t.Error("the counter stays")
	}
}

func TestSb11VoyagerGlidecarScriesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Voyager Glidecar", "Artifact — Vehicle", effectstest.Sb11GlidecarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 1 prompt")
	}
}

// --- Sabotage Strategist ------------------------------------------------------

func TestSb11SabotageStrategistShrinksOnlyTheCreaturesThatAttackedYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, third := g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Sabotage Strategist", "Creature — Vedalken Ranger", effectstest.Sb11StrategistOracle, 2, 2)
	a := effectstest.B12Creature(g, g.Seats[0].ID, "Attacker A", "Creature — Bear", 3, 3)
	b := effectstest.B12Creature(g, g.Seats[0].ID, "Attacker B", "Creature — Bear", 3, 3)
	c := effectstest.B12Creature(g, g.Seats[0].ID, "Attacker C", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, at := range []struct {
		id  uuid.UUID
		def uuid.UUID
	}{{a, me.ID}, {b, me.ID}, {c, third.ID}} {
		if err := g.DeclareAttacker(at.id, at.def); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if p, _ := effectstest.PtOf(t, g, id); p != 2 {
			t.Errorf("a creature that attacked you has power %d, want 2", p)
		}
	}
	if p, _ := effectstest.PtOf(t, g, c); p != 3 {
		t.Errorf("a creature that attacked another player has power %d, want 3", p)
	}
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 3 {
		t.Errorf("-1/-0 should end with the turn: power %d", p)
	}
}

func TestSb11SabotageStrategistIgnoresAttacksOnOtherPlayers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, third := g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Sabotage Strategist", "Creature — Vedalken Ranger", effectstest.Sb11StrategistOracle, 2, 2)
	a := effectstest.B12Creature(g, g.Seats[0].ID, "Attacker A", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, third.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 3 {
		t.Errorf("an attack on another player shrank the attacker: power %d", p)
	}
}

func TestSb11SabotageStrategistExhaustPutsThreeCountersOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sb := effectstest.B12Push(g, me.ID, "Sabotage Strategist", "Creature — Vedalken Ranger", effectstest.Sb11StrategistOracle, 2, 2)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C", "U", "U", "C", "C", "C", "C", "C", "U", "U")
	effectstest.B16Activate(t, g, me.ID, sb, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, sb, game.CounterPlusOne); got != 3 {
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
	effectstest.CastCatalogSpell(t, g, "Spikeshell Harrier", "Artifact Creature — Robot Turtle", effectstest.Sb11HarrierOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSb11SpikeshellHarrierBouncesAndReducesTheHighestSpeedByOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.SetSpeedForTest(opp.ID, 3)
	g.SetSpeedForTest(g.Seats[2].ID, 3)
	sb11Harrier(t, g, bear)
	if got := g.SpeedOf(opp.ID); got != 3 {
		t.Errorf("speed is %d, want 3: it is not greater than each other player's", got)
	}
}

func TestSb11SpikeshellHarrierNeverReducesSpeedBelowOneAndCanTakeAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	car := effectstest.PushVehicleForTest(g, opp.ID, "Their Vehicle", "", 3, 3)
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
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.SetSpeedForTest(opp.ID, game.MaxSpeed)
	sb11Harrier(t, g, bear)
	if g.SpeedOf(opp.ID) != game.MaxSpeed-1 {
		t.Errorf("speed is %d, want %d", g.SpeedOf(opp.ID), game.MaxSpeed-1)
	}
}

// --- Waxen Shapethief ---------------------------------------------------------

func TestSb11WaxenShapethiefOffersOnlyYourArtifactsAndCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 4, 4)
	ring := effectstest.B12Permanent(g, me.ID, "Sol Ring", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Pacifism", "Enchantment — Aura")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.SeedCopyableCreature(g, opp.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Waxen Shapethief", "Creature — Shapeshifter", effectstest.Sb11ShapethiefOracle, nil)
	for i := 0; i < 8 && effectstest.CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := effectstest.CopyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 2 || !effectstest.HasID(pc.CopyOptions, bear) || !effectstest.HasID(pc.CopyOptions, ring) {
		t.Fatalf("CopyOptions = %v, want only my Bear and my Sol Ring", pc.CopyOptions)
	}
}

func TestSb11WaxenShapethiefEntersAsACopyOfACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 4, 4)
	id := effectstest.CastCatalogSpell(t, g, "Waxen Shapethief", "Creature — Shapeshifter", effectstest.Sb11ShapethiefOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, bear)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Grizzly Bears" || got.CurrentPower() != 4 || got.CurrentToughness() != 4 || !got.IsCopy() {
		t.Errorf("entered as %q %d/%d (copy %v), want a 4/4 Grizzly Bears copy", got.Name, got.CurrentPower(), got.CurrentToughness(), got.IsCopy())
	}
}

func TestSb11WaxenShapethiefCopyingAnArtifactStaysANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ring := effectstest.B12Permanent(g, me.ID, "Sol Ring", "Artifact")
	id := effectstest.CastCatalogSpell(t, g, "Waxen Shapethief", "Creature — Shapeshifter", effectstest.Sb11ShapethiefOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, ring)
	got := effectstest.CopyBattlefieldCard(t, g, id)
	if got.Name != "Sol Ring" || !got.IsArtifact() || got.IsCreature() {
		t.Errorf("entered as %q (artifact %v, creature %v), want a Sol Ring copy", got.Name, got.IsArtifact(), got.IsCreature())
	}
}

// --- Demonic Junker -------------------------------------------------------------

func sb11JunkerTable(t *testing.T, mineToo bool) (g *game.Game, junker, mine, theirs uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine = effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs = effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+1)%4].ID, "Their Bear", "Creature — Bear", 2, 2)
	junker = effectstest.CastHoldCreature(t, g, "Demonic Junker", "Artifact — Vehicle", effectstest.Sb11JunkerOracle, 4, 3)
	for i := 0; i < 4 && effectstest.LatestPickTarget(g, me.ID) != nil; i++ {
		pick := effectstest.LatestPickTarget(g, me.ID)
		switch {
		case effectstest.HasID(pick.PickTargetCards, mine) && mineToo:
			effectstest.PickCard(t, g, me.ID, mine)
		case effectstest.HasID(pick.PickTargetCards, theirs):
			effectstest.PickCard(t, g, me.ID, theirs)
		default:
			effectstest.B26DeclinePickTarget(t, g, me.ID)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, junker, mine, theirs
}

func TestSb11DemonicJunkerDestroysOneCreaturePerPlayerAndGrowsWhenYoursDied(t *testing.T) {
	g, junker, mine, theirs := sb11JunkerTable(t, true)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("a chosen creature survived")
	}
	if got := effectstest.CountersOn(g, junker, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2: a creature you controlled was destroyed", got)
	}
}

func TestSb11DemonicJunkerGetsNoCountersWhenNoneOfYoursDied(t *testing.T) {
	g, junker, mine, theirs := sb11JunkerTable(t, false)
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature was destroyed though not chosen")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("their creature survived")
	}
	if got := effectstest.CountersOn(g, junker, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters, want 0", got)
	}
}

func TestSb11DemonicJunkerAffinityForArtifactsCutsTheCost(t *testing.T) {
	cast := func(artifacts int) error {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		for i := 0; i < artifacts; i++ {
			effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Demonic Junker", TypeLine: "Artifact — Vehicle", ManaCost: "{6}{B}",
			OracleID: effectstest.Sb11JunkerOracle, Power: 4, Toughness: 3, Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
		effectstest.FloatForTest(g, me, "CCB")
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast(0); err == nil {
		t.Error("{2}{B} paid for a {6}{B} spell with no artifacts")
	}
	if err := cast(4); err != nil {
		t.Errorf("with four artifacts the spell costs {2}{B}: %v", err)
	}
}

// --- Taigam, Master Opportunist -------------------------------------------------

func TestSb11TaigamCopiesTheSecondSpellAndExilesItWithFourTimeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Taigam, Master Opportunist", "Legendary Creature — Human Monk", effectstest.Sb11TaigamOracle, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	second := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-6 {
		t.Errorf("opponent life %d, want 34: the first Bolt and the copy resolve, the exiled original does not", opp.Life)
	}
	var c game.Card
	for _, e := range g.Exile.Cards {
		if e.InstanceID == second {
			c = e
		}
	}
	if c.InstanceID == uuid.Nil {
		t.Fatal("the second spell was not exiled")
	}
	if c.Counters[game.CounterTime] != 4 {
		t.Errorf("%d time counters, want 4", c.Counters[game.CounterTime])
	}
	if !game.CardHasSuspend(c) {
		t.Error("the exiled spell did not gain suspend")
	}
	if me.Graveyard.Contains(second) {
		t.Error("the original went to the graveyard")
	}
}

func TestSb11TaigamIgnoresTheThirdSpellAndTheFirst(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Taigam, Master Opportunist", "Legendary Creature — Human Monk", effectstest.Sb11TaigamOracle, 2, 2)
	for i := 0; i < 2; i++ {
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		if i == 0 && len(g.Exile.Cards) != 0 {
			t.Fatal("the first spell was exiled")
		}
	}
	third := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Exile.Contains(third) {
		t.Error("the third spell was exiled")
	}
	if opp.Life != 40-9 {
		t.Errorf("opponent life %d, want 31: three Bolts and one copy would be 12", opp.Life)
	}
}

// --- Flamehold Grappler ----------------------------------------------------------

func TestSb11FlameholdGrapplerCopiesTheNextSpellYouCastThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Flamehold Grappler", "Creature — Human Monk", effectstest.Sb11GrapplerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	pc := effectstest.CopyPromptFor(g, me.ID)
	if pc == nil {
		t.Fatal("no new-targets prompt for the copy")
	}
	if err := g.ResolvePickTarget(pc.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("answer the new-targets prompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-6 {
		t.Errorf("opponent life %d, want 34: the Bolt and its copy", opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-9 {
		t.Errorf("only the NEXT spell is copied: opponent life %d, want 31", opp.Life)
	}
}

func TestSb11FlameholdGrapplerCopyEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Flamehold Grappler", "Creature — Human Monk", effectstest.Sb11GrapplerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToMainOf(t, g, 0)
	_ = me
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-3 {
		t.Errorf("a spell on a later turn was copied: opponent life %d, want 37", opp.Life)
	}
}

var (
	_ = uuid.Nil
	_ = game.ZoneHand
)
