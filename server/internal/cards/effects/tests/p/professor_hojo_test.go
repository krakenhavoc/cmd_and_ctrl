package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

type hojoTable struct {
	g             *game.Game
	me, opp       *game.Player
	hojo          uuid.UUID
	bear, oppBear uuid.UUID
	probe, probe2 uuid.UUID
	handAtStart   int
}

// newHojoTable seats Hojo and a bear for seat 0 in their main phase, an
// opposing bear for seat 1, and two probes for seat 0.
func newHojoTable(t *testing.T) hojoTable {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	h := hojoTable{g: g, me: me, opp: opp}
	h.hojo = effectstest.PushCatalogPermanent(g, me.ID, "Professor Hojo", effectstest.HojoTypeLine, effectstest.ProfessorHojoOracle, false)
	h.bear = effectstest.SeedBear(g, me.ID)
	h.oppBear = effectstest.SeedBear(g, opp.ID)
	h.probe = effectstest.PushCatalogPermanent(g, me.ID, "Hojo Probe", "Artifact", effectstest.HojoProbeOracle, false)
	h.probe2 = effectstest.PushCatalogPermanent(g, me.ID, "Hojo Probe", "Artifact", effectstest.HojoProbeOracle, false)
	h.handAtStart = len(me.Hand.Cards)
	return h
}

func handDelta(h hojoTable) int { return len(h.me.Hand.Cards) - h.handAtStart }

func TestProfessorHojoIsFull(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.ProfessorHojoOracle)
	if !ok {
		t.Fatal("Professor Hojo is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %+v, want Full and none", spec.Completeness, spec.Caveats)
	}
}

// The first activation that targets my own creature is {2} cheaper; the
// second pays the printed price.
func TestProfessorHojoDiscountsOnlyTheFirstOwnCreatureActivation(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me

	// {3} less {2} = {1}: {C} is exactly enough, nothing is not.
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "", h.bear); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("first activation with an empty pool: err = %v, want insufficient mana (it still costs {1})", err)
	}
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatalf("first own-creature activation with {1} (discounted from {3}): %v", err)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 0 {
		t.Errorf("pool after the discounted activation = %d, want 0 — it charged exactly {1}", got)
	}

	// The second pays the printed {3}: {2} is refused and spends nothing.
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe2, "{C}{C}", h.bear); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("second own-creature activation with {2}: err = %v, want insufficient mana", err)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 2 {
		t.Errorf("a refused activation spent mana: pool = %d, want 2", got)
	}
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe2, "{C}", h.bear); err != nil {
		t.Fatalf("second own-creature activation with the full {3}: %v", err)
	}
}

// An activation at someone else's creature neither gets the discount nor
// uses it up: the first one that does target my creature is still cheap.
func TestProfessorHojoOpposingTargetIsNotDiscountedAndDoesNotSpendIt(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me

	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.oppBear); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("activating at an opposing creature with {1}: err = %v, want insufficient mana", err)
	}
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}{C}", h.oppBear); err != nil {
		t.Fatalf("activating at an opposing creature with the full {3}: %v", err)
	}
	effectstest.DrainPool(t, g, me.ID)
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe2, "{C}", h.bear); err != nil {
		t.Fatalf("the first own-creature activation after an opposing one is still {1}: %v", err)
	}
}

// A refused activation is not an activation: it leaves the discount for
// the next one (the record is written at the announce).
func TestProfessorHojoRefusedActivationKeepsTheDiscount(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me

	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "", h.bear); err == nil {
		t.Fatal("an unpaid activation was accepted")
	}
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatalf("the discount should still be there after a refusal: %v", err)
	}
}

// "During your turn": on an opponent's turn the same activation is the
// printed {3}.
func TestProfessorHojoDoesNotDiscountOnAnotherPlayersTurn(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me
	effectstest.AdvanceToMainOf(t, g, 1)

	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("own-creature activation on an opponent's turn with {1}: err = %v, want insufficient mana", err)
	}
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}{C}", h.bear); err != nil {
		t.Fatalf("own-creature activation on an opponent's turn with {3}: %v", err)
	}
}

// "You activate": an opponent's Hojo discounts nobody else.
func TestProfessorHojoOnlyDiscountsItsControllersActivations(t *testing.T) {
	h := newHojoTable(t)
	g, me, opp := h.g, h.me, h.opp
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == h.hojo {
				g.Battlefield.Cards[i].Controller = opp.ID
				g.Battlefield.Cards[i].Owner = opp.ID
			}
		}
	})
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("an opponent's Hojo discounted my activation: err = %v", err)
	}
}

// The record resets with the turn: the next turn's first one is cheap
// again.
func TestProfessorHojoDiscountRefreshesNextTurn(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatalf("turn 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DrainPool(t, g, me.ID)
	effectstest.NextTurnOfMine(t, g)
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatalf("the discount did not come back next turn: %v", err)
	}
}

