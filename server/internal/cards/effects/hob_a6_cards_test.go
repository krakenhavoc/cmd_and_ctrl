package effects

import (
	"testing"

	"github.com/google/uuid"

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
		hobA6MirkwoodOracle:                   "Mirkwood",
		hobA6ElvenkingsHallsOracle:            "Elvenking's Halls",
		hobA6LakeTownOracle:                   "Lake-town",
		hobA6HobbitHoleOracle:                 "Hobbit Hole",
		gollumSilentSlinkerOracleID:           "Gollum, Silent Slinker",
		gollumSilentSlinkerOracleID + "#1":    "Meager Meal",
		greatUglyLookingGoblinOracleID:        "Great Ugly-Looking Goblin",
		greatUglyLookingGoblinOracleID + "#1": "Clap! Snap!",
		gloinTheMightyOracleID:                "Glóin the Mighty",
		gloinTheMightyOracleID + "#1":         "Easy Pickings",
		bilboBagginsBurglarOracleID:           "Bilbo Baggins, Burglar",
		bilboBagginsBurglarOracleID + "#1":    "Take a Glance",
		bofurReliableGuardianOracleID:         "Bofur, Reliable Guardian",
		bofurReliableGuardianOracleID + "#1":  "Concerted Care",
		velvetwingButterfliesOracleID:         "Velvetwing Butterflies",
		velvetwingButterfliesOracleID + "#1":  "Gaze in Wonder",
		myPreciousOracleID:                    "My Precious",
		myPreciousOracleID + "#1":             "Allure of Power",
	} {
		spec, ok := Lookup(key)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull {
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
		passPriorityAroundTable(t, g)
	}
	return err
}

func TestHobA6CounterLandsEnterTappedAndTapForEitherColor(t *testing.T) {
	for _, tc := range hobA6CounterLands {
		g := newCatalogGame(t)
		land := b31PlayLand(t, g, tc.name, "Land", tc.oracle)
		if !b31Tapped(t, g, land) {
			t.Errorf("%s should enter tapped", tc.name)
		}
		spec, _ := Lookup(tc.oracle)
		if len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != tc.produced || !spec.ManaAbilities[0].Cost.Tap {
			t.Errorf("%s mana abilities = %+v, want one {T} ability producing %s", tc.name, spec.ManaAbilities, tc.produced)
		}
	}
}

func TestHobA6CounterLandsPutTwoCountersOnTheirTribe(t *testing.T) {
	for _, tc := range hobA6CounterLands {
		for _, typeLine := range tc.good {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			creature := b16Creature(g, me.ID, "Target", typeLine, 2, 2)
			if err := hobA6CrackLand(t, g, me, land, tc.mana, creature); err != nil {
				t.Fatalf("%s targeting %q: %v", tc.name, typeLine, err)
			}
			if got := countersOn(g, creature, game.CounterPlusOne); got != 2 {
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
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
		wrong := b16Creature(g, me.ID, "Wrong", tc.bad, 2, 2)
		theirs := b16Creature(g, opp.ID, "Theirs", tc.good[0], 2, 2)
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
		g := newCatalogGame(t)
		advanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
		creature := b16Creature(g, me.ID, "Target", tc.good[0], 2, 2)
		advanceTo(t, g, game.StepBeginCombat)
		if err := hobA6CrackLand(t, g, me, land, tc.mana, creature); err == nil {
			t.Errorf("%s activated outside a main phase", tc.name)
		}
	}
}

// --- Hobbit Hole ------------------------------------------------------

func TestHobA6HobbitHoleFetchesABasicLandTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hole := pushCatalogPermanent(g, me.ID, "Hobbit Hole", "Land", hobA6HobbitHoleOracle, false)
	dual := stapleLibraryCard(me, "Watery Grave", "Land — Island Swamp")
	forest := stapleLibraryCard(me, "Forest", "Basic Land — Forest")

	if err := g.ActivateCatalogAbility(me.ID, hole, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)

	if g.Battlefield.Contains(hole) {
		t.Error("Hobbit Hole should be sacrificed")
	}
	got, ok := aangCardOnBF(g, forest)
	if !ok || !got.Tapped {
		t.Errorf("the basic Forest should be on the battlefield tapped: %+v ok=%v", got, ok)
	}
	if _, wrong := aangCardOnBF(g, dual); wrong {
		t.Error("a nonbasic land was fetched")
	}
}

func TestHobA6HobbitHoleHalflingcyclesForAHalfling(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Hobbit Rogue", TypeLine: "Creature — Halfling Rogue"},
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Halfling Charm", TypeLine: "Legendary Artifact — Halfling"},
	)
	rogue, charm := ids[0], ids[3]
	hole := pushCatalogHandCard(me, "Hobbit Hole", "Land", hobA6HobbitHoleOracle)
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
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
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

// --- Adventure plumbing -----------------------------------------------

func hobA6Face(name, typeLine, cost string, power, toughness int, colors ...string) game.Face {
	return game.Face{Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors, Power: power, Toughness: toughness}
}

// hobA6HandAdventure seats an Adventure card (creature face 0, Adventure
// face 1) in the active player's hand at a main phase.
func hobA6HandAdventure(t *testing.T, g *game.Game, oracle string, creature, adventure game.Face) (*game.Player, uuid.UUID) {
	t.Helper()
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutAdventure,
		Owner: me.ID, Controller: me.ID, Faces: []game.Face{creature, adventure},
	}
	c.SetFace(0)
	me.Hand.PushTop(c)
	return me, c.InstanceID
}

