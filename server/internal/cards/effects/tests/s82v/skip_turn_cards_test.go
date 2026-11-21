package testss82v

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSkipTurnCardsAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]string{
		effectstest.MagosiOracle:       "Magosi, the Waterveil",
		effectstest.LethalVaporsOracle: "Lethal Vapors",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q completeness %v", want, spec.Name, spec.Completeness)
		}
	}
}

// --- Lethal Vapors ---------------------------------------------------

func TestLethalVaporsDestroysEveryCreatureThatEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTokenReplacementCard(g, effectstest.LethalVaporsOracle, "Lethal Vapors", "Enchantment", me.ID)

	effectstest.MakeTokens(t, g, opp.ID, effects.TokenCard("2/2 black Zombie"), 2)
	effectstest.MakeTokens(t, g, me.ID, effects.TokenCard("2/2 black Zombie"), 1)
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("creatures entering did not put triggers on the stack")
	}
	effectstest.DrcSettle(t, g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Turn.ActiveSeat
	opp := (owner + 2) % len(g.Seats)
	vapors := effectstest.PushCatalogPermanent(g, g.Seats[owner].ID, "Lethal Vapors", "Enchantment", effectstest.LethalVaporsOracle, false)
	effectstest.ApbRequireAnyPlayerRows(t, g, vapors, 0)

	if err := g.ActivateCatalogAbility(g.Seats[opp].ID, vapors, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("an opponent activates Lethal Vapors: %v", err)
	}
	if got := g.TurnSkipsFor(g.Seats[opp].ID); got != 0 {
		t.Fatalf("the skip was owed before the ability resolved: %d", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vapors) {
		t.Fatal("Lethal Vapors survived its own ability")
	}
	if got := g.TurnSkipsFor(g.Seats[opp].ID); got != 1 {
		t.Errorf("the activator owes %d skips, want 1", got)
	}
	if got := g.TurnSkipsFor(g.Seats[owner].ID); got != 0 {
		t.Errorf("the controller owes %d skips, want 0", got)
	}
	effectstest.AssertTurnOf(t, g, owner, false, "the turn the ability resolved in is not cut short")

	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (owner+1)%4, false, "next turn")
	effectstest.EndTurn(t, g)
	// The opponent's turn is the next one: skipped.
	effectstest.AssertTurnOf(t, g, (opp+1)%4, false, "after the opponent's skipped turn")
	if got := g.TurnSkipsFor(g.Seats[opp].ID); got != 0 {
		t.Errorf("the debt is not spent: %d", got)
	}
}

// The destruction and the skip are separate sentences: an enchantment
// that is already gone when the ability resolves still costs its
// activator the turn.
func TestLethalVaporsSkipHappensEvenIfItWasAlreadyDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Turn.ActiveSeat
	vapors := effectstest.PushCatalogPermanent(g, g.Seats[owner].ID, "Lethal Vapors", "Enchantment", effectstest.LethalVaporsOracle, false)
	if err := g.ActivateCatalogAbility(g.Seats[owner].ID, vapors, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(vapors) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[owner].ID); got != 1 {
		t.Errorf("owner owes %d skips, want 1", got)
	}
}

func addBlue(p *game.Player) { p.ManaPool.AddMana(game.ManaToken{Color: "U"}) }

func TestMagosiEntersTappedAndTapsForBlue(t *testing.T) {
	spec, _ := effects.Lookup(effectstest.MagosiOracle)
	if len(spec.Replacements) != 1 || len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != "{U}" {
		t.Fatalf("spec = %d replacements, %d mana abilities", len(spec.Replacements), len(spec.ManaAbilities))
	}
}

// The third ability puts an eon counter and writes the skip; the turn in
// progress goes on, and the controller's next turn is the one skipped.
func TestMagosiSkipYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := effectstest.MagosiOnBattlefield(g, g.Seats[me].ID)
	addBlue(g.Seats[me])

	if err := g.ActivateCatalogAbility(g.Seats[me].ID, magosi, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the skip ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.ApbCounters(g, magosi, "eon"); n != 1 {
		t.Errorf("eon counters = %d, want 1", n)
	}
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
	effectstest.AssertTurnOf(t, g, me, false, "the turn the ability resolved in is not cut short")

	// 1, 2, 3, then this player's turn is skipped and seat 1 is up again.
	for _, want := range []int{1, 2, 3, 1} {
		effectstest.EndTurn(t, g)
		effectstest.AssertTurnOf(t, g, want, false, "rotation with a skipped turn")
	}
}

// The skip is a sentence of its own: it does not depend on the counter,
// or on Magosi still being there when the ability resolves (the
// 2009-10-01 rulings).
func TestMagosiSkipAppliesEvenIfMagosiHasLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := effectstest.MagosiOnBattlefield(g, g.Seats[me].ID)
	addBlue(g.Seats[me])
	if err := g.ActivateCatalogAbility(g.Seats[me].ID, magosi, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(magosi) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips after Magosi left, want 1", got)
	}
}

// The fourth ability costs the counter and Magosi itself, and takes an
// extra turn.
func TestMagosiExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := effectstest.MagosiOnBattlefield(g, g.Seats[me].ID)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(magosi, "eon", 1) })

	if err := g.ActivateCatalogAbility(g.Seats[me].ID, magosi, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the extra-turn ability: %v", err)
	}
	if g.Battlefield.Contains(magosi) {
		t.Error("Magosi stayed on the battlefield: returning it is a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "Magosi's extra turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%4, false, "rotation resumes")
}

// Without a counter the fourth ability cannot be paid.
func TestMagosiExtraTurnNeedsAnEonCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	magosi := effectstest.MagosiOnBattlefield(g, g.Seats[me].ID)
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
			g := effectstest.NewCatalogGame(t)
			me := g.Turn.ActiveSeat
			skipper := effectstest.MagosiOnBattlefield(g, g.Seats[me].ID)
			taker := effectstest.MagosiOnBattlefield(g, g.Seats[me].ID)
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
			effectstest.PassPriorityAroundTable(t, g)
			if err := g.ActivateCatalogAbility(g.Seats[me].ID, second, secondIdx, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("second activation: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)

			// No extra turn: seat 1 follows, then 2, 3, and the player's
			// own normal turn comes back around.
			for _, want := range []int{1, 2, 3, me} {
				effectstest.EndTurn(t, g)
				effectstest.AssertTurnOf(t, g, want, false, "the extra turn is the one skipped")
			}
			if got := g.TurnSkipsFor(g.Seats[me].ID); got != 0 {
				t.Errorf("a debt is left over: %d", got)
			}
		})
	}
}
