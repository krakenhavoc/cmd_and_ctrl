package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// voltron_test.go — S25 (#77). The engine's indestructible rules are
// pinned in server/internal/game/indestructible_test.go; these cases
// drive the CARDS through a real cast and check the grants land on
// the right permanents and nothing else.

const (
	HeroicInterventionOracle = "24882fa2-3fe9-4c1b-aa3d-0e6488b9db27"
	BlossomingDefenseOracle  = "5a851367-1c4a-4cc9-a9c3-3d2775986b4c"
	SnakeskinVeilOracle      = "1e6a24be-8281-41c1-a5ba-b68f0ef1d7b8"
	MakeAStandOracle         = "531f78d5-5004-4b02-99c7-b390cb342fd9"
	TamiyosSafekeepingOracle = "bb2b324c-970a-4920-884e-c92ba49669f0"
	AvacynOracle             = "216cb26e-8da9-478b-bfbc-8030f7adee72"
	BastionProtectorOracle   = "858f53ef-3fec-4aa0-867d-b7f040614b3c"
	ZurgoOracle              = "6c48d888-9f5d-43f4-adbd-61dbdba09260"
	UrilOracle               = "4308a020-48cf-45fe-8074-dd5d0ac6d12d"
)

// pushVoltronCreature puts a 2/2 on the battlefield under `owner`
// with a battlefield timestamp, so the layer engine sorts it.
func PushVoltronCreature(g *game.Game, owner *game.Player, name string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
}

// currentPower reads the counter-inclusive power the combat engine
// and the SBAs see (Effective().Power plus the 7d counter delta),
// as distinct from effectivePower's post-layer-only read.
func CurrentPower(t *testing.T, g *game.Game, cardID uuid.UUID) int {
	t.Helper()
	var p int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			p = c.CurrentPower()
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return p
}
