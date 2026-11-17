package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dice_5681_test.go — #5681: the static "roll that many dice plus one and
// ignore the lowest roll" (Pixie Guide, Barbarian Class), a roll's
// modifier beside its natural result (CR 706.2), and the cards that read
// one or the other.

const (
	pixieGuideOracle          = "b6003419-1a2a-46f3-b32c-47b825324f78"
	barbarianClassOracle      = "59faa1b9-17aa-4f2c-a8f8-17ab50392b36"
	faridehOracle             = "faf5c18c-2642-4df3-a21f-3b237527e6a6"
	netheresePuzzleWardOracle = "561e8ecc-1f2b-4fd4-8f92-a0acea449816"
	criticalHitOracle         = "992e6845-1c6c-4b74-839c-317823bde15b"
)

// d5681Key finds an RNG key under which seat 0's next n d<sides> from
// `source` are rolls `want` accepts, by peeking on a throwaway table.
func d5681Key(t *testing.T, source uuid.UUID, sides, n int, want func(rolls []int) bool) [32]byte {
	t.Helper()
	for k := 0; k < 256; k++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(byte(k)))
		var rolls []int
		g.WithWriteLock(func() {
			rolls, _ = g.RollDiceForEffect(game.RandomDraw{Player: g.Seats[0].ID, Source: source}, sides, n)
		})
		if len(rolls) == n && want(rolls) {
			return randomModesKey(byte(k))
		}
	}
	t.Fatal("no key in 0..255 gives the wanted rolls")
	return [32]byte{}
}

// d5681Roll has player roll n d<sides> plus modifier from source through
// the dice door, then lets the triggers it caused resolve.
func d5681Roll(t *testing.T, g *game.Game, player, source uuid.UUID, sides, n, modifier int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		err = g.RollDiceThenForEffect(game.DiceRoll{Player: player, Source: source, Sides: sides, N: n, Modifier: modifier})
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
}

// d5681Rolled lists the EventRollDie events after event index from.
func d5681Rolled(g *game.Game, from int) []game.Event {
	var out []game.Event
	g.ReadSnapshot(func() {
		for _, ev := range g.Events[from:] {
			if ev.Kind == game.EventRollDie {
				out = append(out, ev)
			}
		}
	})
	return out
}

func d5681InZone(g *game.Game, id uuid.UUID, kind game.ZoneKind) bool {
	in := false
	g.WithWriteLock(func() {
		z := g.FindCardZoneForEffect(id)
		in = z != nil && z.Kind == kind
	})
	return in
}

// --- Pixie Guide ------------------------------------------------------

// One die becomes two and only the higher is rolled as far as the game is
// concerned; nobody is asked. An opponent's roll is not replaced.
func TestPixieGuideRollsOneMoreAndIgnoresTheLowest(t *testing.T) {
	source := uuid.New()
	key := d5681Key(t, source, 20, 2, func(r []int) bool { return r[0] < r[1] })
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(key)
	me, opp := g.Seats[0], g.Seats[1]
	control := g.Clone()
	var raw []int
	control.WithWriteLock(func() {
		raw, _ = control.RollDiceForEffect(game.RandomDraw{Player: me.ID, Source: source}, 20, 2)
	})
	p2Push(g, me.ID, "Pixie Guide", "Creature — Faerie", pixieGuideOracle, 1, 3)

	from := sfkEventCount(g)
	d5681Roll(t, g, me.ID, source, 20, 1, 0)
	if sfkPrompt(g, "dice/ignore-roll") != nil {
		t.Fatal("ignoring the lowest roll asked a question")
	}
	got := d5681Rolled(g, from)
	if len(got) != 1 || got[0].Amount != raw[1] {
		t.Fatalf("announced %+v, want only the higher of %v", got, raw)
	}

	from = sfkEventCount(g)
	d5681Roll(t, g, opp.ID, uuid.New(), 20, 3, 0)
	if got := d5681Rolled(g, from); len(got) != 3 {
		t.Errorf("an opponent's three dice announced %d rolls, want 3", len(got))
	}
}

