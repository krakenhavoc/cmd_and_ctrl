package testsg

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// halo_fountain_test.go — #5669: "Untap N tapped creatures you control"
// as an activation cost (game.TapOthersCost.Untap), proven on Halo
// Fountain.

const haloFountainOracle = "36a9de4c-8073-475d-8679-e50844eef09e"

// haloTable is a Halo Fountain under the active player in a main phase,
// with `tapped` tapped Bears.
func haloTable(t *testing.T, typeLine string, tapped int) (g *game.Game, me *game.Player, fountain uuid.UUID, bears []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	fountain = effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Halo Fountain", haloFountainOracle, typeLine, 0, 3))
	for i := 0; i < tapped; i++ {
		bears = append(bears, haloBear(g, me.ID, true))
	}
	effectstest.ToMainForCost(t, g)
	return g, me, fountain, bears
}

func haloBear(g *game.Game, owner uuid.UUID, tapped bool) uuid.UUID {
	id := effectstest.ApaPush(g, owner, owner, effectstest.StCard("Bear", "", "Creature — Bear", 2, 2))
	if tapped {
		effectstest.TapCard(g, id)
	}
	return id
}

func haloActivate(g *game.Game, me *game.Player, fountain uuid.UUID, index int, ids ...uuid.UUID) error {
	return g.ActivateCatalogAbility(me.ID, fountain, index, game.ActivateAbilityParams{Strict: true, TapIDs: ids})
}

func haloWhite(me *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.B06AddMana(me, "W")
	}
}

func TestHaloFountainIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(haloFountainOracle)
	if !ok || spec.Name != "Halo Fountain" {
		t.Fatalf("Lookup = %q, %v", spec.Name, ok)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
}

// The first ability untaps one tapped creature, taps the Fountain, and
// makes a Citizen. Only a tapped creature the activator controls pays.
func TestHaloFountainUntapsATappedCreatureForACitizen(t *testing.T) {
	g, me, fountain, bears := haloTable(t, "Artifact", 1)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	upright := haloBear(g, me.ID, false)
	theirs := haloBear(g, opp.ID, true)
	rock := effectstest.ApaPush(g, me.ID, me.ID, effectstest.StCard("Rock", "", "Artifact", 0, 0))
	effectstest.TapCard(g, rock)

	// A pool that can't pay {W}.
	if err := haloActivate(g, me, fountain, 0, bears[0]); err == nil {
		t.Fatal("activated with no mana")
	}
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, upright); !errors.Is(err, game.ErrNotTapped) {
		t.Fatalf("an untapped creature: err = %v, want ErrNotTapped", err)
	}
	if err := haloActivate(g, me, fountain, 0, theirs); err == nil {
		t.Fatal("an opponent's tapped creature paid the cost")
	}
	if err := haloActivate(g, me, fountain, 0, rock); err == nil {
		t.Fatal("a tapped noncreature paid the cost")
	}
	if err := haloActivate(g, me, fountain, 0); err == nil {
		t.Fatal("activated without untapping a creature")
	}
	effectstest.AssertTapped(t, g, true, bears[0], theirs, rock)
	effectstest.AssertTapped(t, g, false, fountain)

	if err := haloActivate(g, me, fountain, 0, bears[0]); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssertTapped(t, g, false, bears[0])
	effectstest.AssertTapped(t, g, true, fountain)
	if n := effectstest.UntapEventsFor(g, bears[0]); n != 1 {
		t.Errorf("%d untap events for the creature, want 1", n)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %v, want the {W} spent", me.ManaPool)
	}
	effectstest.PassPriorityAroundTable(t, g)
	n, tapped := effectstest.CountTokensNamed(g, me.ID, "Citizen")
	if n != 1 || tapped != 0 {
		t.Fatalf("%d Citizens (%d tapped), want one untapped", n, tapped)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Citizen" && (c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 2 || !strings.Contains(c.TypeLine, "Citizen")) {
			t.Errorf("the token is %q %d/%d %v, want a 1/1 green and white Citizen", c.TypeLine, c.Power, c.Toughness, c.Colors)
		}
	}

	// The Fountain is tapped now: {T} can't be paid again (CR 118.3).
	other := haloBear(g, me.ID, true)
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, other); !errors.Is(err, game.ErrAlreadyTapped) {
		t.Fatalf("a tapped Fountain: err = %v, want ErrAlreadyTapped", err)
	}
}

