package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// self_cost_modifier_test.go — ADR 0048 addendum (#746): a spell's own
// cost modifier. What a bug here would hide:
//
//  1. THE SLOT. A self modifier applies to its own card and to nothing
//     else: not to another spell while it sits in a hand, not from the
//     battlefield. A Sphere of Resistance in a hand taxes nobody.
//  2. THE PASSES. A self modifier joins the board's increase,
//     reduction and floor passes, so every increase lands before any
//     reduction and a reduction still stops at generic mana.
//  3. THE CARD AS CAST. Any zone, any alternative cost, and the face
//     being cast — never the other face's modifiers.
//  4. TARGETS. Shown only to a modifier that declares ReadsTargets.

// withCatalogDefs wires a stub catalog for the duration of a test: the
// one CatalogLookup hook the per-slot defaults read through.
func withCatalogDefs(t *testing.T, defs map[string]*CardDef) {
	t.Helper()
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

// selfReducer is a CostReduction of a fixed amount.
func selfReducer(label string, n int) CostModifier {
	return CostModifier{Kind: CostReduction, Label: label, Amount: fixed(n)}
}

// catalogSpellInHand seeds a spell carrying `oracle` in the seat's hand
// at a main phase.
func catalogSpellInHand(t *testing.T, g *Game, p *Player, name, typeLine, manaCost, oracle string) uuid.UUID {
	t.Helper()
	id := spellInHand(t, g, p, name, typeLine, manaCost)
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].OracleID = oracle
		}
	}
	return id
}

