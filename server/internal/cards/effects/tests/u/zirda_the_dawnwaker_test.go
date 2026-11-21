package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// zirda_the_dawnwaker_test.go — Zirda, the Dawnwaker (#5908, #5909).

const (
	zirdaOracle   = "7b515c56-7b1e-42a8-82b0-1b30201fb162"
	zirdaName     = "Zirda, the Dawnwaker"
	zirdaTypeLine = "Legendary Creature — Elemental Fox"
)

func TestZirdaIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(zirdaOracle)
	if !ok {
		t.Fatal("Zirda is not registered")
	}
	if spec.Name != zirdaName || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

func pushZirda(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.B12Push(g, owner, zirdaName, zirdaTypeLine, zirdaOracle, 3, 3)
}

// zirdaPrice prices a made-up non-mana ability of `cost` on card `id`
// in `zone` for `activator`, through the pricer the activation pays.
func zirdaPrice(t *testing.T, g *game.Game, activator, id uuid.UUID, zone game.ZoneKind, cost string) string {
	t.Helper()
	var (
		q   game.ParsedCost
		err error
	)
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("no card %s", id)
		}
		ab := game.ActivatedAbilityShape{Label: cost + ": test", Cost: game.AbilityCost{Mana: cost}}
		q, err = g.AbilityManaCostAtXForEffect(activator, c, zone, ab, nil, 0)
	})
	if err != nil {
		t.Fatalf("price %s: %v", cost, err)
	}
	return q.String()
}

// Generic mana only, never below one mana, never raised to one, and of
// any permanent's ability its controller activates, creature or not.
func TestZirdaTakesTwoGenericOffButLeavesOneMana(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	cases := []struct{ cost, want string }{
		{"{2}{G}", "{G}"},
		{"{1}{G}", "{G}"},
		{"{3}{R}", "{1}{R}"},
		{"{4}", "{2}"},
		{"{3}", "{1}"},
		{"{2}", "{1}"},
		{"{1}", "{1}"},
		{"{R}{R}", "{R}{R}"},
	}
	for _, id := range []uuid.UUID{rock, bear} {
		for _, tc := range cases {
			if got := zirdaPrice(t, g, me.ID, id, game.ZoneBattlefield, tc.cost); got != tc.cost {
				t.Errorf("without Zirda %s costs %s, want it unchanged", tc.cost, got)
			}
		}
	}
	pushZirda(g, me.ID)
	for _, id := range []uuid.UUID{rock, bear} {
		for _, tc := range cases {
			if got := zirdaPrice(t, g, me.ID, id, game.ZoneBattlefield, tc.cost); got != tc.want {
				t.Errorf("%s under Zirda costs %s, want %s", tc.cost, got, tc.want)
			}
		}
	}
}

// "Abilities you activate": an ability that works from the hand (cycling)
// is one, and an opponent's activation is not.
func TestZirdaReachesTheHandButNotAnOpponent(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	pushZirda(g, me.ID)
	cycler := effectstest.PushCatalogHandCard(me, "Cycler", "Creature — Beast", "")
	if got := zirdaPrice(t, g, me.ID, cycler, game.ZoneHand, "{3}"); got != "{1}" {
		t.Errorf("a {3} hand ability costs %s, want {1}", got)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if got := zirdaPrice(t, g, opp.ID, theirs, game.ZoneBattlefield, "{3}"); got != "{3}" {
		t.Errorf("an opponent's {3} ability costs %s, want {3}", got)
	}
}

// A mana ability is not reduced (CR 605.1a): its {3} stays {3}.
func TestZirdaLeavesManaAbilitiesAlone(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	pushZirda(g, me.ID)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	var (
		q   game.ParsedCost
		err error
	)
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(rock)
		q, err = g.ManaAbilityManaCostForEffect(me.ID, c, game.ManaAbilityShape{ManaCost: "{3}", TapCost: true})
	})
	if err != nil {
		t.Fatalf("ManaAbilityManaCostForEffect: %v", err)
	}
	if got := q.String(); got != "{3}" {
		t.Errorf("a {3} mana ability costs %s under Zirda, want {3}", got)
	}
}

// Engine-driven: Sower of Chaos's {2}{R} is paid with {R} alone under
// Zirda, and is refused that pool without it.
func TestZirdaDiscountIsWhatTheActivationPays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sower := effectstest.B12Push(g, me.ID, "Sower of Chaos", "Creature — Devil", effectstest.FdnSSowerOracle, 4, 3)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, sower, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(wall)}); err == nil {
		t.Fatal("{2}{R} was paid with {R} without Zirda")
	}
	pushZirda(g, me.ID)
	effectstest.B16Activate(t, g, me.ID, sower, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(wall)})
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.EffectiveOf(t, g, wall).Restrictions; r&game.CantBlock == 0 {
		t.Errorf("restrictions %v, want CantBlock", r)
	}
}

// "{1}, {T}: Target creature can't block this turn." Zirda's own {1}
// stays {1} under its reduction.
func TestZirdaStopsATargetCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	z := pushZirda(g, me.ID)
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, z, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(wall)}); err == nil {
		t.Fatal("the ability was activated with no mana")
	}
	effectstest.B06AddMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, z, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(wall)})
	if !effectstest.IsTapped(g, z) {
		t.Error("Zirda did not tap for its ability")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.EffectiveOf(t, g, wall).Restrictions; r&game.CantBlock == 0 {
		t.Errorf("restrictions %v, want CantBlock", r)
	}
	if r := effectstest.EffectiveOf(t, g, z).Restrictions; r&game.CantBlock != 0 {
		t.Error("Zirda itself can't block")
	}
}
