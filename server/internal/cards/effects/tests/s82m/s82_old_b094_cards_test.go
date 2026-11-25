package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b094_cards_test.go — Zendikar Rising, batch 094 (#5508). Every test
// plays the real catalog entry through the engine.

const (
	oldB094GrotagBugCatcherOracle    = "1b6bfd22-2964-4274-a566-81964348b003"
	oldB094ResoluteStrikeOracle      = "563e9574-3f60-4558-88cb-aaf97caa7515"
	oldB094SoaringThoughtThiefOracle = "5ed445c1-8a04-4fab-91ef-4d52db95c3b5"
	oldB094NimanaSkitterSneakOracle  = "d3168d01-12ea-4a93-9519-82a3544b5bf6"
	oldB094SeafloorStalkerOracle     = "d3ee7f5d-56b4-4d87-9642-dcd8449dfb45"
	oldB094UmaraMysticOracle         = "e1c33cfe-1d48-463d-83bb-111374505194"
)

func TestOldB094EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB094GrotagBugCatcherOracle, oldB094ResoluteStrikeOracle, oldB094SoaringThoughtThiefOracle,
		oldB094NimanaSkitterSneakOracle, oldB094SeafloorStalkerOracle, oldB094UmaraMysticOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// oldB094FillGraveyard leaves p with n cards in their graveyard, the last one
// arriving by a real zone move, which drops the layer cache.
func oldB094FillGraveyard(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n-1; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
	effectstest.B27Kill(g, effectstest.B12Creature(g, p.ID, "Doomed Bear", "Creature — Bear", 1, 1))
}

// --- Grotag Bug-Catcher -------------------------------------------------------

func TestOldB094GrotagBugCatcherGrowsByThePartyWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bug := effectstest.B12Push(g, me.ID, "Grotag Bug-Catcher", "Creature — Goblin Warrior", oldB094GrotagBugCatcherOracle, 1, 2)
	effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, bug, "trample") {
		t.Error("no trample")
	}
	effectstest.FinABAttack(t, g, opp.ID, bug)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bug); p != 3 {
		t.Errorf("the Bug-Catcher is %d power attacking with a party of two, want 3", p)
	}
}

func TestOldB094GrotagBugCatcherAloneIsAPartyOfOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	bug := effectstest.B12Push(g, me.ID, "Grotag Bug-Catcher", "Creature — Goblin Warrior", oldB094GrotagBugCatcherOracle, 1, 2)
	effectstest.FinABAttack(t, g, opp.ID, bug)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bug); p != 2 {
		t.Errorf("the Bug-Catcher is %d power attacking alone, want 2", p)
	}
}

// --- Resolute Strike ----------------------------------------------------------

func TestOldB094ResoluteStrikePumpsAndMayAttachToAWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	warrior := effectstest.B12Creature(g, me.ID, "Fighter", "Creature — Human Warrior", 2, 2)
	sword := effectstest.FinABEquipment(g, me.ID, "Sword")
	effectstest.CastCatalogSpell(t, g, "Resolute Strike", "Instant", oldB094ResoluteStrikeOracle, effectstest.CardRefs(warrior))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, warrior); p != 4 || tough != 4 {
		t.Errorf("the Warrior is %d/%d, want 4/4", p, tough)
	}
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no attach prompt for a Warrior: %+v", g.PendingChoices)
	}
	if c.ChooseMin != 0 || c.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want a \"may\" of one", c.ChooseMin, c.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, sword)
	effectstest.PassPriorityAroundTable(t, g)
	if eq, _ := g.LookupCardForEffect(sword); !eq.IsAttachedTo(warrior) {
		t.Error("the Equipment did not attach to the Warrior")
	}
}

func TestOldB094ResoluteStrikeMayDeclineTheAttach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	warrior := effectstest.B12Creature(g, me.ID, "Fighter", "Creature — Human Warrior", 2, 2)
	sword := effectstest.FinABEquipment(g, me.ID, "Sword")
	effectstest.CastCatalogSpell(t, g, "Resolute Strike", "Instant", oldB094ResoluteStrikeOracle, effectstest.CardRefs(warrior))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if eq, _ := g.LookupCardForEffect(sword); eq.IsAttachedTo(warrior) {
		t.Error("declining still attached the Equipment")
	}
}

func TestOldB094ResoluteStrikeOnANonWarriorOnlyPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FinABEquipment(g, me.ID, "Sword")
	effectstest.CastCatalogSpell(t, g, "Resolute Strike", "Instant", oldB094ResoluteStrikeOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("a non-Warrior was offered an Equipment")
	}
}

// --- Soaring Thought-Thief ----------------------------------------------------

func TestOldB094SoaringThoughtThiefPumpsRoguesOnlyWithEightInAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	thief := effectstest.B12Push(g, me.ID, "Soaring Thought-Thief", "Creature — Human Rogue", oldB094SoaringThoughtThiefOracle, 1, 3)
	other := effectstest.B12Creature(g, me.ID, "Pickpocket", "Creature — Human Rogue", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 2, 2)
	for _, kw := range []string{"flash", "flying"} {
		if !effectstest.HasEffectiveKeyword(t, g, thief, kw) {
			t.Errorf("no %s", kw)
		}
	}
	oldB094FillGraveyard(g, opp, 7)
	if p := effectstest.EffectivePower(t, g, thief); p != 1 {
		t.Errorf("with seven cards the Thief is %d power, want 1", p)
	}
	oldB094FillGraveyard(g, opp, 1)
	for id, want := range map[uuid.UUID]int{thief: 2, other: 3, bear: 2, theirs: 2} {
		if p := effectstest.EffectivePower(t, g, id); p != want {
			t.Errorf("with eight cards a creature is %d power, want %d", p, want)
		}
	}
}

