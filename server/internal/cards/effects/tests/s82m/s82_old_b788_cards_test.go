package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b788_cards_test.go — S82 OLD batch 788 (#5631, Visions: enchantments,
// activated abilities). Each test plays the real catalog entry.

const (
	oldB788Blanket    = "10e95489-a94d-4523-964c-ec9753103a62"
	oldB788City       = "671458ac-7441-4332-9e59-a42ec8b62f31"
	oldB788Equipoise  = "aab905ef-b945-4523-b26f-fb5e6932bb56"
	oldB788RightWar   = "acaab3fe-8b12-4f11-9714-7ca278674975"
	oldB788DreamTides = "d9583aab-3eff-4d88-84d5-7219ab982636"
	oldB788SnakeBask  = "12236231-7c6e-4edc-87dc-737115843185"
	oldB788Kaleido    = "bffa9256-aff4-476f-86f4-4f4c9e57fa69"
	oldB788Guiding    = "c474b52f-2f50-416e-b6c2-07043c7b339d"
	oldB788Quirion    = "cdc0621a-5a88-46a0-90fa-87ba1e7e1eb2"
	oldB788Squandered = "620d0572-6d22-43e2-83a9-000e605c8a2b"
)

func TestOldB788EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB788Blanket: "Blanket of Night", oldB788City: "City of Solitude", oldB788Equipoise: "Equipoise",
		oldB788RightWar: "Righteous War", oldB788DreamTides: "Dream Tides", oldB788SnakeBask: "Snake Basket",
		oldB788Kaleido: "Diamond Kaleidoscope", oldB788Guiding: "Guiding Spirit", oldB788Quirion: "Quirion Druid",
		oldB788Squandered: "Squandered Resources",
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

func oldB788HasMana(c game.Card, produced string) bool {
	for _, ab := range game.ManaAbilitiesForCard(c) {
		if ab.Produced == produced {
			return true
		}
	}
	return false
}

// --- Blanket of Night --------------------------------------------------------

func TestOldB788BlanketMakesEveryLandASwampInAdditionAndLeavesCleanly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	forest := effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	theirs := effectstest.SeedManaLand(g, opp.ID, "Island", "Basic Land — Island", "U")
	blanket := effectstest.B12Push(g, me.ID, "Blanket of Night", "Enchantment", oldB788Blanket, 0, 0)
	for _, id := range []uuid.UUID{forest, theirs} {
		c := effectstest.LayeredCard(t, g, id)
		if !c.HasSubtype("Swamp") {
			t.Errorf("%s is no Swamp under Blanket of Night: %v", c.Name, c.Effective().Subtypes)
		}
		if !oldB788HasMana(c, "{B}") {
			t.Errorf("%s does not tap for {B}", c.Name)
		}
	}
	if c := effectstest.LayeredCard(t, g, forest); !c.HasSubtype("Forest") || !oldB788HasMana(c, "{G}") {
		t.Error("the Forest lost its own land type or its {G}")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(blanket) })
	if c := effectstest.LayeredCard(t, g, forest); c.HasSubtype("Swamp") {
		t.Error("the Forest is still a Swamp after Blanket of Night left")
	}
}

// --- City of Solitude --------------------------------------------------------

func oldB788GuidingSpirit(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, "Guiding Spirit", "Creature — Angel Spirit", oldB788Guiding, 1, 2)
}

func oldB788ActivateSpirit(g *game.Game, who, spirit, target uuid.UUID) error {
	return g.ActivateCatalogAbility(who, spirit, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target}},
	})
}

func TestOldB788CityOfSolitudeLimitsActivationsToTheActivatorsOwnTurn(t *testing.T) {
	// Control: with no City the non-active player can activate at instant speed.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	spirit := oldB788GuidingSpirit(g, opp.ID)
	if err := oldB788ActivateSpirit(g, opp.ID, spirit, me.ID); err != nil {
		t.Fatalf("control: the opponent could not activate on my turn: %v", err)
	}

	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "City of Solitude", "Enchantment", oldB788City, 0, 0)
	theirs := oldB788GuidingSpirit(g, opp.ID)
	mine := oldB788GuidingSpirit(g, me.ID)
	if err := oldB788ActivateSpirit(g, opp.ID, theirs, me.ID); err == nil {
		t.Error("the opponent activated an ability on my turn under City of Solitude")
	}
	if err := oldB788ActivateSpirit(g, me.ID, mine, opp.ID); err != nil {
		t.Errorf("I could not activate on my own turn: %v", err)
	}
}

