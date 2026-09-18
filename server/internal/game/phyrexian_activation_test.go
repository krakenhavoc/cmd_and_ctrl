package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// phyrexian_activation_test.go — #917, the activation half of
// CR 107.4's "or 2 life".
//
// #915 gave the CAST an announce for it (CastSpellParams.PhyrexianLife,
// CR 601.2b) and left the activation without one, so Birthing Pod's
// "{1}{G/P}" could only ever be paid with {G}. CR 602.2b asks the
// same question in the same indivisible announcement, and these tests
// are phyrexian_mana_test.go's cast cases asked one rule number over
// — deliberately, because the point of #917 is that both paths run
// ONE strike-and-pay helper and therefore cannot answer differently.

// phyrexianPodSource seats Birthing Pod's shape: a permanent whose
// one ability costs `cost`, taps nothing, sacrifices nothing, and on
// resolution marks itself so a test can see the effect ran.
func phyrexianPodSource(g *Game, owner *Player, cost AbilityCost) uuid.UUID {
	c := NewCard("Phyrexian Pod", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "pay the Phyrexian symbol: mark",
		Cost:  cost,
		Effect: func(g *Game, item *StackItem) error {
			return g.applyCounterLocked(item.SourceCardID, "effect-ran", 1)
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// abilityCostMana is the one-component cost these tests activate.
func abilityCostMana(s string) AbilityCost { return AbilityCost{Mana: s} }

// lastAbilityPaid returns the PaidCost stamped on the one activation
// these tests announce. An activated ability's stack item is
// synthetic — it lives in StackMeta, not in Stack.Cards — so the
// record is read from there.
func lastAbilityPaid(t *testing.T, g *Game) PaidCost {
	t.Helper()
	if len(g.StackMeta) != 1 {
		t.Fatalf("StackMeta has %d items, want the one activation", len(g.StackMeta))
	}
	for _, item := range g.StackMeta {
		return item.Paid
	}
	return PaidCost{}
}

// TestActivationPaysAPhyrexianSymbolWithMana — the coloured half, and
// the baseline the life half is measured against: an activator with
// the green mana spends it and pays no life.
func TestActivationPaysAPhyrexianSymbolWithMana(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G/P}"))
	life := me.Life
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "G"})

	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activation paying the Phyrexian symbol with mana: %v", err)
	}
	if me.Life != life {
		t.Errorf("life = %d, want %d — the mana half must not charge life", me.Life, life)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the activation = %v, want empty", me.ManaPool)
	}
	if paid := lastAbilityPaid(t, g); paid.LifePaid != 0 {
		t.Errorf("PaidCost.LifePaid = %d, want 0", paid.LifePaid)
	}
}

// TestActivationPaysAPhyrexianSymbolWithLife — CR 107.4f through the
// activation announce: one short of the coloured half, the claim
// makes it payable, and the 2 life leaves through PayLifeForEffect.
func TestActivationPaysAPhyrexianSymbolWithLife(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G/P}"))
	life := me.Life
	// No green anywhere: the symbol is payable only with life.
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	// Without the claim the activation is short exactly the symbol.
	err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{Strict: true})
	var short *InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("activation with no life claim: got %v, want *InsufficientManaError", err)
	}
	if me.Life != life {
		t.Fatalf("a refused activation cost %d life", life-me.Life)
	}
	if len(me.ManaPool) != 1 {
		t.Fatalf("a refused activation spent mana: %v", me.ManaPool)
	}

	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("activation paying the Phyrexian symbol with 2 life: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d (CR 107.4f: 2 life)", me.Life, life-PhyrexianLifePerSymbol)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the activation = %v, want empty", me.ManaPool)
	}
	// ADR 0020's #958 addendum: the record carries what was paid.
	if paid := lastAbilityPaid(t, g); paid.LifePaid != PhyrexianLifePerSymbol {
		t.Errorf("PaidCost.LifePaid = %d, want %d", paid.LifePaid, PhyrexianLifePerSymbol)
	}
}

// TestActivationPhyrexianLifeSumsWithThePrintedLifeCost — an ability
// that prints BOTH a life component and a Phyrexian symbol records
// one LifePaid, because LifePaid is every point the announcement
// paid, not the printed component alone.
func TestActivationPhyrexianLifeSumsWithThePrintedLifeCost(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, AbilityCost{Mana: "{G/P}", Life: 3})
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	want := 3 + PhyrexianLifePerSymbol
	if me.Life != life-want {
		t.Errorf("life = %d, want %d", me.Life, life-want)
	}
	if paid := lastAbilityPaid(t, g); paid.LifePaid != want {
		t.Errorf("PaidCost.LifePaid = %d, want %d", paid.LifePaid, want)
	}
}

// TestActivationRefusesAnOverclaimedPhyrexianLife — CR 601.2b's rule
// at CR 602.2b: the announcement may only claim a symbol the cost
// prints, and CR 119.4 caps the payment at the life total. Both
// reject with nothing paid.
func TestActivationRefusesAnOverclaimedPhyrexianLife(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G/P}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"}, ManaToken{Color: "G"})

	// The cost prints one Phyrexian symbol; two is malformed.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 2,
	}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming two symbols on a one-symbol cost: got %v, want ErrInvalidParam", err)
	}
	if len(g.StackMeta) != 0 {
		t.Fatalf("the refused activation reached the stack")
	}
	if len(me.ManaPool) != 2 {
		t.Errorf("pool = %v, want both tokens back", me.ManaPool)
	}

	// Negative is malformed too.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: -1,
	}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming a negative count: got %v, want ErrInvalidParam", err)
	}

	// An ability with no mana component has nothing to claim
	// against, and a claim against it is refused rather than dropped.
	free := phyrexianPodSource(g, me, AbilityCost{Tap: true})
	if err := g.ActivateCatalogAbility(me.ID, free, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming a symbol on a costless ability: got %v, want ErrInvalidParam", err)
	}
}

