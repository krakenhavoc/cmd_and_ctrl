package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func sacNCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, name, "Creature — Bear", "", false)
}

// The count lives on the clause, and Plus carries the clause whole —
// a merge that dropped the count would price "Sacrifice two" at one.
func TestSacrificeNConstructorsCarryTheCount(t *testing.T) {
	cost := effects.Plus(effects.TapCost(), effects.SacrificeN(3, "three artifacts", effects.Artifact()), effects.ManaCost("{1}"))
	if cost.SacrificeOther == nil {
		t.Fatal("Plus dropped the sacrifice clause")
	}
	if got := game.SacrificeCostCount(cost.SacrificeOther); got != 3 {
		t.Errorf("SacrificeCostCount = %d, want 3", got)
	}
	if cost.SacrificeOther.Min != 3 || cost.SacrificeOther.Max != 3 || cost.SacrificeOther.Label != "three artifacts" {
		t.Errorf("clause = %d..%d %q, want 3..3 \"three artifacts\"", cost.SacrificeOther.Min, cost.SacrificeOther.Max, cost.SacrificeOther.Label)
	}
	if !cost.Tap || cost.Mana != "{1}" {
		t.Errorf("the other components were lost: %+v", cost)
	}
	add := effects.SacrificeNCost(2, "two creatures", effects.Creature())
	if game.SacrificeCostCount(add.Sacrifice) != 2 || add.Label != "Sacrifice two creatures" {
		t.Errorf("SacrificeNCost = %+v", add)
	}
	if got := game.SacrificeCostCount(effects.SacrificeACreature().SacrificeOther); got != 1 {
		t.Errorf("the one-permanent constructors still count 1, got %d", got)
	}
}

func TestSacrificeNActivatedAbilityPaysExactlyN(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	outlet := effectstest.PushCatalogPermanent(g, me.ID, "Two-Creature Outlet", "Enchantment", effectstest.SacNAbilityOracle, false)
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")
	c := sacNCreature(g, me.ID, "Bear C")
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, outlet, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both named creatures are paid at announce")
	}
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("both sacrificed creatures go to the graveyard")
	}
	if !g.Battlefield.Contains(c) {
		t.Error("the unnamed creature is untouched")
	}
	if got := effectstest.CountEventsOf(g, game.EventSacrifice, a, b); got != 2 {
		t.Errorf("%d EventSacrifice, want one per permanent", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want the ability to resolve once", life, me.Life)
	}
}

// Every malformed payment refuses the whole activation with nothing
// paid: no creature moved, the source not sacrificed, nothing on the
// stack.
func TestSacrificeNRefusedPaymentsPayNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	outlet := effectstest.PushCatalogPermanent(g, me.ID, "Two-Creature Outlet", "Enchantment", effectstest.SacNAbilityOracle, false)
	selfToo := effectstest.PushCatalogPermanent(g, me.ID, "Self And Two Outlet", "Creature — Horror", effectstest.SacNSelfTooOracle, false)
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")
	c := sacNCreature(g, me.ID, "Bear C")
	theirs := sacNCreature(g, opp.ID, "Their Bear")
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

	cases := []struct {
		name   string
		source uuid.UUID
		ids    []uuid.UUID
		want   error
	}{
		{"one ID for a two-permanent cost", outlet, []uuid.UUID{a}, game.ErrInvalidParam},
		{"three IDs", outlet, []uuid.UUID{a, b, c}, game.ErrInvalidParam},
		{"no IDs", outlet, nil, game.ErrInvalidParam},
		{"the same ID twice", outlet, []uuid.UUID{a, a}, game.ErrInvalidParam},
		{"an opponent's creature", outlet, []uuid.UUID{a, theirs}, game.ErrCardCallerMismatch},
		{"a permanent the clause doesn't match", outlet, []uuid.UUID{a, rock}, game.ErrIllegalTarget},
		{"a card not on the battlefield", outlet, []uuid.UUID{a, uuid.New()}, game.ErrCardNotFound},
		{"the source when it is sacrificed too", selfToo, []uuid.UUID{a, selfToo}, game.ErrInvalidParam},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stack := len(g.StackMeta)
			err := g.ActivateCatalogAbility(me.ID, tc.source, 0, game.ActivateAbilityParams{SacrificeIDs: tc.ids})
			if err != tc.want {
				t.Errorf("got %v, want %v", err, tc.want)
			}
			for _, id := range []uuid.UUID{outlet, selfToo, a, b, c, theirs, rock} {
				if !g.Battlefield.Contains(id) {
					t.Errorf("%s left the battlefield on a refused payment", effectstest.CardNameOf(g, id))
				}
			}
			if len(g.StackMeta) != stack {
				t.Error("a refused activation put something on the stack")
			}
		})
	}
	if effectstest.CountEventsOf(g, game.EventSacrifice, outlet, selfToo, a, b, c, theirs, rock) != 0 {
		t.Error("a refused payment emitted EventSacrifice")
	}
	// The same board pays the self-too cost legally: the source plus
	// two others, three sacrifices.
	if err := g.ActivateCatalogAbility(me.ID, selfToo, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("legal self-too payment: %v", err)
	}
	if got := effectstest.CountEventsOf(g, game.EventSacrifice, selfToo, a, b); got != 3 {
		t.Errorf("%d EventSacrifice, want 3 (the source and both creatures)", got)
	}
}