func TestOldB788CityOfSolitudeReachesManaAbilitiesAndFollowsTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "City of Solitude", "Enchantment", oldB788City, 0, 0)
	mineLand := effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	theirLand := effectstest.SeedManaLand(g, opp.ID, "Island", "Basic Land — Island", "U")
	if err := g.ActivateManaAbility(opp.ID, theirLand, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("the opponent tapped a land for mana on my turn under City of Solitude")
	}
	if err := g.ActivateManaAbility(me.ID, mineLand, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("I could not tap a land on my own turn: %v", err)
	}
	// On the opponent's turn it reverses.
	effectstest.AangAdvanceToMain(t, g, 1)
	spirit := oldB788GuidingSpirit(g, me.ID)
	theirSpirit := oldB788GuidingSpirit(g, opp.ID)
	if err := oldB788ActivateSpirit(g, me.ID, spirit, opp.ID); err == nil {
		t.Error("I activated an ability on the opponent's turn under City of Solitude")
	}
	if err := oldB788ActivateSpirit(g, opp.ID, theirSpirit, me.ID); err != nil {
		t.Errorf("the opponent could not activate on their own turn: %v", err)
	}
}

// --- Righteous War -----------------------------------------------------------

func TestOldB788RighteousWarGivesEachColourItsProtection(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Righteous War", "Enchantment", oldB788RightWar, 0, 0)
	mk := func(owner uuid.UUID, name string, colors ...string) uuid.UUID {
		id := effectstest.B12Creature(g, owner, name, "Creature — Bear", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Colors = colors
				}
			}
		})
		return id
	}
	white := mk(me.ID, "White", "W")
	black := mk(me.ID, "Black", "B")
	both := mk(me.ID, "Both", "W", "B")
	green := mk(me.ID, "Green", "G")
	theirWhite := mk(opp.ID, "Their White", "W")
	has := func(id uuid.UUID, kw string) bool {
		c := effectstest.LayeredCard(t, g, id)
		return game.HasKeyword(&c, kw)
	}
	for _, tc := range []struct {
		id        uuid.UUID
		name      string
		wantBlack bool
		wantWhite bool
	}{
		{white, "white", true, false}, {black, "black", false, true}, {both, "white and black", true, true},
		{green, "green", false, false}, {theirWhite, "an opponent's white", false, false},
	} {
		if got := has(tc.id, "protection from black"); got != tc.wantBlack {
			t.Errorf("%s creature: protection from black = %v, want %v", tc.name, got, tc.wantBlack)
		}
		if got := has(tc.id, "protection from white"); got != tc.wantWhite {
			t.Errorf("%s creature: protection from white = %v, want %v", tc.name, got, tc.wantWhite)
		}
	}
}

// --- Snake Basket ------------------------------------------------------------

func TestOldB788SnakeBasketMakesXSnakesAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	basket := effectstest.B12Push(g, me.ID, "Snake Basket", "Artifact", oldB788SnakeBask, 0, 0)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, basket, 0, game.ActivateAbilityParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	snakes := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Snake" && c.Controller == me.ID && c.IsToken() && c.IsCreature() {
			snakes++
			if c.Power != 1 || c.Toughness != 1 || !c.HasColor("G") {
				t.Errorf("Snake is %d/%d colours %v, want a 1/1 green", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if snakes != 3 {
		t.Errorf("%d Snakes, want 3", snakes)
	}
	if g.Battlefield.Contains(basket) {
		t.Error("the Basket was not sacrificed")
	}
}

func TestOldB788SnakeBasketRefusesAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	basket := effectstest.B12Push(g, me.ID, "Snake Basket", "Artifact", oldB788SnakeBask, 0, 0)
	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, basket, 0, game.ActivateAbilityParams{Strict: true, XValue: 3}); err == nil {
		t.Error("Snake Basket activated on an opponent's turn")
	}
}

