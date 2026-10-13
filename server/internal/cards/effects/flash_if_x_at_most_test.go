package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flash_if_x_at_most_test.go — #3225: a per-offer X ceiling on an
// alternative cost, the shape behind "You may cast this spell as though
// it had flash if X is 3 or less" (Silver Scrutiny, CR 601.3c). The
// enumerator and the view halves live in internal/legal and
// internal/protocol; this file is the announce gate and the card.

const silverScrutinyFlashOracle = "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e"

func ssFlashTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, _ := p7Table(t)
	card := eacHandCard(g, me, "Silver Scrutiny", "Sorcery", "{X}{U}{U}")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == card {
				me.Hand.Cards[i].OracleID = silverScrutinyFlashOracle
			}
		}
	})
	return g, me, card
}

func ssFund(g *game.Game, me *game.Player, n int) {
	g.WithWriteLock(func() {
		me.ManaPool = nil
		for i := 0; i < n; i++ {
			c := "C"
			if i < 2 {
				c = "U"
			}
			me.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}

func ssToBeginCombat(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

// At instant speed the flash offer opens the window, and an X at the
// ceiling is accepted.
func TestSilverScrutinyFlashOfferAcceptsXUpToThree(t *testing.T) {
	g, me, card := ssFlashTable(t)
	ssToBeginCombat(t, g)
	// The printed cost stays sorcery-speed only.
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{XValue: 3}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("printed cost at instant speed: err = %v, want ErrSorcerySpeedRequired", err)
	}
	ssFund(g, me, 5)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 3, AlternativeCost: "flash-x-at-most"}); err != nil {
		t.Fatalf("flash cast with X=3: %v", err)
	}
	if got := g.StackMeta[card].XValue; got != 3 {
		t.Errorf("XValue = %d, want 3", got)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana left: %d, want 0 (the offer charges the printed cost)", len(me.ManaPool))
	}
}

// X = 4 under the flash offer is refused, and nothing moves.
func TestSilverScrutinyFlashOfferRefusesXFour(t *testing.T) {
	g, me, card := ssFlashTable(t)
	ssToBeginCombat(t, g)
	ssFund(g, me, 6)
	err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 4, AlternativeCost: "flash-x-at-most"})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("flash cast with X=4: err = %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(card) {
		t.Errorf("a refused cast reached the stack")
	}
	var inHand bool
	g.ReadSnapshot(func() { inHand = me.Hand.Contains(card) })
	if !inHand {
		t.Errorf("the refused cast left the hand")
	}
}

// The sorcery-speed cast for the printed cost is not bound by the
// offer: X = 10 is fine in a main phase.
func TestSilverScrutinyPrintedCastTakesAnyX(t *testing.T) {
	g, me, card := ssFlashTable(t)
	advanceToMain(t, g)
	ssFund(g, me, 12)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 10}); err != nil {
		t.Fatalf("printed cast with X=10: %v", err)
	}
	if got := g.StackMeta[card].XValue; got != 10 {
		t.Errorf("XValue = %d, want 10", got)
	}
}

// Claiming the offer in a main phase is still the offer: the ceiling
// goes with the claim, not with the speed it is cast at.
func TestSilverScrutinyFlashOfferIsBoundedAtSorcerySpeedToo(t *testing.T) {
	g, me, card := ssFlashTable(t)
	advanceToMain(t, g)
	ssFund(g, me, 12)
	err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, XValue: 10, AlternativeCost: "flash-x-at-most"})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("offer claimed with X=10: err = %v, want ErrInvalidParam", err)
	}
}

// The card is fully implemented now: the flash clause was its only
// caveat (#3225).
func TestSilverScrutinyIsFullAndDrawsX(t *testing.T) {
	spec, ok := Lookup(silverScrutinyFlashOracle)
	if !ok {
		t.Fatal("Silver Scrutiny is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness = %v, caveats = %v; want Full and none", spec.Completeness, spec.Caveats)
	}

	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Silver Scrutiny", "Sorcery", silverScrutinyFlashOracle, "{X}{U}{U}", 4, nil)
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 4 {
		t.Errorf("drew %d cards, want 4", got)
	}
}

// The constructor refuses a ceiling it cannot bound (no {X} in the
// price) and a negative one, at boot.
func TestFlashIfXAtMostRegisterChecks(t *testing.T) {
	cases := map[string]game.AlternativeCost{
		"no {X} to bound": {Key: "k", ManaCost: "{U}{U}", AsThoughFlash: true, XCeiling: 3},
		"negative":        {Key: "k", ManaCost: "{X}{U}", AsThoughFlash: true, XCeiling: -1},
	}
	for name, ac := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("Register accepted the offer")
				}
				if msg, _ := r.(string); !strings.Contains(msg, "XCeiling") {
					t.Errorf("panic = %v, want it to name XCeiling", r)
				}
			}()
			Register(Spec{
				OracleID:         "test-flash-x-register-" + strings.ReplaceAll(name, " ", "-"),
				Name:             "Test Flash X " + name,
				AlternativeCosts: []game.AlternativeCost{ac},
			})
		})
	}
}
