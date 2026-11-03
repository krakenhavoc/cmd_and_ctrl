package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_mutagen_cards_test.go — the Teenage Mutant Ninja Turtles cards that
// make a Mutagen token (#4676, #4457). Each test plays the real catalog
// entry and asserts what its text says.

const (
	tmtMuCommandoOracle   = "b24a87af-407f-4c58-80b4-caab9c65a233"
	tmtMuGenghisOracle    = "1d6d0348-46ad-46ed-ae7b-d7510f0ebe70"
	tmtMuBFFOracle        = "fe1eb8b3-c605-4711-a057-ed3565bfe75c"
	tmtMuWeirdnessOracle  = "170ab932-9d50-4ce7-9a42-08e1edce7e7c"
	tmtMuMutagenManOracle = "472d9a18-19a0-46c0-a4bf-328ebec7ce41"
	tmtMuChainOracle      = "ea9fc868-374b-4acc-be0d-4283907b4524"
	tmtMuOozeSpillOracle  = "c1c5f9da-1396-4406-b803-b64f66b8e49d"
	tmtMuRayFilletOracle  = "698098b8-b8e4-4e88-b13a-6b039579d192"
	tmtMuSewersOracle     = "a3ee8565-f34c-4073-bc94-3fe9ec6bbe8a"
	tmtMuCryptidOracle    = "f570bac8-9987-4963-af02-476d18abc847"
	tmtMuTheOozeOracle    = "82a6fa6f-13c2-4ea3-bec2-c51b56b2783b"
	tmtMuZooOracle        = "107e9ada-4e2e-4032-addd-274bca956621"
)

func TestTmtMutagenEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmtMuCommandoOracle:   "Crustacean Commando",
		tmtMuGenghisOracle:    "Genghis Frog",
		tmtMuBFFOracle:        "Michelangelo, Mutant BFF",
		tmtMuWeirdnessOracle:  "Michelangelo, Weirdness to 11",
		tmtMuMutagenManOracle: "Mutagen Man, Living Ooze",
		tmtMuChainOracle:      "Mutant Chain Reaction",
		tmtMuOozeSpillOracle:  "Ooze Spill",
		tmtMuRayFilletOracle:  "Ray Fillet, Man Ray",
		tmtMuSewersOracle:     "Return to the Sewers",
		tmtMuCryptidOracle:    "Slithering Cryptid",
		tmtMuTheOozeOracle:    "The Ooze",
		tmtMuZooOracle:        "Zoo Escapees",
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

// tmtMuMutagens counts the Mutagen tokens `controller` controls.
func tmtMuMutagens(g *game.Game, controller uuid.UUID) int {
	return tmtB1Count(g, controller, "Mutagen")
}

// tmtMuCrack activates a Mutagen strictly, paying its {1} from the pool,
// aimed at `target`.
func tmtMuCrack(t *testing.T, g *game.Game, p *game.Player, mutagen, target uuid.UUID) {
	t.Helper()
	p.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(p.ID, mutagen, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate the Mutagen: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// The four creatures whose only Mutagen is "when this enters, create a
// Mutagen token" make exactly one, and it works.
func TestTmtMutagenEntersTriggersMakeOneWorkingMutagen(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Crustacean Commando", "Creature — Crab Mutant Soldier", tmtMuCommandoOracle},
		{"Slithering Cryptid", "Creature — Fish Mutant", tmtMuCryptidOracle},
		{"Ray Fillet, Man Ray", "Legendary Creature — Fish Mutant", tmtMuRayFilletOracle},
		{"Michelangelo, Mutant BFF", "Legendary Creature — Mutant Ninja Turtle", tmtMuBFFOracle},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		castCatalogSpell(t, g, c.name, c.typeLine, c.oracle, nil)
		passPriorityAroundTable(t, g)
		ids := tmtB1Named(g, "Mutagen")
		if len(ids) != 1 {
			t.Errorf("%s: %d Mutagen tokens, want 1", c.name, len(ids))
			continue
		}
		tok := tmtB1Card(t, g, ids[0])
		if tok.Controller != me.ID || !tok.IsToken() || !tok.HasSubtype("Mutagen") {
			t.Errorf("%s: the token is %q under %v", c.name, tok.TypeLine, tok.Controller)
		}
		tmtMuCrack(t, g, me, ids[0], bear)
		if got := plusOnes(g, bear); got != 1 {
			t.Errorf("%s: the Mutagen put %d counters, want 1", c.name, got)
		}
	}
}

// --- Zoo Escapees ------------------------------------------------------------

func TestTmtMutagenZooEscapeesMakesOneWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zoo := b27Push(g, me.ID, "Zoo Escapees", "Creature — Boar Rhino", tmtMuZooOracle, "{1}{G}", 2, 2, "G")
	if got := tmtMuMutagens(g, me.ID); got != 0 {
		t.Fatalf("%d Mutagens before it left", got)
	}
	b27Kill(g, zoo)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens after it died, want 1", got)
	}

	other := b27Push(g, me.ID, "Zoo Escapees", "Creature — Boar Rhino", tmtMuZooOracle, "{1}{G}", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(other) })
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 2 {
		t.Errorf("%d Mutagens after the second one was bounced, want 2", got)
	}
}

