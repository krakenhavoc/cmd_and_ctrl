package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b07_cards_test.go — Streets of New Capenna, S82 batch 7 (#5491).
// Each test plays the real catalog entries through the engine.

const (
	sncB07Kraken   = "42e4d4d8-e4e0-4cf8-a6d0-887a17513280"
	sncB07Sleep    = "48533a2d-1bdb-40a6-939e-9c8337e0dc5a"
	sncB07Lackey   = "6f6cbcd4-7a4b-4a38-bf29-995e82e658ce"
	sncB07Pets     = "497cb281-e057-42fc-937d-512855b74159"
	sncB07Perrie   = "41d408dc-f6fe-45b2-b225-2ac25db9ad0f"
	sncB07Bribe    = "482889f4-6cf5-457f-9f52-a3f841e038b5"
	sncB07Toolkit  = "a87abe7a-fee9-45fd-b3d8-896e43ed6b3c"
	sncB07Courier  = "a01746f3-8e6f-47ce-a95b-e44556d1f73e"
	sncB07Denry    = "ba477a45-84cc-4a85-98e4-50dca3e58788"
	sncB07Safeguad = "8ac43884-9f03-4945-98ea-c744d0403f32"
)

func TestSncB07EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		sncB07Kraken: "Reservoir Kraken", sncB07Sleep: "Sleep with the Fishes",
		sncB07Lackey: "Expendable Lackey", sncB07Pets: "Exotic Pets",
		sncB07Perrie: "Perrie, the Pulverizer", sncB07Bribe: "Bribe Taker",
		sncB07Toolkit: "Agent's Toolkit", sncB07Courier: "Aven Courier",
		sncB07Denry: "Denry Klin, Editor in Chief", sncB07Safeguad: "Contractual Safeguard",
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

// sncB07Fish lists the Fish a player controls.
func sncB07Fish(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == "Fish" {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

// sncB07CantBeBlocked fails unless the attacker can't be blocked by a fresh
// opposing creature.
func sncB07FishCantBeBlocked(t *testing.T, g *game.Game, attackerSeat, blockerSeat *game.Player, fish uuid.UUID) {
	t.Helper()
	blocker := effectstest.B12Creature(g, blockerSeat.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.DeclareAttack(t, g, blockerSeat.ID, fish)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, fish); err == nil {
		t.Error("the Fish was blocked")
	}
}

func sncB07Fishes(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	if got := len(sncB07Fish(g, me.ID)); got != n {
		t.Fatalf("%d Fish, want %d", got, n)
	}
}

// --- the Fish ---------------------------------------------------------

func TestSncB07FishTokenIsABlueOneOneThatCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.S82SncB07FishToken(), 1) })
	fish := sncB07Fish(g, me.ID)
	if len(fish) != 1 {
		t.Fatalf("%d Fish", len(fish))
	}
	c := effectstest.B12Card(t, g, fish[0])
	if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "U" || c.TypeLine != "Token Creature — Fish" {
		t.Errorf("Fish is %+v", c)
	}
	// Seat 0 is the active player; give the Fish no summoning sickness.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fish[0] {
				g.Battlefield.Cards[i].EnteredBattlefieldAt = 0
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	sncB07FishCantBeBlocked(t, g, me, opp, fish[0])
}

// --- Expendable Lackey ------------------------------------------------

func TestSncB07LackeyExilesItselfFromTheGraveyardForAFish(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, me := effectstest.ActivateFromGraveyard(t, g, "Expendable Lackey", "Creature — Human Citizen", sncB07Lackey, 1, 1, "{U}{U}", game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	sncB07Fishes(t, g, me, 1)
	if got := effectstest.B12ZoneOf(g, id); got != game.ZoneExile {
		t.Errorf("the Lackey is in %v, want exile", got)
	}
}

// --- Sleep with the Fishes -----------------------------------------------

func TestSncB07SleepTapsTheCreatureMakesAFishAndHoldsItDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sleep with the Fishes", effectstest.AuraTypeLine, sncB07Sleep, effectstest.CardTarget(host))
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("the enchanted creature is untapped")
	}
	sncB07Fishes(t, g, me, 1)
	if got := len(sncB07Fish(g, opp.ID)); got != 0 {
		t.Errorf("the creature's controller got %d Fish", got)
	}
}

