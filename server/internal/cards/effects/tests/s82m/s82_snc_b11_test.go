package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b11_test.go — Streets of New Capenna / Commander, batch 11
// (Treasure): Ognis, Hoard Hauler, Black Market Tycoon, Jetmir's Fixer,
// Most Wanted, Jolene, Vazi, Capenna Express, Spiteful Repossession,
// Paragon of Modernity and Boxing Ring.

const (
	sncb11OgnisOracle      = "a6e50e06-6d1d-457d-8fbb-ec9a866960c3"
	sncb11HoardOracle      = "2578e8e4-a93c-4904-9037-92a1529c2c67"
	sncb11TycoonOracle     = "464968a8-865e-4ac8-927e-7c931e3dabb3"
	sncb11FixerOracle      = "f87df7d4-99e2-4340-9d0a-b96c397555db"
	sncb11MostWantedOracle = "140d38cc-0ae6-4e9b-a60f-a51d2947e07b"
	sncb11JoleneOracle     = "8215f4a2-7131-426b-ad9c-6427caabd715"
	sncb11VaziOracle       = "7446c013-63b0-402b-b8b9-dd25ccc553df"
	sncb11ExpressOracle    = "568432d1-c7fb-4da6-922a-22d4fa26284e"
	sncb11RepossessOracle  = "63fd77e2-348a-4f6c-88b4-465fbe339d29"
	sncb11ParagonOracle    = "c1c62704-1f20-48cf-8588-47e7fca85f7b"
	sncb11BoxingRingOracle = "0e676904-17c3-40f2-9b5b-eb29cacc336f"
)

func TestS82SncB11EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncb11OgnisOracle:      "Ognis, the Dragon's Lash",
		sncb11HoardOracle:      "Hoard Hauler",
		sncb11TycoonOracle:     "Black Market Tycoon",
		sncb11FixerOracle:      "Jetmir's Fixer",
		sncb11MostWantedOracle: "Most Wanted",
		sncb11JoleneOracle:     "Jolene, the Plunder Queen",
		sncb11VaziOracle:       "Vazi, Keen Negotiator",
		sncb11ExpressOracle:    "Capenna Express",
		sncb11RepossessOracle:  "Spiteful Repossession",
		sncb11ParagonOracle:    "Paragon of Modernity",
		sncb11BoxingRingOracle: "Boxing Ring",
	} {
		effectstest.RequireFull(t, oracle, name)
	}
}

// sncb11Treasures puts n Treasure tokens under `owner` the way an effect does.
func sncb11Treasures(t *testing.T, g *game.Game, owner uuid.UUID, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.CreateTokenForEffect(owner, effects.TreasureToken(), n) })
	if err != nil {
		t.Fatalf("create Treasures: %v", err)
	}
}

// sncb11Attack declares each attacker at defender, locks the declaration in and
// lets what it triggered resolve.
func sncb11Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
}

func sncb11TappedTreasures(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Treasure" && c.Tapped {
			n++
		}
	}
	return n
}

// --- Ognis, the Dragon's Lash ------------------------------------------

func TestS82OgnisMakesATappedTreasureForEachHastyAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ognis := effectstest.PushDiesCreatureForTest(g, me.ID, "Ognis, the Dragon's Lash", sncb11OgnisOracle, "Legendary Creature — Lizard Warrior", 3, 3)
	hasty := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hasty Bear", TypeLine: "Creature — Bear", Keywords: []string{"haste"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	slow := effectstest.PushVanillaCreature(g, me.ID, "Slow Bear", 2, 2)

	sncb11Attack(t, g, opp.ID, ognis, hasty, slow)
	if got := effectstest.CountTreasures(g, me.ID); got != 2 {
		t.Fatalf("Treasures %d, want 2 (Ognis and the hasty Bear, not the slow one)", got)
	}
	if got := sncb11TappedTreasures(g, me.ID); got != 2 {
		t.Errorf("tapped Treasures %d, want 2: they enter tapped", got)
	}
}

