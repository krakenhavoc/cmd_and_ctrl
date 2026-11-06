package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_ab_cards_test.go — Tarkir: Dragonstorm, the cards from A to B
// (#4463). Each test plays the real catalog entry and asserts what its
// text says.

const (
	tdmAbzanDevoteeOracle     = "3d02ee2f-b7bd-40b0-b9be-2420991b2be0"
	tdmAbzanMonumentOracle    = "93563b7b-5685-4832-b164-0701dbf3cf18"
	tdmAdornedCrocodileOracle = "fb49c7d2-d750-4676-b90e-76e7fae70a25"
	tdmAegisSculptorOracle    = "3a75c56a-bdf7-490d-aa41-51d5eb8be46c"
	tdmAgentOfKotisOracle     = "584fe1e4-029d-43fe-8c9e-6995dd392d65"
	tdmAinokWayfarerOracle    = "8134f3fd-197a-4b44-a370-311cc955dea3"
	tdmAlchemistOracle        = "2844d274-cb60-46c1-a73b-e763c633e496"
	tdmAleshasLegacyOracle    = "c4229cc8-11e5-4221-b00d-86969701e21b"
	tdmAllOutAssaultOracle    = "de8c0e2d-3207-4bc6-83ef-91e33b26a981"
	tdmStormshellOracle       = "75070eaf-c931-46e5-9aa1-b849e33580c6"
	tdmAnafenzaOracle         = "659bc01c-fa60-4d00-bdad-5bca1d2a978f"
	tdmAttunedHunterOracle    = "f652c593-e8fa-42ff-a8b4-206556b273de"
	tdmAvengerOracle          = "158d0272-a850-4399-8afa-d0caa143c3cb"
	tdmAwakenOracle           = "5f694917-f434-4459-bc86-caf9ac1c80e7"
	tdmBearerOfGloryOracle    = "963947d5-5fe4-49cd-be0c-0bdb5f7bcbb6"
	tdmBlizzardOracle         = "84469042-d48d-409f-a2c9-047058f13bc5"
	tdmBoneCairnOracle        = "449db6bf-d766-4d09-a7bd-9c45b3d4c4c3"
	tdmBoulderbornOracle      = "1654cab1-0697-4547-b1fe-2d0a61385199"
	tdmBreachingOracle        = "5370ee80-0354-454f-a16e-ec133e432ecb"
)

func TestTdmABEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmAbzanDevoteeOracle:     "Abzan Devotee",
		tdmAbzanMonumentOracle:    "Abzan Monument",
		tdmAdornedCrocodileOracle: "Adorned Crocodile",
		tdmAegisSculptorOracle:    "Aegis Sculptor",
		tdmAgentOfKotisOracle:     "Agent of Kotis",
		tdmAinokWayfarerOracle:    "Ainok Wayfarer",
		tdmAlchemistOracle:        "Alchemist's Assistant",
		tdmAleshasLegacyOracle:    "Alesha's Legacy",
		tdmAllOutAssaultOracle:    "All-Out Assault",
		tdmStormshellOracle:       "Ambling Stormshell",
		tdmAnafenzaOracle:         "Anafenza, Unyielding Lineage",
		tdmAttunedHunterOracle:    "Attuned Hunter",
		tdmAvengerOracle:          "Avenger of the Fallen",
		tdmAwakenOracle:           "Awaken the Honored Dead",
		tdmBearerOfGloryOracle:    "Bearer of Glory",
		tdmBlizzardOracle:         "Bewildering Blizzard",
		tdmBoneCairnOracle:        "Bone-Cairn Butcher",
		tdmBoulderbornOracle:      "Boulderborn Dragon",
		tdmBreachingOracle:        "Breaching Dragonstorm",
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

// --- fixtures ------------------------------------------------------------------

// tdmTap taps a battlefield permanent directly.
func tdmTap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// tdmPlaceWarriors answers every "which player does this token attack"
// prompt with the first offer, as the mobilize tests do.
func tdmPlaceWarriors(t *testing.T, g *game.Game, me uuid.UUID) {
	t.Helper()
	for c := latestChoiceOfKind(g, game.PendingChoicePayAmount); c != nil; c = latestChoiceOfKind(g, game.PendingChoicePayAmount) {
		if err := g.ResolvePayAmount(c.ID, me, c.PayAmount.Max); err != nil {
			t.Fatalf("place the Warriors: %v", err)
		}
	}
}

