package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func castFreshProbe(t *testing.T, g *game.Game, oracle string, target uuid.UUID) {
	t.Helper()
	effectstest.FreshRead.Power, effectstest.FreshRead.Toughness, effectstest.FreshRead.B31Power = -1, -1, -1
	effectstest.CastCatalogSpell(t, g, "Probe", "Instant", oracle, []game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestResolutionReadsPowerAfterAPumpInTheSameResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castFreshProbe(t, g, effectstest.FreshPumpProbe, bear)
	if effectstest.FreshRead.Power != 3 || effectstest.FreshRead.B31Power != 3 {
		t.Errorf("power read after +1/+0 in the same resolution: LookupCardForEffect %d, b31CurrentPowerOf %d, want 3",
			effectstest.FreshRead.Power, effectstest.FreshRead.B31Power)
	}
}

func TestResolutionReadsToughnessAfterACounterInTheSameResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castFreshProbe(t, g, effectstest.FreshCounterProbe, bear)
	if effectstest.FreshRead.Power != 3 || effectstest.FreshRead.Toughness != 3 {
		t.Errorf("P/T read after a +1/+1 counter in the same resolution = %d/%d, want 3/3", effectstest.FreshRead.Power, effectstest.FreshRead.Toughness)
	}
}

func TestResolutionReadsAnAnthemThatEnteredInTheSameResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Glorious Anthem", TypeLine: "Enchantment",
			OracleID: effectstest.GloriousAnthemOracle, Owner: me.ID, Controller: me.ID,
		})
	})
	castFreshProbe(t, g, effectstest.FreshAnthemProbe, bear)
	if !effectstest.OnBattlefieldByName(g, "Glorious Anthem") {
		t.Fatal("the probe did not put Glorious Anthem onto the battlefield")
	}
	if effectstest.FreshRead.Power != 3 || effectstest.FreshRead.Toughness != 3 {
		t.Errorf("P/T read after Glorious Anthem entered in the same resolution = %d/%d, want 3/3", effectstest.FreshRead.Power, effectstest.FreshRead.Toughness)
	}
}
