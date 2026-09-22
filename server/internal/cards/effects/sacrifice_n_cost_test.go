package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_n_cost_test.go — #747 (ADR 0020 addendum §12–§15, ADR
// 0021 addendum): a sacrifice cost of N permanents, at all three cost
// sites — an activated ability, a mana ability, and an additional cost
// to cast — through fixture specs, so the engine rules are pinned
// apart from any one card. The cards that use it are in
// sacrifice_n_cards_test.go.

const (
	sacNAbilityOracle  = "test-sacrifice-n-ability"
	sacNSelfTooOracle  = "test-sacrifice-n-self-too"
	sacNManaOracle     = "test-sacrifice-n-mana"
	sacNSpellOracle    = "test-sacrifice-n-spell"
	sacNWatcherOracle  = "test-sacrifice-n-dies-watcher"
	sacSelfOtherOracle = "test-sacrifice-self-and-one-other"
)

func init() {
	// "Sacrifice two creatures: You gain 1 life."
	Register(Spec{
		OracleID: sacNAbilityOracle,
		Name:     "Two-Creature Outlet",
		Activated: []ActivatedAbility{{
			Label:  "Sacrifice two creatures: You gain 1 life.",
			Cost:   SacrificeN(2, "two creatures", Creature()),
			Effect: Do(GainLife{Amount: 1}),
		}},
	})
	// "{T}, Sacrifice this and two other creatures: You gain 1 life."
	// The source is itself a creature, so naming it among the two is
	// the "source named twice" case.
	Register(Spec{
		OracleID: sacNSelfTooOracle,
		Name:     "Self And Two Outlet",
		Activated: []ActivatedAbility{{
			Label:  "Sacrifice this and two creatures: You gain 1 life.",
			Cost:   Plus(SacrificeThis(), SacrificeN(2, "two creatures", Creature())),
			Effect: Do(GainLife{Amount: 1}),
		}},
	})
	// "Sacrifice two creatures: Add {C}{C}{C}."
	Register(Spec{
		OracleID: sacNManaOracle,
		Name:     "Two-Creature Altar",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{SacrificeOther: SacrificeN(2, "two creatures", Creature()).SacrificeOther},
			Produced: "{C}{C}{C}",
			Label:    "Sacrifice two creatures: Add {C}{C}{C}",
		}},
	})
	// "As an additional cost to cast this spell, sacrifice two
	// creatures. Draw two cards."
	Register(Spec{
		OracleID:       sacNSpellOracle,
		Name:           "Two-Creature Rites",
		AdditionalCost: SacrificeNCost(2, "two creatures", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{N: 2}.Apply(ctx)
		},
	})
	// "Whenever this or another creature dies, you gain 1 life." An
	// untargeted Blood Artist, so a test can count the triggers by
	// life total without answering prompts.
	Register(Spec{
		OracleID: sacNWatcherOracle,
		Name:     "Dies Watcher",
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			}, "Dies Watcher — gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
	// "Whenever this or another creature dies, you gain 1 life.
	// Sacrifice this and a creature: no effect." The pre-#747
	// two-permanent shape — SacrificeSelf plus a one-permanent
	// SacrificeOther — on a card that is its own dies-watcher.
	Register(Spec{
		OracleID: sacSelfOtherOracle,
		Name:     "Self And One Outlet",
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedCreature(ev, g)
				return ok
			}, "Self And One Outlet — gain 1 life", Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:  "Sacrifice this and a creature: no effect.",
			Cost:   Plus(SacrificeThis(), SacrificeACreature()),
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}

func sacNCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return pushCatalogPermanent(g, owner, name, "Creature — Bear", "", false)
}

func countEventsOf(g *game.Game, kind game.EventKind, ids ...uuid.UUID) int {
	want := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == kind && want[ev.CardID] {
			n++
		}
	}
	return n
}

// The count lives on the clause, and Plus carries the clause whole —
// a merge that dropped the count would price "Sacrifice two" at one.
func TestSacrificeNConstructorsCarryTheCount(t *testing.T) {
	cost := Plus(TapCost(), SacrificeN(3, "three artifacts", Artifact()), ManaCost("{1}"))
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
	add := SacrificeNCost(2, "two creatures", Creature())
	if game.SacrificeCostCount(add.Sacrifice) != 2 || add.Label != "Sacrifice two creatures" {
		t.Errorf("SacrificeNCost = %+v", add)
	}
	if got := game.SacrificeCostCount(SacrificeACreature().SacrificeOther); got != 1 {
		t.Errorf("the one-permanent constructors still count 1, got %d", got)
	}
}

func TestSacrificeNActivatedAbilityPaysExactlyN(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	outlet := pushCatalogPermanent(g, me.ID, "Two-Creature Outlet", "Enchantment", sacNAbilityOracle, false)
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
	if got := countEventsOf(g, game.EventSacrifice, a, b); got != 2 {
		t.Errorf("%d EventSacrifice, want one per permanent", got)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want the ability to resolve once", life, me.Life)
	}
}