// The refusal leaves the tap state and the mana pool alone too: a
// {T} plus sacrifice-three cost (Kuldotha Forgemaster) and a mana plus
// sacrifice-two cost (Sai) named with too few permanents neither tap
// the source nor spend the floating mana.
func TestSacrificeNRefusalLeavesTapAndPoolAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forge := effectstest.PushCatalogPermanent(g, me.ID, "Kuldotha Forgemaster", "Artifact Creature — Construct", effectstest.KuldothaForgemasterOracle, false)
	sai := effectstest.PushCatalogPermanent(g, me.ID, "Sai, Master Thopterist", "Legendary Creature — Human Artificer", effectstest.B07SaiMasterThopteristOracle, false)
	arts := effectstest.PushTokens(g, me.ID, effects.TreasureToken(), 2)
	g.WithWriteLock(func() {
		_ = g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}")
	})
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{SacrificeIDs: arts}); err != game.ErrInvalidParam {
		t.Fatalf("two artifacts for three: got %v, want ErrInvalidParam", err)
	}
	if effectstest.B16Tapped(t, g, forge) {
		t.Error("a refused payment tapped the source")
	}
	if err := g.ActivateCatalogAbility(me.ID, sai, 0, game.ActivateAbilityParams{
		SacrificeIDs: arts[:1], Strict: true,
	}); err != game.ErrInvalidParam {
		t.Fatalf("one artifact for two: got %v, want ErrInvalidParam", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool = %d after a refused payment, want the 2 floating mana untouched", me.ManaPool.Total())
	}
	for _, id := range arts {
		if !g.Battlefield.Contains(id) {
			t.Error("a refused payment sacrificed an artifact")
		}
	}
}

func TestSacrificeNManaAbilityPaysExactlyN(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	altar := effectstest.PushCatalogPermanent(g, me.ID, "Two-Creature Altar", "Artifact", effectstest.SacNManaOracle, false)
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")

	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{a}}); err != game.ErrInvalidParam {
		t.Fatalf("one creature for a two-creature mana cost: got %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{b, b}}); err != game.ErrInvalidParam {
		t.Fatalf("the same creature twice: got %v, want ErrInvalidParam", err)
	}
	if me.ManaPool.Total() != 0 || !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("a refused mana payment produced mana or ate a creature")
	}
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.ManaPool.Total() != 3 {
		t.Errorf("pool = %d, want 3", me.ManaPool.Total())
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both creatures are sacrificed")
	}
	if got := effectstest.CountEventsOf(g, game.EventSacrifice, a, b); got != 2 {
		t.Errorf("%d EventSacrifice, want 2", got)
	}
}