// tdmExileFromGraveyard exiles a graveyard card by an effect `by`
// controls.
func tdmExileFromGraveyard(t *testing.T, g *game.Game, by, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: by, Owner: by})
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
}

// --- Abzan Devotee -------------------------------------------------------------

func TestTdmAbzanDevoteeFiltersOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	devotee := b12Push(g, me.ID, "Abzan Devotee", "Creature — Dog Cleric", tdmAbzanDevoteeOracle, 2, 1)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "R", 2)
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("{1}: add {G}: %v", err)
	}
	if got := mcPoolCount(me, "G"); got != 1 {
		t.Errorf("green in the pool %d, want 1", got)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool %d, want 2 (one {R} paid, one {G} made)", me.ManaPool.Total())
	}
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

func TestTdmAbzanDevoteeRefusesAColourItDoesNotName(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	devotee := b12Push(g, me.ID, "Abzan Devotee", "Creature — Dog Cleric", tdmAbzanDevoteeOracle, 2, 1)
	advanceToMain(t, g)
	fillPoolColored(me, "R", 1)
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("the Devotee made {U}; it adds {W}, {B}, or {G}")
	}
}

func TestTdmAbzanDevoteeReturnsFromTheGraveyardToHand(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Abzan Devotee", "Creature — Dog Cleric",
		tdmAbzanDevoteeOracle, 2, 1, "{B}{B}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the Devotee did not return from the graveyard to its owner's hand")
	}
}

// --- Abzan Monument ------------------------------------------------------------

func TestTdmAbzanMonumentFindsABasicPlainsSwampOrForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Murky Swamp", "Land — Swamp"),
		searchTestLand("Forest", "Basic Land — Forest"),
	)
	castCatalogSpell(t, g, "Abzan Monument", "Artifact", tdmAbzanMonumentOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Monument's entry asked for no search")
	}
	if searchOptionNamed(g, c, "Island") != uuid.Nil || searchOptionNamed(g, c, "Murky Swamp") != uuid.Nil {
		t.Error("only a basic Plains, Swamp, or Forest may be found")
	}
	answerSearchNamed(t, g, me.ID, "Swamp")
	found := false
	for _, card := range me.Hand.Cards {
		if card.Name == "Swamp" {
			found = true
		}
	}
	if !found {
		t.Error("the Swamp is not in hand")
	}
}

func TestTdmAbzanMonumentMakesASpiritAsToughAsYourToughest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := pushCatalogPermanent(g, me.ID, "Abzan Monument", "Artifact", tdmAbzanMonumentOracle, false)
	b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 5)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 2)
	advanceToMain(t, g)
	fillPoolColored(me, "W", 1)
	fillPoolColored(me, "B", 1)
	fillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(monument) {
		t.Fatal("the Monument was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	spirits := socB8Named(g, me.ID, "Spirit")
	if len(spirits) != 1 {
		t.Fatalf("%d Spirits, want 1", len(spirits))
	}
	s := spirits[0]
	if s.Power != 5 || s.Toughness != 5 || len(s.Colors) != 1 || s.Colors[0] != "W" {
		t.Errorf("Spirit %d/%d %v, want a 5/5 white Spirit", s.Power, s.Toughness, s.Colors)
	}
}

func TestTdmAbzanMonumentActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := pushCatalogPermanent(g, me.ID, "Abzan Monument", "Artifact", tdmAbzanMonumentOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	fillPoolColored(me, "W", 1)
	fillPoolColored(me, "B", 1)
	fillPoolColored(me, "G", 2)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Monument's ability was activated in combat")
	}
}

// --- Adorned Crocodile ---------------------------------------------------------

