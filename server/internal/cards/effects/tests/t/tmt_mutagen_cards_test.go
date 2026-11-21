package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTmtMutagenEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.TmtMuCommandoOracle:   "Crustacean Commando",
		effectstest.TmtMuGenghisOracle:    "Genghis Frog",
		effectstest.TmtMuBFFOracle:        "Michelangelo, Mutant BFF",
		effectstest.TmtMuWeirdnessOracle:  "Michelangelo, Weirdness to 11",
		effectstest.TmtMuMutagenManOracle: "Mutagen Man, Living Ooze",
		effectstest.TmtMuChainOracle:      "Mutant Chain Reaction",
		effectstest.TmtMuOozeSpillOracle:  "Ooze Spill",
		effectstest.TmtMuRayFilletOracle:  "Ray Fillet, Man Ray",
		effectstest.TmtMuSewersOracle:     "Return to the Sewers",
		effectstest.TmtMuCryptidOracle:    "Slithering Cryptid",
		effectstest.TmtMuTheOozeOracle:    "The Ooze",
		effectstest.TmtMuZooOracle:        "Zoo Escapees",
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
	effectstest.PassPriorityAroundTable(t, g)
}

// The four creatures whose only Mutagen is "when this enters, create a
// Mutagen token" make exactly one, and it works.
func TestTmtMutagenEntersTriggersMakeOneWorkingMutagen(t *testing.T) {
	for _, c := range []struct{ name, typeLine, oracle string }{
		{"Crustacean Commando", "Creature — Crab Mutant Soldier", effectstest.TmtMuCommandoOracle},
		{"Slithering Cryptid", "Creature — Fish Mutant", effectstest.TmtMuCryptidOracle},
		{"Ray Fillet, Man Ray", "Legendary Creature — Fish Mutant", effectstest.TmtMuRayFilletOracle},
		{"Michelangelo, Mutant BFF", "Legendary Creature — Mutant Ninja Turtle", effectstest.TmtMuBFFOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, c.name, c.typeLine, c.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		ids := effectstest.TmtB1Named(g, "Mutagen")
		if len(ids) != 1 {
			t.Errorf("%s: %d Mutagen tokens, want 1", c.name, len(ids))
			continue
		}
		tok := effectstest.TmtB1Card(t, g, ids[0])
		if tok.Controller != me.ID || !tok.IsToken() || !tok.HasSubtype("Mutagen") {
			t.Errorf("%s: the token is %q under %v", c.name, tok.TypeLine, tok.Controller)
		}
		tmtMuCrack(t, g, me, ids[0], bear)
		if got := effectstest.PlusOnes(g, bear); got != 1 {
			t.Errorf("%s: the Mutagen put %d counters, want 1", c.name, got)
		}
	}
}

// --- Zoo Escapees ------------------------------------------------------------

func TestTmtMutagenZooEscapeesMakesOneWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zoo := effectstest.B27Push(g, me.ID, "Zoo Escapees", "Creature — Boar Rhino", effectstest.TmtMuZooOracle, "{1}{G}", 2, 2, "G")
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 0 {
		t.Fatalf("%d Mutagens before it left", got)
	}
	effectstest.B27Kill(g, zoo)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens after it died, want 1", got)
	}

	other := effectstest.B27Push(g, me.ID, "Zoo Escapees", "Creature — Boar Rhino", effectstest.TmtMuZooOracle, "{1}{G}", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 2 {
		t.Errorf("%d Mutagens after the second one was bounced, want 2", got)
	}
}

// --- Genghis Frog ------------------------------------------------------------

func TestTmtMutagenGenghisFrogTriggersOnEnteringItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	frog := effectstest.CastCatalogSpell(t, g, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", effectstest.TmtMuGenghisOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertKeywords(t, g, frog, "trample")
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens after the Frog entered, want 1", got)
	}
}

func TestTmtMutagenGenghisFrogTriggersOnAnotherMutantYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", effectstest.TmtMuGenghisOracle, "{G}{U}", 1, 3, "G", "U")
	// A Mutant of mine: the Frog's trigger and the Commando's own.
	effectstest.CastCatalogSpell(t, g, "Crustacean Commando", "Creature — Crab Mutant Soldier", effectstest.TmtMuCommandoOracle, nil)
	effectstest.B27Settle(t, g, me.ID, uuid.Nil)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 2 {
		t.Errorf("%d Mutagens after a Mutant entered, want 2", got)
	}
}

func TestTmtMutagenGenghisFrogIgnoresANonMutant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", effectstest.TmtMuGenghisOracle, "{G}{U}", 1, 3, "G", "U")
	effectstest.CastCatalogSpell(t, g, "Zoo Escapees", "Creature — Boar Rhino", effectstest.TmtMuZooOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 0 {
		t.Errorf("%d Mutagens after a Boar Rhino entered, want 0", got)
	}
}

