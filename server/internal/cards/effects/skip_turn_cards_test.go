package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skip_turn_cards_test.go — "skip your next turn" (CR 614.10, ADR 0188
// Decision 1, #3559): Magosi, the Waterveil and Lethal Vapors. Each card
// is driven through the real activation and the real rotation seam, and
// the skip is asserted by who is active after the turn ends.

const (
	magosiOracle       = "4bdffa67-e6b3-4588-b76e-c11db6f043ca"
	lethalVaporsOracle = "9fe9650d-6c9b-4988-9ea4-79ba0d55a77b"
)

func TestSkipTurnCardsAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]string{
		magosiOracle:       "Magosi, the Waterveil",
		lethalVaporsOracle: "Lethal Vapors",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q completeness %v", want, spec.Name, spec.Completeness)
		}
	}
}

// --- Lethal Vapors ---------------------------------------------------

func TestLethalVaporsDestroysEveryCreatureThatEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTokenReplacementCard(g, lethalVaporsOracle, "Lethal Vapors", "Enchantment", me.ID)

	makeTokens(t, g, opp.ID, TokenCard("2/2 black Zombie"), 2)
	makeTokens(t, g, me.ID, TokenCard("2/2 black Zombie"), 1)
	if stackFullyEmpty(g) {
		t.Fatal("creatures entering did not put triggers on the stack")
	}
	drcSettle(t, g, me.ID)
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() {
			t.Errorf("%s survived Lethal Vapors", c.Name)
		}
	}
}

// An opponent activates it, the enchantment is destroyed, and the
// opponent — "you", the activator (CR 109.5) — skips their next turn, not
// the controller.
func TestAnOpponentWhoDestroysLethalVaporsSkipsTheirNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Turn.ActiveSeat
	opp := (owner + 2) % len(g.Seats)
	vapors := pushCatalogPermanent(g, g.Seats[owner].ID, "Lethal Vapors", "Enchantment", lethalVaporsOracle, false)
	apbRequireAnyPlayerRows(t, g, vapors, 0)

	if err := g.ActivateCatalogAbility(g.Seats[opp].ID, vapors, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("an opponent activates Lethal Vapors: %v", err)
	}
	if got := g.TurnSkipsFor(g.Seats[opp].ID); got != 0 {
		t.Fatalf("the skip was owed before the ability resolved: %d", got)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vapors) {
		t.Fatal("Lethal Vapors survived its own ability")
	}
	if got := g.TurnSkipsFor(g.Seats[opp].ID); got != 1 {
		t.Errorf("the activator owes %d skips, want 1", got)
	}
	if got := g.TurnSkipsFor(g.Seats[owner].ID); got != 0 {
		t.Errorf("the controller owes %d skips, want 0", got)
	}
	assertTurnOf(t, g, owner, false, "the turn the ability resolved in is not cut short")

	endTurn(t, g)
	assertTurnOf(t, g, (owner+1)%4, false, "next turn")
	endTurn(t, g)
	// The opponent's turn is the next one: skipped.
	assertTurnOf(t, g, (opp+1)%4, false, "after the opponent's skipped turn")
	if got := g.TurnSkipsFor(g.Seats[opp].ID); got != 0 {
		t.Errorf("the debt is not spent: %d", got)
	}
}

// The destruction and the skip are separate sentences: an enchantment
// that is already gone when the ability resolves still costs its
// activator the turn.
func TestLethalVaporsSkipHappensEvenIfItWasAlreadyDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Turn.ActiveSeat
	vapors := pushCatalogPermanent(g, g.Seats[owner].ID, "Lethal Vapors", "Enchantment", lethalVaporsOracle, false)
	if err := g.ActivateCatalogAbility(g.Seats[owner].ID, vapors, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(vapors) })
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[owner].ID); got != 1 {
		t.Errorf("owner owes %d skips, want 1", got)
	}
}

// --- Magosi, the Waterveil ------------------------------------------

func magosiOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Magosi, the Waterveil", "Land", magosiOracle, false)
}

func addBlue(p *game.Player) { p.ManaPool.AddMana(game.ManaToken{Color: "U"}) }

func TestMagosiEntersTappedAndTapsForBlue(t *testing.T) {
	spec, _ := Lookup(magosiOracle)
	if len(spec.Replacements) != 1 || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{U}" {
		t.Fatalf("spec = %d replacements, %d mana abilities", len(spec.Replacements), len(spec.ManaAbilities))
	}
}

