package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b11_cards_test.go — Star Trek, batch 11 (#4973). Each test plays
// the real catalog entry through the engine and asserts what its text
// says.

const (
	trekB11BloodletterOracle = "ae5ee92a-5915-49ce-aa4c-611ab711b2ef"
	trekB11DeathOracle       = "aeec05f1-43ea-4012-82d8-cded5494535d"
	trekB11PraxisOracle      = "028d6812-55db-40df-9288-89c2e478c76c"
	trekB11DeEvolveOracle    = "663e120a-a74f-47fb-aa0a-9d1a75db626e"
	trekB11EjectOracle       = "c06c17c2-203c-4b4a-8564-f29bc41c9eb5"
	trekB11ListenerOracle    = "e6939d1e-6ff8-4559-8526-0e4d0e749e20"
	trekB11ChangOracle       = "72217c2e-4f6b-41da-8de1-bb6892d3c452"
	trekB11GeorgiouOracle    = "31932df4-0be8-45c4-8b22-84119d9c20ba"
)

func TestTrekB11EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB11BloodletterOracle: "D'Ghor Bloodletter",
		trekB11DeathOracle:       "Death Is Irrelevant",
		trekB11PraxisOracle:      "The Destruction of Praxis",
		trekB11DeEvolveOracle:    "De-evolve",
		trekB11EjectOracle:       "Eject the Warp Core",
		trekB11ListenerOracle:    "Fugitive Listener",
		trekB11ChangOracle:       "General Chang, Cold Warrior",
		trekB11GeorgiouOracle:    "Georgiou, Terran Emperor",
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

// trekB11CastCreature puts a creature card with its printed P/T into
// seat 0's hand in the precombat main phase and casts it (no mana gate),
// then passes until the stack is empty or a prompt is owed.
func trekB11CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return id
}

// trekB11Gone reports that a permanent has left the battlefield.
func trekB11Gone(g *game.Game, id uuid.UUID) bool {
	return findBattlefieldCardForTest(g, id) == nil
}

// --- D'Ghor Bloodletter ------------------------------------------------------

func trekB11Bloodletter(t *testing.T, g *game.Game, mode int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	id := trekB11CastCreature(t, g, "D'Ghor Bloodletter", "Creature — Klingon Assassin", trekB11BloodletterOracle, 2, 2)
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("D'Ghor Bloodletter asked for no mode")
	}
	answerMode(t, g, c, me.ID, mode)
	return id
}

func TestTrekB11BloodletterShrinksEveryOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	small := b12Creature(g, opp.ID, "Squirrel", "Creature — Squirrel", 1, 1)
	letter := trekB11Bloodletter(t, g, 0)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if p, tough := effectivePTOf(t, g, letter); p != 2 || tough != 2 {
		t.Errorf("the Bloodletter is %d/%d, want an unchanged 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("your own Bear is %d/%d, want 1/1", p, tough)
	}
	if !trekB11Gone(g, small) {
		t.Error("the 1/1 survived -1/-1")
	}
}

func TestTrekB11BloodletterMakesEachPlayerSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	letter := trekB11Bloodletter(t, g, 1)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, mine)
	answerSacrifice(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if !trekB11Gone(g, mine) || trekB11Gone(g, letter) {
		t.Errorf("you keep the Bloodletter and sacrifice the Bear: bear gone %v, Bloodletter gone %v", trekB11Gone(g, mine), trekB11Gone(g, letter))
	}
	if !trekB11Gone(g, theirs) {
		t.Error("the opponent sacrificed nothing")
	}
	if p, tough := effectivePTOf(t, g, letter); p != 2 || tough != 2 {
		t.Errorf("the sacrifice bullet shrank the Bloodletter to %d/%d", p, tough)
	}
}

// --- Fugitive Listener -------------------------------------------------------

func TestTrekB11ListenerMakesEachOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	trekB11CastCreature(t, g, "Fugitive Listener", "Creature — El-Aurian Scout", trekB11ListenerOracle, 3, 1)
	for i := 0; i < 8; i++ {
		passPriorityAroundTable(t, g)
		answered := false
		for _, p := range g.Seats[1:] {
			if discardChoiceFor(g, p.ID) != nil {
				discardFromHand(t, g, p.ID)
				answered = true
			}
		}
		if !answered {
			break
		}
	}
	if discardChoiceFor(g, me.ID) != nil {
		t.Fatal("the Listener's controller was asked to discard")
	}
	for i, p := range g.Seats[1:] {
		if got := p.Hand.Size(); got != before[i+1]-1 {
			t.Errorf("opponent %d holds %d cards, want %d", i+1, got, before[i+1]-1)
		}
	}
}