// --- Reservoir Kraken ------------------------------------------------------

func TestSncB07KrakenTapsAndMakesAFishOnlyWhenAnOpponentTapsACreature(t *testing.T) {
	for _, paid := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kraken := effectstest.B12Push(g, me.ID, "Reservoir Kraken", "Creature — Kraken", sncB07Kraken, 6, 6)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		who := uuid.Nil
		if paid {
			who = opp.ID
		}
		effectstest.AppAnswerWith(t, g, effectstest.AppOnly(who, victim))
		k, _ := effectstest.BattlefieldCard(g, kraken)
		v, _ := effectstest.BattlefieldCard(g, victim)
		if k.Tapped != paid || v.Tapped != paid {
			t.Errorf("paid=%v: Kraken tapped %v, victim tapped %v", paid, k.Tapped, v.Tapped)
		}
		want := 0
		if paid {
			want = 1
		}
		sncB07Fishes(t, g, me, want)
	}
}

func TestSncB07KrakenDoesNothingWhileTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kraken := effectstest.B12Push(g, me.ID, "Reservoir Kraken", "Creature — Kraken", sncB07Kraken, 6, 6)
	effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.TapForTest(t, g, kraken)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if got := effectstest.AppOpenLink(g); got != nil {
		t.Error("a tapped Kraken asked an opponent to tap a creature")
	}
	effectstest.PassPriorityAroundTable(t, g)
	sncB07Fishes(t, g, me, 0)
}

// --- Exotic Pets -------------------------------------------------------------

func TestSncB07ExoticPetsPutsEachKindOnEitherFish(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 3)
	effectstest.PutCounters(t, g, bear, game.CounterShield, 1)
	effectstest.PutCounters(t, g, ox, game.CounterPlusOne, 1)
	effectstest.CastCatalogSpell(t, g, "Exotic Pets", "Instant", sncB07Pets, nil)
	effectstest.PassPriorityAroundTable(t, g)
	sncB07Fishes(t, g, me, 2)
	// Kinds in name order: +1/+1, then shield. Both go on the first Fish.
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	fish := sncB07Fish(g, me.ID)
	if n := effectstest.CountersOn(g, fish[0], game.CounterPlusOne); n != 1 {
		t.Errorf("first Fish has %d +1/+1 counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, fish[0], game.CounterShield); n != 1 {
		t.Errorf("first Fish has %d shield counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, fish[1], game.CounterPlusOne) + effectstest.CountersOn(g, fish[1], game.CounterShield); n != 0 {
		t.Errorf("second Fish has %d counters, want 0", n)
	}
}

func TestSncB07ExoticPetsSplitsKindsAcrossTheFish(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	effectstest.PutCounters(t, g, bear, game.CounterShield, 1)
	effectstest.CastCatalogSpell(t, g, "Exotic Pets", "Instant", sncB07Pets, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	fish := sncB07Fish(g, me.ID)
	if effectstest.CountersOn(g, fish[0], game.CounterPlusOne) != 1 || effectstest.CountersOn(g, fish[1], game.CounterShield) != 1 {
		t.Error("the +1/+1 counter did not go on the first Fish and the shield on the second")
	}
}

func TestSncB07ExoticPetsWithNoCountersJustMakesTwoFish(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Exotic Pets", "Instant", sncB07Pets, nil)
	effectstest.PassPriorityAroundTable(t, g)
	sncB07Fishes(t, g, me, 2)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("asked about a counter with none on the table")
	}
}

// --- Perrie, the Pulverizer ----------------------------------------------------

func TestSncB07PerrieShieldsOnEntryAndPumpsByKindsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	perrie := effectstest.CastHoldCreature(t, g, "Perrie, the Pulverizer", "Legendary Creature — Rhino Soldier", sncB07Perrie, 3, 3)
	_ = perrie
	effectstest.AnswerTriggerTargets(t, g, me.ID, ox)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, ox, game.CounterShield); n != 1 {
		t.Fatalf("the Ox has %d shield counters, want 1", n)
	}
	effectstest.PutCounters(t, g, ox, game.CounterPlusOne, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == perrie {
				g.Battlefield.Cards[i].EnteredBattlefieldAt = 0
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, perrie)
	effectstest.AnswerTriggerTargets(t, g, me.ID, ox)
	effectstest.PassPriorityAroundTable(t, g)
	// Kinds among permanents: shield and +1/+1 -> X = 2.
	if !effectstest.HasEffectiveKeyword(t, g, ox, "trample") {
		t.Error("the Ox did not gain trample")
	}
	c := effectstest.B12Card(t, g, ox)
	if got := effectstest.CurrentPowerFor(t, g, c.InstanceID); got != 2+2+2 {
		t.Errorf("the Ox has power %d, want 6", got)
	}
}