// --- Genghis Frog ------------------------------------------------------------

func TestTmtMutagenGenghisFrogTriggersOnEnteringItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	frog := castCatalogSpell(t, g, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", tmtMuGenghisOracle, nil)
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, frog, "trample")
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens after the Frog entered, want 1", got)
	}
}

func TestTmtMutagenGenghisFrogTriggersOnAnotherMutantYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", tmtMuGenghisOracle, "{G}{U}", 1, 3, "G", "U")
	// A Mutant of mine: the Frog's trigger and the Commando's own.
	castCatalogSpell(t, g, "Crustacean Commando", "Creature — Crab Mutant Soldier", tmtMuCommandoOracle, nil)
	b27Settle(t, g, me.ID, uuid.Nil)
	if got := tmtMuMutagens(g, me.ID); got != 2 {
		t.Errorf("%d Mutagens after a Mutant entered, want 2", got)
	}
}

func TestTmtMutagenGenghisFrogIgnoresANonMutant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", tmtMuGenghisOracle, "{G}{U}", 1, 3, "G", "U")
	castCatalogSpell(t, g, "Zoo Escapees", "Creature — Boar Rhino", tmtMuZooOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 0 {
		t.Errorf("%d Mutagens after a Boar Rhino entered, want 0", got)
	}
}

// An opponent's Frog does not see my Mutant.
func TestTmtMutagenGenghisFrogIgnoresAMutantItsControllerDoesNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", tmtMuGenghisOracle, "{G}{U}", 1, 3, "G", "U")
	castCatalogSpell(t, g, "Crustacean Commando", "Creature — Crab Mutant Soldier", tmtMuCommandoOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the Frog's controller got %d Mutagens from my Mutant, want 0", got)
	}
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("I have %d Mutagens (the Commando's own), want 1", got)
	}
}

// --- Michelangelo, Mutant BFF ----------------------------------------------

func TestTmtMutagenMichelangeloBFFMakesOneOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mikey := b27Push(g, me.ID, "Michelangelo, Mutant BFF", "Legendary Creature — Mutant Ninja Turtle", tmtMuBFFOracle, "{2}{G}{G}", 4, 4, "G")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mikey, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens after the attack, want 1", got)
	}
}

func TestTmtMutagenMichelangeloBFFLimitsBlockersOnCreaturesWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Michelangelo, Mutant BFF", "Legendary Creature — Mutant Ninja Turtle", tmtMuBFFOracle, "{2}{G}{G}", 4, 4, "G")
	countered := b12Creature(g, me.ID, "Countered Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, game.CounterStun, 1) })
	a := b12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
	b := b12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{countered, plain} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	// The one with a counter (a stun counter: any kind) can be blocked by
	// one creature only.
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: countered}, {Blocker: b, Attacker: countered},
	}), game.BlockReasonTooManyBlockers)
	// The one without a counter can be double-blocked.
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: plain}, {Blocker: b, Attacker: plain},
	}); err != nil {
		t.Errorf("a double block on a creature with no counter was refused: %v", err)
	}
}

// --- Michelangelo, Weirdness to 11 -------------------------------------------