func hobA6Cast(g *game.Game, me *game.Player, id uuid.UUID, face int, targets []game.TargetRef, sacrifice ...uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{Face: face, Targets: targets, SacrificeIDs: sacrifice})
}

func hobA6MustCast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, face int, targets []game.TargetRef, sacrifice ...uuid.UUID) {
	t.Helper()
	if err := hobA6Cast(g, me, id, face, targets, sacrifice...); err != nil {
		t.Fatalf("cast face %d: %v", face, err)
	}
	passPriorityAroundTable(t, g)
}

// hobA6WentOnAnAdventure asserts CR 715.3d / 715.4: the card is in exile
// and its owner may cast the creature face from there.
func hobA6WentOnAnAdventure(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if !g.Exile.Contains(id) {
		t.Fatal("CR 715.3d: the resolved Adventure is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Errorf("grant faces = %v, want just the creature face [0]", perm.Faces)
	}
}

func hobA6Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- Gollum, Silent Slinker // Meager Meal ----------------------------

var (
	hobA6GollumCreature = hobA6Face("Gollum, Silent Slinker", "Legendary Creature — Halfling Horror", "{3}{B}", 4, 3, "B")
	hobA6MeagerMeal     = hobA6Face("Meager Meal", "Sorcery — Adventure", "{B}", 0, 0, "B")
)

func TestHobA6MeagerMealPutsACounterAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, gollumSilentSlinkerOracleID, hobA6GollumCreature, hobA6MeagerMeal)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := seedCreature(g, "Bear", me.ID)
	before := opp.Life

	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{hobA6Card(bear), {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}})

	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the creature has %d +1/+1 counters, want 1", got)
	}
	if opp.Life != before+2 {
		t.Errorf("target player's life %d -> %d, want +2", before, opp.Life)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6MeagerMealWithNoCreatureStillGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, gollumSilentSlinkerOracleID, hobA6GollumCreature, hobA6MeagerMeal)
	before := me.Life

	// The optional creature clause (slot 0) is skipped, so the player answers slot 1.
	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID, Slot: 1}})

	if me.Life != before+2 {
		t.Errorf("life %d -> %d, want +2 with no creature chosen", before, me.Life)
	}
}

func TestHobA6GollumHasMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, gollumSilentSlinkerOracleID, hobA6GollumCreature, hobA6MeagerMeal)
	hobA6MustCast(t, g, me, id, 0, nil)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Gollum should have resolved onto the battlefield")
	}
	if !hasKeywordOnBattlefield(t, g, id, "menace") {
		t.Error("Gollum should have menace")
	}
}

// --- Great Ugly-Looking Goblin // Clap! Snap! -------------------------

var (
	hobA6GoblinCreature = hobA6Face("Great Ugly-Looking Goblin", "Creature — Goblin Soldier", "{5}{B}", 4, 4, "B")
	hobA6ClapSnap       = hobA6Face("Clap! Snap!", "Sorcery — Adventure", "{1}{B}", 0, 0, "B")
)

