package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b3_cards_test.go — Final Fantasy (FIN), batches 3 and 4 (#4462). Each
// test plays the real catalog entry and asserts what its text says.

const (
	finB3NinjasBladesOracle    = "cdaffb13-c2a4-4bc5-a9d6-c1a0d1686a81"
	finB3RenoAndRudeOracle     = "f749a6d6-2c55-439f-bb02-e95fa54c2ef3"
	finB3ResentfulOracle       = "71d6aa12-4888-458d-9dc8-9954b247a246"
	finB3HuntTheMarkOracle     = "d74f9a18-8dab-4549-be3f-9697ff68d73c"
	finB3VincentsOracle        = "814679b4-51bd-4b07-a7b8-d3375d36ff45"
	finB3CoralSwordOracle      = "d4f0dd14-bc2a-4c6d-bfd5-bc5838ab18a5"
	finB3FreyaOracle           = "7e3085d8-9d46-41e0-ad9d-54c44d9ef267"
	finB3HillGigasOracle       = "137594f7-2422-461b-a3ae-7ce3c135e0c7"
	finB3ItemShopkeepOracle    = "7532f0ea-9d13-43ca-9b7d-2d22f2a66017"
	finB3LightOfJudgmentOracle = "932177b8-70f8-4908-a48f-fe03204daddc"
	finB3RandomEncounterOracle = "324493fc-a205-4177-aa94-cf4672844c81"
	finB3BlitzballOracle       = "9c0255ff-6579-4ce4-810d-01df9922ab33"
	finB3SchemesOracle         = "945e56b9-709f-4bb3-8a6b-0a04d9202a03"
	finB3UnexpectedOracle      = "f4497af5-8b81-46b3-bb52-eaed1999809b"
	finB3VaanOracle            = "24fd1f47-bb15-45c6-a7b1-ba8eb3bb8b1d"
	finB3ClashOracle           = "b1c38b70-575d-4d87-bc50-0b77b48ee823"
	finB3DiamondWeaponOracle   = "3beba02f-4d50-4a93-a13b-78d4595bfbfc"
	finB3EsperOriginsOracle    = "947b40a2-e272-4ef6-85fb-3d767a21bc7a"
	finB3GalufOracle           = "9268d369-173e-4325-98f6-348beeb8ba62"
)

func TestFinB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finB3NinjasBladesOracle:        "Ninja's Blades",
		finB3RenoAndRudeOracle:         "Reno and Rude",
		finB3ResentfulOracle:           "Resentful Revelation",
		finB3HuntTheMarkOracle:         "Sidequest: Hunt the Mark",
		finB3HuntTheMarkOracle + "#1":  "Yiazmat, Ultimate Mark",
		finB3VincentsOracle:            "Vincent's Limit Break",
		finB3CoralSwordOracle:          "Coral Sword",
		finB3FreyaOracle:               "Freya Crescent",
		finB3HillGigasOracle:           "Hill Gigas",
		finB3ItemShopkeepOracle:        "Item Shopkeep",
		finB3LightOfJudgmentOracle:     "Light of Judgment",
		finB3RandomEncounterOracle:     "Random Encounter",
		finB3BlitzballOracle:           "Sidequest: Play Blitzball",
		finB3BlitzballOracle + "#1":    "World Champion, Celestial Weapon",
		finB3SchemesOracle:             "Sorceress's Schemes",
		finB3UnexpectedOracle:          "Unexpected Request",
		finB3VaanOracle:                "Vaan, Street Thief",
		finB3ClashOracle:               "Clash of the Eikons",
		finB3DiamondWeaponOracle:       "Diamond Weapon",
		finB3EsperOriginsOracle:        "Esper Origins",
		finB3EsperOriginsOracle + "#1": "Summon: Esper Maduin",
		finB3GalufOracle:               "Galuf's Final Act",
	} {
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

// finB3Has reports whether the card, read through the layers, has the
// ability.
func finB3Has(c game.Card, ability string) bool {
	for _, a := range c.Effective().Abilities {
		if a == ability {
			return true
		}
	}
	return false
}

// finB3Settle passes priority around until the stack is empty, answering
// the target prompt a triggered ability raises with `pick` (a card), as
// many times as it asks.
func finB3Settle(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 6; i++ {
		passPriorityAroundTable(t, g)
		if latestPickTarget(g, chooser) == nil {
			return
		}
		pickCard(t, g, chooser, pick)
	}
	t.Fatal("the target prompt kept coming back")
}

// finB3Landscape seeds a creature under `owner`, not summoning sick.
func finB3Creature(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	return b12Creature(g, owner, name, typeLine, p, tough)
}

// --- Coral Sword ---------------------------------------------------------------

func TestFinB3CoralSwordAttachesToTheTargetAndGivesItFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := finB3Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	sword := castCatalogSpell(t, g, "Coral Sword", "Artifact — Equipment", finB3CoralSwordOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Fatal("the enters trigger did not ask for a target")
	}
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	sw, ok := battlefieldCard(g, sword)
	if !ok || !sw.IsAttachedTo(bear) {
		t.Fatalf("Coral Sword is not attached to the bear: %+v", sw.AttachedTo)
	}
	v := s58p6View(t, g, bear)
	if v.CurrentPower() != 3 || !finB3Has(v, "first strike") {
		t.Errorf("bear is %d/%d first strike=%v, want 3/2 with first strike", v.CurrentPower(), v.CurrentToughness(), finB3Has(v, "first strike"))
	}
	if o := s58p6View(t, g, other); o.CurrentPower() != 2 || finB3Has(o, "first strike") {
		t.Error("the other creature got the Sword's bonus or the first strike")
	}
}

func TestFinB3CoralSwordFirstStrikeLastsOnlyUntilEndOfTurnButTheBonusStays(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Coral Sword", "Artifact — Equipment", finB3CoralSwordOracle, nil)
	finB3Settle(t, g, me.ID, bear)
	for i := 0; i < 40 && !(g.Turn.ActiveSeat != seat); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	v := s58p6View(t, g, bear)
	if finB3Has(v, "first strike") {
		t.Error("first strike outlasted the turn")
	}
	if v.CurrentPower() != 3 {
		t.Errorf("power %d after the turn, want the Sword's +1/+0 still on it", v.CurrentPower())
	}
}

// --- Freya Crescent ------------------------------------------------------------

func TestFinB3FreyaFliesOnlyOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	freya := pushCatalogPermanent(g, me.ID, "Freya Crescent", "Legendary Creature — Rat Knight", finB3FreyaOracle, false)
	if !finB3Has(s58p6View(t, g, freya), "flying") {
		t.Error("Freya lacks flying on her controller's turn")
	}
	g.WithWriteLock(func() { g.Turn.ActiveSeat = (seat + 1) % len(g.Seats) })
	if finB3Has(s58p6View(t, g, freya), "flying") {
		t.Error("Freya has flying on another player's turn")
	}
}

