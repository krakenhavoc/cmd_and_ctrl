package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s58_sauron_orcs_test.go — S58 PR 7: the Orc and amass cards of the
// Sauron deck (Mauhúr, Warg Rider, Saruman, Corsairs of Umbar, Orcish
// Medicine, Foray of Orcs, Lazotep Plating, Fall of Cair Andros,
// Barad-dûr).

const (
	mauhurOracle          = "9966cac0-331f-4627-be5a-5060a6ac5a32"
	wargRiderOracle       = "3f4dd9e2-6460-413e-a3b8-96b20bf57f8b"
	sarumanOracle         = "2bade11e-04e0-42a2-8861-9257c99a7c08"
	corsairsOracle        = "edc74fb9-a368-4f73-8147-f318a32a3d06"
	orcishMedicineOracle  = "2f9a5e7c-f463-4773-bf67-a07339ce9b5d"
	forayOfOrcsOracle     = "60bfcaef-9014-4707-b681-cc3920bf223e"
	lazotepPlatingOracle  = "ba0082fb-2d4c-489e-8140-93a6fa693fd0"
	fallOfCairOracle      = "33c0a8c0-d1d9-4b03-869d-65dab8ace3df"
	baradDurOracle        = "88159872-d37d-4847-b048-e4a9af6437bd"
	sauronTestOrcTypeLine = "Creature — Orc Warrior"
)

// --- Mauhúr -----------------------------------------------------------

func TestMauhurAddsACounterToAmassedArmiesAndOrcsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Mauhúr, Uruk-hai Captain", "Legendary Creature — Orc Soldier", mauhurOracle, false)
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	orc := pushDiesCreatureForTest(g, me.ID, "Orc", "", sauronTestOrcTypeLine, 2, 2)
	oppOrc := pushDiesCreatureForTest(g, g.Seats[1].ID, "Opposing Orc", "", sauronTestOrcTypeLine, 2, 2)

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{bear, orc, oppOrc} {
			if err := g.AddCounterByForEffect(me.ID, id, game.CounterPlusOne, 1); err != nil {
				t.Fatalf("AddCounterForEffect: %v", err)
			}
		}
	})
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("a Bear got %d +1/+1 counters, want 1 (Mauhúr only helps Armies, Goblins and Orcs)", got)
	}
	if got := countersOn(g, orc, game.CounterPlusOne); got != 2 {
		t.Errorf("an Orc got %d +1/+1 counters, want 2", got)
	}
	if got := countersOn(g, oppOrc, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's Orc got %d +1/+1 counters, want 1 (\"you control\")", got)
	}

	// An amass makes the Army first, so its first counter is bumped too.
	castCatalogSpell(t, g, "Orcish Medicine", "Instant", orcishMedicineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("an amass of 1 with Mauhúr out put %d counters on the new Army, want 2", got)
	}
}

// --- Warg Rider -------------------------------------------------------

func TestWargRiderGivesOtherOrcsAndGoblinsMenace(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rider := pushCatalogPermanent(g, me.ID, "Warg Rider", "Creature — Orc Warrior", wargRiderOracle, false)
	orc := pushDiesCreatureForTest(g, me.ID, "Orc", "", sauronTestOrcTypeLine, 2, 2)
	goblin := pushDiesCreatureForTest(g, me.ID, "Goblin", "", "Creature — Goblin", 1, 1)
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	oppOrc := pushDiesCreatureForTest(g, g.Seats[1].ID, "Opposing Orc", "", sauronTestOrcTypeLine, 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	has := func(id uuid.UUID) bool {
		c := findBattlefieldCardByID(g, id)
		return c != nil && game.HasKeyword(c, "menace")
	}
	if !has(rider) {
		t.Error("Warg Rider prints menace")
	}
	if !has(orc) || !has(goblin) {
		t.Errorf("other Orcs and Goblins you control have menace (orc=%v goblin=%v)", has(orc), has(goblin))
	}
	if has(bear) || has(oppOrc) {
		t.Errorf("a Bear or an opposing Orc got menace (bear=%v oppOrc=%v)", has(bear), has(oppOrc))
	}
}

func TestWargRiderAmassesOrcsAtTheBeginningOfYourCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Warg Rider", "Creature — Orc Warrior", wargRiderOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)
	if !army.HasSubtype("Orc") {
		t.Errorf("the Army is %q, want an Orc Army", army.TypeLine)
	}
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
	if !game.HasKeyword(army, "menace") {
		t.Error("the Orc Army has menace from Warg Rider")
	}
}

// --- Saruman ----------------------------------------------------------