// A reduction bigger than the generic component spends the generic
// component and stops: Ghalta, Primal Hunger with twenty power is
// {G}{G}, never less, and never touches a colour.
func TestSelfCostReductionFloorsAtGeneric(t *testing.T) {
	const ghalta = "test-ghalta"
	withCatalogDefs(t, map[string]*CardDef{
		ghalta: {SelfCostModifiers: []CostModifier{selfReducer("{X} less", 20)}},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	id := catalogSpellInHand(t, g, me, "Ghalta", "Legendary Creature — Dinosaur", "{10}{G}{G}", ghalta)

	got := priceOf(t, g, me, id, CastSpellParams{})
	if got.Generic != 0 || len(got.Required) != 2 {
		t.Fatalf("Ghalta under a 20 reduction = %+v, want exactly {G}{G}", got)
	}
	for _, r := range got.Required {
		if len(r.Options) != 1 || r.Options[0] != "G" {
			t.Errorf("a coloured requirement changed: %+v", got.Required)
		}
	}
}

// CR 601.2f: every increase, from the board and from the spell, lands
// before any reduction. Ghalta at {10}{G}{G} under Sphere of Resistance
// with an 11 reduction: increase first is {11} − 11 = {G}{G}; the
// reduction first would be {10} − 11 → {0}, + 1 = {1}{G}{G}.
func TestSelfCostReductionRunsAfterBoardIncreases(t *testing.T) {
	const ghalta, sphere = "test-ghalta", "test-sphere"
	withCatalogDefs(t, map[string]*CardDef{
		ghalta: {SelfCostModifiers: []CostModifier{selfReducer("{X} less", 11)}},
	})
	withCatalogCostModifiers(t, modifiersFor(sphere, CostModifier{Kind: CostIncrease, Label: "{1} more", Amount: fixed(1)}))
	g := newActiveGame(t)
	me := g.Seats[0]
	modifierSource(t, g, me, "Sphere of Resistance", sphere)
	id := catalogSpellInHand(t, g, me, "Ghalta", "Legendary Creature — Dinosaur", "{10}{G}{G}", ghalta)

	if got := priceOf(t, g, me, id, CastSpellParams{}); got.Generic != 0 || len(got.Required) != 2 {
		t.Errorf("Ghalta under a Sphere with 11 reduction = %+v, want {G}{G} (increase before reduction)", got)
	}
}

// The same order with the reducer on a noncreature spell and the
// increase from Thalia: a {R} spell with "costs {1} less" under Thalia
// is {1}{R} − 1 = {R}. Reversed it would be {1}{R}.
func TestSelfCostReductionEatsThaliaTax(t *testing.T) {
	const bolt, thalia = "test-self-bolt", "test-thalia"
	withCatalogDefs(t, map[string]*CardDef{
		bolt: {SelfCostModifiers: []CostModifier{selfReducer("{1} less", 1)}},
	})
	withCatalogCostModifiers(t, modifiersFor(thalia, CostModifier{
		Kind: CostIncrease, Label: "Noncreature spells cost {1} more",
		AppliesTo: func(q CostQuery) bool { return !q.Card.IsCreature() },
		Amount:    fixed(1),
	}))
	g := newActiveGame(t)
	me := g.Seats[0]
	modifierSource(t, g, me, "Thalia", thalia)
	id := catalogSpellInHand(t, g, me, "Self Bolt", "Instant", "{R}", bolt)
	if got := priceOf(t, g, me, id, CastSpellParams{}); got.Generic != 0 || len(got.Required) != 1 {
		t.Errorf("self-reducing {R} under Thalia = %+v, want {R}", got)
	}
}

// A modifier in the battlefield slot works from the battlefield only:
// a Sphere of Resistance in the caster's hand, or an opponent's, taxes
// nothing. And a self modifier on a card in hand changes nothing about
// another spell.
func TestModifiersInHandTaxNobody(t *testing.T) {
	const sphere, reducer = "test-sphere", "test-reducer"
	withCatalogCostModifiers(t, modifiersFor(sphere, CostModifier{Kind: CostIncrease, Label: "{1} more", Amount: fixed(1)}))
	withCatalogDefs(t, map[string]*CardDef{
		sphere:  {CostModifiers: []CostModifier{{Kind: CostIncrease, Label: "{1} more", Amount: fixed(1)}}},
		reducer: {SelfCostModifiers: []CostModifier{selfReducer("{5} less", 5)}},
	})
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	catalogSpellInHand(t, g, me, "Sphere of Resistance", "Artifact", "{2}", sphere)
	catalogSpellInHand(t, g, them, "Sphere of Resistance", "Artifact", "{2}", sphere)
	catalogSpellInHand(t, g, me, "Self Reducer", "Sorcery", "{5}{U}", reducer)
	bolt := spellInHand(t, g, me, "Lightning Bolt", "Instant", "{R}")
	big := spellInHand(t, g, me, "Big Spell", "Sorcery", "{4}{U}")

	if got := priceOf(t, g, me, bolt, CastSpellParams{}); got.Generic != 0 || len(got.Required) != 1 {
		t.Errorf("Bolt with Spheres in two hands = %+v, want {R}", got)
	}
	if got := priceOf(t, g, me, big, CastSpellParams{}); got.Generic != 4 {
		t.Errorf("another spell next to a self reducer in hand: generic = %d, want 4", got.Generic)
	}
}

// A self modifier is never read from the battlefield: Thought Monitor
// on the battlefield does not give another artifact spell affinity.
func TestSelfModifierOnTheBattlefieldDoesNothing(t *testing.T) {
	const monitor = "test-monitor"
	withCatalogDefs(t, map[string]*CardDef{
		monitor: {SelfCostModifiers: []CostModifier{selfReducer("Affinity for artifacts", 3)}},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	modifierSource(t, g, me, "Thought Monitor", monitor)
	id := spellInHand(t, g, me, "Other Artifact", "Artifact Creature — Construct", "{4}")
	if got := priceOf(t, g, me, id, CastSpellParams{}); got.Generic != 4 {
		t.Errorf("artifact spell next to a battlefield Thought Monitor: generic = %d, want 4", got.Generic)
	}
}

// CR 113.6d: the modifier functions on the stack, so it applies from
// the command zone (after the commander tax, which is part of what the
// spell would cost) and from the graveyard.
func TestSelfModifierAppliesFromCommandZoneAndGraveyard(t *testing.T) {
	const cmdr, flash = "test-self-commander", "test-self-flashback"
	withCatalogDefs(t, map[string]*CardDef{
		cmdr:  {SelfCostModifiers: []CostModifier{selfReducer("{5} less", 5)}},
		flash: {SelfCostModifiers: []CostModifier{selfReducer("{2} less", 2)}, CastableZones: []ZoneKind{ZoneGraveyard}},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)

	c := NewCard("Test Commander", me.ID)
	c.TypeLine = "Legendary Creature — Elemental"
	c.ManaCost = "{4}{G}"
	c.OracleID = cmdr
	me.Command.PushTop(c)
	if me.CommanderCasts == nil {
		me.CommanderCasts = make(map[uuid.UUID]int)
	}
	me.CommanderCasts[c.InstanceID] = 2 // +{4}
	if got := priceOf(t, g, me, c.InstanceID, CastSpellParams{FromZone: "command"}); got.Generic != 3 {
		t.Errorf("taxed commander with a {5} self reduction: generic = %d, want 3 ({4} + {4} tax − 5)", got.Generic)
	}

	gy := NewCard("Test Flashback", me.ID)
	gy.TypeLine = "Sorcery"
	gy.ManaCost = "{3}{R}"
	gy.OracleID = flash
	me.Graveyard.PushTop(gy)
	if got := priceOf(t, g, me, gy.InstanceID, CastSpellParams{FromZone: "graveyard"}); got.Generic != 1 {
		t.Errorf("graveyard cast with a {2} self reduction: generic = %d, want 1", got.Generic)
	}
}

// CR 118.9d: increases and reductions apply to an alternative cost too.
func TestSelfModifierAppliesToAnAlternativeCost(t *testing.T) {
	const card = "test-self-alt"
	withCatalogDefs(t, map[string]*CardDef{
		card: {
			SelfCostModifiers: []CostModifier{selfReducer("{2} less", 2)},
			AlternativeCosts:  []AlternativeCost{{Key: "overload", Label: "Overload {5}{U}", ManaCost: "{5}{U}"}},
		},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	id := catalogSpellInHand(t, g, me, "Test Overload", "Instant", "{1}{U}", card)
	if got := priceOf(t, g, me, id, CastSpellParams{AlternativeCost: "overload"}); got.Generic != 3 {
		t.Errorf("overload {5}{U} with a {2} self reduction: generic = %d, want 3", got.Generic)
	}
	if got := priceOf(t, g, me, id, CastSpellParams{}); got.Generic != 0 {
		t.Errorf("printed {1}{U} with a {2} self reduction: generic = %d, want 0", got.Generic)
	}
}

// An MDFC uses the self modifiers of the face being cast and no others.
func TestSelfModifierReadsTheFaceBeingCast(t *testing.T) {
	const oracle = "test-mdfc"
	withCatalogDefs(t, map[string]*CardDef{
		oracle:                       {SelfCostModifiers: []CostModifier{selfReducer("{2} less", 2)}},
		CatalogKeyForFace(oracle, 1): {},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	c := NewCard("Front // Back", me.ID)
	c.OracleID = oracle
	c.Layout = LayoutModalDFC
	c.Faces = []Face{
		{Name: "Front", TypeLine: "Sorcery", ManaCost: "{3}{B}"},
		{Name: "Back", TypeLine: "Sorcery", ManaCost: "{3}{U}"},
	}
	c.SetFace(0)
	me.Hand.PushTop(c)

	front, _ := g.LookupCardForEffect(c.InstanceID)
	back := front
	back.SetFace(1)
	g.mu.Lock()
	defer g.mu.Unlock()
	gotFront, err := g.effectiveCostLocked(me, front, CastSpellParams{})
	if err != nil {
		t.Fatal(err)
	}
	gotBack, err := g.effectiveCostLocked(me, back, CastSpellParams{Face: 1})
	if err != nil {
		t.Fatal(err)
	}
	if gotFront.Generic != 1 {
		t.Errorf("front face with a {2} self reduction: generic = %d, want 1", gotFront.Generic)
	}
	if gotBack.Generic != 3 {
		t.Errorf("back face must not read the front's reduction: generic = %d, want 3", gotBack.Generic)
	}
}

// Fireball: a per-target increase that reads targets. Only a modifier
// that declares ReadsTargets sees them; any other sees nil.
func TestPerTargetIncreaseReadsTargetsOnlyWhenDeclared(t *testing.T) {
	const fireball, peeker = "test-fireball", "test-peeker"
	perTarget := func(q CostQuery) int {
		if n := RealTargetCount(q.Targets); n > 1 {
			return n - 1
		}
		return 0
	}
	sawTargets := -1
	withCatalogDefs(t, map[string]*CardDef{
		fireball: {SelfCostModifiers: []CostModifier{{Kind: CostIncrease, Label: "{1} more per target beyond the first", Amount: perTarget, ReadsTargets: true}}},
		peeker: {SelfCostModifiers: []CostModifier{{Kind: CostIncrease, Label: "reads targets without saying so", Amount: func(q CostQuery) int {
			sawTargets = len(q.Targets)
			return perTarget(q)
		}}}},
	})
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := catalogSpellInHand(t, g, me, "Fireball", "Sorcery", "{X}{R}", fireball)
	refs := []TargetRef{
		{Kind: TargetPlayer, ID: them.ID},
		{Kind: TargetPlayer, ID: me.ID},
		{Kind: TargetCard, ID: uuid.New()},
	}
	for n, want := range map[int]int{0: 0, 1: 0, 2: 1, 3: 2} {
		got := priceOf(t, g, me, id, CastSpellParams{XValue: 4, Targets: refs[:n]})
		if got.Generic != want || got.XSlots != 1 || len(got.Required) != 1 {
			t.Errorf("Fireball at %d targets = %+v, want {X}{%d}{R}", n, got, want)
		}
	}
	// The placeholders the validator ignores are not targets.
	got := priceOf(t, g, me, id, CastSpellParams{Targets: []TargetRef{{Kind: TargetNone}, {Kind: TargetSelf}, refs[0]}})
	if got.Generic != 0 {
		t.Errorf("one real target among placeholders: generic = %d, want 0", got.Generic)
	}

	sneaky := catalogSpellInHand(t, g, me, "Peeker", "Sorcery", "{R}", peeker)
	if got := priceOf(t, g, me, sneaky, CastSpellParams{Targets: refs}); got.Generic != 0 || sawTargets != 0 {
		t.Errorf("a modifier without ReadsTargets saw %d targets and charged %d; want nil targets and no surcharge", sawTargets, got.Generic)
	}
}

// Strive: a coloured unit adds its coloured symbols per point, and a
// reduction on the board still only spends generic mana.
func TestColouredIncreaseUnitAddsColouredRequirements(t *testing.T) {
	const coppercoats, mancer = "test-coppercoats", "test-mancer"
	unit, err := ParseCost("{1}{W}")
	if err != nil {
		t.Fatal(err)
	}
	withCatalogDefs(t, map[string]*CardDef{
		coppercoats: {SelfCostModifiers: []CostModifier{{
			Kind: CostIncrease, Label: "Strive", ReadsTargets: true, Unit: &unit,
			Amount: func(q CostQuery) int { return RealTargetCount(q.Targets) - 1 },
		}}},
	})
	withCatalogCostModifiers(t, modifiersFor(mancer, CostModifier{Kind: CostReduction, Label: "{5} less", Amount: fixed(5)}))
	g := newActiveGame(t)
	me := g.Seats[0]
	id := catalogSpellInHand(t, g, me, "Call the Coppercoats", "Instant", "{2}{W}", coppercoats)
	three := []TargetRef{{Kind: TargetPlayer, ID: g.Seats[1].ID}, {Kind: TargetPlayer, ID: uuid.New()}, {Kind: TargetPlayer, ID: uuid.New()}}

	got := priceOf(t, g, me, id, CastSpellParams{Targets: three})
	whites := 0
	for _, r := range got.Required {
		if len(r.Options) == 1 && r.Options[0] == "W" {
			whites++
		}
	}
	if got.Generic != 4 || whites != 3 {
		t.Fatalf("strive at three targets = %+v, want {4}{W}{W}{W}", got)
	}

	modifierSource(t, g, me, "Reducer", mancer)
	got = priceOf(t, g, me, id, CastSpellParams{Targets: three})
	if got.Generic != 0 || len(got.Required) != 3 {
		t.Errorf("strive at three targets under a {5} reduction = %+v, want {W}{W}{W}", got)
	}
}

// mustParse parses a mana string a test spells out by hand.
func mustParse(t *testing.T, s string) *ParsedCost {
	t.Helper()
	c, err := ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return &c
}

// ADR 0048 §4, for the self slot: a negative amount refuses the cast
// and names the card; so does a mana unit on a reduction, and a unit on
// an increase that carries {X}, a hybrid or a Phyrexian symbol (§16).
// effects.Register refuses all of these at boot; this is the engine's
// own refusal for a catalog hook that got past it.
func TestBrokenSelfModifierRefusesTheCast(t *testing.T) {
	const broken, badUnit = "test-self-broken", "test-self-bad-unit"
	const xUnit, hybridUnit, phyrexianUnit = "test-self-x-unit", "test-self-hybrid-unit", "test-self-phyrexian-unit"
	increaseOf := func(unit string) []CostModifier {
		return []CostModifier{{Kind: CostIncrease, Label: "odd surcharge", Amount: fixed(1), Unit: mustParse(t, unit)}}
	}
	withCatalogDefs(t, map[string]*CardDef{
		broken:        {SelfCostModifiers: []CostModifier{{Kind: CostReduction, Label: "sign error", Amount: fixed(-2)}}},
		badUnit:       {SelfCostModifiers: []CostModifier{{Kind: CostReduction, Label: "coloured reduction", Amount: fixed(1), Unit: mustParse(t, "{W}")}}},
		xUnit:         {SelfCostModifiers: increaseOf("{X}")},
		hybridUnit:    {SelfCostModifiers: increaseOf("{1}{W/U}")},
		phyrexianUnit: {SelfCostModifiers: increaseOf("{W/P}")},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	for _, oracle := range []string{broken, badUnit, xUnit, hybridUnit, phyrexianUnit} {
		id := catalogSpellInHand(t, g, me, "Broken Spell", "Sorcery", "{2}{R}", oracle)
		err := g.CastSpell(me.ID, id, CastSpellParams{})
		if !errors.Is(err, ErrCostModifier) {
			t.Fatalf("%s: got %v, want ErrCostModifier", oracle, err)
		}
		if !strings.Contains(err.Error(), "Broken Spell") {
			t.Errorf("%s: refusal does not name the card: %v", oracle, err)
		}
		if !me.Hand.Contains(id) {
			t.Errorf("%s: refused cast left the card out of hand", oracle)
		}
	}
}

// A real strict cast charges the target-priced total: two targets at
// X=2 need {2}{1}{R}, and a pool one short is refused.
func TestStrictCastPaysTheTargetSurcharge(t *testing.T) {
	const fireball = "test-fireball"
	withCatalogDefs(t, map[string]*CardDef{
		fireball: {
			Targets: &TargetSpec{Mode: "player", Players: true, Min: 0, Max: 0},
			SelfCostModifiers: []CostModifier{{
				Kind: CostIncrease, Label: "{1} more per target beyond the first", ReadsTargets: true,
				Amount: func(q CostQuery) int {
					if n := RealTargetCount(q.Targets); n > 1 {
						return n - 1
					}
					return 0
				},
			}},
		},
	})
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := catalogSpellInHand(t, g, me, "Fireball", "Sorcery", "{X}{R}", fireball)
	two := []TargetRef{{Kind: TargetPlayer, ID: them.ID}, {Kind: TargetPlayer, ID: me.ID}}

	me.ManaPool.AddMana(ManaToken{Color: "R"}, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	err := g.CastSpell(me.ID, id, CastSpellParams{XValue: 2, Targets: two, Strict: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("three mana for {2}{1}{R}: got %v, want InsufficientManaError", err)
	}
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{XValue: 2, Targets: two, Strict: true}); err != nil {
		t.Fatalf("four mana for {2}{1}{R}: %v", err)
	}
}

// strive is the Call the Coppercoats fixture: {2}{W}, "{1}{W} more for
// each target beyond the first", targeting any number of players.
func strive(t *testing.T) *CardDef {
	return &CardDef{
		Targets: &TargetSpec{Mode: "player", Players: true, Min: 0, Max: 0},
		SelfCostModifiers: []CostModifier{{
			Kind: CostIncrease, Label: "Strive", ReadsTargets: true, Unit: mustParse(t, "{1}{W}"),
			Amount: func(q CostQuery) int {
				if n := RealTargetCount(q.Targets); n > 1 {
					return n - 1
				}
				return 0
			},
		}},
	}
}

func whiteSymbols(c ParsedCost) int {
	n := 0
	for _, r := range c.Required {
		if len(r.Options) == 1 && r.Options[0] == "W" {
			n++
		}
	}
	return n
}

// ADR 0048 addendum §16's ordering cases for a coloured increase. A
// two-target Call the Coppercoats is {3}{W}{W}. Under Goblin
// Electromancer ({1} less) it is {2}{W}{W}: the reduction spends
// generic mana and both {W} stay. Under Sphere of Resistance and the
// Electromancer it is {3}{W}{W}: both increases land before the
// reduction (CR 601.2f).
func TestColouredIncreaseUnderElectromancerAndSphere(t *testing.T) {
	const coppercoats, mancer, sphere = "test-coppercoats", "test-electromancer", "test-sphere"
	withCatalogDefs(t, map[string]*CardDef{coppercoats: strive(t)})
	mods := map[string][]CostModifier{
		mancer: {{Kind: CostReduction, Label: "{1} less", Amount: fixed(1)}},
		sphere: {{Kind: CostIncrease, Label: "{1} more", Amount: fixed(1)}},
	}
	withCatalogCostModifiers(t, func(id string) []CostModifier { return mods[id] })
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	id := catalogSpellInHand(t, g, me, "Call the Coppercoats", "Instant", "{2}{W}", coppercoats)
	two := []TargetRef{{Kind: TargetPlayer, ID: them.ID}, {Kind: TargetPlayer, ID: me.ID}}

	modifierSource(t, g, me, "Goblin Electromancer", mancer)
	if got := priceOf(t, g, me, id, CastSpellParams{Targets: two}); got.Generic != 2 || whiteSymbols(got) != 2 || len(got.Required) != 2 {
		t.Errorf("two-target strive under Electromancer = %+v, want {2}{W}{W}", got)
	}
	modifierSource(t, g, them, "Sphere of Resistance", sphere)
	if got := priceOf(t, g, me, id, CastSpellParams{Targets: two}); got.Generic != 3 || whiteSymbols(got) != 2 || len(got.Required) != 2 {
		t.Errorf("two-target strive under Sphere and Electromancer = %+v, want {3}{W}{W}", got)
	}
}

// The added {W} is a real requirement at payment. A pool with enough
// mana in total but one {W} short refuses the strict two-target cast,
// and the auto-tapper taps a second white source for it.
func TestColouredIncreaseIsPaidInColour(t *testing.T) {
	const coppercoats = "test-coppercoats"
	withCatalogDefs(t, map[string]*CardDef{coppercoats: strive(t)})

	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	two := []TargetRef{{Kind: TargetPlayer, ID: them.ID}, {Kind: TargetPlayer, ID: me.ID}}
	id := catalogSpellInHand(t, g, me, "Call the Coppercoats", "Instant", "{2}{W}", coppercoats)
	me.ManaPool.AddMana(ManaToken{Color: "W"}, ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	err := g.CastSpell(me.ID, id, CastSpellParams{Targets: two, Strict: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("{W} and five {C} for {3}{W}{W}: got %v, want InsufficientManaError", err)
	}
	if !me.Hand.Contains(id) {
		t.Fatal("refused cast left the card out of hand")
	}

	g = newActiveGame(t)
	me, them = g.Seats[0], g.Seats[1]
	two = []TargetRef{{Kind: TargetPlayer, ID: them.ID}, {Kind: TargetPlayer, ID: me.ID}}
	id = catalogSpellInHand(t, g, me, "Call the Coppercoats", "Instant", "{2}{W}", coppercoats)
	plains := []uuid.UUID{
		pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", ""),
		pushBattlefieldForTest(g, me.ID, "Plains", "Basic Land — Plains", ""),
	}
	for i := 0; i < 4; i++ {
		pushBattlefieldForTest(g, me.ID, "Mountain", "Basic Land — Mountain", "")
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Targets: two, Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap two-target strive off two Plains and four Mountains: %v", err)
	}
	for _, p := range plains {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == p && !c.Tapped {
				t.Errorf("a Plains stayed untapped: the added {W} was not paid in white")
			}
		}
	}
}

// The enumerator's switch: true only when the card or the board
// declares a target-reading modifier.
func TestCastPriceReadsTargets(t *testing.T) {
	const plain, fireball, kopala = "test-plain", "test-fireball", "test-kopala"
	withCatalogDefs(t, map[string]*CardDef{
		plain:    {SelfCostModifiers: []CostModifier{selfReducer("{1} less", 1)}},
		fireball: {SelfCostModifiers: []CostModifier{{Kind: CostIncrease, Label: "per target", ReadsTargets: true, Amount: fixed(0)}}},
	})
	g := newActiveGame(t)
	me := g.Seats[0]
	g.mu.Lock()
	if g.CastPriceReadsTargetsForEffect(Card{OracleID: plain}) {
		t.Error("a self modifier without ReadsTargets must not ask for per-target pricing")
	}
	if !g.CastPriceReadsTargetsForEffect(Card{OracleID: fireball}) {
		t.Error("a target-reading self modifier must ask for per-target pricing")
	}
	g.mu.Unlock()

	withCatalogCostModifiers(t, modifiersFor(kopala, CostModifier{Kind: CostIncrease, Label: "targets a Merfolk", ReadsTargets: true, Amount: fixed(2)}))
	modifierSource(t, g, me, "Kopala", kopala)
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.CastPriceReadsTargetsForEffect(Card{OracleID: plain}) {
		t.Error("a target-reading battlefield modifier must ask for per-target pricing of every spell")
	}
	if got := TargetPricedCostClauses(fireball); len(got) != 1 || got[0] != "per target" {
		t.Errorf("TargetPricedCostClauses = %v, want the target-reading clause", got)
	}
	if got := TargetPricedCostClauses(plain); got != nil {
		t.Errorf("TargetPricedCostClauses for a plain reducer = %v, want nil", got)
	}
}