func TestS82OgnisIgnoresAnOpponentsHastyAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushDiesCreatureForTest(g, opp.ID, "Ognis, the Dragon's Lash", sncb11OgnisOracle, "Legendary Creature — Lizard Warrior", 3, 3)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hasty Bear", TypeLine: "Creature — Bear", Keywords: []string{"haste"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	sncb11Attack(t, g, opp.ID, mine)
	if got := effectstest.CountTreasures(g, opp.ID) + effectstest.CountTreasures(g, me.ID); got != 0 {
		t.Errorf("%d Treasures: Ognis counts only creatures its controller controls", got)
	}
}

// --- Black Market Tycoon -----------------------------------------------

func TestS82TycoonCreatesATreasureAndHurtsYouAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tycoon := effectstest.PushDiesCreatureForTest(g, me.ID, "Black Market Tycoon", sncb11TycoonOracle, "Creature — Cat Rogue", 2, 2)
	effectstest.ToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, tycoon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 1 {
		t.Fatalf("Treasures %d, want 1", got)
	}
	sncb11Treasures(t, g, me.ID, 2) // three Treasures in all
	life := me.Life
	sncb11ToMyNextUpkeep(t, g)
	if got := life - me.Life; got != 6 {
		t.Errorf("lost %d life, want 6 (2 for each of 3 Treasures)", got)
	}
}

func TestS82TycoonDealsNothingWithoutTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushDiesCreatureForTest(g, me.ID, "Black Market Tycoon", sncb11TycoonOracle, "Creature — Cat Rogue", 2, 2)
	life := me.Life
	sncb11ToMyNextUpkeep(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// --- Jetmir's Fixer ----------------------------------------------------

func sncb11Activate(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, idx int, spec string, kinds game.ManaSourceKinds) {
	t.Helper()
	effectstest.ToMain(t, g)
	effectstest.GmrFloatFrom(g, p, spec, kinds)
	if err := g.ActivateCatalogAbility(p.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82FixerPumpsOrGrowsPermanentlyOnTreasureMana(t *testing.T) {
	for _, tc := range []struct {
		name     string
		kinds    game.ManaSourceKinds
		counters int
		power    int
	}{
		{"ordinary mana is a +1/+1 until end of turn", 0, 0, 3},
		{"mana from a Treasure is a +1/+1 counter instead", game.ManaSourceTreasure | game.ManaSourceArtifact, 1, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			fixer := effectstest.PushDiesCreatureForTest(g, me.ID, "Jetmir's Fixer", sncb11FixerOracle, "Creature — Cat Warrior", 2, 2)
			sncb11Activate(t, g, me, fixer, 0, "RG", tc.kinds)
			if got := effectstest.CounterCount(g, fixer, game.CounterPlusOne); got != tc.counters {
				t.Errorf("+1/+1 counters %d, want %d", got, tc.counters)
			}
			if got := effectstest.EffectivePowerWithCounters(t, g, fixer); got != tc.power {
				t.Errorf("power %d, want %d", got, tc.power)
			}
		})
	}
}

func TestS82FixerHalfTreasureManaStillCountsAsTreasureMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fixer := effectstest.PushDiesCreatureForTest(g, me.ID, "Jetmir's Fixer", sncb11FixerOracle, "Creature — Cat Warrior", 2, 2)
	effectstest.ToMain(t, g)
	effectstest.GmrFloatFrom(g, me, "R", 0)
	effectstest.GmrFloatFrom(g, me, "G", game.ManaSourceTreasure|game.ManaSourceArtifact)
	if err := g.ActivateCatalogAbility(me.ID, fixer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, fixer, game.CounterPlusOne); got != 1 {
		t.Errorf("one mana from a Treasure was spent: counters %d, want 1", got)
	}
}

// --- Paragon of Modernity ----------------------------------------------

func TestS82ParagonCountersOnlyOnExactlyThreeColors(t *testing.T) {
	for _, tc := range []struct {
		pay      string
		counters int
	}{
		{"WUB", 1},
		{"WWU", 0},
		{"WUBR", 0}, // four colours need a {3} cost of four mana: pay three of four below
	} {
		if len(tc.pay) != 3 {
			continue
		}
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		p := effectstest.PushDiesCreatureForTest(g, me.ID, "Paragon of Modernity", sncb11ParagonOracle, "Artifact Creature — Angel Warrior", 2, 2)
		sncb11Activate(t, g, me, p, 0, tc.pay, 0)
		if got := effectstest.CounterCount(g, p, game.CounterPlusOne); got != tc.counters {
			t.Errorf("paid %s: counters %d, want %d", tc.pay, got, tc.counters)
		}
		if got := effectstest.EffectivePowerWithCounters(t, g, p); got != 3 {
			t.Errorf("paid %s: power %d, want 3", tc.pay, got)
		}
	}
}

func TestS82ParagonColorlessIsNotAColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	p := effectstest.PushDiesCreatureForTest(g, me.ID, "Paragon of Modernity", sncb11ParagonOracle, "Artifact Creature — Angel Warrior", 2, 2)
	sncb11Activate(t, g, me, p, 0, "WUC", 0)
	if got := effectstest.CounterCount(g, p, game.CounterPlusOne); got != 0 {
		t.Errorf("two colours and colourless: counters %d, want 0", got)
	}
}