func TestSarumanAmassesByTheManaValueOfANoncreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Saruman, the White Hand", "Legendary Creature — Avatar Wizard", sarumanOracle, false)

	// A creature spell does nothing.
	id := putInHand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Bear: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && game.IsArmy(c) {
			t.Fatal("a creature spell amassed")
		}
	}

	// A noncreature spell amasses its mana value.
	sorc := putInHand(me, game.Card{Name: "Divination-ish", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	if err := g.CastSpell(me.ID, sorc, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 3 {
		t.Errorf("Army counters = %d, want 3 (the spell's mana value)", got)
	}
	if !army.HasSubtype("Orc") {
		t.Errorf("the Army is %q, want an Orc Army", army.TypeLine)
	}
}

func TestSarumanGivesGoblinsAndOrcsWardTwoButNotOtherCreatures(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		warded   bool
	}{
		{"an Orc", sauronTestOrcTypeLine, true},
		{"a Goblin", "Creature — Goblin", true},
		{"a Bear", "Creature — Bear", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushCatalogPermanent(g, me.ID, "Saruman, the White Hand", "Legendary Creature — Avatar Wizard", sarumanOracle, false)
			victim := pushDiesCreatureForTest(g, me.ID, "Victim", "", tc.typeLine, 2, 2)
			for g.Turn.Step != game.StepPrecombatMain {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			castAtWardedCreature(t, g, opp, victim)
			for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID) && !stackFullyEmpty(g); i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if got := hasPayUnlessFor(g, opp.ID); got != tc.warded {
				t.Errorf("ward prompt = %v, want %v", got, tc.warded)
			}
		})
	}
}

// --- Corsairs of Umbar ------------------------------------------------

func TestCorsairsOfUmbarAmassesThreeOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	corsairs := pushCatalogPermanent(g, me.ID, "Corsairs of Umbar", "Creature — Human Pirate", corsairsOracle, false)
	if c := findBattlefieldCardByID(g, corsairs); c != nil {
		c.Power, c.Toughness, c.PrintedPTKnown = 3, 3, true
	}
	attackWith(t, g, opp.ID, corsairs)
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 3 {
		t.Errorf("Army counters = %d, want 3", got)
	}
}

func TestCorsairsOfUmbarMakesAGoblinOrcOrPirateUnblockable(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	corsairs := pushCatalogPermanent(g, me.ID, "Corsairs of Umbar", "Creature — Human Pirate", corsairsOracle, false)
	orc := pushDiesCreatureForTest(g, me.ID, "Orc", "", sauronTestOrcTypeLine, 2, 2)
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	toMain(t, g)
	fillPoolColored(me, "U", 3)

	if err := g.ActivateCatalogAbility(me.ID, corsairs, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("a Bear is not a Goblin, Orc or Pirate, so it is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, corsairs, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: orc}},
	}); err != nil {
		t.Fatalf("activate on an Orc: %v", err)
	}
	passPriorityAroundTable(t, g)
	if auraRestrictions(t, g, orc)&game.CantBeBlocked == 0 {
		t.Error("the Orc can't be blocked this turn")
	}
}

// --- Orcish Medicine --------------------------------------------------

func TestOrcishMedicineGrantsTheChosenKeywordThenAmasses(t *testing.T) {
	for _, tc := range []struct {
		name    string
		index   int
		keyword string
	}{
		{"lifelink", 0, "lifelink"},
		{"indestructible", 1, "indestructible"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
			castCatalogSpell(t, g, "Orcish Medicine", "Instant", orcishMedicineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			passPriorityAroundTable(t, g)
			answerOptionPick(t, g, me.ID, tc.index)
			passPriorityAroundTable(t, g)

			c := findBattlefieldCardByID(g, bear)
			if c == nil || !game.HasKeyword(c, tc.keyword) {
				t.Errorf("the Bear should have %s until end of turn", tc.keyword)
			}
			army := armyOf(t, g, me.ID)
			if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
				t.Errorf("Army counters = %d, want 1", got)
			}
		})
	}
}

func TestOrcishMedicineFizzlesWithoutItsTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Orcish Medicine", "Instant", orcishMedicineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	exileCreature(t, g, bear)
	passPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && game.IsArmy(c) {
			t.Fatal("a spell with no legal target does nothing, so there is no amass")
		}
	}
}

// --- Foray of Orcs ----------------------------------------------------