// --- General Chang, Cold Warrior --------------------------------------------

func TestTrekB11ChangGrowsEachKlingon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chang := b27Push(g, me.ID, "General Chang, Cold Warrior", "Legendary Creature — Klingon Warrior Officer", trekB11ChangOracle, "{B}", 1, 2, "B")
	klingon := b12Creature(g, me.ID, "Klingon Ally", "Creature — Klingon Warrior", 2, 2)
	human := b12Creature(g, me.ID, "Human", "Creature — Human", 2, 2)
	theirKlingon := b12Creature(g, opp.ID, "Rival Klingon", "Creature — Klingon", 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, chang, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !trekB11Gone(g, fodder) {
		t.Error("the creature was not sacrificed as the cost")
	}
	for id, want := range map[uuid.UUID]int{chang: 1, klingon: 1, human: 0, theirKlingon: 0} {
		if got := plusOneCounters(g, id); got != want {
			t.Errorf("%s has %d +1/+1 counters, want %d", findBattlefieldCardForTest(g, id).Name, got, want)
		}
	}
}

func TestTrekB11ChangNeedsItsMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	chang := b27Push(g, me.ID, "General Chang, Cold Warrior", "Legendary Creature — Klingon Warrior Officer", trekB11ChangOracle, "{B}", 1, 2, "B")
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, chang, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("activated with an empty mana pool")
	}
	if trekB11Gone(g, fodder) {
		t.Error("a refused activation still sacrificed the creature")
	}
}

// --- Eject the Warp Core -----------------------------------------------------

func TestTrekB11EjectShrinksByTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	advanceToMain(t, g)
	b06AddMana(me, "B", "C")
	castFromHandForTest(t, g, me, "Eject the Warp Core", "Instant", "{1}{B}", trekB11EjectOracle,
		game.CastSpellParams{Targets: cardRefs(victim), Strict: true})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, victim); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
}

func TestTrekB11EjectKickedShrinksBySixInstead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Wurm", "Creature — Wurm", 7, 7)
	rock := b12Push(g, me.ID, "Mind Stone", "Artifact", "", 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "B", "C")
	castFromHandForTest(t, g, me, "Eject the Warp Core", "Instant", "{1}{B}", trekB11EjectOracle,
		game.CastSpellParams{Targets: cardRefs(victim), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{rock}, Strict: true})
	passPriorityAroundTable(t, g)
	if !trekB11Gone(g, rock) {
		t.Error("the artifact was not sacrificed for the kicker")
	}
	if p, tough := effectivePTOf(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1 (-6/-6, not -8/-8)", p, tough)
	}
}

func TestTrekB11EjectRefusals(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 5, 5)
	land := b12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	advanceToMain(t, g)
	cast := func(params game.CastSpellParams) error {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Eject the Warp Core", TypeLine: "Instant", ManaCost: "{1}{B}",
				OracleID: trekB11EjectOracle, Owner: me.ID, Controller: me.ID})
		})
		params.Strict = true
		return g.CastSpell(me.ID, id, params)
	}
	if err := cast(game.CastSpellParams{Targets: cardRefs(victim)}); err == nil {
		t.Error("cast with an empty mana pool")
	}
	b06AddMana(me, "B", "C")
	if err := cast(game.CastSpellParams{Targets: cardRefs(land)}); err == nil {
		t.Error("cast at a land")
	}
	if err := cast(game.CastSpellParams{Targets: cardRefs(victim), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("kicked by sacrificing a land")
	}
}

// --- The Destruction of Praxis ----------------------------------------------

func trekB11Lands(g *game.Game, owner uuid.UUID, names ...string) {
	for _, n := range names {
		b12Push(g, owner, n, "Land", "", 0, 0)
	}
}

func TestTrekB11PraxisCountsDifferentlyNamedLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trekB11Lands(g, me.ID, "Swamp", "Swamp", "Island")
	trekB11Lands(g, opp.ID, "Forest", "Plains", "Mountain")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	castCatalogSpell(t, g, "The Destruction of Praxis", "Sorcery", trekB11PraxisOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !trekB11Gone(g, bear) {
		t.Error("the 2/2 survived -2/-2")
	}
	if p, tough := effectivePTOf(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("the Ogre is %d/%d, want 1/1 (X is 2: Swamp and Island)", p, tough)
	}
	if n := b27CountNamed(g, me.ID, "Klingon Warrior"); n != 0 {
		t.Errorf("%d Klingon Warriors with X of 2", n)
	}
}

