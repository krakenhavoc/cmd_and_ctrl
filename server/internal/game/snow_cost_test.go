package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// snow_cost_test.go — #5986, CR 107.4h: "When used in a cost, the snow
// mana symbol {S} represents a cost that can be paid with one mana of
// any type produced by a snow source. Effects that reduce the amount of
// generic mana you pay don't affect {S} costs."
//
// The pool solver, the auto-tapper, the pool top-up, the any-type
// widening and the cost reducers each get a case, and the headline
// cases run through the real cast and activation paths with a
// Snow-Covered Forest beside a nonsnow Forest and a Sol Ring.

// snowMana is one mana a snow land made: the snapshot manaSourceKindsOf
// takes of a Snow-Covered Forest.
func snowMana(color string) ManaToken {
	return ManaToken{Color: color, SourceKinds: ManaSourceSnow | ManaSourceLand | ManaSourceBasicLand}
}

// snowForestForTest puts a Snow-Covered Forest under `p`'s control.
func snowForestForTest(g *Game, p *Player) uuid.UUID {
	return pushBattlefieldForTest(g, p.ID, "Snow-Covered Forest", "Basic Snow Land — Forest", "")
}

// The pool solver: an {S} takes mana of any colour from a snow source,
// and no mana from anything else, colourless included.
func TestPoolPaysSnowOnlyWithSnowMana(t *testing.T) {
	for _, tc := range []struct {
		name string
		pool ManaPool
		cost string
		ok   bool
	}{
		{"a snow green pays {S}", ManaPool{snowMana("G")}, "{S}", true},
		{"a snow colourless pays {S}", ManaPool{snowMana("C")}, "{S}", true},
		{"a nonsnow colourless does not", ManaPool{{Color: "C"}}, "{S}", false},
		{"a nonsnow green does not", ManaPool{{Color: "G"}}, "{S}", false},
		{"one snow mana is one {S}", ManaPool{snowMana("G")}, "{S}{S}", false},
		{"snow mana still pays generic", ManaPool{snowMana("U"), snowMana("U")}, "{1}{S}", true},
		{"the {G} takes the nonsnow Forest", ManaPool{snowMana("G"), {Color: "G"}}, "{S}{G}", true},
		{"whichever order the pool holds them", ManaPool{snowMana("G"), {Color: "G"}}, "{G}{S}", true},
		{"a {C} beside the {S} is still colourless", ManaPool{snowMana("G"), {Color: "C"}}, "{C}{S}", true},
		{"and the snow mana cannot be both", ManaPool{snowMana("C")}, "{C}{S}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.pool.CanPay(costFor(t, tc.cost), 0); got != tc.ok {
				t.Errorf("CanPay(%s) out of %+v = %v, want %v", tc.cost, tc.pool, got, tc.ok)
			}
		})
	}
}

// The spent record says the snow mana paid, and the breakdown a player
// sees names the symbol as printed.
func TestPoolSpendsTheSnowManaOnTheSnowSymbol(t *testing.T) {
	pool := ManaPool{{Color: "G"}, snowMana("G")}
	spent, ok := pool.SpendManaFor(costFor(t, "{S}"), 0, ManaSpendContext{})
	if !ok {
		t.Fatal("{S} out of a Forest's {G} and a Snow-Covered Forest's {G}: refused")
	}
	if len(spent) != 1 || !spent[0].SourceKinds.Has(ManaSourceSnow) {
		t.Errorf("spent %+v, want the snow {G}", spent)
	}
	if len(pool) != 1 || pool[0].SourceKinds.Has(ManaSourceSnow) {
		t.Errorf("left %+v, want the nonsnow {G}", pool)
	}

	missing := ManaPool{{Color: "C"}}.Missing(costFor(t, "{S}"), 0)
	if !slices.Equal(missing, []string{"{S}"}) {
		t.Errorf("Missing = %v, want [{S}]", missing)
	}
}