func TestFinB3FreyaManaPaysEquipmentSpellsAndEquipOnly(t *testing.T) {
	g, me, _ := spendTable(t)
	freya := pushCatalogPermanent(g, me.ID, "Freya Crescent", "Legendary Creature — Rat Knight", finB3FreyaOracle, false)
	if err := g.ActivateManaAbility(me.ID, freya, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Freya's mana ability: %v", err)
	}
	if me.ManaPool.Total() != 1 {
		t.Fatalf("pool = %v, want {R}", me.ManaPool)
	}
	bear := handSpell(me, "Test Bear", "Creature — Bear", "{R}")
	refusedForMana(t, g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}), "a creature spell")
	rock := handSpell(me, "Test Rock", "Artifact", "{R}")
	refusedForMana(t, g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}), "a non-Equipment artifact")
	blade := handSpell(me, "Test Blade", "Artifact — Equipment", "{R}")
	if err := g.CastSpell(me.ID, blade, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an Equipment spell with Freya's mana: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want it spent", me.ManaPool)
	}
}

func TestFinB3FreyaManaPaysAnEquipAbility(t *testing.T) {
	g, me, _ := spendTable(t)
	freya := pushCatalogPermanent(g, me.ID, "Freya Crescent", "Legendary Creature — Rat Knight", finB3FreyaOracle, false)
	spear := pushCatalogPermanent(g, me.ID, "Shadowspear", equipTypeLine, shadowspearOracle, false)
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, freya, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("Freya's mana ability: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, spear, 1, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("equip with Freya's mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if sw, _ := battlefieldCard(g, spear); !sw.IsAttachedTo(bear) {
		t.Error("the Spear did not attach")
	}
}

// --- Hill Gigas ----------------------------------------------------------------

func TestFinB3HillGigasMountaincyclingFetchesAMountain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain := uuid.New()
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: mountain, Name: "Mountain", TypeLine: "Basic Land — Mountain", Owner: me.ID, Controller: me.ID})
	id, _ := cycleFromHand(t, g, "Hill Gigas", "Creature — Giant", finB3HillGigasOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Hill Gigas is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no search prompt")
	}
	if hasID(pick.ChooseCards, forest) || !hasID(pick.ChooseCards, mountain) {
		t.Errorf("the search offers %v, want only the Mountain", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, mountain)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(mountain) {
		t.Error("the Mountain did not come to hand")
	}
}

func TestFinB3HillGigasHasTrampleAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gigas := pushCatalogPermanent(g, me.ID, "Hill Gigas", "Creature — Giant", finB3HillGigasOracle, true)
	v := s58p6View(t, g, gigas)
	if !finB3Has(v, "trample") || !finB3Has(v, "haste") {
		t.Errorf("abilities %v, want trample and haste", v.Effective().Abilities)
	}
}

// --- Resentful Revelation ------------------------------------------------------

func finB3LibraryCards(me *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestFinB3ResentfulRevelationTakesOneOfTheTopThreeAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := finB3LibraryCards(me, "Fourth", "Third", "Second", "First") // First is on top
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Resentful Revelation", "Sorcery", finB3ResentfulOracle, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if len(pick.ChooseCards) != 3 || hasID(pick.ChooseCards, ids[0]) {
		t.Errorf("offered %d cards (fourth card offered: %v), want only the top three", len(pick.ChooseCards), hasID(pick.ChooseCards, ids[0]))
	}
	answerChooseCards(t, g, me.ID, ids[2])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[2]) {
		t.Error("the chosen card did not come to hand")
	}
	if !me.Graveyard.Contains(ids[1]) || !me.Graveyard.Contains(ids[3]) {
		t.Error("the rest did not go to the graveyard")
	}
	if got := lib - me.Library.Size(); got != 3 {
		t.Errorf("library lost %d cards, want 3", got)
	}
	if !me.Library.Contains(ids[0]) {
		t.Error("the fourth card left the library")
	}
}

func TestFinB3ResentfulRevelationFlashbackExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := finB3LibraryCards(me, "C", "B", "A")
	id := seedGraveyardCard(t, g, "Resentful Revelation", "Sorcery", finB3ResentfulOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, ids[0])
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the chosen card did not come to hand")
	}
}

// --- Item Shopkeep -------------------------------------------------------------

func finB3Equip(t *testing.T, g *game.Game, equipment, creature uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: creature}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
}

func TestFinB3ItemShopkeepGivesAnAttackingEquippedCreatureMenace(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], finABOpp(g)
	pushCatalogPermanent(g, me.ID, "Item Shopkeep", "Creature — Human Citizen", finB3ItemShopkeepOracle, false)
	armed := finB3Creature(g, me.ID, "Armed", "Creature — Bear", 2, 2)
	bare := finB3Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	finB3Equip(t, g, finABEquipment(g, me.ID, "Sword"), armed)
	advanceToPrecombatMainOf(t, g, seat)
	finABAttack(t, g, opp.ID, armed, bare)
	finB3Settle(t, g, me.ID, armed)
	if !finB3Has(s58p6View(t, g, armed), "menace") {
		t.Error("the equipped attacker did not gain menace")
	}
	if finB3Has(s58p6View(t, g, bare), "menace") {
		t.Error("the unequipped attacker gained menace")
	}
}

func TestFinB3ItemShopkeepHasNoTargetWithoutAnEquippedAttacker(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], finABOpp(g)
	pushCatalogPermanent(g, me.ID, "Item Shopkeep", "Creature — Human Citizen", finB3ItemShopkeepOracle, false)
	armed := finB3Creature(g, me.ID, "Armed", "Creature — Bear", 2, 2)
	bare := finB3Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	finB3Equip(t, g, finABEquipment(g, me.ID, "Sword"), armed)
	advanceToPrecombatMainOf(t, g, seat)
	finABAttack(t, g, opp.ID, bare)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the trigger asked for a target with no attacking equipped creature")
	}
	if finB3Has(s58p6View(t, g, armed), "menace") || finB3Has(s58p6View(t, g, bare), "menace") {
		t.Error("a creature gained menace")
	}
}

// --- Light of Judgment ---------------------------------------------------------