func TestTdmAdornedCrocodileDiesIntoAZombieThenRenews(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	croc := b12Push(g, me.ID, "Adorned Crocodile", "Creature — Crocodile", tdmAdornedCrocodileOracle, 5, 3)
	advanceToMain(t, g)
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, croc, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("renew was activated from the battlefield")
	}
	b18Kill(t, g, croc)
	zombies := socB8Named(g, me.ID, "Zombie Druid")
	if len(zombies) != 1 {
		t.Fatalf("%d Zombie Druids, want 1", len(zombies))
	}
	z := zombies[0]
	if z.Power != 2 || z.Toughness != 2 || len(z.Colors) != 1 || z.Colors[0] != "B" {
		t.Errorf("Zombie Druid %d/%d %v, want a 2/2 black", z.Power, z.Toughness, z.Colors)
	}
	if g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		advanceTo(t, g, game.StepPostcombatMain)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, croc, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: z.InstanceID}},
	}); err != nil {
		t.Fatalf("renew from the graveyard: %v", err)
	}
	if !g.Exile.Contains(croc) {
		t.Fatal("renew's cost did not exile the Crocodile")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, z.InstanceID, game.CounterPlusOne); got != 1 {
		t.Errorf("the Zombie has %d +1/+1 counters, want 1", got)
	}
}

// --- Agent of Kotis ------------------------------------------------------------

func TestTdmAgentOfKotisRenewsTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushBattlefieldCreature(g, me.ID)
	id, _ := activateFromGraveyard(t, g, "Agent of Kotis", "Creature — Human Rogue",
		tdmAgentOfKotisOracle, 2, 1, "{U}{U}{U}{U}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Agent")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, target, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
}

func TestTdmAgentOfKotisRenewIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := pushBattlefieldCreature(g, me.ID)
	id := pushCatalogGraveyardCard(me, "Agent of Kotis", "Creature — Human Rogue", tdmAgentOfKotisOracle, 2, 1)
	advanceTo(t, g, game.StepBeginCombat)
	fillPoolColored(me, "U", 4)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Error("renew was activated in combat")
	}
}

// --- Alchemist's Assistant -----------------------------------------------------

func TestTdmAlchemistsAssistantHasLifelinkAndRenewsALifelinkCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monkey := b12Push(g, me.ID, "Alchemist's Assistant", "Creature — Monkey", tdmAlchemistOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, monkey, "lifelink") {
		t.Error("Alchemist's Assistant has no lifelink")
	}
	target := pushBattlefieldCreature(g, me.ID)
	activateFromGraveyard(t, g, "Alchemist's Assistant", "Creature — Monkey",
		tdmAlchemistOracle, 2, 1, "{B}{B}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, target, game.CounterLifelink); got != 1 {
		t.Errorf("%d lifelink counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, target, "lifelink") {
		t.Error("the lifelink counter did not give lifelink")
	}
}

// --- Aegis Sculptor ------------------------------------------------------------

func TestTdmAegisSculptorExilesTwoForACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sculptor := b12Push(g, me.ID, "Aegis Sculptor", "Creature — Bird Wizard", tdmAegisSculptorOracle, 2, 3)
	a := pushGraveyardCard(g, me, "Old One")
	b := pushGraveyardCard(g, me, "Old Two")
	c := pushGraveyardCard(g, me, "Old Three")
	advanceToUpkeepOfSeat(t, g, 0)
	if n := triggersOnStackFrom(g, sculptor); n != 1 {
		t.Fatalf("the upkeep trigger is on the stack %d times, want 1", n)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt to exile two cards")
	}
	answerChooseCards(t, g, me.ID, a, b)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || !me.Graveyard.Contains(c) {
		t.Error("the two chosen cards were not exiled, or the third was")
	}
	if got := countersOn(g, sculptor, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, sculptor, "flying") {
		t.Error("Aegis Sculptor has no flying")
	}
}

