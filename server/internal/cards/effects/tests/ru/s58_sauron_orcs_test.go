package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Mauhúr -----------------------------------------------------------

func TestMauhurAddsACounterToAmassedArmiesAndOrcsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Mauhúr, Uruk-hai Captain", "Legendary Creature — Orc Soldier", effectstest.MauhurOracle, false)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	orc := effectstest.PushDiesCreatureForTest(g, me.ID, "Orc", "", effectstest.SauronTestOrcTypeLine, 2, 2)
	oppOrc := effectstest.PushDiesCreatureForTest(g, g.Seats[1].ID, "Opposing Orc", "", effectstest.SauronTestOrcTypeLine, 2, 2)

	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{bear, orc, oppOrc} {
			if err := g.AddCounterByForEffect(me.ID, id, game.CounterPlusOne, 1); err != nil {
				t.Fatalf("AddCounterForEffect: %v", err)
			}
		}
	})
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("a Bear got %d +1/+1 counters, want 1 (Mauhúr only helps Armies, Goblins and Orcs)", got)
	}
	if got := effectstest.CountersOn(g, orc, game.CounterPlusOne); got != 2 {
		t.Errorf("an Orc got %d +1/+1 counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, oppOrc, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's Orc got %d +1/+1 counters, want 1 (\"you control\")", got)
	}

	// An amass makes the Army first, so its first counter is bumped too.
	effectstest.CastCatalogSpell(t, g, "Orcish Medicine", "Instant", effectstest.OrcishMedicineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("an amass of 1 with Mauhúr out put %d counters on the new Army, want 2", got)
	}
}

// --- Warg Rider -------------------------------------------------------

func TestWargRiderGivesOtherOrcsAndGoblinsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rider := effectstest.PushCatalogPermanent(g, me.ID, "Warg Rider", "Creature — Orc Warrior", effectstest.WargRiderOracle, false)
	orc := effectstest.PushDiesCreatureForTest(g, me.ID, "Orc", "", effectstest.SauronTestOrcTypeLine, 2, 2)
	goblin := effectstest.PushDiesCreatureForTest(g, me.ID, "Goblin", "", "Creature — Goblin", 1, 1)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	oppOrc := effectstest.PushDiesCreatureForTest(g, g.Seats[1].ID, "Opposing Orc", "", effectstest.SauronTestOrcTypeLine, 2, 2)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })

	has := func(id uuid.UUID) bool {
		c := effectstest.FindBattlefieldCardByID(g, id)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Warg Rider", "Creature — Orc Warrior", effectstest.WargRiderOracle, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if !army.HasSubtype("Orc") {
		t.Errorf("the Army is %q, want an Orc Army", army.TypeLine)
	}
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
	if !game.HasKeyword(army, "menace") {
		t.Error("the Orc Army has menace from Warg Rider")
	}
}

// --- Saruman ----------------------------------------------------------

func TestSarumanAmassesByTheManaValueOfANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Saruman, the White Hand", "Legendary Creature — Avatar Wizard", effectstest.SarumanOracle, false)

	// A creature spell does nothing.
	id := effectstest.PutInHand(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	effectstest.ToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && game.IsArmy(c) {
			t.Fatal("a creature spell amassed")
		}
	}

	// A noncreature spell amasses its mana value.
	sorc := effectstest.PutInHand(me, game.Card{Name: "Divination-ish", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	if err := g.CastSpell(me.ID, sorc, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell sorcery: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 3 {
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
		{"an Orc", effectstest.SauronTestOrcTypeLine, true},
		{"a Goblin", "Creature — Goblin", true},
		{"a Bear", "Creature — Bear", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, me.ID, "Saruman, the White Hand", "Legendary Creature — Avatar Wizard", effectstest.SarumanOracle, false)
			victim := effectstest.PushDiesCreatureForTest(g, me.ID, "Victim", "", tc.typeLine, 2, 2)
			for g.Turn.Step != game.StepPrecombatMain {
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
			effectstest.CastAtWardedCreature(t, g, opp, victim)
			for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID) && !effectstest.StackFullyEmpty(g); i++ {
				if err := g.PassPriority(); err != nil {
					t.Fatalf("PassPriority: %v", err)
				}
			}
			if got := effectstest.HasPayUnlessFor(g, opp.ID); got != tc.warded {
				t.Errorf("ward prompt = %v, want %v", got, tc.warded)
			}
		})
	}
}