func TestFinB3LightOfJudgmentDamagesAndDestroysAChosenEquipment(t *testing.T) {
	g := newCatalogGame(t)
	opp := finABOpp(g)
	me := g.Seats[g.Turn.ActiveSeat]
	big := finB3Creature(g, opp.ID, "Big", "Creature — Giant", 7, 7)
	keep := finABEquipment(g, opp.ID, "Keep")
	doom := finABEquipment(g, opp.ID, "Doom")
	elsewhere := finABEquipment(g, opp.ID, "Elsewhere")
	finB3Equip(t, g, keep, big)
	finB3Equip(t, g, doom, big)
	castCatalogSpell(t, g, "Light of Judgment", "Instant", finB3LightOfJudgmentOracle, tdCard(big))
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no Equipment prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 || !hasID(pick.ChooseCards, keep) || !hasID(pick.ChooseCards, doom) || hasID(pick.ChooseCards, elsewhere) {
		t.Errorf("prompt %d..%d over %v, want up to one of the two attached Equipment", pick.ChooseMin, pick.ChooseMax, pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, doom)
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, big); v.DamageMarked != 6 {
		t.Errorf("the creature has %d damage, want 6", v.DamageMarked)
	}
	if g.Battlefield.Contains(doom) || !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(elsewhere) {
		t.Error("exactly the chosen Equipment should have been destroyed")
	}
}

func TestFinB3LightOfJudgmentCanDestroyNothingAndTargetsABareCreature(t *testing.T) {
	g := newCatalogGame(t)
	opp := finABOpp(g)
	me := g.Seats[g.Turn.ActiveSeat]
	big := finB3Creature(g, opp.ID, "Big", "Creature — Giant", 7, 7)
	sword := finABEquipment(g, opp.ID, "Sword")
	finB3Equip(t, g, sword, big)
	castCatalogSpell(t, g, "Light of Judgment", "Instant", finB3LightOfJudgmentOracle, tdCard(big))
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Error("the Equipment was destroyed though none was chosen")
	}

	g2 := newCatalogGame(t)
	bare := finB3Creature(g2, finABOpp(g2).ID, "Bare", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g2, "Light of Judgment", "Instant", finB3LightOfJudgmentOracle, tdCard(bare))
	passPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(bare) {
		t.Error("6 damage did not kill the bare 2/2")
	}
}

// --- Vincent's Limit Break -----------------------------------------------------

func TestFinB3VincentsLimitBreakSetsBasePTByTierAndReturnsTheCreature(t *testing.T) {
	for _, tc := range []struct {
		tier    int
		p, tgh  int
		tierWas string
	}{{0, 3, 2, "Galian Beast"}, {1, 5, 2, "Death Gigas"}, {2, 7, 2, "Hellmasker"}} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
		g.WithWriteLock(func() {
			if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 1); err != nil {
				t.Fatal(err)
			}
		})
		if _, err := castModalErr(t, g, "Vincent's Limit Break", "Instant", finB3VincentsOracle, []int{tc.tier}, tdCard(bear)); err != nil {
			t.Fatalf("%s: %v", tc.tierWas, err)
		}
		passPriorityAroundTable(t, g)
		v := s58p6View(t, g, bear)
		if v.CurrentPower() != tc.p+1 || v.CurrentToughness() != tc.tgh+1 {
			t.Errorf("%s: %d/%d, want base %d/%d with its +1/+1 counter on top", tc.tierWas, v.CurrentPower(), v.CurrentToughness(), tc.p, tc.tgh)
		}
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(bear, game.DestroyOptions{}); err != nil {
				t.Fatal(err)
			}
		})
		passPriorityAroundTable(t, g)
		back := findBattlefieldByName(g, "Bear")
		c, ok := battlefieldCard(g, back)
		if back == uuid.Nil || !ok || !c.Tapped {
			t.Errorf("%s: the creature did not return tapped", tc.tierWas)
		}
	}
}

func TestFinB3VincentsLimitBreakOnlyTargetsYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	theirs := finB3Creature(g, finABOpp(g).ID, "Theirs", "Creature — Bear", 2, 2)
	if _, err := castModalErr(t, g, "Vincent's Limit Break", "Instant", finB3VincentsOracle, []int{0}, tdCard(theirs)); err == nil {
		t.Error("an opponent's creature was accepted as the target")
	}
}

// --- Ninja's Blades ------------------------------------------------------------

func TestFinB3NinjasBladesJobSelectMakesANinjaHero(t *testing.T) {
	g := newCatalogGame(t)
	blades := castCatalogSpell(t, g, "Ninja's Blades", "Artifact — Equipment", finB3NinjasBladesOracle, nil)
	passPriorityAroundTable(t, g)
	hero := ysBHero(t, g, blades)
	v := s58p6View(t, g, hero)
	if v.CurrentPower() != 2 || v.CurrentToughness() != 2 || !v.HasSubtype("Ninja") {
		t.Errorf("Hero is %d/%d %v, want a 2/2 Hero Ninja", v.CurrentPower(), v.CurrentToughness(), v.Effective().Subtypes)
	}
}

// finB3BladesAttack attacks the first opponent with a bear carrying the
// Blades and returns the player, the bear and the hand card the loot can
// discard (a card of the given cost).
func finB3BladesAttack(t *testing.T, g *game.Game, cost string) (me, opp *game.Player, discard uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me, opp = g.Seats[seat], finABOpp(g)
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blades := pushCatalogPermanent(g, me.ID, "Ninja's Blades", "Artifact — Equipment", finB3NinjasBladesOracle, false)
	finB3Equip(t, g, blades, bear)
	me.Hand.Cards = nil
	discard = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: discard, Name: "Fodder", TypeLine: "Sorcery", ManaCost: cost, Owner: me.ID, Controller: me.ID})
	advanceToPrecombatMainOf(t, g, seat)
	finABAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	return me, opp, discard
}

func TestFinB3NinjasBladesLootsAndTheVictimLosesLifeEqualToTheDiscardedManaValue(t *testing.T) {
	g := newCatalogGame(t)
	opp := finABOpp(g)
	life := opp.Life
	me, _, discard := finB3BladesAttack(t, g, "{2}{R}")
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no discard prompt after the draw")
	}
	if len(pick.ChooseCards) != 2 {
		t.Errorf("the discard offers %d cards, want the old card and the drawn one", len(pick.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, discard)
	passPriorityAroundTable(t, g)
	// 2 combat damage from the bear (3 with the Blades' +1/+1) and 3 life lost.
	if got := life - opp.Life; got != 3+3 {
		t.Errorf("the victim lost %d life, want 3 combat damage and 3 for the discarded mana value", got)
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand size %d, want the one drawn card", me.Hand.Size())
	}
}

func TestFinB3NinjasBladesDiscardingALandCostsNoLife(t *testing.T) {
	g := newCatalogGame(t)
	opp := finABOpp(g)
	life := opp.Life
	me, _, discard := finB3BladesAttack(t, g, "")
	answerChooseCards(t, g, me.ID, discard)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the victim lost %d life, want only the 3 combat damage", got)
	}
}