func TestSacrificeNAdditionalCostPaysExactlyN(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}

	for _, bad := range [][]uuid.UUID{{a}, {a, a}, {a, b, uuid.New()}} {
		id, err := effectstest.TryCastWithSacrifice(g, "Two-Creature Rites", "Sorcery", effectstest.SacNSpellOracle, bad)
		if err == nil {
			t.Fatalf("cast accepted with sacrifice_ids %v", bad)
		}
		if !me.Hand.Contains(id) || g.Stack.Contains(id) {
			t.Error("a refused cast left the caster's hand")
		}
		me.Hand.Cards = nil
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("a refused cast ate a creature")
	}

	handBefore := me.Hand.Size()
	spell, err := effectstest.TryCastWithSacrifice(g, "Two-Creature Rites", "Sorcery", effectstest.SacNSpellOracle, []uuid.UUID{b, a})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell is not on the stack")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both creatures are paid while the spell is on the stack")
	}
	if got := effectstest.CountEventsOf(g, game.EventSacrifice, a, b); got != 2 {
		t.Errorf("%d EventSacrifice, want 2", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// CR 603.10a: the N permanents of one payment leave together, so a
// dies-watcher paid in alongside another creature sees both deaths —
// in either wire order. Paid one at a time, the watcher going first
// would be gone before the second death and one trigger would be lost.
func TestSacrificeNIsOneSimultaneousExitInEitherOrder(t *testing.T) {
	for _, watcherFirst := range []bool{true, false} {
		name := "watcher named second"
		if watcherFirst {
			name = "watcher named first"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			outlet := effectstest.PushCatalogPermanent(g, me.ID, "Two-Creature Outlet", "Enchantment", effectstest.SacNAbilityOracle, false)
			watcher := effectstest.PushCatalogPermanent(g, me.ID, "Dies Watcher", "Creature — Vampire", effectstest.SacNWatcherOracle, false)
			goblin := sacNCreature(g, me.ID, "Goblin")
			ids := []uuid.UUID{goblin, watcher}
			if watcherFirst {
				ids = []uuid.UUID{watcher, goblin}
			}
			life := me.Life
			if err := g.ActivateCatalogAbility(me.ID, outlet, 0, game.ActivateAbilityParams{SacrificeIDs: ids}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			// Two watcher triggers (its own death and the Goblin's)
			// plus the outlet's own 1 life.
			if got := me.Life - life; got != 3 {
				t.Errorf("gained %d life, want 3: the watcher sees both deaths", got)
			}
		})
	}
}

// ADR 0021 addendum: the additional cost's N permanents are one exit
// too, paid with the spell already on the stack.
func TestSacrificeNAdditionalCostIsOneSimultaneousExit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watcher := effectstest.PushCatalogPermanent(g, me.ID, "Dies Watcher", "Creature — Vampire", effectstest.SacNWatcherOracle, false)
	goblin := sacNCreature(g, me.ID, "Goblin")
	for i := 0; i < 4; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	life := me.Life
	if _, err := effectstest.TryCastWithSacrifice(g, "Two-Creature Rites", "Sorcery", effectstest.SacNSpellOracle, []uuid.UUID{watcher, goblin}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: the watcher sees both deaths", got)
	}
}

// The pre-#747 two-permanent case, SacrificeSelf plus a one-permanent
// SacrificeOther, is paid as one batch too. The source is the watcher
// here, and the source leads the payment, so it is sacrificed first:
// only the batch lets it see the other creature die after it.
func TestSacrificeSelfAndOtherIsOneSimultaneousExit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	outlet := effectstest.PushCatalogPermanent(g, me.ID, "Self And One Outlet", "Creature — Horror", effectstest.SacSelfOtherOracle, false)
	goblin := sacNCreature(g, me.ID, "Goblin")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, outlet, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{goblin}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: the source sees itself and the Goblin die", got)
	}
}

