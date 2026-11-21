package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b129_test.go — GAP batch 129 (S82 #5948), The Lord of the Rings:
// Tales of Middle-earth: Éowyn, Lady of Rohan; March from the Black Gate;
// Morgul-Knife Wound; Oath of the Grey Host; Fall of Gil-galad; Book of
// Mazarbul; The Bath Song; Lost Isle Calling; Storm of Saruman.

const (
	gapEowynOracle         = "1ee3753c-3b6e-4182-9305-2ab757f485f0"
	gapMarchOracle         = "9259e918-ace5-4f13-9035-d8bc8c820d25"
	gapMorgulOracle        = "f48b2349-797e-4e79-aea7-2413e78129e1"
	gapOathOracle          = "228fe7f0-54ad-411b-bedb-65a1bc2a099b"
	gapGilGaladOracle      = "c248537f-e28f-47cb-9e47-a6eb05e1af63"
	gapBookMazarbulOracle  = "4598c221-b111-4bde-b4dc-673186c4efe6"
	gapBathSongOracle      = "bc475376-54eb-4846-9cfc-5fa2e7908905"
	gapLostIsleOracle      = "0efbf29c-09a0-4472-83ad-8d692c0232a5"
	gapStormSarumanOracle  = "cf5f4860-e805-46a3-9352-a2c583e33403"
	gapSagaTypeLine        = "Enchantment — Saga"
	gapAuraTypeLine        = "Enchantment — Aura"
	gapEowynTypeLine       = "Legendary Creature — Human Noble"
	gapBonesplitterEquipID = effectstest.BonesplitterOracle
)

func TestGapB129CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapEowynOracle:        "Éowyn, Lady of Rohan",
		gapMarchOracle:        "March from the Black Gate",
		gapMorgulOracle:       "Morgul-Knife Wound",
		gapOathOracle:         "Oath of the Grey Host",
		gapGilGaladOracle:     "Fall of Gil-galad",
		gapBookMazarbulOracle: "Book of Mazarbul",
		gapBathSongOracle:     "The Bath Song",
		gapLostIsleOracle:     "Lost Isle Calling",
		gapStormSarumanOracle: "Storm of Saruman",
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

// gapAttach attaches an Equipment or Aura already on the battlefield to a
// creature.
func gapAttach(g *game.Game, attachment, host uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == attachment {
				g.Battlefield.Cards[i].AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
			}
		}
	})
}

// gapLibraryCards stacks n plain cards on top of a library.
func gapLibraryCards(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.R38LibraryTop(p, "Filler", "Creature — Test")
	}
}

// --- Éowyn, Lady of Rohan ---------------------------------------------------

func TestGapEowynGivesYourChoiceOfFirstStrikeOrVigilance(t *testing.T) {
	for index, want := range []string{"first strike", "vigilance"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Éowyn, Lady of Rohan", gapEowynTypeLine, gapEowynOracle, 2, 4)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, me.ID, index)
		effectstest.PassPriorityAroundTable(t, g)
		other := "vigilance"
		if index == 1 {
			other = "first strike"
		}
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, want) {
			t.Errorf("choice %d: the bear lacks %s", index, want)
		}
		if effectstest.HasKeywordOnBattlefield(t, g, bear, other) {
			t.Errorf("choice %d: the bear has %s too", index, other)
		}
	}
}

func TestGapEowynGivesBothToAnEquippedCreatureWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Éowyn, Lady of Rohan", gapEowynTypeLine, gapEowynOracle, 2, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	sword := effectstest.SeedEquipment(g, opp.ID, "Bonesplitter", effectstest.BonesplitterOracle)
	gapAttach(g, sword, theirs)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Fatal("an equipped creature still asked for a choice")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, theirs, "first strike") || !effectstest.HasKeywordOnBattlefield(t, g, theirs, "vigilance") {
		t.Error("the equipped creature did not get both keywords")
	}
}

func TestGapEowynDoesNotTriggerOnAnotherPlayersCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	e := effectstest.B12Push(g, me.ID, "Éowyn, Lady of Rohan", gapEowynTypeLine, gapEowynOracle, 2, 4)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.Turn.ActiveSeat = 1
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if effectstest.TriggerOnStack(g, e) != nil {
		t.Error("the trigger went on the stack on an opponent's turn")
	}
}

func TestGapEowynMakesEquipAbilitiesCostOneLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.SeedBear(g, me.ID)
	sword := effectstest.SeedEquipment(g, me.ID, "Bonesplitter", effectstest.BonesplitterOracle)
	if got := effectstest.EquipPrice(t, g, me.ID, sword, bear); got != "{1}" {
		t.Fatalf("equip before Éowyn: %s, want the printed {1}", got)
	}
	effectstest.B12Push(g, me.ID, "Éowyn, Lady of Rohan", gapEowynTypeLine, gapEowynOracle, 2, 4)
	if got := effectstest.EquipPrice(t, g, me.ID, sword, bear); got != "{0}" {
		t.Errorf("equip with Éowyn: %s, want {0}", got)
	}
}