// --- Corsairs of Umbar ------------------------------------------------

func TestCorsairsOfUmbarAmassesThreeOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	corsairs := effectstest.PushCatalogPermanent(g, me.ID, "Corsairs of Umbar", "Creature — Human Pirate", effectstest.CorsairsOracle, false)
	if c := effectstest.FindBattlefieldCardByID(g, corsairs); c != nil {
		c.Power, c.Toughness, c.PrintedPTKnown = 3, 3, true
	}
	effectstest.AttackWith(t, g, opp.ID, corsairs)
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 3 {
		t.Errorf("Army counters = %d, want 3", got)
	}
}

func TestCorsairsOfUmbarMakesAGoblinOrcOrPirateUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	corsairs := effectstest.PushCatalogPermanent(g, me.ID, "Corsairs of Umbar", "Creature — Human Pirate", effectstest.CorsairsOracle, false)
	orc := effectstest.PushDiesCreatureForTest(g, me.ID, "Orc", "", effectstest.SauronTestOrcTypeLine, 2, 2)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, g)
	effectstest.FillPoolColored(me, "U", 3)

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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraRestrictions(t, g, orc)&game.CantBeBlocked == 0 {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
			effectstest.CastCatalogSpell(t, g, "Orcish Medicine", "Instant", effectstest.OrcishMedicineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerOptionPick(t, g, me.ID, tc.index)
			effectstest.PassPriorityAroundTable(t, g)

			c := effectstest.FindBattlefieldCardByID(g, bear)
			if c == nil || !game.HasKeyword(c, tc.keyword) {
				t.Errorf("the Bear should have %s until end of turn", tc.keyword)
			}
			army := effectstest.ArmyOf(t, g, me.ID)
			if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
				t.Errorf("Army counters = %d, want 1", got)
			}
		})
	}
}

func TestOrcishMedicineFizzlesWithoutItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Orcish Medicine", "Instant", effectstest.OrcishMedicineOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.ExileCreature(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && game.IsArmy(c) {
			t.Fatal("a spell with no legal target does nothing, so there is no amass")
		}
	}
}

// --- Foray of Orcs ----------------------------------------------------

func TestForayOfOrcsDamagesByTheAmassedArmysPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	// An existing Army with 3 counters becomes a 5/5 with the amass.
	existing := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Orc Army", TypeLine: "Token Creature — Orc Army",
		PrintedPTKnown: true, Counters: map[string]int{game.CounterPlusOne: 3},
		Owner: me.ID, Controller: me.ID,
	})
	victim := effectstest.PushDiesCreatureForTest(g, opp.ID, "Big Bear", "", "Creature — Bear", 5, 5)
	mine := effectstest.PushDiesCreatureForTest(g, me.ID, "Mine", "", "Creature — Bear", 1, 1)

	effectstest.CastCatalogSpell(t, g, "Foray of Orcs", "Sorcery", effectstest.ForayOfOrcsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		// Only an opponent's creature is a legal target.
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("Foray of Orcs may only target a creature an opponent controls")
		}
		effectstest.PickCard(t, g, me.ID, victim)
	} else {
		t.Fatal("the reflexive trigger should ask for its target")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountersOn(g, existing, game.CounterPlusOne); got != 5 {
		t.Fatalf("Army counters = %d, want 5", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("5 damage from a 5-power Army should have destroyed the 5/5")
	}
}

func TestForayOfOrcsWithNoOpposingCreatureStillAmasses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Foray of Orcs", "Sorcery", effectstest.ForayOfOrcsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
}

