package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_mp_cards_test.go — Murders at Karlov Manor cards M–P (#4469). Each
// test plays the real catalog entry and asserts what its text says.

const (
	mkmMPGlassOracle       = "f2ce9e55-07c0-4f36-a59e-127693fc0f62"
	mkmMPBindingOracle     = "942f3f06-1df0-40b4-8537-71cbec111852"
	mkmMPPhantomOracle     = "6b0d4192-bc76-4ef1-92b2-b5fca81970bd"
	mkmMPGuardianOracle    = "ef414347-bbe3-47cb-b900-88275561c43e"
	mkmMPWatchOracle       = "2d400c01-d2d0-442a-bc3d-cf106a754dab"
	mkmMPInspectorOracle   = "c206bdcb-d2c4-4bda-a979-6c0627ef8f73"
	mkmMPOnTheJobOracle    = "e4733889-4d4c-4bf7-8a9b-fb5ad7a38395"
	mkmMPNightwatchOracle  = "3891a914-e597-4126-846f-39bf8534eb92"
	mkmMPMoroiiOracle      = "07d830a7-f0a2-4d46-9a53-101dca681603"
	mkmMPOffenderOracle    = "c8c42851-1b1b-49dd-a97c-7eda57dadb13"
	mkmMPGardenerOracle    = "15dca6a7-8367-4b86-a86c-e44f10fb3c53"
	mkmMPPerformerOracle   = "78f9ec0f-6015-4024-aca6-9a14cba373fa"
	mkmMPPerformerTypeLine = "Creature — Lizard Assassin"
)

func TestMkmMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmMPGlassOracle:      "Magnifying Glass",
		mkmMPBindingOracle:    "Makeshift Binding",
		mkmMPPhantomOracle:    "Marketwatch Phantom",
		mkmMPGuardianOracle:   "Neighborhood Guardian",
		mkmMPWatchOracle:      "Not on My Watch",
		mkmMPInspectorOracle:  "Novice Inspector",
		mkmMPOnTheJobOracle:   "On the Job",
		mkmMPNightwatchOracle: "Museum Nightwatch",
		mkmMPMoroiiOracle:     "Nightdrinker Moroii",
		mkmMPOffenderOracle:   "Offender at Large",
		mkmMPGardenerOracle:   "Nervous Gardener",
		mkmMPPerformerOracle:  "Pyrotechnic Performer",
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

// TestMkmMPDisguiseDeclarationsAreWired pins each disguise card's
// keyword: the {3} face-down cast, the disguise kind (ward {2}) and the
// card's own face-up cost.
func TestMkmMPDisguiseDeclarationsAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Museum Nightwatch", mkmMPNightwatchOracle, "{1}{W}"},
		{"Nightdrinker Moroii", mkmMPMoroiiOracle, "{B}{B}"},
		{"Offender at Large", mkmMPOffenderOracle, "{4}{R}"},
		{"Nervous Gardener", mkmMPGardenerOracle, "{G}"},
		{"Pyrotechnic Performer", mkmMPPerformerOracle, "{R}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "disguise")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no disguise cast", tc.name)
			continue
		}
		if ac.ManaCost != MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s: cast %q, kind %q, face-up %q; want %q, disguised, %q",
				tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCost, MorphCastCost, tc.faceUp)
		}
	}
}

// --- shared helpers ------------------------------------------------------------

// mkmMPCastFaceDown casts a catalog card from the active player's hand
// face down for its disguise and resolves it.
func mkmMPCastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: "disguise"}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

// mkmMPTurnFaceUp pays the disguise cost from conjured mana and turns
// the permanent face up.
func mkmMPTurnFaceUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...string) {
	t.Helper()
	g.WithWriteLock(func() { p.ManaPool.AddMana(manaTokens(mana...)...) })
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	if findCardAnywhere(t, g, id).FaceDown {
		t.Fatal("still face down after the special action")
	}
}

// mkmMPPT is a battlefield permanent's power and toughness after the
// layers and its counters.
func mkmMPPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := tmtB1Card(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// mkmMPOpponents is every seat but the active one.
func mkmMPOpponents(g *game.Game) []*game.Player {
	var out []*game.Player
	for i, p := range g.Seats {
		if i != g.Turn.ActiveSeat {
			out = append(out, p)
		}
	}
	return out
}

// mkmMPLives reads each player's life total.
func mkmMPLives(ps []*game.Player) []int {
	out := make([]int, len(ps))
	for i, p := range ps {
		out[i] = p.Life
	}
	return out
}

// --- Magnifying Glass ----------------------------------------------------------

func TestMkmMPMagnifyingGlassTapsForColorlessAndInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	glass := b12Push(g, me.ID, "Magnifying Glass", "Artifact", mkmMPGlassOracle, 0, 0)

	if err := g.ActivateManaAbility(me.ID, glass, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 1 {
		t.Errorf("%d colorless in the pool, want 1", got)
	}
	if !tmtB1Card(t, g, glass).Tapped {
		t.Error("the Glass is untapped after its mana ability")
	}
}

func TestMkmMPMagnifyingGlassInvestigatesForFourAndATap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	glass := b12Push(g, me.ID, "Magnifying Glass", "Artifact", mkmMPGlassOracle, 0, 0)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C", "C", "C", "C")...) })

	if err := g.ActivateCatalogAbility(me.ID, glass, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {4}, {T}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
	if !tmtB1Card(t, g, glass).Tapped {
		t.Error("the Glass is untapped after its {T} ability")
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, want the {4} spent", got)
	}
}

// --- Makeshift Binding ---------------------------------------------------------

func TestMkmMPMakeshiftBindingExilesUntilItLeavesAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life

	binding := castHoldCreature(t, g, "Makeshift Binding", "Enchantment", mkmMPBindingOracle, 0, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) || onBattlefieldNamed(g, "Their Bear") != 0 {
		t.Fatal("the creature was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(binding) })
	g.RunStateChecksForTest()
	if tmtB1Count(g, opp.ID, "Their Bear") != 1 {
		t.Error("the creature did not return when the Binding left")
	}
}

func TestMkmMPMakeshiftBindingOffersOnlyAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Makeshift Binding", "Enchantment", mkmMPBindingOracle, 0, 0)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("my own creature was accepted as the target")
		}
	}
}

