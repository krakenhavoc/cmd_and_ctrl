package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b167_cards_test.go — War of the Spark: +1/+1 and other counters, S82
// older sets, batch 167 (#5520). Each test plays the real catalog entry.

const (
	oldB167Company   = "083dcc19-3357-48e4-94e8-8da5bbfba210"
	oldB167Hydra     = "1f79de2a-6cef-4285-b008-5b5527c406ab"
	oldB167Awakening = "5c326102-005d-463c-b9f7-14b034b66baa"
	oldB167Stray     = "6781166b-f6ab-4076-8e97-0cd9342f1b41"
	oldB167BattleCry = "7528f47d-c758-42bc-bdfe-5aca805c4126"
	oldB167Hellion   = "8f6620c3-2c15-467c-92ca-fdf9042a63f7"
	oldB167Stealth   = "92419fbf-7a1b-428f-92d6-e9fd510dbfac"
	oldB167Ambush    = "a481f910-59a8-48b0-a8e4-c24675e1a67d"
	oldB167Pledge    = "ad501bcb-b838-4716-9759-98c484772285"
	oldB167Butcher   = "d254514b-0688-4f05-9473-3090a2ce7ced"
	oldB167Elderspel = "eff0a1ad-ecc6-416f-a4ae-96b26cd8a905"
	oldB167Twist     = "f5479d4f-bc4b-4452-a944-41be623f8367"
	oldB167Projectn  = "ff8db8aa-1d5a-42f8-8056-8da0da9846bc"
	oldB167Bear      = "Creature — Bear"
)

func TestOldB167EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB167Company:   "Gideon's Company",
		oldB167Hydra:     "Bioessence Hydra",
		oldB167Awakening: "Awakening of Vitu-Ghazi",
		oldB167Stray:     "Charmed Stray",
		oldB167BattleCry: "Gideon's Battle Cry",
		oldB167Hellion:   "Devouring Hellion",
		oldB167Stealth:   "Stealth Mission",
		oldB167Ambush:    "Domri's Ambush",
		oldB167Pledge:    "Pledge of Unity",
		oldB167Butcher:   "Dreadhorde Butcher",
		oldB167Elderspel: "The Elderspell",
		oldB167Twist:     "Teferi's Time Twist",
		oldB167Projectn:  "Jace's Projection",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB167Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[kind]
}

func oldB167Walker(g *game.Game, owner uuid.UUID, name, subtype string, loyalty int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Legendary Planeswalker — " + subtype,
		Counters: map[string]int{game.CounterLoyalty: loyalty}, Owner: owner, Controller: owner,
	})
}

// --- Charmed Stray ----------------------------------------------------

func TestOldB167CharmedStrayCountersOnlyTheOtherStrays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	first := effectstest.B12Push(g, me.ID, "Charmed Stray", "Creature — Cat", oldB167Stray, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other Cat", "Creature — Cat", 1, 1)
	theirs := effectstest.B12Push(g, opp.ID, "Charmed Stray", "Creature — Cat", oldB167Stray, 1, 1)
	third := effectstest.EnterFromHand(t, g, me.ID, "Charmed Stray", "Creature — Cat", oldB167Stray)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, first, game.CounterPlusOne); got != 1 {
		t.Errorf("the other Charmed Stray has %d counters, want 1", got)
	}
	if got := oldB167Counters(t, g, third, game.CounterPlusOne); got != 0 {
		t.Errorf("the entering Stray put a counter on itself: %d", got)
	}
	if oldB167Counters(t, g, other, game.CounterPlusOne) != 0 || oldB167Counters(t, g, theirs, game.CounterPlusOne) != 0 {
		t.Error("a creature with another name, or an opponent's Stray, got a counter")
	}
}

// --- Pledge of Unity and Gideon's Battle Cry ---------------------------