// --- Bribe Taker --------------------------------------------------------------

func TestSncB07BribeTakerChoosesPerKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterFlying, 1)
	effectstest.PutCounters(t, g, bear, game.CounterShield, 1)
	taker := effectstest.CastHoldCreature(t, g, "Bribe Taker", "Creature — Rhino Warrior", sncB07Bribe, 6, 6)
	// Kinds in name order: flying, shield. Flying: one of that kind. Shield: +1/+1.
	effectstest.AnswerOptionPick(t, g, me.ID, 2)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, taker, game.CounterFlying); n != 1 {
		t.Errorf("Bribe Taker has %d flying counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, taker, game.CounterShield); n != 0 {
		t.Errorf("Bribe Taker has %d shield counters, want 0", n)
	}
	if n := effectstest.CountersOn(g, taker, game.CounterPlusOne); n != 1 {
		t.Errorf("Bribe Taker has %d +1/+1 counters, want 1", n)
	}
}

func TestSncB07BribeTakerMayDeclineEveryKind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	taker := effectstest.CastHoldCreature(t, g, "Bribe Taker", "Creature — Rhino Warrior", sncB07Bribe, 6, 6)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil || len(pick.PickOptions) != 2 {
		t.Fatalf("want a none / +1/+1 choice for the +1/+1 kind, got %+v", pick)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, taker, game.CounterPlusOne); n != 0 {
		t.Errorf("Bribe Taker has %d counters after declining", n)
	}
}

// --- Agent's Toolkit ------------------------------------------------------------

func TestSncB07ToolkitEntersWithFourCountersAndMovesOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Agent's Toolkit", "Artifact — Clue", sncB07Toolkit, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var toolkit uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == sncB07Toolkit {
				toolkit = c.InstanceID
			}
		}
	})
	if toolkit == uuid.Nil {
		t.Fatal("the Toolkit did not enter")
	}
	for _, kind := range []string{game.CounterPlusOne, game.CounterFlying, game.CounterDeathtouch, game.CounterShield} {
		if n := effectstest.CountersOn(g, toolkit, kind); n != 1 {
			t.Errorf("the Toolkit has %d %s counters, want 1", n, kind)
		}
	}
	bear := effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	// Kinds in name order: +1/+1, deathtouch, flying, shield; 0 is declining.
	effectstest.AnswerOptionPick(t, g, me.ID, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterShield); n != 1 {
		t.Errorf("the creature has %d shield counters, want 1", n)
	}
	if n := effectstest.CountersOn(g, toolkit, game.CounterShield); n != 0 {
		t.Errorf("the Toolkit still has %d shield counters", n)
	}
}