func TestTdmAegisSculptorDeclineAndShortGraveyardGiveNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sculptor := b12Push(g, me.ID, "Aegis Sculptor", "Creature — Bird Wizard", tdmAegisSculptorOracle, 2, 3)
	only := pushGraveyardCard(g, me, "Lonely Card")
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked to exile two cards from a graveyard of one")
	}
	if got := countersOn(g, sculptor, game.CounterPlusOne); got != 0 || !me.Graveyard.Contains(only) {
		t.Error("one card in the graveyard still bought a counter")
	}

	pushGraveyardCard(g, me, "Second Card")
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("no prompt with two cards in the graveyard")
	}
	answerChooseCards(t, g, me.ID)
	if got := countersOn(g, sculptor, game.CounterPlusOne); got != 0 {
		t.Errorf("declining still put %d counters on", got)
	}
}

// --- Ainok Wayfarer ------------------------------------------------------------

func TestTdmAinokWayfarerTakesAMilledLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Milled Bolt", TypeLine: "Instant"},
		game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Kept Land", TypeLine: "Basic Land — Plains"},
	)
	wayfarer := castCatalogSpell(t, g, "Ainok Wayfarer", "Creature — Dog Scout", tdmAinokWayfarerOracle, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for a land from among the milled cards")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[0] {
		t.Fatalf("offered %v, want only the milled land", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, ids[0])
	if !me.Hand.Contains(ids[0]) {
		t.Error("the land did not go to hand")
	}
	if !me.Graveyard.Contains(ids[1]) || !me.Graveyard.Contains(ids[2]) {
		t.Error("the other milled cards are not in the graveyard")
	}
	if got := countersOn(g, wayfarer, game.CounterPlusOne); got != 0 {
		t.Errorf("a land was taken and the Wayfarer still has %d counters", got)
	}
}

func TestTdmAinokWayfarerGrowsWithoutALand(t *testing.T) {
	for _, tc := range []struct {
		name    string
		land    bool
		decline bool
	}{
		{"no land milled", false, false},
		{"land declined", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			first := game.Card{Name: "Milled Bolt", TypeLine: "Instant"}
			if tc.land {
				first = game.Card{Name: "Milled Forest", TypeLine: "Basic Land — Forest"}
			}
			seedSearchLibrary(me, first,
				game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear"},
				game.Card{Name: "Milled Ox", TypeLine: "Creature — Ox"},
			)
			wayfarer := castCatalogSpell(t, g, "Ainok Wayfarer", "Creature — Dog Scout", tdmAinokWayfarerOracle, nil)
			passPriorityAroundTable(t, g)
			if tc.decline {
				if chooseCardsChoiceFor(g, me.ID) == nil {
					t.Fatal("no prompt with a land milled")
				}
				answerChooseCards(t, g, me.ID)
			} else if chooseCardsChoiceFor(g, me.ID) != nil {
				t.Fatal("a prompt with no land among the milled cards")
			}
			if got := countersOn(g, wayfarer, game.CounterPlusOne); got != 1 {
				t.Errorf("%d +1/+1 counters, want 1", got)
			}
		})
	}
}

// --- Alesha's Legacy -----------------------------------------------------------

func TestTdmAleshasLegacyGrantsDeathtouchAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if _, err := castFromHandErr(g, me, "Alesha's Legacy", "Instant", "", tdmAleshasLegacyOracle, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("Alesha's Legacy targeted a creature you don't control")
	}
	castCatalogSpell(t, g, "Alesha's Legacy", "Instant", tdmAleshasLegacyOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"deathtouch", "indestructible"} {
		if !hasEffectiveKeyword(t, g, bear, kw) {
			t.Errorf("the target did not gain %s", kw)
		}
	}
}

// --- All-Out Assault -----------------------------------------------------------

func TestTdmAllOutAssaultPumpsAndGivesDeathtouchToYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "All-Out Assault", "Enchantment", tdmAllOutAssaultOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("your Bear is %d/%d, want 3/3", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("your Bear has no deathtouch")
	}
	if effectivePower(t, g, theirs) != 2 || hasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's creature was pumped")
	}
}