// --- Lazotep Plating --------------------------------------------------

func TestLazotepPlatingAmassesZombiesAndGivesHexproof(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lazotep Plating", "Instant", effectstest.LazotepPlatingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	army := effectstest.ArmyOf(t, g, me.ID)
	if !army.HasSubtype("Zombie") {
		t.Errorf("the Army is %q, want a Zombie Army", army.TypeLine)
	}
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("Army counters = %d, want 1", got)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, id := range []uuid.UUID{bear, army.InstanceID} {
		c := effectstest.FindBattlefieldCardByID(g, id)
		if c == nil || !game.HasKeyword(c, "hexproof") {
			t.Errorf("%v should have hexproof until end of turn", id)
		}
	}
}

// --- Fall of Cair Andros ---------------------------------------------

func TestFallOfCairAndrosAmassesTheExcessNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	fall := effectstest.PushCatalogPermanent(g, me.ID, "Fall of Cair Andros", "Enchantment", effectstest.FallOfCairOracle, false)
	victim := effectstest.PushDiesCreatureForTest(g, opp.ID, "Squishy", "", "Creature — Bear", 2, 3)
	effectstest.ToMain(t, g)
	effectstest.FillPoolColored(me, "R", 8)

	if err := g.ActivateCatalogAbility(me.ID, fall, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 4 {
		t.Errorf("Army counters = %d, want 4 (7 damage to a 3-toughness creature)", got)
	}
	if !army.HasSubtype("Orc") {
		t.Errorf("the Army is %q, want an Orc Army", army.TypeLine)
	}
}

func TestFallOfCairAndrosIgnoresNoExcessCombatAndYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	fall := effectstest.PushCatalogPermanent(g, me.ID, "Fall of Cair Andros", "Enchantment", effectstest.FallOfCairOracle, false)
	wall := effectstest.PushDiesCreatureForTest(g, opp.ID, "Wall", "", "Creature — Wall", 0, 10)
	mine := effectstest.PushDiesCreatureForTest(g, me.ID, "Mine", "", "Creature — Bear", 2, 3)
	effectstest.ToMain(t, g)
	effectstest.FillPoolColored(me, "R", 16)

	for _, target := range []uuid.UUID{wall, mine} {
		if err := g.ActivateCatalogAbility(me.ID, fall, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			if tc.legendary {
				effectstest.PushDiesCreatureForTest(g, me.ID, "Legend", "", "Legendary Creature — Human", 2, 2)
			}
			id := effectstest.B12PlayFromHand(t, g, "Barad-dûr", "Legendary Land", effectstest.BaradDurOracle, game.CastSpellParams{})
			if got := effectstest.B12Card(t, g, id).Tapped; got != tc.tapped {
				t.Errorf("tapped = %v, want %v", got, tc.tapped)
			}
		})
	}
}

func TestBaradDurAmassesXOnlyIfACreatureDiedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Barad-dûr", "Legendary Land", effectstest.BaradDurOracle, false)
	effectstest.ToMain(t, g)
	effectstest.FillPoolColored(me, "B", 8)

	err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{XValue: 2})
	if err == nil {
		t.Fatal("the ability can't be activated before a creature has died this turn")
	}

	victim := effectstest.PushDiesCreatureForTest(g, me.ID, "Victim", "", "Creature — Bear", 1, 1)
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{XValue: 2}); err != nil {
		t.Fatalf("activate after a creature died: %v", err)
	}
	if got := me.ManaPool.Total(); got != 8-5 {
		t.Errorf("pool = %d after paying {X}{X}{B} at X=2, want 3", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	army := effectstest.ArmyOf(t, g, me.ID)
	if got := effectstest.CountersOn(g, army.InstanceID, game.CounterPlusOne); got != 2 {
		t.Errorf("Army counters = %d, want 2", got)
	}
}
