package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// autotap_topup_test.go — ADR 0118 §1, the pool top-up. The
// auto-tapper plans only what the floating pool is missing, so mana
// already in the pool is spent first and the plan pays the rest.
// The legal-move half (the enumerator and the bot agree) is in
// internal/legal/pool_top_up_test.go.

// zigguratToken is Ancient Ziggurat's mana: "Spend this mana only to
// cast a creature spell."
func zigguratToken(color string) ManaToken {
	return ManaToken{Color: color, Restrictions: []string{ManaRestrictCast, ManaRestrictType("Creature")}}
}

// countTappedForTest counts p's tapped permanents.
func countTappedForTest(g *Game, p *Player) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p.ID && c.Tapped {
			n++
		}
	}
	return n
}

// The headline case: {G} floating plus one untapped Forest casts a
// {1}{G} creature. Before the top-up the pool alone was short and the
// Forest alone was short, and the cast was refused.
func TestAutoTapTopsUpAFloatingGreen(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	id := pushTypedCardToHandWithCost(p, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	p.ManaPool.AddMana(ManaToken{Color: "G"})

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {1}{G} off a floating {G} and a Forest: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Error("Grizzly Bears did not reach the stack")
	}
	if !cardTappedForTest(g, forest) {
		t.Error("the Forest was not tapped for the {1} the pool was missing")
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool after the cast: %+v, want empty — the floating {G} is spent first", p.ManaPool)
	}
}

// A pool that covers the cost on its own taps nothing: the shortcut
// the auto-tapper always had.
func TestAutoTapFundedPoolTapsNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	id := pushTypedCardToHandWithCost(p, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	p.ManaPool.AddMana(ManaToken{Color: "G"}, ManaToken{Color: "C"})

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped cast on a funded pool: %v", err)
	}
	if cardTappedForTest(g, forest) {
		t.Error("the Forest was tapped for a cost the pool already covered")
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool after the cast: %+v, want empty", p.ManaPool)
	}
}

// Floating mana the cost cannot use stays in the pool, and the plan
// pays everything it cannot: {U} floating does not pay {G}{G}.
func TestAutoTapTopUpLeavesUnusableFloatingMana(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	id := pushTypedCardToHandWithCost(p, "Twin Bears", "Creature — Bear", "{G}{G}")
	p.ManaPool.AddMana(ManaToken{Color: "U"})

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {G}{G} with a floating {U}: %v", err)
	}
	if n := countTappedForTest(g, p); n != 2 {
		t.Errorf("tapped %d Forests, want 2", n)
	}
	if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "U" {
		t.Errorf("pool after the cast: %+v, want the floating {U} left alone", p.ManaPool)
	}
}

// Floating mana pays the generic half too: {U} floating plus one
// Forest pays {1}{G}, and nothing is left over.
func TestAutoTapTopUpSpendsFloatingManaOnGeneric(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	id := pushTypedCardToHandWithCost(p, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	p.ManaPool.AddMana(ManaToken{Color: "U"})

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {1}{G} with a floating {U}: %v", err)
	}
	if n := countTappedForTest(g, p); n != 1 {
		t.Errorf("tapped %d Forests, want 1 — the {U} pays the {1}", n)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool after the cast: %+v, want empty", p.ManaPool)
	}
}

// Restricted pool mana is credited only where it may pay. Ancient
// Ziggurat's {G} funds a creature spell's {G}, and the Forest pays the
// {1}.
func TestAutoTapTopUpCreditsRestrictedManaWhereItMayPay(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	id := pushTypedCardToHandWithCost(p, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	p.ManaPool.AddMana(zigguratToken("G"))

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped creature with Ziggurat mana floating: %v", err)
	}
	if !cardTappedForTest(g, forest) {
		t.Error("the Forest was not tapped")
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool after the cast: %+v, want the Ziggurat {G} spent", p.ManaPool)
	}
}

