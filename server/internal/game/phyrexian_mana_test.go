package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// phyrexian_mana_test.go — CR 107.4's ten hybrid Phyrexian symbols
// (#787), from the parser out to a cast that pays one with life.
//
// Before this, {G/W/P} was an `unknown token`: ParseCost refused it,
// so legal/cast.go offered no cast, the engine rejected one with
// ErrUnparseableCost, and Card.ManaValue() read 0 for a card CR 202.3g
// says is four.

// hybridPhyrexianSymbols is CR 107.4's list, in the order the rule
// prints it — the five allied pairs then the five enemy pairs.
var hybridPhyrexianSymbols = []struct {
	symbol string
	colors [2]string
}{
	{"{W/U/P}", [2]string{"W", "U"}},
	{"{U/B/P}", [2]string{"U", "B"}},
	{"{B/R/P}", [2]string{"B", "R"}},
	{"{R/G/P}", [2]string{"R", "G"}},
	{"{G/W/P}", [2]string{"G", "W"}},
	{"{W/B/P}", [2]string{"W", "B"}},
	{"{U/R/P}", [2]string{"U", "R"}},
	{"{B/G/P}", [2]string{"B", "G"}},
	{"{R/W/P}", [2]string{"R", "W"}},
	{"{G/U/P}", [2]string{"G", "U"}},
}

// TestHybridPhyrexianSymbolsParse — one symbol, two colour options,
// the Phyrexian flag, for all ten. The symbol model is the existing
// ColorRequirement with no new field (mana_cost.go).
func TestHybridPhyrexianSymbolsParse(t *testing.T) {
	for _, tc := range hybridPhyrexianSymbols {
		got, err := ParseCost(tc.symbol)
		if err != nil {
			t.Errorf("ParseCost(%s): %v", tc.symbol, err)
			continue
		}
		if len(got.Required) != 1 {
			t.Errorf("ParseCost(%s): %d requirements, want 1 — a hybrid Phyrexian symbol is ONE symbol", tc.symbol, len(got.Required))
			continue
		}
		req := got.Required[0]
		if len(req.Options) != 2 || req.Options[0] != tc.colors[0] || req.Options[1] != tc.colors[1] {
			t.Errorf("ParseCost(%s): options = %v, want %v", tc.symbol, req.Options, tc.colors)
		}
		if !req.Phyrexian {
			t.Errorf("ParseCost(%s): Phyrexian = false, want true (CR 107.4f)", tc.symbol)
		}
		if !got.HasPhyrexian {
			t.Errorf("ParseCost(%s): HasPhyrexian = false, want true", tc.symbol)
		}
		if got.Generic != 0 || got.XSlots != 0 || got.HasSnow {
			t.Errorf("ParseCost(%s): stray generic/X/snow: %+v", tc.symbol, got)
		}
		// Lowercase is the same symbol (Scryfall never writes it that
		// way; the parser has always been case-insensitive).
		lower, err := ParseCost(strings.ToLower(tc.symbol))
		if err != nil || len(lower.Required) != 1 || !lower.Required[0].Phyrexian {
			t.Errorf("ParseCost(%s) lowercased: %+v, %v", tc.symbol, lower, err)
		}
	}
}

// TestHybridPhyrexianManaValueIsOne — CR 202.3g: each Phyrexian symbol
// counts as 1 toward mana value, hybrid Phyrexian included. Tamiyo,
// Compleated Sage is four, not three and not zero.
func TestHybridPhyrexianManaValueIsOne(t *testing.T) {
	for _, tc := range hybridPhyrexianSymbols {
		cost, err := ParseCost(tc.symbol)
		if err != nil {
			t.Fatalf("ParseCost(%s): %v", tc.symbol, err)
		}
		if got := cost.ManaValue(); got != 1 {
			t.Errorf("%s mana value = %d, want 1 (CR 202.3g)", tc.symbol, got)
		}
	}
	for _, tc := range []struct {
		name, cost string
		want       int
	}{
		{"Tamiyo, Compleated Sage", "{2}{G}{G/U/P}{U}", 5},
		{"Ajani, Sleeper Agent", "{1}{G}{G/W/P}{W}", 4},
		{"Lukka, Bound to Ruin", "{2}{R}{R/G/P}{G}", 5},
		{"Nahiri, the Unforgiving", "{1}{R}{R/W/P}{W}", 4},
	} {
		c := NewCard(tc.name, uuid.New())
		c.ManaCost = tc.cost
		if got := c.ManaValue(); got != tc.want {
			t.Errorf("%s (%s) mana value = %d, want %d", tc.name, tc.cost, got, tc.want)
		}
		if _, ok := c.ParsedManaValue(); !ok {
			t.Errorf("%s (%s): ParsedManaValue says the cost is unreadable", tc.name, tc.cost)
		}
	}
}