func TestTdmAllOutAssaultAddsACombatAndUntapsOnTheNextAttack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	idle := pushVanillaCreature(g, me.ID, "Idle", 2, 2)
	attackWith(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepPostcombatMain)
	tdmTap(g, idle)

	castCatalogSpell(t, g, "All-Out Assault", "Enchantment", tdmAllOutAssaultOracle, nil)
	passPriorityAroundTable(t, g)
	if !isTapped(g, bear) || !isTapped(g, idle) {
		t.Fatal("the enters trigger untapped something before the next attack")
	}
	next := stepOnce(t, g)
	if next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after All-Out Assault's main phase: %+v, want the second combat", next)
	}
	// The Bear is still tapped from the first combat; the Idle one attacks.
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(idle) })
	declareAttack(t, g, opp.ID, idle)
	passPriorityAroundTable(t, g)
	if isTapped(g, idle) || isTapped(g, bear) {
		t.Error("attacking again did not untap each creature you control")
	}
}

// "If it's your main phase" is the controller's own main phase and no
// other step: an opponent's main phase and your combat both fail it.
func TestTdmAllOutAssaultMainPhaseCondition(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	if !allOutAssaultYourMainPhase(g, me.ID) {
		t.Error("your precombat main phase is not your main phase")
	}
	if allOutAssaultYourMainPhase(g, opp.ID) {
		t.Error("your main phase counted as an opponent's")
	}
	advanceTo(t, g, game.StepBeginCombat)
	if allOutAssaultYourMainPhase(g, me.ID) {
		t.Error("the beginning of combat counted as a main phase")
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if !allOutAssaultYourMainPhase(g, me.ID) {
		t.Error("your postcombat main phase is not your main phase")
	}
}

// --- Ambling Stormshell --------------------------------------------------------

func TestTdmAmblingStormshellStunsItselfAndDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	shell := b12Push(g, me.ID, "Ambling Stormshell", "Creature — Turtle", tdmStormshellOracle, 5, 9)
	hand := len(me.Hand.Cards)
	declareAttack(t, g, opp.ID, shell)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, shell, game.CounterStun); got != 3 {
		t.Errorf("%d stun counters, want 3", got)
	}
	if got := len(me.Hand.Cards) - hand; got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
}

func TestTdmAmblingStormshellUntapsOnATurtleSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shell := b12Push(g, me.ID, "Ambling Stormshell", "Creature — Turtle", tdmStormshellOracle, 5, 9)
	advanceToMain(t, g)
	tdmTap(g, shell)
	castCatalogSpell(t, g, "Snapping Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if !isTapped(g, shell) {
		t.Fatal("a non-Turtle spell untapped the Stormshell")
	}
	castCatalogSpell(t, g, "Little Turtle", "Creature — Turtle", "", nil)
	passPriorityAroundTable(t, g)
	if isTapped(g, shell) {
		t.Fatal("a Turtle spell did not untap the Stormshell")
	}

	// With a stun counter, the untap removes the counter instead.
	tdmTap(g, shell)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(shell, game.CounterStun, 2) })
	castCatalogSpell(t, g, "Second Turtle", "Creature — Turtle", "", nil)
	passPriorityAroundTable(t, g)
	if !isTapped(g, shell) || countersOn(g, shell, game.CounterStun) != 1 {
		t.Errorf("a stunned Stormshell: tapped=%v, stun=%d; want tapped with one stun counter",
			isTapped(g, shell), countersOn(g, shell, game.CounterStun))
	}
}

// --- Anafenza, Unyielding Lineage ----------------------------------------------

func TestTdmAnafenzaEnduresWithCountersOrASpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	anafenza := b12Push(g, me.ID, "Anafenza, Unyielding Lineage", "Legendary Creature — Spirit Soldier", tdmAnafenzaOracle, 2, 2)
	for _, kw := range []string{"flash", "first strike"} {
		if !hasEffectiveKeyword(t, g, anafenza, kw) {
			t.Errorf("Anafenza has no %s", kw)
		}
	}
	advanceToMain(t, g)

	// Counters.
	b18Kill(t, g, b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no endure choice when a nontoken creature died")
	}
	answerOptionPick(t, g, me.ID, 1)
	if got := countersOn(g, anafenza, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}

	// The Spirit.
	b18Kill(t, g, b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2))
	answerOptionPick(t, g, me.ID, 0)
	spirits := socB8Named(g, me.ID, "Spirit")
	if len(spirits) != 1 || spirits[0].Power != 2 || spirits[0].Toughness != 2 ||
		len(spirits[0].Colors) != 1 || spirits[0].Colors[0] != "W" {
		t.Fatalf("Spirits %+v, want one 2/2 white Spirit", spirits)
	}
	if got := countersOn(g, anafenza, game.CounterPlusOne); got != 2 {
		t.Errorf("choosing the Spirit changed the counters to %d", got)
	}

	// A token dying, or an opponent's creature, is not a trigger.
	b18Kill(t, g, spirits[0].InstanceID)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b18Kill(t, g, b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("a token or an opponent's creature dying made Anafenza endure")
	}
}

func TestTdmAnafenzaGoneMakesTheSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	anafenza := b12Push(g, me.ID, "Anafenza, Unyielding Lineage", "Legendary Creature — Spirit Soldier", tdmAnafenzaOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	// Both die at once: Anafenza sees the Bear die, and endures from the
	// graveyard, where only the Spirit is possible.
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear)
		_ = g.DestroyPermanentForEffect(anafenza)
	})
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Fatal("asked to choose counters for an Anafenza that is gone")
	}
	if n := len(socB8Named(g, me.ID, "Spirit")); n != 1 {
		t.Errorf("%d Spirits, want 1", n)
	}
}

// --- Attuned Hunter ------------------------------------------------------------

func TestTdmAttunedHunterGrowsWhenCardsLeaveYourGraveyardOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	hunter := b12Push(g, me.ID, "Attuned Hunter", "Creature — Human Ranger", tdmAttunedHunterOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, hunter, "trample") {
		t.Error("Attuned Hunter has no trample")
	}
	a := pushGraveyardCard(g, me, "First")
	b := pushGraveyardCard(g, me, "Second")
	advanceToMainOf(t, g, 0)
	tdmExileFromGraveyard(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, hunter, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d +1/+1 counters after a card left on your turn, want 1", got)
	}
	advanceToMainOf(t, g, 1)
	tdmExileFromGraveyard(t, g, opp.ID, b)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, hunter, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters after a card left on an opponent's turn, want still 1", got)
	}
}

func TestTdmAttunedHunterCountsARenewCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hunter := b12Push(g, me.ID, "Attuned Hunter", "Creature — Human Ranger", tdmAttunedHunterOracle, 3, 3)
	activateFromGraveyard(t, g, "Agent of Kotis", "Creature — Human Rogue",
		tdmAgentOfKotisOracle, 2, 1, "{U}{U}{U}{U}", game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: hunter}},
		})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, hunter, game.CounterPlusOne); got != 3 {
		t.Errorf("%d +1/+1 counters, want 3 (one for the card leaving, two from renew)", got)
	}
}

// --- Avenger of the Fallen -----------------------------------------------------

func TestTdmAvengerOfTheFallenMobilizesOnePerCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	avenger := b12Push(g, me.ID, "Avenger of the Fallen", "Creature — Human Warrior", tdmAvengerOracle, 2, 4)
	if !hasEffectiveKeyword(t, g, avenger, "deathtouch") {
		t.Error("Avenger of the Fallen has no deathtouch")
	}
	for i := 0; i < 3; i++ {
		pushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	}
	pushGraveyardCard(g, me, "Dead Spell")
	declareAttack(t, g, opp.ID, avenger)
	passPriorityAroundTable(t, g)
	tdmPlaceWarriors(t, g, me.ID)
	warriors := socB8Named(g, me.ID, "Warrior")
	if len(warriors) != 3 {
		t.Fatalf("mobilize X made %d Warriors, want 3", len(warriors))
	}
	for _, w := range warriors {
		if !w.Tapped || w.AttackingTarget == uuid.Nil {
			t.Error("a Warrior is not tapped and attacking")
		}
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if n := len(socB8Named(g, me.ID, "Warrior")); n != 0 {
		t.Errorf("%d Warriors survived the end step", n)
	}
}

func TestTdmAvengerOfTheFallenWithNoCreatureCardsMakesNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	avenger := b12Push(g, me.ID, "Avenger of the Fallen", "Creature — Human Warrior", tdmAvengerOracle, 2, 4)
	declareAttack(t, g, opp.ID, avenger)
	passPriorityAroundTable(t, g)
	if n := len(socB8Named(g, me.ID, "Warrior")); n != 0 {
		t.Errorf("%d Warriors with no creature card in the graveyard", n)
	}
}