// …and nowhere else. An instant cannot spend Ziggurat mana, so one
// Forest cannot top it up to {1}{G}: the cast is refused with nothing
// tapped, and the restricted mana stays in the pool. With a second
// Forest the plan pays the whole cost and the Ziggurat {G} still floats.
func TestAutoTapTopUpDoesNotCreditRestrictedManaElsewhere(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	forest := pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	id := pushTypedCardToHandWithCost(p, "Giant Growth Plus", "Instant", "{1}{G}")
	p.ManaPool.AddMana(zigguratToken("G"))

	err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("instant off Ziggurat mana and one Forest: got %v, want *InsufficientManaError", err)
	}
	if cardTappedForTest(g, forest) {
		t.Error("a refused cast tapped the Forest")
	}
	if len(p.ManaPool) != 1 {
		t.Errorf("pool after the refusal: %+v, want the Ziggurat {G} untouched", p.ManaPool)
	}

	pushBattlefieldForTest(g, p.ID, "Forest", "Basic Land — Forest", "")
	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("instant off two Forests with Ziggurat mana floating: %v", err)
	}
	if n := countTappedForTest(g, p); n != 2 {
		t.Errorf("tapped %d Forests, want 2", n)
	}
	if len(p.ManaPool) != 1 || len(p.ManaPool[0].Restrictions) == 0 {
		t.Errorf("pool after the cast: %+v, want the Ziggurat {G} still floating", p.ManaPool)
	}
}

// A hybrid symbol and a single-colour one compete for one floating
// {G}. The {G} must go to {G}: the Plains can pay {G/W} and cannot pay
// {G}. The pool then holds the floating {G} in front of the Plains'
// {W}, which the payment has to assign the right way round.
func TestAutoTapTopUpGivesTheFloatingManaToTheSymbolOnlyItCanPay(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	plains := pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")
	id := pushTypedCardToHandWithCost(p, "Hybrid Bear", "Creature — Bear", "{G/W}{G}")
	p.ManaPool.AddMana(ManaToken{Color: "G"})

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {G/W}{G} off a floating {G} and a Plains: %v (pool %+v)", err, p.ManaPool)
	}
	if !cardTappedForTest(g, plains) {
		t.Error("the Plains was not tapped for the {G/W}")
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("pool after the cast: %+v, want empty", p.ManaPool)
	}
}

// Two hybrids, one floating {G}: which one the {G} pays depends on the
// board. With an Island it must pay {G/W} (the Island pays {G/U}); with
// a Plains it must pay {G/U}.
func TestAutoTapTopUpChoosesTheHybridTheBoardCannotPay(t *testing.T) {
	for _, land := range []struct{ name, sub string }{{"Island", "Island"}, {"Plains", "Plains"}} {
		t.Run(land.name, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			p := g.Seats[0]
			src := pushBattlefieldForTest(g, p.ID, land.name, "Basic Land — "+land.sub, "")
			id := pushTypedCardToHandWithCost(p, "Two Guilds", "Creature — Elf", "{G/U}{G/W}")
			p.ManaPool.AddMana(ManaToken{Color: "G"})

			if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
				t.Fatalf("auto-tapped {G/U}{G/W} off a floating {G} and an %s: %v", land.name, err)
			}
			if !cardTappedForTest(g, src) {
				t.Errorf("the %s was not tapped", land.name)
			}
			if len(p.ManaPool) != 0 {
				t.Errorf("pool after the cast: %+v, want empty", p.ManaPool)
			}
		})
	}
}

// A board that cannot pay what the pool is missing is refused with
// nothing tapped and the pool untouched: the top-up is atomic.
func TestAutoTapTopUpRefusalTapsNothing(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	mountain := pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	id := pushTypedCardToHandWithCost(p, "Twin Bears", "Creature — Bear", "{G}{G}")
	p.ManaPool.AddMana(ManaToken{Color: "G"})

	err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("{G}{G} off a floating {G} and a Mountain: got %v, want *InsufficientManaError", err)
	}
	if cardTappedForTest(g, mountain) {
		t.Error("a refused cast tapped the Mountain")
	}
	if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "G" {
		t.Errorf("pool after the refusal: %+v, want the floating {G}", p.ManaPool)
	}
}

// An activation tops up too (payAbilityManaCostLocked, which a special
// action and an attack tax also pay through).
func TestAutoTapTopUpForAnActivation(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G}"))
	forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	me.ManaPool.AddMana(ManaToken{Color: "G"})

	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {1}{G} activation off a floating {G} and a Forest: %v", err)
	}
	if !cardTappedForTest(g, forest) {
		t.Error("the Forest was not tapped")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the activation: %+v, want empty", me.ManaPool)
	}
}

// A special action tops up: foretell's {2} off a floating {C} and one
// Forest.
func TestAutoTapTopUpForASpecialAction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	withForetellCard(t, "{1}{U}")
	advanceTo(t, g, StepPrecombatMain)
	card := seedHandCard(me, "Saw It Coming", foretellOracle, "Instant", "{1}{U}{U}")
	forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.PerformSpecialAction(me.ID, card.InstanceID, SpecialActionForetell, SpecialActionParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped foretell off a floating {C} and a Forest: %v", err)
	}
	if !cardTappedForTest(g, forest) {
		t.Error("the Forest was not tapped")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the special action: %+v, want empty", me.ManaPool)
	}
}

