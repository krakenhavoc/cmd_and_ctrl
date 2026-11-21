package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	cardsbl "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bl"
	cardsg "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/g"
	cardsgr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/gr"
	cardsmi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/mi"
	cardsv "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/v"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hob_a6_cards_test.go — The Hobbit, half A, batch 6 (#4454): the three
// counter lands, Hobbit Hole, and seven Adventure cards. Each test plays
// or casts the real catalog entry and asserts what its text says.

const (
	hobA6MirkwoodOracle        = "cd49aa99-bf84-4edd-aecc-6dae78b73412"
	hobA6ElvenkingsHallsOracle = "a91e0154-14a9-4681-8236-04db231592a4"
	hobA6LakeTownOracle        = "717c6beb-81c6-43ed-aab0-aedfc1cbac33"
	hobA6HobbitHoleOracle      = "17492186-9814-4c41-8111-1f000a96c212"
)

func TestHobA6EveryCardIsRegisteredFull(t *testing.T) {
	for key, name := range map[string]string{
		hobA6MirkwoodOracle:                           "Mirkwood",
		hobA6ElvenkingsHallsOracle:                    "Elvenking's Halls",
		hobA6LakeTownOracle:                           "Lake-town",
		hobA6HobbitHoleOracle:                         "Hobbit Hole",
		cardsg.GollumSilentSlinkerOracleID:            "Gollum, Silent Slinker",
		cardsg.GollumSilentSlinkerOracleID + "#1":     "Meager Meal",
		cardsgr.GreatUglyLookingGoblinOracleID:        "Great Ugly-Looking Goblin",
		cardsgr.GreatUglyLookingGoblinOracleID + "#1": "Clap! Snap!",
		cardsg.GloinTheMightyOracleID:                 "Glóin the Mighty",
		cardsg.GloinTheMightyOracleID + "#1":          "Easy Pickings",
		cardsb.BilboBagginsBurglarOracleID:            "Bilbo Baggins, Burglar",
		cardsb.BilboBagginsBurglarOracleID + "#1":     "Take a Glance",
		cardsbl.BofurReliableGuardianOracleID:         "Bofur, Reliable Guardian",
		cardsbl.BofurReliableGuardianOracleID + "#1":  "Concerted Care",
		cardsv.VelvetwingButterfliesOracleID:          "Velvetwing Butterflies",
		cardsv.VelvetwingButterfliesOracleID + "#1":   "Gaze in Wonder",
		cardsmi.MyPreciousOracleID:                    "My Precious",
		cardsmi.MyPreciousOracleID + "#1":             "Allure of Power",
	} {
		spec, ok := effects.Lookup(key)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// --- the counter lands ------------------------------------------------

type hobA6CounterLandCase struct {
	name, oracle string
	mana         string // the activation cost as a produced-mana string
	produced     string
	good         []string // type lines of creatures it may target
	bad          string   // a type line it may not
}

var hobA6CounterLands = []hobA6CounterLandCase{
	{"Mirkwood", hobA6MirkwoodOracle, "{C}{C}{B}{G}", "{B|G}",
		[]string{"Creature — Bear", "Creature — Spider", "Creature — Wolf"}, "Creature — Elf"},
	{"Elvenking's Halls", hobA6ElvenkingsHallsOracle, "{C}{C}{G}{U}", "{G|U}",
		[]string{"Creature — Elf Warrior"}, "Creature — Human"},
	{"Lake-town", hobA6LakeTownOracle, "{C}{C}{W}{U}", "{W|U}",
		[]string{"Creature — Human Soldier"}, "Creature — Elf"},
}

// hobA6CrackLand pays the land's sacrifice ability and settles it.
func hobA6CrackLand(t *testing.T, g *game.Game, me *game.Player, land uuid.UUID, mana string, target uuid.UUID) error {
	t.Helper()
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestHobA6CounterLandsEnterTappedAndTapForEitherColor(t *testing.T) {
	for _, tc := range hobA6CounterLands {
		g := effectstest.NewCatalogGame(t)
		land := effectstest.B31PlayLand(t, g, tc.name, "Land", tc.oracle)
		if !effectstest.B31Tapped(t, g, land) {
			t.Errorf("%s should enter tapped", tc.name)
		}
		spec, _ := effects.Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced || !spec.ManaAbilities[0].Cost.Tap {
			t.Errorf("%s mana abilities = %+v, want one {T} ability producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

func TestHobA6CounterLandsPutTwoCountersOnTheirTribe(t *testing.T) {
	for _, tc := range hobA6CounterLands {
		for _, typeLine := range tc.good {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			creature := effectstest.B16Creature(g, me.ID, "Target", typeLine, 2, 2)
			if err := hobA6CrackLand(t, g, me, land, tc.mana, creature); err != nil {
				t.Fatalf("%s targeting %q: %v", tc.name, typeLine, err)
			}
			if got := effectstest.CountersOn(g, creature, game.CounterPlusOne); got != 2 {
				t.Errorf("%s on %q: %d +1/+1 counters, want 2", tc.name, typeLine, got)
			}
			if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
				t.Errorf("%s should be sacrificed as a cost", tc.name)
			}
		}
	}
}

func TestHobA6CounterLandsRefuseOtherTribesAndOpponentsCreatures(t *testing.T) {
	for _, tc := range hobA6CounterLands {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
		wrong := effectstest.B16Creature(g, me.ID, "Wrong", tc.bad, 2, 2)
		theirs := effectstest.B16Creature(g, opp.ID, "Theirs", tc.good[0], 2, 2)
		if err := hobA6CrackLand(t, g, me, land, tc.mana, wrong); err == nil {
			t.Errorf("%s accepted a %q", tc.name, tc.bad)
		}
		if err := hobA6CrackLand(t, g, me, land, tc.mana, theirs); err == nil {
			t.Errorf("%s accepted a creature an opponent controls", tc.name)
		}
		if !g.Battlefield.Contains(land) {
			t.Errorf("%s was sacrificed by a refused activation", tc.name)
		}
	}
}

func TestHobA6CounterLandsActivateOnlyAsASorcery(t *testing.T) {
	for _, tc := range hobA6CounterLands {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
		creature := effectstest.B16Creature(g, me.ID, "Target", tc.good[0], 2, 2)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		if err := hobA6CrackLand(t, g, me, land, tc.mana, creature); err == nil {
			t.Errorf("%s activated outside a main phase", tc.name)
		}
	}
}

// --- Hobbit Hole ------------------------------------------------------

func TestHobA6HobbitHoleFetchesABasicLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hole := effectstest.PushCatalogPermanent(g, me.ID, "Hobbit Hole", "Land", hobA6HobbitHoleOracle, false)
	dual := effectstest.StapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")
	forest := effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, hole, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)

	if g.Battlefield.Contains(hole) {
		t.Error("Hobbit Hole should be sacrificed")
	}
	got, ok := effectstest.AangCardOnBF(g, forest)
	if !ok || !got.Tapped {
		t.Errorf("the basic Forest should be on the battlefield tapped: %+v ok=%v", got, ok)
	}
	if _, wrong := effectstest.AangCardOnBF(g, dual); wrong {
		t.Error("a nonbasic land was fetched")
	}
}

func TestHobA6HobbitHoleHalflingcyclesForAHalfling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Hobbit Rogue", TypeLine: "Creature — Halfling Rogue"},
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Halfling Charm", TypeLine: "Legendary Artifact — Halfling"},
	)
	rogue, charm := ids[0], ids[3]
	hole := effectstest.PushCatalogHandCard(me, "Hobbit Hole", "Land", hobA6HobbitHoleOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	// Ability 0 is the battlefield fetch; 1 is the hand-only typecycling.
	if err := g.ActivateCatalogAbility(me.ID, hole, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Halflingcycling: %v", err)
	}
	if !me.Graveyard.Contains(hole) {
		t.Fatal("the cycled Hobbit Hole should be in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt after Halflingcycling")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the two Halflings", len(c.SearchCards))
	}
	for _, got := range c.SearchCards {
		if got != rogue && got != charm {
			t.Errorf("search offers %v, which is not a Halfling card", got)
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{rogue}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(rogue) {
		t.Error("the Halfling did not reach the hand")
	}
}

func TestHobA6MeagerMealPutsACounterAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsg.GollumSilentSlinkerOracleID, effectstest.HobA6GollumCreature, effectstest.HobA6MeagerMeal)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	before := opp.Life

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(bear), {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}})

	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the creature has %d +1/+1 counters, want 1", got)
	}
	if opp.Life != before+2 {
		t.Errorf("target player's life %d -> %d, want +2", before, opp.Life)
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6MeagerMealWithNoCreatureStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsg.GollumSilentSlinkerOracleID, effectstest.HobA6GollumCreature, effectstest.HobA6MeagerMeal)
	before := me.Life

	// The optional creature clause (slot 0) is skipped, so the player answers slot 1.
	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 1}})

	if me.Life != before+2 {
		t.Errorf("life %d -> %d, want +2 with no creature chosen", before, me.Life)
	}
}

