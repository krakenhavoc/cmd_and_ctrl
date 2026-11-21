package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_fresh_layers_test.go pins #4909: inside ONE resolution, a
// characteristic read through the effect API sees every earlier step of
// that resolution (CR 608.2), with no explicit
// RecomputeLayersIfStaleLocked between the step and the read. Each probe
// is a synthetic instant whose body does one step to its target and then
// reads the target's power and toughness the way card bodies do.

const (
	FreshPumpProbe    = "test-4909-pump-then-read"
	FreshCounterProbe = "test-4909-counter-then-read"
	FreshAnthemProbe  = "test-4909-anthem-enters-then-read"
)

// freshRead is what the last probe read: LookupCardForEffect's P/T, and
// b31CurrentPowerOf's power.
var FreshRead struct{ Power, Toughness, B31Power int }

func freshReadTarget(ctx *effects.Context) (uuid.UUID, bool) {
	t, ok := ctx.ClauseTarget(0)
	return t.ID, ok && t.Kind == game.TargetCard
}

func recordFreshRead(ctx *effects.Context, id uuid.UUID) {
	c, _ := ctx.Game.LookupCardForEffect(id)
	FreshRead.Power, FreshRead.Toughness = c.CurrentPower(), c.CurrentToughness()
	FreshRead.B31Power = effects.B31CurrentPowerOf(ctx.Game, id)
}

func init() {
	probe := func(oracle, name string, step func(ctx *effects.Context, id uuid.UUID) error) {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     name,
			Targets:  effects.Clauses(effects.TargetCreature("target creature")),
			OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
				id, ok := freshReadTarget(ctx)
				if !ok {
					return nil
				}
				if err := step(ctx, id); err != nil {
					return err
				}
				recordFreshRead(ctx, id)
				return nil
			},
		})
	}
	probe(FreshPumpProbe, "Pump Then Read", func(ctx *effects.Context, id uuid.UUID) error {
		return effects.BoostUntilEOT{Target: id, Power: 1, Label: "+1/+0 until end of turn"}.Apply(ctx)
	})
	probe(FreshCounterProbe, "Counter Then Read", func(ctx *effects.Context, id uuid.UUID) error {
		return ctx.Game.AddCounterForEffect(id, "+1/+1", 1)
	})
	// "Put the top card of your library onto the battlefield" with a
	// Glorious Anthem there: the anthem applies at once (CR 611.3a).
	probe(FreshAnthemProbe, "Anthem Then Read", func(ctx *effects.Context, id uuid.UUID) error {
		p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
		top, err := p.Library.Top()
		if err != nil {
			return nil
		}
		_, err = ctx.Game.PutCardsFromLibraryOntoBattlefieldForEffect([]uuid.UUID{top.InstanceID}, game.LibraryEntryOptions{})
		return err
	})
}