// Every malformed payment refuses the whole activation with nothing
// paid: no creature moved, the source not sacrificed, nothing on the
// stack.
func TestSacrificeNRefusedPaymentsPayNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	outlet := pushCatalogPermanent(g, me.ID, "Two-Creature Outlet", "Enchantment", sacNAbilityOracle, false)
	selfToo := pushCatalogPermanent(g, me.ID, "Self And Two Outlet", "Creature — Horror", sacNSelfTooOracle, false)
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")
	c := sacNCreature(g, me.ID, "Bear C")
	theirs := sacNCreature(g, opp.ID, "Their Bear")
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)

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
					t.Errorf("%s left the battlefield on a refused payment", cardNameOf(g, id))
				}
			}
			if len(g.StackMeta) != stack {
				t.Error("a refused activation put something on the stack")
			}
		})
	}
	if countEventsOf(g, game.EventSacrifice, outlet, selfToo, a, b, c, theirs, rock) != 0 {
		t.Error("a refused payment emitted EventSacrifice")
	}
	// The same board pays the self-too cost legally: the source plus
	// two others, three sacrifices.
	if err := g.ActivateCatalogAbility(me.ID, selfToo, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("legal self-too payment: %v", err)
	}
	if got := countEventsOf(g, game.EventSacrifice, selfToo, a, b); got != 3 {
		t.Errorf("%d EventSacrifice, want 3 (the source and both creatures)", got)
	}
}

// The refusal leaves the tap state and the mana pool alone too: a
// {T} plus sacrifice-three cost (Kuldotha Forgemaster) and a mana plus
// sacrifice-two cost (Sai) named with too few permanents neither tap
// the source nor spend the floating mana.
func TestSacrificeNRefusalLeavesTapAndPoolAlone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forge := pushCatalogPermanent(g, me.ID, "Kuldotha Forgemaster", "Artifact Creature — Construct", kuldothaForgemasterOracle, false)
	sai := pushCatalogPermanent(g, me.ID, "Sai, Master Thopterist", "Legendary Creature — Human Artificer", b07SaiMasterThopteristOracle, false)
	arts := pushTokens(g, me.ID, TreasureToken(), 2)
	g.WithWriteLock(func() {
		_ = g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}")
	})
	if err := g.ActivateCatalogAbility(me.ID, forge, 0, game.ActivateAbilityParams{SacrificeIDs: arts}); err != game.ErrInvalidParam {
		t.Fatalf("two artifacts for three: got %v, want ErrInvalidParam", err)
	}
	if b16Tapped(t, g, forge) {
		t.Error("a refused payment tapped the source")
	}
	if err := g.ActivateCatalogAbility(me.ID, sai, 0, game.ActivateAbilityParams{
		SacrificeIDs: arts[:1], Strict: true,
	}); err != game.ErrInvalidParam {
		t.Fatalf("one artifact for two: got %v, want ErrInvalidParam", err)
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %d after a refused payment, want the 2 floating mana untouched", len(me.ManaPool))
	}
	for _, id := range arts {
		if !g.Battlefield.Contains(id) {
			t.Error("a refused payment sacrificed an artifact")
		}
	}
}

func cardNameOf(g *game.Game, id uuid.UUID) string {
	if c, ok := g.LookupCardForEffect(id); ok {
		return c.Name
	}
	return id.String()
}

func TestSacrificeNManaAbilityPaysExactlyN(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	altar := pushCatalogPermanent(g, me.ID, "Two-Creature Altar", "Artifact", sacNManaOracle, false)
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")

	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{a}}); err != game.ErrInvalidParam {
		t.Fatalf("one creature for a two-creature mana cost: got %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{b, b}}); err != game.ErrInvalidParam {
		t.Fatalf("the same creature twice: got %v, want ErrInvalidParam", err)
	}
	if len(me.ManaPool) != 0 || !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Fatal("a refused mana payment produced mana or ate a creature")
	}
	if err := g.ActivateManaAbility(me.ID, altar, 0, game.ManaAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if len(me.ManaPool) != 3 {
		t.Errorf("pool = %d, want 3", len(me.ManaPool))
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both creatures are sacrificed")
	}
	if got := countEventsOf(g, game.EventSacrifice, a, b); got != 2 {
		t.Errorf("%d EventSacrifice, want 2", got)
	}
}