// ADR 0020 addendum §12: a sacrifice clause is a fixed count of at
// least one. Every other shape panics at boot, at all three sites, so
// no card can declare a variable count and have it read as something
// else — except the shapes #1213 and ADR 0100 §3 opened, each on the
// site that can announce it.
func TestRegisterRejectsVariableSacrificeClauses(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	ability := func(id string, spec *game.TargetSpec) effects.Spec {
		return effects.Spec{OracleID: id, Name: id, Activated: []effects.ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{SacrificeOther: spec}, Effect: noop,
		}}}
	}
	fromX := func(label string) *game.TargetSpec {
		s := effects.SacrificeSpec(label, effects.Creature())
		s.CountFromX = true
		return s
	}
	cases := []struct {
		name string
		spec effects.Spec
		want string
	}{
		{"count zero", ability("sac-guard-zero", effects.SacrificeSpec("none", effects.Creature()).WithCount(0, 0)), "at least one"},
		{"ceiling below floor", ability("sac-guard-inverted", effects.SacrificeSpec("three to two", effects.Creature()).WithCount(3, 2)), "ceiling is below the floor"},
		{"count from X beside an {X} mana cost", effects.Spec{OracleID: "sac-guard-two-x", Name: "sac-guard-two-x", Activated: []effects.ActivatedAbility{{
			Label:  "x",
			Cost:   game.AbilityCost{Mana: "{X}{B}", SacrificeOther: fromX("X creatures")},
			Effect: noop,
		}}}, "one announced X cannot pay both"},
		{"allow same", ability("sac-guard-same", func() *game.TargetSpec {
			s := effects.SacrificeSpec("two creatures", effects.Creature()).WithCount(2, 2)
			s.AllowSame = true
			return s
		}()), "AllowSame"},
		{"players", ability("sac-guard-players", func() *game.TargetSpec {
			s := effects.SacrificeSpec("a creature", effects.Creature())
			s.Players = true
			return s
		}()), "admits players"},
		{"count from X on a mana ability", effects.Spec{OracleID: "sac-guard-mana-x", Name: "sac-guard-mana-x", ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{SacrificeOther: fromX("X creatures")},
			Produced: "{C}",
		}}}, "no X to announce"},
		// ADR 0100 §3: a cast's mandatory slot prints "sacrifice X" and
		// "any number", never "one or more" with a floor above zero.
		{"one or more on a cast", effects.Spec{OracleID: "sac-guard-spell", Name: "sac-guard-spell", AdditionalCost: &game.AdditionalCost{
			Sacrifice: effects.SacrificeSpec("one or more creatures", effects.Creature()).WithCount(1, 0),
		}}, "additional cost"},
		// #1224: an AdditionalCost in OptionalCosts carries the same
		// Sacrifice clause the mandatory slot does (Constant Mists'
		// "Buyback—Sacrifice a land"), so the same guard has to catch a
		// variable count there too.
		{"optional cost site", effects.Spec{OracleID: "sac-guard-optional", Name: "sac-guard-optional", OptionalCosts: []game.AdditionalCost{{
			Optional:  true,
			Key:       game.BuybackKey,
			Label:     "Buyback—Sacrifice any number of creatures",
			Sacrifice: effects.SacrificeSpec("any number of creatures", effects.Creature()).WithCount(0, 0),
		}}}, "optional cost"},
		// "Sacrifice X" on a cast's MANDATORY slot is legal since ADR 0100
		// §3 (TestRegisterAcceptsCastVariableSacrifices); in an optional
		// cost there is still no X to announce for it.
		{"count from X on an optional cost", effects.Spec{OracleID: "sac-guard-spell-x", Name: "sac-guard-spell-x", OptionalCosts: []game.AdditionalCost{{
			Optional:  true,
			Key:       game.KickerKey,
			Label:     "Kicker—Sacrifice X creatures",
			Sacrifice: fromX("X creatures"),
		}}}, "no X to announce"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			effectstest.MustPanic(t, tc.want, func() { effects.Register(tc.spec) })
			if effects.Has(tc.spec.OracleID) {
				t.Error("a refused spec was registered anyway")
			}
		})
	}
}

// TestRegisterAcceptsFixedOptionalSacrificeClause is the positive
// control for #1224's guard: a well-formed optional sacrifice clause
// (fixed count of one, exactly what BuybackSacrifice / KickerSacrifice
// build) must still register cleanly. The catalog already proves this
// live — Constant Mists and Gatekeeper of Malakir register at package
// init and their own tests cast them — but this pins it directly
// against the guard so a future change to checkSacrificeClause cannot
// start refusing every optional sacrifice cost and have nothing catch
// it here.
func TestRegisterAcceptsFixedOptionalSacrificeClause(t *testing.T) {
	spec := effects.Spec{
		OracleID: "sac-guard-optional-good",
		Name:     "sac-guard-optional-good",
		OptionalCosts: []game.AdditionalCost{{
			Optional:  true,
			Key:       game.BuybackKey,
			Label:     "Buyback—Sacrifice a land",
			Sacrifice: effects.SacrificeSpec("a land", effects.Land()),
		}},
	}
	effects.Register(spec)
	if !effects.Has(spec.OracleID) {
		t.Error("a legal optional sacrifice clause was refused")
	}
}