func TestOldB167PledgeOfUnityCountersEachCreatureAndGainsLifeForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", oldB167Bear, 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", oldB167Bear, 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB167Bear, 1, 1)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Pledge of Unity", "Instant", oldB167Pledge, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB167Counters(t, g, a, game.CounterPlusOne) != 1 || oldB167Counters(t, g, b, game.CounterPlusOne) != 1 {
		t.Error("my creatures did not each get a counter")
	}
	if oldB167Counters(t, g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's creature got a counter")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

func TestOldB167GideonsBattleCryCountersThenFetchesGideonFromTheGraveyardOrLibrary(t *testing.T) {
	for _, tc := range []struct {
		name  string
		where string
	}{{"library", "library"}, {"graveyard", "graveyard"}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bear := effectstest.B12Creature(g, me.ID, "Bear", oldB167Bear, 2, 2)
			var gideon uuid.UUID
			if tc.where == "library" {
				gideon = effectstest.SeedLibraryTop(me, "Gideon, the Oathsworn", "Legendary Planeswalker — Gideon", "{3}{W}{W}")
			} else {
				gideon = effectstest.PushGraveyardCardTyped(me, "Gideon, the Oathsworn", "Legendary Planeswalker — Gideon")
			}
			effectstest.CastCatalogSpell(t, g, "Gideon's Battle Cry", "Sorcery", oldB167BattleCry, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if oldB167Counters(t, g, bear, game.CounterPlusOne) != 1 {
				t.Error("the creature got no counter")
			}
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			effectstest.AnswerSearchByID(t, g, me.ID, gideon)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(gideon) {
				t.Errorf("Gideon, the Oathsworn did not reach my hand from the %s", tc.where)
			}
		})
	}
}

// --- Stealth Mission --------------------------------------------------

func TestOldB167StealthMissionPutsTwoCountersAndMakesItUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB167Bear, 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB167Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Stealth Mission", "Sorcery", oldB167Stealth, effectstest.CardTargets(attacker))
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, attacker, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters, want 2", got)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err == nil {
		t.Error("an unblockable creature was blocked")
	}
}

// --- Domri's Ambush ---------------------------------------------------

func TestOldB167DomrisAmbushCountersThenHitsACreatureOrPlaneswalkerForItsPower(t *testing.T) {
	for _, atWalker := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.B12Creature(g, me.ID, "Mine", oldB167Bear, 2, 2)
		var victim uuid.UUID
		if atWalker {
			victim = oldB167Walker(g, opp.ID, "Their Walker", "Test", 3)
		} else {
			victim = effectstest.B12Creature(g, opp.ID, "Victim", oldB167Bear, 3, 3)
		}
		effectstest.CastCatalogSpell(t, g, "Domri's Ambush", "Sorcery", oldB167Ambush, effectstest.CardTargets(mine, victim))
		effectstest.PassPriorityAroundTable(t, g)
		if oldB167Counters(t, g, mine, game.CounterPlusOne) != 1 {
			t.Fatal("no counter on my creature")
		}
		if effectstest.BattlefieldHas(g, victim) {
			t.Errorf("atWalker=%v: the 3-power hit (2 plus the counter) left the target alive", atWalker)
		}
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B12Creature(g, me.ID, "A", oldB167Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB167Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Domri's Ambush", "Sorcery", oldB167Ambush, effectstest.CardTargets(a, b)); err == nil {
		t.Error("the damage may be aimed at my own creature")
	}
}

// --- Dreadhorde Butcher -----------------------------------------------

func TestOldB167DreadhordeButcherGrowsOnCombatDamageAndDiesForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butcher := effectstest.B12Push(g, me.ID, "Dreadhorde Butcher", "Creature — Zombie Warrior", oldB167Butcher, 1, 1)
	effectstest.DealCombatDamageToPlayer(g, butcher, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamageToPlayer(g, butcher, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, butcher, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters, want 2", got)
	}
	opp.Life = 40
	effectstest.KillCreature(t, g, me.ID, butcher)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("opponent life = %d, want 37 (the Butcher's last power, 3)", opp.Life)
	}
}

func TestOldB167DreadhordeButcherIgnoresNoncombatAndCreatureDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butcher := effectstest.B12Push(g, me.ID, "Dreadhorde Butcher", "Creature — Zombie Warrior", oldB167Butcher, 1, 1)
	effectstest.DealNoncombatDamage(g, me.ID, butcher, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB167Counters(t, g, butcher, game.CounterPlusOne) != 0 {
		t.Error("noncombat damage grew the Butcher")
	}
}

// --- Devouring Hellion ------------------------------------------------

func TestOldB167DevouringHellionSacrificesCreaturesAndPlaneswalkersForTwiceThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB167Bear, 1, 1)
	walker := oldB167Walker(g, me.ID, "My Walker", "Test", 3)
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	id := effectstest.HandCardFull(me, "Devouring Hellion", "Creature — Hellion", "{2}{R}", oldB167Hellion, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EntryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	if effectstest.HasID(c.ChooseCards, land) {
		t.Error("a land was offered")
	}
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, fodder, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, id, game.CounterPlusOne); got != 4 {
		t.Errorf("the Hellion has %d counters, want 4 (twice two)", got)
	}
	if effectstest.BattlefieldHas(g, fodder) || effectstest.BattlefieldHas(g, walker) {
		t.Error("what was sacrificed is still on the battlefield")
	}
}

func TestOldB167DevouringHellionCanDeclineTheSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB167Bear, 1, 1)
	id := effectstest.HandCardFull(me, "Devouring Hellion", "Creature — Hellion", "{2}{R}", oldB167Hellion, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB167Counters(t, g, id, game.CounterPlusOne) != 0 || !effectstest.BattlefieldHas(g, fodder) {
		t.Error("declining still sacrificed or put counters on it")
	}
}

// --- Bioessence Hydra -------------------------------------------------

func TestOldB167BioessenceHydraEntersWithCountersAndGrowsWithLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := oldB167Walker(g, me.ID, "My Walker", "Test", 0)
	oldB167Walker(g, opp.ID, "Their Walker", "Other", 5)
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: mine, Kind: game.CounterLoyalty, N: 3}.Apply(effects.KeyedContext(g, me.ID, mine))
	})
	hydra := effectstest.EnterFromHand(t, g, me.ID, "Bioessence Hydra", "Creature — Hydra Mutant", oldB167Hydra)
	if got := oldB167Counters(t, g, hydra, game.CounterPlusOne); got != 3 {
		t.Fatalf("the Hydra entered with %d counters, want 3", got)
	}
	ctx := func(f func(c *effects.Context) error) {
		g.WithWriteLock(func() { _ = f(effects.KeyedContext(g, me.ID, mine)) })
	}
	ctx(func(c *effects.Context) error {
		return effects.AddCounter{Target: mine, Kind: game.CounterLoyalty, N: 2}.Apply(c)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, hydra, game.CounterPlusOne); got != 5 {
		t.Errorf("after two loyalty counters the Hydra has %d, want 5", got)
	}
}

func TestOldB167BioessenceHydraIgnoresAnOpponentsPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hydra := effectstest.B12Push(g, me.ID, "Bioessence Hydra", "Creature — Hydra Mutant", oldB167Hydra, 4, 4)
	theirs := oldB167Walker(g, opp.ID, "Their Walker", "Other", 5)
	g.WithWriteLock(func() {
		_ = effects.AddCounter{Target: theirs, Kind: game.CounterLoyalty, N: 2}.Apply(effects.KeyedContext(g, me.ID, hydra))
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, hydra, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's loyalty counters grew the Hydra by %d", got)
	}
}

// --- Gideon's Company and Jace's Projection ---------------------------

func TestOldB167GideonsCompanyGrowsOnLifeGainAndAddsLoyaltyToAGideon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	company := effectstest.B12Push(g, me.ID, "Gideon's Company", "Creature — Human Soldier", oldB167Company, 3, 3)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(company, me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, company, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters after gaining life, want 2", got)
	}
	gideon := oldB167Walker(g, g.Seats[1].ID, "Gideon Test", "Gideon", 4)
	jace := oldB167Walker(g, me.ID, "Jace Test", "Jace", 4)
	effectstest.B06AddMana(me, "W", "W", "W", "W", "W", "W", "W", "W")
	if err := g.ActivateCatalogAbility(me.ID, company, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(jace)}); err == nil {
		t.Error("a non-Gideon planeswalker was targeted")
	}
	effectstest.B16Activate(t, g, me.ID, company, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(gideon)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, gideon, game.CounterLoyalty); got != 5 {
		t.Errorf("Gideon has %d loyalty, want 5", got)
	}
}

func TestOldB167JacesProjectionGrowsOnEachDrawAndAddsLoyaltyToAJace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	proj := effectstest.B12Push(g, me.ID, "Jace's Projection", "Creature — Wizard Illusion", oldB167Projectn, 2, 2)
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, proj, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters after two draws, want 2", got)
	}
	jace := oldB167Walker(g, me.ID, "Jace Test", "Jace", 4)
	gideon := oldB167Walker(g, me.ID, "Gideon Test", "Gideon", 4)
	effectstest.B06AddMana(me, "U", "U", "U", "U", "U", "U", "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, proj, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(gideon)}); err == nil {
		t.Error("a non-Jace planeswalker was targeted")
	}
	effectstest.B16Activate(t, g, me.ID, proj, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(jace)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB167Counters(t, g, jace, game.CounterLoyalty); got != 5 {
		t.Errorf("Jace has %d loyalty, want 5", got)
	}
}

