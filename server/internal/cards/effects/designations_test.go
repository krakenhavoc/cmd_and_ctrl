package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// designations_test.go is the Class and Case half of ADR 0071's
// catalog side — the three cards that prove the one gate on real
// printed text: a gated static's worth of level lines (Wizard Class),
// a gated cost modifier (Fortune Teller's Talent) and a gated trigger
// (Case of the Shattered Pact). The station half is station_test.go;
// the engine contract behind both — the predicate, the filter, the
// layer invalidation, the snapshot — is game/designations_test.go.

const (
	wizardClassOracle           = "36f68aa3-9955-46f1-bc87-497f16ef5222"
	fortuneTellersTalentOracle  = "1b430f67-3686-4452-8594-b060f1a5a04e"
	caseOfTheShatteredPactOracl = "d00a089b-7d0d-4b89-902b-7b914e892178"
)

// pushClass seeds a Class permanent that is not summoning sick and
// carries a CR 613 timestamp, so the layer pass sees it.
func pushClass(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracle,
		Owner:      owner,
		Controller: owner,
	})
}

// levelOf reads a battlefield permanent's CR 716.2 level through the
// engine's accessor, which is the same one the wire and the level-up
// condition read.
func levelOf(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() { n = g.ClassLevelFor(id) })
	return n
}

// levelUpTo activates the Class's "Level N" ability and resolves it,
// paying with mana conjured into the pool. Returns the activation
// error so the timing and CR 716.2e tests can assert on it.
func levelUpTo(t *testing.T, g *game.Game, controller, classID uuid.UUID, index int, mana string) error {
	t.Helper()
	g.WithWriteLock(func() { _ = g.AddManaForEffect(controller, uuid.Nil, mana) })
	if err := g.ActivateCatalogAbility(controller, classID, index, game.ActivateAbilityParams{}); err != nil {
		return err
	}
	passPriorityAroundTable(t, g)
	return nil
}

// --- Wizard Class (CR 716) -------------------------------------------

// TestWizardClassStartsAtLevelOneWithOnlyItsLevelOneAbility is the
// first half of the ADR: a Class enters level 1 (CR 716.2b) and its
// later lines are not abilities it has.
func TestWizardClassStartsAtLevelOneWithOnlyItsLevelOneAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushClass(g, me.ID, "Wizard Class", "Enchantment — Class", wizardClassOracle)

	if got := levelOf(g, id); got != 1 {
		t.Errorf("a Class that has never been levelled is level %d, want 1 (CR 716.2b)", got)
	}
	// Level 1 — "You have no maximum hand size" — is on from the
	// start, and it is the printed NoMaxHandSize slot rather than a
	// gated static, because a level-1 line has nothing to wait for.
	if !game.CatalogNoMaxHandSize(wizardClassOracle) {
		t.Error("Wizard Class level 1 should declare NoMaxHandSize")
	}

	card := layeredCard(t, g, id)
	live := game.TriggersForCard(card)
	if len(live) != 0 {
		t.Errorf("at level 1 the level-2 and level-3 triggers must not exist: got %d", len(live))
	}
	// Both level-up abilities ARE there — they are printed on the
	// card from the moment it enters (CR 716.2a) and are limited by
	// their activation condition, not by a gate.
	if got := len(game.ActivatedAbilitiesForCard(card)); got != 2 {
		t.Errorf("level-up abilities offered = %d, want 2", got)
	}
}