// TestHybridPhyrexianIsBothColors — CR 202.2d / CR 903.4: the card is
// every colour in the symbol, so a hybrid Phyrexian planeswalker is
// two colours and its colour identity names both.
func TestHybridPhyrexianIsBothColors(t *testing.T) {
	for _, tc := range hybridPhyrexianSymbols {
		c := NewCard("Compleated Test", uuid.New())
		c.ManaCost = tc.symbol
		c.TypeLine = "Planeswalker"
		got := map[string]bool{}
		for _, col := range c.EffectiveColors() {
			got[col] = true
		}
		if len(got) != 2 || !got[tc.colors[0]] || !got[tc.colors[1]] {
			t.Errorf("%s: colours = %v, want exactly %v (CR 202.2d)", tc.symbol, c.EffectiveColors(), tc.colors)
		}
	}
}

// TestHybridPhyrexianPaysWithEitherColor — CR 107.4f, the mana half:
// one mana of EITHER colour satisfies the symbol, and a third colour
// does not.
func TestHybridPhyrexianPaysWithEitherColor(t *testing.T) {
	for _, tc := range hybridPhyrexianSymbols {
		cost, err := ParseCost(tc.symbol)
		if err != nil {
			t.Fatalf("ParseCost(%s): %v", tc.symbol, err)
		}
		for _, col := range tc.colors {
			if !(ManaPool{{Color: col}}).CanPay(cost, 0) {
				t.Errorf("%s: a {%s} in the pool cannot pay it", tc.symbol, col)
			}
		}
		for _, col := range []string{"W", "U", "B", "R", "G", "C"} {
			if col == tc.colors[0] || col == tc.colors[1] {
				continue
			}
			if (ManaPool{{Color: col}}).CanPay(cost, 0) {
				t.Errorf("%s: a {%s} in the pool paid it, and %s is not one of its colours", tc.symbol, col, col)
			}
		}
	}
}

// TestPhyrexianMissingSymbolNamesTheLifeOption — the breakdown a
// player sees when the mana half is short spells the whole symbol,
// so the life option is visible in the message.
func TestPhyrexianMissingSymbolNamesTheLifeOption(t *testing.T) {
	for _, tc := range []struct{ cost, want string }{
		{"{G/W/P}", "{G/W/P}"},
		{"{U/P}", "{U/P}"},
		{"{W/U}", "{W/U}"},
		{"{W}", "{W}"},
	} {
		parsed, err := ParseCost(tc.cost)
		if err != nil {
			t.Fatalf("ParseCost(%s): %v", tc.cost, err)
		}
		missing := ManaPool{}.Missing(parsed, 0)
		if len(missing) != 1 || missing[0] != tc.want {
			t.Errorf("Missing(%s) = %v, want [%s]", tc.cost, missing, tc.want)
		}
	}
}

// --- the life half (CR 107.4f, CR 601.2b) ------------------------

// TestPhyrexianSymbolsCount is the ceiling CastSpellParams.PhyrexianLife
// is validated against.
func TestPhyrexianSymbolsCount(t *testing.T) {
	for _, tc := range []struct {
		cost string
		want int
	}{
		{"{1}{G}{G/W/P}{W}", 1},
		{"{1}{R/P}{R/P}", 2},
		{"{G/W/P}{U/P}", 2},
		{"{2}{W}{U}", 0},
		{"{W/U}", 0},
		{"{2/W}", 0},
	} {
		parsed, err := ParseCost(tc.cost)
		if err != nil {
			t.Fatalf("ParseCost(%s): %v", tc.cost, err)
		}
		if got := parsed.PhyrexianSymbols(); got != tc.want {
			t.Errorf("PhyrexianSymbols(%s) = %d, want %d", tc.cost, got, tc.want)
		}
	}
}

