package testsa

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// activation_limit_cards_test.go — #5752: "Activate no more than twice
// each turn" on the activation tally (AtMostNActivationsPerTurn with
// n = 2, Roterothopter), and the two Visions cards whose only missing
// piece was "Activate only once each turn" (n = 1): Spitting Drake and
// Knight of Valor.

const (
	roterothopterOracle = "69c69d69-3565-4d9c-82c0-a96d0fb80513"
	spittingDrakeOracle = "bfbc12ec-ecc0-4e23-a9d1-d13babd99a46"
	knightOfValorOracle = "f2edb543-de33-4eed-892c-98038eec96e2"
)

func TestActivationLimitCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		roterothopterOracle: "Roterothopter",
		spittingDrakeOracle: "Spitting Drake",
		knightOfValorOracle: "Knight of Valor",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Lookup = %+v, %v", name, spec, ok)
			continue
		}
		if len(spec.Activated) != 1 || spec.Activated[0].Condition == nil {
			t.Errorf("%s: want one activated ability with an activation limit", name)
		}
	}
}

// rotoActivationOffered reports whether the legal-move enumerator offers
// `seat` the source's activated ability.
func rotoActivationOffered(g *game.Game, seat, thopter uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeActivateAbility && m.Source == thopter {
			return true
		}
	}
	return false
}

// Two activations held on the stack at once are both allowed; the third
// is refused, greyed in the view and not offered. Both pumps resolve, and
// the count starts over the next turn.
func TestRoterothopterActivatesNoMoreThanTwiceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	thopter := effectstest.PushCatalogPermanent(g, me.ID, "Roterothopter", "Artifact Creature — Thopter", roterothopterOracle, false)
	effectstest.FindBattlefieldCardForTest(g, thopter).Power = 0
	effectstest.ActivatedRowOfMarkKnown(g)

	effectstest.FillPool(me, 6)
	for i := 1; i <= 2; i++ {
		if !rotoActivationOffered(g, me.ID, thopter) {
			t.Fatalf("activation %d is not offered", i)
		}
		if err := g.ActivateCatalogAbility(me.ID, thopter, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, thopter, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("third activation: err = %v, want ErrConditionNotMet", err)
	}
	if rotoActivationOffered(g, me.ID, thopter) {
		t.Error("the enumerator offers a third activation")
	}
	if row := effectstest.ActivatedRowOf(t, g, me.ID, thopter, 0); !row.ConditionUnmet {
		t.Error("the row is not greyed after two activations")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, thopter); got != 2 {
		t.Errorf("power %d after two pumps, want 2", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, thopter, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("third activation after both resolved: err = %v, want ErrConditionNotMet", err)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if row := effectstest.ActivatedRowOf(t, g, me.ID, thopter, 0); row.ConditionUnmet {
		t.Error("the row is still greyed on the next turn")
	}
	effectstest.FillPool(me, 2)
	if !rotoActivationOffered(g, me.ID, thopter) {
		t.Fatal("the ability is not offered on the next turn")
	}
	if err := g.ActivateCatalogAbility(me.ID, thopter, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation on the next turn: %v", err)
	}
}

// Spitting Drake pumps once; the second activation this turn is refused
// and not offered, and the next turn it may pump again.
func TestSpittingDrakeActivatesOnlyOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Spitting Drake", "Creature — Drake", spittingDrakeOracle, false)
	effectstest.FindBattlefieldCardForTest(g, drake).Power = 2

	effectstest.B06AddMana(me, "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("second activation: err = %v, want ErrConditionNotMet", err)
	}
	if rotoActivationOffered(g, me.ID, drake) {
		t.Error("the enumerator offers a second activation")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, drake); got != 3 {
		t.Errorf("power %d after one pump, want 3", got)
	}

	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, drake, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation on the next turn: %v", err)
	}
}

// Knight of Valor's ability shrinks each blocker without flanking, once
// each turn. Flanking already shrank the plain blocker as it blocked; the
// blocker with flanking and a creature outside the combat are untouched.
func TestKnightOfValorShrinksItsBlockersWithoutFlankingOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	knight := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Knight of Valor", OracleID: knightOfValorOracle, TypeLine: "Creature — Human Knight", Power: 2, Toughness: 2, Keywords: []string{"flanking"}})
	plain := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Plain", "", 2, 4))
	flanker := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flanker", TypeLine: "Creature — Test", Power: 2, Toughness: 4, Keywords: []string{"flanking"}})
	bystander := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bystander", "", 2, 4))

	effectstest.DeclareAttack(t, g, opp.ID, knight)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{plain, flanker} {
		if err := g.DeclareBlocker(b, knight); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ApaLive(g, plain).CurrentToughness(); got != 3 {
		t.Fatalf("plain blocker toughness %d after flanking, want 3", got)
	}

	effectstest.ApaMana(me, "C", "W", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("second activation: err = %v, want ErrConditionNotMet", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"plain blocker", plain, 2}, {"blocker with flanking", flanker, 4}, {"bystander", bystander, 4}} {
		if got := effectstest.ApaLive(g, tc.id).CurrentToughness(); got != tc.want {
			t.Errorf("%s toughness %d, want %d", tc.name, got, tc.want)
		}
	}
}