// --- March from the Black Gate ----------------------------------------------

func TestGapMarchAmassesWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "March from the Black Gate", "Enchantment", gapMarchOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 || !army.HasSubtype("Orc") {
		t.Fatalf("after entering: %d counters, Orc %v; want 1 and an Orc", got, army.HasSubtype("Orc"))
	}
}

func TestGapMarchAmassesOncePerAttackingArmy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "March from the Black Gate", "Enchantment", gapMarchOracle, 0, 0)
	army := effectstest.B12Creature(g, me.ID, "Orc Army", "Token Creature — Orc Army", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, army, game.CounterPlusOne); got != 0 {
		t.Fatalf("a non-Army attacked and the Army got %d counters", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	opp2 := g2.Seats[(g2.Turn.ActiveSeat+1)%len(g2.Seats)]
	effectstest.B12Push(g2, me2.ID, "March from the Black Gate", "Enchantment", gapMarchOracle, 0, 0)
	army2 := effectstest.B12Creature(g2, me2.ID, "Orc Army", "Token Creature — Orc Army", 1, 1)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(army2, opp2.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if got := effectstest.CountersOn(g2, army2, game.CounterPlusOne); got != 1 {
		t.Errorf("the Army attacked: %d counters, want 1", got)
	}
}

// --- Morgul-Knife Wound -----------------------------------------------------

func gapMorgulOn(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Morgul-Knife Wound", gapAuraTypeLine, gapMorgulOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 1 {
		t.Fatalf("enchanted power %d, want 4-3=1", got)
	}
	return bear
}

func TestGapMorgulPayTwoLifeKeepsTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := gapMorgulOn(t, g, me)
	life := me.Life
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, bear) != game.ZoneBattlefield {
		t.Error("the creature left although its controller paid")
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want 2 paid", life, me.Life)
	}
}

func TestGapMorgulDecliningExilesTheCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := gapMorgulOn(t, g, me)
	life := me.Life
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, bear) != game.ZoneExile {
		t.Errorf("the creature is in %s, want exile", effectstest.B12ZoneOf(g, bear))
	}
	if me.Life != life {
		t.Errorf("life changed %d → %d without paying", life, me.Life)
	}
}

func TestGapMorgulExilesWithoutAskingWhenYouCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := gapMorgulOn(t, g, me)
	me.Life = 1
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Fatal("asked to pay 2 life at 1 life")
	}
	if effectstest.B12ZoneOf(g, bear) != game.ZoneExile {
		t.Errorf("the creature is in %s, want exile", effectstest.B12ZoneOf(g, bear))
	}
}

// --- Oath of the Grey Host --------------------------------------------------

func TestGapOathPlaysItsThreeChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	other := g.Seats[(seat+2)%len(g.Seats)]
	saga := effectstest.CastCatalogSpell(t, g, "Oath of the Grey Host", gapSagaTypeLine, gapOathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	foodOf := func(p *game.Player) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Food" && c.Controller == p.ID {
				n++
			}
		}
		return n
	}
	if foodOf(me) != 1 || foodOf(opp) != 1 || foodOf(other) != 0 {
		t.Fatalf("Food after I: me %d, target %d, other %d; want 1, 1, 0", foodOf(me), foodOf(opp), foodOf(other))
	}
	lifes := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lifes[p.ID] = p.Life
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lifes[p.ID]
		if p.ID != me.ID {
			want -= 3
		}
		if p.Life != want {
			t.Errorf("after II %s life %d, want %d", p.Name, p.Life, want)
		}
	}
	if effectstest.CountBattlefieldByName(g, "Treasure") != 1 {
		t.Errorf("Treasures after II = %d, want 1", effectstest.CountBattlefieldByName(g, "Treasure"))
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.IsToken() {
			spirits++
			if !c.Tapped || c.Power != 1 || c.Toughness != 1 || !game.HasKeyword(&c, "flying") {
				t.Errorf("Spirit = tapped %v, %d/%d", c.Tapped, c.Power, c.Toughness)
			}
		}
	}
	if spirits != 3 {
		t.Errorf("Spirits after III = %d, want 3", spirits)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after III")
	}
}

// --- Fall of Gil-galad ------------------------------------------------------

func TestGapGilGaladPlaysItsChapters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 5)
	gapLibraryCards(me, 6)
	effectstest.CastCatalogSpell(t, g, "Fall of Gil-galad", gapSagaTypeLine, gapGilGaladOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Fatalf("after II: %d counters, want 2", got)
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	// Mine is 4/4 with two counters and fights a 1/5: it takes 1 and deals 4.
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneBattlefield {
		t.Fatalf("their 1/5 should survive 4 damage")
	}
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two cards from the granted dies trigger", hand, me.Hand.Size())
	}
}