func TestFinB3NinjasBladesXInTheDiscardedCostCountsAsZero(t *testing.T) {
	g := newCatalogGame(t)
	opp := finABOpp(g)
	life := opp.Life
	me, _, discard := finB3BladesAttack(t, g, "{X}{R}")
	answerChooseCards(t, g, me.ID, discard)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3+1 {
		t.Errorf("the victim lost %d life, want 3 combat damage and 1 for {X}{R} with X = 0", got)
	}
}

// --- Reno and Rude -------------------------------------------------------------

func finB3RenoAttack(t *testing.T, g *game.Game) (me, opp *game.Player, reno uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me, opp = g.Seats[seat], finABOpp(g)
	reno = pushCatalogPermanent(g, me.ID, "Reno and Rude", "Legendary Creature — Human Assassin", finB3RenoAndRudeOracle, false)
	advanceToPrecombatMainOf(t, g, seat)
	finABAttack(t, g, opp.ID, reno)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	return me, opp, reno
}

func TestFinB3RenoAndRudeSacrificeLetsYouCastTheExiledCardWithAnyMana(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	opp := finABOpp(g)
	me := g.Seats[seat]
	top := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: top, Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	fodder := finB3Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	finB3RenoAttack(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card of the damaged player's library was not exiled")
	}
	answerChooseCards(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	// Mana of any type: two red mana pay for {1}{G}.
	g.WithWriteLock(func() { g.Turn.Step = game.StepPostcombatMain })
	floatMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting the exiled card with red mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Error("the exiled creature did not resolve onto the battlefield")
	}
}

func TestFinB3RenoAndRudeNoSacrificeMeansNoPermission(t *testing.T) {
	g := newCatalogGame(t)
	opp := finABOpp(g)
	me := g.Seats[g.Turn.ActiveSeat]
	top := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: top, Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	fodder := finB3Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	finB3RenoAttack(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fodder) {
		t.Error("a creature was sacrificed though none was chosen")
	}
	g.WithWriteLock(func() { g.Turn.Step = game.StepPostcombatMain })
	floatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the exiled card was castable without a sacrifice")
	}
}

func TestFinB3RenoAndRudeSacrificeOfAnArtifactAndNeverTheSource(t *testing.T) {
	g := newCatalogGame(t)
	opp := finABOpp(g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	land := b12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	_, _, reno := finB3RenoAttack(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no sacrifice prompt")
	}
	if !hasID(pick.ChooseCards, rock) || hasID(pick.ChooseCards, reno) || hasID(pick.ChooseCards, land) {
		t.Errorf("the sacrifice offers %v, want the artifact and neither Reno and Rude nor the land", pick.ChooseCards)
	}
}

// --- Sidequest: Hunt the Mark // Yiazmat, Ultimate Mark ------------------------

func TestFinB3HuntTheMarkDestroysUpToOneCreatureWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := finB3Creature(g, finABOpp(g).ID, "Victim", "Creature — Bear", 2, 2)
	importAndCast(t, g, finB3HuntRow(), me)
	finB3Settle(t, g, me.ID, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
}

func finB3HuntRow() cards.Card {
	return transformRow(finB3HuntTheMarkOracle, "Sidequest: Hunt the Mark", "Enchantment", "{3}{B}{B}",
		"Yiazmat, Ultimate Mark", "Legendary Creature — Dragon", "5", "6", []string{"B"})
}

func finB3EndStep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
}

func TestFinB3HuntTheMarkMakesATreasureAfterAnOpponentCreatureDiedAndTransformsAtThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := finB3Creature(g, finABOpp(g).ID, "Victim", "Creature — Bear", 2, 2)
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Treasure", "Token Artifact — Treasure")
	}
	quest := importAndCast(t, g, finB3HuntRow(), me)
	finB3Settle(t, g, me.ID, victim)
	finB3EndStep(t, g)
	if got := countOnBattlefield(g, "Treasure", me.ID); got != 3 {
		t.Errorf("%d Treasures, want the two seeded plus a new one", got)
	}
	c, ok := battlefieldCard(g, quest)
	if !ok || c.ActiveFace != 1 || c.Name != "Yiazmat, Ultimate Mark" {
		t.Errorf("the enchantment is %q on face %d, want Yiazmat on face 1", c.Name, c.ActiveFace)
	}
}

func TestFinB3HuntTheMarkDoesNotTransformBelowThreeTreasures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := finB3Creature(g, finABOpp(g).ID, "Victim", "Creature — Bear", 2, 2)
	quest := importAndCast(t, g, finB3HuntRow(), me)
	finB3Settle(t, g, me.ID, victim)
	finB3EndStep(t, g)
	if got := countOnBattlefield(g, "Treasure", me.ID); got != 1 {
		t.Errorf("%d Treasures, want 1", got)
	}
	if c, _ := battlefieldCard(g, quest); c.ActiveFace != 0 {
		t.Error("the enchantment transformed with one Treasure")
	}
}

func TestFinB3HuntTheMarkIgnoresYourOwnCreaturesDying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := finB3Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	importAndCast(t, g, finB3HuntRow(), me)
	finB3Settle(t, g, me.ID, mine)
	if g.Battlefield.Contains(mine) {
		t.Fatal("setup: the creature should have been destroyed")
	}
	finB3EndStep(t, g)
	if got := countOnBattlefield(g, "Treasure", me.ID); got != 0 {
		t.Errorf("%d Treasures after only your own creature died, want none", got)
	}
}

func TestFinB3YiazmatSacrificesToBecomeIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	victim := finB3Creature(g, finABOpp(g).ID, "Victim", "Creature — Bear", 2, 2)
	for i := 0; i < 2; i++ {
		b12Permanent(g, me.ID, "Treasure", "Token Artifact — Treasure")
	}
	quest := importAndCast(t, g, finB3HuntRow(), me)
	finB3Settle(t, g, me.ID, victim)
	finB3EndStep(t, g)
	fodder := finB3Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	floatMana(t, g, me, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, quest, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{quest}}); err == nil {
		t.Error("Yiazmat sacrificed itself to its own cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, quest, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("Yiazmat's ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	v := s58p6View(t, g, quest)
	if !finB3Has(v, "indestructible") || !v.Tapped {
		t.Errorf("Yiazmat indestructible=%v tapped=%v, want both", finB3Has(v, "indestructible"), v.Tapped)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	if v.CurrentPower() != 5 || v.CurrentToughness() != 6 {
		t.Errorf("Yiazmat is %d/%d, want 5/6", v.CurrentPower(), v.CurrentToughness())
	}
}

// --- Random Encounter ----------------------------------------------------------

func TestFinB3RandomEncounterPutsMilledCreaturesOntoTheBattlefieldWithHasteThenReturnsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mk := func(name, typeLine string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID}
	}
	c1, c2 := mk("Milled One", "Creature — Bear"), mk("Milled Two", "Creature — Bear")
	land, spell := mk("Milled Land", "Basic Land — Forest"), mk("Milled Spell", "Sorcery")
	left := mk("Stays Put", "Creature — Bear")
	g.WithWriteLock(func() { me.Library.Cards = []game.Card{left, c1, land, c2, spell} })
	castCatalogSpell(t, g, "Random Encounter", "Sorcery", finB3RandomEncounterOracle, nil)
	passPriorityAroundTable(t, g)
	// The whole library is shuffled, so which four were milled varies; the
	// invariants do not: every creature card milled is on the battlefield,
	// hasty, and no other card is.
	milled := 0
	for _, c := range []game.Card{left, c1, land, c2, spell} {
		if me.Library.Contains(c.InstanceID) {
			continue
		}
		milled++
		isCreature := c.TypeLine == "Creature — Bear"
		if isCreature != g.Battlefield.Contains(c.InstanceID) {
			t.Errorf("%s: creature=%v onBattlefield=%v", c.Name, isCreature, g.Battlefield.Contains(c.InstanceID))
		}
		if isCreature {
			if !finB3Has(s58p6View(t, g, c.InstanceID), "haste") {
				t.Errorf("%s lacks haste", c.Name)
			}
		} else if !me.Graveyard.Contains(c.InstanceID) {
			t.Errorf("%s was milled but is not in the graveyard", c.Name)
		}
	}
	if milled != 4 {
		t.Fatalf("%d cards left the library, want 4", milled)
	}
	var entered []uuid.UUID
	for _, c := range []game.Card{left, c1, c2} {
		if g.Battlefield.Contains(c.InstanceID) {
			entered = append(entered, c.InstanceID)
		}
	}
	if len(entered) == 0 {
		t.Fatal("no creature card was milled; the shuffle cannot hide all three from four cards")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range entered {
		if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
			t.Errorf("creature %s was not returned to its owner's hand at the end step", id)
		}
	}
}

func TestFinB3RandomEncounterFlashbackExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Random Encounter", "Sorcery", finB3RandomEncounterOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}
}

func TestFinB3RandomEncounterReturnsOnlyTheCreaturesThatStayed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		me.Library.Cards = nil
		for i := 0; i < 4; i++ {
			id := uuid.New()
			ids = append(ids, id)
			me.Library.Cards = append(me.Library.Cards, game.Card{InstanceID: id, Name: "Milled", TypeLine: "Creature — Bear",
				Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		}
	})
	castCatalogSpell(t, g, "Random Encounter", "Sorcery", finB3RandomEncounterOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("creature %s did not enter", id)
		}
	}
	// One is exiled before the end step: it is not "those creatures" any
	// more and must not be dragged back.
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(ids[0]) })
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(ids[0]) {
		t.Error("an exiled creature came back to hand")
	}
	for _, id := range ids[1:] {
		if !me.Hand.Contains(id) {
			t.Errorf("creature %s was not returned", id)
		}
	}
}

// --- Sidequest: Play Blitzball // World Champion, Celestial Weapon -------------

func finB3BlitzballRow() cards.Card {
	return transformRow(finB3BlitzballOracle, "Sidequest: Play Blitzball", "Enchantment", "{2}{R}",
		"World Champion, Celestial Weapon", "Legendary Artifact — Equipment", "", "", []string{"R"})
}

func TestFinB3BlitzballPumpsATargetAtTheBeginningOfYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	importAndCast(t, g, finB3BlitzballRow(), me)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	finB3Settle(t, g, me.ID, bear)
	if v := s58p6View(t, g, bear); v.CurrentPower() != 4 {
		t.Errorf("power %d, want 4 (+2/+0)", v.CurrentPower())
	}
}

// finB3Blitzball attacks the first opponent with a creature of the given
// power, with Sidequest: Play Blitzball (and its +2/+0) on the table, and
// leaves the game at the end of combat.
func finB3Blitzball(t *testing.T, g *game.Game, power int) (me *game.Player, quest, bear uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp := finABOpp(g)
	bear = finB3Creature(g, me.ID, "Bear", "Creature — Bear", power, power)
	quest = importAndCast(t, g, finB3BlitzballRow(), me)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	finB3Settle(t, g, me.ID, bear)
	finABAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	return me, quest, bear
}

func TestFinB3BlitzballTransformsAfterSixCombatDamageAndAttachesToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, quest, bear := finB3Blitzball(t, g, 4)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no attach prompt after the transform")
	}
	answerChooseCards(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, quest)
	if !ok || c.ActiveFace != 1 || c.Name != "World Champion, Celestial Weapon" {
		t.Fatalf("the enchantment is %q on face %d, want the Equipment", c.Name, c.ActiveFace)
	}
	if !c.IsAttachedTo(bear) {
		t.Error("the Equipment is not attached to the creature")
	}
	v := s58p6View(t, g, bear)
	if !finB3Has(v, "double strike") {
		t.Error("the equipped creature lacks double strike")
	}
	// 4 base + 2 from the combat trigger (until end of turn) + 2 from the Equipment.
	if v.CurrentPower() != 8 {
		t.Errorf("power %d, want 8", v.CurrentPower())
	}
}

func TestFinB3BlitzballStaysPutBelowSixCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	_, quest, _ := finB3Blitzball(t, g, 3) // 3 + 2 = 5
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("an attach prompt came up with only 5 combat damage dealt")
	}
	if c, _ := battlefieldCard(g, quest); c.ActiveFace != 0 {
		t.Error("the enchantment transformed after 5 combat damage")
	}
}

func TestFinB3BlitzballSixDamageCanComeFromSeveralSources(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], finABOpp(g)
	a := finB3Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := finB3Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	quest := importAndCast(t, g, finB3BlitzballRow(), me)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepBeginCombat)
	finB3Settle(t, g, me.ID, a) // a is 4/2, b 2/2: 6 in all
	finABAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, b)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, quest); c.ActiveFace != 1 || !c.IsAttachedTo(b) {
		t.Errorf("face %d, attached to b: %v", c.ActiveFace, c.IsAttachedTo(b))
	}
}

// --- Sorceress's Schemes -------------------------------------------------------

