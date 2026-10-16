package effects

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entwine_test.go — CR 702.42 (#3866, ADR 0065's 2026-10-16
// amendment): an entwine spell chooses its printed count, or — with the
// entwine cost announced and paid — every mode, never anything between.
// Every refusal is paired with an acceptance on the same board, so a
// refusal can only be the mode count or the price talking.

const spectralShiftOracle = "99eb4f63-841f-4b6b-9794-82bc2ed7b758"

// castEntwineSpell puts `name` in the caster's hand and casts it with
// exactly `pool` in the mana pool, announcing `modes` and — when
// `entwined` — the card's entwine cost (its only optional cost). A
// refused cast takes the card back out of the hand and empties the pool,
// so the next attempt starts from the same board.
func castEntwineSpell(g *game.Game, me *game.Player, name, typeLine, manaCost, oracle string, colors []string,
	pool string, modes []int, entwined bool, targets []game.TargetRef) (uuid.UUID, error) {
	id := handCardFull(me, name, typeLine, manaCost, oracle, colors)
	me.ManaPool.EmptyPool()
	if pool != "" {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, pool); err != nil {
			return id, err
		}
	}
	var optional []int
	if entwined {
		optional = []int{0}
	}
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, Modes: modes, OptionalCosts: optional, Targets: targets,
	})
	if err != nil {
		_, _ = me.Hand.Remove(id)
		me.ManaPool.EmptyPool()
	}
	return id, err
}

type shiftBoard struct {
	me, opp      *game.Player
	forest, bear uuid.UUID
}

func newShiftBoard(t *testing.T) (*game.Game, shiftBoard) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := cmcSeats(t, g)
	b := shiftBoard{me: me, opp: opp}
	b.forest = twPut(g, opp.ID, "Forest", "Basic Land — Forest", nil)
	b.bear = twPut(g, opp.ID, "Green Bear", "Creature — Bear", []string{"G"})
	return g, b
}

// targets is the announcement's target list: the land-type bullet on
// the Forest, the colour bullet on the bear, each stamped with the
// occurrence that chose it.
func (b shiftBoard) targets(modes []int) []game.TargetRef {
	out := make([]game.TargetRef, 0, len(modes))
	for occ, m := range modes {
		id := b.forest
		if m == 1 {
			id = b.bear
		}
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occ})
	}
	return out
}

func (b shiftBoard) cast(g *game.Game, pool string, modes []int, entwined bool) error {
	_, err := castEntwineSpell(g, b.me, "Spectral Shift", "Instant", "{1}{U}", spectralShiftOracle,
		[]string{"U"}, pool, modes, entwined, b.targets(modes))
	return err
}

// Not entwined: the printed "choose one". Both bullets are refused,
// and so is no bullet; either one alone is legal.
func TestSpectralShiftUnentwinedChoosesExactlyOne(t *testing.T) {
	g, b := newShiftBoard(t)
	if err := b.cast(g, "{C}{C}{U}{U}", []int{0, 1}, false); err == nil {
		t.Fatal("not entwined: both bullets must be refused")
	}
	if err := b.cast(g, "{C}{U}", nil, false); err == nil {
		t.Fatal("not entwined: no bullet must be refused")
	}
	if err := b.cast(g, "{C}{U}", []int{1}, false); err != nil {
		t.Fatalf("not entwined: one bullet is legal: %v", err)
	}
}

// Entwined: every bullet, no fewer. One bullet with the entwine cost
// announced is refused — CR 702.42a's "choose all modes" is the whole
// of what the cost buys.
func TestSpectralShiftEntwinedChoosesBoth(t *testing.T) {
	g, b := newShiftBoard(t)
	if err := b.cast(g, "{C}{C}{C}{U}", []int{0}, true); err == nil {
		t.Fatal("entwined: one bullet must be refused")
	}
	if err := b.cast(g, "{C}{C}{C}{U}", []int{0, 1}, true); err != nil {
		t.Fatalf("entwined: both bullets are legal: %v", err)
	}
}

// The entwine cost is paid on top of the mana cost (CR 601.2f): the
// printed {1}{U} alone does not cast it entwined.
func TestSpectralShiftEntwineCostsTwoMore(t *testing.T) {
	g, b := newShiftBoard(t)
	if err := b.cast(g, "{C}{U}", []int{0, 1}, true); err == nil {
		t.Fatal("entwined for {1}{U}: the {2} entwine cost is unpaid")
	}
	if err := b.cast(g, "{C}{C}{C}{U}", []int{0, 1}, true); err != nil {
		t.Fatalf("entwined for {3}{U}: %v", err)
	}
	if left := len(b.me.ManaPool.Units()); left != 0 {
		t.Errorf("%d mana left in the pool, want 0 — {1}{U} plus {2}", left)
	}
}