// The equip ability is the printed use: it targets my creature, is the
// first such activation (so it is cheaper), and draws.
func TestProfessorHojoEquipIsDiscountedAndDraws(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me
	sword := effectstest.SeedEquipment(g, me.ID, "Sword of Fire and Ice", effectstest.SwordFireOracle)
	// Equip {2}, less {2}: free.
	if err := effectstest.StrictEquip(t, g, me.ID, sword, h.bear, ""); err != nil {
		t.Fatalf("equip with an empty pool under Hojo: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := handDelta(h); got != 1 {
		t.Errorf("hand +%d after equipping, want +1", got)
	}
}

// --- the draw ---------------------------------------------------------

// A creature of mine becoming the target of an activated ability draws a
// card — once, however many times it happens in the turn.
func TestProfessorHojoDrawsOncePerTurn(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me

	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := handDelta(h); got != 1 {
		t.Fatalf("hand +%d after the first activation at my creature, want +1", got)
	}
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe2, "{C}{C}{C}", h.bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := handDelta(h); got != 1 {
		t.Errorf("hand +%d after a second activation the same turn, want still +1 (once each turn)", got)
	}
}

// ...and the limit lifts the next turn.
func TestProfessorHojoDrawsAgainNextTurn(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DrainPool(t, g, me.ID)
	effectstest.NextTurnOfMine(t, g)
	before := len(me.Hand.Cards)
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 1 {
		t.Errorf("hand +%d after next turn's activation, want +1", got)
	}
}

// One activation naming two of my creatures is one trigger: "one or more"
// collapses the batch.
func TestProfessorHojoOneOrMoreCreaturesIsOneDraw(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me
	pair := effectstest.PushCatalogPermanent(g, me.ID, "Hojo Pair Probe", "Artifact", effectstest.HojoPairProbeOracle, false)
	other := effectstest.SeedBear(g, me.ID)

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, pair, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.HojoRef(h.bear), effectstest.HojoRef(other)}, Strict: true,
	}); err != nil {
		t.Fatalf("two-target activation ({3} less {2}): %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := handDelta(h); got != 1 {
		t.Errorf("hand +%d after one activation at two creatures, want +1", got)
	}
}

// An activation at the opponent's creature draws nothing: the clause is
// "a creature YOU control".
func TestProfessorHojoNoDrawForAnOpposingCreature(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me
	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}{C}{C}", h.oppBear); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := handDelta(h); got != 0 {
		t.Errorf("hand +%d after targeting an opposing creature, want +0", got)
	}
}

// Whose ability it is does not matter: an opponent's activation at my
// creature draws me a card.
func TestProfessorHojoDrawsWhenAnOpponentActivatesAtMyCreature(t *testing.T) {
	h := newHojoTable(t)
	g, opp := h.g, h.opp
	oppProbe := effectstest.PushCatalogPermanent(g, opp.ID, "Hojo Probe", "Artifact", effectstest.HojoProbeOracle, false)
	if err := effectstest.StrictProbe(t, g, opp.ID, oppProbe, "{C}{C}{C}", h.bear); err != nil {
		t.Fatalf("opponent's activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := handDelta(h); got != 1 {
		t.Errorf("hand +%d after an opponent's activation at my creature, want +1", got)
	}
}

// A SPELL that targets my creature is not an activated ability.
func TestProfessorHojoNoDrawForASpell(t *testing.T) {
	h := newHojoTable(t)
	g := h.g
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", "5748ebf1-24e3-499d-ab7c-c2cebd462a24",
		[]game.TargetRef{effectstest.HojoRef(h.bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := handDelta(h); got != 0 {
		t.Errorf("hand +%d after a spell targeted my creature, want +0 — Hojo reads activated abilities only", got)
	}
}

// The row prices each legal target on its own, so the client shows {1}
// on my creature and {3} on the opposing one before the click, and
// stops showing the discount once it is spent.
func TestProfessorHojoViewPricesEachLegalTarget(t *testing.T) {
	h := newHojoTable(t)
	g, me := h.g, h.me

	row := effectstest.ActivatedRowOf(t, g, me.ID, h.probe, 0)
	want := map[string]string{h.bear.String(): "{1}", h.oppBear.String(): "{3}"}
	for id, price := range want {
		if got := row.TargetChargedManaCosts[id]; got != price {
			t.Errorf("target_charged_mana_costs[%s] = %q, want %q (all: %v)", id, got, price, row.TargetChargedManaCosts)
		}
	}

	if err := effectstest.StrictProbe(t, g, me.ID, h.probe, "{C}", h.bear); err != nil {
		t.Fatal(err)
	}
	row = effectstest.ActivatedRowOf(t, g, me.ID, h.probe2, 0)
	if got := row.TargetChargedManaCosts[h.bear.String()]; got != "{3}" {
		t.Errorf("after the discount is spent my creature prices at %q, want {3}", got)
	}
}