// TestPhyrexianLifePlanStrikesTheUnpayableSymbolFirst — which symbol
// the engine charges to life when the caster names a count.
func TestPhyrexianLifePlanStrikesTheUnpayableSymbolFirst(t *testing.T) {
	// {G/P}{W/P} with only a {W} floating: one life payment has to
	// take the green pip, or the cast cannot be made at all.
	cost, err := ParseCost("{G/P}{W/P}")
	if err != nil {
		t.Fatal(err)
	}
	pool := ManaPool{{Color: "W"}}
	reduced, life := PhyrexianLifePlan(cost, pool, ManaSpendContext{}, 1)
	if life != 2 {
		t.Fatalf("life = %d, want 2 (CR 107.4f)", life)
	}
	if len(reduced.Required) != 1 || reduced.Required[0].Options[0] != "W" {
		t.Fatalf("remaining requirement = %+v, want the {W/P}", reduced.Required)
	}
	if !pool.CanPay(reduced, 0) {
		t.Errorf("the reduced cost is still unpayable from %v", pool)
	}
	// A caster who wants the free cast with mana available still gets
	// it: the second pass takes a payable symbol.
	free, freeLife := PhyrexianLifePlan(cost, ManaPool{{Color: "G"}, {Color: "W"}}, ManaSpendContext{}, 2)
	if freeLife != 4 || len(free.Required) != 0 {
		t.Errorf("two symbols by life = %d life, %d requirements left; want 4 and 0", freeLife, len(free.Required))
	}
	// Nothing claimed changes nothing.
	same, none := PhyrexianLifePlan(cost, pool, ManaSpendContext{}, 0)
	if none != 0 || len(same.Required) != 2 {
		t.Errorf("claiming nothing changed the cost: %+v, %d life", same, none)
	}
}

// compleatedSage seeds a hybrid Phyrexian card in `owner`'s hand —
// Tamiyo, Compleated Sage's printed cost, on a sorcery so the cast
// needs no planeswalker machinery.
func compleatedSage(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Compleated Sage", owner.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{2}{G}{G/U/P}{U}"
	c.Layout = "normal"
	c.OracleID = "test-compleated-sage"
	owner.Hand.PushTop(c)
	return c.InstanceID
}

// TestCastPaysAHybridPhyrexianSymbolWithMana — end to end, the mana
// half: five mana, no life spent, the spell reaches the stack.
func TestCastPaysAHybridPhyrexianSymbolWithMana(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := compleatedSage(t, g, me)
	life := me.Life
	me.ManaPool.AddMana(
		ManaToken{Color: "C"}, ManaToken{Color: "C"},
		ManaToken{Color: "G"}, ManaToken{Color: "G"}, ManaToken{Color: "U"},
	)

	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "hand"}); err != nil {
		t.Fatalf("cast paying the Phyrexian symbol with mana: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatalf("the spell is not on the stack")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the cast = %v, want empty", me.ManaPool)
	}
	if me.Life != life {
		t.Errorf("life = %d, want %d — the mana half must not charge life", me.Life, life)
	}
}

// TestCastPaysAHybridPhyrexianSymbolWithLife — end to end, the life
// half (CR 107.4f): four mana and 2 life, through PayLifeForEffect.
func TestCastPaysAHybridPhyrexianSymbolWithLife(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := compleatedSage(t, g, me)
	life := me.Life
	// One short of the mana cost: no second green, no second blue.
	me.ManaPool.AddMana(
		ManaToken{Color: "C"}, ManaToken{Color: "C"},
		ManaToken{Color: "G"}, ManaToken{Color: "U"},
	)

	// Without the claim the cast is short exactly the hybrid symbol.
	err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "hand"})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("cast with no life claim: got %v, want *InsufficientManaError", err)
	}
	// One symbol short. WHICH one the greedy walk names is its own
	// business — it assigns the pool's {U} to the hybrid and reports
	// the plain {U} — and TestPhyrexianMissingSymbolNamesTheLifeOption
	// pins the formatting on its own.
	if len(short.Missing) != 1 {
		t.Errorf("missing = %v, want exactly one symbol", short.Missing)
	}
	if me.Life != life {
		t.Fatalf("a refused cast cost %d life", life-me.Life)
	}

	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "hand", PhyrexianLife: 1}); err != nil {
		t.Fatalf("cast paying the Phyrexian symbol with 2 life: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatalf("the spell is not on the stack")
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d (CR 107.4f: 2 life)", me.Life, life-PhyrexianLifePerSymbol)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the cast = %v, want empty", me.ManaPool)
	}
}