// --- Awakening of Vitu-Ghazi ------------------------------------------

func TestOldB167AwakeningOfVituGhaziMakesAHasteyNineNineLandCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Awakening of Vitu-Ghazi", "Instant", oldB167Awakening, effectstest.CardTargets(land))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, land)
	if c.Effective().Name != "Vitu-Ghazi" || !c.IsCreature() || !c.IsLand() || !c.IsLegendary() || !c.HasSubtype("Elemental") || !game.HasKeyword(&c, "haste") {
		t.Fatalf("land is %q creature=%v land=%v legendary=%v elemental=%v haste=%v", c.Effective().Name, c.IsCreature(), c.IsLand(), c.IsLegendary(), c.HasSubtype("Elemental"), game.HasKeyword(&c, "haste"))
	}
	if c.CurrentPower() != 9 || c.CurrentToughness() != 9 {
		t.Errorf("it is %d/%d, want 9/9 (nine counters on a 0/0)", c.CurrentPower(), c.CurrentToughness())
	}
	// It lasts: still a creature on a later turn.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if c := effectstest.LayeredCard(t, g, land); !c.IsCreature() {
		t.Error("it stopped being a creature at end of turn")
	}
}

func TestOldB167AwakeningOfVituGhaziOnlyTargetsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Permanent(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, g, "Awakening of Vitu-Ghazi", "Instant", oldB167Awakening, effectstest.CardTargets(theirs)); err == nil {
		t.Error("targeted an opponent's land")
	}
}

// --- The Elderspell ---------------------------------------------------

func TestOldB167TheElderspellDestroysWalkersAndAddsTwoLoyaltyEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := oldB167Walker(g, me.ID, "My Walker", "Test", 3)
	a := oldB167Walker(g, opp.ID, "Walker A", "A", 3)
	b := oldB167Walker(g, opp.ID, "Walker B", "B", 3)
	effectstest.CastCatalogSpell(t, g, "The Elderspell", "Sorcery", oldB167Elderspel, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); p != nil {
		effectstest.AnswerMayCounter(t, g, me.ID, mine)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, a) || effectstest.BattlefieldHas(g, b) {
		t.Error("a targeted planeswalker survived")
	}
	if got := oldB167Counters(t, g, mine, game.CounterLoyalty); got != 7 {
		t.Errorf("my walker has %d loyalty, want 7 (3 + 2 x 2)", got)
	}
}

func TestOldB167TheElderspellCountsOnlyWhatWasReallyDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := oldB167Walker(g, me.ID, "My Walker", "Test", 3)
	a := oldB167Walker(g, opp.ID, "Walker A", "A", 3)
	b := oldB167Walker(g, opp.ID, "Walker B", "B", 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == b {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "indestructible")
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "The Elderspell", "Sorcery", oldB167Elderspel, effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); p != nil {
		effectstest.AnswerMayCounter(t, g, me.ID, mine)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, b) {
		t.Fatal("the indestructible walker died")
	}
	if got := oldB167Counters(t, g, mine, game.CounterLoyalty); got != 5 {
		t.Errorf("my walker has %d loyalty, want 5 (one destroyed)", got)
	}
}

// --- Teferi's Time Twist ----------------------------------------------

func TestOldB167TeferisTimeTwistReturnsAtTheEndStepWithACounterOnlyAsACreature(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		want     int
	}{{"a creature", oldB167Bear, 1}, {"an artifact", "Artifact", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			perm := effectstest.B12Permanent(g, me.ID, "Subject", tc.typeLine)
			effectstest.CastCatalogSpell(t, g, "Teferi's Time Twist", "Instant", oldB167Twist, effectstest.CardTargets(perm))
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.BattlefieldHas(g, perm) || !g.Exile.Contains(perm) {
				t.Fatal("the permanent was not exiled")
			}
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			var back *game.Card
			g.ReadSnapshot(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].Name == "Subject" {
						c := g.Battlefield.Cards[i]
						back = &c
					}
				}
			})
			if back == nil {
				t.Fatal("it did not return at the end step")
			}
			if got := back.Counters[game.CounterPlusOne]; got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

func TestOldB167TeferisTimeTwistOnlyTargetsYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", oldB167Bear, 1, 1)
	if err := effectstest.CastCatalogSpellErr(t, g, "Teferi's Time Twist", "Instant", oldB167Twist, effectstest.CardTargets(theirs)); err == nil {
		t.Error("targeted an opponent's permanent")
	}
}
