package effectstest

import (
	"encoding/json"
	"errors"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// combat_limits_test.go — #1507's proof cards: the CR 508.1c / 509.1b
// whole-combat count limits (ADR 0045 amendment of 2026-09-24,
// Decisions 43-45). Silent Arbiter, Dueling Grounds and Caverns of
// Despair carry both halves; Crawlspace and Judoon Enforcers the
// per-defender attack half.
//
// Every shape is pinned the way block_rules_test.go pins #750's: the
// VERB accepts the legal declaration and refuses the illegal one with
// its reason, and the ENUMERATOR offers exactly what the verb accepts —
// checked exhaustively, by trying every candidate on a clone.

const (
	SilentArbiterOracle    = "1cdf30de-d88c-421a-80df-4917bbd2f09e"
	CrawlspaceOracle       = "2296370c-fe34-4df6-92a5-260f1634bede"
	DuelingGroundsOracle   = "eb2df4a1-6b63-4f6a-a830-e9487afe59f8"
	CavernsOfDespairOracle = "a1034a02-36cf-4586-a001-9dc3fb76e904"
	JudoonEnforcersOracle  = "ca04089c-24b6-465e-9303-ea28c0d6f3c7"
)

// clAttack is one (attacker, target) pair.
type clAttack struct{ attacker, target uuid.UUID }

func clSorted(in []clAttack) []clAttack {
	sort.Slice(in, func(i, j int) bool {
		if in[i].attacker != in[j].attacker {
			return in[i].attacker.String() < in[j].attacker.String()
		}
		return in[i].target.String() < in[j].target.String()
	})
	return in
}

// clOfferedAttacks is every attack the enumerator offers `seat`.
func clOfferedAttacks(t *testing.T, g *game.Game, seat uuid.UUID) []clAttack {
	t.Helper()
	var out []clAttack
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind != legal.KindAttack {
			continue
		}
		var p struct {
			Attacker string `json:"attacker"`
			Target   string `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("attack params %s: %v", m.Params, err)
		}
		out = append(out, clAttack{uuid.MustParse(p.Attacker), uuid.MustParse(p.Target)})
	}
	return clSorted(out)
}

// clAcceptedAttacks is every (creature of seat, opponent) attack the
// VERB accepts right now, each tried on its own clone.
func clAcceptedAttacks(g *game.Game, seat uuid.UUID) []clAttack {
	var out []clAttack
	for _, c := range g.Battlefield.Cards {
		if c.Controller != seat || !c.IsCreature() || c.AttackingTarget != uuid.Nil {
			continue
		}
		for _, p := range g.Seats {
			if p.ID == seat {
				continue
			}
			if err := g.Clone().DeclareAttacker(c.InstanceID, p.ID); err == nil {
				out = append(out, clAttack{c.InstanceID, p.ID})
			}
		}
	}
	return clSorted(out)
}

// clAgreeOnAttacks asserts the enumerator offers exactly what the verb
// accepts, and returns the offer.
func ClAgreeOnAttacks(t *testing.T, g *game.Game, seat uuid.UUID) []clAttack {
	t.Helper()
	offered, accepted := clOfferedAttacks(t, g, seat), clAcceptedAttacks(g, seat)
	if len(offered) != len(accepted) {
		t.Fatalf("the enumerator offers %d attacks, the verb accepts %d:\noffered  %v\naccepted %v", len(offered), len(accepted), offered, accepted)
	}
	for i := range offered {
		if offered[i] != accepted[i] {
			t.Fatalf("the enumerator and the verb disagree:\noffered  %v\naccepted %v", offered, accepted)
		}
	}
	return offered
}

func ClAttackRefusal(t *testing.T, err error) *game.AttackLimitError {
	t.Helper()
	var le *game.AttackLimitError
	if err == nil || !errors.Is(err, game.ErrAttackLimit) || !errors.As(err, &le) {
		t.Fatalf("want an attack-limit refusal, got %v", err)
	}
	return le
}

func ClTargets(offered []clAttack) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, a := range offered {
		out[a.target]++
	}
	return out
}