// It is not the {Q} symbol: a creature that came under your control
// this turn may be untapped to pay (CR 302.6).
func TestHaloFountainUntapsACreatureThatArrivedThisTurn(t *testing.T) {
	g, me, fountain, bears := haloTable(t, "Artifact", 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bears[0]).SummonedThisTurn = true })
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, bears[0]); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssertTapped(t, g, false, bears[0])
}

// The second ability untaps exactly two different tapped creatures and
// draws a card.
func TestHaloFountainUntapsTwoToDrawACard(t *testing.T) {
	g, me, fountain, bears := haloTable(t, "Artifact", 3)
	haloWhite(me, 2)
	if err := haloActivate(g, me, fountain, 1, bears[0]); err == nil {
		t.Fatal("one creature paid for two")
	}
	if err := haloActivate(g, me, fountain, 1, bears[0], bears[0]); err == nil {
		t.Fatal("one creature named twice paid for two")
	}
	if err := haloActivate(g, me, fountain, 1, bears...); err == nil {
		t.Fatal("three creatures were untapped for a cost of two")
	}
	effectstest.AssertTapped(t, g, true, bears...)
	hand := effectstest.HandCount(g, me)
	if err := haloActivate(g, me, fountain, 1, bears[0], bears[1]); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssertTapped(t, g, false, bears[0], bears[1])
	effectstest.AssertTapped(t, g, true, bears[2], fountain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, got)
	}
}

// The third ability: fourteen is not fifteen, and fifteen wins.
func TestHaloFountainUntapsFifteenToWinTheGame(t *testing.T) {
	g, me, fountain, bears := haloTable(t, "Artifact", 15)
	haloWhite(me, 5)
	if err := haloActivate(g, me, fountain, 2, bears[:14]...); err == nil {
		t.Fatal("fourteen creatures paid for fifteen")
	}
	if err := haloActivate(g, me, fountain, 2, bears...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssertTapped(t, g, false, bears...)
	if g.State != game.StateActive {
		t.Fatal("the game ended before the ability resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.State == game.StateActive || g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("state %v outcome %+v, want the game won by the Fountain's controller", g.State, g.Outcome)
	}
}

// CR 122.1d and CR 118.11: a stun counter is removed in place of the
// untap, the creature stays tapped, and the cost is still paid.
func TestHaloFountainCostIsPaidByACreatureWithAStunCounter(t *testing.T) {
	g, me, fountain, bears := haloTable(t, "Artifact", 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bears[0], game.CounterStun, 1) })
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, bears[0]); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AssertTapped(t, g, true, bears[0], fountain)
	if n := effectstest.CountersOn(g, bears[0], game.CounterStun); n != 0 {
		t.Errorf("%d stun counters left, want the one removed", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Citizen"); n != 1 {
		t.Errorf("%d Citizens, want 1", n)
	}
}

