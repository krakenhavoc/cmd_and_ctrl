package effects

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ormos_archive_keeper_test.go — #4896: a discard cost "with different
// names", and Ormos, Archive Keeper's draw replacement.

const ormosOracle = "06d60b71-a9f9-4a7f-ad69-15d760b6e5c3"

type ormosTable struct {
	g     *game.Game
	me    *game.Player
	ormos uuid.UUID
}

func newOrmosTable(t *testing.T) ormosTable {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	ormos := b12Push(g, me.ID, "Ormos, Archive Keeper", "Legendary Creature — Sphinx", ormosOracle, 5, 5)
	advanceToMain(t, g)
	me.Hand.Cards = nil
	return ormosTable{g: g, me: me, ormos: ormos}
}

func ormosHand(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	return id
}

func ormosActivate(t *testing.T, tb ormosTable, picks ...uuid.UUID) error {
	t.Helper()
	floatMana(t, tb.g, tb.me, "{C}{U}{U}")
	return tb.g.ActivateCatalogAbility(tb.me.ID, tb.ormos, 0, game.ActivateAbilityParams{DiscardIDs: picks})
}

func TestOrmosArchiveKeeperIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(ormosOracle)
	if !ok {
		t.Fatal("Ormos, Archive Keeper is not registered")
	}
	if spec.Name != "Ormos, Archive Keeper" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%q: completeness %v caveats %v, want Full and none", spec.Name, spec.Completeness, spec.Caveats)
	}
	if !hasAbility(spec.PrintedKeywords, "flying") {
		t.Error("Ormos prints flying")
	}
	dc := spec.Activated[0].Cost.DiscardCards
	if dc == nil || dc.N != 3 || !dc.DifferentNames {
		t.Fatalf("discard clause = %+v, want three cards with different names", dc)
	}
}

func TestOrmosDiscardsThreeDifferentNamesAndDrawsFive(t *testing.T) {
	tb := newOrmosTable(t)
	a := ormosHand(tb.me, "Opt")
	b := ormosHand(tb.me, "Shock")
	c := ormosHand(tb.me, "Ponder")
	if tb.me.Library.Size() < 5 {
		t.Fatalf("library %d, the test needs five cards", tb.me.Library.Size())
	}
	if err := ormosActivate(t, tb, a, b, c); err != nil {
		t.Fatalf("three names were refused: %v", err)
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if !inGraveyard(tb.me, id) {
			t.Errorf("%v was not discarded as the cost", id)
		}
	}
	passPriorityAroundTable(t, tb.g)
	if got := handCount(tb.g, tb.me); got != 5 {
		t.Errorf("hand %d, want 5 (three discarded, five drawn)", got)
	}
}

// ormosRefused pins a refusal and that nothing was paid.
func ormosRefused(t *testing.T, tb ormosTable, picks ...uuid.UUID) {
	t.Helper()
	handBefore := handCount(tb.g, tb.me)
	err := ormosActivate(t, tb, picks...)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("activation = %v, want ErrInvalidParam", err)
	}
	for _, id := range picks {
		if inGraveyard(tb.me, id) || !tb.me.Hand.Contains(id) {
			t.Errorf("the refused discard moved %v", id)
		}
	}
	if got := handCount(tb.g, tb.me); got != handBefore {
		t.Errorf("hand %d, want %d unchanged", got, handBefore)
	}
	if n := tb.g.Stack.Size(); n != 0 {
		t.Errorf("stack has %d items after a refusal", n)
	}
}

func TestOrmosRefusesTwoCardsSharingAName(t *testing.T) {
	tb := newOrmosTable(t)
	a := ormosHand(tb.me, "Opt")
	b := ormosHand(tb.me, "Opt")
	c := ormosHand(tb.me, "Shock")
	ormosHand(tb.me, "Ponder")
	ormosRefused(t, tb, a, b, c)
}