// An opponent's Frog does not see my Mutant.
func TestTmtMutagenGenghisFrogIgnoresAMutantItsControllerDoesNotControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, opp.ID, "Genghis Frog", "Legendary Creature — Frog Mutant Rogue", effectstest.TmtMuGenghisOracle, "{G}{U}", 1, 3, "G", "U")
	effectstest.CastCatalogSpell(t, g, "Crustacean Commando", "Creature — Crab Mutant Soldier", effectstest.TmtMuCommandoOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the Frog's controller got %d Mutagens from my Mutant, want 0", got)
	}
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("I have %d Mutagens (the Commando's own), want 1", got)
	}
}

// --- Michelangelo, Mutant BFF ----------------------------------------------

func TestTmtMutagenMichelangeloBFFMakesOneOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mikey := effectstest.B27Push(g, me.ID, "Michelangelo, Mutant BFF", "Legendary Creature — Mutant Ninja Turtle", effectstest.TmtMuBFFOracle, "{2}{G}{G}", 4, 4, "G")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mikey, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens after the attack, want 1", got)
	}
}

func TestTmtMutagenMichelangeloBFFLimitsBlockersOnCreaturesWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Michelangelo, Mutant BFF", "Legendary Creature — Mutant Ninja Turtle", effectstest.TmtMuBFFOracle, "{2}{G}{G}", 4, 4, "G")
	countered := effectstest.B12Creature(g, me.ID, "Countered Bear", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(countered, game.CounterStun, 1) })
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{countered, plain} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	// The one with a counter (a stun counter: any kind) can be blocked by
	// one creature only.
	effectstest.BrRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
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
	g, me, opp := effectstest.SpendTable(t)
	effectstest.B27Push(g, me.ID, "Michelangelo, Weirdness to 11", "Legendary Creature — Mutant Ninja Turtle", effectstest.TmtMuWeirdnessOracle, "{1}{G}", 1, 1, "G")
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	tmtMuCrack(t, g, me, effectstest.PushToken(g, me.ID, effects.MutagenToken()), mine)
	if got := effectstest.PlusOnes(g, mine); got != 2 {
		t.Errorf("my creature got %d counters from one Mutagen, want 2", got)
	}
	tmtMuCrack(t, g, me, effectstest.PushToken(g, me.ID, effects.MutagenToken()), theirs)
	if got := effectstest.PlusOnes(g, theirs); got != 1 {
		t.Errorf("an opponent's creature got %d counters, want 1", got)
	}
}

func TestTmtMutagenMichelangeloWeirdnessMakesAMutagenOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Michelangelo, Weirdness to 11", "Legendary Creature — Mutant Ninja Turtle", effectstest.TmtMuWeirdnessOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}

// --- Mutagen Man, Living Ooze ------------------------------------------------

func TestTmtMutagenMutagenManMakesXMutagens(t *testing.T) {
	for _, x := range []int{0, 3} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		man := effectstest.CastXSpell(t, g, "Mutagen Man, Living Ooze", "Legendary Creature — Ooze Mutant", effectstest.TmtMuMutagenManOracle, "{X}{G}{G}", x, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(man) {
			t.Fatalf("X=%d: Mutagen Man is not on the battlefield", x)
		}
		if got := effectstest.TmtMuMutagens(g, me.ID); got != x {
			t.Errorf("X=%d: %d Mutagens, want %d", x, got, x)
		}
	}
}

