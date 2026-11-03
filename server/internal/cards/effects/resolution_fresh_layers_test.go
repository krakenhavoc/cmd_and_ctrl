package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_fresh_layers_test.go pins #4909: inside ONE resolution, a
// characteristic read through the effect API sees every earlier step of
// that resolution (CR 608.2), with no explicit
// RecomputeLayersIfStaleLocked between the step and the read. Each probe
// is a synthetic instant whose body does one step to its target and then
// reads the target's power and toughness the way card bodies do.

const (
	freshPumpProbe    = "test-4909-pump-then-read"
	freshCounterProbe = "test-4909-counter-then-read"
	freshAnthemProbe  = "test-4909-anthem-enters-then-read"
)

// freshRead is what the last probe read: LookupCardForEffect's P/T, and
// b31CurrentPowerOf's power.
var freshRead struct{ power, toughness, b31Power int }

func freshReadTarget(ctx *Context) (uuid.UUID, bool) {
	t, ok := ctx.ClauseTarget(0)
	return t.ID, ok && t.Kind == game.TargetCard
}

func recordFreshRead(ctx *Context, id uuid.UUID) {
	c, _ := ctx.Game.LookupCardForEffect(id)
	freshRead.power, freshRead.toughness = c.CurrentPower(), c.CurrentToughness()
	freshRead.b31Power = b31CurrentPowerOf(ctx.Game, id)
}

func init() {
	probe := func(oracle, name string, step func(ctx *Context, id uuid.UUID) error) {
		Register(Spec{
			OracleID: oracle,
			Name:     name,
			Targets:  Clauses(TargetCreature("target creature")),
			OnResolve: func(item *game.StackItem, ctx *Context) error {
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
	probe(freshPumpProbe, "Pump Then Read", func(ctx *Context, id uuid.UUID) error {
		return BoostUntilEOT{Target: id, Power: 1, Label: "+1/+0 until end of turn"}.Apply(ctx)
	})
	probe(freshCounterProbe, "Counter Then Read", func(ctx *Context, id uuid.UUID) error {
		return ctx.Game.AddCounterForEffect(id, "+1/+1", 1)
	})
	// "Put the top card of your library onto the battlefield" with a
	// Glorious Anthem there: the anthem applies at once (CR 611.3a).
	probe(freshAnthemProbe, "Anthem Then Read", func(ctx *Context, id uuid.UUID) error {
		p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
		top, err := p.Library.Top()
		if err != nil {
			return nil
		}
		_, err = ctx.Game.PutCardsFromLibraryOntoBattlefieldForEffect([]uuid.UUID{top.InstanceID}, game.LibraryEntryOptions{})
		return err
	})
}

func castFreshProbe(t *testing.T, g *game.Game, oracle string, target uuid.UUID) {
	t.Helper()
	freshRead.power, freshRead.toughness, freshRead.b31Power = -1, -1, -1
	castCatalogSpell(t, g, "Probe", "Instant", oracle, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
}

func TestResolutionReadsPowerAfterAPumpInTheSameResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castFreshProbe(t, g, freshPumpProbe, bear)
	if freshRead.power != 3 || freshRead.b31Power != 3 {
		t.Errorf("power read after +1/+0 in the same resolution: LookupCardForEffect %d, b31CurrentPowerOf %d, want 3",
			freshRead.power, freshRead.b31Power)
	}
}

func TestResolutionReadsToughnessAfterACounterInTheSameResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castFreshProbe(t, g, freshCounterProbe, bear)
	if freshRead.power != 3 || freshRead.toughness != 3 {
		t.Errorf("P/T read after a +1/+1 counter in the same resolution = %d/%d, want 3/3", freshRead.power, freshRead.toughness)
	}
}

func TestResolutionReadsAnAnthemThatEnteredInTheSameResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
			OracleID: gloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	castFreshProbe(t, g, freshAnthemProbe, bear)
	if !onBattlefieldByName(g, "Glorious Anthem") {
		t.Fatal("the probe did not put Glorious Anthem onto the battlefield")
	}
	if freshRead.power != 3 || freshRead.toughness != 3 {
		t.Errorf("P/T read after Glorious Anthem entered in the same resolution = %d/%d, want 3/3", freshRead.power, freshRead.toughness)
	}
}