func TestGapGilGaladChapterThreeWithNoSecondTargetStillGrants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	gapLibraryCards(me, 8)
	effectstest.CastCatalogSpell(t, g, "Fall of Gil-galad", gapSagaTypeLine, gapGilGaladOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, mine)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the second target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two cards", hand, me.Hand.Size())
	}
}

// --- Book of Mazarbul -------------------------------------------------------

func TestGapBookOfMazarbulAmassesThenPumpsEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	opp := g.Seats[(seat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	saga := effectstest.CastCatalogSpell(t, g, "Book of Mazarbul", gapSagaTypeLine, gapBookMazarbulOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
		t.Fatalf("after I: %d counters, want 1", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, effectstest.ArmyOf(t, g, me.ID).InstanceID, game.CounterPlusOne); got != 3 {
		t.Fatalf("after II: %d counters, want 3 (amass 2 onto the same Army)", got)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("bear power %d, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "menace") {
		t.Error("the bear lacks menace")
	}
	if got := effectstest.EffectivePower(t, g, theirs); got != 2 || effectstest.HasKeywordOnBattlefield(t, g, theirs, "menace") {
		t.Errorf("an opponent's creature was affected (power %d)", got)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after III")
	}
}

// --- The Bath Song ----------------------------------------------------------

func TestGapBathSongLootsTwiceThenShufflesAndAddsMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	gapLibraryCards(me, 12)
	gy := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: gy, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "The Bath Song", gapSagaTypeLine, gapBathSongOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1 {
		t.Fatalf("chapter I: hand %d before the discard, %d after, want one fewer", hand, me.Hand.Size())
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PickCard(t, g, me.ID, gy)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, gy) != game.ZoneLibrary {
		t.Errorf("the targeted card is in %s, want the library", effectstest.B12ZoneOf(g, gy))
	}
	if effectstest.PoolCount(me, "U") != 2 {
		t.Errorf("blue mana in the pool = %d, want 2", effectstest.PoolCount(me, "U"))
	}
}

// --- Lost Isle Calling ------------------------------------------------------

func gapScry(g *game.Game, p *game.Player, source uuid.UUID) {
	g.WithWriteLock(func() { g.ScryForEffect(p.ID, source, 1) })
}

func TestGapLostIsleCallingCountsScriesAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapLibraryCards(me, 12)
	isle := effectstest.B12Push(g, me.ID, "Lost Isle Calling", "Enchantment", gapLostIsleOracle, 0, 0)
	for i := 0; i < 3; i++ {
		gapScry(g, me, isle)
		effectstest.AnswerScryKeepAll(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.B12Counter(t, g, isle, "verse"); got != 3 {
		t.Fatalf("verse counters = %d, want 3", got)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "U", "U", "U", "U", "U")
	hand := me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, isle, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want 3 cards", hand, me.Hand.Size())
	}
	if effectstest.B12ZoneOf(g, isle) != game.ZoneExile {
		t.Errorf("the enchantment is in %s, want exile", effectstest.B12ZoneOf(g, isle))
	}
}

func TestGapLostIsleCallingExtraTurnNeedsSevenCounters(t *testing.T) {
	for _, n := range []int{6, 7} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		gapLibraryCards(me, 12)
		isle := effectstest.B12Push(g, me.ID, "Lost Isle Calling", "Enchantment", gapLostIsleOracle, 0, 0)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(isle, "verse", n) })
		effectstest.AdvanceToMain(t, g)
		effectstest.B06AddMana(me, "U", "U", "U", "U", "U", "U")
		effectstest.B16Activate(t, g, me.ID, isle, 0, game.ActivateAbilityParams{})
		effectstest.PassPriorityAroundTable(t, g)
		extra := len(g.ExtraTurns)
		if (n >= 7) != (extra > 0) {
			t.Errorf("%d counters: %d extra turns pending", n, extra)
		}
	}
}

func TestGapLostIsleCallingRefusesInstantTiming(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	isle := effectstest.B12Push(g, me.ID, "Lost Isle Calling", "Enchantment", gapLostIsleOracle, 0, 0)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.B06AddMana(me, "U", "U", "U", "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, isle, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the ability was activated at instant speed")
	}
}

// --- Storm of Saruman -------------------------------------------------------

func TestGapStormOfSarumanCopiesOnlyTheSecondSpellEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Storm of Saruman", "Enchantment", gapStormSarumanOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != life-3 {
		t.Fatalf("the first spell: life %d → %d, want a single bolt", life, opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-9 {
		t.Errorf("after the second spell: life %d, want %d (bolt and its copy)", opp.Life, life-9)
	}
}