func TestFinB3SorceresssSchemesReturnsAnInstantFromYourGraveyardAndAddsRed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := seedGraveyardCard(t, g, "Shock", "Instant", "")
	castCatalogSpell(t, g, "Sorceress's Schemes", "Sorcery", finB3SchemesOracle, tdCard(bolt))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Error("the instant did not return to hand")
	}
	if mcPoolCount(me, "R") != 1 {
		t.Errorf("pool %v, want {R}", me.ManaPool)
	}
}

func TestFinB3SorceresssSchemesReturnsAnExiledCardWithFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	exiled := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: exiled, Name: "Resentful Revelation", TypeLine: "Sorcery", OracleID: finB3ResentfulOracle,
		Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Sorceress's Schemes", "Sorcery", finB3SchemesOracle, tdCard(exiled))
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(exiled) || g.Exile.Contains(exiled) {
		t.Error("the exiled flashback card did not return to hand")
	}
	if mcPoolCount(me, "R") != 1 {
		t.Errorf("pool %v, want {R}", me.ManaPool)
	}
}

func TestFinB3SorceresssSchemesRefusesTheWrongCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	// An exiled card without flashback, an exiled flashback card you don't
	// own, a creature in your graveyard and an opponent's instant.
	plain := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: plain, Name: "Plain", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: theirs, Name: "Resentful Revelation", TypeLine: "Sorcery", OracleID: finB3ResentfulOracle,
		Owner: opp.ID, Controller: opp.ID})
	bear := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	oppBolt := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: oppBolt, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	for name, id := range map[string]uuid.UUID{"plain exile": plain, "their flashback card": theirs, "your creature": bear, "their instant": oppBolt} {
		if err := castCatalogSpellErr(t, g, "Sorceress's Schemes", "Sorcery", finB3SchemesOracle, tdCard(id)); err == nil {
			t.Errorf("%s was accepted as the target", name)
		}
	}
}

func TestFinB3SorceresssSchemesFlashbackExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := seedGraveyardCard(t, g, "Shock", "Instant", "")
	id := seedGraveyardCard(t, g, "Sorceress's Schemes", "Sorcery", finB3SchemesOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: tdCard(bolt)}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || !me.Hand.Contains(bolt) {
		t.Error("the flashed-back spell should be exiled after returning the instant")
	}
}

// --- Unexpected Request --------------------------------------------------------

func TestFinB3UnexpectedRequestStealsUntapsAndMayAttachAnEquipmentUntilTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], finABOpp(g)
	theirs := finB3Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	sword := finABEquipment(g, me.ID, "Sword")
	theirSword := finABEquipment(g, opp.ID, "Their Sword")
	castCatalogSpell(t, g, "Unexpected Request", "Sorcery", finB3UnexpectedOracle, tdCard(theirs))
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, theirs)
	if c.Controller != me.ID || c.Tapped {
		t.Fatalf("controller %v tapped %v, want mine and untapped", c.Controller, c.Tapped)
	}
	v := s58p6View(t, g, theirs)
	if !finB3Has(v, "haste") {
		t.Error("the stolen creature lacks haste")
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no attach prompt")
	}
	if !hasID(pick.ChooseCards, sword) || hasID(pick.ChooseCards, theirSword) {
		t.Errorf("the attach offers %v, want only the Equipment I control", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	if s, _ := battlefieldCard(g, sword); !s.IsAttachedTo(theirs) {
		t.Fatal("the Equipment is not attached to the stolen creature")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if s, _ := battlefieldCard(g, sword); s.IsAttached() {
		t.Error("the Equipment was not unattached at the beginning of the end step")
	}
	if !g.Battlefield.Contains(sword) {
		t.Error("the Equipment left the battlefield")
	}
}

func TestFinB3UnexpectedRequestDecliningTheAttachSchedulesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	theirs := finB3Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	sword := finABEquipment(g, me.ID, "Sword")
	other := finB3Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	finB3Equip(t, g, sword, other)
	castCatalogSpell(t, g, "Unexpected Request", "Sorcery", finB3UnexpectedOracle, tdCard(theirs))
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if s, _ := battlefieldCard(g, sword); !s.IsAttachedTo(other) {
		t.Error("the Equipment on my other creature was unattached though nothing was attached by the spell")
	}
}

func TestFinB3UnexpectedRequestReturnsControlAtTheEndOfTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	opp := finABOpp(g)
	theirs := finB3Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Unexpected Request", "Sorcery", finB3UnexpectedOracle, tdCard(theirs))
	passPriorityAroundTable(t, g)
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if c, _ := battlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("control did not return when the turn ended")
	}
}

// --- Vaan, Street Thief --------------------------------------------------------

// finB3VaanAttack puts Vaan (and the other creatures given) onto the
// battlefield, stacks the opponent's library with `top`, and attacks with
// Vaan. It returns once the combat damage trigger has put the exile (and
// any cast offer) up.
func finB3VaanAttack(t *testing.T, g *game.Game, top game.Card) (me, opp *game.Player, vaan uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	me, opp = g.Seats[seat], finABOpp(g)
	top.InstanceID = uuid.New()
	top.Owner, top.Controller = opp.ID, opp.ID
	opp.Library.PushTop(top)
	vaan = pushCatalogPermanent(g, me.ID, "Vaan, Street Thief", "Legendary Creature — Human Scout", finB3VaanOracle, false)
	advanceToPrecombatMainOf(t, g, seat)
	finABAttack(t, g, opp.ID, vaan)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	return me, opp, vaan
}

func TestFinB3VaanCastingTheExiledCardMakesNoTreasureAndPutsCountersOnYourThieves(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pirate := finB3Creature(g, me.ID, "Pirate", "Creature — Human Pirate", 1, 1)
	cleric := finB3Creature(g, me.ID, "Cleric", "Creature — Human Cleric", 1, 1)
	theirCard := game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	_, _, vaan := finB3VaanAttack(t, g, theirCard)
	offer := resolutionCastOfferFor(g, me.ID)
	if offer == nil {
		t.Fatal("no cast offer for the exiled card")
	}
	exiled := offer.DiscardOptions[0]
	if !g.Exile.Contains(exiled) {
		t.Fatal("the offered card is not in exile")
	}
	floatMana(t, g, me, "{R}{G}")
	if err := castThroughResolutionCast(t, g, me.ID, exiled, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting the exiled card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Treasure", me.ID); got != 0 {
		t.Errorf("%d Treasures after casting the card, want none", got)
	}
	if got := countersOn(g, vaan, game.CounterPlusOne); got != 1 {
		t.Errorf("Vaan has %d +1/+1 counters, want 1 (a spell you don't own was cast)", got)
	}
	if got := countersOn(g, pirate, game.CounterPlusOne); got != 1 {
		t.Errorf("the Pirate has %d counters, want 1", got)
	}
	if got := countersOn(g, cleric, game.CounterPlusOne); got != 0 {
		t.Errorf("the Cleric has %d counters, want none", got)
	}
}