// --- Dream Tides -------------------------------------------------------------

func oldB788Tapped(g *game.Game, ids ...uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			for _, id := range ids {
				if g.Battlefield.Cards[i].InstanceID == id {
					g.Battlefield.Cards[i].Tapped = true
				}
			}
		}
	})
}

func oldB788ChooseCards(t *testing.T, g *game.Game, picks ...uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceChooseCards {
				if err := g.ResolveChooseCards(c.ID, c.Chooser, picks); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				return c
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no choose_cards prompt")
	return nil
}

func oldB788Tap(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Tapped
}

func TestOldB788DreamTidesKeepsCreaturesTappedUntilPaidFor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Dream Tides", "Enchantment", oldB788DreamTides, 0, 0)
	bearA := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	bearB := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	bearC := effectstest.B12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	elf := effectstest.B12Creature(g, me.ID, "Green Elf", "Creature — Elf", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elf {
				g.Battlefield.Cards[i].Colors = []string{"G"}
			}
		}
	})
	oldB788Tapped(g, bearA, bearB, bearC, elf)
	effectstest.B12ToMyNextUpkeep(t, g)
	// The untap step did not untap any creature.
	for _, id := range []uuid.UUID{bearA, bearB, bearC, elf} {
		if !oldB788Tap(t, g, id) {
			t.Fatalf("a creature untapped in the untap step")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The green creature is not offered.
	var offered []uuid.UUID
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			offered = c.ChooseCards
		}
	}
	for _, id := range offered {
		if id == elf {
			t.Error("the green creature was offered")
		}
	}
	if len(offered) != 3 {
		t.Errorf("%d creatures offered, want the three nongreen ones", len(offered))
	}
	effectstest.FloatForTest(g, me, "CCCC")
	oldB788ChooseCards(t, g, bearA, bearB)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB788Tap(t, g, bearA) || oldB788Tap(t, g, bearB) {
		t.Error("a paid-for creature stayed tapped")
	}
	if !oldB788Tap(t, g, bearC) || !oldB788Tap(t, g, elf) {
		t.Error("an unchosen creature untapped")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left in the pool, want {4} spent for two creatures", me.ManaPool.Total())
	}
}

func TestOldB788DreamTidesDecliningLeavesEverythingTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Dream Tides", "Enchantment", oldB788DreamTides, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	oldB788Tapped(g, bear)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FloatForTest(g, me, "CC")
	oldB788ChooseCards(t, g, bear)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB788Tap(t, g, bear) {
		t.Error("the creature untapped without payment")
	}
}

// --- Equipoise ---------------------------------------------------------------

func oldB788Phased(g *game.Game, id uuid.UUID) bool {
	return g.PhasedOut.Contains(id)
}

func TestOldB788EquipoisePhasesOutTheExcessLandsAndArtifactsAndCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Equipoise", "Enchantment", oldB788Equipoise, 0, 0)
	effectstest.B12Permanent(g, me.ID, "My Island", "Basic Land — Island")
	var lands []uuid.UUID
	for i := 0; i < 3; i++ {
		lands = append(lands, effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island"))
	}
	// Artifacts: they have two, I have none, so both go without a choice.
	art1 := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	art2 := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	// Creatures: they have three, I have two, so one is chosen.
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "My Cat", "Creature — Cat", 2, 2)
	var theirs []uuid.UUID
	for i := 0; i < 3; i++ {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// Lands: two excess, I choose two of three.
	oldB788ChooseCards(t, g, lands[0], lands[2])
	effectstest.PassPriorityAroundTable(t, g)
	// Creatures: one excess, I choose one of three.
	oldB788ChooseCards(t, g, theirs[1])
	effectstest.PassPriorityAroundTable(t, g)
	wantPhased := map[uuid.UUID]bool{lands[0]: true, lands[2]: true, art1: true, art2: true, theirs[1]: true}
	for _, id := range append(append(append([]uuid.UUID{}, lands...), art1, art2), theirs...) {
		if got := oldB788Phased(g, id); got != wantPhased[id] {
			t.Errorf("%s: phased out = %v, want %v", id, got, wantPhased[id])
		}
	}
}