func TestSacrificeNAdditionalCostPaysExactlyN(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := sacNCreature(g, me.ID, "Bear A")
	b := sacNCreature(g, me.ID, "Bear B")
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}

	for _, bad := range [][]uuid.UUID{{a}, {a, a}, {a, b, uuid.New()}} {
		id, err := tryCastWithSacrifice(g, "Two-Creature Rites", "Sorcery", sacNSpellOracle, bad)
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
	spell, err := tryCastWithSacrifice(g, "Two-Creature Rites", "Sorcery", sacNSpellOracle, []uuid.UUID{b, a})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell is not on the stack")
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both creatures are paid while the spell is on the stack")
	}
	if got := countEventsOf(g, game.EventSacrifice, a, b); got != 2 {
		t.Errorf("%d EventSacrifice, want 2", got)
	}
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			outlet := pushCatalogPermanent(g, me.ID, "Two-Creature Outlet", "Enchantment", sacNAbilityOracle, false)
			watcher := pushCatalogPermanent(g, me.ID, "Dies Watcher", "Creature — Vampire", sacNWatcherOracle, false)
			goblin := sacNCreature(g, me.ID, "Goblin")
			ids := []uuid.UUID{goblin, watcher}
			if watcherFirst {
				ids = []uuid.UUID{watcher, goblin}
			}
			life := me.Life
			if err := g.ActivateCatalogAbility(me.ID, outlet, 0, game.ActivateAbilityParams{SacrificeIDs: ids}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watcher := pushCatalogPermanent(g, me.ID, "Dies Watcher", "Creature — Vampire", sacNWatcherOracle, false)
	goblin := sacNCreature(g, me.ID, "Goblin")
	for i := 0; i < 4; i++ {
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	life := me.Life
	if _, err := tryCastWithSacrifice(g, "Two-Creature Rites", "Sorcery", sacNSpellOracle, []uuid.UUID{watcher, goblin}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: the watcher sees both deaths", got)
	}
}

// The pre-#747 two-permanent case, SacrificeSelf plus a one-permanent
// SacrificeOther, is paid as one batch too. The source is the watcher
// here, and the source leads the payment, so it is sacrificed first:
// only the batch lets it see the other creature die after it.
func TestSacrificeSelfAndOtherIsOneSimultaneousExit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	outlet := pushCatalogPermanent(g, me.ID, "Self And One Outlet", "Creature — Horror", sacSelfOtherOracle, false)
	goblin := sacNCreature(g, me.ID, "Goblin")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, outlet, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{goblin}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 2 {
		t.Errorf("gained %d life, want 2: the source sees itself and the Goblin die", got)
	}
}

// ADR 0020 addendum §12: a sacrifice clause is a fixed count of at
// least one. Every other shape panics at boot, at all three sites, so
// no card can declare a variable count and have it read as something
// else.
func TestRegisterRejectsVariableSacrificeClauses(t *testing.T) {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	ability := func(id string, spec *game.TargetSpec) Spec {
		return Spec{OracleID: id, Name: id, Activated: []ActivatedAbility{{
			Label: "x", Cost: game.AbilityCost{SacrificeOther: spec}, Effect: noop,
		}}}
	}
	cases := []struct {
		name string
		spec Spec
		want string
	}{
		{"min differs from max", ability("sac-guard-range", sacrificeSpec("one or two", Creature()).WithCount(1, 2)), "fixed count"},
		{"one or more", ability("sac-guard-any", sacrificeSpec("one or more", Creature()).WithCount(1, 0)), "fixed count"},
		{"count zero", ability("sac-guard-zero", sacrificeSpec("none", Creature()).WithCount(0, 0)), "fixed count"},
		{"count from X", ability("sac-guard-x", func() *game.TargetSpec {
			s := sacrificeSpec("X creatures", Creature())
			s.CountFromX = true
			return s
		}()), "from X"},
		{"allow same", ability("sac-guard-same", func() *game.TargetSpec {
			s := sacrificeSpec("two creatures", Creature()).WithCount(2, 2)
			s.AllowSame = true
			return s
		}()), "AllowSame"},
		{"players", ability("sac-guard-players", func() *game.TargetSpec {
			s := sacrificeSpec("a creature", Creature())
			s.Players = true
			return s
		}()), "admits players"},
		{"mana ability site", Spec{OracleID: "sac-guard-mana", Name: "sac-guard-mana", ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{SacrificeOther: sacrificeSpec("one or more", Creature()).WithCount(1, 0)},
			Produced: "{C}",
		}}}, "mana ability 0"},
		{"additional cost site", Spec{OracleID: "sac-guard-spell", Name: "sac-guard-spell", AdditionalCost: &game.AdditionalCost{
			Sacrifice: sacrificeSpec("any number of creatures", Creature()).WithCount(0, 0),
		}}, "additional cost"},
		// #1224: an AdditionalCost in OptionalCosts carries the same
		// Sacrifice clause the mandatory slot does (Constant Mists'
		// "Buyback—Sacrifice a land"), so the same guard has to catch a
		// variable count there too.
		{"optional cost site", Spec{OracleID: "sac-guard-optional", Name: "sac-guard-optional", OptionalCosts: []game.AdditionalCost{{
			Optional:  true,
			Key:       game.BuybackKey,
			Label:     "Buyback—Sacrifice any number of creatures",
			Sacrifice: sacrificeSpec("any number of creatures", Creature()).WithCount(0, 0),
		}}}, "optional cost"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() { Register(tc.spec) })
			if Has(tc.spec.OracleID) {
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
	spec := Spec{
		OracleID: "sac-guard-optional-good",
		Name:     "sac-guard-optional-good",
		OptionalCosts: []game.AdditionalCost{{
			Optional:  true,
			Key:       game.BuybackKey,
			Label:     "Buyback—Sacrifice a land",
			Sacrifice: sacrificeSpec("a land", Land()),
		}},
	}
	Register(spec)
	if !Has(spec.OracleID) {
		t.Error("a legal optional sacrifice clause was refused")
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
	for _, spec := range All() {
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