// Two Pixie Guides: two more dice and the two lowest ignored (CR 614.5:
// each applies once), on a restored table too.
func TestPixieGuideTwoOfThemStackAndSurviveARestore(t *testing.T) {
	source := uuid.New()
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(randomModesKey(7))
	me := g.Seats[0]
	p2Push(g, me.ID, "Pixie Guide", "Creature — Faerie", pixieGuideOracle, 1, 3)
	p2Push(g, me.ID, "Pixie Guide", "Creature — Faerie", pixieGuideOracle, 1, 3)
	g = restoreRoundTrip(t, g, true)
	control := g.Clone()
	var raw []int
	control.WithWriteLock(func() {
		raw, _ = control.RollDiceIgnoringLowestForEffect(game.RandomDraw{Player: me.ID, Source: source}, 6, 4, 2)
	})
	from := sfkEventCount(g)
	d5681Roll(t, g, me.ID, source, 6, 2, 0)
	got := d5681Rolled(g, from)
	if len(got) != 2 || got[0].Amount != raw[0] || got[1].Amount != raw[1] {
		t.Fatalf("announced %+v, want the two highest of four, %v", got, raw)
	}
}

// --- Barbarian Class ----------------------------------------------------

func TestBarbarianClassLevels(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	class := p2Push(g, me.ID, "Barbarian Class", "Enchantment — Class", barbarianClassOracle, 0, 0)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	// Level 1: one more die, the lowest ignored, and no trigger yet.
	from := sfkEventCount(g)
	d5681Roll(t, g, me.ID, uuid.New(), 20, 1, 0)
	if got := d5681Rolled(g, from); len(got) != 1 {
		t.Fatalf("level 1 announced %d rolls, want 1 of the 2 rolled", len(got))
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("the level 2 ability triggered at level 1")
	}
	if p2Keyword(g, bear, "haste") {
		t.Fatal("level 3's haste at level 1")
	}

	// Level 2, through the printed level-up.
	if err := p2Activate(t, g, me, class, 0, "RR"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	passPriorityAroundTable(t, g)
	var level int
	g.ReadSnapshot(func() { level = g.ClassLevelFor(class) })
	if level != 2 {
		t.Fatalf("level = %d, want 2", level)
	}
	var err error
	g.WithWriteLock(func() {
		err = g.RollDiceThenForEffect(game.DiceRoll{Player: me.ID, Source: uuid.New(), Sides: 6, N: 2})
	})
	if err != nil {
		t.Fatalf("roll: %v", err)
	}
	g.RunStateChecksForTest()
	answerTriggerTargets(t, g, me.ID, bear)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("two dice in one instruction triggered the level 2 ability twice")
	}
	passPriorityAroundTable(t, g)
	if p, tough := p2PT(g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
	if !p2Keyword(g, bear, "menace") {
		t.Error("the target did not gain menace")
	}

	// Level 3: creatures you control have haste.
	if err := p2Activate(t, g, me, class, 1, "RRR"); err != nil {
		t.Fatalf("level 3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !p2Keyword(g, bear, "haste") {
		t.Error("no haste at level 3")
	}
	theirs := pushVanillaCreature(g, g.Seats[1].ID, "Their Bear", 2, 2)
	if p2Keyword(g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

// --- Farideh, Devil's Chosen --------------------------------------------

// faridehRoll rolls one d20 showing `natural` plus modifier for seat 0
// with Farideh out, and reports whether a card was drawn and whether she
// has flying and menace.
func faridehRoll(t *testing.T, natural func(int) bool, modifier int) (drew, keywords bool) {
	t.Helper()
	source := uuid.New()
	key := d5681Key(t, source, 20, 1, func(r []int) bool { return natural(r[0]) })
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(key)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	farideh := p2Push(g, me.ID, "Farideh, Devil's Chosen", "Legendary Creature — Tiefling Warlock", faridehOracle, 3, 3)
	before := handSize(me)
	d5681Roll(t, g, me.ID, source, 20, 1, modifier)
	return handSize(me) == before+1, p2Keyword(g, farideh, "flying") && p2Keyword(g, farideh, "menace")
}

func TestFaridehReadsTheResultNotTheNaturalRoll(t *testing.T) {
	for _, tc := range []struct {
		name     string
		natural  func(int) bool
		modifier int
		draws    bool
	}{
		{"a 10 or higher draws", func(v int) bool { return v >= 10 }, 0, true},
		{"a 9 or lower does not", func(v int) bool { return v <= 9 }, 0, false},
		{"a 5 plus 5 is a result of 10", func(v int) bool { return v == 5 }, 5, true},
		{"a 12 minus 3 is a result of 9", func(v int) bool { return v == 12 }, -3, false},
	} {
		drew, keywords := faridehRoll(t, tc.natural, tc.modifier)
		if drew != tc.draws {
			t.Errorf("%s: drew = %v", tc.name, drew)
		}
		if !keywords {
			t.Errorf("%s: Farideh did not gain flying and menace", tc.name)
		}
	}
}

// One trigger per instruction, one card at most, and an opponent's roll
// is not hers.
func TestFaridehOncePerInstruction(t *testing.T) {
	source := uuid.New()
	key := d5681Key(t, source, 20, 3, func(r []int) bool { return r[0] >= 10 && r[1] >= 10 })
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(key)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	farideh := p2Push(g, me.ID, "Farideh, Devil's Chosen", "Legendary Creature — Tiefling Warlock", faridehOracle, 3, 3)
	before := handSize(me)
	d5681Roll(t, g, g.Seats[1].ID, uuid.New(), 20, 2, 30)
	if handSize(me) != before || p2Keyword(g, farideh, "flying") {
		t.Fatal("an opponent's roll triggered Farideh")
	}
	d5681Roll(t, g, me.ID, source, 20, 3, 0)
	if got := handSize(me); got != before+1 {
		t.Errorf("hand = %d, want %d: one card for the instruction", got, before+1)
	}
}

// --- Netherese Puzzle-Ward ------------------------------------------------

// Focus Beam scries the d4's result; a 4 is also the die's highest
// natural result and draws.
func TestNetheresePuzzleWardFocusBeam(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	p2Push(g, me.ID, "Netherese Puzzle-Ward", "Enchantment", netheresePuzzleWardOracle, 0, 0)
	advanceToUpkeepOfSeat(t, g, 0)
	from := sfkEventCount(g)
	before := handSize(me)
	passPriorityAroundTable(t, g)
	rolled := d5681Rolled(g, from)
	if len(rolled) != 1 || rolled[0].Sides != 4 {
		t.Fatalf("rolled %+v, want one d4", rolled)
	}
	c := pendingOfKind(g, game.PendingChoiceScry)
	if c == nil || c.Chooser != me.ID || len(c.ScryCards) != rolled[0].Amount {
		t.Fatalf("scry prompt %+v, want %d cards for the roller", c, rolled[0].Amount)
	}
	answerScryKeepAll(t, g, me.ID)
	passPriorityAroundTable(t, g)
	want := before
	if rolled[0].Amount == 4 {
		want++
	}
	if got := handSize(me); got != want {
		t.Errorf("hand = %d, want %d after a d4 showing %d", got, want, rolled[0].Amount)
	}
}

// Perfect Illumination reads the die, not the result: a 20 minus 3 draws,
// a 17 plus 3 does not, each highest face of one instruction draws, and
// an opponent's roll is not yours.
func TestNetheresePuzzleWardPerfectIllumination(t *testing.T) {
	for _, tc := range []struct {
		name            string
		sides, n        int
		rolls           func(r []int) bool
		modifier, draws int
	}{
		{"a natural 20 with a modifier", 20, 1, func(r []int) bool { return r[0] == 20 }, -3, 1},
		{"a 17 plus 3", 20, 1, func(r []int) bool { return r[0] == 17 }, 3, 0},
		{"a 6 on a d6", 6, 1, func(r []int) bool { return r[0] == 6 }, 0, 1},
		{"two sixes", 6, 2, func(r []int) bool { return r[0] == 6 && r[1] == 6 }, 0, 2},
		{"a 6 on a d20", 20, 1, func(r []int) bool { return r[0] == 6 }, 0, 0},
	} {
		source := uuid.New()
		key := d5681Key(t, source, tc.sides, tc.n, tc.rolls)
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(key)
		me := g.Seats[0]
		advanceTo(t, g, game.StepPrecombatMain)
		p2Push(g, me.ID, "Netherese Puzzle-Ward", "Enchantment", netheresePuzzleWardOracle, 0, 0)
		before := handSize(me)
		d5681Roll(t, g, me.ID, source, tc.sides, tc.n, tc.modifier)
		if got := handSize(me) - before; got != tc.draws {
			t.Errorf("%s: drew %d, want %d", tc.name, got, tc.draws)
		}
	}
}

// An ignored roll never happened (CR 706.6): with Pixie Guide a d6 rolls
// two dice, and only the kept one can be the highest natural result.
func TestNetheresePuzzleWardIgnoredRollDoesNotTrigger(t *testing.T) {
	source := uuid.New()
	key := d5681Key(t, source, 6, 3, func(r []int) bool { return r[0] == 6 && r[1] == 6 && r[2] < 6 })
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(key)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	p2Push(g, me.ID, "Netherese Puzzle-Ward", "Enchantment", netheresePuzzleWardOracle, 0, 0)
	p2Push(g, me.ID, "Pixie Guide", "Creature — Faerie", pixieGuideOracle, 1, 3)
	p2Push(g, me.ID, "Pixie Guide", "Creature — Faerie", pixieGuideOracle, 1, 3)
	before := handSize(me)
	d5681Roll(t, g, me.ID, source, 6, 1, 0)
	if got := handSize(me) - before; got != 1 {
		t.Errorf("drew %d, want 1: of 6, 6 and a lower roll only one six was kept", got)
	}
}

// --- Critical Hit ---------------------------------------------------------

func TestCriticalHitGrantsDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Critical Hit", TypeLine: "Instant", ManaCost: "{1}{R}",
		OracleID: criticalHitOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{p2Card(bear)}}); err == nil {
		t.Fatal("cast with an empty pool")
	}
	floatForTest(g, me, "RR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{p2Card(bear)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !p2Keyword(g, bear, "double strike") {
		t.Error("the target did not gain double strike")
	}
	if !d5681InZone(g, id, game.ZoneGraveyard) {
		t.Error("Critical Hit is not in the graveyard")
	}
}

// From the graveyard: a die showing 20 returns it, whatever the modifier;
// a result of 20 on a lower face does not, nor does an opponent's 20, and
// it does nothing from the hand's side of things once returned.
func TestCriticalHitReturnsOnANatural20(t *testing.T) {
	for _, tc := range []struct {
		name     string
		natural  int
		modifier int
		opponent bool
		returns  bool
	}{
		{"a natural 20", 20, 0, false, true},
		{"a natural 20 minus 5", 20, -5, false, true},
		{"a 17 plus 3", 17, 3, false, false},
		{"an opponent's natural 20", 20, 0, true, false},
	} {
		source := uuid.New()
		key := d5681Key(t, source, 20, 1, func(r []int) bool { return r[0] == tc.natural })
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(key)
		me := g.Seats[0]
		advanceTo(t, g, game.StepPrecombatMain)
		owner := me
		if tc.opponent {
			owner = g.Seats[1]
		}
		id := uuid.New()
		g.WithWriteLock(func() {
			owner.Graveyard.PushTop(game.Card{
				InstanceID: id, Name: "Critical Hit", TypeLine: "Instant", ManaCost: "{1}{R}",
				OracleID: criticalHitOracle, Owner: owner.ID, Controller: owner.ID,
			})
		})
		from := sfkEventCount(g)
		d5681Roll(t, g, me.ID, source, 20, 1, tc.modifier)
		if got := d5681Rolled(g, from); len(got) != 1 || got[0].Amount != tc.natural {
			t.Fatalf("%s: rolled %+v", tc.name, got)
		}
		if got := d5681InZone(g, id, game.ZoneHand); got != tc.returns {
			t.Errorf("%s: returned to hand = %v, want %v", tc.name, got, tc.returns)
		}
	}
}

// Diviner's Portent's hand is the roll's modifier now: the die is
// announced with its natural result and the hand beside it.
func TestDivinersPortentAnnouncesItsModifier(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Diviner's Portent", TypeLine: "Instant", ManaCost: "{X}{U}{U}{U}",
		OracleID: "119585c7-ddfa-47ed-b2f8-488ebc156222", Owner: me.ID, Controller: me.ID,
	})
	floatForTest(g, me, "UUU")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := handSize(me)
	from := sfkEventCount(g)
	passPriorityAroundTable(t, g)
	got := d5681Rolled(g, from)
	if len(got) != 1 || got[0].Modifier != hand || got[0].RollResult() != got[0].Amount+hand {
		t.Fatalf("rolled %+v, want one d20 with the hand of %d as its modifier", got, hand)
	}
}