func TestOldB788EquipoiseDoesNothingWhenIControlAsMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Equipoise", "Enchantment", oldB788Equipoise, 0, 0)
	effectstest.B12Permanent(g, me.ID, "My Island", "Basic Land — Island")
	land := effectstest.B12Permanent(g, opp.ID, "Their Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PickPlayerTarget(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	// They have a creature and I have none, so only the creature goes.
	if oldB788Phased(g, land) {
		t.Error("an equal number of lands still phased one out")
	}
	if !oldB788Phased(g, bear) {
		t.Error("the creature I have no match for did not phase out")
	}
}

// --- Guiding Spirit ----------------------------------------------------------

func TestOldB788GuidingSpiritMovesATopCreatureCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	spirit := oldB788GuidingSpirit(g, me.ID)
	creature := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	if err := oldB788ActivateSpirit(g, me.ID, spirit, opp.ID); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(creature) || !opp.Library.Contains(creature) {
		t.Fatal("the top creature card did not go to the top of its owner's library")
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != creature {
		t.Error("the creature card is not on top of the library")
	}
	// Not a creature on top: nothing moves.
	spirit2 := oldB788GuidingSpirit(g, me.ID)
	below := uuid.New()
	land := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: below, Name: "Dead Bear 2", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Dead Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	if err := oldB788ActivateSpirit(g, me.ID, spirit2, opp.ID); err != nil {
		t.Fatalf("second activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(land) || !opp.Graveyard.Contains(below) {
		t.Error("a graveyard whose top card is a land lost a card")
	}
}

func TestOldB788GuidingSpiritWithAnEmptyGraveyardDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	spirit := oldB788GuidingSpirit(g, me.ID)
	before := opp.Library.Size()
	if err := oldB788ActivateSpirit(g, me.ID, spirit, opp.ID); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != before {
		t.Error("the library changed")
	}
}

// --- Quirion Druid -----------------------------------------------------------

func TestOldB788QuirionDruidMakesALandA22GreenCreatureIndefinitely(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMain(t, g)
	druid := effectstest.B12Push(g, me.ID, "Quirion Druid", "Creature — Elf Druid", oldB788Quirion, 1, 2)
	land := effectstest.SeedManaLand(g, opp.ID, "Island", "Basic Land — Island", "U")
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	check := func(when string) {
		t.Helper()
		c := effectstest.LayeredCard(t, g, land)
		if !c.IsCreature() || !c.IsLand() || c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.HasColor("G") {
			t.Errorf("%s: the land is creature=%v land=%v %d/%d colours %v, want a 2/2 green creature land",
				when, c.IsCreature(), c.IsLand(), c.CurrentPower(), c.CurrentToughness(), c.EffectiveColors())
		}
	}
	check("on resolution")
	effectstest.AangAdvanceToMain(t, g, 1)
	check("on the next turn")
	effectstest.AangAdvanceToMain(t, g, 0)
	check("two turns later")
}

// --- Diamond Kaleidoscope ----------------------------------------------------

func TestOldB788DiamondKaleidoscopeMakesPrismsAndSacrificesThemForMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	kal := effectstest.B12Push(g, me.ID, "Diamond Kaleidoscope", "Artifact", oldB788Kaleido, 0, 0)
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, kal, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("make a Prism: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var prism uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Prism" && c.Controller == me.ID {
			prism = c.InstanceID
			if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || c.Power != 0 || c.Toughness != 1 || !c.IsColorless() {
				t.Errorf("Prism is %q %d/%d, want a 0/1 colorless artifact creature token", c.TypeLine, c.Power, c.Toughness)
			}
		}
	}
	if prism == uuid.Nil {
		t.Fatal("no Prism token")
	}
	// A non-Prism cannot pay the sacrifice.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, kal, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Colors: []string{"R"}}); err == nil {
		t.Error("sacrificed a Bear to the Kaleidoscope")
	}
	if err := g.ActivateManaAbility(me.ID, kal, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{prism}, Colors: []string{"R"}}); err != nil {
		t.Fatalf("sacrifice the Prism: %v", err)
	}
	if g.Battlefield.Contains(prism) {
		t.Error("the Prism is still on the battlefield")
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want {R}", got)
	}
}