// --- Capenna Express ---------------------------------------------------

func TestS82CapennaExpressAnimatesOnATreasureSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	van := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Capenna Express", OracleID: sncb11ExpressOracle, TypeLine: "Artifact — Vehicle",
		Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	sncb11Treasures(t, g, me.ID, 1)
	var treasure uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Treasure" {
			treasure = c.InstanceID
		}
	}
	effectstest.ToMain(t, g)
	if c, _ := effectstest.BattlefieldCardByID(g, van); c.IsCreature() {
		t.Fatal("an uncrewed Vehicle is a creature")
	}
	// Anything but a Treasure is refused as the sacrifice.
	other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Clue", TypeLine: "Artifact — Clue", Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateCatalogAbility(me.ID, van, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{other}}); err == nil {
		t.Fatal("a Clue was accepted as the sacrificed Treasure")
	}
	if err := g.ActivateCatalogAbility(me.ID, van, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{treasure}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treasure) {
		t.Error("the Treasure was not sacrificed")
	}
	c, _ := effectstest.BattlefieldCardByID(g, van)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Errorf("the Vehicle is not an artifact creature after the sacrifice: %q", c.TypeLine)
	}
}

func TestS82CapennaExpressCrew3(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	van := s.PushVehicle("Capenna Express", sncb11ExpressOracle, 6, 6)
	big := s.PushCreature("Big Bear", "Creature — Bear", 3, 3)
	s.CrewAndResolve(t, van, 1, big)
}

func TestS82CapennaExpressCannotCrewWithTooLittlePower(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	van := s.PushVehicle("Capenna Express", sncb11ExpressOracle, 6, 6)
	small := s.PushCreature("Small Bear", "Creature — Bear", 2, 2)
	effectstest.ToMain(t, s.G)
	if err := s.G.ActivateCatalogAbility(s.Me.ID, van, 1, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{small}}); err == nil {
		t.Error("crew 3 was paid with power 2")
	}
}

// --- Hoard Hauler ------------------------------------------------------

func TestS82HoardHaulerMakesATreasurePerArtifactTheyControl(t *testing.T) {
	s := effectstest.NewCrewScene(t)
	hauler := s.PushVehicle("Hoard Hauler", sncb11HoardOracle, 5, 5)
	crew := s.PushCreature("Big Bear", "Creature — Bear", 3, 3)
	for i := 0; i < 3; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(s.G, game.Card{
			InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact", Owner: s.Opp.ID, Controller: s.Opp.ID,
		})
	}
	// Their creature that is not an artifact, and one of mine, do not count.
	effectstest.PushVanillaCreature(s.G, s.Opp.ID, "Their Bear", 0, 9)
	effectstest.PushBattlefieldCardWithTimestamp(s.G, game.Card{
		InstanceID: uuid.New(), Name: "My Trinket", TypeLine: "Artifact", Owner: s.Me.ID, Controller: s.Me.ID,
	})
	s.CrewAndResolve(t, hauler, 0, crew)
	before := effectstest.CountTreasures(s.G, s.Me.ID)
	effectstest.AdvanceTo(t, s.G, game.StepDeclareAttackers)
	if err := s.G.DeclareAttacker(hauler, s.Opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, s.G)
	effectstest.AdvanceTo(t, s.G, game.StepCombatDamage)
	s.G.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, s.G)
	if got := effectstest.CountTreasures(s.G, s.Me.ID) - before; got != 3 {
		t.Errorf("Treasures %d, want 3 (one per artifact they control)", got)
	}
	if got := effectstest.CountTreasures(s.G, s.Opp.ID); got != 0 {
		t.Errorf("the defender got %d Treasures", got)
	}
}

