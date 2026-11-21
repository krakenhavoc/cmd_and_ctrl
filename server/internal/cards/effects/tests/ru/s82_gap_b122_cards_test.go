package testsru

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b122_cards_test.go — The Lord of the Rings: Tales of Middle-earth
// artifacts and creatures (S82 GAP batch 122, #5948). Every test plays the
// real catalog entry through the engine.

const (
	gapB122BarrowBladeOracle     = "867c5592-50d5-420c-ac71-50014bc390eb"
	gapB122MirrorOracle          = "7695ef41-fa58-4890-b3fe-de5751831d8e"
	gapB122GrondOracle           = "ea52d9e8-5c25-429a-8de4-feb4558bdb99"
	gapB122GaladhrimBowOracle    = "92225b59-4f39-4175-ad0d-88e96f4d906a"
	gapB122DoorsOracle           = "d82c77a1-2b81-415d-9f18-47e059c15e01"
	gapB122GimlisAxeOracle       = "773b40b1-72f0-48a0-b210-2afa5df5485e"
	gapB122ScarecrowOracle       = "4b9f6a6e-c168-4732-b454-2272d56636da"
	gapB122CirithUngolOracle     = "c9610d33-4cef-4171-bbf4-48d720492d55"
	gapB122CrebainOracle         = "9dd755a5-0bef-4bde-92fa-14ef2a1a1410"
	gapB122EasterlingOracle      = "149887fa-6b2d-495e-b3e0-b2ac7d3dc838"
	gapB122EquipmentType         = "Artifact — Equipment"
	gapB122LegendaryCreatureType = "Legendary Creature — Elf Noble"
)

func TestGapB122EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB122BarrowBladeOracle:  "Barrow-Blade",
		gapB122MirrorOracle:       "Mirror of Galadriel",
		gapB122GrondOracle:        "Grond, the Gatebreaker",
		gapB122GaladhrimBowOracle: "Galadhrim Bow",
		gapB122DoorsOracle:        "Doors of Durin",
		gapB122GimlisAxeOracle:    "Gimli's Axe",
		gapB122ScarecrowOracle:    "Shire Scarecrow",
		gapB122CirithUngolOracle:  "Cirith Ungol Patrol",
		gapB122CrebainOracle:      "Dunland Crebain",
		gapB122EasterlingOracle:   "Easterling Vanguard",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func gapB122Equip(t *testing.T, g *game.Game, me *game.Player, name, oracle string, host uuid.UUID, mana string) uuid.UUID {
	t.Helper()
	eq := effectstest.PushCatalogPermanent(g, me.ID, name, gapB122EquipmentType, oracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.StrictEquip(t, g, me.ID, eq, host, mana); err != nil {
		t.Fatalf("equip %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, eq, host) {
		t.Fatalf("%s did not attach", name)
	}
	return eq
}

func TestGapB122BarrowBladePumpsAndStripsTheCreatureItBlocksOrIsBlockedBy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	holder := effectstest.PushVanillaCreature(g, me.ID, "Holder", 2, 2)
	gapB122Equip(t, g, me, "Barrow-Blade", gapB122BarrowBladeOracle, holder, "{G}")
	if c := effectstest.LayeredCard(t, g, holder); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("equipped creature is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Winged Blocker", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.DeclareAttack(t, g, opp.ID, holder)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if !effectstest.EffectiveAbilitiesContain(t, g, flyer, "flying") {
		t.Fatal("the blocker lacked flying before it blocked")
	}
	// Flying creatures can block a ground attacker.
	if err := g.DeclareBlocker(flyer, holder); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, flyer, "flying") {
		t.Error("the blocking creature kept its abilities")
	}
	effectstest.EndTurn(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, flyer, "flying") {
		t.Error("the blocker did not get its abilities back at end of turn")
	}
}

func TestGapB122BarrowBladeStripsTheBlockerOfAnEquippedBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	// The equipped creature is the one that blocks.
	eq := effectstest.PushCatalogPermanent(g, opp.ID, "Barrow-Blade", gapB122EquipmentType, gapB122BarrowBladeOracle, false)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 1, 3)
	g.WithWriteLock(func() { _ = g.AttachForEffect(eq, game.TargetRef{Kind: game.TargetCard, ID: blocker}) })
	attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Trampler", TypeLine: "Creature — Beast",
		Power: 2, Toughness: 2, Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID})
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, attacker, "trample") {
		t.Error("the creature blocked by the equipped creature kept its abilities")
	}
}