func TestTrekB11PraxisMakesTwoKlingonsAtSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB11Lands(g, me.ID, "Swamp", "Island", "Forest", "Plains", "Mountain", "Wastes")
	big := b12Creature(g, me.ID, "Titan", "Creature — Giant", 7, 7)
	castCatalogSpell(t, g, "The Destruction of Praxis", "Sorcery", trekB11PraxisOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if p, tough := effectivePTOf(t, g, big); p != 1 || tough != 1 {
		t.Errorf("the Titan is %d/%d, want 1/1", p, tough)
	}
	var warriors []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Klingon Warrior" {
			warriors = append(warriors, layeredCard(t, g, c.InstanceID))
		}
	}
	if len(warriors) != 2 {
		t.Fatalf("%d Klingon Warriors, want 2", len(warriors))
	}
	for _, w := range warriors {
		if !w.IsToken() || w.CurrentPower() != 2 || w.CurrentToughness() != 1 || !game.HasKeyword(&w, "haste") ||
			!w.HasSubtype("Klingon") || !w.HasSubtype("Warrior") {
			t.Errorf("token %q is %d/%d haste=%v, want a 2/1 Klingon Warrior with haste the -6/-6 missed",
				w.TypeLine, w.CurrentPower(), w.CurrentToughness(), game.HasKeyword(&w, "haste"))
		}
	}
}

// --- De-evolve ---------------------------------------------------------------

func trekB11DeEvolve(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	advanceToMain(t, g)
	b06AddMana(me, "B", "C")
	id := castFromHandForTest(t, g, me, "De-evolve", "Enchantment — Aura", "{1}{B}", trekB11DeEvolveOracle,
		game.CastSpellParams{Targets: cardRefs(host), Strict: true})
	passPriorityAroundTable(t, g)
	return id
}

func TestTrekB11DeEvolveMakesASalamanderThatCantBlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Push(g, opp.ID, "Steel Golem", "Artifact Creature — Golem Soldier", "", 5, 5)
	trekB11DeEvolve(t, g, host)
	got := gaLayered(t, g, host)
	if !got.HasSubtype("Salamander") || got.HasSubtype("Golem") || got.HasSubtype("Soldier") {
		t.Errorf("the host is %q, want a Salamander and no other creature type", got.TypeLine)
	}
	if !got.IsArtifact() || !got.IsCreature() {
		t.Error("the host lost a card type")
	}
	if p, tough := effectivePTOf(t, g, host); p != 1 || tough != 5 {
		t.Errorf("the host is %d/%d, want 1/5", p, tough)
	}
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	other := b12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(host, attacker); err == nil {
		t.Error("the enchanted creature blocked")
	}
	if err := g.DeclareBlocker(other, attacker); err != nil {
		t.Errorf("another creature could not block: %v", err)
	}
}

func TestTrekB11DeEvolveThresholdExilesAtEachUpkeep(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cards int
		exile bool
	}{
		{"six cards in your graveyard", 6, false},
		{"seven cards in your graveyard", 7, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			host := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
			trekB11DeEvolve(t, g, host)
			// The Aura went to the battlefield; the graveyard is seeded after.
			for i := 0; i < tc.cards; i++ {
				pushToGraveyard(g, me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
			}
			// An opponent's upkeep is "each upkeep" too.
			advanceToUpkeepOf(t, g, 1)
			passPriorityAroundTable(t, g)
			if got := g.Exile.Contains(host); got != tc.exile {
				t.Errorf("host exiled %v, want %v", got, tc.exile)
			}
		})
	}
}

func TestTrekB11DeEvolveRechecksThresholdOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	trekB11DeEvolve(t, g, host)
	var filler []uuid.UUID
	for i := 0; i < 7; i++ {
		filler = append(filler, pushToGraveyard(g, me, game.Card{Name: "Filler", TypeLine: "Sorcery"}))
	}
	advanceToUpkeepOf(t, g, 1)
	// Seven as it triggers; one card leaves before it resolves (CR 603.4).
	g.WithWriteLock(func() { me.Graveyard.Remove(filler[0]) })
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(host) || trekB11Gone(g, host) {
		t.Error("the trigger exiled the creature with six cards left in the graveyard")
	}
}

// --- Death Is Irrelevant -----------------------------------------------------