// Tapping a snow land stamps ManaSourceSnow on its mana (#1212's
// snapshot), and that mana pays an activation's {S} through the real
// payment. A nonsnow Forest's does not.
func TestSnowLandManaPaysAnActivationsSnowSymbol(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	pod := phyrexianPodSource(g, me, abilityCostMana("{S}"))
	forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	snow := snowForestForTest(g, me)

	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].SourceKinds.Has(ManaSourceSnow) {
		t.Fatalf("the Forest's mana: %+v, want one nonsnow {G}", me.ManaPool)
	}
	if err := g.ActivateCatalogAbility(me.ID, pod, 0, ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("a strict {S} activation paid with a nonsnow Forest's {G}: accepted")
	}

	if err := g.ActivateManaAbility(me.ID, snow, 0, ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Snow-Covered Forest: %v", err)
	}
	var snowTok *ManaToken
	for i := range me.ManaPool {
		if me.ManaPool[i].Source == snow {
			snowTok = &me.ManaPool[i]
		}
	}
	if snowTok == nil || !snowTok.SourceKinds.Has(ManaSourceSnow) {
		t.Fatalf("the Snow-Covered Forest's mana carries no snow bit: %+v", me.ManaPool)
	}
	if err := g.ActivateCatalogAbility(me.ID, pod, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("a strict {S} activation off the Snow-Covered Forest's {G}: %v", err)
	}
	paid := lastAbilityPaid(t, g)
	if len(paid.Mana) != 1 || paid.Mana[0].Source != snow {
		t.Errorf("paid %+v, want the Snow-Covered Forest's mana", paid.Mana)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Source != forest {
		t.Errorf("pool after: %+v, want the Forest's {G} left floating", me.ManaPool)
	}
}