// CR 614.17b: a creature that can't become untapped can't pay.
func TestHaloFountainCostRefusesACreatureThatCantBecomeUntapped(t *testing.T) {
	g, me, fountain, _ := haloTable(t, "Artifact", 0)
	host := haloBear(g, me.ID, false)
	effectstest.CastCatalogSpell(t, g, "Frozen in Ice", effectstest.AuraTypeLine, effectstest.OracleFrozenInIce,
		[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AssertTapped(t, g, true, host)
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, host); !errors.Is(err, game.ErrNotTapped) {
		t.Fatalf("a creature that can't become untapped: err = %v, want ErrNotTapped", err)
	}
	effectstest.AssertTapped(t, g, true, host)
	effectstest.AssertTapped(t, g, false, fountain)
}

// CR 601.2h: a Fountain that has become a creature pays both halves of
// its own cost, in the order that works. Untapped, it is tapped for the
// {T} and then untapped; tapped, it is untapped and then tapped. A
// tapped one with a stun counter would stay tapped, so it can't.
func TestHaloFountainAsACreaturePaysItsOwnUntapCost(t *testing.T) {
	const animated = "Artifact Creature — Golem"

	g, me, fountain, _ := haloTable(t, animated, 0)
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, fountain); err != nil {
		t.Fatalf("untapped Fountain naming itself: %v", err)
	}
	effectstest.AssertTapped(t, g, false, fountain)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Citizen"); n != 1 {
		t.Errorf("%d Citizens, want 1", n)
	}

	g, me, fountain, bears := haloTable(t, animated, 1)
	effectstest.TapCard(g, fountain)
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, bears[0]); !errors.Is(err, game.ErrAlreadyTapped) {
		t.Fatalf("tapped Fountain untapping another creature: err = %v, want ErrAlreadyTapped", err)
	}
	if err := haloActivate(g, me, fountain, 0, fountain); err != nil {
		t.Fatalf("tapped Fountain naming itself: %v", err)
	}
	effectstest.AssertTapped(t, g, true, fountain, bears[0])
	if n := effectstest.UntapEventsFor(g, fountain); n != 1 {
		t.Errorf("%d untap events for the Fountain, want 1", n)
	}

	g, me, fountain, _ = haloTable(t, animated, 0)
	effectstest.TapCard(g, fountain)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(fountain, game.CounterStun, 1) })
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, fountain); err == nil {
		t.Fatal("a tapped Fountain with a stun counter paid its own {T}")
	}
	if n := effectstest.CountersOn(g, fountain, game.CounterStun); n != 1 {
		t.Errorf("%d stun counters, want the refused activation to leave 1", n)
	}

	// A summoning-sick creature Fountain can't pay its {T} (CR 302.6).
	g, me, fountain, bears = haloTable(t, animated, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, fountain).SummonedThisTurn = true })
	haloWhite(me, 1)
	if err := haloActivate(g, me, fountain, 0, bears[0]); !errors.Is(err, game.ErrSummoningSick) {
		t.Fatalf("summoning-sick Fountain: err = %v, want ErrSummoningSick", err)
	}
}

// The untap form is an activated ability's alone, with a fixed count.
func TestRegisterRefusesAnUntapCostOutsideAnActivatedAbility(t *testing.T) {
	untap := effects.UntapTapped(1, "a tapped creature you control", effects.Creature())
	for name, spec := range map[string]effects.Spec{
		"mana ability": {ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{TapOthers: untap}, Produced: "{G}", Label: "Add {G}",
		}}},
		"additional cost":  {AdditionalCost: &game.AdditionalCost{TapOthers: untap}},
		"alternative cost": {AlternativeCosts: []game.AlternativeCost{{Key: "untap", Label: "Untap", TapOthers: untap}}},
	} {
		spec.OracleID = uuid.NewString()
		spec.Name = "Untap Outside " + name
		if msg := effectstest.RegisterPanics(spec); !strings.Contains(msg, "untaps tapped permanents") {
			t.Errorf("%s: Register said %q, want the #5669 refusal", name, msg)
		}
	}
	zero := effects.UntapTapped(0, "tapped creatures you control", effects.Creature())
	spec := effects.Spec{OracleID: uuid.NewString(), Name: "Untap Nothing", Activated: []effects.ActivatedAbility{{
		Label: "Untap: nothing", Cost: game.AbilityCost{TapOthers: zero}, Effect: effects.Do(effects.DrawCards{N: 1}),
	}}}
	if msg := effectstest.RegisterPanics(spec); msg == "" {
		t.Error("Register accepted an untap cost of zero permanents")
	}
}