func TestTrekB11DeathIsIrrelevantGrantsIndestructibleToArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Death Is Irrelevant", "Enchantment", trekB11DeathOracle, false)
	robot := b12Push(g, me.ID, "Robot", "Artifact Creature — Robot", "", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Push(g, me.ID, "Mind Stone", "Artifact", "", 0, 0)
	theirs := b12Push(g, opp.ID, "Enemy Robot", "Artifact Creature — Robot", "", 2, 2)
	for id, want := range map[uuid.UUID]bool{robot: true, bear: false, rock: false, theirs: false} {
		if got := hasAbility(effectiveAbilities(t, g, id), "indestructible"); got != want {
			t.Errorf("%s indestructible %v, want %v", findBattlefieldCardForTest(g, id).Name, got, want)
		}
	}
}

func TestTrekB11DeathIsIrrelevantAssimilatesAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Death Is Irrelevant", "Enchantment", trekB11DeathOracle, false)
	dead := pushToGraveyard(g, opp, game.Card{Name: "Fallen Knight", TypeLine: "Enchantment Creature — Human Knight", Power: 2, Toughness: 2})
	advanceToEndStepOf(t, g, 0)
	pickTriggerTarget(t, g, me.ID, dead)
	passPriorityAroundTable(t, g)
	got := findBattlefieldCardForTest(g, dead)
	if got == nil {
		t.Fatal("the creature card was not put onto the battlefield")
	}
	if got.Controller != me.ID {
		t.Errorf("it entered under %v, want you", got.Controller)
	}
	if n := plusOneCounters(g, dead); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	v := gaLayered(t, g, dead)
	if !v.IsArtifact() || !v.IsCreature() || !v.IsEnchantment() {
		t.Errorf("it is %q, want an artifact creature that keeps its enchantment type (CR 205.1b)", v.TypeLine)
	}
	if !v.HasSubtype("Borg") || v.HasSubtype("Human") || v.HasSubtype("Knight") {
		t.Errorf("it is %q, want a Borg and no other creature type", v.TypeLine)
	}
	if !hasAbility(effectiveAbilities(t, g, dead), "indestructible") {
		t.Error("the assimilated artifact creature is not indestructible")
	}
	if lc := layeredCard(t, g, dead); lc.CurrentPower() != 3 || lc.CurrentToughness() != 3 {
		t.Errorf("it is %d/%d, want 3/3 with its counter", lc.CurrentPower(), lc.CurrentToughness())
	}
}

func TestTrekB11DeathIsIrrelevantWaitsForYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The opponent controls it; seat 0's end step is not theirs.
	pushCatalogPermanent(g, opp.ID, "Death Is Irrelevant", "Enchantment", trekB11DeathOracle, false)
	dead := pushToGraveyard(g, me, game.Card{Name: "Fallen Knight", TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2})
	advanceToEndStepOf(t, g, 0)
	if latestPickTarget(g, me.ID) != nil || latestPickTarget(g, opp.ID) != nil {
		t.Fatal("it triggered on an opponent's end step")
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("the card left the graveyard on an opponent's end step")
	}
}

// --- Georgiou, Terran Emperor ------------------------------------------------

func trekB11Georgiou(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Georgiou, Terran Emperor", "Legendary Creature — Human Rogue Noble", trekB11GeorgiouOracle, "{4}{B}{B}", 4, 6, "B")
}

func TestTrekB11GeorgiouMakesTenOfficersAtZeroLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB11Georgiou(g, me.ID)
	stSetLife(t, g, me, 3)
	advanceToEndStepOf(t, g, 0)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if me.Eliminated || me.Life != 0 {
		t.Fatalf("eliminated %v at life %d; want in the game at 0", me.Eliminated, me.Life)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
	officers := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Officer" {
			v := layeredCard(t, g, c.InstanceID)
			if v.IsToken() && v.CurrentPower() == 1 && v.CurrentToughness() == 1 && v.HasSubtype("Officer") {
				officers++
			}
		}
	}
	if officers != 10 {
		t.Errorf("%d 1/1 Officer tokens, want 10", officers)
	}
}

func TestTrekB11GeorgiouAboveZeroMakesNoOfficers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB11Georgiou(g, me.ID)
	stSetLife(t, g, me, 10)
	advanceToEndStepOf(t, g, 0)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Life != 7 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d hand %d, want 7 and %d", me.Life, me.Hand.Size(), hand+1)
	}
	if n := b27CountNamed(g, me.ID, "Officer"); n != 0 {
		t.Errorf("%d Officers at 7 life", n)
	}
}

func TestTrekB11GeorgiouLeavingAtZeroLifeLoses(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	georgiou := trekB11Georgiou(g, me.ID)
	stSetLife(t, g, me, 0)
	g.RunStateChecksForTest()
	if me.Eliminated {
		t.Fatal("lost at 0 life with Georgiou on the battlefield")
	}
	b27Kill(g, georgiou)
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Error("still in the game at 0 life after Georgiou left")
	}
}