func TestTmtMutagenMutagenManDiscountsArtifactTokensYouControl(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	man := effectstest.B27Push(g, me.ID, "Mutagen Man, Living Ooze", "Legendary Creature — Ooze Mutant", effectstest.TmtMuMutagenManOracle, "{X}{G}{G}", 2, 3, "G")
	effectstest.AssertKeywords(t, g, man, "trample")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mutagen := effectstest.PushToken(g, me.ID, effects.MutagenToken())
	// No mana at all: the {1} is gone.
	if err := g.ActivateCatalogAbility(me.ID, mutagen, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("a Mutagen under Mutagen Man needed mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	// A Clue still needs {1} of its {2}.
	clue := effectstest.PushToken(g, me.ID, effects.ClueToken())
	if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a Clue was cracked for nothing")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, clue, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Errorf("a Clue under Mutagen Man would not crack for {1}: %v", err)
	}
}

func TestTmtMutagenMutagenManDoesNotDiscountAnOpponentsToken(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	effectstest.B27Push(g, me.ID, "Mutagen Man, Living Ooze", "Legendary Creature — Ooze Mutant", effectstest.TmtMuMutagenManOracle, "{X}{G}{G}", 2, 3, "G")
	theirs := effectstest.PushToken(g, opp.ID, effects.MutagenToken())
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Mutant Chain Reaction", "Sorcery", effectstest.TmtMuChainOracle, effectstest.B16TargetCard(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}

func TestTmtMutagenMutantChainReactionWithNoTargetStillMakesAMutagen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Mutant Chain Reaction", "Sorcery", effectstest.TmtMuChainOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}

func TestTmtMutagenMutantChainReactionTargetsOnlyFliersAmongCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ground := effectstest.B12Creature(g, opp.ID, "Ground Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Mutant Chain Reaction", "Sorcery", effectstest.TmtMuChainOracle, effectstest.B16TargetCard(ground)); err == nil {
		t.Error("a creature without flying was accepted as a target")
	}
	flier := effectstest.BrCreature(g, opp.ID, "Bird", "Creature — Bird", 1, 1, "flying")
	effectstest.CastCatalogSpell(t, g, "Mutant Chain Reaction", "Sorcery", effectstest.TmtMuChainOracle, effectstest.B16TargetCard(flier))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(flier) {
		t.Error("the flier was not destroyed")
	}
}

// --- Ooze Spill --------------------------------------------------------------

func TestTmtMutagenOozeSpillCountersAndMakesAMutagen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spell := effectstest.B13OpponentCasts(t, g, opp, "Their Opt", "Instant", "", "{U}", nil)
	effectstest.CastCatalogSpell(t, g, "Ooze Spill", "Instant", effectstest.TmtMuOozeSpillOracle, effectstest.B16TargetCard(spell))
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(spell) {
		t.Error("the spell was not countered")
	}
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
	if got := effectstest.TmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the opponent got %d Mutagens", got)
	}
}

// --- Ray Fillet, Man Ray -----------------------------------------------------

func TestTmtMutagenRayFilletRemovesACounterFromACreatureYouControlToDraw(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	ray := effectstest.B27Push(g, me.ID, "Ray Fillet, Man Ray", "Legendary Creature — Fish Mutant", effectstest.TmtMuRayFilletOracle, "{3}{U}", 3, 3, "U")
	effectstest.AssertKeywords(t, g, ray, "flying")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
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
	if got := effectstest.PlusOnes(g, bear); got != 0 {
		t.Errorf("the bear still has %d counters", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if err := activate(bear); err == nil {
		t.Error("activated with no +1/+1 counter to remove")
	}
}

// --- Return to the Sewers ----------------------------------------------------

func TestTmtMutagenReturnToTheSewersTucksThenMakesTheCastersMutagen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Return to the Sewers", "Instant", effectstest.TmtMuSewersOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PutInLibraryChoiceFor(g, opp.ID)
	if c == nil || c.LibraryPlacement != game.LibraryPlaceTopOrBottom {
		t.Fatalf("the owner was not asked top or bottom: %+v", c)
	}
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 0 {
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
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("the caster has %d Mutagens, want 1", got)
	}
	if got := effectstest.TmtMuMutagens(g, opp.ID); got != 0 {
		t.Errorf("the owner has %d Mutagens, want 0", got)
	}
}

// --- The Ooze ----------------------------------------------------------------

func TestTmtMutagenTheOozeMakesOnePerCounterOnACreatureThatLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "The Ooze", "Legendary Artifact", effectstest.TmtMuTheOozeOracle, "{2}", 0, 0)
	grown := effectstest.B12Creature(g, me.ID, "Grown Bear", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare Bear", "Creature — Bear", 2, 2)
	stunned := effectstest.B12Creature(g, me.ID, "Stunned Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(grown, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(stunned, game.CounterStun, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 3)
	})
	for _, id := range []uuid.UUID{bare, stunned, theirs} {
		effectstest.B27Kill(g, id)
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 0 {
		t.Fatalf("%d Mutagens from creatures without a +1/+1 counter of mine, want 0", got)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(grown) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 2 {
		t.Errorf("%d Mutagens after a creature with two counters left, want 2", got)
	}
}

func TestTmtMutagenTheOozeExilesACardFromAGraveyardAndMakesAMutagen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ooze := effectstest.B27Push(g, me.ID, "The Ooze", "Legendary Artifact", effectstest.TmtMuTheOozeOracle, "{2}", 0, 0)
	dead := game.NewCard("Their Dead Bear", opp.ID)
	dead.TypeLine = "Creature — Bear"
	opp.Graveyard.PushTop(dead)
	if err := g.ActivateCatalogAbility(me.ID, ooze, 0, game.ActivateAbilityParams{
		Strict:  true,
		Targets: effectstest.B16TargetCard(dead.InstanceID),
	}); err != nil {
		t.Fatalf("activate The Ooze: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(dead.InstanceID) {
		t.Error("the card was not exiled")
	}
	if got := effectstest.TmtMuMutagens(g, me.ID); got != 1 {
		t.Errorf("%d Mutagens, want 1", got)
	}
}