// --- Awaken the Honored Dead ---------------------------------------------------

func TestTdmAwakenTheHonoredDeadWalksItsChapters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirRock := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	land := pushCatalogPermanent(g, opp.ID, "Their Land", "Land", "", false)

	// I — destroy target nonland permanent.
	saga := castCatalogSpell(t, g, "Awaken the Honored Dead", "Enchantment — Saga", tdmAwakenOracle, nil)
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter I asked for no target")
	}
	if hasID(p.PickTargetCards, land) || !hasID(p.PickTargetCards, theirRock) {
		t.Errorf("chapter I targets %v: a nonland permanent only", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, theirRock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirRock) {
		t.Fatal("chapter I did not destroy its target")
	}

	// II — mill three.
	graveyard := len(me.Graveyard.Cards)
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := len(me.Graveyard.Cards) - graveyard; got != 3 {
		t.Fatalf("chapter II milled %d, want 3", got)
	}

	// III — discard, then return a creature or land card.
	creature := pushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	discard := pushHandCard(g, me)
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter III offered no discard")
	}
	answerDiscard(t, g, me.ID, discard)
	passPriorityAroundTable(t, g)
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger asked for no target")
	}
	pickCard(t, g, me.ID, creature)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) {
		t.Error("the creature card did not return to hand")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after III")
	}
}

func TestTdmAwakenTheHonoredDeadChapterThreeDeclinedReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirRock := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	creature := pushGraveyardCreature(g, me.ID, "Dead Bear", "{1}{G}")
	pushHandCard(g, me)
	castCatalogSpell(t, g, "Awaken the Honored Dead", "Enchantment — Saga", tdmAwakenOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, theirRock)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("chapter III offered no discard")
	}
	answerDiscard(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || me.Hand.Contains(creature) {
		t.Error("declining the discard still returned a card")
	}
}

// --- Bearer of Glory -----------------------------------------------------------

func TestTdmBearerOfGloryFirstStrikeOnYourTurnAndTeamPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bearer := b12Push(g, me.ID, "Bearer of Glory", "Creature — Human Soldier", tdmBearerOfGloryOracle, 2, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 0)
	if !hasEffectiveKeyword(t, g, bearer, "first strike") {
		t.Error("no first strike during your turn")
	}
	fillPoolColored(me, "W", 5)
	if err := g.ActivateCatalogAbility(me.ID, bearer, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 3 {
		t.Errorf("the Bear's power is %d, want 3", p)
	}
	if p := effectivePower(t, g, bearer); p != 3 {
		t.Errorf("the Bearer's power is %d, want 3", p)
	}
	advanceToMainOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bearer, "first strike") {
		t.Error("first strike on an opponent's turn")
	}
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

// --- Bewildering Blizzard ------------------------------------------------------

func TestTdmBewilderingBlizzardDrawsThreeAndWeakensOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	advanceToMain(t, g)
	hand := len(me.Hand.Cards)
	castCatalogSpell(t, g, "Bewildering Blizzard", "Instant", tdmBlizzardOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - hand; got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
	if p, tough := effectivePower(t, g, theirs), effectiveToughness(t, g, theirs); p != 2 || tough != 5 {
		t.Errorf("their Giant is %d/%d, want 2/5", p, tough)
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("your Bear was weakened to %d", p)
	}
}

// --- Bone-Cairn Butcher --------------------------------------------------------