func TestFinB3VaanDecliningTheCastMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	theirCard := game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	finB3VaanAttack(t, g, theirCard)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Treasure", me.ID); got != 1 {
		t.Errorf("%d Treasures after declining, want 1", got)
	}
}

func TestFinB3VaanAnExiledLandMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"}
	finB3VaanAttack(t, g, land)
	passPriorityAroundTable(t, g)
	if got := countOnBattlefield(g, "Treasure", me.ID); got != 1 {
		t.Errorf("%d Treasures after a land was exiled, want 1", got)
	}
}

func TestFinB3VaanOnlyScoutsPiratesAndRoguesSetOffTheDamageTrigger(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], finABOpp(g)
	top := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: top, Name: "Their Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	pushCatalogPermanent(g, me.ID, "Vaan, Street Thief", "Legendary Creature — Human Scout", finB3VaanOracle, false)
	cleric := finB3Creature(g, me.ID, "Cleric", "Creature — Human Cleric", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	finABAttack(t, g, opp.ID, cleric)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Error("a Cleric connecting set off Vaan's trigger")
	}
}

func TestFinB3VaanYourOwnSpellsPutNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vaan := pushCatalogPermanent(g, me.ID, "Vaan, Street Thief", "Legendary Creature — Human Scout", finB3VaanOracle, false)
	b24CastNoncreature(t, g, "Own Spell", "Instant", "{R}")
	passPriorityAroundTable(t, g)
	if got := countersOn(g, vaan, game.CounterPlusOne); got != 0 {
		t.Errorf("Vaan has %d counters after my own spell, want none", got)
	}
}

// --- Clash of the Eikons -------------------------------------------------------

func finB3RamuhSaga(g *game.Game, owner uuid.UUID, lore int) uuid.UUID {
	return finS3Saga(g, owner, "Summon: Esper Ramuh", "Enchantment Creature — Saga Wizard", finS3RamuhOracle, 3, 3, lore)
}

func TestFinB3ClashOfTheEikonsFightsRemovesAndPutsLoreCountersInPrintedOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], finABOpp(g)
	mine := finB3Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	theirs := finB3Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	down := finB3RamuhSaga(g, me.ID, 2)
	up := finB3RamuhSaga(g, me.ID, 1)
	castModal(t, g, "Clash of the Eikons", "Sorcery", finB3ClashOracle, []int{0, 1, 2}, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: theirs, Mode: 0, Slot: 1},
		{Kind: game.TargetCard, ID: down, Mode: 1},
		{Kind: game.TargetCard, ID: up, Mode: 2},
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the fight did not kill their 2/2")
	}
	if v := s58p6View(t, g, mine); v.DamageMarked != 2 {
		t.Errorf("my creature has %d damage, want 2 from the fight", v.DamageMarked)
	}
	if got := loreCountersOn(g, down); got != 1 {
		t.Errorf("the first Saga has %d lore counters, want 1 after the removal", got)
	}
	if got := loreCountersOn(g, up); got != 2 {
		t.Errorf("the second Saga has %d lore counters, want 2 after the addition", got)
	}
}

func TestFinB3ClashOfTheEikonsRemovingALoreCounterTriggersNothingButAddingOneDoes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := finB3RamuhSaga(g, me.ID, 2)
	castModal(t, g, "Clash of the Eikons", "Sorcery", finB3ClashOracle, []int{1}, tdCard(saga))
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, saga); loreCountersOn(g, saga) != 1 || v.CurrentPower() != 3 {
		t.Fatalf("after the removal: lore %d, power %d; want 1 and an untouched 3", loreCountersOn(g, saga), v.CurrentPower())
	}
	// Chapter II (Wizards get +1/+0) fires when the counter goes back on.
	castModal(t, g, "Clash of the Eikons", "Sorcery", finB3ClashOracle, []int{2}, tdCard(saga))
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, saga); loreCountersOn(g, saga) != 2 || v.CurrentPower() != 4 {
		t.Errorf("after the addition: lore %d, power %d; want 2 and chapter II's +1/+0", loreCountersOn(g, saga), v.CurrentPower())
	}
}

func TestFinB3ClashOfTheEikonsNeedsAtLeastOneMode(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := castModalErr(t, g, "Clash of the Eikons", "Sorcery", finB3ClashOracle, nil, nil); err == nil {
		t.Error("cast with no mode chosen")
	}
}

// --- Diamond Weapon ------------------------------------------------------------

func TestFinB3DiamondWeaponCostsOneLessPerPermanentCardInYourGraveyard(t *testing.T) {
	cast := func(t *testing.T, permanents int, pool string) error {
		t.Helper()
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		for i := 0; i < permanents; i++ {
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		}
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spent Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
		advanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Diamond Weapon", TypeLine: "Legendary Artifact Creature — Elemental",
			OracleID: finB3DiamondWeaponOracle, ManaCost: "{7}{G}{G}", Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID})
		floatMana(t, g, me, pool)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast(t, 5, "{G}{G}{C}{C}{C}{C}"); err != nil {
		t.Errorf("five permanent cards should make it {2}{G}{G}: %v", err)
	}
	refusedForMana(t, cast(t, 5, "{G}{G}{C}{C}{C}"), "one mana short of {2}{G}{G}")
	// The reduction takes only generic mana: nine cards still leave {G}{G}.
	if err := cast(t, 9, "{G}{G}"); err != nil {
		t.Errorf("nine permanent cards should make it {G}{G}: %v", err)
	}
	refusedForMana(t, cast(t, 9, "{G}"), "below the coloured pips")
	refusedForMana(t, cast(t, 0, "{G}{G}{C}{C}{C}{C}{C}{C}"), "an empty graveyard (the instant doesn't count)")
}

func TestFinB3DiamondWeaponPreventsCombatDamageButNotOtherDamage(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], finABOpp(g)
	weapon := pushCatalogPermanent(g, me.ID, "Diamond Weapon", "Legendary Artifact Creature — Elemental", finB3DiamondWeaponOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == weapon {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 8, 8
			}
		}
	})
	blocker := finB3Creature(g, opp.ID, "Blocker", "Creature — Giant", 6, 6)
	advanceToPrecombatMainOf(t, g, seat)
	pr7aBlockedCombat(t, g, opp.ID, weapon, blocker)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the 8/8 did not kill the blocker")
	}
	if v := s58p6View(t, g, weapon); v.DamageMarked != 0 {
		t.Errorf("Diamond Weapon has %d damage from combat, want it all prevented", v.DamageMarked)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, weapon, 3); err != nil {
			t.Fatal(err)
		}
	})
	if v := s58p6View(t, g, weapon); v.DamageMarked != 3 {
		t.Errorf("Diamond Weapon has %d damage from a spell, want 3", v.DamageMarked)
	}
}