func TestGapB122GimlisAxePumpsAndGivesMenaceOnlyToALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	hero := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Hero", TypeLine: gapB122LegendaryCreatureType,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	axe := gapB122Equip(t, g, me, "Gimli's Axe", gapB122GimlisAxeOracle, plain, "{G}{G}")
	if c := effectstest.LayeredCard(t, g, plain); c.CurrentPower() != 5 || c.CurrentToughness() != 2 || game.HasKeyword(&c, "menace") {
		t.Errorf("non-legendary host is %d/%d menace=%v, want 5/2 with no menace", c.Power, c.Toughness, game.HasKeyword(&c, "menace"))
	}
	if err := effectstest.StrictEquip(t, g, me.ID, axe, hero, "{G}{G}"); err != nil {
		t.Fatalf("re-equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, hero); c.CurrentPower() != 5 || !game.HasKeyword(&c, "menace") {
		t.Errorf("legendary host is %d/%d menace=%v, want 5 power and menace", c.Power, c.Toughness, game.HasKeyword(&c, "menace"))
	}
}

func TestGapB122GaladhrimBowHasFlashAttachesAndUntaps(t *testing.T) {
	spec, ok := effects.Lookup(gapB122GaladhrimBowOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Fatalf("Galadhrim Bow keywords: %+v", spec.PrintedKeywords)
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	bow := effectstest.TdCastFromHand(t, g, game.Card{Name: "Galadhrim Bow", TypeLine: gapB122EquipmentType, OracleID: gapB122GaladhrimBowOracle, ManaCost: "{2}{G}"},
		game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	answerPendingTargetIfAny(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, bow, bear) {
		t.Fatal("the Bow did not attach to the creature")
	}
	if effectstest.IsTapped(g, bear) {
		t.Error("the creature was not untapped")
	}
	c := effectstest.LayeredCard(t, g, bear)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "reach") {
		t.Errorf("equipped creature is %d/%d reach=%v, want 3/4 with reach", c.Power, c.Toughness, game.HasKeyword(&c, "reach"))
	}
}

// answerPendingTargetIfAny picks `pick` for an open target prompt of `who`,
// if the engine asked (it does not when only one target is legal).
func answerPendingTargetIfAny(t *testing.T, g *game.Game, who, pick uuid.UUID) {
	t.Helper()
	if p := effectstest.LatestPickTarget(g, who); p != nil {
		if err := g.ResolvePickTarget(p.ID, who, game.TargetRef{Kind: game.TargetCard, ID: pick}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}

func TestGapB122MirrorOfGaladrielCostsOneLessPerLegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mirror := effectstest.PushCatalogPermanent(g, me.ID, "Mirror of Galadriel", "Legendary Artifact", gapB122MirrorOracle, false)
	effectstest.AdvanceToMain(t, g)
	activate := func(mana string) error {
		if mana != "" {
			if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
				t.Fatalf("AddMana: %v", err)
			}
		}
		return g.ActivateCatalogAbility(me.ID, mirror, 0, game.ActivateAbilityParams{Strict: true})
	}
	if err := activate("{G}{G}{G}{G}"); err == nil {
		t.Fatal("activated for {4} with no legendary creature")
	}
	g.WithWriteLock(func() { me.ManaPool = game.ManaPool{} })
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: fmt.Sprintf("Legend %d", i), TypeLine: gapB122LegendaryCreatureType,
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	if err := activate("{G}{G}{G}"); err != nil {
		t.Fatalf("activate for {3} with two legendary creatures: %v", err)
	}
	// Scry 1, then draw a card.
	effectstest.StockLibraryFor(g, me, 5)
	before := me.Hand.Size()
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand grew by %d, want 1 (the draw after the scry)", me.Hand.Size()-before)
	}
}