// TestWizardClassLevelTwoDrawsAndTurnsOnItsTrigger walks the whole
// CR 716.2 lifecycle on a real card: level up, the level event, the
// "becomes level 2" trigger, and the level-3 trigger still absent.
func TestWizardClassLevelTwoDrawsAndTurnsOnItsTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushClass(g, me.ID, "Wizard Class", "Enchantment — Class", wizardClassOracle)
	advanceTo(t, g, game.StepPrecombatMain)

	before := len(me.Hand.Cards)
	if err := levelUpTo(t, g, me.ID, id, 0, "{2}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if got := levelOf(g, id); got != 2 {
		t.Fatalf("level after the level-2 activation = %d, want 2", got)
	}
	// "When this Class becomes level 2, draw two cards."
	if got := len(me.Hand.Cards) - before; got != 2 {
		t.Errorf("cards drawn = %d, want 2", got)
	}

	card := layeredCard(t, g, id)
	live := game.TriggersForCard(card)
	if len(live) != 1 {
		t.Fatalf("at level 2 exactly the level-2 trigger exists: got %d", len(live))
	}
	// The level-3 "whenever you draw a card" line must NOT have been
	// live for the two cards the level-2 ability just drew — which is
	// the difference between a gate and an `if level >= 3` inside
	// AppliesTo.
	if pendingTriggerCount(g) != 0 {
		t.Errorf("the level-3 draw trigger fired at level 2: %d pending", pendingTriggerCount(g))
	}
}

// TestWizardClassLevelThreeActivatesEveryLine — at level 3 all three
// printed lines are live, and the draw trigger is one of them.
func TestWizardClassLevelThreeActivatesEveryLine(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushClass(g, me.ID, "Wizard Class", "Enchantment — Class", wizardClassOracle)
	creature := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	advanceTo(t, g, game.StepPrecombatMain)

	if err := levelUpTo(t, g, me.ID, id, 0, "{2}{U}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	if err := levelUpTo(t, g, me.ID, id, 1, "{4}{U}"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	if got := levelOf(g, id); got != 3 {
		t.Fatalf("level = %d, want 3", got)
	}
	if got := len(game.TriggersForCard(layeredCard(t, g, id))); got != 2 {
		t.Errorf("at level 3 both triggers exist: got %d", got)
	}

	// Draw a card and let the level-3 trigger put a counter on the
	// Bear. It targets, so the harvester opens a pick_target prompt
	// as the ability goes on the stack (CR 603.3d).
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	pickCard(t, g, me.ID, creature)
	passPriorityAroundTable(t, g)
	if got := counterCountOn(g, creature, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters on the target = %d, want 1", got)
	}
}

// TestWizardClassLevelUpIsSorceryTimedAndInOrder — CR 716.2d and
// 716.2e, the two rules LevelUp carries so no card file has to.
func TestWizardClassLevelUpIsSorceryTimedAndInOrder(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushClass(g, me.ID, "Wizard Class", "Enchantment — Class", wizardClassOracle)
	advanceTo(t, g, game.StepPrecombatMain)

	// CR 716.2e: level 3 cannot be activated from level 1.
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{4}{U}") })
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err == nil {
		t.Error("level 3 from level 1 must be refused (CR 716.2e)")
	}
	if got := levelOf(g, id); got != 1 {
		t.Errorf("a refused activation must not change the level: %d", got)
	}

	// CR 716.2d: sorcery speed. The upkeep is not a main phase.
	advanceTo(t, g, game.StepUpkeep)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{2}{U}") })
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a level-up at instant speed must be refused (CR 716.2d)")
	}
}