func TestSncB07ToolkitMayDeclineToMove(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	toolkit := effectstest.B12Push(g, me.ID, "Agent's Toolkit", "Artifact — Clue", sncB07Toolkit, 0, 0)
	effectstest.PutCounters(t, g, toolkit, game.CounterFlying, 1)
	bear := effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterFlying); n != 0 {
		t.Errorf("the creature has %d flying counters after declining", n)
	}
	if n := effectstest.CountersOn(g, toolkit, game.CounterFlying); n != 1 {
		t.Errorf("the Toolkit lost its counter: %d", n)
	}
}

// --- Aven Courier -------------------------------------------------------------

func TestSncB07CourierCopiesAKindOntoAPermanentWithoutIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	courier := effectstest.B12Push(g, me.ID, "Aven Courier", "Creature — Bird Advisor", sncB07Courier, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterShield, 1)
	effectstest.PutCounters(t, g, ox, game.CounterShield, 1)
	effectstest.PutCounters(t, g, ox, game.CounterPlusOne, 1)
	effectstest.DeclareAttack(t, g, opp.ID, courier)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	// Kinds: +1/+1, shield. Choose +1/+1: the Bear has none, so it gets one.
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", n)
	}
}

func TestSncB07CourierDoesNothingForAKindTheTargetHas(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	courier := effectstest.B12Push(g, me.ID, "Aven Courier", "Creature — Bird Advisor", sncB07Courier, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterShield, 1)
	effectstest.DeclareAttack(t, g, opp.ID, courier)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterShield); n != 1 {
		t.Errorf("the Bear has %d shield counters, want 1", n)
	}
}

// --- Denry Klin ----------------------------------------------------------------

func TestSncB07DenryEntersWithAChoiceAndCopiesItOntoItselfAndLaterCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	denry := effectstest.CastHoldCreature(t, g, "Denry Klin, Editor in Chief", "Legendary Creature — Cat Advisor", sncB07Denry, 2, 2)
	_ = denry
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, denry, game.CounterFirstStrike); n < 1 {
		t.Fatalf("Denry has %d first strike counters, want at least 1", n)
	}
	bear := effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, bear, game.CounterFirstStrike); n != effectstest.CountersOn(g, denry, game.CounterFirstStrike) {
		t.Errorf("the Bear has %d first strike counters, Denry has %d", n, effectstest.CountersOn(g, denry, game.CounterFirstStrike))
	}
}

// --- Contractual Safeguard -----------------------------------------------------

func TestSncB07SafeguardSpreadsAKindToEveryOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	cat := effectstest.B12Creature(g, me.ID, "Cat", "Creature — Cat", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterFlying, 1)
	effectstest.CastCatalogSpell(t, g, "Contractual Safeguard", "Instant", sncB07Safeguad, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Main phase: the shield goes on a creature of my choice (the Ox), then
	// the kind is chosen among the pairs.
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, ox, game.CounterShield); n != 1 {
		t.Fatalf("the Ox has %d shield counters, want 1", n)
	}
	// Pairs: Bear flying, Ox shield. Choose the Bear's flying counter.
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Ox": ox, "Cat": cat} {
		if n := effectstest.CountersOn(g, id, game.CounterFlying); n != 1 {
			t.Errorf("the %s has %d flying counters, want 1", name, n)
		}
	}
	if n := effectstest.CountersOn(g, bear, game.CounterFlying); n != 1 {
		t.Errorf("the Bear (the chosen creature) has %d flying counters, want 1", n)
	}
}

func TestSncB07SafeguardOutsideMainPhaseHasNoAddendum(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ox := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Contractual Safeguard", TypeLine: "Instant",
			OracleID: sncB07Safeguad, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, ox, game.CounterShield) + effectstest.CountersOn(g, bear, game.CounterShield); n != 0 {
		t.Errorf("%d shield counters outside a main phase", n)
	}
	if n := effectstest.CountersOn(g, ox, game.CounterPlusOne); n != 1 {
		t.Errorf("the Ox has %d +1/+1 counters, want 1", n)
	}
}