func TestGapB122ShireScarecrowAddsOneManaOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	scarecrow := effectstest.PushCatalogPermanent(g, me.ID, "Shire Scarecrow", "Artifact Creature — Scarecrow", gapB122ScarecrowOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatalf("AddMana: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, scarecrow, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceMana); c != nil {
		effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 2 {
		t.Errorf("pool = %d after paying {1} and adding one, want 2", got)
	}
	if err := g.ActivateManaAbility(me.ID, scarecrow, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the second activation in one turn was allowed")
	}
}

func TestGapB122GrondIsACreatureOnlyOnYourTurnWithAnArmy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	grond := effectstest.PushCatalogPermanent(g, me.ID, "Grond, the Gatebreaker", "Legendary Artifact — Vehicle", gapB122GrondOracle, false)
	isCreature := func() bool { c := effectstest.LayeredCard(t, g, grond); return c.IsCreature() }
	if isCreature() {
		t.Fatal("Grond is a creature with no Army")
	}
	army := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Orc Army", TypeLine: "Token Creature — Orc Army",
		Power: 0, Toughness: 0, Owner: opp.ID, Controller: opp.ID})
	if isCreature() {
		t.Error("Grond became a creature for an opponent's Army")
	}
	g.WithWriteLock(func() { _ = g.GainControlForEffect(uuid.Nil, army, me.ID, game.IndefiniteDuration(), "test") })
	if !isCreature() {
		t.Error("Grond is not a creature on my turn while I control an Army")
	}
	spec, _ := effects.Lookup(gapB122GrondOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "trample" {
		t.Errorf("keywords = %v, want trample", spec.PrintedKeywords)
	}
	effectstest.EndTurn(t, g)
	if isCreature() {
		t.Error("Grond is a creature on an opponent's turn")
	}
}

func TestGapB122GrondCrewsWithThreePower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grond := effectstest.PushCatalogPermanent(g, me.ID, "Grond, the Gatebreaker", "Legendary Artifact — Vehicle", gapB122GrondOracle, false)
	crew := effectstest.PushVanillaCreature(g, me.ID, "Crew", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, grond, 0, game.ActivateAbilityParams{Strict: true, CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, grond); !c.IsCreature() {
		t.Error("crewing did not make Grond a creature")
	}
}

func TestGapB122CirithUngolPatrolSacrificesForACardAndFood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	patrol := effectstest.PushCatalogPermanent(g, me.ID, "Cirith Ungol Patrol", "Creature — Orc Soldier", gapB122CirithUngolOracle, false)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("AddMana: %v", err)
	}
	// "Another": the Patrol can't pay with itself.
	if err := g.ActivateCatalogAbility(me.ID, patrol, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{patrol}}); err == nil {
		t.Fatal("the Patrol sacrificed itself")
	}
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, patrol, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, fodder) {
		t.Error("the creature was not sacrificed")
	}
	if me.Hand.Size() != before+1 {
		t.Errorf("hand grew by %d, want 1", me.Hand.Size()-before)
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Food")); n != 1 {
		t.Errorf("%d Food tokens, want 1", n)
	}
}

func gapB122Army(g *game.Game, owner uuid.UUID) (uuid.UUID, int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype("Army") {
			return c.InstanceID, effectstest.CountersOn(g, c.InstanceID, game.CounterPlusOne)
		}
	}
	return uuid.Nil, 0
}

func TestGapB122DunlandCrebainAmassesOrcsTwoWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Dunland Crebain", TypeLine: "Creature — Bird Horror", OracleID: gapB122CrebainOracle,
		ManaCost: "{2}{B}", Power: 1, Toughness: 1}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	id, n := gapB122Army(g, me.ID)
	if id == uuid.Nil || n != 2 {
		t.Fatalf("Army %v with %d counters, want an Orc Army with 2", id, n)
	}
	if c, _ := effectstest.BattlefieldCard(g, id); !c.HasSubtype("Orc") {
		t.Error("the Army is not an Orc")
	}
	spec, _ := effects.Lookup(gapB122CrebainOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flying" {
		t.Errorf("keywords = %v, want flying", spec.PrintedKeywords)
	}
}