// --- Esper Origins // Summon: Esper Maduin ---------------------------------------

func finB3EsperRow() cards.Card {
	return transformRow(finB3EsperOriginsOracle, "Esper Origins", "Sorcery", "{1}{G}",
		"Summon: Esper Maduin", "Enchantment Creature — Saga Elemental", "4", "4", []string{"G"})
}

func finB3SurveilKeep(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	c := surveilChoiceFor(g, who)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if len(c.ScryCards) != 2 {
		t.Errorf("surveil offers %d cards, want 2", len(c.ScryCards))
	}
	if err := g.ResolveSurveil(c.ID, who, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
}

func TestFinB3EsperOriginsFromHandSurveilsGainsLifeAndGoesToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := importAndCast(t, g, finB3EsperRow(), me)
	passPriorityAroundTable(t, g)
	finB3SurveilKeep(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if !me.Graveyard.Contains(id) || g.Battlefield.Contains(id) {
		t.Error("a spell cast from hand should end up in the graveyard")
	}
}

func finB3EsperInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := importToHand(finB3EsperRow(), me)
	var card game.Card
	g.WithWriteLock(func() {
		for i, c := range me.Hand.Cards {
			if c.InstanceID == id {
				card = c
				me.Hand.Cards = append(me.Hand.Cards[:i], me.Hand.Cards[i+1:]...)
				break
			}
		}
		me.Graveyard.PushTop(card)
	})
	return id
}

func TestFinB3EsperOriginsFlashbackPutsItOntoTheBattlefieldTransformedWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := finB3EsperInGraveyard(g, me)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	finB3SurveilKeep(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	var saga game.Card
	var found bool
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Summon: Esper Maduin" {
			saga, found = c, true
		}
	}
	if !found {
		t.Fatalf("Summon: Esper Maduin is not on the battlefield (exile has it: %v)", g.Exile.Contains(id))
	}
	if saga.ActiveFace != 1 || saga.Counters[game.CounterFinality] != 1 {
		t.Errorf("face %d, finality counters %d; want its back face and one finality counter", saga.ActiveFace, saga.Counters[game.CounterFinality])
	}
	if saga.Controller != me.ID {
		t.Error("it did not enter under its owner's control")
	}
	if g.Exile.Contains(id) {
		t.Error("the card is still in exile")
	}
	// Chapter I: the Saga entered with a lore counter, so the top card (a
	// permanent card) went to hand.
	if !me.Hand.Contains(top) {
		t.Error("chapter I did not put the revealed permanent card into hand")
	}
}

func TestFinB3SummonEsperMaduinChapterOneLeavesANonpermanentOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := finB3EsperInGraveyard(g, me)
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Keep both surveilled cards on top, in the order offered, so the
	// sorcery stays the top card the chapter reveals.
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(top) {
		t.Error("a nonpermanent card was put into hand")
	}
}

func finB3MaduinOnBattlefield(g *game.Game, owner uuid.UUID, lore int) uuid.UUID {
	c := deck.ToGameCard(finB3EsperRow(), false)
	c.Owner, c.Controller = owner, owner
	c.SetFace(1)
	c.Counters = map[string]int{game.CounterLore: lore}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestFinB3SummonEsperMaduinChapterTwoAddsGGAndThreePumpsTheOthers(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	saga := finB3MaduinOnBattlefield(g, me.ID, 1)
	ally := finB3Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	foe := finB3Creature(g, finABOpp(g).ID, "Foe", "Creature — Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if got := mcPoolCount(me, "G"); got != 2 {
		t.Errorf("chapter II added %d green mana, want 2", got)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	late := finB3Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	if v := s58p6View(t, g, ally); v.CurrentPower() != 4 || v.CurrentToughness() != 4 || !finB3Has(v, "trample") {
		t.Errorf("ally is %d/%d trample=%v, want 4/4 with trample", v.CurrentPower(), v.CurrentToughness(), finB3Has(v, "trample"))
	}
	if v := s58p6View(t, g, saga); v.CurrentPower() != 4 {
		t.Errorf("the Saga itself is %d/%d, want it unchanged (other creatures only)", v.CurrentPower(), v.CurrentToughness())
	}
	if v := s58p6View(t, g, foe); v.CurrentPower() != 2 {
		t.Error("an opponent's creature got the pump")
	}
	if v := s58p6View(t, g, late); v.CurrentPower() != 2 {
		t.Error("a creature that entered after chapter III resolved got the pump")
	}
}

func TestFinB3SummonEsperMaduinDiesToExileWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := finB3MaduinOnBattlefield(g, me.ID, 1)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(saga, game.CounterFinality, 1); err != nil {
			t.Fatal(err)
		}
		if err := g.DestroyPermanentForEffect(saga, game.DestroyOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(saga) || me.Graveyard.Contains(saga) || !g.Exile.Contains(saga) {
		t.Error("a creature with a finality counter should be exiled instead of dying")
	}
}

// --- Galuf's Final Act ---------------------------------------------------------

func TestFinB3GalufsFinalActPumpsAndPutsCountersEqualToItsPowerWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	friend := finB3Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Galuf's Final Act", "Instant", finB3GalufOracle, tdCard(bear))
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, bear); v.CurrentPower() != 3 {
		t.Fatalf("power %d, want 3", v.CurrentPower())
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear, game.DestroyOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	finB3Settle(t, g, me.ID, friend)
	if got := countersOn(g, friend, game.CounterPlusOne); got != 3 {
		t.Errorf("the target got %d +1/+1 counters, want 3 (its power as it died, +1/+0 included)", got)
	}
}

func TestFinB3GalufsFinalActTheDiesTriggerMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	friend := finB3Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Galuf's Final Act", "Instant", finB3GalufOracle, tdCard(bear))
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear, game.DestroyOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("choosing no target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, friend, game.CounterPlusOne); got != 0 {
		t.Errorf("a counter went on a creature nobody targeted: %d", got)
	}
}

func TestFinB3GalufsFinalActGrantEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	bear := finB3Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	friend := finB3Creature(g, me.ID, "Friend", "Creature — Bear", 1, 1)
	castCatalogSpell(t, g, "Galuf's Final Act", "Instant", finB3GalufOracle, tdCard(bear))
	passPriorityAroundTable(t, g)
	for i := 0; i < 40 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear, game.DestroyOptions{}); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || countersOn(g, friend, game.CounterPlusOne) != 0 {
		t.Error("the granted dies trigger outlived the turn")
	}
}
