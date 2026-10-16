package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soul_burn_x_test.go — #3941 (ADR 0068, amendment 2026-10-16): the
// payment records which mana paid the restricted {X}, and Soul Burn
// reads "the amount of {B} spent on X" off it instead of guessing the
// least it could have been.

// castSoulBurn floats `pool`, casts Soul Burn at X with `params`' payment
// choices, resolves it and returns the life gained and the record.
func castSoulBurn(t *testing.T, pool string, x int, params game.CastSpellParams) (gain int, paid game.PaidCost) {
	t.Helper()
	g, me, opp := spendTable(t)
	spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
	floatForTest(g, me, pool)
	params.XValue = x
	params.Targets = atPlayer(opp)
	before := me.Life
	if err := g.CastSpell(me.ID, spell, params); err != nil {
		t.Fatalf("cast X=%d off %q: %v", x, pool, err)
	}
	paid = g.StackMeta[spell].Paid
	passPriorityAroundTable(t, g)
	return me.Life - before, paid
}

// The record says which mana paid X, and the life gained is the black
// among it. The {B} is paid before X, so with enough black the X is
// all black even when red sits ahead of it in the pool.
func TestSoulBurnReadsTheBlackThatPaidX(t *testing.T) {
	for _, tc := range []struct {
		name   string
		pool   string
		x      int
		onXB   int
		onXR   int
		wantLP int
	}{
		// {B} takes a black, X takes the other two black, red pays {2}.
		{"black first, red in front", "RRBBB", 2, 2, 0, 2},
		// One black only, and it pays the printed {B}: none on X.
		{"the only black pays the printed B", "BRRRR", 2, 0, 2, 0},
		// Two black: one for {B}, one for X, the second X is red.
		{"as much black as there is", "BBRRR", 2, 1, 1, 1},
		// All black: every X mana is black.
		{"all black", "BBBBBR", 3, 3, 0, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gain, paid := castSoulBurn(t, tc.pool, tc.x, game.CastSpellParams{Strict: true})
			onX := paid.Spent().OnX()
			if got := onX.Total(); got != tc.x {
				t.Errorf("%d mana recorded against X, want %d: %+v", got, tc.x, paid.Mana)
			}
			if b, r := onX.Count("B"), onX.Count("R"); b != tc.onXB || r != tc.onXR {
				t.Errorf("on X: %d black and %d red, want %d and %d", b, r, tc.onXB, tc.onXR)
			}
			if got := paid.Spent().Total(); got != tc.x+3 {
				t.Errorf("%d mana spent in all, want %d", got, tc.x+3)
			}
			if gain != tc.wantLP {
				t.Errorf("gained %d, want %d", gain, tc.wantLP)
			}
		})
	}
}

// Only the {X} part is marked: the printed {2}{B} never is.
func TestSoulBurnMarksOnlyTheXPart(t *testing.T) {
	_, paid := castSoulBurn(t, "BBBBB", 2, game.CastSpellParams{Strict: true})
	marked := 0
	for _, tok := range paid.Mana {
		if tok.ForX {
			marked += tok.N()
		}
	}
	if marked != 2 || paid.Spent().Total() != 5 {
		t.Errorf("marked %d of %d spent, want 2 of 5: %+v", marked, paid.Spent().Total(), paid.Mana)
	}
}

// Auto-tap taps the lands that put the most black on X: three Swamps
// and three Mountains, X = 2, and all three Swamps are tapped, so two
// black are on X and the Mountains pay the {2}.
func TestSoulBurnAutoPayPutsTheBlackOnX(t *testing.T) {
	g, me, opp := spendTable(t)
	// The Mountains come first, so a plan that ignored the colour of X
	// would tap them for it.
	mountains := basics(g, me, "Mountain", 3)
	swamps := basics(g, me, "Swamp", 3)
	spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
	before := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: 2, Targets: atPlayer(opp)}); err != nil {
		t.Fatal(err)
	}
	if n := tappedCount(t, g, swamps); n != 3 {
		t.Errorf("%d Swamps tapped, want all three: black on X needs them", n)
	}
	if n := tappedCount(t, g, mountains); n != 2 {
		t.Errorf("%d Mountains tapped, want the two that pay {2}", n)
	}
	onX := g.StackMeta[spell].Paid.Spent().OnX()
	if onX.Count("B") != 2 {
		t.Errorf("%d black on X, want 2: %+v", onX.Count("B"), g.StackMeta[spell].Paid.Mana)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 2 {
		t.Errorf("gained %d, want 2", got)
	}
}

// The gain is still capped by the damage dealt, and by the target's
// life total before it: black on X cannot make it more than printed.
func TestSoulBurnGainStaysCappedByTheTarget(t *testing.T) {
	g, me, opp := spendTable(t)
	spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
	floatForTest(g, me, "BBBBBB")
	opp.Life = 2
	before := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 3, Targets: atPlayer(opp)}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 2 {
		t.Errorf("gained %d off three black on X into a 2-life target, want 2", got)
	}
}

// --- a waived payment ------------------------------------------------