func TestGapB122EasterlingVanguardAmassesOrcsOneWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vanguard := effectstest.PushCatalogPermanent(g, me.ID, "Easterling Vanguard", "Creature — Human Warrior", gapB122EasterlingOracle, false)
	if id, _ := gapB122Army(g, me.ID); id != uuid.Nil {
		t.Fatal("an Army existed before the Vanguard died")
	}
	effectstest.P1Destroy(t, g, vanguard)
	effectstest.PassPriorityAroundTable(t, g)
	if id, n := gapB122Army(g, me.ID); id == uuid.Nil || n != 1 {
		t.Errorf("Army %v with %d counters, want an Orc Army with 1", id, n)
	}
}

// gapB122DoorsTable is a table with Doors of Durin, an attacker and a library
// whose top card is `top`.
func gapB122DoorsTable(t *testing.T, top game.Card) (g *game.Game, me *game.Player, attacker uuid.UUID, topID uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Doors of Durin", "Legendary Artifact", gapB122DoorsOracle, false)
	attacker = effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	topID = uuid.New()
	top.InstanceID, top.Owner, top.Controller = topID, me.ID, me.ID
	g.WithWriteLock(func() { me.Library.PushTop(top) })
	return g, me, attacker, topID
}

// gapB122AnswerDoors runs the trigger: keep the scry, accept the reveal and
// attack the first opponent if asked what the creature attacks.
func gapB122AnswerDoors(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	for i := 0; i < 8; i++ {
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); c != nil {
			if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			continue
		}
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick); c != nil {
			if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
				t.Fatalf("ResolveOptionPick: %v", err)
			}
			continue
		}
		break
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestGapB122DoorsOfDurinPutsARevealedCreatureOntoTheBattlefieldTappedAndAttacking(t *testing.T) {
	g, me, attacker, top := gapB122DoorsTable(t, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}{G}"})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gimli", TypeLine: "Legendary Creature — Dwarf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Legolas", TypeLine: "Legendary Creature — Elf",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, attacker)
	gapB122AnswerDoors(t, g, me)
	c, ok := effectstest.BattlefieldCard(g, top)
	if !ok {
		t.Fatal("the revealed creature card was not put onto the battlefield")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("entered tapped=%v attacking=%v, want tapped and attacking", c.Tapped, c.AttackingTarget != uuid.Nil)
	}
	lc := effectstest.LayeredCard(t, g, top)
	if !game.HasKeyword(&lc, "trample") || !game.HasKeyword(&lc, "hexproof") {
		t.Errorf("trample=%v hexproof=%v, want both with a Dwarf and an Elf", game.HasKeyword(&lc, "trample"), game.HasKeyword(&lc, "hexproof"))
	}
}

func TestGapB122DoorsOfDurinGivesNoKeywordsWithoutADwarfOrElf(t *testing.T) {
	g, me, attacker, top := gapB122DoorsTable(t, game.Card{Name: "Top Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}{G}"})
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, attacker)
	gapB122AnswerDoors(t, g, me)
	if _, ok := effectstest.BattlefieldCard(g, top); !ok {
		t.Fatal("the revealed creature card was not put onto the battlefield")
	}
	lc := effectstest.LayeredCard(t, g, top)
	if game.HasKeyword(&lc, "trample") || game.HasKeyword(&lc, "hexproof") {
		t.Error("the creature gained a keyword with no Dwarf or Elf")
	}
}

func TestGapB122DoorsOfDurinLeavesANonCreatureOnTop(t *testing.T) {
	g, me, attacker, top := gapB122DoorsTable(t, game.Card{Name: "Top Island", TypeLine: "Basic Land — Island"})
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, attacker)
	gapB122AnswerDoors(t, g, me)
	if effectstest.BattlefieldHas(g, top) {
		t.Error("a land was put onto the battlefield")
	}
	if c, err := me.Library.Top(); err != nil || c.InstanceID != top {
		t.Error("the revealed land is not still on top of the library")
	}
}