func TestTmtMutagenMichelangeloWeirdnessAddsOneToCountersOnYourCreatures(t *testing.T) {
	g, me, opp := spendTable(t)
	b27Push(g, me.ID, "Michelangelo, Weirdness to 11", "Legendary Creature — Mutant Ninja Turtle", tmtMuWeirdnessOracle, "{1}{G}", 1, 1, "G")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	tmtMuCrack(t, g, me, pushToken(g, me.ID, MutagenToken()), mine)
	if got := plusOnes(g, mine); got != 2 {
		t.Errorf("my creature got %d counters from one Mutagen, want 2", got)
	}
	tmtMuCrack(t, g, me, pushToken(g, me.ID, MutagenToken()), theirs)
	if got := plusOnes(g, theirs); got != 1 {
		t.Errorf("an opponent's creature got %d counters, want 1", got)
	}
}

func TestTmtMutagenMichelangeloWeirdnessMakesAMutagenOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Michelangelo, Weirdness to 11", "Legendary Creature — Mutant Ninja Turtle", tmtMuWeirdnessOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}

// --- Mutagen Man, Living Ooze ------------------------------------------------

func TestTmtMutagenMutagenManMakesXMutagens(t *testing.T) {
	for _, x := range []int{0, 3} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		man := castXSpell(t, g, "Mutagen Man, Living Ooze", "Legendary Creature — Ooze Mutant", tmtMuMutagenManOracle, "{X}{G}{G}", x, nil)
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(man) {
			t.Fatalf("X=%d: Mutagen Man is not on the battlefield", x)
		}
		if got := tmtMuMutagens(g, me.ID); got != x {
			t.Errorf("X=%d: %d Mutagens, want %d", x, got, x)
		}
	}
}

func TestTmtMutagenMutagenManDiscountsArtifactTokensYouControl(t *testing.T) {
	g, me, _ := spendTable(t)
	man := b27Push(g, me.ID, "Mutagen Man, Living Ooze", "Legendary Creature — Ooze Mutant", tmtMuMutagenManOracle, "{X}{G}{G}", 2, 3, "G")
	assertKeywords(t, g, man, "trample")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mutagen := pushToken(g, me.ID, MutagenToken())
	// No mana at all: the {1} is gone.
	if err := g.ActivateCatalogAbility(me.ID, mutagen, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("a Mutagen under Mutagen Man needed mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	// A Clue still needs {1} of its {2}.
	clue := pushToken(g, me.ID, ClueToken())
	if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a Clue was cracked for nothing")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Errorf("a Clue under Mutagen Man would not crack for {1}: %v", err)
	}
}

func TestTmtMutagenMutagenManDoesNotDiscountAnOpponentsToken(t *testing.T) {
	g, me, opp := spendTable(t)
	b27Push(g, me.ID, "Mutagen Man, Living Ooze", "Legendary Creature — Ooze Mutant", tmtMuMutagenManOracle, "{X}{G}{G}", 2, 3, "G")
	theirs := pushToken(g, opp.ID, MutagenToken())
	var q game.ParsedCost
	var err error
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(theirs)
		q, err = g.AbilityManaCostForEffect(opp.ID, c, game.ZoneBattlefield, game.ActivatedAbilitiesForCard(c)[0])
	})
	if err != nil {
		t.Fatalf("AbilityManaCostForEffect: %v", err)
	}
	if q.Generic != 1 {
		t.Errorf("an opponent's Mutagen costs %d generic, want 1", q.Generic)
	}
}

// --- Mutant Chain Reaction ---------------------------------------------------

func TestTmtMutagenMutantChainReactionDestroysAnArtifactAndMakesAMutagen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	castCatalogSpell(t, g, "Mutant Chain Reaction", "Sorcery", tmtMuChainOracle, b16TargetCard(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}

func TestTmtMutagenMutantChainReactionWithNoTargetStillMakesAMutagen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Mutant Chain Reaction", "Sorcery", tmtMuChainOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}

func TestTmtMutagenMutantChainReactionTargetsOnlyFliersAmongCreatures(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ground := b12Creature(g, opp.ID, "Ground Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Mutant Chain Reaction", "Sorcery", tmtMuChainOracle, b16TargetCard(ground)); err == nil {
		t.Error("a creature without flying was accepted as a target")
	}
	flier := brCreature(g, opp.ID, "Bird", "Creature — Bird", 1, 1, "flying")
	castCatalogSpell(t, g, "Mutant Chain Reaction", "Sorcery", tmtMuChainOracle, b16TargetCard(flier))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the flier was not destroyed")
	}
}