func TestHobA6GreatUglyLookingGoblinGivesMenaceToYourCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goblin := pushCatalogPermanent(g, me.ID, "Great Ugly-Looking Goblin", "Creature — Goblin Soldier", greatUglyLookingGoblinOracleID, false)
	mine := seedCreatureWithCounters(g, me.ID, map[string]int{game.CounterPlusOne: 1})
	plain := seedCreature(g, "Plain", me.ID)
	theirs := seedCreatureWithCounters(g, opp.ID, map[string]int{game.CounterPlusOne: 1})
	otherCounter := seedCreatureWithCounters(g, me.ID, map[string]int{"charge": 2})

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
		if got := hasKeywordOnBattlefield(t, g, tc.id, "menace"); got != tc.want {
			t.Errorf("%s: menace = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestHobA6ClapSnapAmassesGoblinsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, greatUglyLookingGoblinOracleID, hobA6GoblinCreature, hobA6ClapSnap)

	hobA6MustCast(t, g, me, id, 1, nil)

	army := battlefieldIDNamed(g, me.ID, "Goblin Army")
	if army == uuid.Nil {
		t.Fatal("no Goblin Army token was created")
	}
	if got := countersOn(g, army, game.CounterPlusOne); got != 2 {
		t.Errorf("the Army has %d +1/+1 counters, want 2", got)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

// --- Glóin the Mighty // Easy Pickings ---------------------------------

var (
	hobA6GloinCreature = hobA6Face("Glóin the Mighty", "Legendary Creature — Dwarf Warrior", "{3}{R}", 4, 3, "R")
	hobA6EasyPickings  = hobA6Face("Easy Pickings", "Sorcery — Adventure", "{2}{R}", 0, 0, "R")
)

func TestHobA6GloinAddsRedRedAtYourFirstMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "Glóin the Mighty", "Legendary Creature — Dwarf Warrior", gloinTheMightyOracleID, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	red := 0
	for _, c := range poolColors(me) {
		if c == "R" {
			red++
		}
	}
	if red != 2 {
		t.Errorf("pool has %d red, want 2 (%v)", red, poolColors(me))
	}
}

func TestHobA6GloinDoesNotTriggerOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, opp.ID, "Glóin the Mighty", "Legendary Creature — Dwarf Warrior", gloinTheMightyOracleID, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	passPriorityAroundTable(t, g)

	if got := len(poolColors(opp)); got != 0 {
		t.Errorf("Glóin's controller got %d mana on someone else's main phase", got)
	}
}

func TestHobA6EasyPickingsDamagesOnlyYourOpponentsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, gloinTheMightyOracleID, hobA6GloinCreature, hobA6EasyPickings)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := seedCreature(g, "Mine", me.ID)
	small := b16Creature(g, opp.ID, "Small", "Creature — Rat", 1, 1)
	big := seedCreature(g, "Big", opp.ID)

	hobA6MustCast(t, g, me, id, 1, nil)

	if g.Battlefield.Contains(small) {
		t.Error("an opponent's 1/1 should die")
	}
	if got := damageMarkedOn(g, big); got != 1 {
		t.Errorf("an opponent's 2/2 has %d damage, want 1", got)
	}
	if got := damageMarkedOn(g, mine); got != 0 {
		t.Errorf("your own creature took %d damage", got)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

// --- Bilbo Baggins, Burglar // Take a Glance ---------------------------

var (
	hobA6BilboCreature = hobA6Face("Bilbo Baggins, Burglar", "Legendary Creature — Halfling Rogue", "{2}{U}", 2, 1, "U")
	hobA6TakeAGlance   = hobA6Face("Take a Glance", "Sorcery — Adventure", "{U}", 0, 0, "U")
)

func TestHobA6BilboDrawsACardWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, bilboBagginsBurglarOracleID, hobA6BilboCreature, hobA6TakeAGlance)
	seedLibrary(me, "One", "Two", "Three")
	before := me.Hand.Size()

	hobA6MustCast(t, g, me, id, 0, nil)

	if !g.Battlefield.Contains(id) {
		t.Fatal("Bilbo should be on the battlefield")
	}
	// The card left the hand and one was drawn.
	if got := me.Hand.Size(); got != before {
		t.Errorf("hand size %d -> %d, want the cast card replaced by a draw", before, got)
	}
}

func TestHobA6TakeAGlanceScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, bilboBagginsBurglarOracleID, hobA6BilboCreature, hobA6TakeAGlance)
	seedLibrary(me, "One", "Two", "Three")

	hobA6MustCast(t, g, me, id, 1, nil)

	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("scry 2 offered %d cards", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

// --- Bofur, Reliable Guardian // Concerted Care ------------------------

var (
	hobA6BofurCreature = hobA6Face("Bofur, Reliable Guardian", "Legendary Creature — Dwarf Scout", "{W}", 1, 1, "W")
	hobA6ConcertedCare = hobA6Face("Concerted Care", "Instant — Adventure", "{1}{W}", 0, 0, "W")
)

func TestHobA6BofurHasLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, bofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	hobA6MustCast(t, g, me, id, 0, nil)
	if !hasKeywordOnBattlefield(t, g, id, "lifelink") {
		t.Error("Bofur should have lifelink")
	}
}

