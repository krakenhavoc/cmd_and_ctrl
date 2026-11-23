package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// professor_hojo_test.go — Professor Hojo (#3779): the discount on the
// first activation each turn that targets a creature you control, and the
// draw when a creature you control becomes the target of an activated
// ability, once each turn.
//
// The cost-side tests run with Strict set, so the payment is enforced: a
// discount that did not apply shows up as ErrInsufficientMana with the
// pool untouched, not as a free activation the sandbox waved through.

const (
	ProfessorHojoOracle = "4ca66b33-28c2-4d0d-8a44-9d891ab13e18"
	HojoProbeOracle     = "test-professor-hojo-probe"
	HojoPairProbeOracle = "test-professor-hojo-pair-probe"
	HojoTypeLine        = "Legendary Creature — Human Scientist"
)

func init() {
	noop := func(*game.Game, *game.StackItem) error { return nil }
	// {3}: target creature. Instant speed, so it can be activated on
	// either player's turn and at any creature.
	effects.Register(effects.Spec{
		OracleID: HojoProbeOracle,
		Name:     "Hojo Probe",
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Target creature does nothing.",
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  noop,
		}},
	})
	// {3}: two target creatures — one activation, two targets.
	effects.Register(effects.Spec{
		OracleID: HojoPairProbeOracle,
		Name:     "Hojo Pair Probe",
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}: Two target creatures do nothing.",
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCreature("two target creatures").WithCount(2, 2),
			Effect:  noop,
		}},
	})
}

func HojoRef(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// strictProbe activates the probe's one ability at `targets`, with the
// payment enforced and `mana` floated first.
func StrictProbe(t *testing.T, g *game.Game, who, probe uuid.UUID, mana string, targets ...uuid.UUID) error {
	t.Helper()
	if mana != "" {
		if err := g.AddManaForEffect(who, uuid.Nil, mana); err != nil {
			t.Fatalf("AddManaForEffect(%s): %v", mana, err)
		}
	}
	var refs []game.TargetRef
	for _, id := range targets {
		refs = append(refs, HojoRef(id))
	}
	return g.ActivateCatalogAbility(who, probe, 0, game.ActivateAbilityParams{Targets: refs, Strict: true})
}

func DrainPool(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.ID == who {
				p.ManaPool = game.ManaPool{}
			}
		}
	})
}
