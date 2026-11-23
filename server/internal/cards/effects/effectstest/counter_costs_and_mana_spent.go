package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_costs_and_mana_spent_test.go — the catalog half of #789 and
// #761. One assertion per converted card, against the real registered
// spec: the engine-side contract is pinned in game/counter_cost_rest_test.go
// and game/mana_spent_test.go.

const (
	VividCreekOracle     = "2da7c49f-cc1e-45d9-9cbf-067e92b0daef"
	VividGroveOracle     = "b7a68899-c0d3-49e0-854b-19268ae9b89d"
	RamosOracle          = "3ed41d2d-211b-4013-8562-8c64d54cc43a"
	MageRingOracle       = "136596a0-b179-40be-b42d-c0b992621c95"
	DevotedDruidOracle   = "cb814e16-acf7-41d5-a357-1323dcc369f3"
	HopefulInitiateOrcle = "317169b8-9014-48e6-862b-ca21b706846e"
	PainfulTruthsOracle  = "58a2d15f-1b4b-4303-b524-8cc0d1b3e7d2"
	EtchedOracleOracle   = "7ecfa47e-1165-46a6-884c-290b1c14d020"
	VexingBaubleOracle   = "4514777d-0589-4631-978b-ff244167c176"
	SlayingFireOracle    = "1e94e647-9150-4b22-aa9f-d195f64fb20a"
)

// floatForTest drops one token per rune into a seat's pool.
func FloatForTest(g *game.Game, p *game.Player, spec string) {
	g.WithWriteLock(func() {
		for _, r := range spec {
			p.ManaPool.AddMana(game.ManaToken{Color: string(r), Source: uuid.New()})
		}
	})
}

// untapForTest untaps a battlefield permanent directly — b16Tap's
// inverse, for a fixture that needs an untapped land after an
// enters-tapped replacement.
func UntapForTest(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = false
		}
	}
}

// activateManaFor fires a mana ability and fails the test on refusal.
func ActivateManaFor(t *testing.T, g *game.Game, controller, card uuid.UUID, idx int, params game.ManaAbilityParams) {
	t.Helper()
	if err := g.ActivateManaAbility(controller, card, idx, params); err != nil {
		t.Fatalf("activate mana ability %d: %v", idx, err)
	}
}