// --- Ooze Spill --------------------------------------------------------------

func TestTmtMutagenOozeSpillCountersAndMakesAMutagen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := b13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	castCatalogSpell(t, g, "Ooze Spill", "Instant", tmtMuOozeSpillOracle, b16TargetCard(spell))
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the spell was not countered")
	}
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
	if got := tmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the opponent got %d Mutagens", got)
	}
}

// --- Ray Fillet, Man Ray -----------------------------------------------------

func TestTmtMutagenRayFilletRemovesACounterFromACreatureYouControlToDraw(t *testing.T) {
	g, me, opp := spendTable(t)
	ray := b27Push(g, me.ID, "Ray Fillet, Man Ray", "Legendary Creature — Fish Mutant", tmtMuRayFilletOracle, "{3}{U}", 3, 3, "U")
	assertKeywords(t, g, ray, "flying")
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	activate := func(from uuid.UUID) error {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		return g.ActivateCatalogAbility(me.ID, ray, 0, game.ActivateAbilityParams{
			Strict:           true,
			CounterSourceIDs: []uuid.UUID{from},
		})
	}
	if err := activate(theirs); err == nil {
		t.Error("a counter was removed from an opponent's creature")
	}
	me.ManaPool = nil
	hand := me.Hand.Size()
	if err := activate(bear); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := plusOnes(g, bear); got != 0 {
		t.Errorf("the bear still has %d counters", got)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if err := activate(bear); err == nil {
		t.Error("activated with no +1/+1 counter to remove")
	}
}

// --- Return to the Sewers ----------------------------------------------------

func TestTmtMutagenReturnToTheSewersTucksThenMakesTheCastersMutagen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Return to the Sewers", "Instant", tmtMuSewersOracle, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	c := putInLibraryChoiceFor(g, opp.ID)
	if c == nil || c.LibraryPlacement != game.LibraryPlaceTopOrBottom {
		t.Fatalf("the owner was not asked top or bottom: %+v", c)
	}
	if got := tmtMuMutagens(g, me.ID); got != 0 {
		t.Errorf("the Mutagen was made before the owner answered (%d)", got)
	}
	if err := g.ResolvePutInLibrary(c.ID, opp.ID, nil, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature is still on the battlefield")
	}
	if top, err := opp.Library.Top(); err != nil || top.InstanceID != theirs {
		t.Error("the creature is not on top of its owner's library")
	}
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("the caster has %d Mutagens, want 1", got)
	}
	if got := tmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the owner has %d Mutagens, want 0", got)
	}
}

// --- The Ooze ----------------------------------------------------------------

func TestTmtMutagenTheOozeMakesOnePerCounterOnACreatureThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "The Ooze", "Legendary Artifact", tmtMuTheOozeOracle, "{2}", 0, 0)
	grown := b12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	bare := b12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	stunned := b12Creature(g, me.ID, "Stunned Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(grown, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(stunned, game.CounterStun, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 3)
	})
	for _, id := range []uuid.UUID{bare, stunned, theirs} {
		b27Kill(g, id)
		g.RunStateChecksForTest()
		passPriorityAroundTable(t, g)
	}
	if got := tmtMuMutagens(g, me.ID); got != 0 {
		t.Fatalf("%d Mutagens from creatures without a +1/+1 counter of mine, want 0", got)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(grown) })
	passPriorityAroundTable(t, g)
	if got := tmtMuMutagens(g, me.ID); got != 2 {
		t.Errorf("%d Mutagens after a creature with two counters left, want 2", got)
	}
}

func TestTmtMutagenTheOozeExilesACardFromAGraveyardAndMakesAMutagen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := b27Push(g, me.ID, "The Ooze", "Legendary Artifact", tmtMuTheOozeOracle, "{2}", 0, 0)
	dead := game.NewCard("Their Dead Bear", opp.ID)
	dead.TypeLine = "Creature — Bear"
	opp.Graveyard.PushTop(dead)
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: b16TargetCard(dead.InstanceID),
	}); err != nil {
		t.Fatalf("activate The Ooze: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead.InstanceID) {
		t.Error("the card was not exiled")
	}
	if got := tmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}