// --- Marketwatch Phantom -------------------------------------------------------

func TestMkmMPMarketwatchPhantomFliesWhenASmallCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	phantom := b12Push(g, me.ID, "Marketwatch Phantom", "Creature — Spirit Detective", mkmMPPhantomOracle, 2, 2)
	if hasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Fatal("the Phantom flies before anything entered")
	}

	castHoldCreature(t, g, "Neighborhood Guardian", "Creature — Unicorn", mkmMPGuardianOracle, 2, 2)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Error("a 2/2 entered and the Phantom has no flying")
	}
}

func TestMkmMPMarketwatchPhantomIgnoresAThreePowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	phantom := b12Push(g, me.ID, "Marketwatch Phantom", "Creature — Spirit Detective", mkmMPPhantomOracle, 2, 2)
	castHoldCreature(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Error("a 3/2 entering gave the Phantom flying")
	}
}

func TestMkmMPMarketwatchPhantomCountsAFaceDownTwoTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	phantom := b12Push(g, me.ID, "Marketwatch Phantom", "Creature — Spirit Detective", mkmMPPhantomOracle, 2, 2)
	// The Performer is 3/2 face up and a 2/2 face down (CR 708.2a).
	mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Error("a face-down 2/2 entered and the Phantom has no flying")
	}
}

// --- Neighborhood Guardian -----------------------------------------------------

func TestMkmMPNeighborhoodGuardianPumpsTheChosenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guardian := b12Push(g, me.ID, "Neighborhood Guardian", "Creature — Unicorn", mkmMPGuardianOracle, 2, 2)

	castHoldCreature(t, g, "Marketwatch Phantom", "Creature — Spirit Detective", mkmMPPhantomOracle, 2, 2)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, guardian)
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, guardian); p != 3 || tough != 3 {
		t.Errorf("the Guardian is %d/%d, want 3/3", p, tough)
	}
}

func TestMkmMPNeighborhoodGuardianIgnoresAThreePowerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Neighborhood Guardian", "Creature — Unicorn", mkmMPGuardianOracle, 2, 2)
	castHoldCreature(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a 3/2 entering triggered the Guardian")
	}
}

// --- Not on My Watch -----------------------------------------------------------

func TestMkmMPNotOnMyWatchExilesAnAttackingCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := b12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, attacker)

	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Not on My Watch", TypeLine: "Instant",
		OracleID: mkmMPWatchOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}},
	}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	}); err != nil {
		t.Fatalf("cast Not on My Watch: %v", err)
	}
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(attacker) || !g.Exile.Contains(attacker) {
		t.Error("the attacking creature was not exiled")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the other creature left the battlefield")
	}
}

// --- Novice Inspector ----------------------------------------------------------

func TestMkmMPNoviceInspectorInvestigatesOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Novice Inspector", "Creature — Human Detective", mkmMPInspectorOracle, 1, 2)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- On the Job ----------------------------------------------------------------

func TestMkmMPOnTheJobPumpsYourCreaturesAndInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	castCatalogSpell(t, g, "On the Job", "Instant", mkmMPOnTheJobOracle, nil)
	passPriorityAroundTable(t, g)

	if p, tough := mkmMPPT(t, g, a); p != 4 || tough != 3 {
		t.Errorf("my 2/2 is %d/%d, want 4/3", p, tough)
	}
	if p, tough := mkmMPPT(t, g, b); p != 5 || tough != 4 {
		t.Errorf("my 3/3 is %d/%d, want 5/4", p, tough)
	}
	if p, tough := mkmMPPT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's 2/2 is %d/%d, want unchanged", p, tough)
	}
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Museum Nightwatch ---------------------------------------------------------

func TestMkmMPMuseumNightwatchMakesADetectiveWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watch := b12Push(g, me.ID, "Museum Nightwatch", "Creature — Centaur Soldier", mkmMPNightwatchOracle, 3, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(watch) })
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
}