// The third ability puts an eon counter and writes the skip; the turn in
// progress goes on, and the controller's next turn is the one skipped.
func TestMagosiSkipYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := magosiOnBattlefield(g, g.Seats[me].ID)
	addBlue(g.Seats[me])

	if err := g.ActivateCatalogAbility(g.Seats[me].ID, magosi, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the skip ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := apbCounters(g, magosi, "eon"); n != 1 {
		t.Errorf("eon counters = %d, want 1", n)
	}
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
	assertTurnOf(t, g, me, false, "the turn the ability resolved in is not cut short")

	// 1, 2, 3, then this player's turn is skipped and seat 1 is up again.
	for _, want := range []int{1, 2, 3, 1} {
		endTurn(t, g)
		assertTurnOf(t, g, want, false, "rotation with a skipped turn")
	}
}

// The skip is a sentence of its own: it does not depend on the counter,
// or on Magosi still being there when the ability resolves (the
// 2009-10-01 rulings).
func TestMagosiSkipAppliesEvenIfMagosiHasLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := magosiOnBattlefield(g, g.Seats[me].ID)
	addBlue(g.Seats[me])
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, magosi, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(magosi) })
	passPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips after Magosi left, want 1", got)
	}
}

// The fourth ability costs the counter and Magosi itself, and takes an
// extra turn.
func TestMagosiExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := magosiOnBattlefield(g, g.Seats[me].ID)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(magosi, "eon", 1) })

	if err := g.ActivateCatalogAbility(g.Seats[me].ID, magosi, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the extra-turn ability: %v", err)
	}
	if g.Battlefield.Contains(magosi) {
		t.Error("Magosi stayed on the battlefield: returning it is a cost")
	}
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "Magosi's extra turn")
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%4, false, "rotation resumes")
}

// Without a counter the fourth ability cannot be paid.
func TestMagosiExtraTurnNeedsAnEonCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := magosiOnBattlefield(g, g.Seats[me].ID)
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, magosi, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the extra-turn ability was paid with no eon counter")
	}
}

// The 2009-10-01 ruling: using both in one turn cancels out. The next
// turn the player would begin is the extra turn the fourth ability made,
// and that is the one the skip is spent on. Either order of resolution.
func TestMagosiBothAbilitiesInOneTurnCancelOut(t *testing.T) {
	for name, skipFirst := range map[string]bool{"skip resolves first": true, "extra turn resolves first": false} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Turn.ActiveSeat
			skipper := magosiOnBattlefield(g, g.Seats[me].ID)
			taker := magosiOnBattlefield(g, g.Seats[me].ID)
			g.WithWriteLock(func() { _ = g.AddCounterForEffect(taker, "eon", 1) })
			addBlue(g.Seats[me])

			first, second := skipper, taker
			firstIdx, secondIdx := 0, 1
			if !skipFirst {
				first, second, firstIdx, secondIdx = taker, skipper, 1, 0
			}
			if err := g.ActivateCatalogAbility(g.Seats[me].ID, first, firstIdx, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("first activation: %v", err)
			}
			passPriorityAroundTable(t, g)
			if err := g.ActivateCatalogAbility(g.Seats[me].ID, second, secondIdx, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("second activation: %v", err)
			}
			passPriorityAroundTable(t, g)

			// No extra turn: seat 1 follows, then 2, 3, and the player's
			// own normal turn comes back around.
			for _, want := range []int{1, 2, 3, me} {
				endTurn(t, g)
				assertTurnOf(t, g, want, false, "the extra turn is the one skipped")
			}
			if got := g.TurnSkipsFor(g.Seats[me].ID); got != 0 {
				t.Errorf("a debt is left over: %d", got)
			}
		})
	}
}

// --- The interim Trouble in Pairs default ---------------------------

// A player who owes a skip and whose extra turn an opponent's Trouble in
// Pairs would also skip: both are replacements on one event (CR 616.1).
// Until ADR 0188 slice 3 asks the player, the debt is spent first, so the
// charge is spent and EventExtraTurnSkipped is not emitted. That slice
// replaces this with the question and changes this test.
func TestASkipDebtIsSpentBeforeTroubleInPairsSeesTheExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Turn.ActiveSeat
	other := (caster + 1) % len(g.Seats)
	seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)
	g.WithWriteLock(func() { g.SkipNextTurnsForEffect(g.Seats[caster].ID, uuid.Nil, 1, "probe") })

	castCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", temporalManipulationOracle, nil)
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, other, false, "after the extra turn is skipped")
	if got := g.TurnSkipsFor(g.Seats[caster].ID); got != 0 {
		t.Errorf("the charge was not spent: %d owed", got)
	}
	if got := skippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0: the debt took the turn", got)
	}
}