func TestForayOfOrcsDamagesByTheAmassedArmysPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	// An existing Army with 3 counters becomes a 5/5 with the amass.
	existing := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Orc Army", TypeLine: "Token Creature — Orc Army",
		PrintedPTKnown: true, Counters: map[string]int{game.CounterPlusOne: 3},
		Owner: me.ID, Controller: me.ID,
	})
	victim := pushDiesCreatureForTest(g, opp.ID, "Big Bear", "", "Creature — Bear", 5, 5)
	mine := pushDiesCreatureForTest(g, me.ID, "Mine", "", "Creature — Bear", 1, 1)

	castCatalogSpell(t, g, "Foray of Orcs", "Sorcery", forayOfOrcsOracle, nil)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		// Only an opponent's creature is a legal target.
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("Foray of Orcs may only target a creature an opponent controls")
		}
		pickCard(t, g, me.ID, victim)
	} else {
		t.Fatal("the reflexive trigger should ask for its target")
	}
	passPriorityAroundTable(t, g)

	if got := countersOn(g, existing, game.CounterPlusOne); got != 5 {
		t.Fatalf("Army counters = %d, want 5", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("5 damage from a 5-power Army should have destroyed the 5/5")
	}
}

func TestForayOfOrcsWithNoOpposingCreatureStillAmasses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Foray of Orcs", "Sorcery", forayOfOrcsOracle, nil)
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
}

// --- Lazotep Plating --------------------------------------------------

func TestLazotepPlatingAmassesZombiesAndGivesHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Lazotep Plating", "Instant", lazotepPlatingOracle, nil)
	passPriorityAroundTable(t, g)

	army := armyOf(t, g, me.ID)
	if !army.HasSubtype("Zombie") {
		t.Errorf("the Army is %q, want a Zombie Army", army.TypeLine)
	}
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("Army counters = %d, want 1", got)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, id := range []uuid.UUID{bear, army.InstanceID} {
		c := findBattlefieldCardByID(g, id)
		if c == nil || !game.HasKeyword(c, "hexproof") {
			t.Errorf("%v should have hexproof until end of turn", id)
		}
	}
}

// --- Fall of Cair Andros ---------------------------------------------

func TestFallOfCairAndrosAmassesTheExcessNoncombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	fall := pushCatalogPermanent(g, me.ID, "Fall of Cair Andros", "Enchantment", fallOfCairOracle, false)
	victim := pushDiesCreatureForTest(g, opp.ID, "Squishy", "", "Creature — Bear", 2, 3)
	toMain(t, g)
	fillPoolColored(me, "R", 8)

	if err := g.ActivateCatalogAbility(me.ID, fall, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 4 {
		t.Errorf("Army counters = %d, want 4 (7 damage to a 3-toughness creature)", got)
	}
	if !army.HasSubtype("Orc") {
		t.Errorf("the Army is %q, want an Orc Army", army.TypeLine)
	}
}

func TestFallOfCairAndrosIgnoresNoExcessCombatAndYourOwnCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	fall := pushCatalogPermanent(g, me.ID, "Fall of Cair Andros", "Enchantment", fallOfCairOracle, false)
	wall := pushDiesCreatureForTest(g, opp.ID, "Wall", "", "Creature — Wall", 0, 10)
	mine := pushDiesCreatureForTest(g, me.ID, "Mine", "", "Creature — Bear", 2, 3)
	toMain(t, g)
	fillPoolColored(me, "R", 16)

	for _, target := range []uuid.UUID{wall, mine} {
		if err := g.ActivateCatalogAbility(me.ID, fall, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && game.IsArmy(c) {
			t.Error("no excess damage was dealt to an opponent's creature, so there is no Army")
		}
	}
}

// --- Barad-dûr --------------------------------------------------------

func TestBaradDurEntersTappedUnlessYouControlALegendaryCreature(t *testing.T) {
	for _, tc := range []struct {
		name      string
		legendary bool
		tapped    bool
	}{
		{"without a legend", false, true},
		{"with a legend", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			if tc.legendary {
				pushDiesCreatureForTest(g, me.ID, "Legend", "", "Legendary Creature — Human", 2, 2)
			}
			id := b12PlayFromHand(t, g, "Barad-dûr", "Legendary Land", baradDurOracle, game.CastSpellParams{})
			if got := b12Card(t, g, id).Tapped; got != tc.tapped {
				t.Errorf("tapped = %v, want %v", got, tc.tapped)
			}
		})
	}
}

func TestBaradDurAmassesXOnlyIfACreatureDiedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Barad-dûr", "Legendary Land", baradDurOracle, false)
	toMain(t, g)
	fillPoolColored(me, "B", 8)

	err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{XValue: 2})
	if err == nil {
		t.Fatal("the ability can't be activated before a creature has died this turn")
	}

	victim := pushDiesCreatureForTest(g, me.ID, "Victim", "", "Creature — Bear", 1, 1)
	killCreature(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate after a creature died: %v", err)
	}
	if got := len(me.ManaPool); got != 8-5 {
		t.Errorf("pool = %d after paying {X}{X}{B} at X=2, want 3", got)
	}
	passPriorityAroundTable(t, g)
	army := armyOf(t, g, me.ID)
	if got := countersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
}