// --- Squandered Resources ----------------------------------------------------

func oldB788Squander(t *testing.T, g *game.Game, me *game.Player, enchant, land uuid.UUID, color string) error {
	t.Helper()
	params := game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{land}}
	if color != "" {
		params.Colors = []string{color}
	}
	return g.ActivateManaAbility(me.ID, enchant, 0, params)
}

func TestOldB788SquanderedResourcesAddsATypeTheSacrificedLandCouldProduce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	sq := effectstest.B12Push(g, me.ID, "Squandered Resources", "Enchantment", oldB788Squandered, 0, 0)
	forest := effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	if err := oldB788Squander(t, g, me, sq, forest, ""); err != nil {
		t.Fatalf("sacrifice a Forest: %v", err)
	}
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want {G}", got)
	}
}

func TestOldB788SquanderedResourcesOffersEveryTypeAMultiColourLandCouldProduce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	sq := effectstest.B12Push(g, me.ID, "Squandered Resources", "Enchantment", oldB788Squandered, 0, 0)
	dual := effectstest.SeedManaLand(g, me.ID, "Tundra", "Land — Plains Island", "W", "U")
	// A colour the land cannot produce is refused.
	if err := oldB788Squander(t, g, me, sq, dual, "R"); err == nil {
		t.Fatal("took {R} from a land that cannot produce it")
	}
	if !g.Battlefield.Contains(dual) {
		t.Fatal("a refused activation still sacrificed the land")
	}
	if err := oldB788Squander(t, g, me, sq, dual, "U"); err != nil {
		t.Fatalf("take {U}: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want {U}", got)
	}
}

func TestOldB788SquanderedResourcesReadsTheLandAsItLastExistedUnderBlanketOfNight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	sq := effectstest.B12Push(g, me.ID, "Squandered Resources", "Enchantment", oldB788Squandered, 0, 0)
	effectstest.B12Push(g, me.ID, "Blanket of Night", "Enchantment", oldB788Blanket, 0, 0)
	forest := effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	if err := oldB788Squander(t, g, me, sq, forest, "B"); err != nil {
		t.Fatalf("take {B} from a Forest that is also a Swamp: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want {B}", got)
	}
}

func TestOldB788SquanderedResourcesRefusesANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	sq := effectstest.B12Push(g, me.ID, "Squandered Resources", "Enchantment", oldB788Squandered, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB788Squander(t, g, me, sq, bear, "G"); err == nil {
		t.Error("sacrificed a creature to Squandered Resources")
	}
}

func TestOldB788SquanderedResourcesAddsNothingForALandThatProducesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	sq := effectstest.B12Push(g, me.ID, "Squandered Resources", "Enchantment", oldB788Squandered, 0, 0)
	waste := effectstest.B12Permanent(g, me.ID, "Barren Plot", "Land")
	if err := g.ActivateManaAbility(me.ID, sq, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{waste}}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	if g.Battlefield.Contains(waste) || me.ManaPool.Total() != 0 {
		t.Errorf("land on battlefield=%v, pool %v; want it sacrificed for no mana", g.Battlefield.Contains(waste), effectstest.PoolColors(me))
	}
}

func TestOldB788SquanderedResourcesIsOfferedOnlyWithALandToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMain(t, g)
	sq := effectstest.B12Push(g, me.ID, "Squandered Resources", "Enchantment", oldB788Squandered, 0, 0)
	addsNoMana := func() bool {
		var out bool
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(sq)
			out = game.ManaAbilityAddsNoMana(g, me.ID, sq, game.ManaAbilitiesForCard(c)[0])
		})
		return out
	}
	if !addsNoMana() {
		t.Error("offered with no land to sacrifice")
	}
	effectstest.SeedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	if addsNoMana() {
		t.Error("not offered with a Forest to sacrifice")
	}
}