// Both halves of a split card name it (CR 709.4, CR 201.2a).
func TestOrmosRefusesASplitCardBesideOneOfItsHalves(t *testing.T) {
	tb := newOrmosTable(t)
	split := game.Card{
		InstanceID: uuid.New(), Owner: tb.me.ID, Controller: tb.me.ID, Layout: game.LayoutSplit,
		Faces: []game.Face{
			{Name: "Fire", TypeLine: "Instant", ManaCost: "{1}{R}"},
			{Name: "Ice", TypeLine: "Instant", ManaCost: "{1}{U}"},
		},
	}
	split.SettleImported()
	tb.me.Hand.PushTop(split)
	ice := ormosHand(tb.me, "Ice")
	opt := ormosHand(tb.me, "Opt")
	ormosRefused(t, tb, split.InstanceID, ice, opt)
}

// ormosMoves lists the enumerator's moves for the draw ability.
func ormosMoves(t *testing.T, tb ormosTable) [][]string {
	t.Helper()
	var out [][]string
	for _, m := range legal.EnumerateFor(tb.g, tb.me.ID) {
		if m.Type != "activate_ability" || m.Source != tb.ormos {
			continue
		}
		var p struct {
			AbilityIndex int      `json:"ability_index"`
			DiscardIDs   []string `json:"discard_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if p.AbilityIndex == 0 {
			out = append(out, p.DiscardIDs)
		}
	}
	return out
}

func TestOrmosIsOnlyEnumeratedWithThreeNamesInHand(t *testing.T) {
	tb := newOrmosTable(t)
	floatMana(t, tb.g, tb.me, "{C}{U}{U}")
	ormosHand(tb.me, "Opt")
	ormosHand(tb.me, "Opt")
	ormosHand(tb.me, "Shock")
	ormosHand(tb.me, "Shock")

	if moves := ormosMoves(t, tb); len(moves) != 0 {
		t.Fatalf("four cards with two names were offered as the cost: %v", moves)
	}

	ponder := ormosHand(tb.me, "Ponder")
	moves := ormosMoves(t, tb)
	if len(moves) != 1 {
		t.Fatalf("%d draw moves with three names in hand, want 1", len(moves))
	}
	got := map[string]bool{}
	for _, id := range moves[0] {
		got[id] = true
	}
	if len(moves[0]) != 3 || !got[ponder.String()] {
		t.Errorf("discard_ids = %v, want three cards with the Ponder among them", moves[0])
	}
	names := map[string]bool{}
	for _, c := range tb.me.Hand.Cards {
		if got[c.InstanceID.String()] {
			names[c.Name] = true
		}
	}
	if len(names) != 3 {
		t.Errorf("the enumerated payment names %v, want three different names", names)
	}

	// The move the enumerator offers is one the engine accepts.
	ids := make([]uuid.UUID, 0, 3)
	for _, s := range moves[0] {
		ids = append(ids, uuid.MustParse(s))
	}
	if err := tb.g.ActivateCatalogAbility(tb.me.ID, tb.ormos, 0, game.ActivateAbilityParams{DiscardIDs: ids}); err != nil {
		t.Fatalf("the enumerated payment was refused: %v", err)
	}
}

func TestOrmosDifferentNamesRuleIsRefusedOnTheWrongForms(t *testing.T) {
	for name, dc := range map[string]*game.DiscardCost{
		"one card":     {N: 1, DifferentNames: true},
		"random":       {N: 3, Random: true, DifferentNames: true},
		"hand":         {Hand: true, DifferentNames: true},
		"x cards":      {CountFromX: true, DifferentNames: true},
		"mana value x": {N: 1, ManaValueX: true, DifferentNames: true},
		"legend name":  {N: 2, SameNameAsLegendaryPermanent: true, DifferentNames: true},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: Register accepted the different-names rule beside it", name)
				}
			}()
			checkDiscardClause("Fixture", "ability 0", dc)
		}()
	}
	checkDiscardClause("Fixture", "ability 0", DiscardNDifferentNames(3, "three cards with different names").DiscardCards)
}

func TestOrmosDifferentNamesRuleIsRefusedOnAManaAbility(t *testing.T) {
	if msg := registerPanics(Spec{
		OracleID: "00000000-0000-0000-0000-000000004896",
		Name:     "Different Names Mana Fixture",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{DiscardCards: &game.DiscardCost{N: 2, Label: "two cards with different names", DifferentNames: true}},
			Produced: "{C}",
		}},
	}); !strings.Contains(msg, "different names") {
		t.Errorf("a mana ability discarding cards with different names registered (panic %q)", msg)
	}
}

// --- the draw replacement ---

func TestOrmosPutsFiveCountersInsteadOfDrawingFromAnEmptyLibrary(t *testing.T) {
	tb := newOrmosTable(t)
	emptyLibrary(tb.me)
	tb.g.WithWriteLock(func() { _ = tb.g.DrawNForEffect(tb.me.ID, 1) })
	if got := plusOnes(tb.g, tb.ormos); got != 5 {
		t.Errorf("+1/+1 counters = %d, want 5", got)
	}
	if tb.me.AttemptedEmptyDraw {
		t.Error("the replaced draw set the empty-draw flag")
	}
	checkState(t, tb.g)
	if tb.me.Eliminated {
		t.Error("lost for a draw Ormos replaced")
	}
}

// A draw with cards left is an ordinary draw.
func TestOrmosLeavesADrawFromALibraryAlone(t *testing.T) {
	tb := newOrmosTable(t)
	if tb.me.Library.Size() == 0 {
		t.Fatal("the library is empty")
	}
	tb.g.WithWriteLock(func() { _ = tb.g.DrawNForEffect(tb.me.ID, 1) })
	if got := handCount(tb.g, tb.me); got != 1 {
		t.Errorf("hand %d, want the one card drawn", got)
	}
	if got := plusOnes(tb.g, tb.ormos); got != 0 {
		t.Errorf("+1/+1 counters = %d, want none", got)
	}
}

// "Draw five cards" off a two-card library: the first two are drawn and
// each of the other three is replaced on its own (CR 121.2): fifteen
// counters.
func TestOrmosDrawFiveFromATwoCardLibraryReplacesEachMissingDraw(t *testing.T) {
	tb := newOrmosTable(t)
	for tb.me.Library.Size() > 2 {
		_, _ = tb.me.Library.PopTop()
	}
	a := ormosHand(tb.me, "Opt")
	b := ormosHand(tb.me, "Shock")
	c := ormosHand(tb.me, "Ponder")
	if err := ormosActivate(t, tb, a, b, c); err != nil {
		t.Fatalf("activation: %v", err)
	}
	passPriorityAroundTable(t, tb.g)
	if got := handCount(tb.g, tb.me); got != 2 {
		t.Errorf("hand %d, want the two cards the library held", got)
	}
	if got := plusOnes(tb.g, tb.ormos); got != 15 {
		t.Errorf("+1/+1 counters = %d, want 15 (five for each of three replaced draws)", got)
	}
	if tb.me.AttemptedEmptyDraw {
		t.Error("a replaced draw set the empty-draw flag")
	}
}

// The counters can pause on a CR 616 ordering prompt (Doubling Season
// beside Hardened Scales); the rest of the draw waits behind it.
func TestOrmosCountersPauseOnAnOrderingPromptAndTheDrawWaits(t *testing.T) {
	tb := newOrmosTable(t)
	hs := seedReplacementPermanent(tb.g, hardenedScalesOracle, "Hardened Scales", tb.me.ID)
	ds := seedReplacementPermanent(tb.g, doublingSeasonOracle, "Doubling Season", tb.me.ID)
	emptyLibrary(tb.me)
	tb.g.WithWriteLock(func() { _ = tb.g.DrawNForEffect(tb.me.ID, 2) })
	if got := plusOnes(tb.g, tb.ormos); got != 0 {
		t.Fatalf("+1/+1 counters = %d with the placement owed to the prompt", got)
	}
	// Doubling Season first: 5 → 10 → 11.
	answerOrderBySource(t, tb.g, ds, hs)
	if got := plusOnes(tb.g, tb.ormos); got != 11 {
		t.Fatalf("+1/+1 counters = %d after the first draw, want 11", got)
	}
	// The second draw opens its own window and asks again.
	answerOrderBySource(t, tb.g, hs, ds)
	if got := plusOnes(tb.g, tb.ormos); got != 11+12 {
		t.Errorf("+1/+1 counters = %d after the second draw, want 23 ((5+1)*2 with Hardened Scales first)", got)
	}
	if len(tb.g.PendingChoices) != 0 {
		t.Errorf("%d prompts left after both draws", len(tb.g.PendingChoices))
	}
	if tb.me.AttemptedEmptyDraw {
		t.Error("a replaced draw set the empty-draw flag")
	}
}