// A pay-unless prompt (payCostLocked) and the attack-tax check top up
// too.
func TestAutoTapTopUpForPayUnlessAndAttackTax(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")

	me.ManaPool.AddMana(ManaToken{Color: "C"})
	if err := g.attackTaxAffordableLocked(me.ID, AttackTaxPrice{Cost: "{2}"}, DeclareAttackersParams{AutoTap: true}); err != nil {
		t.Errorf("attack tax {2} off a floating {C} and a Forest: %v", err)
	}

	if !g.payCostLocked(me, costFor(t, "{2}"), uuid.Nil, nil) {
		t.Fatal("pay-unless {2} off a floating {C} and a Forest was refused")
	}
	if n := countTappedForTest(g, me); n != 1 {
		t.Errorf("tapped %d Forests, want 1", n)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the payment: %+v, want empty", me.ManaPool)
	}
}

// poolShortfalls is the unit: what a plan must pay once the pool has
// paid what it can.
func TestPoolShortfalls(t *testing.T) {
	cases := []struct {
		name string
		pool ManaPool
		cost string
		x    int
		ctx  ManaSpendContext
		want []string
	}{
		{"covered", ManaPool{{Color: "G"}, {Color: "G"}}, "{1}{G}", 0, ManaSpendContext{}, nil},
		{"colored first", ManaPool{{Color: "G"}}, "{1}{G}", 0, ManaSpendContext{}, []string{"{1}"}},
		{"generic from an off colour", ManaPool{{Color: "U"}}, "{1}{G}", 0, ManaSpendContext{}, []string{"{G}"}},
		{"unusable", ManaPool{{Color: "U"}}, "{G}{G}", 0, ManaSpendContext{}, []string{"{G}{G}"}},
		{"X is folded into generic", ManaPool{{Color: "R"}}, "{X}{R}", 3, ManaSpendContext{}, []string{"{3}"}},
		{"single colour before hybrid", ManaPool{{Color: "G"}}, "{G/W}{G}", 0, ManaSpendContext{}, []string{"{G/W}"}},
		{"two hybrids, both ways", ManaPool{{Color: "G"}}, "{G/U}{G/W}", 0, ManaSpendContext{}, []string{"{G/W}", "{G/U}"}},
		{"restricted mana unseen", ManaPool{zigguratToken("G")}, "{1}{G}", 0, instantSpendContext(), []string{"{1}{G}"}},
		{"restricted mana credited", ManaPool{zigguratToken("G")}, "{1}{G}", 0, creatureSpendContext(), []string{"{1}"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := poolShortfalls(tc.pool, costFor(t, tc.cost), tc.x, tc.ctx)
			if len(got) != len(tc.want) {
				t.Fatalf("poolShortfalls(%v, %s) = %v, want %v", tc.pool, tc.cost, got, tc.want)
			}
			for i := range got {
				if got[i].String() != tc.want[i] {
					t.Errorf("shortfall %d: got %s, want %s", i, got[i].String(), tc.want[i])
				}
			}
		})
	}
}

// The pool solver pays a cost whenever SOME assignment of tokens to
// symbols could. The first-match walk takes the {G} for {G/W} and then
// finds no {G} for the second symbol; the repair moves the hybrid onto
// the {W}. This is the pool the top-up hands the payment when a {G}
// floated before a Plains was tapped.
func TestPoolSolverReassignsAHybridToPayACost(t *testing.T) {
	pool := ManaPool{{Color: "G"}, {Color: "W"}}
	cost := costFor(t, "{G/W}{G}")
	if !pool.CanPay(cost, 0) {
		t.Fatalf("[{G} {W}] cannot pay {G/W}{G}; missing %v", pool.Missing(cost, 0))
	}
	if m := pool.Missing(cost, 0); m != nil {
		t.Errorf("Missing = %v, want nil", m)
	}
	spent, ok := pool.SpendManaFor(cost, 0, ManaSpendContext{})
	if !ok || len(spent) != 2 {
		t.Fatalf("SpendManaFor = %v, %v; want both tokens spent", spent, ok)
	}
	// A cost no assignment pays is still refused, and named.
	if m := (ManaPool{{Color: "G"}, {Color: "U"}}).Missing(cost, 0); len(m) != 1 || m[0] != "{G}" {
		t.Errorf("Missing over [{G} {U}] = %v, want [{G}]", m)
	}
}