// TestActivationRefusesLifeBelowZero — CR 119.4: a player may pay
// life only down to 0, and the check runs before anything is paid, so
// the refusal costs neither a point nor a token.
func TestActivationRefusesLifeBelowZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G/P}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"})
	me.Life = 1

	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("claiming 2 life at 1 life: got %v, want ErrInvalidParam", err)
	}
	if me.Life != 1 {
		t.Errorf("life = %d, want 1 — a refused activation paid anyway", me.Life)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the token back", me.ManaPool)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("the refused activation reached the stack")
	}

	// Down to EXACTLY 0 is legal (CR 119.4 caps the payment at the
	// life total, it does not require a survivor), and it is the same
	// bound the cast path enforces.
	me.Life = PhyrexianLifePerSymbol
	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("paying down to exactly 0: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life = %d, want 0", me.Life)
	}
}

// TestActivationPhyrexianLifeIsNotAutoTapped — the auto-tapper plans
// the mana the activation still owes, so a claimed symbol does not
// strand a land. The activation half of
// TestPhyrexianLifeIsNotAutoTapped.
func TestActivationPhyrexianLifeIsNotAutoTapped(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G/P}"))
	pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	pushBattlefieldForTest(g, me.ID, "Forest", "Basic Land — Forest", "")
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, AutoTap: true, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("auto-tapped activation with the symbol paid by life: %v", err)
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
	if tapped != 1 {
		t.Errorf("tapped %d lands, want 1 — the {1} the activation still owes", tapped)
	}
}

// TestUndoAcrossAPhyrexianActivation — the rewind takes the life back
// with the activation, and the replay charges it once more rather
// than twice.
func TestUndoAcrossAPhyrexianActivation(t *testing.T) {
	g := newActiveGame(t)
	// RestoreFrom swaps g.Seats wholesale, so the seat is re-read
	// after the rewind rather than captured once.
	seat := func() *Player { return g.Seats[0] }
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, seat(), abilityCostMana("{1}{G/P}"))
	life := seat().Life
	fund := func() { seat().ManaPool.AddMana(ManaToken{Color: "C"}) }
	fund()

	before := g.Clone()
	if err := g.ActivateCatalogAbility(seat().ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if seat().Life != life-PhyrexianLifePerSymbol {
		t.Fatalf("life after the activation = %d, want %d", seat().Life, life-PhyrexianLifePerSymbol)
	}

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if seat().Life != life {
		t.Errorf("life after the undo = %d, want %d — the payment did not rewind", seat().Life, life)
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("the activation survived the rewind on the stack")
	}
	if len(seat().ManaPool) != 1 {
		t.Errorf("pool after the undo = %v, want the token back", seat().ManaPool)
	}

	if err := g.ActivateCatalogAbility(seat().ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("replayed activation: %v", err)
	}
	if seat().Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life after the replay = %d, want %d — the payment doubled", seat().Life, life-PhyrexianLifePerSymbol)
	}
}

// TestActivationPhyrexianLifeInPermissiveMode — a life total is
// engine state in every mode, so the claimed life is paid even when
// the mana charge is waived and the record says OnPaper.
func TestActivationPhyrexianLifeInPermissiveMode(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{1}{G/P}"))
	life := me.Life

	// Permissive (the default): no mana at all, no Strict, no
	// AutoTap. The charge is waived; the life is not.
	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{PhyrexianLife: 1}); err != nil {
		t.Fatalf("permissive activation: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d — permissive mode waives mana, not life", me.Life, life-PhyrexianLifePerSymbol)
	}
	paid := lastAbilityPaid(t, g)
	if !paid.OnPaper {
		t.Errorf("PaidCost.OnPaper = false, want true — the mana charge was waived")
	}
	if paid.LifePaid != PhyrexianLifePerSymbol {
		t.Errorf("PaidCost.LifePaid = %d, want %d", paid.LifePaid, PhyrexianLifePerSymbol)
	}
}

// TestActivationStrikesTheUnpayableSymbolFirst — the shared strike
// order, reached through the activation announce: a cost printing two
// Phyrexian symbols of different colours against a pool that can pay
// one of them spends the life on the one it cannot.
func TestActivationStrikesTheUnpayableSymbolFirst(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	id := phyrexianPodSource(g, me, abilityCostMana("{G/P}{W/P}"))
	life := me.Life
	// Only white floating: the green pip has to be the one life buys.
	me.ManaPool.AddMana(ManaToken{Color: "W"})

	if err := g.ActivateCatalogAbility(me.ID, id, 0, ActivateAbilityParams{
		Strict: true, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("activation: %v", err)
	}
	if me.Life != life-PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-PhyrexianLifePerSymbol)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want the {W} spent on the {W/P}", me.ManaPool)
	}
}
