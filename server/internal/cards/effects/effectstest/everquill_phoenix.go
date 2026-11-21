package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// everquill_phoenix_test.go — ADR 0223 (#5910), CR 702.140 and CR 730:
// mutate on the real catalog, with Everquill Phoenix. Over and under,
// abilities from the cards underneath, "whenever this creature
// mutates", an illegal target, the exits, a copy and a restore.

const (
	EverquillOracle = "56ba78c1-d1eb-4cc1-b70d-98c35d2972ba"
	llanowarOracle  = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
)

// llanowarOn puts a catalog Llanowar Elves ("{T}: Add {G}.") on the
// battlefield under its owner, since an earlier turn.
func LlanowarOn(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: llanowarOracle, ManaCost: "{G}", Colors: []string{"G"},
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

func CastMutating(g *game.Game, caster, spell, host uuid.UUID, strict bool) error {
	return g.CastSpell(caster, spell, game.CastSpellParams{
		AlternativeCost: game.MutateKey,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: host}},
		Strict:          strict,
	})
}

// resolveMutate passes priority until the over/under question, answers
// it (0 over, 1 under), and resolves whatever is left.
func ResolveMutate(t *testing.T, g *game.Game, controller uuid.UUID, under bool) {
	t.Helper()
	PassPriorityAroundTable(t, g)
	idx := 0
	if under {
		idx = 1
	}
	AnswerOptionPick(t, g, controller, idx)
	PassPriorityAroundTable(t, g)
}

// mergedView reads a battlefield permanent's merged facts under the
// write lock, with the layer cache caught up.
type mergedFacts struct {
	OnBattlefield    bool
	Name             string
	Power, Toughness int
	Flying           bool
	TapsForMana      bool
	Components       []uuid.UUID
	ComponentNames   []string
}

func ReadMerged(t *testing.T, g *game.Game, id uuid.UUID) mergedFacts {
	t.Helper()
	var f mergedFacts
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := FindBattlefieldCardByID(g, id)
	if c == nil {
		return f
	}
	f.OnBattlefield = true
	f.Name = c.Name
	f.Power, f.Toughness = c.PowerForComparison(), c.CurrentToughness()
	f.Flying = game.HasKeyword(c, "flying")
	f.TapsForMana = len(game.ManaAbilitiesForCard(*c)) > 0
	for _, part := range game.MergedComponentsForEffect(*c) {
		f.Components = append(f.Components, part.InstanceID)
		f.ComponentNames = append(f.ComponentNames, part.Name)
	}
	return f
}

func ZoneHolds(z *game.Zone, id uuid.UUID) bool { return z != nil && z.Contains(id) }