func TestOldB094SoaringThoughtThiefMillsEachOpponentOncePerAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	thief := effectstest.B12Push(g, me.ID, "Soaring Thought-Thief", "Creature — Human Rogue", oldB094SoaringThoughtThiefOracle, 1, 3)
	other := effectstest.B12Creature(g, me.ID, "Pickpocket", "Creature — Human Rogue", 2, 2)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Library.Size()
	}
	effectstest.FinABAttack(t, g, opp.ID, thief, other)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := 0
		if p.ID != me.ID {
			want = 2
		}
		if got := before[p.ID] - p.Library.Size(); got != want {
			t.Errorf("%s milled %d, want %d (one trigger for two Rogues)", p.Name, got, want)
		}
	}
}

func TestOldB094SoaringThoughtThiefIgnoresANonRogueAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	effectstest.B12Push(g, me.ID, "Soaring Thought-Thief", "Creature — Human Rogue", oldB094SoaringThoughtThiefOracle, 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	library := opp.Library.Size()
	effectstest.FinABAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != library {
		t.Error("a non-Rogue attack milled the opponent")
	}
}

// --- Nimana Skitter-Sneak -----------------------------------------------------

func TestOldB094NimanaSkitterSneakGrowsAndGainsMenaceWithEightInAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
	sneak := effectstest.B12Push(g, me.ID, "Nimana Skitter-Sneak", "Creature — Human Rogue", oldB094NimanaSkitterSneakOracle, 3, 4)
	oldB094FillGraveyard(g, opp, 7)
	if p := effectstest.EffectivePower(t, g, sneak); p != 3 || effectstest.HasEffectiveKeyword(t, g, sneak, "menace") {
		t.Errorf("with seven cards: power %d, menace %v; want 3 and no menace", p, effectstest.HasEffectiveKeyword(t, g, sneak, "menace"))
	}
	oldB094FillGraveyard(g, opp, 1)
	if p := effectstest.EffectivePower(t, g, sneak); p != 4 || !effectstest.HasEffectiveKeyword(t, g, sneak, "menace") {
		t.Errorf("with eight cards: power %d, menace %v; want 4 and menace", p, effectstest.HasEffectiveKeyword(t, g, sneak, "menace"))
	}
}

func TestOldB094NimanaSkitterSneakIgnoresItsOwnersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	sneak := effectstest.B12Push(g, me.ID, "Nimana Skitter-Sneak", "Creature — Human Rogue", oldB094NimanaSkitterSneakOracle, 3, 4)
	oldB094FillGraveyard(g, me, 10)
	if p := effectstest.EffectivePower(t, g, sneak); p != 3 || effectstest.HasEffectiveKeyword(t, g, sneak, "menace") {
		t.Errorf("your own graveyard turned it on: power %d", p)
	}
}

// --- Seafloor Stalker ---------------------------------------------------------

func TestOldB094SeafloorStalkerCostsLessPerPartyMember(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	stalker := effectstest.PushCatalogPermanent(g, me.ID, "Seafloor Stalker", "Creature — Merfolk Rogue", oldB094SeafloorStalkerOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {3}{U} with three mana and a party of one")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, stalker); p != 2 {
		t.Errorf("the Stalker is %d power, want 2 (1/1 stand-in +1)", p)
	}
	effectstest.AssertRestrictions(t, g, stalker, game.CantBeBlocked)
}

func TestOldB094SeafloorStalkerFullPartyPaysOnlyBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	stalker := effectstest.PushCatalogPermanent(g, me.ID, "Seafloor Stalker", "Creature — Merfolk Rogue", oldB094SeafloorStalkerOracle, false)
	effectstest.B12Creature(g, me.ID, "Healer", "Creature — Human Cleric", 1, 1)
	effectstest.B12Creature(g, me.ID, "Fighter", "Creature — Human Warrior", 1, 1)
	effectstest.B12Creature(g, me.ID, "Mage", "Creature — Human Wizard", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")
	effectstest.B16Activate(t, g, me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertRestrictions(t, g, stalker, game.CantBeBlocked)
}

// --- Umara Mystic -------------------------------------------------------------

func TestOldB094UmaraMysticGrowsOnInstantSorceryAndWizardSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	mystic := effectstest.B12Push(g, me.ID, "Umara Mystic", "Creature — Merfolk Wizard", oldB094UmaraMysticOracle, 1, 3)
	if !effectstest.HasEffectiveKeyword(t, g, mystic, "flying") {
		t.Error("no flying")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastHoldCreature(t, g, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mystic); p != 1 {
		t.Errorf("a Bear spell pumped the Mystic to %d", p)
	}
	effectstest.CastHoldCreature(t, g, "Apprentice", "Creature — Human Wizard", "", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mystic); p != 3 {
		t.Errorf("a Wizard spell left the Mystic at %d, want 3", p)
	}
	effectstest.CastCatalogSpell(t, g, "Resolute Strike", "Instant", oldB094ResoluteStrikeOracle, effectstest.CardRefs(mystic))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, mystic); p != 7 {
		t.Errorf("an instant spell left the Mystic at %d, want 7 (+2 trigger, +2 spell)", p)
	}
}
