package testsbr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6170 — Capricopian: "{2}: Put a +1/+1 counter on this creature, then
// you may reselect which player this creature is attacking. Only the
// player this creature is attacking may activate this ability and only
// during the declare attackers step."

const capricopianOracle = "3a5a6732-bdfa-4b64-b897-6817983dbb0e"

func capricopianAttacking(t *testing.T, g *game.Game, defender uuid.UUID) uuid.UUID {
	t.Helper()
	goat := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Capricopian", "Creature — Goat Hydra", capricopianOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == goat {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
			}
		}
	})
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(goat, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	g.RunStateChecksForTest()
	return goat
}

func capricopianCounters(g *game.Game, id uuid.UUID) int {
	n := 0
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c != nil {
			n = c.Counters[game.CounterPlusOne]
		}
	})
	return n
}

func capricopianTarget(g *game.Game, id uuid.UUID) uuid.UUID {
	var out uuid.UUID
	g.WithWriteLock(func() {
		if c := effectstest.FindBattlefieldCardForTest(g, id); c != nil {
			out = c.AttackingTarget
		}
	})
	return out
}

// CR 602.2: only the object's controller may activate its ability
// "unless the object specifically says otherwise". Capricopian names the
// player it is attacking: its controller and a player it is not
// attacking are refused, and the attacked player may activate it.
func TestCapricopianOnlyTheAttackedPlayerMayActivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	goat := capricopianAttacking(t, g, opp.ID)

	for _, p := range []*game.Player{me, third} {
		effectstest.ApaMana(p, "C", "C")
		err := g.ActivateCatalogAbility(p.ID, goat, 0, game.ActivateAbilityParams{Strict: true})
		if !errors.Is(err, game.ErrCardCallerMismatch) {
			t.Fatalf("%s activating the Goat it is not attacking: err = %v, want ErrCardCallerMismatch", p.Name, err)
		}
		p.ManaPool = nil
	}

	effectstest.ApaMana(opp, "C", "C")
	if err := g.ActivateCatalogAbility(opp.ID, goat, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the attacked player activating the Goat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := capricopianCounters(g, goat); got != 3 {
		t.Errorf("the Goat has %d +1/+1 counters, want 3", got)
	}

	// CR 508.7: the activator, not the Goat's controller, may reselect
	// it onto another player. Players only, and never its controller
	// (CR 508.7c).
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatal("the activator was not asked to reselect the Goat's attack")
	}
	choice := -1
	for i, o := range pick.PickOptions {
		if len(o.Cards) > 0 {
			t.Errorf("option %d (%q) is a permanent; Capricopian reselects a player", i, o.Label)
		}
		if o.Player == me.ID {
			t.Errorf("option %d offers the Goat's own controller", i)
		}
		if o.Player == third.ID {
			choice = i
		}
	}
	if choice < 0 {
		t.Fatalf("no option for the third player: %+v", pick.PickOptions)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, choice)
	if got := capricopianTarget(g, goat); got != third.ID {
		t.Fatalf("the Goat attacks %v after the reselection, want the third player", got)
	}

	// Now it is the third player's to activate, and no longer the first
	// defender's.
	opp.ManaPool = nil
	effectstest.ApaMana(opp, "C", "C")
	if err := g.ActivateCatalogAbility(opp.ID, goat, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("the player the Goat stopped attacking: err = %v, want ErrCardCallerMismatch", err)
	}
	effectstest.ApaMana(third, "C", "C")
	if err := g.ActivateCatalogAbility(third.ID, goat, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("the newly attacked player activating the Goat: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, third.ID, 0) // keep attacking
	if got := capricopianTarget(g, goat); got != third.ID {
		t.Errorf("declining the reselection moved the attack to %v", got)
	}
	if got := capricopianCounters(g, goat); got != 4 {
		t.Errorf("the Goat has %d +1/+1 counters, want 4", got)
	}
}

// "Only during the declare attackers step": in the declare blockers step
// the attacked player is refused, and a Goat that is not attacking has
// no activator at all.
func TestCapricopianOnlyDuringTheDeclareAttackersStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	goat := capricopianAttacking(t, g, opp.ID)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.ApaMana(opp, "C", "C")
	if err := g.ActivateCatalogAbility(opp.ID, goat, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the attacked player activated the Goat in the declare blockers step")
	}

	idle := effectstest.NewCatalogGame(t)
	still := effectstest.PushCatalogPermanent(idle, idle.Seats[0].ID, "Capricopian", "Creature — Goat Hydra", capricopianOracle, false)
	effectstest.AdvanceTo(t, idle, game.StepDeclareAttackers)
	for _, p := range idle.Seats {
		effectstest.ApaMana(p, "C", "C")
		if err := idle.ActivateCatalogAbility(p.ID, still, 0, game.ActivateAbilityParams{Strict: true}); !errors.Is(err, game.ErrCardCallerMismatch) {
			t.Errorf("%s activated a Goat that is not attacking: err = %v, want ErrCardCallerMismatch", p.Name, err)
		}
	}
}

// MayActivate is the one predicate the engine, the enumerator and the
// view share: it names the attacked seat and nobody else.
func TestDefendingPlayerOnlyMayActivate(t *testing.T) {
	ctrl, attacked, other := uuid.New(), uuid.New(), uuid.New()
	ab := game.ActivatedAbilityShape{DefendingPlayerOnly: true}
	src := game.Card{Controller: ctrl, Owner: ctrl, AttackingTarget: attacked}
	if !game.MayActivate(attacked, src, game.ZoneBattlefield, ab, game.AbilityOrigin{}) {
		t.Error("the attacked player may not activate it")
	}
	for name, p := range map[string]uuid.UUID{"its controller": ctrl, "another player": other, "nobody": uuid.Nil} {
		if game.MayActivate(p, src, game.ZoneBattlefield, ab, game.AbilityOrigin{}) {
			t.Errorf("%s may activate it", name)
		}
	}
	src.AttackingTarget = uuid.Nil
	if game.MayActivate(attacked, src, game.ZoneBattlefield, ab, game.AbilityOrigin{}) {
		t.Error("a creature that is not attacking still names an activator")
	}
	if !ab.ReachesAcross() {
		t.Error("a defending-player row does not reach across the table")
	}
}
