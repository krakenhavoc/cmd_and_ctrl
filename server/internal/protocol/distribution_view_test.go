package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// distribution_view_test.go — #4114: a divided or distributed effect
// (Armament Dragon's +1/+1 counters, a Fireball's damage) puts its
// division on the wire, keyed by target id, so the stack can say how
// much each target gets. The client reads it from stack_items[].distribution.

func TestStackItemCarriesItsDistribution(t *testing.T) {
	g := buildActiveGame(t)
	id := pushStackSpell(g, "Fireball", "fireball", game.PaidCost{}, game.StackItemSpell)
	a, b := uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		g.StackMeta[id].Distribution = map[uuid.UUID]int{a: 3, b: 1}
	})

	v := stackItemView(t, g, id)
	if len(v.Distribution) != 2 || v.Distribution[a.String()] != 3 || v.Distribution[b.String()] != 1 {
		t.Fatalf("distribution = %v, want {%s:3 %s:1}", v.Distribution, a, b)
	}
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"distribution":{`) {
		t.Fatalf("the wire item has no distribution: %s", raw)
	}
}

func TestStackItemWithoutADivisionSendsNoDistribution(t *testing.T) {
	g := buildActiveGame(t)
	id := pushStackSpell(g, "Shock", "shock", game.PaidCost{}, game.StackItemSpell)
	raw, err := json.Marshal(stackItemView(t, g, id))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "distribution") {
		t.Fatalf("an undivided item sends a distribution: %s", raw)
	}
}