func TestHobA6GollumHasMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsg.GollumSilentSlinkerOracleID, effectstest.HobA6GollumCreature, effectstest.HobA6MeagerMeal)
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Gollum should have resolved onto the battlefield")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "menace") {
		t.Error("Gollum should have menace")
	}
}

// --- Great Ugly-Looking Goblin // Clap! Snap! -------------------------

var (
	hobA6GoblinCreature = effectstest.HobA6Face("Great Ugly-Looking Goblin", "Creature — Goblin Soldier", "{5}{B}", 4, 4, "B")
	hobA6ClapSnap       = effectstest.HobA6Face("Clap! Snap!", "Sorcery — Adventure", "{1}{B}", 0, 0, "B")
)

func TestHobA6GreatUglyLookingGoblinGivesMenaceToYourCreaturesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := effectstest.PushCatalogPermanent(g, me.ID, "Great Ugly-Looking Goblin", "Creature — Goblin Soldier", cardsgr.GreatUglyLookingGoblinOracleID, false)
	mine := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	plain := effectstest.SeedCreature(g, "Plain", me.ID)
	theirs := effectstest.SeedCreatureWithCounters(g, opp.ID, map[string]int{game.CounterPlusOne: 1})
	otherCounter := effectstest.SeedCreatureWithCounters(g, me.ID, map[string]int{"charge": 2})

	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want bool
	}{
		{"your creature with a +1/+1 counter", mine, true},
		{"your creature with no counter", plain, false},
		{"the Goblin itself, with no counter", goblin, false},
		{"your creature with a different counter", otherCounter, false},
		{"an opponent's creature with a +1/+1 counter", theirs, false},
	} {
		if got := effectstest.HasKeywordOnBattlefield(t, g, tc.id, "menace"); got != tc.want {
			t.Errorf("%s: menace = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestHobA6ClapSnapAmassesGoblinsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsgr.GreatUglyLookingGoblinOracleID, hobA6GoblinCreature, hobA6ClapSnap)

	effectstest.HobA6MustCast(t, g, me, id, 1, nil)

	army := effectstest.BattlefieldIDNamed(g, me.ID, "Goblin Army")
	if army == uuid.Nil {
		t.Fatal("no Goblin Army token was created")
	}
	if got := effectstest.CountersOn(g, army, game.CounterPlusOne); got != 2 {
		t.Errorf("the Army has %d +1/+1 counters, want 2", got)
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

// --- Glóin the Mighty // Easy Pickings ---------------------------------

var (
	hobA6GloinCreature = effectstest.HobA6Face("Glóin the Mighty", "Legendary Creature — Dwarf Warrior", "{3}{R}", 4, 3, "R")
	hobA6EasyPickings  = effectstest.HobA6Face("Easy Pickings", "Sorcery — Adventure", "{2}{R}", 0, 0, "R")
)

func TestHobA6GloinAddsRedRedAtYourFirstMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Glóin the Mighty", "Legendary Creature — Dwarf Warrior", cardsg.GloinTheMightyOracleID, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	red := 0
	for _, c := range effectstest.PoolColors(me) {
		if c == "R" {
			red++
		}
	}
	if red != 2 {
		t.Errorf("pool has %d red, want 2 (%v)", red, effectstest.PoolColors(me))
	}
}

func TestHobA6GloinDoesNotTriggerOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.PushCatalogPermanent(g, opp.ID, "Glóin the Mighty", "Legendary Creature — Dwarf Warrior", cardsg.GloinTheMightyOracleID, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(effectstest.PoolColors(opp)); got != 0 {
		t.Errorf("Glóin's controller got %d mana on someone else's main phase", got)
	}
}

func TestHobA6EasyPickingsDamagesOnlyYourOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsg.GloinTheMightyOracleID, hobA6GloinCreature, hobA6EasyPickings)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.SeedCreature(g, "Mine", me.ID)
	small := effectstest.B16Creature(g, opp.ID, "Small", "Creature — Rat", 1, 1)
	big := effectstest.SeedCreature(g, "Big", opp.ID)

	effectstest.HobA6MustCast(t, g, me, id, 1, nil)

	if g.Battlefield.Contains(small) {
		t.Error("an opponent's 1/1 should die")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 1 {
		t.Errorf("an opponent's 2/2 has %d damage, want 1", got)
	}
	if got := effectstest.DamageMarkedOn(g, mine); got != 0 {
		t.Errorf("your own creature took %d damage", got)
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

// --- Bilbo Baggins, Burglar // Take a Glance ---------------------------

var (
	hobA6BilboCreature = effectstest.HobA6Face("Bilbo Baggins, Burglar", "Legendary Creature — Halfling Rogue", "{2}{U}", 2, 1, "U")
	hobA6TakeAGlance   = effectstest.HobA6Face("Take a Glance", "Sorcery — Adventure", "{U}", 0, 0, "U")
)

func TestHobA6BilboDrawsACardWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsb.BilboBagginsBurglarOracleID, hobA6BilboCreature, hobA6TakeAGlance)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	before := me.Hand.Size()

	effectstest.HobA6MustCast(t, g, me, id, 0, nil)

	if !g.Battlefield.Contains(id) {
		t.Fatal("Bilbo should be on the battlefield")
	}
	// The card left the hand and one was drawn.
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size %d -> %d, want the cast card replaced by a draw", before, got)
	}
}

func TestHobA6TakeAGlanceScriesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsb.BilboBagginsBurglarOracleID, hobA6BilboCreature, hobA6TakeAGlance)
	effectstest.SeedLibrary(me, "One", "Two", "Three")

	effectstest.HobA6MustCast(t, g, me, id, 1, nil)

	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("scry 2 offered %d cards", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

// --- Bofur, Reliable Guardian // Concerted Care ------------------------

var (
	hobA6BofurCreature = effectstest.HobA6Face("Bofur, Reliable Guardian", "Legendary Creature — Dwarf Scout", "{W}", 1, 1, "W")
	hobA6ConcertedCare = effectstest.HobA6Face("Concerted Care", "Instant — Adventure", "{1}{W}", 0, 0, "W")
)

func TestHobA6BofurHasLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsbl.BofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "lifelink") {
		t.Error("Bofur should have lifelink")
	}
}