// TestClassLevelIsNotCopiedByACopyEffect — CR 716.2c through the
// catalog: the copy is level 1 and therefore has only the level-1
// line, whatever the original had climbed to.
func TestClassLevelIsNotCopiedByACopyEffect(t *testing.T) {
	original := game.Card{
		InstanceID: uuid.New(),
		Name:       "Wizard Class",
		TypeLine:   "Enchantment — Class",
		OracleID:   wizardClassOracle,
		ClassLevel: 2,
	}
	if got := len(game.TriggersForCard(original)); got != 1 {
		t.Fatalf("the original at level 2 has %d triggers, want 1", got)
	}
	// A copy takes the COPIABLE VALUES and nothing else (CR 707.2),
	// and the level is not among them — game.PrintedValues has no
	// slot for it, which is CR 716.2c enforced by the type rather
	// than by a rule in the copy path. What a copy therefore looks
	// like is a permanent carrying those values at level 1:
	v := game.CopiableValuesOf(original)
	copied := game.Card{
		InstanceID: uuid.New(),
		OracleID:   v.OracleID,
		Name:       v.Name,
		TypeLine:   v.TypeLine,
	}
	if copied.OracleID != original.OracleID {
		t.Fatal("the copy should carry the original's catalog identity")
	}
	if copied.ClassLevel != 0 {
		t.Errorf("the copy took the level designation: %d (CR 716.2c says it must not)", copied.ClassLevel)
	}
	if got := len(game.TriggersForCard(copied)); got != 0 {
		t.Errorf("a copy of a level-2 Class has %d triggers, want 0 — it is level 1", got)
	}
}

// --- Fortune Teller's Talent (#333) ----------------------------------

// TestFortuneTellersTalentReductionWaitsForLevelThree is the gated
// COST MODIFIER, the fourth slot the gate covers. The reduction must
// not apply at level 1 or 2 and must apply at level 3 — and only to
// a spell cast from outside its controller's hand.
func TestFortuneTellersTalentReductionWaitsForLevelThree(t *testing.T) {
	seat := uuid.New()
	talent := game.Card{
		InstanceID: uuid.New(),
		Name:       "Fortune Teller's Talent",
		TypeLine:   "Enchantment — Class",
		OracleID:   fortuneTellersTalentOracle,
		Owner:      seat,
		Controller: seat,
	}
	if got := len(game.CostModifiersForCard(talent)); got != 0 {
		t.Errorf("at level 1 the reduction must not exist: got %d modifiers", got)
	}
	talent.ClassLevel = 2
	if got := len(game.CostModifiersForCard(talent)); got != 0 {
		t.Errorf("at level 2 the reduction must not exist: got %d modifiers", got)
	}
	talent.ClassLevel = 3
	mods := game.CostModifiersForCard(talent)
	if len(mods) != 1 {
		t.Fatalf("at level 3 the reduction must exist: got %d modifiers", len(mods))
	}
	m := mods[0]
	q := game.CostQuery{
		Controller: seat,
		Source:     talent,
		FromZone:   game.ZoneGraveyard,
	}
	if m.AppliesTo == nil || !m.AppliesTo(q) {
		t.Error("a spell cast from the graveyard should be discounted")
	}
	q.FromZone = game.ZoneHand
	if m.AppliesTo(q) {
		t.Error("a spell cast from HAND must not be discounted — the card says 'anywhere other than your hand'")
	}
	q.FromZone = game.ZoneGraveyard
	q.Controller = uuid.New()
	if m.AppliesTo(q) {
		t.Error("an opponent's spell must not be discounted — the card says 'spells you cast'")
	}
	if m.Amount == nil || m.Amount(q) != 2 {
		t.Error("the reduction is {2}")
	}
}