func TestMkmMPMuseumNightwatchDyingFaceDownMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watch := mkmMPCastFaceDown(t, g, "Museum Nightwatch", "Creature — Centaur Soldier", mkmMPNightwatchOracle, 3, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(watch) })
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Detective"); got != 0 {
		t.Errorf("%d Detectives from a face-down Nightwatch, want 0 (CR 708.2a)", got)
	}
}

func TestMkmMPMuseumNightwatchTurnedFaceUpIsAThreeTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watch := mkmMPCastFaceDown(t, g, "Museum Nightwatch", "Creature — Centaur Soldier", mkmMPNightwatchOracle, 3, 2)
	mkmMPTurnFaceUp(t, g, me, watch, "W", "C")
	if p, tough := mkmMPPT(t, g, watch); p != 3 || tough != 2 {
		t.Errorf("face up the Nightwatch is %d/%d, want 3/2", p, tough)
	}
}

// --- Nightdrinker Moroii -------------------------------------------------------

func TestMkmMPNightdrinkerMoroiiFliesAndCostsThreeLifeOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := castHoldCreature(t, g, "Nightdrinker Moroii", "Creature — Vampire", mkmMPMoroiiOracle, 4, 2)
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, me.Life)
	}
	assertKeywords(t, g, id, "flying")
}

func TestMkmMPNightdrinkerMoroiiFaceDownNeverLosesTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := mkmMPCastFaceDown(t, g, "Nightdrinker Moroii", "Creature — Vampire", mkmMPMoroiiOracle, 4, 2)
	if hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the face-down Moroii kept flying (CR 708.2a)")
	}
	mkmMPTurnFaceUp(t, g, me, id, "B", "B")
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d; neither the face-down entry nor turning it face up loses life (CR 708.3, CR 708.8)", life, me.Life)
	}
	assertKeywords(t, g, id, "flying")
	if p, tough := mkmMPPT(t, g, id); p != 4 || tough != 2 {
		t.Errorf("face up the Moroii is %d/%d, want 4/2", p, tough)
	}
}

// --- Offender at Large ---------------------------------------------------------

func TestMkmMPOffenderAtLargePumpsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Offender at Large", "Creature — Giant Rogue", mkmMPOffenderOracle, 5, 4)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
}

func TestMkmMPOffenderAtLargePumpsWhenTurnedFaceUpNotWhenEnteringFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := mkmMPCastFaceDown(t, g, "Offender at Large", "Creature — Giant Rogue", mkmMPOffenderOracle, 5, 4)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Offender")
	}
	mkmMPTurnFaceUp(t, g, me, id, "R", "C", "C", "C", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
}

func TestMkmMPOffenderAtLargeMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Offender at Large", "Creature — Giant Rogue", mkmMPOffenderOracle, 5, 4)
	b04WaitForPick(t, g, me.ID)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("with no target chosen the bear is %d/%d, want 2/2", p, tough)
	}
}

// --- Nervous Gardener ----------------------------------------------------------

func TestMkmMPNervousGardenerFetchesALandWithABasicLandTypeWhenTurnedUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dual := stapleLibraryCard(me, "Breeding Pool", "Land — Forest Island")
	stapleLibraryCard(me, "Wastes", "Basic Land")
	id := mkmMPCastFaceDown(t, g, "Nervous Gardener", "Creature — Dryad", mkmMPGardenerOracle, 2, 2)
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("entering face down asked for a search")
	}

	mkmMPTurnFaceUp(t, g, me, id, "G")
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("turning the Gardener face up asked for no search")
	}
	if searchOptionNamed(g, c, "Wastes") != uuid.Nil {
		t.Error("Wastes, which has no basic land type, was offered")
	}
	answerSearchByID(t, g, me.ID, dual)
	if !me.Hand.Contains(dual) {
		t.Error("the land did not reach the hand")
	}
}

// --- Pyrotechnic Performer -----------------------------------------------------

func TestMkmMPPyrotechnicPerformerBurnsEachOpponentWhenItTurnsUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := mkmMPOpponents(g)
	id := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	before := mkmMPLives(opps)
	myLife := me.Life

	mkmMPTurnFaceUp(t, g, me, id, "R")
	passPriorityAroundTable(t, g)
	for i, p := range opps {
		if p.Life != before[i]-3 {
			t.Errorf("opponent %d: life %d -> %d, want -3", i, before[i], p.Life)
		}
	}
	if me.Life != myLife {
		t.Errorf("my life changed %d -> %d", myLife, me.Life)
	}
}

func TestMkmMPPyrotechnicPerformerUsesTheOtherCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := mkmMPOpponents(g)
	b12Push(g, me.ID, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	moroii := mkmMPCastFaceDown(t, g, "Nightdrinker Moroii", "Creature — Vampire", mkmMPMoroiiOracle, 4, 2)
	before := mkmMPLives(opps)

	mkmMPTurnFaceUp(t, g, me, moroii, "B", "B")
	passPriorityAroundTable(t, g)
	for i, p := range opps {
		if p.Life != before[i]-4 {
			t.Errorf("opponent %d: life %d -> %d, want -4 (the Moroii's power)", i, before[i], p.Life)
		}
	}
}