func TestHobA6ConcertedCareGrantsHexproofAndIndestructibleToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsbl.BofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	bear := effectstest.SeedCreature(g, "Bear", me.ID)

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(bear)})

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the creature should have %s", kw)
		}
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6ConcertedCareCanTargetAnArtifactYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsbl.BofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(rock)})

	if !effectstest.HasKeywordOnBattlefield(t, g, rock, "hexproof") {
		t.Error("the artifact should have hexproof")
	}
}

func TestHobA6ConcertedCareRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsbl.BofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.SeedCreature(g, "Theirs", opp.ID)

	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(theirs)}); err == nil {
		t.Error("Concerted Care accepted a creature you do not control")
	}
}

// --- Velvetwing Butterflies // Gaze in Wonder --------------------------

var (
	hobA6ButterfliesCreature = effectstest.HobA6Face("Velvetwing Butterflies", "Creature — Insect", "{2}{W}", 2, 2, "W")
	hobA6GazeInWonder        = effectstest.HobA6Face("Gaze in Wonder", "Instant — Adventure", "{1}{W}", 0, 0, "W")
)

func TestHobA6VelvetwingButterfliesFly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsv.VelvetwingButterfliesOracleID, hobA6ButterfliesCreature, hobA6GazeInWonder)
	effectstest.HobA6MustCast(t, g, me, id, 0, nil)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("Velvetwing Butterflies should have flying")
	}
}

func TestHobA6GazeInWonderTapsOneOrTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsv.VelvetwingButterfliesOracleID, hobA6ButterfliesCreature, hobA6GazeInWonder)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.SeedCreature(g, "A", opp.ID)
	b := effectstest.SeedCreature(g, "B", opp.ID)
	c := effectstest.SeedCreature(g, "C", opp.ID)

	effectstest.HobA6MustCast(t, g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(a), effectstest.HobA6Card(b)})

	if !effectstest.B16Tapped(t, g, a) || !effectstest.B16Tapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
	if effectstest.B16Tapped(t, g, c) {
		t.Error("an untargeted creature was tapped")
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6GazeInWonderRefusesThreeTargetsAndNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsv.VelvetwingButterfliesOracleID, hobA6ButterfliesCreature, hobA6GazeInWonder)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.SeedCreature(g, "A", opp.ID)
	b := effectstest.SeedCreature(g, "B", opp.ID)
	c := effectstest.SeedCreature(g, "C", opp.ID)

	if err := effectstest.HobA6Cast(g, me, id, 1, []game.TargetRef{effectstest.HobA6Card(a), effectstest.HobA6Card(b), effectstest.HobA6Card(c)}); err == nil {
		t.Error("Gaze in Wonder accepted three targets")
	}
	if err := effectstest.HobA6Cast(g, me, id, 1, nil); err == nil {
		t.Error("Gaze in Wonder accepted no target")
	}
}