// #1213 narrowed the guard above: the two printed VARIABLE shapes are
// declarable now, on the surfaces that can announce them. This is the
// other side of the same contract, and it is a test rather than a
// comment because "#747 refuses this" was true for a year and the
// panic message still says which shapes are mistakes.
func TestRegisterAcceptsAnnouncedSacrificeCounts(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	ability := func(id string, spec *game.TargetSpec) effects.Spec {
		return effects.Spec{OracleID: id, Name: id, Activated: []effects.ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{SacrificeOther: spec}, Effect: noop,
		}}}
	}
	fromX := func(label string) *game.TargetSpec {
		s := effects.SacrificeSpec(label, effects.Creature())
		s.CountFromX = true
		return s
	}
	cases := []struct {
		name string
		spec effects.Spec
	}{
		{"one or more on an activated ability", ability("sac-ok-open", effects.SacrificeSpec("one or more creatures", effects.Creature()).WithCount(1, 0))},
		{"a printed range", ability("sac-ok-range", effects.SacrificeSpec("one or two creatures", effects.Creature()).WithCount(1, 2))},
		{"count from X on an activated ability", ability("sac-ok-x", fromX("X creatures"))},
		{"one or more on a mana ability", effects.Spec{OracleID: "sac-ok-mana-open", Name: "sac-ok-mana-open", ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{SacrificeOther: effects.SacrificeSpec("one or more creatures", effects.Creature()).WithCount(1, 0)},
			Produced: "{C}",
		}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			effectstest.RegisterForTest(t, tc.spec)
			if !effects.Has(tc.spec.OracleID) {
				t.Fatal("a legal variable-count clause was not registered")
			}
		})
	}
}

// The bounds a variable clause reads are the WHOLE of the design, so
// they are pinned here beside the guard rather than left to the cards.
func TestSacrificeCostBoundsReadTheAnnouncement(t *testing.T) {
	fixed := effects.SacrificeSpec("two creatures", effects.Creature()).WithCount(2, 2)
	open := effects.SacrificeSpec("one or more creatures", effects.Creature()).WithCount(1, 0)
	fromX := effects.SacrificeSpec("X creatures", effects.Creature())
	fromX.CountFromX = true

	cases := []struct {
		name     string
		spec     *game.TargetSpec
		x        int
		lo, hi   int
		legal    []int
		notLegal []int
	}{
		{"fixed", fixed, 0, 2, 2, []int{2}, []int{0, 1, 3}},
		{"open", open, 0, 1, 0, []int{1, 2, 7}, []int{0}},
		{"from X at 3", fromX, 3, 3, 3, []int{3}, []int{0, 2, 4}},
		{"from X at 0", fromX, 0, 0, 0, []int{0}, []int{1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lo, hi := game.SacrificeCostBounds(tc.spec, tc.x)
			if lo != tc.lo || hi != tc.hi {
				t.Fatalf("bounds = %d/%d, want %d/%d", lo, hi, tc.lo, tc.hi)
			}
			for _, n := range tc.legal {
				if !game.SacrificeCountLegal(tc.spec, tc.x, n) {
					t.Errorf("naming %d is refused, want accepted", n)
				}
			}
			for _, n := range tc.notLegal {
				if game.SacrificeCountLegal(tc.spec, tc.x, n) {
					t.Errorf("naming %d is accepted, want refused", n)
				}
			}
		})
	}
}

// Every clause the catalog ships is a fixed count; the guard above
// would have panicked at init otherwise. This pins the other half:
// the cards #747 enabled really declare the count they print.
func TestCatalogSacrificeCountsMatchThePrintedCards(t *testing.T) {
	want := map[string]int{
		"Savvy Hunter":             2,
		"Samwise Gamgee":           3,
		"Sai, Master Thopterist":   2,
		"Magda, the Hoardmaster":   3,
		"Magda, Brazen Outlaw":     5,
		"Priest of Forgotten Gods": 2,
		"Kuldotha Forgemaster":     3,
		"Tamiyo's Journal":         3,
		"Teysa, Orzhov Scion":      3,
		"Hedron Detonator":         2,
		"Olivia, Opulent Outlaw":   2,
		"Breya, Etherium Shaper":   2,
		"Whisper, Blood Liturgist": 2,
	}
	seen := map[string]bool{}
	for _, spec := range effects.All() {
		n, ok := want[spec.Name]
		if !ok {
			continue
		}
		seen[spec.Name] = true
		found := false
		for _, ab := range spec.Activated {
			if ab.Cost.SacrificeOther != nil {
				found = true
				if got := game.SacrificeCostCount(ab.Cost.SacrificeOther); got != n {
					t.Errorf("%s: %q sacrifices %d, want %d", spec.Name, ab.Label, got, n)
				}
			}
		}
		if !found {
			t.Errorf("%s declares no sacrifice-N ability", spec.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("%s is not registered", name)
		}
	}
}