func TestTdmBoneCairnButcherGivesAttackingTokensDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	butcher := b12Push(g, me.ID, "Bone-Cairn Butcher", "Creature — Demon", tdmBoneCairnOracle, 4, 4)
	idle := pushTokens(g, me.ID, TokenCard("1/1 red Warrior"), 1)[0]
	declareAttack(t, g, opp.ID, butcher)
	passPriorityAroundTable(t, g)
	tdmPlaceWarriors(t, g, me.ID)
	var attackers []uuid.UUID
	for _, c := range socB8Named(g, me.ID, "Warrior") {
		if c.InstanceID != idle {
			attackers = append(attackers, c.InstanceID)
		}
	}
	if len(attackers) != 2 {
		t.Fatalf("mobilize 2 made %d Warriors", len(attackers))
	}
	for _, id := range attackers {
		if !hasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Error("an attacking Warrior has no deathtouch")
		}
	}
	if hasEffectiveKeyword(t, g, idle, "deathtouch") {
		t.Error("a token that is not attacking has deathtouch")
	}
	if hasEffectiveKeyword(t, g, butcher, "deathtouch") {
		t.Error("the Butcher, not a token, has deathtouch")
	}
}

// --- Boulderborn Dragon --------------------------------------------------------

func TestTdmBoulderbornDragonSurveilsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	dragon := b12Push(g, me.ID, "Boulderborn Dragon", "Artifact Creature — Dragon", tdmBoulderbornOracle, 3, 3)
	for _, kw := range []string{"flying", "vigilance"} {
		if !hasEffectiveKeyword(t, g, dragon, kw) {
			t.Errorf("Boulderborn Dragon has no %s", kw)
		}
	}
	declareAttack(t, g, opp.ID, dragon)
	passPriorityAroundTable(t, g)
	if surveilChoiceFor(g, me.ID) == nil {
		t.Error("attacking did not surveil 1")
	}
	if isTapped(g, dragon) {
		t.Error("a vigilant Dragon tapped to attack")
	}
}

// --- Breaching Dragonstorm -----------------------------------------------------

func TestTdmBreachingDragonstormCastsTheNonlandCardFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Exiled Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Wanted Spell", TypeLine: "Sorcery", ManaCost: "{6}{R}{R}"},
		game.Card{Name: "Under It", TypeLine: "Sorcery", ManaCost: "{R}"},
	)
	castCatalogSpell(t, g, "Breaching Dragonstorm", "Enchantment", tdmBreachingOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(ids[0]) || !g.Exile.Contains(ids[1]) {
		t.Fatal("the run did not exile the land and the nonland card")
	}
	if !me.Library.Contains(ids[2]) {
		t.Fatal("the run went past the first nonland card")
	}
	c := keywordCastPromptFor(g, me.ID)
	if c == nil || c.MayCastKeyword != game.MayCastKeywordFreeOrHand {
		t.Fatalf("prompt = %+v, want a free cast of a mana value 8 spell", c)
	}
	answerMayCastPrompt(t, g, me.ID, true)
	if !g.Stack.Contains(ids[1]) {
		t.Error("the free cast is not on the stack")
	}
	if !g.Exile.Contains(ids[0]) {
		t.Error("the exiled land left exile")
	}
}

func TestTdmBreachingDragonstormPutsADeclinedOrTooBigCardIntoHand(t *testing.T) {
	for _, tc := range []struct {
		name, cost string
		offered    bool
	}{
		{"declined", "{2}{R}", true},
		{"mana value 9", "{7}{R}{R}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			ids := seedSearchLibrary(me, game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: tc.cost})
			castCatalogSpell(t, g, "Breaching Dragonstorm", "Enchantment", tdmBreachingOracle, nil)
			passPriorityAroundTable(t, g)
			if got := keywordCastPromptFor(g, me.ID) != nil; got != tc.offered {
				t.Fatalf("free cast offered = %v, want %v", got, tc.offered)
			}
			if tc.offered {
				answerMayCastPrompt(t, g, me.ID, false)
			}
			if !me.Hand.Contains(ids[0]) {
				t.Error("the card is not in hand")
			}
		})
	}
}

func TestTdmBreachingDragonstormReturnsWhenADragonEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	storm := pushCatalogPermanent(g, me.ID, "Breaching Dragonstorm", "Enchantment", tdmBreachingOracle, false)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(storm) {
		t.Fatal("a Bear is not a Dragon")
	}
	castCatalogSpell(t, g, "Dragon", "Creature — Dragon", "", nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(storm) || !me.Hand.Contains(storm) {
		t.Error("a Dragon you control entering did not return the enchantment to your hand")
	}
}