// --- My Precious // Allure of Power ------------------------------------

var (
	hobA6PreciousArtifact = effectstest.HobA6Face("My Precious", "Legendary Artifact — Equipment", "{3}", 0, 0)
	hobA6AllureOfPower    = effectstest.HobA6Face("Allure of Power", "Instant — Adventure", "{1}{B}", 0, 0, "B")
)

func TestHobA6MyPreciousGrantsHexproofAndUnblockableForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	precious := effectstest.PushCatalogPermanent(g, me.ID, "My Precious", "Legendary Artifact — Equipment", cardsmi.MyPreciousOracleID, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	before := me.Life

	effectstest.EquipTo(t, g, me.ID, precious, bear)

	if me.Life != before-2 {
		t.Errorf("life %d -> %d, want the equip's 2 life paid", before, me.Life)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "hexproof") {
		t.Error("the equipped creature should have hexproof")
	}
	effectstest.AssertRestrictions(t, g, bear, game.CantBeBlocked)
}

func TestHobA6AllureOfPowerSacrificesACreatureToDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsmi.MyPreciousOracleID, hobA6PreciousArtifact, hobA6AllureOfPower)
	effectstest.SeedLibrary(me, "One", "Two", "Three")
	fodder := effectstest.SeedCreature(g, "Fodder", me.ID)
	before := me.Hand.Size()

	effectstest.HobA6MustCast(t, g, me, id, 1, nil, fodder)

	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	// The cast card left the hand and two were drawn.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size %d -> %d, want +1 net", before, got)
	}
	effectstest.HobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6AllureOfPowerNeedsACreatureToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.HobA6HandAdventure(t, g, cardsmi.MyPreciousOracleID, hobA6PreciousArtifact, hobA6AllureOfPower)
	if err := effectstest.HobA6Cast(g, me, id, 1, nil); err == nil {
		t.Error("Allure of Power was cast without sacrificing a creature")
	}
}