// The auto-tapper books a snow source for the {S}, wherever it sits in
// the source order, and finds no plan without one: a Sol Ring's {C}
// and a Forest's {G} pay nothing of an {S}.
func TestAutoTapChoosesASnowSourceForTheSnowSymbol(t *testing.T) {
	withCatalogHook(t, solRingHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	forest := pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	pushBattlefieldForTest(g, me.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487")

	if _, ok := g.AutoTapForCost(me.ID, costFor(t, "{S}"), 0); ok {
		t.Fatal("planned an {S} with only a Forest and a Sol Ring")
	}
	id := pushTypedCardToHandWithCost(me, "Snow Test Bear", "Snow Creature — Bear", "{1}{S}{G}")
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("auto-tapped {1}{S}{G} with no snow source: accepted")
	}
	if countTappedForTest(g, me) != 0 {
		t.Fatal("a refused auto-tap tapped something")
	}

	snow := snowForestForTest(g, me)
	plan, ok := g.AutoTapForCost(me.ID, costFor(t, "{S}"), 0)
	if !ok || len(plan) != 1 || plan[0] != snow {
		t.Fatalf("plan for {S} = %v (%v), want the Snow-Covered Forest %v", plan, ok, snow)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped {1}{S}{G} off a Snow-Covered Forest, a Forest and a Sol Ring: %v", err)
	}
	if !g.Stack.Contains(id) {
		t.Fatal("the bear did not reach the stack")
	}
	if !cardTappedForTest(g, snow) || !cardTappedForTest(g, forest) {
		t.Error("the plan did not tap the Snow-Covered Forest for {S} and the Forest for {G}")
	}
}

// The pool top-up credits floating snow mana to an {S} and floating
// nonsnow mana to nothing of it.
func TestPoolTopUpCreditsOnlySnowManaToTheSnowSymbol(t *testing.T) {
	cost := costFor(t, "{1}{S}")
	short := poolShortfalls(ManaPool{snowMana("G")}, cost, 0, ManaSpendContext{})
	if len(short) != 1 || short[0].String() != "{1}" {
		t.Errorf("shortfall after a floating snow {G} = %v, want {1}", short)
	}
	short = poolShortfalls(ManaPool{{Color: "C"}}, cost, 0, ManaSpendContext{})
	if len(short) != 1 || short[0].String() != "{S}" {
		t.Errorf("shortfall after a floating Sol Ring {C} = %v, want {S}", short)
	}
	// A coloured symbol takes the nonsnow mana of its colour first, so
	// the snow mana is left for the {S}.
	short = poolShortfalls(ManaPool{snowMana("G"), {Color: "G"}}, costFor(t, "{G}{S}{1}"), 0, ManaSpendContext{})
	if len(short) != 1 || short[0].String() != "{1}" {
		t.Errorf("shortfall after a snow {G} and a {G} = %v, want {1}", short)
	}
}

// CR 107.4h: a generic reduction never reaches an {S}, and a reduction
// by {C} is not one by {S} either. Snow is neither a colour nor a type,
// so neither an any-colour nor an any-type grant widens it.
func TestCostReductionsAndSpendGrantsLeaveTheSnowSymbol(t *testing.T) {
	if got := reduceGeneric(costFor(t, "{1}{S}"), 3).String(); got != "{S}" {
		t.Errorf("{1}{S} less {3} = %s, want {S}", got)
	}
	if got := reduceByManaCost(costFor(t, "{2}{S}"), costFor(t, "{C}")).String(); got != "{1}{S}" {
		t.Errorf("{2}{S} less {C} = %s, want {1}{S} (CR 118.7b)", got)
	}
	for name, widened := range map[string]ParsedCost{
		"any colour":             widenForAnyColorSpend(costFor(t, "{S}{G}"), false),
		"any type":               widenForAnyColorSpend(costFor(t, "{S}{G}"), true),
		"as any type permission": asAnyTypeCost(costFor(t, "{S}{G}")),
	} {
		for _, r := range widened.Required {
			if r.Snow && r.AnyMana {
				t.Errorf("%s: the {S} was widened to any mana", name)
			}
		}
		if (ManaPool{{Color: "C"}, {Color: "G"}}).CanPay(widened, 0) {
			t.Errorf("%s: nonsnow mana paid the {S}", name)
		}
	}

	const mancer = "test-snow-mancer"
	g := newActiveGame(t)
	me := g.Seats[0]
	withCatalogCostModifiers(t, modifiersFor(mancer, CostModifier{
		Kind:   CostReduction,
		Label:  "Creature spells you cast cost {2} less to cast.",
		Amount: fixed(2),
	}))
	modifierSource(t, g, me, "Snow Test Summoning", mancer)
	id := spellInHand(t, g, me, "Snow Test Bear", "Snow Creature — Bear", "{1}{S}{G}")
	got := priceOf(t, g, me, id, CastSpellParams{})
	if got.String() != "{S}{G}" {
		t.Fatalf("{1}{S}{G} under a {2} reduction priced %s, want {S}{G}", got.String())
	}
	me.ManaPool.AddMana(ManaToken{Color: "G"}, ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err == nil {
		t.Fatal("the reduced {S}{G} paid with nonsnow mana: accepted")
	}
	me.ManaPool.AddMana(snowMana("U"))
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the reduced {S}{G} with a snow {U} and a {G}: %v", err)
	}
}

// A waived payment's paper picker offers any mana for an {S}, and a
// declaration of it is taken as snow mana: paper mana has no source to
// read, so the declaration is the caster's word for it.
func TestPaperPaymentDeclaresSnowMana(t *testing.T) {
	short := ManaPool{}.shortfallFor(costFor(t, "{S}"), 0, ManaSpendContext{}, SpendPreserveColors)
	slots := short.slots()
	if len(slots) != 1 || slots[0].Symbol != "{S}" || len(slots[0].Options) != len(paperSymbols) {
		t.Fatalf("paper slots for {S} = %+v, want one {S} slot offering every symbol", slots)
	}
	toks, err := short.tokensFor(PaperPayment{Mana: []PaperMana{{Color: "R"}}})
	if err != nil {
		t.Fatalf("declare a {R} for the {S}: %v", err)
	}
	if len(toks) != 1 || !toks[0].SourceKinds.Has(ManaSourceSnow) {
		t.Errorf("declared tokens %+v, want one snow {R}", toks)
	}
}