// Entwined, both changes are made, in the order written (CR 702.42b):
// the land type's two words are asked first, then the colour's, and
// the Forest is an Island and the bear's green is blue.
func TestSpectralShiftEntwinedMakesBothChangesInPrintedOrder(t *testing.T) {
	g, b := newShiftBoard(t)
	// Announced colour first: the order the bullets RUN in is still
	// the printed one.
	if err := b.cast(g, "{C}{C}{C}{U}", []int{1, 0}, true); err != nil {
		t.Fatalf("entwined: %v", err)
	}
	passPriorityAroundTable(t, g)
	first := latestOptionPickFor(g, b.me.ID)
	if first == nil {
		t.Fatal("the first bullet asks for the word to replace")
	}
	if !slices.ContainsFunc(first.PickOptions, func(o game.ChoiceOption) bool { return o.Label == "Forest" }) {
		t.Fatalf("the land-type bullet runs first (CR 702.42b); its options were %v", first.PickOptions)
	}
	answerOptionPick(t, g, b.me.ID, 4) // Forest
	answerOptionPick(t, g, b.me.ID, 1) // Plains, Island, Swamp, Mountain: Island
	answerOptionPick(t, g, b.me.ID, twGreen)
	answerOptionPick(t, g, b.me.ID, twBlue) // W, U, B, R: blue
	g.SettleResolution()
	if subs := effectiveSubtypes(t, g, b.forest); !slices.Equal(subs, []string{"Island"}) {
		t.Errorf("the Forest is a %v, want [Island]", subs)
	}
	if changes := twTextChanges(t, g, b.bear); len(changes) == 0 {
		t.Error("the bear's colour word was not changed")
	}
}

// The hand card's view carries both counts: one with nothing announced,
// two on if_optional_paid, which the picker switches to when the caster
// ticks "Entwine {2}" among the optional costs.
func TestSpectralShiftViewOffersEntwine(t *testing.T) {
	g, b := newShiftBoard(t)
	id := handCardFull(b.me, "Spectral Shift", "Instant", "{1}{U}", spectralShiftOracle, []string{"U"})
	b.me.Hand.Cards[len(b.me.Hand.Cards)-1].KnownBy = map[uuid.UUID]bool{b.me.ID: true}
	m := handModesFor(t, g, b.me, id)
	if m == nil || m.Min != 1 || m.Max != 1 {
		t.Fatalf("not entwined: %+v, want 1..1", m)
	}
	if m.IfOptionalPaid == nil || m.IfOptionalPaid.Min != 2 || m.IfOptionalPaid.Max != 2 {
		t.Errorf("entwined: %+v, want 2..2", m.IfOptionalPaid)
	}
	if offers := game.OptionalCostsFor(spectralShiftOracle); len(offers) != 1 || offers[0].Label != "Entwine {2}" || offers[0].Key != game.EntwineKey {
		t.Errorf("optional costs %+v, want one \"Entwine {2}\"", offers)
	}
}

// Register keeps the two halves together.
func TestRegisterRefusesHalfAnEntwine(t *testing.T) {
	bullets := func() []game.ModeOption {
		return []game.ModeOption{Mode("You gain 1 life."), Mode("You lose 1 life.")}
	}
	for _, tc := range []struct {
		name string
		spec Spec
		want string
	}{
		{"cost without the count", Spec{
			OptionalCosts: []game.AdditionalCost{Entwine("{2}")},
			Modes:         ChooseOne(bullets()...),
		}, "do not rise"},
		{"count without the cost", Spec{
			Modes: Entwining(ChooseOne(bullets()...)),
		}, "no entwine cost"},
		{"two entwine costs", Spec{
			OptionalCosts: []game.AdditionalCost{Entwine("{2}"), Entwine("{3}")},
			Modes:         Entwining(ChooseOne(bullets()...)),
		}, "entwine costs"},
	} {
		tc.spec.OracleID = "entwine-test-" + strings.ReplaceAll(tc.name, " ", "-")
		tc.spec.Name = "Entwine " + tc.name
		if msg := registerPanics(tc.spec); !strings.Contains(msg, tc.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", tc.name, msg, tc.want)
		}
	}
}
