package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// per_target_life_test.go — #6206, the enumerator's half of "This spell
// costs 3 life more to cast for each target" (Phyrexian Purge, CR
// 601.2f, 119.4).
//
// The contract is #544's: a move the enumerator offers is a cast the
// engine accepts. A per-target life cost makes the PRICE a function of
// the target set, exactly as a per-target mana surcharge and a Terror of
// the Peaks on the board do, so each set has to be measured against the
// life it would leave the seat — otherwise a bot at 5 life is offered
// three targets, the engine refuses the cast, and the soak logs a
// rejected move.

const perTargetLifeOracle = "test-6206-legal-purge"

func perTargetLifeTable(t *testing.T, life int) (g *game.Game, active *game.Player, spell uuid.UUID) {
	t.Helper()
	defs := map[string]*game.CardDef{
		perTargetLifeOracle: {LifePerTarget: game.LifePerTargetCost{
			Life:  3,
			Label: "This spell costs 3 life more to cast for each target.",
		}},
	}
	prevDef, prevSpec := game.CatalogLookup, game.CatalogTargetSpec
	game.CatalogLookup = func(key string) *game.CardDef { return defs[key] }
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != perTargetLifeOracle {
			return nil
		}
		return &game.TargetSpec{
			Mode: "creature", Label: "any number of target creatures",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 0,
		}
	}
	t.Cleanup(func() { game.CatalogLookup, game.CatalogTargetSpec = prevDef, prevSpec })

	g = newTable(t)
	active = g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	spell = handCard(active, game.Card{
		Name: "Test Purge", TypeLine: "Sorcery",
		ManaCost: "{2}{R}", OracleID: perTargetLifeOracle,
	})
	for range 3 {
		battlefieldCard(g, active, basic("Mountain", "Mountain"))
	}
	them := g.Seats[1]
	for _, name := range []string{"Their Bear", "Their Bull", "Their Boar"} {
		battlefieldCard(g, them, creature(name, "{2}", 2, 2))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	active.Life = life
	return g, active, spell
}

// At a comfortable life total every arity is offered, and every offered
// move is one the engine accepts.
func TestPerTargetLifeOffersEveryArityItCanPay(t *testing.T) {
	g, active, spell := perTargetLifeTable(t, 20)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, spell)
	if len(casts) == 0 {
		t.Fatalf("the spell is not offered at 20 life: %v", labels(moves))
	}
	sizes := map[int]int{}
	for _, m := range casts {
		_, n := castXAndTargets(t, m)
		sizes[n]++
	}
	for want := 1; want <= 3; want++ {
		if sizes[want] == 0 {
			t.Errorf("no %d-target cast offered at 20 life (sizes %v)", want, sizes)
		}
	}
	// The life it charges is in the move's line, so a bot trace, a model
	// seat and an MCP client read what the targets cost.
	named := false
	for _, m := range casts {
		if strings.Contains(m.Label, "life for what it targets") {
			named = true
		}
	}
	if !named {
		t.Errorf("no cast names the life it pays: %v", labels(casts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// CR 119.4, and the enumerator's contract: at 5 life the seat can pay
// for one target and not for two, so only the one-target casts are
// offered — and each of them is accepted.
func TestPerTargetLifeDropsTheSetsItCannotPay(t *testing.T) {
	g, active, spell := perTargetLifeTable(t, 5)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, spell)
	if len(casts) == 0 {
		t.Fatalf("one target is payable at 5 life: %v", labels(moves))
	}
	for _, m := range casts {
		if _, n := castXAndTargets(t, m); n != 1 {
			t.Errorf("at 5 life %q names %d targets, which costs %d life", m.Label, n, 3*n)
		}
	}
	dispatchAll(t, g, active.ID, moves)
}