// TestCastRefusesAnOverclaimedPhyrexianLife — CR 601.2b announces a
// payment the cost actually prints, and CR 119.4 caps it at the
// caster's life total. Both reject without charging anything.
func TestCastRefusesAnOverclaimedPhyrexianLife(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := compleatedSage(t, g, me)
	me.ManaPool.AddMana(
		ManaToken{Color: "C"}, ManaToken{Color: "C"},
		ManaToken{Color: "G"}, ManaToken{Color: "G"}, ManaToken{Color: "U"},
	)

	// The cost prints one Phyrexian symbol; two is a malformed
	// announce.
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "hand", PhyrexianLife: 2}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming two symbols on a one-symbol cost: got %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(id) {
		t.Fatalf("the refused cast reached the stack")
	}

	// CR 119.4: only down to 0. At 1 life, 2 life cannot be paid.
	life := me.Life
	me.Life = 1
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, FromZone: "hand", PhyrexianLife: 1}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming 2 life at 1 life: got %v, want ErrInvalidParam", err)
	}
	if me.Life != 1 {
		t.Errorf("life = %d, want 1 — a refused cast paid anyway", me.Life)
	}
	if len(me.ManaPool) != 5 {
		t.Errorf("pool = %v, want the five tokens back — a refused cast spent mana", me.ManaPool)
	}
	me.Life = life
}

// TestPhyrexianLifeIsNotAutoTapped — the auto-tapper plans the cost
// the caster is actually paying with mana, so a claimed symbol does
// not strand a land.
func TestPhyrexianLifeIsNotAutoTapped(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := compleatedSage(t, g, me)
	for _, land := range []struct{ name, typeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Island", "Basic Land — Island"},
		{"Plains", "Basic Land — Plains"},
		{"Swamp", "Basic Land — Swamp"},
	} {
		pushBattlefieldForTest(g, me.ID, land.name, land.typeLine, "")
	}
	life := me.Life

	if err := g.CastSpell(me.ID, id, CastSpellParams{
		Strict: true, AutoTap: true, FromZone: "hand", PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("auto-tapped cast with one symbol paid by life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-PhyrexianLifePerSymbol)
	}
	tapped := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Tapped {
			tapped++
		}
	}
	if tapped != 4 {
		t.Errorf("tapped %d lands, want 4 — the four mana the cast still owes", tapped)
	}
}

// TestUndoAcrossAPhyrexianLifePayment — the rewind takes the life
// back with the spell, and the replay charges it once more, not twice
// (#806's clone fix is what makes the second half meaningful).
func TestUndoAcrossAPhyrexianLifePayment(t *testing.T) {
	g := newActiveGame(t)
	// RestoreFrom swaps g.Seats wholesale, so the seat is re-read
	// after the rewind rather than captured once.
	seat := func() *Player { return g.Seats[0] }
	advanceTo(t, g, StepPrecombatMain)
	id := compleatedSage(t, g, seat())
	life := seat().Life
	fund := func() {
		seat().ManaPool.AddMana(
			ManaToken{Color: "C"}, ManaToken{Color: "C"},
			ManaToken{Color: "G"}, ManaToken{Color: "U"},
		)
	}
	fund()

	beforeCast := g.Clone()
	if err := g.CastSpell(seat().ID, id, CastSpellParams{Strict: true, FromZone: "hand", PhyrexianLife: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if seat().Life != life-PhyrexianLifePerSymbol {
		t.Fatalf("life after the cast = %d, want %d", seat().Life, life-PhyrexianLifePerSymbol)
	}

	g.WithWriteLock(func() { g.RestoreFrom(beforeCast) })
	if seat().Life != life {
		t.Errorf("life after the undo = %d, want %d — the payment did not rewind", seat().Life, life)
	}
	if g.Stack.Contains(id) {
		t.Fatalf("the spell survived the rewind on the stack")
	}
	if !seat().Hand.Contains(id) {
		t.Fatalf("the rewind did not put the card back in hand")
	}
	if len(seat().ManaPool) != 4 {
		t.Errorf("pool after the undo = %v, want the four tokens back", seat().ManaPool)
	}

	if err := g.CastSpell(seat().ID, id, CastSpellParams{Strict: true, FromZone: "hand", PhyrexianLife: 1}); err != nil {
		t.Fatalf("replayed cast: %v", err)
	}
	if seat().Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life after the replay = %d, want %d — the payment doubled", seat().Life, life-PhyrexianLifePerSymbol)
	}
}