func TestHobA6ConcertedCareGrantsHexproofAndIndestructibleToACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, bofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	bear := seedCreature(g, "Bear", me.ID)

	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{hobA6Card(bear)})

	for _, kw := range []string{"hexproof", "indestructible"} {
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the creature should have %s", kw)
		}
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6ConcertedCareCanTargetAnArtifactYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, bofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{hobA6Card(rock)})

	if !hasKeywordOnBattlefield(t, g, rock, "hexproof") {
		t.Error("the artifact should have hexproof")
	}
}

func TestHobA6ConcertedCareRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, bofurReliableGuardianOracleID, hobA6BofurCreature, hobA6ConcertedCare)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := seedCreature(g, "Theirs", opp.ID)

	if err := hobA6Cast(g, me, id, 1, []game.TargetRef{hobA6Card(theirs)}); err == nil {
		t.Error("Concerted Care accepted a creature you do not control")
	}
}

// --- Velvetwing Butterflies // Gaze in Wonder --------------------------

var (
	hobA6ButterfliesCreature = hobA6Face("Velvetwing Butterflies", "Creature — Insect", "{2}{W}", 2, 2, "W")
	hobA6GazeInWonder        = hobA6Face("Gaze in Wonder", "Instant — Adventure", "{1}{W}", 0, 0, "W")
)

func TestHobA6VelvetwingButterfliesFly(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, velvetwingButterfliesOracleID, hobA6ButterfliesCreature, hobA6GazeInWonder)
	hobA6MustCast(t, g, me, id, 0, nil)
	if !hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("Velvetwing Butterflies should have flying")
	}
}

func TestHobA6GazeInWonderTapsOneOrTwoCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, velvetwingButterfliesOracleID, hobA6ButterfliesCreature, hobA6GazeInWonder)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := seedCreature(g, "A", opp.ID)
	b := seedCreature(g, "B", opp.ID)
	c := seedCreature(g, "C", opp.ID)

	hobA6MustCast(t, g, me, id, 1, []game.TargetRef{hobA6Card(a), hobA6Card(b)})

	if !b16Tapped(t, g, a) || !b16Tapped(t, g, b) {
		t.Error("both targets should be tapped")
	}
	if b16Tapped(t, g, c) {
		t.Error("an untargeted creature was tapped")
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6GazeInWonderRefusesThreeTargetsAndNone(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, velvetwingButterfliesOracleID, hobA6ButterfliesCreature, hobA6GazeInWonder)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := seedCreature(g, "A", opp.ID)
	b := seedCreature(g, "B", opp.ID)
	c := seedCreature(g, "C", opp.ID)

	if err := hobA6Cast(g, me, id, 1, []game.TargetRef{hobA6Card(a), hobA6Card(b), hobA6Card(c)}); err == nil {
		t.Error("Gaze in Wonder accepted three targets")
	}
	if err := hobA6Cast(g, me, id, 1, nil); err == nil {
		t.Error("Gaze in Wonder accepted no target")
	}
}

// --- My Precious // Allure of Power ------------------------------------

var (
	hobA6PreciousArtifact = hobA6Face("My Precious", "Legendary Artifact — Equipment", "{3}", 0, 0)
	hobA6AllureOfPower    = hobA6Face("Allure of Power", "Instant — Adventure", "{1}{B}", 0, 0, "B")
)

func TestHobA6MyPreciousGrantsHexproofAndUnblockableForTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := seedCreature(g, "Bear", me.ID)
	precious := pushCatalogPermanent(g, me.ID, "My Precious", "Legendary Artifact — Equipment", myPreciousOracleID, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	before := me.Life

	equipTo(t, g, me.ID, precious, bear)

	if me.Life != before-2 {
		t.Errorf("life %d -> %d, want the equip's 2 life paid", before, me.Life)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "hexproof") {
		t.Error("the equipped creature should have hexproof")
	}
	assertRestrictions(t, g, bear, game.CantBeBlocked)
}

func TestHobA6AllureOfPowerSacrificesACreatureToDrawTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, myPreciousOracleID, hobA6PreciousArtifact, hobA6AllureOfPower)
	seedLibrary(me, "One", "Two", "Three")
	fodder := seedCreature(g, "Fodder", me.ID)
	before := me.Hand.Size()

	hobA6MustCast(t, g, me, id, 1, nil, fodder)

	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
	// The cast card left the hand and two were drawn.
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand size %d -> %d, want +1 net", before, got)
	}
	hobA6WentOnAnAdventure(t, g, id)
}

func TestHobA6AllureOfPowerNeedsACreatureToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, id := hobA6HandAdventure(t, g, myPreciousOracleID, hobA6PreciousArtifact, hobA6AllureOfPower)
	if err := hobA6Cast(g, me, id, 1, nil); err == nil {
		t.Error("Allure of Power was cast without sacrificing a creature")
	}
}