// soulBurnShortBy casts Soul Burn at X = 2 with two black floating and
// returns the prompt, so the shortfall is one X slot and two generic.
func soulBurnShortBy(t *testing.T) (*game.Game, *game.Player, *game.Player, *game.PaperPaymentRequiredError, func(game.PaperPayment) (int, game.PaidCost)) {
	t.Helper()
	g, me, opp := spendTable(t)
	spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
	floatForTest(g, me, "BB")
	base := game.CastSpellParams{XValue: 2, Targets: atPlayer(opp)}
	ask := ppAsks(t, g.CastSpell(me.ID, spell, base))
	answer := func(d game.PaperPayment) (int, game.PaidCost) {
		t.Helper()
		params := base
		params.PaperPayment = &d
		before := me.Life
		if err := g.CastSpell(me.ID, spell, params); err != nil {
			t.Fatalf("declared %+v: %v", d, err)
		}
		paid := g.StackMeta[spell].Paid
		passPriorityAroundTable(t, g)
		return me.Life - before, paid
	}
	return g, me, opp, ask, answer
}

// The picker is told which slots are the X part.
func TestSoulBurnPaperPaymentNamesTheXSlots(t *testing.T) {
	_, _, _, ask, _ := soulBurnShortBy(t)
	if len(ask.Slots) != 3 {
		t.Fatalf("slots = %+v, want the X slot and two generic", ask.Slots)
	}
	if !ask.Slots[0].X || ask.Slots[0].Symbol != "{B/R}" {
		t.Errorf("first slot = %+v, want the {B/R} slot marked X", ask.Slots[0])
	}
	if ask.Slots[1].X || ask.Slots[2].X {
		t.Errorf("generic slots marked X: %+v", ask.Slots)
	}
}

// A declaration that marks its X mana keeps the caster's own allocation:
// black on X gains the life, red on X does not.
func TestSoulBurnPaperPaymentKeepsTheCastersAllocation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		onX     string
		wantGet int
	}{
		{"black on X", "B", 2},
		{"red on X", "R", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, _, answer := soulBurnShortBy(t)
			gain, paid := answer(game.PaperPayment{Mana: []game.PaperMana{
				{Color: tc.onX, X: true}, {Color: "B"}, {Color: "C"},
			}})
			if !paid.Spent().Known() {
				t.Error("a declared payment should be exact")
			}
			onX := paid.Spent().OnX()
			if onX.Total() != 2 {
				t.Errorf("%d recorded against X, want 2: %+v", onX.Total(), paid.Mana)
			}
			if gain != tc.wantGet {
				t.Errorf("gained %d, want %d", gain, tc.wantGet)
			}
		})
	}
}

// A declaration that names no X mana leaves the allocation to the
// engine, which gives X the black: the three declared mana are red,
// black and colourless, and the black is put on X.
func TestSoulBurnPaperPaymentUnmarkedGivesXTheBlack(t *testing.T) {
	_, _, _, _, answer := soulBurnShortBy(t)
	gain, paid := answer(game.PaperPayment{Mana: []game.PaperMana{{Color: "R"}, {Color: "B"}, {Color: "C"}}})
	if got := paid.Spent().OnX().Count("B"); got != 2 {
		t.Errorf("%d black on X, want 2 (the floating one and the declared one): %+v", got, paid.Mana)
	}
	if gain != 2 {
		t.Errorf("gained %d, want 2", gain)
	}
}

// A mark that is not a legal payment of the X slots is refused, and
// nothing is paid.
func TestSoulBurnPaperPaymentRefusesABadMark(t *testing.T) {
	for _, tc := range []struct {
		name string
		mana []game.PaperMana
	}{
		{"green cannot pay X", []game.PaperMana{{Color: "G", X: true}, {Color: "B"}, {Color: "C"}}},
		{"too many marked", []game.PaperMana{{Color: "B", X: true}, {Color: "B", X: true}, {Color: "C"}}},
		{"the generic marked, X not", []game.PaperMana{{Color: "B"}, {Color: "C", X: true}, {Color: "C"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := spendTable(t)
			spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
			floatForTest(g, me, "BB")
			params := game.CastSpellParams{XValue: 2, Targets: atPlayer(opp)}
			ppAsks(t, g.CastSpell(me.ID, spell, params))
			params.PaperPayment = &game.PaperPayment{Mana: tc.mana}
			err := g.CastSpell(me.ID, spell, params)
			if !errors.Is(err, game.ErrInvalidPaperPayment) {
				t.Fatalf("err = %v, want the declaration refused", err)
			}
			if got := len(me.ManaPool); got != 2 {
				t.Errorf("a refused declaration spent the pool: %d tokens left, want 2", got)
			}
		})
	}
}

// A cast declared free whose pool still paid part of the cost keeps that
// share (CR 601.2h): the black the pool put on X counts, and the mana the
// caster says they never paid does not.
func TestSoulBurnDeclaredFreeKeepsThePoolsShare(t *testing.T) {
	g, me, opp := spendTable(t)
	spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
	// Free, declared: the record is the pool's tokens alone and exact.
	floatForTest(g, me, "BB")
	params := game.CastSpellParams{XValue: 2, Targets: atPlayer(opp)}
	ppAsks(t, g.CastSpell(me.ID, spell, params))
	params.PaperPayment = &game.PaperPayment{Free: true}
	before := me.Life
	if err := g.CastSpell(me.ID, spell, params); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	// The pool paid {B} and one X: a black on X, the rest declared free.
	if got := me.Life - before; got != 1 {
		t.Errorf("gained %d, want 1: the one black the pool put on X", got)
	}
}