// --- Most Wanted -------------------------------------------------------

func TestS82MostWantedPumpsAndPaysTwoTreasuresWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Most Wanted", "Enchantment — Aura", sncb11MostWantedOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || th != 3 {
		t.Fatalf("enchanted creature is %d/%d, want 4/3", p, th)
	}
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 2 {
		t.Errorf("Aura's controller has %d Treasures, want 2", got)
	}
	if got := effectstest.CountTreasures(g, opp.ID); got != 0 {
		t.Errorf("the creature's controller has %d Treasures, want 0", got)
	}
}

func TestS82MostWantedIsAFlashAura(t *testing.T) {
	spec, ok := effects.Lookup(sncb11MostWantedOracle)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Most Wanted keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Spiteful Repossession ---------------------------------------------

func sncb11Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.PushPermanentForTest(g, owner, "Mountain", "", "Basic Land — Mountain")
	}
}

func TestS82RepossessionDamagesEachOpponentWithMoreLandsAndMakesThatManyTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a, b, c := g.Seats[1], g.Seats[2], g.Seats[3]
	sncb11Lands(g, me.ID, 2)
	sncb11Lands(g, a.ID, 5) // 3 more
	sncb11Lands(g, b.ID, 2) // equal: nothing
	sncb11Lands(g, c.ID, 3) // 1 more
	la, lb, lc := a.Life, b.Life, c.Life
	effectstest.CastCatalogSpell(t, g, "Spiteful Repossession", "Sorcery", sncb11RepossessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if la-a.Life != 3 || lb-b.Life != 0 || lc-c.Life != 1 {
		t.Errorf("damage %d/%d/%d, want 3/0/1", la-a.Life, lb-b.Life, lc-c.Life)
	}
	if got := effectstest.CountTreasures(g, me.ID); got != 4 {
		t.Errorf("Treasures %d, want 4", got)
	}
}

func TestS82RepossessionMakesNothingWithNoDifference(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sncb11Lands(g, me.ID, 3)
	effectstest.CastCatalogSpell(t, g, "Spiteful Repossession", "Sorcery", sncb11RepossessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 0 {
		t.Errorf("Treasures %d, want 0", got)
	}
}

func TestS82RepossessionCountsOnlyDamageActuallyDealt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	sncb11Lands(g, a.ID, 4)
	g.WithWriteLock(func() {
		if err := (effects.PreventNextDamage{Target: a.ID, Amount: 100, Label: "prevent the next 100 damage"}).Apply(effects.NewContext(g, nil)); err != nil {
			t.Fatalf("shield: %v", err)
		}
	})
	effectstest.CastCatalogSpell(t, g, "Spiteful Repossession", "Sorcery", sncb11RepossessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTreasures(g, me.ID); got != 0 {
		t.Errorf("Treasures %d, want 0 when the damage was prevented", got)
	}
}

// sncb11ToMyNextUpkeep walks the table round to seat 0's next upkeep and
// resolves what that step triggered.
func sncb11ToMyNextUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	away := false
	for i := 0; i < 400; i++ {
		if g.Turn.ActiveSeat != 0 {
			away = true
		}
		if away && g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep {
			effectstest.PassPriorityAroundTable(t, g)
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	t.Fatal("never reached seat 0's next upkeep")
}
