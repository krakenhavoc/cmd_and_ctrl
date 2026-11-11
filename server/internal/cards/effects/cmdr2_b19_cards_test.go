package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b19_cards_test.go — Commander wave 2, batch 19 (#5368): simple Auras.

const (
	cmdr2VowOfDuty      = "1e7bddbd-e448-4aea-9d18-b44bd45837c5"
	cmdr2VowOfLightning = "bfed67f3-70d2-49c2-8f25-16e4d10fa5be"
)

func TestCmdr2B19EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		cmdr2VowOfDuty:      "Vow of Duty",
		cmdr2VowOfLightning: "Vow of Lightning",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// cmdr2B19Attach puts an Aura of the catalog onto the battlefield attached
// to host, controlled by owner.
func cmdr2B19Attach(g *game.Game, owner, host uuid.UUID, name, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: auraTypeLine,
		OracleID: oracle, Owner: owner, Controller: owner, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
}

func TestCmdr2B19VowsPumpGrantAndBarAttackingTheirController(t *testing.T) {
	for _, tc := range []struct{ name, oracle, kw string }{
		{"Vow of Duty", cmdr2VowOfDuty, "vigilance"},
		{"Vow of Lightning", cmdr2VowOfLightning, "first strike"},
	} {
		g := newCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		host := b12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
		other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
		cmdr2B19Attach(g, me.ID, host, tc.name, tc.oracle)
		if p, to := effectivePTOf(t, g, host); p != 4 || to != 4 {
			t.Errorf("%s: host is %d/%d, want 4/4", tc.name, p, to)
		}
		if !hasEffectiveKeyword(t, g, host, tc.kw) {
			t.Errorf("%s: host lacks %s", tc.name, tc.kw)
		}
		var walker uuid.UUID
		g.WithWriteLock(func() {
			walker = pushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
				Counters: map[string]int{game.CounterLoyalty: 3}})
		})
		advanceToDeclareAttackersOf(t, g, 1)
		for _, target := range []uuid.UUID{me.ID, walker} {
			if err := g.Clone().DeclareAttacker(host, target); !errors.Is(err, game.ErrIllegalAttackTarget) {
				t.Errorf("%s: enchanted creature attacking %s: err = %v, want ErrIllegalAttackTarget", tc.name, target, err)
			}
		}
		if err := g.Clone().DeclareAttacker(host, third.ID); err != nil {
			t.Errorf("%s: enchanted creature attacking another player: %v", tc.name, err)
		}
		if err := g.Clone().DeclareAttacker(other, me.ID); err != nil {
			t.Errorf("%s: an unenchanted creature attacking me: %v", tc.name, err)
		}
	}
}