// TestFortuneTellersTalentDeclaresItsLibraryTopCaveats — the two
// clauses that are NOT implemented are declared, not approximated
// (ADR 0037 §5). If #765 lands and the clauses are built, this test
// is the one that should be deleted along with the caveats.
func TestFortuneTellersTalentDeclaresItsLibraryTopCaveats(t *testing.T) {
	spec, ok := Lookup(fortuneTellersTalentOracle)
	if !ok {
		t.Fatal("Fortune Teller's Talent is not registered")
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 2 {
		t.Fatalf("want CompletenessCaveats with 2 caveats, got %s with %d", spec.Completeness, len(spec.Caveats))
	}
	for _, cv := range spec.Caveats {
		if !containsFoldASCII(cv, "top card of your library") && !containsFoldASCII(cv, "top of your library") {
			t.Errorf("caveat does not name the library-top clause it covers: %q", cv)
		}
	}
}

// --- Case of the Shattered Pact (CR 719) -----------------------------

// TestCaseSolvesAtTheEndStepWhenItsConditionHolds is the CR 719.3a
// lifecycle: the trigger is the end step, the condition is the
// intervening if, and the solved line turns on.
func TestCaseSolvesAtTheEndStepWhenItsConditionHolds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushClass(g, me.ID, "Case of the Shattered Pact", "Enchantment — Case", caseOfTheShatteredPactOracl)

	// Four colours: not solvable yet.
	for _, color := range []string{"W", "U", "B", "R"} {
		pushColoredPermanent(g, me.ID, "Pip "+color, color)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if isSolved(g, id) {
		t.Fatal("a Case solved with only four colours among its controller's permanents")
	}
	if got := len(game.TriggersForCard(layeredCard(t, g, id))); got != 2 {
		t.Errorf("unsolved, the ETB and to-solve triggers exist and the solved one does not: got %d", got)
	}

	// Fifth colour, and round to this seat's NEXT end step — a full
	// four-seat turn cycle away, because "To solve" is a trigger on
	// the CONTROLLER's end step (CR 719.3a).
	pushColoredPermanent(g, me.ID, "Pip G", "G")
	for i := 0; i < 200 && !isSolved(g, id); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !isSolved(g, id) {
		t.Fatal("the Case never solved with five colours among its controller's permanents")
	}
	if got := len(game.TriggersForCard(layeredCard(t, g, id))); got != 3 {
		t.Errorf("solved, all three triggers exist: got %d", got)
	}
}

// TestSolvedCaseStaysSolvedWhenItsConditionStops — CR 719.3b. The
// condition is a one-time question, not an "as long as".
func TestSolvedCaseStaysSolvedWhenItsConditionStops(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushClass(g, me.ID, "Case of the Shattered Pact", "Enchantment — Case", caseOfTheShatteredPactOracl)
	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
	})
	if !isSolved(g, id) {
		t.Fatal("the Case did not solve")
	}
	// No permanents of any colour at all; the condition is false.
	if !isSolved(g, id) {
		t.Error("a solved Case must stay solved (CR 719.3b)")
	}
	if got := len(game.TriggersForCard(layeredCard(t, g, id))); got != 3 {
		t.Errorf("a solved Case keeps its solved ability: got %d triggers, want 3", got)
	}
}

// --- shared helpers --------------------------------------------------
//
// layeredCard (ability_removal_test.go) is the read that forces a
// recompute first, which is what makes Effective()-backed answers
// (IsCreature, HasKeyword, CurrentPower) reflect the current board.

func isSolved(g *game.Game, id uuid.UUID) bool {
	solved := false
	g.ReadSnapshot(func() { solved = g.IsSolved(id) })
	return solved
}

func pendingTriggerCount(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() { n = len(g.PendingTriggers) })
	return n
}

// pushColoredPermanent seeds a one-colour permanent, for Case of the
// Shattered Pact's "five colors among permanents you control".
func pushColoredPermanent(g *game.Game, owner uuid.UUID, name, color string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Enchantment",
		Colors:     []string{color},
		Owner:      owner,
		Controller: owner,
	})
}

// --- catalog-wide guards ---------------------------------------------

// TestNoRegisteredSpecDeclaresADoorGate — ADR 0071 decision 3. The
// Room door gate is reserved and not built: game.Card has no unlocked
// state, so Designation.Active answers false for it, and a card that
// declared one would ship with that ability switched off forever.
// Register refuses one at boot; this is the whole-catalog sweep that
// says so without needing a card to try.
func TestNoRegisteredSpecDeclaresADoorGate(t *testing.T) {
	for _, spec := range All() {
		for _, d := range specDesignations(spec) {
			if d.Kind == game.DesignationDoorUnlocked {
				t.Errorf("%s gates an ability on an unlocked Room door, which is not built yet (#886)", spec.Name)
			}
		}
	}
}
