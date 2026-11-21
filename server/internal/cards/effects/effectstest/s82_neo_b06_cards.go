package effectstest

import (
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// neoNinjutsu swings a 1/1 at seat 1, then puts the named ninja in with
// ninjutsu and lets the ability resolve. The table is left in the
// declare-blockers step with the ninja attacking.
func NeoNinjutsu(t *testing.T, name, oracle, typeLine string, power, toughness int, mana string) (g *game.Game, ninja, attacker uuid.UUID, me, opp *game.Player) {
	t.Helper()
	g = NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker = B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja = PushNinjaToHand(me, name, typeLine, oracle, power, toughness)
	DeclareAttack(t, g, opp.ID, attacker)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, NeoPool(mana)); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ninja) {
		t.Fatal("the ninja is not on the battlefield")
	}
	return g, ninja, attacker, me, opp
}

// neoPool turns a printed cost into mana the pool takes: generic symbols
// become the cost's last colour ("{1}{B}" is "{B}{B}").
func NeoPool(cost string) string {
	last := cost[strings.LastIndex(cost, "{"):]
	out := ""
	for _, sym := range strings.SplitAfter(cost, "}") {
		if sym == "" {
			continue
		}
		if n, err := strconv.Atoi(strings.Trim(sym, "{}")); err == nil {
			out += strings.Repeat(last, n)
		} else {
			out += sym
		}
	}
	return out
}

func NeoToDamage(t *testing.T, g *game.Game) {
	t.Helper()
	AdvanceTo(t, g, game.StepCombatDamage)
}
